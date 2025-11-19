package test

import (
	"testing"

	"github.com/EIPs-CodeLab/EIP-1559/internal/executor"
	"github.com/EIPs-CodeLab/EIP-1559/internal/types"
)

func TestExecuteTransaction(t *testing.T) {
	// Setup state
	state := types.NewState()
	state.SetAccount("0xAlice", types.NewAccount("0xAlice", 1_000_000_000_000_000))
	state.SetAccount("0xBob", types.NewAccount("0xBob", 0))
	state.SetAccount("0xMiner", types.NewAccount("0xMiner", 0))

	// Create block
	block := &types.Block{
		Number:   1,
		GasLimit: 30_000_000,
		BaseFee:  1_000_000_000,
		Miner:    "0xMiner",
	}

	// Create transaction
	tx := &types.Transaction{
		From:                 "0xAlice",
		To:                   "0xBob",
		Nonce:                0,
		MaxPriorityFeePerGas: 2_000_000_000,
		MaxFeePerGas:         5_000_000_000,
		GasLimit:             21_000,
		Value:                1_000,
	}

	initialAliceBalance := state.GetBalance("0xAlice")

	// Execute transaction
	result := executor.ExecuteTransaction(tx, block, state)

	// Assertions
	if !result.Success {
		t.Fatalf("transaction should succeed, got error: %v", result.Error)
	}

	if result.GasUsed != 21_000 {
		t.Errorf("expected gas used 21000, got %d", result.GasUsed)
	}

	// Check base fee was burned (not given to anyone)
	expectedBurned := result.GasUsed * block.BaseFee
	if result.BaseFeeAmount != expectedBurned {
		t.Errorf("expected burned %d, got %d", expectedBurned, result.BaseFeeAmount)
	}

	// Check miner received tip
	minerBalance := state.GetBalance("0xMiner")
	if minerBalance != result.TipAmount {
		t.Errorf("expected miner balance %d, got %d", result.TipAmount, minerBalance)
	}

	// Check Bob received value
	bobBalance := state.GetBalance("0xBob")
	if bobBalance != tx.Value {
		t.Errorf("expected Bob balance %d, got %d", tx.Value, bobBalance)
	}

	// Check Alice paid correctly
	aliceBalance := state.GetBalance("0xAlice")
	expectedAlicePaid := result.GasUsed*tx.EffectiveGasPrice(block.BaseFee) + tx.Value
	expectedAliceBalance := initialAliceBalance - expectedAlicePaid

	if aliceBalance != expectedAliceBalance {
		t.Errorf("expected Alice balance %d, got %d", expectedAliceBalance, aliceBalance)
	}

	// Check nonce incremented
	if state.GetNonce("0xAlice") != 1 {
		t.Errorf("expected Alice nonce 1, got %d", state.GetNonce("0xAlice"))
	}
}

func TestExecuteTransactionInsufficientFunds(t *testing.T) {
	state := types.NewState()
	state.SetAccount("0xAlice", types.NewAccount("0xAlice", 1_000)) // Very low balance
	state.SetAccount("0xMiner", types.NewAccount("0xMiner", 0))

	block := &types.Block{
		Number:   1,
		GasLimit: 30_000_000,
		BaseFee:  1_000_000_000,
		Miner:    "0xMiner",
	}

	tx := &types.Transaction{
		From:                 "0xAlice",
		To:                   "0xBob",
		Nonce:                0,
		MaxPriorityFeePerGas: 2_000_000_000,
		MaxFeePerGas:         5_000_000_000,
		GasLimit:             21_000,
		Value:                1_000,
	}

	result := executor.ExecuteTransaction(tx, block, state)

	if result.Success {
		t.Error("transaction should fail due to insufficient funds")
	}

	if result.Error == nil {
		t.Error("expected error, got nil")
	}
}
