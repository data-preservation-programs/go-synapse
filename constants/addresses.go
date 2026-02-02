package constants

import (
	"github.com/ethereum/go-ethereum/common"
)

type Network string

const (
	NetworkMainnet Network = "mainnet"
	NetworkCalibration Network = "calibration"
)

const (
	ChainIDMainnet     int64 = 314
	ChainIDCalibration int64 = 314159
)

var (
	WarmStorageAddresses = map[Network]common.Address{
		NetworkMainnet:     common.HexToAddress("0x8408502033C418E1bbC97cE9ac48E5528F371A9f"),
		NetworkCalibration: common.HexToAddress("0x02925630df557F957f70E112bA06e50965417CA0"),
	}

	SPRegistryAddresses = map[Network]common.Address{
		NetworkMainnet:     common.HexToAddress("0xf55dDbf63F1b55c3F1D4FA7e339a68AB7b64A5eB"),
		NetworkCalibration: common.HexToAddress("0x839e5c9988e4e9977d40708d0094103c0839Ac9D"),
	}

	Multicall3Addresses = map[Network]common.Address{
		NetworkMainnet:     common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
		NetworkCalibration: common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
	}

	USDFCAddresses = map[Network]common.Address{
		NetworkMainnet:     common.HexToAddress("0x80B98d3aa09ffff255c3ba4A241111Ff1262F045"),
		NetworkCalibration: common.HexToAddress("0xb3042734b608a1B16e9e86B374A3f3e389B4cDf0"),
	}

	PaymentsAddresses = map[Network]common.Address{
		NetworkMainnet:     common.HexToAddress("0xC8C3C94aa8C60E0dFF060D4c3acBa0bC16e4e0ec"),
		NetworkCalibration: common.HexToAddress("0xD58af75a0F6ed91E8d416CAB72Ebae40E05ecD44"),
	}

	WarmStorageStateViewAddresses = map[Network]common.Address{
		NetworkMainnet:     common.HexToAddress("0x9e4e6699d8F67dFc883d6b0A7344Bd56F7E80B46"),
		NetworkCalibration: common.HexToAddress("0xA5D87b04086B1d591026cCE10255351B5AA4689B"),
	}
)

var RPCURLs = map[Network]string{
	NetworkMainnet:     "https://api.node.glif.io/rpc/v1",
	NetworkCalibration: "https://api.calibration.node.glif.io/rpc/v1",
}

var GenesisTimestamps = map[Network]int64{
	NetworkMainnet:     1598306400,
	NetworkCalibration: 1667326380,
}

var GenesisTimestampsByChainID = map[int64]int64{
	ChainIDMainnet:     1598306400,
	ChainIDCalibration: 1667326380,
}

var USDFCAddressesByChainID = map[int64]common.Address{
	ChainIDMainnet:     common.HexToAddress("0x80B98d3aa09ffff255c3ba4A241111Ff1262F045"),
	ChainIDCalibration: common.HexToAddress("0xb3042734b608a1B16e9e86B374A3f3e389B4cDf0"),
}

// NetworkChainIDs maps network to expected chain ID
var NetworkChainIDs = map[Network]int64{
	NetworkMainnet:     ChainIDMainnet,
	NetworkCalibration: ChainIDCalibration,
}

// ExpectedChainID returns the expected chain ID for a given network.
// Returns the chain ID and true if the network is known, or 0 and false otherwise.
func ExpectedChainID(network Network) (int64, bool) {
	chainID, ok := NetworkChainIDs[network]
	return chainID, ok
}
