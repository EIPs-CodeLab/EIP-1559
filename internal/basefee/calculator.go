package basefee

import (
	"github.com/EIPs-CodeLab/EIP-1559/internal/types"
	"github.com/EIPs-CodeLab/EIP-1559/pkg/constants"
)

// Calculate computes the base fee for the next block based on parent block
func Calculate(parent *types.Block) uint64 {
	// Special case: fork block
	if parent.Number+1 == constants.ForkBlockNumber {
		return constants.InitialBaseFee
	}

	parentGasTarget := parent.GasLimit / constants.ElasticityMultiplier

	// If parent block used exactly the target, base fee stays the same
	if parent.GasUsed == parentGasTarget {
		return parent.BaseFee
	}

	var newBaseFee uint64

	if parent.GasUsed > parentGasTarget {
		// Block used more than target - increase base fee
		gasUsedDelta := parent.GasUsed - parentGasTarget
		baseFeePerGasDelta := max(
			(parent.BaseFee*gasUsedDelta)/parentGasTarget/constants.BaseFeeChangeDenominator,
			1, // Minimum increase of 1 wei
		)
		newBaseFee = parent.BaseFee + baseFeePerGasDelta
	} else {
		// Block used less than target - decrease base fee
		gasUsedDelta := parentGasTarget - parent.GasUsed
		baseFeePerGasDelta := (parent.BaseFee * gasUsedDelta) / parentGasTarget / constants.BaseFeeChangeDenominator

		// Ensure base fee doesn't go negative
		if baseFeePerGasDelta > parent.BaseFee {
			newBaseFee = 0
		} else {
			newBaseFee = parent.BaseFee - baseFeePerGasDelta
		}
	}

	return newBaseFee
}

// max returns the larger of two uint64 values
func max(a, b uint64) uint64 {
	if a > b {
		return a
	}
	return b
}

// CalculateForBlocks simulates base fee changes over multiple blocks
func CalculateForBlocks(initialBlock *types.Block, gasUsedSequence []uint64) []uint64 {
	baseFees := make([]uint64, len(gasUsedSequence))
	currentBlock := initialBlock

	for i, gasUsed := range gasUsedSequence {
		// Calculate next base fee
		nextBaseFee := Calculate(currentBlock)
		baseFees[i] = nextBaseFee

		// Create next block for simulation
		currentBlock = &types.Block{
			Number:   currentBlock.Number + 1,
			GasLimit: currentBlock.GasLimit,
			GasUsed:  gasUsed,
			BaseFee:  nextBaseFee,
		}
	}

	return baseFees
}
