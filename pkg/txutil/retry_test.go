package txutil

import (
	"errors"
	"testing"
	"time"
)

func TestIsRetryableError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "nonce too low",
			err:      errors.New("nonce too low"),
			expected: true,
		},
		{
			name:     "replacement transaction underpriced",
			err:      errors.New("replacement transaction underpriced"),
			expected: true,
		},
		{
			name:     "already known",
			err:      errors.New("already known"),
			expected: true,
		},
		{
			name:     "timeout error",
			err:      errors.New("timeout occurred"),
			expected: true,
		},
		{
			name:     "connection refused",
			err:      errors.New("connection refused"),
			expected: true,
		},
		{
			name:     "non-retryable error",
			err:      errors.New("insufficient funds"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("IsRetryableError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsNonceError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "nonce too low",
			err:      errors.New("nonce too low"),
			expected: true,
		},
		{
			name:     "nonce too high",
			err:      errors.New("nonce too high"),
			expected: true,
		},
		{
			name:     "invalid nonce",
			err:      errors.New("invalid nonce"),
			expected: true,
		},
		{
			name:     "non-nonce error",
			err:      errors.New("insufficient funds"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsNonceError(tt.err)
			if result != tt.expected {
				t.Errorf("IsNonceError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestIsGasError(t *testing.T) {
	tests := []struct {
		name     string
		err      error
		expected bool
	}{
		{
			name:     "nil error",
			err:      nil,
			expected: false,
		},
		{
			name:     "gas too low",
			err:      errors.New("intrinsic gas too low"),
			expected: true,
		},
		{
			name:     "max fee per gas",
			err:      errors.New("max fee per gas less than block base fee"),
			expected: true,
		},
		{
			name:     "non-gas error",
			err:      errors.New("insufficient funds"),
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := IsGasError(tt.err)
			if result != tt.expected {
				t.Errorf("IsGasError() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestDefaultRetryConfig(t *testing.T) {
	config := DefaultRetryConfig()

	if config.MaxRetries != 3 {
		t.Errorf("MaxRetries = %d, want 3", config.MaxRetries)
	}
	if config.InitialBackoff != time.Second {
		t.Errorf("InitialBackoff = %v, want 1s", config.InitialBackoff)
	}
	if config.MaxBackoff != 30*time.Second {
		t.Errorf("MaxBackoff = %v, want 30s", config.MaxBackoff)
	}
	if config.BackoffMultiple != 2.0 {
		t.Errorf("BackoffMultiple = %f, want 2.0", config.BackoffMultiple)
	}
}

func TestCalculateBackoff(t *testing.T) {
	tests := []struct {
		name           string
		attempt        int
		initialBackoff time.Duration
		maxBackoff     time.Duration
		multiplier     float64
		expected       time.Duration
	}{
		{
			name:           "first attempt",
			attempt:        0,
			initialBackoff: time.Second,
			maxBackoff:     30 * time.Second,
			multiplier:     2.0,
			expected:       time.Second,
		},
		{
			name:           "second attempt",
			attempt:        1,
			initialBackoff: time.Second,
			maxBackoff:     30 * time.Second,
			multiplier:     2.0,
			expected:       2 * time.Second,
		},
		{
			name:           "third attempt",
			attempt:        2,
			initialBackoff: time.Second,
			maxBackoff:     30 * time.Second,
			multiplier:     2.0,
			expected:       4 * time.Second,
		},
		{
			name:           "exceeds max backoff",
			attempt:        10,
			initialBackoff: time.Second,
			maxBackoff:     30 * time.Second,
			multiplier:     2.0,
			expected:       30 * time.Second,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := CalculateBackoff(tt.attempt, tt.initialBackoff, tt.maxBackoff, tt.multiplier)
			if result != tt.expected {
				t.Errorf("CalculateBackoff() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestWrapError(t *testing.T) {
	tests := []struct {
		name      string
		operation string
		err       error
		expected  string
	}{
		{
			name:      "nil error",
			operation: "test",
			err:       nil,
			expected:  "",
		},
		{
			name:      "wrap error",
			operation: "create transaction",
			err:       errors.New("insufficient funds"),
			expected:  "create transaction: insufficient funds",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := WrapError(tt.operation, tt.err)
			if result == nil && tt.expected != "" {
				t.Errorf("WrapError() = nil, want error with message %q", tt.expected)
			}
			if result != nil && result.Error() != tt.expected {
				t.Errorf("WrapError() = %q, want %q", result.Error(), tt.expected)
			}
		})
	}
}
