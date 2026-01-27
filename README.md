# go-synapse

A Go SDK for interacting with the FilOzone Synapse protocol, enabling Proof of Data Possession (PDP) deals on Filecoin.

## Features

- **PDP Proof Set Management**: Create and manage proof sets on-chain
- **Contract Interactions**: Direct interaction with PDPVerifier smart contracts
- **Transaction Management**: Robust transaction handling with retry logic and nonce management
- **Storage Provider API**: Upload and manage data with storage providers
- **Network Support**: Filecoin Mainnet and Calibration testnet

## Installation

```bash
go get github.com/data-preservation-programs/go-synapse
```

## Quick Start

### Initialize a Client

```go
package main

import (
    "context"
    "crypto/ecdsa"
    "log"

    "github.com/data-preservation-programs/go-synapse"
    "github.com/ethereum/go-ethereum/crypto"
)

func main() {
    ctx := context.Background()

    // Load your private key
    privateKey, err := crypto.HexToECDSA("your-private-key-hex")
    if err != nil {
        log.Fatal(err)
    }

    // Create a new synapse client
    client, err := synapse.New(ctx, synapse.Options{
        PrivateKey:  privateKey,
        RPCURL:      "https://api.calibration.node.glif.io/rpc/v1",
        ProviderURL: "https://your-storage-provider.example.com",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    log.Printf("Connected to network: %s", client.Network())
    log.Printf("Address: %s", client.Address())
}
```

### Working with Proof Sets

```go
package main

import (
    "context"
    "crypto/ecdsa"
    "log"
    "math/big"

    "github.com/data-preservation-programs/go-synapse/constants"
    "github.com/data-preservation-programs/go-synapse/pdp"
    "github.com/ethereum/go-ethereum/common"
    "github.com/ethereum/go-ethereum/crypto"
    "github.com/ethereum/go-ethereum/ethclient"
    "github.com/ipfs/go-cid"
)

func main() {
    ctx := context.Background()

    // Setup
    privateKey, _ := crypto.HexToECDSA("your-private-key-hex")
    client, _ := ethclient.Dial("https://api.calibration.node.glif.io/rpc/v1")

    // Create proof set manager
    manager, err := pdp.NewManager(client, privateKey, constants.NetworkCalibration)
    if err != nil {
        log.Fatal(err)
    }

    // Create a new proof set
    result, err := manager.CreateProofSet(ctx, pdp.CreateProofSetOptions{
        Listener:  common.HexToAddress("0xYourListenerAddress"),
        ExtraData: []byte{},
    })
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Proof set created: ID=%s, TxHash=%s",
        result.ProofSetID.String(),
        result.TransactionHash.Hex())

    // Add roots to the proof set
    pieceCID, _ := cid.Parse("baga6ea4seaqao7s73y24kcutaosvacpdjgfe5pw76ooefnyqw4ynr3d2y6x2mpq")
    roots := []pdp.Root{
        {PieceCID: pieceCID},
    }

    addResult, err := manager.AddRoots(ctx, result.ProofSetID, roots)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Added %d roots, TxHash=%s",
        addResult.RootsAdded,
        addResult.TransactionHash.Hex())

    // Query proof set details
    proofSet, err := manager.GetProofSet(ctx, result.ProofSetID)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Proof Set Status:")
    log.Printf("  Live: %v", proofSet.Live)
    log.Printf("  Active Pieces: %d", proofSet.ActivePieces)
    log.Printf("  Leaf Count: %d", proofSet.LeafCount)
    log.Printf("  Storage Provider: %s", proofSet.StorageProvider.Hex())
}
```

### Upload Data to Storage Provider

```go
package main

import (
    "context"
    "log"
    "os"

    "github.com/data-preservation-programs/go-synapse"
    "github.com/ethereum/go-ethereum/crypto"
)

func main() {
    ctx := context.Background()

    privateKey, _ := crypto.HexToECDSA("your-private-key-hex")

    client, err := synapse.New(ctx, synapse.Options{
        PrivateKey:  privateKey,
        RPCURL:      "https://api.calibration.node.glif.io/rpc/v1",
        ProviderURL: "https://your-storage-provider.example.com",
    })
    if err != nil {
        log.Fatal(err)
    }
    defer client.Close()

    // Get storage manager
    storage, err := client.Storage()
    if err != nil {
        log.Fatal(err)
    }

    // Upload a file
    file, _ := os.Open("data.txt")
    defer file.Close()

    result, err := storage.UploadFile(ctx, file)
    if err != nil {
        log.Fatal(err)
    }

    log.Printf("Uploaded file: PieceCID=%s, Size=%d",
        result.PieceCID.String(),
        result.Size)
}
```

## API Overview

### Core Components

#### `synapse.Client`
Main client for interacting with the Synapse protocol.

- `New()` - Create a new client
- `Network()` - Get current network
- `Address()` - Get wallet address
- `Storage()` - Get storage manager
- `Close()` - Clean up resources

#### `pdp.ProofSetManager`
Manage proof sets on-chain.

- `CreateProofSet()` - Create a new proof set
- `GetProofSet()` - Retrieve proof set details
- `AddRoots()` - Add piece CIDs to a proof set
- `GetRoots()` - List roots with pagination
- `DeleteProofSet()` - Remove a proof set
- `GetNextChallengeEpoch()` - Query challenge schedule
- `DataSetLive()` - Check if proof set is active

#### `storage.Manager`
Handle file uploads and storage operations.

- `UploadFile()` - Upload a file to storage provider
- `UploadData()` - Upload raw data
- `FindPiece()` - Check if a piece exists
- `DownloadPiece()` - Retrieve piece data

#### `pkg/txutil`
Transaction utilities for robust blockchain interactions.

- `WaitForConfirmation()` - Wait for transaction confirmations
- `EstimateGasWithBuffer()` - Estimate gas with safety margin
- `SendTransactionWithRetry()` - Send transactions with retry logic
- `NonceManager` - Thread-safe nonce management

## Contract Addresses

### Filecoin Mainnet (Chain ID: 314)
- **PDPVerifier**: `0xBADd0B92C1c71d02E7d520f64c0876538fa2557F`

### Filecoin Calibration (Chain ID: 314159)
- **PDPVerifier**: `0x85e366Cf9DD2c0aE37E963d9556F5f4718d6417C`

## Configuration

### Network Detection
The client automatically detects the network based on chain ID:
- Chain ID 314 → Filecoin Mainnet
- Chain ID 314159 → Filecoin Calibration

### RPC Endpoints

**Mainnet:**
- `https://api.node.glif.io/rpc/v1`

**Calibration:**
- `https://api.calibration.node.glif.io/rpc/v1`

## Examples

See the [`examples/`](./examples/) directory for more detailed examples:

- [Create Proof Set](./examples/create-proof-set/) - Create and manage proof sets
- [Upload Data](./examples/upload/) - Upload files to storage providers

## Development

### Building

```bash
go build ./...
```

### Running Tests

```bash
# Run all tests
go test ./...

# Run with verbose output
go test -v ./...

# Run specific package tests
go test ./pkg/txutil/...
```

### Integration Tests

Integration tests require access to a Filecoin testnet:

```bash
export CALIBRATION_RPC=https://api.calibration.node.glif.io/rpc/v1
export TEST_PRIVATE_KEY=your-test-private-key
go test -tags=integration -v ./...
```

## Dependencies

- `github.com/ethereum/go-ethereum` - Ethereum client for contract interaction
- `github.com/ipfs/go-cid` - Content identifiers
- `github.com/filecoin-project/go-commp-utils` - Piece CID utilities

## License

This project is part of the Data Preservation Programs organization.

## Related Projects

- [synapse-sdk](https://github.com/FilOzone/synapse-sdk) - TypeScript/JavaScript SDK
- [singularity](https://github.com/data-preservation-programs/singularity) - Data onboarding tool

## Support

For issues and questions:
- Open an issue on [GitHub](https://github.com/data-preservation-programs/go-synapse/issues)
- Join the [Filecoin Slack](https://filecoin.io/slack)
