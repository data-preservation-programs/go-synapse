package costs

import "math/big"

type EffectiveRate struct {
	// RatePerEpoch uses integer division to match on-chain Solidity truncation.
	RatePerEpoch *big.Int
	// RatePerMonth preserves more precision for display.
	RatePerMonth *big.Int
}

type AdditionalLockup struct {
	RateDelta   *big.Int
	TotalLockup *big.Int
}

type UploadCosts struct {
	Rate                 EffectiveRate
	DepositNeeded        *big.Int
	NeedsFWSSMaxApproval bool
	Ready                bool
}

type UploadCostOptions struct {
	RunwayEpochs int64 // defaults to DefaultRunwayEpochs (3 months)
	BufferEpochs int64 // defaults to DefaultBufferEpochs (1 month)
	EnableCDN    bool
	IsNewDataSet bool
}
