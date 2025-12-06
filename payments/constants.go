package payments

import (
	"math/big"
	"time"

	"github.com/data-preservation-programs/go-synapse/constants"
	"github.com/ethereum/go-ethereum/common"
)

var USDFCAddresses = constants.USDFCAddressesByChainID
var Multicall3Address = constants.Multicall3Addresses[constants.NetworkMainnet]

const (
	EpochDuration  = constants.EpochDuration
	EpochsPerDay   = constants.EpochsPerDay
	EpochsPerMonth = constants.EpochsPerMonth
)

var SettlementFee = big.NewInt(1300000000000000)

var GenesisTimestamps = constants.GenesisTimestampsByChainID

const PermitDeadline = 3600 * time.Second
const TokenDecimals = 18

var (
	CurrentEpoch = constants.CurrentEpoch
	EpochToTime  = constants.EpochToTime
	TimeToEpoch  = constants.TimeToEpoch
)

var PaymentsAddresses = map[int64]common.Address{
	constants.ChainIDMainnet:     constants.PaymentsAddresses[constants.NetworkMainnet],
	constants.ChainIDCalibration: constants.PaymentsAddresses[constants.NetworkCalibration],
}
