package types

import "fmt"

// Block represents a block with EIP-1559 base fee
type Block struct {
	Number       uint64
	ParentHash   string
	Hash         string
	GasLimit     uint64
	GasUsed      uint64
	BaseFee      uint64 // EIP-1559 base fee
	Transactions []*Transaction
	Miner        string
	Timestamp    uint64
}

// NewBlock creates a new block
func NewBlock(number uint64, parentHash string, gasLimit uint64, baseFee uint64, miner string) *Block {
	return &Block{
		Number:       number,
		ParentHash:   parentHash,
		GasLimit:     gasLimit,
		BaseFee:      baseFee,
		Miner:        miner,
		Transactions: make([]*Transaction, 0),
	}
}

// AddTransaction adds a transaction to the block
func (b *Block) AddTransaction(tx *Transaction) error {
	// Check if adding this tx would exceed gas limit
	txGas := tx.GasLimit
	if b.GasUsed+txGas > b.GasLimit {
		return fmt.Errorf("transaction would exceed block gas limit")
	}

	b.Transactions = append(b.Transactions, tx)
	b.GasUsed += txGas
	return nil
}

// GasTarget returns the target gas usage (50% of limit)
func (b *Block) GasTarget() uint64 {
	return b.GasLimit / 2 // ElasticityMultiplier = 2
}

// IsAboveTarget returns true if block used more than target gas
func (b *Block) IsAboveTarget() bool {
	return b.GasUsed > b.GasTarget()
}

// IsBelowTarget returns true if block used less than target gas
func (b *Block) IsBelowTarget() bool {
	return b.GasUsed < b.GasTarget()
}

// Utilization returns gas utilization percentage (0-100)
func (b *Block) Utilization() float64 {
	if b.GasLimit == 0 {
		return 0
	}
	return float64(b.GasUsed) / float64(b.GasLimit) * 100
}
