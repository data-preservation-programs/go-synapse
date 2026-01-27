package constants

import "github.com/ethereum/go-ethereum/common"

// PDPVerifier contract addresses by network
var (
	// PDPVerifierMainnet is the PDPVerifier contract address on Filecoin Mainnet (Chain ID: 314)
	PDPVerifierMainnet = common.HexToAddress("0xBADd0B92C1c71d02E7d520f64c0876538fa2557F")

	// PDPVerifierCalibration is the PDPVerifier contract address on Filecoin Calibration testnet (Chain ID: 314159)
	PDPVerifierCalibration = common.HexToAddress("0x85e366Cf9DD2c0aE37E963d9556F5f4718d6417C")
)

// PDPVerifierAddresses maps network to PDPVerifier contract address
var PDPVerifierAddresses = map[Network]common.Address{
	NetworkMainnet:     PDPVerifierMainnet,
	NetworkCalibration: PDPVerifierCalibration,
}

// GetPDPVerifierAddress returns the PDPVerifier contract address for the given network
func GetPDPVerifierAddress(network Network) common.Address {
	addr, ok := PDPVerifierAddresses[network]
	if !ok {
		return common.Address{} // Return zero address if network not found
	}
	return addr
}
