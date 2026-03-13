package costs

import (
	"math/big"

	"github.com/data-preservation-programs/go-synapse/constants"
)

const (
	DefaultRunwayEpochs = 3 * constants.EpochsPerMonth // 3 months
	DefaultBufferEpochs = constants.EpochsPerMonth     // 1 month
	DefaultLockupPeriod = constants.EpochsPerMonth     // 1 month
)

var (
	CDNFixedLockup       = big.NewInt(700000000000000000) // 0.7 USDFC
	CacheMissFixedLockup = big.NewInt(300000000000000000) // 0.3 USDFC
)
