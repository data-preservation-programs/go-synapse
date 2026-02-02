package txutil

import (
	"context"
	"errors"
	"fmt"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Error types for receipt waiting
var (
	// ErrReceiptTimeout is returned when waiting for a receipt times out
	ErrReceiptTimeout = errors.New("timeout waiting for transaction receipt")
	// ErrReceiptRPCFailure is returned when too many consecutive RPC errors occur
	ErrReceiptRPCFailure = errors.New("receipt fetch failed due to repeated RPC errors")
)

// ReceiptWaitConfig configures the WaitForReceipt behavior
type ReceiptWaitConfig struct {
	Timeout             time.Duration // Total timeout for waiting (default: 5 minutes)
	PollInterval        time.Duration // How often to poll (default: 1 second)
	MaxConsecutiveErrors int          // Max consecutive RPC errors before failing (default: 5)
}

// DefaultReceiptWaitConfig returns the default configuration
func DefaultReceiptWaitConfig() ReceiptWaitConfig {
	return ReceiptWaitConfig{
		Timeout:             5 * time.Minute,
		PollInterval:        time.Second,
		MaxConsecutiveErrors: 5,
	}
}

// WaitForConfirmation waits for a transaction to be confirmed with the specified number of confirmations
func WaitForConfirmation(ctx context.Context, client *ethclient.Client, txHash common.Hash, confirmations uint64) (*types.Receipt, error) {
	ticker := time.NewTicker(2 * time.Second)
	defer ticker.Stop()

	consecutiveErrors := 0
	pollCount := 0
	var lastErr error

	for {
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-ticker.C:
			pollCount++
			receipt, err := client.TransactionReceipt(ctx, txHash)
			if err != nil {
				// Distinguish between "not found yet" and actual RPC errors
				if errors.Is(err, ethereum.NotFound) {
					// Transaction not mined yet - this is expected, reset error counter
					consecutiveErrors = 0
					continue
				}
				// Actual RPC error
				consecutiveErrors++
				lastErr = err
				if consecutiveErrors >= 5 {
					return nil, fmt.Errorf("%w: %d consecutive errors, last error: %v", ErrReceiptRPCFailure, consecutiveErrors, lastErr)
				}
				continue
			}

			consecutiveErrors = 0

			if receipt.Status != types.ReceiptStatusSuccessful {
				return receipt, fmt.Errorf("transaction failed with status %d", receipt.Status)
			}

			if confirmations == 0 {
				return receipt, nil
			}

			currentBlock, err := client.BlockNumber(ctx)
			if err != nil {
				consecutiveErrors++
				lastErr = err
				if consecutiveErrors >= 5 {
					return nil, fmt.Errorf("%w: %d consecutive errors, last error: %v", ErrReceiptRPCFailure, consecutiveErrors, lastErr)
				}
				continue
			}

			consecutiveErrors = 0

			if receipt.BlockNumber.Uint64()+confirmations <= currentBlock {
				return receipt, nil
			}
		}
	}
}

// WaitForReceipt waits for a transaction receipt without confirmation requirements.
// Uses a default timeout of 5 minutes. For custom configuration, use WaitForReceiptWithConfig.
func WaitForReceipt(ctx context.Context, client *ethclient.Client, txHash common.Hash, timeout time.Duration) (*types.Receipt, error) {
	config := DefaultReceiptWaitConfig()
	if timeout > 0 {
		config.Timeout = timeout
	}
	return WaitForReceiptWithConfig(ctx, client, txHash, config)
}

// WaitForReceiptWithConfig waits for a transaction receipt with custom configuration
func WaitForReceiptWithConfig(ctx context.Context, client *ethclient.Client, txHash common.Hash, config ReceiptWaitConfig) (*types.Receipt, error) {
	ctx, cancel := context.WithTimeout(ctx, config.Timeout)
	defer cancel()

	pollInterval := config.PollInterval
	if pollInterval == 0 {
		pollInterval = time.Second
	}
	maxErrors := config.MaxConsecutiveErrors
	if maxErrors == 0 {
		maxErrors = 5
	}

	ticker := time.NewTicker(pollInterval)
	defer ticker.Stop()

	consecutiveErrors := 0
	pollCount := 0
	var lastErr error

	for {
		select {
		case <-ctx.Done():
			return nil, fmt.Errorf("%w after %d polls: %v", ErrReceiptTimeout, pollCount, ctx.Err())
		case <-ticker.C:
			pollCount++
			receipt, err := client.TransactionReceipt(ctx, txHash)
			if err != nil {
				// Distinguish between "not found yet" and actual RPC errors
				if errors.Is(err, ethereum.NotFound) {
					// Transaction not mined yet - this is expected, reset error counter
					consecutiveErrors = 0
					continue
				}
				// Actual RPC error
				consecutiveErrors++
				lastErr = err
				if consecutiveErrors >= maxErrors {
					return nil, fmt.Errorf("%w: %d consecutive errors after %d polls, last error: %v", ErrReceiptRPCFailure, consecutiveErrors, pollCount, lastErr)
				}
				continue
			}

			if receipt.Status != types.ReceiptStatusSuccessful {
				return receipt, fmt.Errorf("transaction failed with status %d", receipt.Status)
			}
			return receipt, nil
		}
	}
}
