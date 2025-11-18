package test

import (
	"testing"

	"github.com/EIPs-CodeLab/EIP-1559/internal/basefee"
	"github.com/EIPs-CodeLab/EIP-1559/internal/types"
	"github.com/EIPs-CodeLab/EIP-1559/pkg/constants"
)

func TestBaseFeeCalculation(t *testing.T) {
	tests := []struct {
		name            string
		gasUsed         uint64
		gasLimit        uint64
		currentBaseFee  uint64
		expectedBaseFee uint64
	}{
		{
			name:            "at target - no change",
			gasUsed:         15_000_000,
			gasLimit:        30_000_000,
			currentBaseFee:  1_000_000_000,
			expectedBaseFee: 1_000_000_000,
		},
		{
			name:            "above target - increase",
			gasUsed:         20_000_000,
			gasLimit:        30_000_000,
			currentBaseFee:  1_000_000_000,
			expectedBaseFee: 1_083_333_333,
		},
		{
			name:            "below target - decrease",
			gasUsed:         10_000_000,
			gasLimit:        30_000_000,
			currentBaseFee:  1_000_000_000,
			expectedBaseFee: 916_666_666,
		},
		{
			name:            "full block - max increase",
			gasUsed:         30_000_000,
			gasLimit:        30_000_000,
			currentBaseFee:  1_000_000_000,
			expectedBaseFee: 1_125_000_000,
		},
		{
			name:            "empty block - max decrease",
			gasUsed:         0,
			gasLimit:        30_000_000,
			currentBaseFee:  1_000_000_000,
			expectedBaseFee: 875_000_000,
		},
		{
			name:            "very low base fee decrease",
			gasUsed:         0,
			gasLimit:        30_000_000,
			currentBaseFee:  100,
			expectedBaseFee: 87,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			parent := &types.Block{
				Number:   constants.ForkBlockNumber,
				GasLimit: tt.gasLimit,
				GasUsed:  tt.gasUsed,
				BaseFee:  tt.currentBaseFee,
			}

			result := basefee.Calculate(parent)

			if result != tt.expectedBaseFee {
				t.Errorf("expected base fee %d, got %d", tt.expectedBaseFee, result)
			}
		})
	}
}

func TestBaseFeeNeverNegative(t *testing.T) {
	parent := &types.Block{
		Number:   constants.ForkBlockNumber,
		GasLimit: 30_000_000,
		GasUsed:  0,
		BaseFee:  10, // Very low base fee
	}

	result := basefee.Calculate(parent)

	if result > parent.BaseFee {
		t.Errorf("base fee should decrease, but increased from %d to %d", parent.BaseFee, result)
	}
}

func TestBaseFeeMinimumIncrease(t *testing.T) {
	parent := &types.Block{
		Number:   constants.ForkBlockNumber,
		GasLimit: 30_000_000,
		GasUsed:  15_000_001, // Just 1 wei above target
		BaseFee:  1,
	}

	result := basefee.Calculate(parent)

	// Should increase by at least 1 wei
	if result <= parent.BaseFee {
		t.Errorf("base fee should increase by at least 1, but got %d", result)
	}
}
