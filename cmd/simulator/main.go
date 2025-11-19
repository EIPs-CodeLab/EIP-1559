package main

import (
	"flag"
	"fmt"

	"github.com/EIPs-CodeLab/EIP-1559/internal/basefee"
	"github.com/EIPs-CodeLab/EIP-1559/internal/executor"
	"github.com/EIPs-CodeLab/EIP-1559/internal/types"
	"github.com/EIPs-CodeLab/EIP-1559/internal/validator"
	"github.com/EIPs-CodeLab/EIP-1559/pkg/constants"
)

func main() {
	// Command line flags
	blocks := flag.Int("blocks", 10, "Number of blocks to simulate")
	gasUsed := flag.Uint64("gas", 15000000, "Gas used per block (target is 15M)")
	verbose := flag.Bool("verbose", false, "Verbose output")
	flag.Parse()

	fmt.Println("EIP-1559 Simulator")
	fmt.Println("=====================")
	fmt.Printf("Simulating %d blocks with %d gas used per block\n\n", *blocks, *gasUsed)

	// Initialize state
	state := types.NewState()

	// Create initial accounts
	minerAddr := "0xMiner"
	senderAddr := "0xAlice"
	recipientAddr := "0xBob"

	state.SetAccount(minerAddr, types.NewAccount(minerAddr, 0))
	// Give sender enough balance to cover several transactions (in wei)
	state.SetAccount(senderAddr, types.NewAccount(senderAddr, 1_000_000_000_000_000))
	state.SetAccount(recipientAddr, types.NewAccount(recipientAddr, 0))

	// Create genesis block
	genesisBlock := &types.Block{
		Number:   constants.ForkBlockNumber - 1,
		Hash:     "0xgenesis",
		GasLimit: 30_000_000,
		GasUsed:  15_000_000,
		BaseFee:  constants.InitialBaseFee,
		Miner:    minerAddr,
	}

	currentBlock := genesisBlock
	totalBurned := uint64(0)
	totalTips := uint64(0)

	fmt.Printf("%-6s | %-12s | %-12s | %-8s | %-12s | %-12s\n",
		"Block", "BaseFee", "GasUsed", "Usage%", "Burned", "Tips")
	fmt.Println("-------|--------------|--------------|----------|--------------|------------")

	// Simulate blocks
	for i := 0; i < *blocks; i++ {
		// Calculate next base fee
		nextBaseFee := basefee.Calculate(currentBlock)

		// Create next block
		nextBlock := types.NewBlock(
			currentBlock.Number+1,
			currentBlock.Hash,
			30_000_000,
			nextBaseFee,
			minerAddr,
		)
		nextBlock.Hash = fmt.Sprintf("0xblock%d", nextBlock.Number)

		// Create transaction
		tx := &types.Transaction{
			ChainID:              1,
			Nonce:                state.GetNonce(senderAddr),
			MaxPriorityFeePerGas: 2_000_000_000,               // 2 Gwei tip
			MaxFeePerGas:         nextBaseFee + 5_000_000_000, // base fee + 5 Gwei
			// Use a realistic per-transaction gas limit (transfer ~21k)
			GasLimit: 21_000,
			To:       recipientAddr,
			Value:    1_000,
			From:     senderAddr,
		}

		// Validate transaction
		if err := validator.ValidateTransaction(tx, nextBaseFee, state); err != nil {
			fmt.Printf("Transaction validation failed: %v\n", err)
			continue
		}

		// Add transaction to block
		if err := nextBlock.AddTransaction(tx); err != nil {
			fmt.Printf("Failed to add transaction: %v\n", err)
			continue
		}

		// Validate block
		if err := validator.ValidateBlock(nextBlock, currentBlock); err != nil {
			fmt.Printf("Block validation failed: %v\n", err)
			continue
		}

		// Execute transaction
		result := executor.ExecuteTransaction(tx, nextBlock, state)
		if !result.Success {
			fmt.Printf("Transaction execution failed: %v\n", result.Error)
			continue
		}

		totalBurned += result.BaseFeeAmount
		totalTips += result.TipAmount

		// Print block info
		utilization := nextBlock.Utilization()
		fmt.Printf("%-6d | %-12d | %-12d | %7.2f%% | %-12d | %-12d\n",
			nextBlock.Number,
			nextBlock.BaseFee,
			nextBlock.GasUsed,
			utilization,
			result.BaseFeeAmount,
			result.TipAmount,
		)

		if *verbose {
			fmt.Printf("  Sender balance: %d\n", state.GetBalance(senderAddr))
			fmt.Printf("  Miner balance:  %d\n", state.GetBalance(minerAddr))
			fmt.Printf("  Gas used:       %d / %d\n", result.GasUsed, tx.GasLimit)
			fmt.Println()
		}

		// Move to next block
		currentBlock = nextBlock
	}

	// Summary
	fmt.Println("\nSummary")
	fmt.Println("=======")
	fmt.Printf("Total ETH burned: %d wei\n", totalBurned)
	fmt.Printf("Total tips paid:  %d wei\n", totalTips)
	fmt.Printf("Final base fee:   %d wei (%.2f Gwei)\n",
		currentBlock.BaseFee,
		float64(currentBlock.BaseFee)/1_000_000_000)
	fmt.Printf("\nFinal balances:\n")
	fmt.Printf("  Alice (sender):    %d wei\n", state.GetBalance(senderAddr))
	fmt.Printf("  Bob (recipient):   %d wei\n", state.GetBalance(recipientAddr))
	fmt.Printf("  Miner:             %d wei\n", state.GetBalance(minerAddr))
}
