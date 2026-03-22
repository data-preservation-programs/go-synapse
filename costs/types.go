package costs

import "math/big"

type EffectiveRate struct {
	// RatePerEpoch uses integer division to match on-chain Solidity truncation.
	RatePerEpoch *big.Int
	// RatePerMonth preserves more precision for display.
	RatePerMonth *big.Int
}

type AdditionalLockup struct {
	RateDelta      *big.Int
	RateLockup     *big.Int // rateDelta * lockupPeriod
	CDNFixedLockup *big.Int // 1.0 USDFC for new CDN datasets, 0 otherwise
	SybilFee       *big.Int // sybil fee for new datasets, 0 otherwise
	TotalLockup    *big.Int // sum of all components
}

type UploadCosts struct {
	Rate                 EffectiveRate
	Lockup               AdditionalLockup
	DepositNeeded        *big.Int
	NeedsFWSSMaxApproval bool
	Ready                bool
}

type UploadCostOptions struct {
	RunwayEpochs int64 // defaults to DefaultRunwayEpochs (0)
	BufferEpochs int64 // defaults to DefaultBufferEpochs (5 epochs)
	EnableCDN    bool
	IsNewDataSet bool
}

type AccountSummary struct {
	Funds              *big.Int
	AvailableFunds     *big.Int
	Debt               *big.Int
	LockupRatePerEpoch *big.Int
	LockupRatePerMonth *big.Int
	FundedUntilEpoch   *big.Int
	CurrentEpoch       *big.Int
}
