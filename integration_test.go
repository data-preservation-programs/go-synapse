//go:build integration

package synapse_test

import (
	"context"
	"math/big"
	"os"
	"testing"
	"time"

	"github.com/data-preservation-programs/go-synapse/constants"
	"github.com/data-preservation-programs/go-synapse/pdp"
	"github.com/data-preservation-programs/go-synapse/signer"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
)

// Integration tests for go-synapse against Calibration testnet
//
// To run these tests, you need:
// 1. A Calibration testnet RPC endpoint
// 2. A test wallet with FIL for gas fees
// 3. A listener address (record keeper contract)
//
// Run with:
//   export CALIBRATION_RPC=https://api.calibration.node.glif.io/rpc/v1
//   export TEST_PRIVATE_KEY=your-test-private-key
//   export TEST_LISTENER_ADDRESS=0xYourListenerAddress
//   go test -tags=integration -v ./...

func getTestConfig(t *testing.T) (string, string, string) {
	rpcURL := os.Getenv("CALIBRATION_RPC")
	if rpcURL == "" {
		t.Skip("CALIBRATION_RPC not set, skipping integration test")
	}

	privateKeyHex := os.Getenv("TEST_PRIVATE_KEY")
	if privateKeyHex == "" {
		t.Skip("TEST_PRIVATE_KEY not set, skipping integration test")
	}

	listenerAddr := os.Getenv("TEST_LISTENER_ADDRESS")
	if listenerAddr == "" {
		t.Skip("TEST_LISTENER_ADDRESS not set, skipping integration test")
	}

	return rpcURL, privateKeyHex, listenerAddr
}

func TestIntegration_ProofSetLifecycle(t *testing.T) {
	rpcURL, privateKeyHex, listenerAddr := getTestConfig(t)

	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()

	// Parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	// Connect to Calibration testnet
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		t.Fatalf("Failed to connect to RPC: %v", err)
	}
	defer client.Close()

	// Create proof set manager
	s, err := signer.NewSecp256k1SignerFromECDSA(privateKey)
	if err != nil {
		t.Fatalf("Failed to create signer: %v", err)
	}
	manager, err := pdp.NewManagerWithContext(ctx, client, s, constants.NetworkCalibration)
	if err != nil {
		t.Fatalf("Failed to create manager: %v", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	t.Logf("Using address: %s", address.Hex())

	// Test 1: Create a proof set
	t.Run("CreateProofSet", func(t *testing.T) {
		result, err := manager.CreateProofSet(ctx, pdp.CreateProofSetOptions{
			Listener:  common.HexToAddress(listenerAddr),
			ExtraData: []byte{},
		})
		if err != nil {
			t.Fatalf("Failed to create proof set: %v", err)
		}

		if result.ProofSetID == nil {
			t.Fatal("Proof set ID is nil")
		}

		t.Logf("Created proof set: ID=%s, TxHash=%s",
			result.ProofSetID.String(),
			result.TransactionHash.Hex())

		// Test 2: Query the proof set
		t.Run("GetProofSet", func(t *testing.T) {
			proofSet, err := manager.GetProofSet(ctx, result.ProofSetID)
			if err != nil {
				t.Fatalf("Failed to get proof set: %v", err)
			}

			if !proofSet.Live {
				t.Error("Expected proof set to be live")
			}

			if proofSet.Listener != common.HexToAddress(listenerAddr) {
				t.Errorf("Expected listener %s, got %s",
					listenerAddr,
					proofSet.Listener.Hex())
			}

			t.Logf("Proof set details: ActivePieces=%d, LeafCount=%d",
				proofSet.ActivePieces,
				proofSet.LeafCount)
		})

		// Test 3: Check if data set is live
		t.Run("DataSetLive", func(t *testing.T) {
			live, err := manager.DataSetLive(ctx, result.ProofSetID)
			if err != nil {
				t.Fatalf("Failed to check if data set is live: %v", err)
			}

			if !live {
				t.Error("Expected data set to be live")
			}
		})

		// Test 4: Get next challenge epoch
		t.Run("GetNextChallengeEpoch", func(t *testing.T) {
			epoch, err := manager.GetNextChallengeEpoch(ctx, result.ProofSetID)
			if err != nil {
				t.Fatalf("Failed to get next challenge epoch: %v", err)
			}

			t.Logf("Next challenge epoch: %d", epoch)
		})
	})
}

func TestIntegration_ContractConnection(t *testing.T) {
	rpcURL, _, _ := getTestConfig(t)

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Connect to Calibration testnet
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		t.Fatalf("Failed to connect to RPC: %v", err)
	}
	defer client.Close()

	// Verify chain ID
	chainID, err := client.ChainID(ctx)
	if err != nil {
		t.Fatalf("Failed to get chain ID: %v", err)
	}

	expectedChainID := big.NewInt(314159) // Calibration
	if chainID.Cmp(expectedChainID) != 0 {
		t.Fatalf("Expected chain ID %s, got %s", expectedChainID.String(), chainID.String())
	}

	t.Logf("Connected to Calibration testnet (Chain ID: %s)", chainID.String())

	// Verify PDPVerifier contract address is set
	pdpVerifierAddr := constants.GetPDPVerifierAddress(constants.NetworkCalibration)
	if pdpVerifierAddr == (common.Address{}) {
		t.Fatal("PDPVerifier address not set for Calibration network")
	}

	t.Logf("PDPVerifier contract address: %s", pdpVerifierAddr.Hex())

	// Check if contract exists by querying its code
	code, err := client.CodeAt(ctx, pdpVerifierAddr, nil)
	if err != nil {
		t.Fatalf("Failed to get contract code: %v", err)
	}

	if len(code) == 0 {
		t.Fatalf("No code at PDPVerifier address %s", pdpVerifierAddr.Hex())
	}

	t.Logf("Contract verified: %d bytes of code", len(code))
}
