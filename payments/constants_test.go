package payments

import (
	"math/big"
	"testing"
	"time"
)

func TestCurrentEpoch(t *testing.T) {
	epoch := CurrentEpoch(314)
	if epoch.Cmp(big.NewInt(0)) <= 0 {
		t.Error("Expected positive epoch for mainnet")
	}

	epoch = CurrentEpoch(314159)
	if epoch.Cmp(big.NewInt(0)) <= 0 {
		t.Error("Expected positive epoch for calibration")
	}

	epoch = CurrentEpoch(999999)
	if epoch.Cmp(big.NewInt(0)) != 0 {
		t.Error("Expected zero epoch for unknown chain")
	}
}

func TestEpochToTime(t *testing.T) {
	t.Run("should convert epoch 0 to genesis timestamp for mainnet", func(t *testing.T) {
		genesis := EpochToTime(314, big.NewInt(0))
		expectedGenesis := time.Unix(GenesisTimestamps[314], 0)
		if !genesis.Equal(expectedGenesis) {
			t.Errorf("Expected genesis %v, got %v", expectedGenesis, genesis)
		}
	})

	t.Run("should convert epoch 0 to genesis timestamp for calibration", func(t *testing.T) {
		genesis := EpochToTime(314159, big.NewInt(0))
		expectedGenesis := time.Unix(GenesisTimestamps[314159], 0)
		if !genesis.Equal(expectedGenesis) {
			t.Errorf("Expected genesis %v, got %v", expectedGenesis, genesis)
		}
	})

	t.Run("should calculate correct date for future epochs", func(t *testing.T) {
		epochsPerDay := int64(2880) // 24 * 60 * 2
		date := EpochToTime(314, big.NewInt(epochsPerDay))
		expectedTime := time.Unix(GenesisTimestamps[314]+epochsPerDay*30, 0)
		if !date.Equal(expectedTime) {
			t.Errorf("Expected %v, got %v", expectedTime, date)
		}
	})

	t.Run("should handle large epoch numbers", func(t *testing.T) {
		largeEpoch := int64(1000000)
		date := EpochToTime(314159, big.NewInt(largeEpoch))
		expectedTime := time.Unix(GenesisTimestamps[314159]+largeEpoch*30, 0)
		if !date.Equal(expectedTime) {
			t.Errorf("Expected %v, got %v", expectedTime, date)
		}
	})

	t.Run("should return zero time for unknown chain", func(t *testing.T) {
		unknown := EpochToTime(999999, big.NewInt(0))
		if !unknown.IsZero() {
			t.Error("Expected zero time for unknown chain")
		}
	})
}

func TestTimeToEpoch(t *testing.T) {
	t.Run("should convert genesis date to epoch 0 for mainnet", func(t *testing.T) {
		genesisDate := time.Unix(GenesisTimestamps[314], 0)
		epoch := TimeToEpoch(314, genesisDate)
		if epoch.Cmp(big.NewInt(0)) != 0 {
			t.Errorf("Expected epoch 0, got %v", epoch)
		}
	})

	t.Run("should convert genesis date to epoch 0 for calibration", func(t *testing.T) {
		genesisDate := time.Unix(GenesisTimestamps[314159], 0)
		epoch := TimeToEpoch(314159, genesisDate)
		if epoch.Cmp(big.NewInt(0)) != 0 {
			t.Errorf("Expected epoch 0, got %v", epoch)
		}
	})

	t.Run("should calculate correct epoch for future dates", func(t *testing.T) {
		futureDate := time.Unix(GenesisTimestamps[314]+3600, 0)
		epoch := TimeToEpoch(314, futureDate)
		if epoch.Cmp(big.NewInt(120)) != 0 {
			t.Errorf("Expected epoch 120, got %v", epoch)
		}
	})

	t.Run("should round down to nearest epoch", func(t *testing.T) {
		partialEpochDate := time.Unix(GenesisTimestamps[314159]+45, 0)
		epoch := TimeToEpoch(314159, partialEpochDate)
		if epoch.Cmp(big.NewInt(1)) != 0 {
			t.Errorf("Expected epoch 1, got %v", epoch)
		}
	})

	t.Run("should return zero for unknown chain", func(t *testing.T) {
		epoch := TimeToEpoch(999999, time.Now())
		if epoch.Cmp(big.NewInt(0)) != 0 {
			t.Errorf("Expected epoch 0 for unknown chain, got %v", epoch)
		}
	})
}

func TestEpochRoundTrip(t *testing.T) {
	chainID := int64(314)
	originalEpoch := big.NewInt(100000)

	epochTime := EpochToTime(chainID, originalEpoch)
	recoveredEpoch := TimeToEpoch(chainID, epochTime)

	if originalEpoch.Cmp(recoveredEpoch) != 0 {
		t.Errorf("Round-trip failed: original %v, recovered %v", originalEpoch, recoveredEpoch)
	}
}

type TimeUntilResult struct {
	Epochs  int64
	Seconds int64
	Minutes float64
	Hours   float64
	Days    float64
}

func TimeUntilEpoch(futureEpoch, currentEpoch int64) TimeUntilResult {
	epochDiff := futureEpoch - currentEpoch
	seconds := epochDiff * 30

	return TimeUntilResult{
		Epochs:  epochDiff,
		Seconds: seconds,
		Minutes: float64(seconds) / 60,
		Hours:   float64(seconds) / 3600,
		Days:    float64(seconds) / 86400,
	}
}

func TestTimeUntilEpoch(t *testing.T) {
	t.Run("should calculate correct time difference", func(t *testing.T) {
		currentEpoch := int64(1000)
		futureEpoch := int64(1120) // 120 epochs in the future = 1 hour
		result := TimeUntilEpoch(futureEpoch, currentEpoch)

		if result.Epochs != 120 {
			t.Errorf("Expected 120 epochs, got %d", result.Epochs)
		}
		if result.Seconds != 3600 {
			t.Errorf("Expected 3600 seconds, got %d", result.Seconds)
		}
		if result.Minutes != 60 {
			t.Errorf("Expected 60 minutes, got %f", result.Minutes)
		}
		if result.Hours != 1 {
			t.Errorf("Expected 1 hour, got %f", result.Hours)
		}
	})

	t.Run("should handle same epoch", func(t *testing.T) {
		result := TimeUntilEpoch(1000, 1000)

		if result.Epochs != 0 {
			t.Errorf("Expected 0 epochs, got %d", result.Epochs)
		}
		if result.Seconds != 0 {
			t.Errorf("Expected 0 seconds, got %d", result.Seconds)
		}
	})

	t.Run("should handle negative differences (past epochs)", func(t *testing.T) {
		result := TimeUntilEpoch(1000, 1120)

		if result.Epochs != -120 {
			t.Errorf("Expected -120 epochs, got %d", result.Epochs)
		}
		if result.Seconds != -3600 {
			t.Errorf("Expected -3600 seconds, got %d", result.Seconds)
		}
		if result.Hours != -1 {
			t.Errorf("Expected -1 hour, got %f", result.Hours)
		}
	})
}

func TestGenesisTimestamps(t *testing.T) {
	t.Run("should have correct timestamp for mainnet", func(t *testing.T) {
		if GenesisTimestamps[314] != 1598306400 {
			t.Errorf("Wrong mainnet genesis: expected 1598306400, got %d", GenesisTimestamps[314])
		}
	})

	t.Run("should have correct timestamp for calibration", func(t *testing.T) {
		if GenesisTimestamps[314159] != 1667326380 {
			t.Errorf("Wrong calibration genesis: expected 1667326380, got %d", GenesisTimestamps[314159])
		}
	})
}

func TestTokenConstants(t *testing.T) {
	t.Run("should have USDFC address for mainnet", func(t *testing.T) {
		addr, ok := USDFCAddresses[314]
		if !ok {
			t.Error("Missing USDFC address for mainnet")
		}
		expected := "0x80B98d3aa09ffff255c3ba4A241111Ff1262F045"
		if addr.Hex() != expected {
			t.Errorf("Wrong mainnet USDFC address: expected %s, got %s", expected, addr.Hex())
		}
	})

	t.Run("should have USDFC address for calibration", func(t *testing.T) {
		addr, ok := USDFCAddresses[314159]
		if !ok {
			t.Error("Missing USDFC address for calibration")
		}
		expected := "0xb3042734b608a1B16e9e86B374A3f3e389B4cDf0"
		if addr.Hex() != expected {
			t.Errorf("Wrong calibration USDFC address: expected %s, got %s", expected, addr.Hex())
		}
	})

	t.Run("should have correct settlement fee value", func(t *testing.T) {
		expected := big.NewInt(1300000000000000)
		if SettlementFee.Cmp(expected) != 0 {
			t.Errorf("Wrong settlement fee: expected %s, got %s", expected.String(), SettlementFee.String())
		}
	})

	t.Run("should have correct epochs per day", func(t *testing.T) {
		if EpochsPerDay != 2880 {
			t.Errorf("Wrong epochs per day: expected 2880, got %d", EpochsPerDay)
		}
	})

	t.Run("should have correct epochs per month", func(t *testing.T) {
		if EpochsPerMonth != 86400 {
			t.Errorf("Wrong epochs per month: expected 86400, got %d", EpochsPerMonth)
		}
	})
}
