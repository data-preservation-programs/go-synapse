package txutil

import (
	"context"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// WaitForConfirmation waits for a transaction to be confirmed with the specified number of confirmations
func WaitForConfirmation(ctx context.Context, client *ethclient.Client, txHash common.Hash, confirmations uint64) (*types.Receipt, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			receipt, err := client.TransactionReceipt(ctx, txHash)
			if err != nil {
				continue
			}

			if receipt.Status != types.ReceiptStatusSuccessful {
				return receipt, fmt.Errorf("transaction failed with status %d", receipt.Status)
			}

			if confirmations == 0 {
				return receipt, nil
			}

			currentBlock, err := client.BlockNumber(ctx)
			if err != nil {
				continue
			}

			if receipt.BlockNumber.Uint64()+confirmations <= currentBlock {
				return receipt, nil
			}
		}
	}
}

// WaitForReceipt waits for a transaction receipt without confirmation requirements
func WaitForReceipt(ctx context.Context, client *ethclient.Client, txHash common.Hash, timeout time.Duration) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("timeout waiting for transaction receipt: %w", ctx.Err())
		case <-ticker.C:
			receipt, err := client.TransactionReceipt(ctx, txHash)
			if err == nil {
				if receipt.Status != types.ReceiptStatusSuccessful {
					return receipt, fmt.Errorf("transaction failed with status %d", receipt.Status)
				}
				return receipt, nil
			}
		}
	}
}
