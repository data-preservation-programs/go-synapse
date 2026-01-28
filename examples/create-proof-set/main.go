package main

import (
	"context"
	"fmt"
	"log"
	"os"

	"github.com/data-preservation-programs/go-synapse/constants"
	"github.com/data-preservation-programs/go-synapse/pdp"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ipfs/go-cid"
)

func main() {
	if err := run(); err != nil {
		log.Fatal(err)
	}
}

func run() error {
	ctx := context.Background()

	// Get configuration from environment
	privateKeyHex := os.Getenv("PRIVATE_KEY")
	if privateKeyHex == "" {
		return fmt.Errorf("PRIVATE_KEY environment variable not set")
	}

	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		rpcURL = "https://api.calibration.node.glif.io/rpc/v1"
		log.Printf("Using default RPC URL: %s", rpcURL)
	}

	listenerAddr := os.Getenv("LISTENER_ADDRESS")
	if listenerAddr == "" {
		return fmt.Errorf("LISTENER_ADDRESS environment variable not set")
	}

	// Parse private key
	privateKey, err := crypto.HexToECDSA(privateKeyHex)
	if err != nil {
		return fmt.Errorf("failed to parse private key: %w", err)
	}

	// Connect to Filecoin network
	client, err := ethclient.Dial(rpcURL)
	if err != nil {
		return fmt.Errorf("failed to connect to RPC: %w", err)
	}
	defer client.Close()

	// Detect network
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return fmt.Errorf("failed to get chain ID: %w", err)
	}

	var network constants.Network
	switch chainID.Int64() {
	case 314:
		network = constants.NetworkMainnet
	case 314159:
		network = constants.NetworkCalibration
	default:
		return fmt.Errorf("unsupported chain ID: %d", chainID.Int64())
	}

	log.Printf("Connected to %s (Chain ID: %d)", network, chainID.Int64())

	// Create proof set manager
	// Note: Using NewManager for backward compatibility, but NewManagerWithContext
	// or NewManagerWithConfig are recommended for new code
	manager, err := pdp.NewManager(client, privateKey, network)
	if err != nil {
		return fmt.Errorf("failed to create manager: %w", err)
	}

	// Alternative: Use NewManagerWithContext for explicit context support
	// manager, err := pdp.NewManagerWithContext(ctx, client, privateKey, network)

	// Alternative: Use NewManagerWithConfig for custom gas buffer
	// config := pdp.DefaultManagerConfig()
	// config.GasBufferPercent = 15  // Custom 15% buffer
	// manager, err := pdp.NewManagerWithConfig(ctx, client, privateKey, network, &config)

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	log.Printf("Using address: %s", address.Hex())

	// Create a new proof set
	log.Println("Creating proof set...")
	result, err := manager.CreateProofSet(ctx, pdp.CreateProofSetOptions{
		Listener:  common.HexToAddress(listenerAddr),
		ExtraData: []byte{},
	})
	if err != nil {
		return fmt.Errorf("failed to create proof set: %w", err)
	}

	log.Printf("✓ Proof set created successfully!")
	log.Printf("  Proof Set ID: %s", result.ProofSetID.String())
	log.Printf("  Transaction Hash: %s", result.TransactionHash.Hex())
	log.Printf("  Block Number: %d", result.Receipt.BlockNumber.Uint64())
	log.Printf("  Gas Used: %d", result.Receipt.GasUsed)

	// Query the proof set
	log.Println("\nQuerying proof set details...")
	proofSet, err := manager.GetProofSet(ctx, result.ProofSetID)
	if err != nil {
		return fmt.Errorf("failed to get proof set: %w", err)
	}

	log.Printf("✓ Proof Set Details:")
	log.Printf("  ID: %s", proofSet.ID.String())
	log.Printf("  Live: %v", proofSet.Live)
	log.Printf("  Listener: %s", proofSet.Listener.Hex())
	log.Printf("  Storage Provider: %s", proofSet.StorageProvider.Hex())
	log.Printf("  Active Pieces: %d", proofSet.ActivePieces)
	log.Printf("  Leaf Count: %d", proofSet.LeafCount)
	log.Printf("  Next Piece ID: %d", proofSet.NextPieceID)

	// Optionally add roots if PIECE_CID is provided
	pieceCIDStr := os.Getenv("PIECE_CID")
	if pieceCIDStr != "" {
		log.Println("\nAdding piece to proof set...")

		pieceCID, err := cid.Parse(pieceCIDStr)
		if err != nil {
			return fmt.Errorf("failed to parse piece CID: %w", err)
		}

		roots := []pdp.Root{
			{PieceCID: pieceCID},
		}

		addResult, err := manager.AddRoots(ctx, result.ProofSetID, roots)
		if err != nil {
			return fmt.Errorf("failed to add roots: %w", err)
		}

		log.Printf("✓ Added %d piece(s) successfully!", addResult.RootsAdded)
		log.Printf("  Transaction Hash: %s", addResult.TransactionHash.Hex())
		log.Printf("  Block Number: %d", addResult.Receipt.BlockNumber.Uint64())
		log.Printf("  Piece IDs: %v", addResult.PieceIDs)

		// Query updated proof set
		log.Println("\nQuerying updated proof set...")
		updatedProofSet, err := manager.GetProofSet(ctx, result.ProofSetID)
		if err != nil {
			return fmt.Errorf("failed to get updated proof set: %w", err)
		}

		log.Printf("✓ Updated Proof Set:")
		log.Printf("  Active Pieces: %d", updatedProofSet.ActivePieces)
		log.Printf("  Leaf Count: %d", updatedProofSet.LeafCount)
	}

	log.Println("\n✓ Complete!")
	return nil
}
