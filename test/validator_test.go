package test

import (
	"testing"

	"github.com/EIPs-CodeLab/EIP-1559/internal/types"
	"github.com/EIPs-CodeLab/EIP-1559/internal/validator"
	"github.com/EIPs-CodeLab/EIP-1559/pkg/constants"
)

func TestValidateTransaction(t *testing.T) {
	state := types.NewState()
	// Give Alice sufficient upfront funds to cover max-fee * gas for test transactions
	state.SetAccount("0xAlice", types.NewAccount("0xAlice", 200_000_000_000_000))

	baseFee := uint64(1_000_000_000)

	tests := []struct {
		name    string
		tx      *types.Transaction
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid transaction",
			tx: &types.Transaction{
				From:                 "0xAlice",
				Nonce:                0,
				MaxPriorityFeePerGas: 2_000_000_000,
				MaxFeePerGas:         5_000_000_000,
				GasLimit:             21_000,
				Value:                1_000,
			},
			wantErr: false,
		},
		{
			name: "max fee less than base fee",
			tx: &types.Transaction{
				From:                 "0xAlice",
				Nonce:                0,
				MaxPriorityFeePerGas: 500_000_000,
				MaxFeePerGas:         500_000_000, // Less than base fee
				GasLimit:             21_000,
				Value:                1_000,
			},
			wantErr: true,
			errMsg:  "less than base fee",
		},
		{
			name: "max fee less than priority fee",
			tx: &types.Transaction{
				From:                 "0xAlice",
				Nonce:                0,
				MaxPriorityFeePerGas: 5_000_000_000,
				MaxFeePerGas:         2_000_000_000, // Less than priority fee
				GasLimit:             21_000,
				Value:                1_000,
			},
			wantErr: true,
			errMsg:  "less than max priority fee",
		},
		{
			name: "insufficient balance",
			tx: &types.Transaction{
				From:                 "0xBob", // No balance
				Nonce:                0,
				MaxPriorityFeePerGas: 2_000_000_000,
				MaxFeePerGas:         5_000_000_000,
				GasLimit:             21_000,
				Value:                1_000,
			},
			wantErr: true,
			errMsg:  "insufficient funds",
		},
		{
			name: "invalid nonce",
			tx: &types.Transaction{
				From:                 "0xAlice",
				Nonce:                5, // Wrong nonce
				MaxPriorityFeePerGas: 2_000_000_000,
				MaxFeePerGas:         5_000_000_000,
				GasLimit:             21_000,
				Value:                1_000,
			},
			wantErr: true,
			errMsg:  "invalid nonce",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateTransaction(tt.tx, baseFee, state)

			if tt.wantErr && err == nil {
				t.Errorf("expected error containing '%s', got nil", tt.errMsg)
			}

			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if tt.wantErr && err != nil {
				// Check error message contains expected text
				if tt.errMsg != "" && !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			}
		})
	}
}

func TestValidateBlock(t *testing.T) {
	parent := &types.Block{
		Number:   constants.ForkBlockNumber,
		Hash:     "0xparent",
		GasLimit: 30_000_000,
		GasUsed:  15_000_000,
		BaseFee:  1_000_000_000,
	}

	tests := []struct {
		name    string
		block   *types.Block
		wantErr bool
		errMsg  string
	}{
		{
			name: "valid block",
			block: &types.Block{
				Number:     parent.Number + 1,
				ParentHash: parent.Hash,
				GasLimit:   30_000_000,
				GasUsed:    20_000_000,
				// Base fee is calculated from the parent block's usage; parent used exactly target
				// so base fee should remain unchanged.
				BaseFee:    1_000_000_000,
			},
			wantErr: false,
		},
		{
			name: "invalid block number",
			block: &types.Block{
				Number:     parent.Number + 2, // Skip a block
				ParentHash: parent.Hash,
				GasLimit:   30_000_000,
				GasUsed:    15_000_000,
				BaseFee:    1_000_000_000,
			},
			wantErr: true,
			errMsg:  "invalid block number",
		},
		{
			name: "gas used exceeds limit",
			block: &types.Block{
				Number:     parent.Number + 1,
				ParentHash: parent.Hash,
				GasLimit:   30_000_000,
				GasUsed:    31_000_000, // Exceeds limit
				BaseFee:    1_000_000_000,
			},
			wantErr: true,
			errMsg:  "exceeds gas limit",
		},
		{
			name: "gas limit increased too much",
			block: &types.Block{
				Number:     parent.Number + 1,
				ParentHash: parent.Hash,
				GasLimit:   35_000_000, // Too much increase
				GasUsed:    15_000_000,
				BaseFee:    1_000_000_000,
			},
			wantErr: true,
			errMsg:  "increased too much",
		},
		{
			name: "invalid base fee",
			block: &types.Block{
				Number:     parent.Number + 1,
				ParentHash: parent.Hash,
				GasLimit:   30_000_000,
				GasUsed:    15_000_000,
				BaseFee:    2_000_000_000, // Wrong base fee
			},
			wantErr: true,
			errMsg:  "invalid base fee",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := validator.ValidateBlock(tt.block, parent)

			if tt.wantErr && err == nil {
				t.Errorf("expected error, got nil")
			}

			if !tt.wantErr && err != nil {
				t.Errorf("expected no error, got %v", err)
			}

			if tt.wantErr && err != nil && tt.errMsg != "" {
				if !contains(err.Error(), tt.errMsg) {
					t.Errorf("expected error containing '%s', got '%s'", tt.errMsg, err.Error())
				}
			}
		})
	}
}

func contains(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || len(substr) == 0 ||
		(len(s) > 0 && len(substr) > 0 && containsSubstring(s, substr)))
}

func containsSubstring(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}
