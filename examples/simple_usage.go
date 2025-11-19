package main

import (
	"fmt"

	"github.com/EIPs-CodeLab/EIP-1559/internal/basefee"
	"github.com/EIPs-CodeLab/EIP-1559/internal/types"
	"github.com/EIPs-CodeLab/EIP-1559/pkg/constants"
)

func main() {
	// Create a parent block (e.g. the most recent block)
	parent := &types.Block{
		Number:   constants.ForkBlockNumber,
		GasLimit: 30_000_000,
		GasUsed:  20_000_000,
		BaseFee:  constants.InitialBaseFee,
	}

	// Calculate the next block's base fee
	nextBaseFee := basefee.Calculate(parent)
	fmt.Printf("Parent base fee: %d\n", parent.BaseFee)
	fmt.Printf("Next base fee:   %d\n", nextBaseFee)

	// Simulate a short sequence of blocks with high usage
	highUsage := []uint64{25_000_000, 28_000_000, 29_000_000}
	fees := basefee.CalculateForBlocks(parent, highUsage)
	fmt.Println("Simulated base fees for high usage sequence:")
	for i, f := range fees {
		fmt.Printf("  block %d -> %d\n", parent.Number+uint64(i)+1, f)
	}

	// Simulate low usage sequence
	lowUsage := []uint64{5_000_000, 3_000_000, 1_000_000}
	// start from last fee
	parent.BaseFee = fees[len(fees)-1]
	fees2 := basefee.CalculateForBlocks(parent, lowUsage)
	fmt.Println("Simulated base fees for low usage sequence:")
	for i, f := range fees2 {
		fmt.Printf("  block %d -> %d\n", parent.Number+uint64(i)+1, f)
	}
}
