package txutil

import (
	"context"
	"errors"
	"testing"
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
		{
			name:     "context deadline exceeded",
			err:      context.DeadlineExceeded,
			expected: false,
		},
		{
			name:     "context canceled",
			err:      context.Canceled,
			expected: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := isRetryableError(tt.err)
			if result != tt.expected {
				t.Errorf("isRetryableError() = %v, want %v", result, tt.expected)
			}
		})
	}
}
