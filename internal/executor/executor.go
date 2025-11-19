package executor

import (
	"fmt"

	"github.com/EIPs-CodeLab/EIP-1559/internal/types"
)

// ExecutionResult holds the result of transaction execution
type ExecutionResult struct {
	GasUsed       uint64
	BaseFeeAmount uint64 // Amount burned
	TipAmount     uint64 // Amount paid to miner
	Success       bool
	Error         error
}

func ExecuteTransaction(tx *types.Transaction, block *types.Block, state *types.State) *ExecutionResult {
	result := &ExecutionResult{
		Success: false,
	}

	// Get accounts
	sender := state.GetAccount(tx.From)
	miner := state.GetAccount(block.Miner)

	// Calculate fees
	effectiveGasPrice := tx.EffectiveGasPrice(block.BaseFee)
	priorityFee := tx.EffectivePriorityFee(block.BaseFee)

	// Deduct upfront cost (gas + value)
	// Deduct upfront cost (gas + value) based on MAX fee
	upfrontGasCost := tx.GasLimit * tx.MaxFeePerGas
	totalCost := upfrontGasCost + tx.Value

	if err := sender.Deduct(totalCost); err != nil {
		result.Error = fmt.Errorf("insufficient funds for gas + value: %w", err)
		return result
	}

	// Execute transaction (simplified - actual execution would call EVM)
	gasUsed := executeTransaction(tx)
	result.GasUsed = gasUsed

	// Calculate actual costs
	// Calculate actual costs
	// Refund = (GasLimit * MaxFee) - (GasUsed * EffectiveFee)
	//        = (GasLimit - GasUsed) * MaxFee + GasUsed * (MaxFee - EffectiveFee)
	// Simplified: Refund unused gas @ MaxFee + Refund overpayment on used gas

	remainderGas := tx.GasLimit - gasUsed
	refundAmount := remainderGas * tx.MaxFeePerGas

	// Add refund for the difference between max fee and effective fee for used gas
	overpaymentPerGas := tx.MaxFeePerGas - effectiveGasPrice
	refundAmount += gasUsed * overpaymentPerGas

	// Refund unused gas to sender
	sender.Add(refundAmount)

	// Transfer value to recipient (if not contract creation)
	if tx.To != "" {
		recipient := state.GetAccount(tx.To)
		recipient.Add(tx.Value)
	}

	// Pay miner the priority fee (tip)
	tipAmount := gasUsed * priorityFee
	miner.Add(tipAmount)
	result.TipAmount = tipAmount

	// Base fee is BURNED (not given to anyone)
	baseFeeAmount := gasUsed * block.BaseFee
	result.BaseFeeAmount = baseFeeAmount
	// Note: baseFeeAmount is effectively burned as it's not added to any account

	// Increment sender nonce
	sender.IncrementNonce()

	result.Success = true
	return result
}

// executeTransaction simulates transaction execution
// In a real implementation, this would call the EVM
func executeTransaction(tx *types.Transaction) uint64 {
	// Simple simulation: use 21000 gas for transfer, more for contract calls
	baseGas := uint64(21000)

	if len(tx.Data) > 0 {
		// Contract call/creation uses more gas
		dataGas := uint64(len(tx.Data)) * 16 // 16 gas per byte
		return baseGas + dataGas
	}

	return baseGas
}

func ExecuteBlock(block *types.Block, state *types.State) ([]*ExecutionResult, error) {
	results := make([]*ExecutionResult, 0, len(block.Transactions))

	for _, tx := range block.Transactions {
		result := ExecuteTransaction(tx, block, state)
		results = append(results, result)

		if !result.Success {
			return results, fmt.Errorf("transaction execution failed: %v", result.Error)
		}
	}

	return results, nil
}
