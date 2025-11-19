# EIP-1559: Fee Market Change Implementation

[![Go Version](https://img.shields.io/badge/Go-1.21+-00ADD8?style=flat&logo=go)](https://go.dev/)
[![License](https://img.shields.io/badge/License-MIT-blue.svg)](LICENSE)
[![Tests](https://img.shields.io/badge/Tests-Passing-brightgreen.svg)]()

A Go implementation of EIP-1559, Ethereum's fee market mechanism that introduced base fees, dynamic block sizes, and fee burning.

## Overview

EIP-1559 fundamentally changed Ethereum's transaction fee mechanism from a simple auction system to a more predictable model with a protocol-enforced base fee that adjusts based on network congestion. This implementation provides a complete simulation of the EIP-1559 fee market dynamics.

The base fee automatically adjusts block by block: increasing when blocks exceed the gas target and decreasing when they fall below it. This predictable adjustment mechanism allows wallets to reliably estimate fees, eliminating the need for complex bidding strategies.

A critical feature is that the base fee is burned rather than paid to miners. This ensures ETH remains the sole currency for transaction fees, reduces miner extractable value manipulation, and can make ETH deflationary during periods of high network usage.

## Key Features

**Dynamic Base Fee Calculation**
- Automatic adjustment based on network congestion
- Maximum 12.5% change per block
- Predictable fee estimation for wallets

**Transaction Validation**
- EIP-1559 transaction type support
- Fee validation (maxFeePerGas, maxPriorityFeePerGas)
- Balance and nonce verification

**Block Validation**
- Gas limit change constraints (1/1024 per block)
- Base fee correctness verification
- Gas usage enforcement

**Fee Burning Mechanism**
- Base fee destruction (not paid to miners)
- Priority fee (tip) paid to block producers
- Accurate accounting of burned vs. distributed fees

**Network Simulation**
- Multi-block processing
- Configurable congestion scenarios
- Detailed metrics and reporting

## Specification

### Base Fee Calculation

The base fee adjusts according to parent block usage:

```
parentGasTarget = parentGasLimit / ELASTICITY_MULTIPLIER (2)

if parentGasUsed == parentGasTarget:
    baseFee = parentBaseFee

if parentGasUsed > parentGasTarget:
    gasUsedDelta = parentGasUsed - parentGasTarget
    baseFeePerGasDelta = max(
        parentBaseFee * gasUsedDelta / parentGasTarget / 8,
        1
    )
    baseFee = parentBaseFee + baseFeePerGasDelta

if parentGasUsed < parentGasTarget:
    gasUsedDelta = parentGasTarget - parentGasUsed
    baseFeePerGasDelta = parentBaseFee * gasUsedDelta / parentGasTarget / 8
    baseFee = max(parentBaseFee - baseFeePerGasDelta, 0)
```

### Transaction Format

EIP-1559 introduces Type 2 transactions with separate fee parameters:

```go
type Transaction struct {
    ChainID              uint64
    Nonce                uint64
    MaxPriorityFeePerGas uint64  // Tip to miner
    MaxFeePerGas         uint64  // Maximum total fee
    GasLimit             uint64
    To                   string
    Value                uint64
    Data                 []byte
}
```

### Fee Calculation

For each transaction:

```
priorityFee = min(maxPriorityFeePerGas, maxFeePerGas - baseFee)
effectiveGasPrice = priorityFee + baseFee

User pays: gasUsed * effectiveGasPrice
Miner receives: gasUsed * priorityFee
Burned: gasUsed * baseFee
```

### Constants

```go
const (
    InitialBaseFee              = 1_000_000_000  // 1 Gwei
    BaseFeeChangeDenominator    = 8              // 12.5% max change
    ElasticityMultiplier        = 2              // 2x target size
    MinGasLimit                 = 5_000
    GasLimitBoundDivisor        = 1_024
)
```

## Project Structure

```
eip-1559/
├── cmd/
│   └── simulator/
│       └── main.go                 # CLI simulator
├── internal/
│   ├── types/
│   │   ├── transaction.go          # Transaction types
│   │   ├── block.go                # Block with BaseFee
│   │   └── account.go              # Account state
│   ├── basefee/
│   │   └── calculator.go           # Base fee calculation
│   ├── validator/
│   │   └── validator.go            # Validation logic
│   └── executor/
│       └── executor.go             # Transaction execution
├── pkg/
│   └── constants/
│       └── params.go               # EIP-1559 constants
├── test/
│   ├── basefee_test.go
│   ├── validator_test.go
│   ├── executor_test.go
│   └── integration_test.go
└── Makefile
```

## Installation

### Prerequisites

- Go 1.21 or higher

### Setup

```bash
# Clone the repository
git clone https://github.com/EIPs-CodeLab/eip-1559.git
cd eip-1559

# Initialize Go module
go mod download

# Build the project
make build
```

## Usage

### Build

Compile the simulator binary:

```bash
make build
```

This creates `bin/simulator` executable.

### Run Tests

Execute the complete test suite:

```bash
make test
```

Run tests with verbose output:

```bash
go test -v ./test/...
```

### Run Simulator

Basic simulation (10 blocks, target gas usage):

```bash
make run
```

Verbose output with detailed metrics:

```bash
make run-verbose
```

Simulate network congestion (20 blocks, high gas usage):

```bash
make run-congestion
```

Custom parameters:

```bash
go run cmd/simulator/main.go -blocks=50 -gas=20000000 -verbose
```

### Command Line Options

```
-blocks int      Number of blocks to simulate (default: 10)
-gas uint        Gas used per block (default: 15000000)
-verbose         Enable verbose output
```

### Example Output (sample run)

```
Block      | BaseFee      | GasUsed      | Usage%   | Burned           | Tips
-----------|--------------|--------------|----------|------------------|------------------
12965000  | 1000000000   | 21000        |    0.07% | 21000000000000   | 42000000000000
12965001  | 875175000    | 21000        |    0.07% | 18378675000000   | 42000000000000
12965002  | 765931281    | 21000        |    0.07% | 16084556901000   | 42000000000000
12965003  | 670323909    | 21000        |    0.07% | 14076802089000   | 42000000000000
12965004  | 586650728    | 21000        |    0.07% | 12319665288000   | 42000000000000

Summary
Total ETH burned: 81859699278000 wei
Total tips paid:  210000000000000 wei
Final base fee:   586650728 wei (0.59 Gwei)

Final balances:
    Alice (sender):    708140300717000 wei
    Bob (recipient):   5000 wei
    Miner:             210000000000000 wei
```

### Clean Build Artifacts

```bash
make clean
```

### Format Code

```bash
make fmt
```

## Rationale

### Why EIP-1559?

**Predictable Fees**
The automatic base fee adjustment creates predictability. Wallets can reliably estimate fees without complex auction strategies, improving user experience significantly.

**Reduced Overpayment**
First-price auctions are inefficient. Users often overpay to ensure inclusion. EIP-1559's separate priority fee mechanism allows users to pay only what is necessary.

**Fee Burning**
Burning the base fee removes miner incentive to manipulate fees. It also ensures only ETH can pay for Ethereum transactions, reinforcing ETH's economic position.

**Elastic Block Sizes**
Allowing blocks to temporarily exceed the target (up to 2x) smooths congestion spikes. This reduces delays without compromising long-term average block size.

### Design Decisions

**12.5% Maximum Change**
The base fee can change by at most 1/8th per block. This constraint ensures predictability while allowing responsiveness to demand changes.

**Base Fee Burn vs. Miner Payment**
Burning prevents miners from stuffing blocks with fake transactions to manipulate fees upward. The priority fee alone provides sufficient miner incentive.

**Elasticity Multiplier of 2**
This allows short-term bursts while maintaining the long-term average. Higher multipliers would require more aggressive base fee adjustments.

## Security Considerations

### Increased Block Complexity

EIP-1559 allows blocks up to twice the target size during congestion. Clients must handle these larger blocks without errors. This implementation includes validation to prevent blocks exceeding the maximum.

### Transaction Ordering

With most users paying similar base fees, transaction ordering depends on priority fees and arrival time. Miners should sort by priority fee to maximize revenue, with timestamps as tiebreakers to prevent spam attacks.

### Empty Block Mining Attack

Miners could theoretically mine empty blocks to drive base fees to zero, then mine half-full blocks and profit from priority fees. However, this attack requires over 50% hash power and any defector profits more. The elasticity multiplier can be increased if this attack is attempted.

### ETH Supply Uncertainty

Burning base fees makes ETH supply unpredictable. If burn exceeds issuance, ETH becomes deflationary. If issuance exceeds burn, it remains inflationary. This removes guarantees about long-term supply but aligns incentives better.

### GASPRICE Opcode Change

The GASPRICE (0x3a) opcode now returns the effective gas price paid by the sender, not the amount received by the miner. Smart contracts relying on GASPRICE must account for this change.

## Testing

The implementation includes comprehensive tests:

**Unit Tests**
- Base fee calculation for various scenarios
- Transaction validation rules
- Block validation constraints
- Fee burning mechanics

**Integration Tests**
- Multi-block processing
- State transitions
- End-to-end transaction flow

**Test Coverage**
- Base fee adjustment (above/below target)
- Minimum and maximum changes
- Edge cases (empty blocks, full blocks)
- Fee validation errors
- Balance verification

Run all tests:

```bash
make test
```

## References

**Official Specification**
- [EIP-1559: Fee market change for ETH 1.0 chain](https://eips.ethereum.org/EIPS/eip-1559)

**Related EIPs**
- [EIP-2718: Typed Transaction Envelope](https://eips.ethereum.org/EIPS/eip-2718)
- [EIP-2930: Optional access lists](https://eips.ethereum.org/EIPS/eip-2930)

**Research & Analysis**
- [EIP-1559 Agent-Based Model](https://ethereum.github.io/abm1559/notebooks/eip1559.html)
- [Ethereum Foundation Research](https://ethereum.org/en/developers/docs/gas/)

**Implementation Resources**
- [Go Ethereum (Geth) Implementation](https://github.com/ethereum/go-ethereum)
- [Ethereum Yellow Paper](https://ethereum.github.io/yellowpaper/paper.pdf)

## License

This project is licensed under the MIT License. See the [LICENSE](LICENSE) file for details.

The EIP-1559 specification is licensed under CC0-1.0.

## Citation

Please cite this implementation as:

```
EIPs-CodeLab, "EIP-1559 Go Implementation," 2024. 
[Online]. Available: https://github.com/EIPs-CodeLab/eip-1559
```

Cite the original EIP as:

```
Vitalik Buterin (@vbuterin), Eric Conner (@econoar), Rick Dudley (@AFDudley), 
Matthew Slipper (@mslipper), Ian Norden (@i-norden), Abdelhamid Bakhta (@abdelhamidbakhta), 
"EIP-1559: Fee market change for ETH 1.0 chain," Ethereum Improvement Proposals, 
no. 1559, April 2019. [Online serial]. Available: https://eips.ethereum.org/EIPS/eip-1559
```

## Contributing

Contributions are welcome. Please follow these guidelines:

1. Fork the repository
2. Create a feature branch
3. Write tests for new functionality
4. Ensure all tests pass
5. Submit a pull request

## Acknowledgments

Thanks to the EIP-1559 authors and the Ethereum community for the specification and research that made this implementation possible.

---

**Built by [EIPs-CodeLab](https://github.com/EIPs-CodeLab) - Translating EIPs into Code**