package synapse

import (
	"context"
	"fmt"
	"math/big"

	"github.com/data-preservation-programs/go-synapse/constants"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

type Network = constants.Network

const (
	NetworkMainnet     = constants.NetworkMainnet
	NetworkCalibration = constants.NetworkCalibration
	ChainIDMainnet     = constants.ChainIDMainnet
	ChainIDCalibration = constants.ChainIDCalibration
)

var (
	WarmStorageAddresses = constants.WarmStorageAddresses
	SPRegistryAddresses  = constants.SPRegistryAddresses
	Multicall3Addresses  = constants.Multicall3Addresses
	USDFCAddresses       = constants.USDFCAddresses
	RPCURLs              = constants.RPCURLs
	GenesisTimestamps    = constants.GenesisTimestamps
)

const (
	KiB           = constants.KiB
	MiB           = constants.MiB
	GiB           = constants.GiB
	TiB           = constants.TiB
	MaxUploadSize = constants.MaxUploadSize
	MinUploadSize = constants.MinUploadSize
)

const (
	EpochDuration                        = constants.EpochDurationSeconds
	EpochsPerDay                         = constants.EpochsPerDay
	EpochsPerMonth                       = constants.EpochsPerMonth
	TransactionPropagationTimeoutMS      = constants.TransactionPropagationTimeoutMS
	DataSetCreationTimeoutMS             = constants.DataSetCreationTimeoutMS
	PieceParkingTimeoutMS                = constants.PieceParkingTimeoutMS
	PieceAdditionTimeoutMS               = constants.PieceAdditionTimeoutMS
	TransactionPropagationPollIntervalMS = constants.TransactionPropagationPollIntervalMS
	DataSetCreationPollIntervalMS        = constants.DataSetCreationPollIntervalMS
	PieceParkingPollIntervalMS           = constants.PieceParkingPollIntervalMS
	PieceAdditionPollIntervalMS          = constants.PieceAdditionPollIntervalMS
)

func DetectNetwork(ctx context.Context, client *ethclient.Client) (Network, int64, error) {
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return "", 0, fmt.Errorf("failed to get chain ID: %w", err)
	}

	return NetworkFromChainID(chainID)
}

func NetworkFromChainID(chainID *big.Int) (Network, int64, error) {
	id := chainID.Int64()

	switch id {
	case ChainIDMainnet:
		return NetworkMainnet, id, nil
	case ChainIDCalibration:
		return NetworkCalibration, id, nil
	default:
		return "", 0, fmt.Errorf("unsupported chain ID: %d (expected %d for mainnet or %d for calibration)",
			id, ChainIDMainnet, ChainIDCalibration)
	}
}

func GetSPRegistryAddress(network Network) common.Address {
	return SPRegistryAddresses[network]
}

func GetWarmStorageAddress(network Network) common.Address {
	return WarmStorageAddresses[network]
}
