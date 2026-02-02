package txutil

import (
	"context"
	cryptorand "crypto/rand"
	"encoding/binary"
	"errors"
	"fmt"
	"math"
	"strings"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
)

// RetryConfig holds configuration for transaction retry logic
type RetryConfig struct {
	MaxRetries      int
	InitialBackoff  time.Duration
	MaxBackoff      time.Duration
	BackoffMultiple float64
}

// DefaultRetryConfig returns a default retry configuration
func DefaultRetryConfig() RetryConfig {
	return RetryConfig{
		MaxRetries:      3,
		InitialBackoff:  time.Second,
		MaxBackoff:      30 * time.Second,
		BackoffMultiple: 2.0,
	}
}

// IsRetryableError checks if an error is retryable
func IsRetryableError(err error) bool {
	if err == nil {
		return false
	}

	errStr := strings.ToLower(err.Error())

	retryableErrors := []string{
		"nonce too low",
		"replacement transaction underpriced",
		"already known",
		"timeout",
		"connection refused",
		"connection reset",
		"broken pipe",
		"i/o timeout",
	}

	for _, retryable := range retryableErrors {
		if strings.Contains(errStr, retryable) {
			return true
		}
	}

	return false
}

// SendTransactionWithRetry sends a transaction with retry logic
func SendTransactionWithRetry(ctx context.Context, client *ethclient.Client, tx *types.Transaction, config RetryConfig) (common.Hash, error) {
	var lastErr error
	backoff := config.InitialBackoff

	for attempt := 0; attempt <= config.MaxRetries; attempt++ {
		if attempt > 0 {
			select {
			case <-ctx.Done():
				return common.Hash{}, ctx.Err()
			case <-time.After(backoff):
			}

			backoff = time.Duration(float64(backoff) * config.BackoffMultiple)
			if backoff > config.MaxBackoff {
				backoff = config.MaxBackoff
			}
		}

		err := client.SendTransaction(ctx, tx)
		if err == nil {
			return tx.Hash(), nil
		}

		lastErr = err
		if !IsRetryableError(err) {
			return common.Hash{}, fmt.Errorf("non-retryable error: %w", err)
		}
	}

	return common.Hash{}, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// WaitForTransactionWithRetry sends a transaction and waits for it to be mined with retry logic
func WaitForTransactionWithRetry(ctx context.Context, client *ethclient.Client, tx *types.Transaction, confirmations uint64, config RetryConfig) (*types.Receipt, error) {
	txHash, err := SendTransactionWithRetry(ctx, client, tx, config)
	if err != nil {
		return nil, fmt.Errorf("failed to send transaction: %w", err)
	}

	receipt, err := WaitForConfirmation(ctx, client, txHash, confirmations)
	if err != nil {
		return nil, fmt.Errorf("failed to wait for confirmation: %w", err)
	}

	return receipt, nil
}

// CalculateBackoff calculates exponential backoff with decorrelated jitter.
// Jitter prevents thundering herd issues when multiple clients retry simultaneously.
// Uses decorrelated jitter: returns backoff/2 + random(0, backoff/2)
func CalculateBackoff(attempt int, initialBackoff, maxBackoff time.Duration, multiplier float64) time.Duration {
	backoff := time.Duration(float64(initialBackoff) * math.Pow(multiplier, float64(attempt)))
	if backoff > maxBackoff {
		backoff = maxBackoff
	}

	// Apply decorrelated jitter to prevent synchronized retry storms
	// Returns backoff/2 + random(0, backoff/2)
	halfBackoff := backoff / 2
	jitter := time.Duration(secureRandomInt64n(int64(halfBackoff) + 1))
	return halfBackoff + jitter
}

// secureRandomInt64n returns a cryptographically secure random int64 in [0, n).
// This is goroutine-safe and suitable for security-sensitive contexts.
func secureRandomInt64n(n int64) int64 {
	if n <= 0 {
		return 0
	}
	var buf [8]byte
	_, err := cryptorand.Read(buf[:])
	if err != nil {
		// Fallback to 0 if crypto/rand fails (extremely rare)
		return 0
	}
	// Use modulo for simplicity - bias is negligible for our use case (jitter)
	return int64(binary.BigEndian.Uint64(buf[:]) % uint64(n))
}

// IsNonceError checks if an error is related to nonce issues
func IsNonceError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "nonce too low") ||
		   strings.Contains(errStr, "nonce too high") ||
		   strings.Contains(errStr, "invalid nonce")
}

// IsGasError checks if an error is related to gas issues
func IsGasError(err error) bool {
	if err == nil {
		return false
	}
	errStr := strings.ToLower(err.Error())
	return strings.Contains(errStr, "gas") ||
		   strings.Contains(errStr, "fee")
}

// WrapError wraps an error with context
func WrapError(operation string, err error) error {
	if err == nil {
		return nil
	}
	return fmt.Errorf("%s: %w", operation, err)
}

// ErrTxFailed is returned when a transaction fails on-chain
var ErrTxFailed = errors.New("transaction failed")

// ErrTxTimeout is returned when waiting for a transaction times out
var ErrTxTimeout = errors.New("transaction timeout")
