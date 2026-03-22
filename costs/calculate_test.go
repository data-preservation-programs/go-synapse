package costs

import (
	"math/big"
	"testing"

	"github.com/data-preservation-programs/go-synapse/constants"
	"github.com/data-preservation-programs/go-synapse/warmstorage"
)

// Helper to create a *big.Int from int64.
func bi(v int64) *big.Int { return big.NewInt(v) }

// usdfc returns n USDFC as attoUSDFC (n * 1e18).
func usdfc(n int64) *big.Int {
	return new(big.Int).Mul(bi(n), big.NewInt(1e18))
}

// usdfcFrac returns a fractional USDFC amount: numerator/10 USDFC.
// e.g. usdfcFrac(1) = 0.1 USDFC, usdfcFrac(25) = 2.5 USDFC.
func usdfcFrac(tenths int64) *big.Int {
	return new(big.Int).Mul(bi(tenths), big.NewInt(1e17))
}

func defaultPricing() *warmstorage.ServicePrice {
	return &warmstorage.ServicePrice{
		PricePerTiBPerMonthNoCDN: usdfcFrac(25), // 2.5 USDFC/TiB/month
		MinimumPricePerMonth:     usdfcFrac(1),   // 0.1 USDFC/month
		EpochsPerMonth:           bi(constants.EpochsPerMonth),
	}
}

// --- CalculateEffectiveRate ---

func TestCalculateEffectiveRate_ExactOneTiB(t *testing.T) {
	pricing := defaultPricing()
	size := bi(constants.TiB)

	rate := CalculateEffectiveRate(
		size,
		pricing.PricePerTiBPerMonthNoCDN,
		pricing.MinimumPricePerMonth,
		pricing.EpochsPerMonth.Int64(),
	)

	// ratePerMonth = 2.5 USDFC * 1 TiB / 1 TiB = 2.5 USDFC
	if rate.RatePerMonth.Cmp(usdfcFrac(25)) != 0 {
		t.Errorf("ratePerMonth: got %s, want %s", rate.RatePerMonth, usdfcFrac(25))
	}

	// ratePerEpoch = 2.5 USDFC / 86400 epochs
	expectedPerEpoch := new(big.Int).Div(usdfcFrac(25), bi(constants.EpochsPerMonth))
	if rate.RatePerEpoch.Cmp(expectedPerEpoch) != 0 {
		t.Errorf("ratePerEpoch: got %s, want %s", rate.RatePerEpoch, expectedPerEpoch)
	}
}

func TestCalculateEffectiveRate_SubTiB_HitsMinimum(t *testing.T) {
	pricing := defaultPricing()
	// Very small size: 1 byte. Natural rate << minimum.
	size := bi(1)

	rate := CalculateEffectiveRate(
		size,
		pricing.PricePerTiBPerMonthNoCDN,
		pricing.MinimumPricePerMonth,
		pricing.EpochsPerMonth.Int64(),
	)

	// Should hit minimum monthly rate
	if rate.RatePerMonth.Cmp(pricing.MinimumPricePerMonth) != 0 {
		t.Errorf("ratePerMonth should be minimum: got %s, want %s",
			rate.RatePerMonth, pricing.MinimumPricePerMonth)
	}

	// ratePerEpoch should be at least 1
	if rate.RatePerEpoch.Cmp(bi(1)) < 0 {
		t.Errorf("ratePerEpoch should be at least 1: got %s", rate.RatePerEpoch)
	}
}

func TestCalculateEffectiveRate_MultiTiB(t *testing.T) {
	pricing := defaultPricing()
	size := new(big.Int).Mul(bi(5), bi(constants.TiB)) // 5 TiB

	rate := CalculateEffectiveRate(
		size,
		pricing.PricePerTiBPerMonthNoCDN,
		pricing.MinimumPricePerMonth,
		pricing.EpochsPerMonth.Int64(),
	)

	// ratePerMonth = 2.5 * 5 = 12.5 USDFC
	expected := usdfcFrac(125) // 12.5 USDFC
	if rate.RatePerMonth.Cmp(expected) != 0 {
		t.Errorf("ratePerMonth: got %s, want %s", rate.RatePerMonth, expected)
	}
}

func TestCalculateEffectiveRate_ZeroSize(t *testing.T) {
	pricing := defaultPricing()
	rate := CalculateEffectiveRate(
		bi(0),
		pricing.PricePerTiBPerMonthNoCDN,
		pricing.MinimumPricePerMonth,
		pricing.EpochsPerMonth.Int64(),
	)

	// Should hit minimum
	if rate.RatePerMonth.Cmp(pricing.MinimumPricePerMonth) != 0 {
		t.Errorf("ratePerMonth should be minimum for zero size")
	}
}

// --- CalculateAdditionalLockupRequired ---

func TestAdditionalLockup_NewDataSet(t *testing.T) {
	pricing := defaultPricing()
	sybilFee := usdfcFrac(1) // 0.1 USDFC

	lockup := CalculateAdditionalLockupRequired(
		bi(constants.TiB), // uploading 1 TiB
		bi(0),             // empty dataset
		pricing,
		DefaultLockupPeriod,
		sybilFee,
		true,  // new dataset
		false, // no CDN
	)

	if lockup.RateDelta.Sign() <= 0 {
		t.Errorf("rateDelta should be positive for new dataset: got %s", lockup.RateDelta)
	}

	// TotalLockup should include sybil fee
	minExpected := new(big.Int).Add(
		new(big.Int).Mul(lockup.RateDelta, bi(DefaultLockupPeriod)),
		sybilFee,
	)
	if lockup.TotalLockup.Cmp(minExpected) != 0 {
		t.Errorf("totalLockup: got %s, want %s", lockup.TotalLockup, minExpected)
	}
}

func TestAdditionalLockup_NewDataSet_WithCDN(t *testing.T) {
	pricing := defaultPricing()
	sybilFee := usdfcFrac(1)

	lockup := CalculateAdditionalLockupRequired(
		bi(constants.TiB),
		bi(0),
		pricing,
		DefaultLockupPeriod,
		sybilFee,
		true, // new dataset
		true, // CDN enabled
	)

	// Should include CDN fixed lockup + sybil fee
	rateLockup := new(big.Int).Mul(lockup.RateDelta, bi(DefaultLockupPeriod))
	expected := new(big.Int).Add(rateLockup, CDNFixedLockup)
	expected.Add(expected, sybilFee)

	if lockup.TotalLockup.Cmp(expected) != 0 {
		t.Errorf("totalLockup with CDN: got %s, want %s", lockup.TotalLockup, expected)
	}

	// verify component breakdown
	if lockup.RateLockup.Cmp(rateLockup) != 0 {
		t.Errorf("RateLockup: got %s, want %s", lockup.RateLockup, rateLockup)
	}
	if lockup.CDNFixedLockup.Cmp(CDNFixedLockup) != 0 {
		t.Errorf("CDNFixedLockup: got %s, want %s", lockup.CDNFixedLockup, CDNFixedLockup)
	}
	if lockup.SybilFee.Cmp(sybilFee) != 0 {
		t.Errorf("SybilFee: got %s, want %s", lockup.SybilFee, sybilFee)
	}
}

func TestAdditionalLockup_ExistingDataSet(t *testing.T) {
	pricing := defaultPricing()
	currentSize := bi(constants.TiB) // 1 TiB already stored

	lockup := CalculateAdditionalLockupRequired(
		bi(constants.TiB), // adding 1 TiB
		currentSize,
		pricing,
		DefaultLockupPeriod,
		usdfcFrac(1),
		false, // existing dataset
		false,
	)

	// rateDelta = rate(2TiB) - rate(1TiB), should be positive
	if lockup.RateDelta.Sign() < 0 {
		t.Errorf("rateDelta should not be negative: got %s", lockup.RateDelta)
	}

	// Should NOT include sybil fee or CDN lockup for existing dataset
	expectedLockup := new(big.Int).Mul(lockup.RateDelta, bi(DefaultLockupPeriod))
	if lockup.TotalLockup.Cmp(expectedLockup) != 0 {
		t.Errorf("totalLockup for existing dataset should not include sybil: got %s, want %s",
			lockup.TotalLockup, expectedLockup)
	}
}

func TestAdditionalLockup_ExistingDataSet_NilSybilFee(t *testing.T) {
	pricing := defaultPricing()

	lockup := CalculateAdditionalLockupRequired(
		bi(constants.TiB),
		bi(0),
		pricing,
		DefaultLockupPeriod,
		nil,   // nil sybil fee
		true,  // new dataset
		false,
	)

	// Should not panic and should equal rateDelta * lockupPeriod
	expectedLockup := new(big.Int).Mul(lockup.RateDelta, bi(DefaultLockupPeriod))
	if lockup.TotalLockup.Cmp(expectedLockup) != 0 {
		t.Errorf("totalLockup with nil sybil: got %s, want %s",
			lockup.TotalLockup, expectedLockup)
	}
}

// --- CalculateDepositNeeded ---

func TestDepositNeeded_InsufficientFunds(t *testing.T) {
	deposit := CalculateDepositNeeded(
		usdfc(10),  // additionalLockup
		bi(100),    // rateDelta per epoch
		bi(50),     // currentLockupRate per epoch
		bi(0),      // no debt
		usdfc(1),   // 1 USDFC available
		DefaultRunwayEpochs,
		DefaultBufferEpochs,
		false, // existing dataset
	)

	if deposit.Sign() <= 0 {
		t.Errorf("deposit should be positive when funds are insufficient: got %s", deposit)
	}
}

func TestDepositNeeded_SufficientFunds(t *testing.T) {
	// give a massive amount of available funds
	huge := new(big.Int).Mul(usdfc(1000000), bi(1e18))
	deposit := CalculateDepositNeeded(
		usdfc(1),  // small lockup
		bi(1),     // tiny rate delta
		bi(1),     // tiny current rate
		bi(0),     // no debt
		huge,      // way more than enough
		DefaultRunwayEpochs,
		DefaultBufferEpochs,
		false,
	)

	// raw = lockup + runway - available + debt → negative → clamped to 0
	// buffer = (1+1) * 5 = 10
	// deposit = 0 + 10 = 10
	expectedBuffer := new(big.Int).Mul(bi(2), bi(DefaultBufferEpochs))
	if deposit.Cmp(expectedBuffer) != 0 {
		t.Errorf("deposit should equal buffer when funds are sufficient: got %s, want %s",
			deposit, expectedBuffer)
	}
}

func TestDepositNeeded_WithDebt(t *testing.T) {
	depositNoDebt := CalculateDepositNeeded(
		usdfc(10), bi(100), bi(50), bi(0), usdfc(1),
		DefaultRunwayEpochs, DefaultBufferEpochs, false,
	)
	depositWithDebt := CalculateDepositNeeded(
		usdfc(10), bi(100), bi(50), usdfc(5), usdfc(1),
		DefaultRunwayEpochs, DefaultBufferEpochs, false,
	)

	if depositWithDebt.Cmp(depositNoDebt) <= 0 {
		t.Errorf("deposit with debt should be larger: debt=%s, noDebt=%s",
			depositWithDebt, depositNoDebt)
	}
}

func TestDepositNeeded_BufferSkipped_NewDataSet_ZeroRate(t *testing.T) {
	depositNew := CalculateDepositNeeded(
		usdfc(10), // additionalLockup
		bi(100),   // rateDelta
		bi(0),     // currentLockupRate = 0
		bi(0),     // no debt
		bi(0),     // no available funds
		DefaultRunwayEpochs,
		DefaultBufferEpochs,
		true, // new dataset → buffer should be skipped
	)

	depositExisting := CalculateDepositNeeded(
		usdfc(10),
		bi(100),
		bi(0),
		bi(0),
		bi(0),
		DefaultRunwayEpochs,
		DefaultBufferEpochs,
		false, // existing → buffer applied
	)

	// for new dataset with zero current rate, buffer is skipped.
	// for existing, buffer = bufferEpochs * (0 + 100) = 5 * 100 = 500
	// so depositExisting should be larger.
	if depositNew.Cmp(depositExisting) >= 0 {
		t.Errorf("new dataset with zero rate should skip buffer and be smaller: new=%s, existing=%s",
			depositNew, depositExisting)
	}
}

func TestDepositNeeded_ZeroEverything(t *testing.T) {
	deposit := CalculateDepositNeeded(
		bi(0), bi(0), bi(0), bi(0), bi(0),
		0, 0,
		true,
	)
	if deposit.Sign() != 0 {
		t.Errorf("deposit should be zero when all inputs are zero: got %s", deposit)
	}
}
