//go:build integration

package constants

import (
	"context"
	"math/big"
	"strings"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum"
	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// minimal ABIs - Payments.accounts is the most reliable test
var paymentsABI = `[{"name":"accounts","type":"function","inputs":[{"name":"token","type":"address"},{"name":"owner","type":"address"}],"outputs":[{"name":"balance","type":"uint256"},{"name":"lockedBalance","type":"uint256"}],"stateMutability":"view"}]`

type contractTest struct {
	name    string
	address common.Address
}

func TestGeneratedAddresses_MainnetContracts(t *testing.T) {
	testContracts(t, NetworkMainnet, RPCURLs[NetworkMainnet])
}

func TestGeneratedAddresses_CalibrationContracts(t *testing.T) {
	testContracts(t, NetworkCalibration, RPCURLs[NetworkCalibration])
}

func testContracts(t *testing.T, network Network, rpcURL string) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	client, err := ethclient.DialContext(ctx, rpcURL)
	if err != nil {
		t.Fatalf("failed to connect to %s: %v", network, err)
	}
	defer client.Close()

	// verify we can reach the network
	chainID, err := client.ChainID(ctx)
	if err != nil {
		t.Fatalf("failed to get chain ID: %v", err)
	}
	t.Logf("connected to %s (chain ID: %d)", network, chainID)

	tests := []contractTest{
		{"Payments", PaymentsAddresses[network]},
		{"StateView", WarmStorageStateViewAddresses[network]},
		{"PDPVerifier", PDPVerifierAddresses[network]},
		{"SPRegistry", SPRegistryAddresses[network]},
		{"SessionKeyRegistry", SessionKeyRegistryAddresses[network]},
	}

	for _, tc := range tests {
		t.Run(tc.name, func(t *testing.T) {
			// verify contract has code deployed
			code, err := client.CodeAt(ctx, tc.address, nil)
			if err != nil {
				t.Fatalf("failed to get code at %s: %v", tc.address.Hex(), err)
			}
			if len(code) == 0 {
				t.Fatalf("no code at %s - not a contract", tc.address.Hex())
			}
			t.Logf("%s: %s (%d bytes)", tc.name, tc.address.Hex(), len(code))
		})
	}

	// deeper test: verify Payments contract responds to known method
	t.Run("Payments_ABI", func(t *testing.T) {
		parsed, err := abi.JSON(strings.NewReader(paymentsABI))
		if err != nil {
			t.Fatalf("failed to parse ABI: %v", err)
		}

		zeroAddr := common.Address{}
		data, err := parsed.Pack("accounts", zeroAddr, zeroAddr)
		if err != nil {
			t.Fatalf("failed to pack accounts call: %v", err)
		}

		paymentsAddr := PaymentsAddresses[network]
		result, err := client.CallContract(ctx, ethereum.CallMsg{
			To:   &paymentsAddr,
			Data: data,
		}, nil)
		if err != nil {
			t.Fatalf("accounts() call failed: %v", err)
		}

		// unpack and verify we get two uint256 values
		var balance, lockedBalance *big.Int
		unpacked, err := parsed.Unpack("accounts", result)
		if err != nil {
			t.Fatalf("failed to unpack: %v", err)
		}
		balance = unpacked[0].(*big.Int)
		lockedBalance = unpacked[1].(*big.Int)
		t.Logf("Payments.accounts(0x0, 0x0) = {balance: %s, locked: %s}", balance, lockedBalance)
	})
}
