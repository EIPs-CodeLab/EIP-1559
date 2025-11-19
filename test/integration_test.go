package test

import (
	"testing"

	"github.com/EIPs-CodeLab/EIP-1559/internal/basefee"
	"github.com/EIPs-CodeLab/EIP-1559/internal/executor"
	"github.com/EIPs-CodeLab/EIP-1559/internal/types"
	"github.com/EIPs-CodeLab/EIP-1559/internal/validator"
	"github.com/EIPs-CodeLab/EIP-1559/pkg/constants"
)

func TestFullBlockProcessing(t *testing.T) {
	// Initialize state
	state := types.NewState()
	// Give Alice enough balance to cover upfront max-fee * gas + value for multiple transactions
	state.SetAccount("0xAlice", types.NewAccount("0xAlice", 1_000_000_000_000_000))
	state.SetAccount("0xBob", types.NewAccount("0xBob", 0))
	state.SetAccount("0xMiner", types.NewAccount("0xMiner", 0))

	// Create genesis block
	genesisBlock := &types.Block{
		Number:   constants.ForkBlockNumber - 1,
		Hash:     "0xgenesis",
		GasLimit: 30_000_000,
		GasUsed:  15_000_000,
		BaseFee:  constants.InitialBaseFee,
		Miner:    "0xMiner",
	}

	// Process 5 blocks
	currentBlock := genesisBlock
	totalBurned := uint64(0)

	for i := 0; i < 5; i++ {
		// Calculate next base fee
		nextBaseFee := basefee.Calculate(currentBlock)

		// Create next block
		nextBlock := types.NewBlock(
			currentBlock.Number+1,
			currentBlock.Hash,
			30_000_000,
			nextBaseFee,
			"0xMiner",
		)

		// Create transaction
		tx := &types.Transaction{
			From:                 "0xAlice",
			To:                   "0xBob",
			Nonce:                state.GetNonce("0xAlice"),
			MaxPriorityFeePerGas: 2_000_000_000,
			MaxFeePerGas:         nextBaseFee + 10_000_000_000,
			GasLimit:             21_000,
			Value:                1_000,
		}

		// Validate transaction
		if err := validator.ValidateTransaction(tx, nextBaseFee, state); err != nil {
			t.Fatalf("block %d: transaction validation failed: %v", i, err)
		}

		// Add transaction to block
		if err := nextBlock.AddTransaction(tx); err != nil {
			t.Fatalf("block %d: failed to add transaction: %v", i, err)
		}

		// Validate block
		if err := validator.ValidateBlock(nextBlock, currentBlock); err != nil {
			t.Fatalf("block %d: block validation failed: %v", i, err)
		}

		// Execute transaction
		result := executor.ExecuteTransaction(tx, nextBlock, state)
		if !result.Success {
			t.Fatalf("block %d: transaction execution failed: %v", i, result.Error)
		}

		totalBurned += result.BaseFeeAmount

		// Move to next block
		currentBlock = nextBlock
	}

	// Verify final state
	if state.GetNonce("0xAlice") != 5 {
		t.Errorf("expected Alice nonce 5, got %d", state.GetNonce("0xAlice"))
	}

	if state.GetBalance("0xBob") != 5000 { // 5 transactions * 1000 wei
		t.Errorf("expected Bob balance 5000, got %d", state.GetBalance("0xBob"))
	}

	if totalBurned == 0 {
		t.Error("expected some ETH to be burned")
	}

	minerBalance := state.GetBalance("0xMiner")
	if minerBalance == 0 {
		t.Error("expected miner to receive tips")
	}

	t.Logf(" Processed 5 blocks successfully")
	t.Logf("   Total burned: %d wei", totalBurned)
	t.Logf("   Miner earned: %d wei", minerBalance)
}

func TestBaseFeeAdjustmentOverTime(t *testing.T) {
	parent := &types.Block{
		Number:   constants.ForkBlockNumber,
		GasLimit: 30_000_000,
		GasUsed:  15_000_000,
		BaseFee:  1_000_000_000,
	}

	// Simulate congestion (high usage)
	highUsage := []uint64{25_000_000, 28_000_000, 29_000_000}
	baseFees := basefee.CalculateForBlocks(parent, highUsage)

	// Base fee should increase with high usage
	for i := 1; i < len(baseFees); i++ {
		if baseFees[i] <= baseFees[i-1] {
			t.Errorf("base fee should increase with high usage, but didn't at index %d", i)
		}
	}

	// Simulate low usage
	parent.BaseFee = baseFees[len(baseFees)-1]
	lowUsage := []uint64{5_000_000, 3_000_000, 1_000_000}
	baseFees = basefee.CalculateForBlocks(parent, lowUsage)

	// Base fee should decrease with low usage
	for i := 1; i < len(baseFees); i++ {
		if baseFees[i] >= baseFees[i-1] {
			t.Errorf("base fee should decrease with low usage, but didn't at index %d", i)
		}
	}
}
