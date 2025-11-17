package constants

const (
	// InitialBaseFee is the starting base fee (1 Gwei)
	InitialBaseFee uint64 = 1_000_000_000

	// BaseFeeChangeDenominator controls max base fee change per block (12.5% = 1/8)
	BaseFeeChangeDenominator uint64 = 8

	// ElasticityMultiplier allows blocks to be 2x target size
	ElasticityMultiplier uint64 = 2

	// MinGasLimit is the minimum gas limit for a block
	MinGasLimit uint64 = 5000

	// ForkBlockNumber is when EIP-1559 activates (London fork)
	ForkBlockNumber uint64 = 12_965_000

	// GasLimitBoundDivisor limits how much gas limit can change per block (1/1024)
	GasLimitBoundDivisor uint64 = 1024
)
