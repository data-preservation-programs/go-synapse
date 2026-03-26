package costs

import (
	"math/big"

	"github.com/data-preservation-programs/go-synapse/constants"
)

const (
	DefaultRunwayEpochs = 0                        // match synapse-sdk: no extra runway
	DefaultBufferEpochs = 5                        // match synapse-sdk: 5 epoch execution buffer
	DefaultLockupPeriod = constants.EpochsPerMonth // 30 days
)

var (
	CDNFixedLockup       = big.NewInt(1000000000000000000) // 1.0 USDFC (combined CDN + cache miss)
	UsdfcSybilFeeDefault = big.NewInt(100000000000000000)  // 0.1 USDFC fallback
)
