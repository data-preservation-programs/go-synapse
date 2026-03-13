package costs

import (
	"math/big"

	"github.com/data-preservation-programs/go-synapse/constants"
	"github.com/data-preservation-programs/go-synapse/warmstorage"
)

var (
	bigOne = big.NewInt(1)
	bigTiB = big.NewInt(constants.TiB)
)

// CalculateEffectiveRate computes the storage rate for the given data size.
// ratePerEpoch uses integer division to match on-chain truncation.
// Both rates are floored to their corresponding minimums.
func CalculateEffectiveRate(
	sizeBytes *big.Int,
	pricePerTiBPerMonth *big.Int,
	minMonthlyRate *big.Int,
	epochsPerMonth int64,
) EffectiveRate {
	ratePerMonth := new(big.Int).Mul(pricePerTiBPerMonth, sizeBytes)
	ratePerMonth.Div(ratePerMonth, bigTiB)
	if ratePerMonth.Cmp(minMonthlyRate) < 0 {
		ratePerMonth.Set(minMonthlyRate)
	}

	epm := big.NewInt(epochsPerMonth)
	ratePerEpoch := new(big.Int).Mul(pricePerTiBPerMonth, sizeBytes)
	ratePerEpoch.Div(ratePerEpoch, bigTiB)
	ratePerEpoch.Div(ratePerEpoch, epm)

	minEpochRate := new(big.Int).Div(minMonthlyRate, epm)
	if minEpochRate.Cmp(bigOne) < 0 {
		minEpochRate.Set(bigOne)
	}
	if ratePerEpoch.Cmp(minEpochRate) < 0 {
		ratePerEpoch.Set(minEpochRate)
	}

	return EffectiveRate{
		RatePerEpoch: ratePerEpoch,
		RatePerMonth: ratePerMonth,
	}
}

func CalculateAdditionalLockupRequired(
	dataSizeBytes *big.Int,
	currentDataSetSizeBytes *big.Int,
	pricing *warmstorage.ServicePrice,
	lockupPeriod int64,
	usdfcSybilFee *big.Int,
	isNewDataSet bool,
	enableCDN bool,
) AdditionalLockup {
	newTotalSize := new(big.Int).Add(currentDataSetSizeBytes, dataSizeBytes)
	epm := pricing.EpochsPerMonth.Int64()

	newRate := CalculateEffectiveRate(
		newTotalSize,
		pricing.PricePerTiBPerMonthNoCDN,
		pricing.MinimumPricePerMonth,
		epm,
	)

	var rateDelta *big.Int
	if isNewDataSet {
		rateDelta = new(big.Int).Set(newRate.RatePerEpoch)
	} else {
		currentRate := CalculateEffectiveRate(
			currentDataSetSizeBytes,
			pricing.PricePerTiBPerMonthNoCDN,
			pricing.MinimumPricePerMonth,
			epm,
		)
		rateDelta = new(big.Int).Sub(newRate.RatePerEpoch, currentRate.RatePerEpoch)
		if rateDelta.Sign() < 0 {
			rateDelta.SetInt64(0)
		}
	}

	totalLockup := new(big.Int).Mul(rateDelta, big.NewInt(lockupPeriod))

	if isNewDataSet && enableCDN {
		totalLockup.Add(totalLockup, CDNFixedLockup)
		totalLockup.Add(totalLockup, CacheMissFixedLockup)
	}

	if isNewDataSet && usdfcSybilFee != nil {
		totalLockup.Add(totalLockup, usdfcSybilFee)
	}

	return AdditionalLockup{
		RateDelta:   rateDelta,
		TotalLockup: totalLockup,
	}
}

// CalculateDepositNeeded computes the USDFC deposit required to cover lockup,
// runway, and buffer. Buffer is skipped when currentLockupRate is zero and
// isNewDataSet is true (deposit lands before any rail is created).
func CalculateDepositNeeded(
	additionalLockup *big.Int,
	rateDelta *big.Int,
	currentLockupRate *big.Int,
	debt *big.Int,
	availableFunds *big.Int,
	runwayEpochs int64,
	bufferEpochs int64,
	isNewDataSet bool,
) *big.Int {
	combinedRate := new(big.Int).Add(currentLockupRate, rateDelta)
	runway := new(big.Int).Mul(combinedRate, big.NewInt(runwayEpochs))

	raw := new(big.Int).Add(additionalLockup, runway)
	raw.Sub(raw, availableFunds)
	raw.Add(raw, debt)

	if raw.Sign() < 0 {
		raw.SetInt64(0)
	}

	var buffer *big.Int
	if currentLockupRate.Sign() == 0 && isNewDataSet {
		buffer = new(big.Int)
	} else {
		buffer = new(big.Int).Mul(combinedRate, big.NewInt(bufferEpochs))
	}

	return new(big.Int).Add(raw, buffer)
}
