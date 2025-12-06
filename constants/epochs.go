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
	TransactionPropagationTimeoutMS = 180000 // 3 minutes

	DataSetCreationTimeoutMS = 7 * 60 * 1000 // 7 minutes

	PieceParkingTimeoutMS = 7 * 60 * 1000 // 7 minutes

	PieceAdditionTimeoutMS = 7 * 60 * 1000 // 7 minutes

	TransactionPropagationPollIntervalMS = 2000 // 2 seconds
	DataSetCreationPollIntervalMS        = 2000 // 2 seconds
	PieceParkingPollIntervalMS           = 5000 // 5 seconds
	PieceAdditionPollIntervalMS          = 1000 // 1 second
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
