package types

import "fmt"

// the Transaction represent an EIP-1559 transaction
type Transaction struct {
	ChainID              uint64
	Nonce                uint64
	MaxPriorityFeePerGas uint64 // Tip to miner
	MaxFeePerGas         uint64 // Max total fee willing to pay
	GasLimit             uint64
	To                   string
	Value                uint64
	Data                 []byte
	From                 string
}

// EffectiveGasPrice calculates the actual gas price paid
func (tx *Transaction) EffectiveGasPrice(baseFee uint64) uint64 {

	// priority fee is capped by (maxFee - baseFee)
	priorityFee := tx.MaxPriorityFeePerGas
	if tx.MaxFeePerGas-baseFee < priorityFee {
		priorityFee = tx.MaxFeePerGas - baseFee
	}

	return priorityFee + baseFee
}

func (tx *Transaction) EffectivePriorityFee(baseFee uint64) uint64 {
	priorityFee := tx.MaxPriorityFeePerGas
	if tx.MaxFeePerGas-baseFee < priorityFee {
		priorityFee = tx.MaxFeePerGas - baseFee
	}
	return priorityFee
}

// Validate performs basic transaction validation
func (tx *Transaction) Validate(baseFee uint64) error {
	if tx.MaxFeePerGas < baseFee {
		return fmt.Errorf("max fee pre gas %d less than base fee %d", tx.MaxFeePerGas, baseFee)
	}

	if tx.MaxFeePerGas < tx.MaxPriorityFeePerGas {
		return fmt.Errorf("max fee pre gas %d less than max priority fee %d", tx.MaxFeePerGas, tx.MaxPriorityFeePerGas)
	}

	if tx.GasLimit == 0 {
		return fmt.Errorf("gas limit cannot be zero")
	}

	return nil
}

func (tx *Transaction) MaxCost() uint64 {
	return tx.GasLimit*tx.MaxFeePerGas + tx.Value
}
