//go:generate go run ../internal/generate/addresses.go

package constants

import (
	"github.com/ethereum/go-ethereum/common"
)

type Network string

const (
	NetworkMainnet     Network = "mainnet"
	NetworkCalibration Network = "calibration"
	NetworkDevnet      Network = "devnet"
)

const (
	ChainIDMainnet     int64 = 314
	ChainIDCalibration int64 = 314159
	ChainIDDevnet      int64 = 31415926
)

// static addresses not derived from FWSS
var (
	Multicall3Addresses = map[Network]common.Address{
		NetworkMainnet:     common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
		NetworkCalibration: common.HexToAddress("0xcA11bde05977b3631167028862bE2a173976CA11"),
		NetworkDevnet:      common.HexToAddress("0x2e1F1424b41ad7b2E34b0a60501edFc82FEf5BE8"),
	}

	USDFCAddresses = map[Network]common.Address{
		NetworkMainnet:     common.HexToAddress("0x80B98d3aa09ffff255c3ba4A241111Ff1262F045"),
		NetworkCalibration: common.HexToAddress("0xb3042734b608a1B16e9e86B374A3f3e389B4cDf0"),
		NetworkDevnet:      common.HexToAddress("0xB514FeE11119E0923950C09A181F1fa3aa62C80b"),
	}
)

var RPCURLs = map[Network]string{
	NetworkMainnet:     "https://api.node.glif.io/rpc/v1",
	NetworkCalibration: "https://api.calibration.node.glif.io/rpc/v1",
	NetworkDevnet:      "http://127.0.0.1:1234/rpc/v1",
}

var GenesisTimestamps = map[Network]int64{
	NetworkMainnet:     1598306400,
	NetworkCalibration: 1667326380,
	// devnet genesis is dynamic per run -- omitted here, query the node
}

var GenesisTimestampsByChainID = map[int64]int64{
	ChainIDMainnet:     1598306400,
	ChainIDCalibration: 1667326380,
}

var USDFCAddressesByChainID = map[int64]common.Address{
	ChainIDMainnet:     common.HexToAddress("0x80B98d3aa09ffff255c3ba4A241111Ff1262F045"),
	ChainIDCalibration: common.HexToAddress("0xb3042734b608a1B16e9e86B374A3f3e389B4cDf0"),
}

var NetworkChainIDs = map[Network]int64{
	NetworkMainnet:     ChainIDMainnet,
	NetworkCalibration: ChainIDCalibration,
	NetworkDevnet:      ChainIDDevnet,
}

// ExpectedChainID returns the expected chain ID for a given network.
// Returns the chain ID and true if the network is known, or 0 and false otherwise.
func ExpectedChainID(network Network) (int64, bool) {
	chainID, ok := NetworkChainIDs[network]
	return chainID, ok
}

// WarmStorageAddresses aliases the FWSS addresses (root of trust)
// WarmStorageAddresses -- initialized in addresses_generated.go from FWSS root of trust vars

// GetPDPVerifierAddress returns the PDPVerifier contract address for the given network
func GetPDPVerifierAddress(network Network) common.Address {
	addr, ok := PDPVerifierAddresses[network]
	if !ok {
		return common.Address{}
	}
	return addr
}
