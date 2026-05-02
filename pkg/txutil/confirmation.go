package txutil

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

var (
	ErrReceiptTimeout    = errors.New("timeout waiting for transaction receipt")
	ErrReceiptRPCFailure = errors.New("receipt fetch failed due to repeated RPC errors")
)

type ReceiptWaitConfig struct {
	Timeout              time.Duration
	PollInterval         time.Duration
	MaxConsecutiveErrors int
}

func DefaultReceiptWaitConfig() ReceiptWaitConfig {
	return ReceiptWaitConfig{
		Timeout:              5 * time.Minute,
		PollInterval:         time.Second,
		MaxConsecutiveErrors: 5,
	}
}

// WaitForReceipt polls until the receipt for txHash is available or timeout
// elapses. Default timeout is 5 minutes when timeout is zero.
func WaitForReceipt(ctx context.Context, client *ethclient.Client, txHash common.Hash, timeout time.Duration) (*types.Receipt, error) {
	config := DefaultReceiptWaitConfig()
	if timeout > 0 {
		config.Timeout = timeout
	}
	return WaitForReceiptWithConfig(ctx, client, txHash, config)
}

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
			if lastErr != nil {
				return nil, fmt.Errorf("%w after %d polls: %v (last error: %v)", ErrReceiptTimeout, pollCount, ctx.Err(), lastErr)
			}
			return nil, fmt.Errorf("%w after %d polls: %v", ErrReceiptTimeout, pollCount, ctx.Err())
		case <-ticker.C:
			pollCount++
			receipt, err := client.TransactionReceipt(ctx, txHash)
			if err != nil {
				if errors.Is(err, ethereum.NotFound) {
					// not mined yet -- expected, reset error counter
					consecutiveErrors = 0
					continue
				}
				if !isRetryableError(err) {
					return nil, fmt.Errorf("%w: non-retryable error: %v", ErrReceiptRPCFailure, err)
				}
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

// isRetryableError returns true for transient RPC errors worth retrying.
// Matches by string fragment because go-ethereum surfaces these as plain errors.
func isRetryableError(err error) bool {
	if err == nil {
		return false
	}
	if errors.Is(err, context.Canceled) || errors.Is(err, context.DeadlineExceeded) {
		return false
	}

	errStr := strings.ToLower(err.Error())
	for _, retryable := range []string{
		"nonce too low",
		"replacement transaction underpriced",
		"already known",
		"timeout",
		"connection refused",
		"connection reset",
		"broken pipe",
		"i/o timeout",
	} {
		if strings.Contains(errStr, retryable) {
			return true
		}
	}
	return false
}
