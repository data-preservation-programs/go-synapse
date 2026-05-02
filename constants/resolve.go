package constants

import (
	"context"
	"fmt"
	"strings"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

const fwssABI = `[
	{"type":"function","name":"paymentsContractAddress","inputs":[],"outputs":[{"name":"","type":"address"}],"stateMutability":"view"},
	{"type":"function","name":"viewContractAddress","inputs":[],"outputs":[{"name":"","type":"address"}],"stateMutability":"view"},
	{"type":"function","name":"pdpVerifierAddress","inputs":[],"outputs":[{"name":"","type":"address"}],"stateMutability":"view"},
	{"type":"function","name":"serviceProviderRegistry","inputs":[],"outputs":[{"name":"","type":"address"}],"stateMutability":"view"},
	{"type":"function","name":"sessionKeyRegistry","inputs":[],"outputs":[{"name":"","type":"address"}],"stateMutability":"view"}
]`

// NetworkAddresses holds the full set of addresses derived from an FWSS contract.
type NetworkAddresses struct {
	FWSS               common.Address
	Payments           common.Address
	StateView          common.Address
	PDPVerifier        common.Address
	SPRegistry         common.Address
	SessionKeyRegistry common.Address
}

// ResolveFromFWSS queries a FWSS contract to derive all dependent addresses.
// This is the runtime equivalent of what `go generate` does at build time
// for mainnet and calibration.
func ResolveFromFWSS(ctx context.Context, client *ethclient.Client, fwssAddr common.Address) (*NetworkAddresses, error) {
	parsed, err := abi.JSON(strings.NewReader(fwssABI))
	if err != nil {
		return nil, fmt.Errorf("parse FWSS abi: %w", err)
	}

	callView := func(method string) (common.Address, error) {
		data, err := parsed.Pack(method)
		if err != nil {
			return common.Address{}, fmt.Errorf("pack %s: %w", method, err)
		}
		result, err := client.CallContract(ctx, ethereum.CallMsg{
			To:   &fwssAddr,
			Data: data,
		}, nil)
		if err != nil {
			return common.Address{}, fmt.Errorf("call %s: %w", method, err)
		}
		var addr common.Address
		// safe: single primitive output, not a named tuple -- the
		// UnpackIntoInterface bug abix.UnpackSingleTuple guards against
		// only manifests for tuple returns.
		if err := parsed.UnpackIntoInterface(&addr, method, result); err != nil {
			return common.Address{}, fmt.Errorf("unpack %s: %w", method, err)
		}
		return addr, nil
	}

	addrs := &NetworkAddresses{FWSS: fwssAddr}

	addrs.Payments, err = callView("paymentsContractAddress")
	if err != nil {
		return nil, err
	}
	addrs.StateView, err = callView("viewContractAddress")
	if err != nil {
		return nil, err
	}
	addrs.PDPVerifier, err = callView("pdpVerifierAddress")
	if err != nil {
		return nil, err
	}
	addrs.SPRegistry, err = callView("serviceProviderRegistry")
	if err != nil {
		return nil, err
	}
	addrs.SessionKeyRegistry, err = callView("sessionKeyRegistry")
	if err != nil {
		return nil, err
	}

	return addrs, nil
}

// RegisterNetwork populates the package-level address maps for a network.
// Use this for devnet or custom deployments where addresses are resolved
// at runtime from FWSS rather than baked in at build time.
func RegisterNetwork(network Network, addrs *NetworkAddresses) {
	WarmStorageAddresses[network] = addrs.FWSS
	PaymentsAddresses[network] = addrs.Payments
	WarmStorageStateViewAddresses[network] = addrs.StateView
	PDPVerifierAddresses[network] = addrs.PDPVerifier
	SPRegistryAddresses[network] = addrs.SPRegistry
	SessionKeyRegistryAddresses[network] = addrs.SessionKeyRegistry
}
