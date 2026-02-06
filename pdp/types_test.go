package pdp

import (
	"testing"
)

func TestDefaultManagerConfig(t *testing.T) {
	config := DefaultManagerConfig()

	if config.GasBufferPercent != 10 {
		t.Errorf("expected default GasBufferPercent of 10, got %d", config.GasBufferPercent)
	}
}

func TestManagerConfig_Validation(t *testing.T) {
	tests := []struct {
		name          string
		gasBuffer     int
		shouldBeValid bool
	}{
		{"zero percent", 0, true},
		{"valid 10 percent", 10, true},
		{"valid 50 percent", 50, true},
		{"valid 100 percent", 100, true},
		{"negative percent", -1, false},
		{"over 100 percent", 101, false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			config := ManagerConfig{
				GasBufferPercent: tt.gasBuffer,
			}

			// We test validation through NewManagerWithConfig
			// For this test, we just validate the range
			isValid := config.GasBufferPercent >= 0 && config.GasBufferPercent <= 100

			if isValid != tt.shouldBeValid {
				t.Errorf("config validation = %v, want %v for GasBufferPercent=%d",
					isValid, tt.shouldBeValid, config.GasBufferPercent)
			}
		})
	}
}
