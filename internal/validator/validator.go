package validator

import (
	"fmt"

	"github.com/EIPs-CodeLab/EIP-1559/internal/types"
	"github.com/EIPs-CodeLab/EIP-1559/pkg/constants"
)

// ValidateTransaction validates an EIP-1559 transaction
func ValidateTransaction(tx *types.Transaction, baseFee uint64, state *types.State) error {
	// Basic transaction validation
	if err := tx.Validate(baseFee); err != nil {
		return fmt.Errorf("invalid transaction: %w", err)
	}

	// Check sender has enough balance
	sender := state.GetAccount(tx.From)
	maxCost := tx.MaxCost()

	if !sender.CanPay(maxCost) {
		return fmt.Errorf("insufficient funds: have %d, need %d", sender.Balance, maxCost)
	}

	// Check nonce
	if tx.Nonce != sender.Nonce {
		return fmt.Errorf("invalid nonce: have %d, expected %d", tx.Nonce, sender.Nonce)
	}

	return nil
}

// ValidateBlock validates a block according to EIP-1559 rules
func ValidateBlock(block *types.Block, parent *types.Block) error {
	// Validate block number
	if block.Number != parent.Number+1 {
		return fmt.Errorf("invalid block number: expected %d, got %d", parent.Number+1, block.Number)
	}

	// Validate parent hash
	if block.ParentHash != parent.Hash {
		return fmt.Errorf("invalid parent hash")
	}

	// Validate gas used doesn't exceed gas limit
	if block.GasUsed > block.GasLimit {
		return fmt.Errorf("gas used (%d) exceeds gas limit (%d)", block.GasUsed, block.GasLimit)
	}

	// Validate gas limit change (max 1/1024 change per block)
	maxIncrease := parent.GasLimit + parent.GasLimit/constants.GasLimitBoundDivisor
	maxDecrease := parent.GasLimit - parent.GasLimit/constants.GasLimitBoundDivisor

	if block.GasLimit > maxIncrease {
		return fmt.Errorf("gas limit increased too much: parent %d, current %d, max %d",
			parent.GasLimit, block.GasLimit, maxIncrease)
	}

	if block.GasLimit < maxDecrease {
		return fmt.Errorf("gas limit decreased too much: parent %d, current %d, min %d",
			parent.GasLimit, block.GasLimit, maxDecrease)
	}

	// Validate minimum gas limit
	if block.GasLimit < constants.MinGasLimit {
		return fmt.Errorf("gas limit (%d) below minimum (%d)", block.GasLimit, constants.MinGasLimit)
	}

	// Validate base fee (must match calculated value)
	expectedBaseFee := calculateExpectedBaseFee(parent)
	if block.BaseFee != expectedBaseFee {
		return fmt.Errorf("invalid base fee: expected %d, got %d", expectedBaseFee, block.BaseFee)
	}

	return nil
}

// calculateExpectedBaseFee calculates what the base fee should be
func calculateExpectedBaseFee(parent *types.Block) uint64 {
	// Import from basefee package to avoid duplication
	// For now, inline the logic
	parentGasTarget := parent.GasLimit / constants.ElasticityMultiplier

	if parent.GasUsed == parentGasTarget {
		return parent.BaseFee
	}

	if parent.GasUsed > parentGasTarget {
		gasUsedDelta := parent.GasUsed - parentGasTarget
		baseFeePerGasDelta := max(
			(parent.BaseFee*gasUsedDelta)/parentGasTarget/constants.BaseFeeChangeDenominator,
			1,
		)
		return parent.BaseFee + baseFeePerGasDelta
	}

	gasUsedDelta := parentGasTarget - parent.GasUsed
	baseFeePerGasDelta := (parent.BaseFee * gasUsedDelta) / parentGasTarget / constants.BaseFeeChangeDenominator

	if baseFeePerGasDelta > parent.BaseFee {
		return 0
	}
	return parent.BaseFee - baseFeePerGasDelta
}

func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}
