package constants

import (
	"math/big"
	"time"
)

const (
	EpochDuration = 30 * time.Second

	EpochDurationSeconds = 30

	EpochsPerDay = 2880

	EpochsPerMonth = 86400
)

const (
	TransactionPropagationTimeoutMS = 180000

	DataSetCreationTimeoutMS = 7 * 60 * 1000

	PieceParkingTimeoutMS = 7 * 60 * 1000

	PieceAdditionTimeoutMS = 7 * 60 * 1000

	TransactionPropagationPollIntervalMS = 2000
	DataSetCreationPollIntervalMS        = 2000
	PieceParkingPollIntervalMS           = 5000
	PieceAdditionPollIntervalMS          = 1000
)

func CurrentEpoch(chainID int64) *big.Int {
	genesis, ok := GenesisTimestampsByChainID[chainID]
	if !ok {
		return big.NewInt(0)
	}
	now := time.Now().Unix()
	epochsSinceGenesis := (now - genesis) / EpochDurationSeconds
	return big.NewInt(epochsSinceGenesis)
}

func EpochToTime(chainID int64, epoch *big.Int) time.Time {
	genesis, ok := GenesisTimestampsByChainID[chainID]
	if !ok {
		return time.Time{}
	}
	epochSeconds := epoch.Int64() * EpochDurationSeconds
	return time.Unix(genesis+epochSeconds, 0)
}

func TimeToEpoch(chainID int64, t time.Time) *big.Int {
	genesis, ok := GenesisTimestampsByChainID[chainID]
	if !ok {
		return big.NewInt(0)
	}
	epochsSinceGenesis := (t.Unix() - genesis) / EpochDurationSeconds
	return big.NewInt(epochsSinceGenesis)
}
