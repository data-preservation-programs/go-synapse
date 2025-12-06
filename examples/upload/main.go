package main

import (
	"bytes"
	"context"
	"encoding/hex"
	"fmt"
	"log"
	"os"
	"strings"
	"time"

	synapse "github.com/data-preservation-programs/go-synapse"
	"github.com/ethereum/go-ethereum/crypto"
)

func main() {
	ctx, cancel := context.WithTimeout(context.Background(), 10*time.Minute)
	defer cancel()

	privateKeyHex := os.Getenv("PRIVATE_KEY")
	if privateKeyHex == "" {
		log.Fatal("PRIVATE_KEY environment variable is required")
	}
	privateKeyHex = strings.TrimPrefix(privateKeyHex, "0x")

	providerURL := os.Getenv("PROVIDER_URL")
	if providerURL == "" {
		log.Fatal("PROVIDER_URL environment variable is required")
	}

	rpcURL := os.Getenv("RPC_URL")
	if rpcURL == "" {
		rpcURL = "https://api.calibration.node.glif.io/rpc/v1"
	}

	privateKeyBytes, err := hex.DecodeString(privateKeyHex)
	if err != nil {
		log.Fatalf("Failed to decode private key: %v", err)
	}
	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		log.Fatalf("Failed to parse private key: %v", err)
	}

	fmt.Println("Connecting to Filecoin network...")
	client, err := synapse.New(ctx, synapse.Options{
		PrivateKey:  privateKey,
		RPCURL:      rpcURL,
		ProviderURL: providerURL,
	})
	if err != nil {
		log.Fatalf("Failed to create Synapse client: %v", err)
	}
	defer client.Close()

	fmt.Printf("Connected to %s (chain ID: %d)\n", client.Network(), client.ChainID())
	fmt.Printf("Client address: %s\n", client.Address().Hex())

	storage, err := client.Storage()
	if err != nil {
		log.Fatalf("Failed to get storage manager: %v", err)
	}

	testData := []byte("Hello, Filecoin! This is a test upload from the Synapse Go SDK.")
	fmt.Printf("\nUploading %d bytes of data...\n", len(testData))

	result, err := storage.UploadBytes(ctx, testData, nil)
	if err != nil {
		log.Fatalf("Upload failed: %v", err)
	}

	fmt.Printf("Upload successful!\n")
	fmt.Printf("  PieceCID: %s\n", result.PieceCID.String())
	fmt.Printf("  Size: %d bytes\n", result.Size)
	fmt.Printf("  PieceID: %d\n", result.PieceID)
	fmt.Printf("  DataSetID: %d\n", result.DataSetID)

	fmt.Printf("\nDownloading data...\n")
	downloadedData, err := storage.Download(ctx, result.PieceCID, nil)
	if err != nil {
		log.Fatalf("Download failed: %v", err)
	}

	fmt.Printf("Download successful!\n")
	fmt.Printf("  Size: %d bytes\n", len(downloadedData))

	if bytes.Equal(testData, downloadedData) {
		fmt.Println("  Data verified: MATCH")
	} else {
		fmt.Println("  Data verified: MISMATCH!")
	}

	fmt.Println("\nDone!")
}
