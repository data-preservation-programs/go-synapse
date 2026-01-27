package txutil

import (
	"testing"

	"github.com/ethereum/go-ethereum"
)

func TestEstimateGasWithBuffer_InvalidPercent(t *testing.T) {
	tests := []struct {
		name          string
		bufferPercent int
		shouldError   bool
	}{
		{
			name:          "negative percent",
			bufferPercent: -1,
			shouldError:   true,
		},
		{
			name:          "too large percent",
			bufferPercent: 101,
			shouldError:   true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := EstimateGasWithBuffer(nil, nil, ethereum.CallMsg{}, tt.bufferPercent)
			if tt.shouldError && err == nil {
				t.Error("EstimateGasWithBuffer() should return error for invalid buffer percent")
			}
			if !tt.shouldError && err != nil && err.Error() == "buffer percent must be between 0 and 100" {
				t.Errorf("EstimateGasWithBuffer() should not return percent error, got: %v", err)
			}
		})
	}
}
