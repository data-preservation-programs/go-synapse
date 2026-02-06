# Create Proof Set Example

This example demonstrates how to create a proof set on the PDPVerifier contract and optionally add pieces to it.

## Prerequisites

- A wallet with FIL for gas fees
- An RPC endpoint (default: Calibration testnet)
- A listener address (record keeper contract)

## Usage

### Basic Usage - Create Proof Set

```bash
export PRIVATE_KEY=your-private-key-hex
export LISTENER_ADDRESS=0xYourListenerAddress
go run main.go
```

### Add Piece to Proof Set

```bash
export PRIVATE_KEY=your-private-key-hex
export LISTENER_ADDRESS=0xYourListenerAddress
export PIECE_CID=baga6ea4seaqao7s73y24kcutaosvacpdjgfe5pw76ooefnyqw4ynr3d2y6x2mpq
go run main.go
```

### Use Custom RPC Endpoint

```bash
export PRIVATE_KEY=your-private-key-hex
export LISTENER_ADDRESS=0xYourListenerAddress
export RPC_URL=https://api.node.glif.io/rpc/v1
go run main.go
```

## Environment Variables

| Variable | Required | Description | Default |
|----------|----------|-------------|---------|
| `PRIVATE_KEY` | Yes | Your wallet private key (hex, without 0x prefix) | - |
| `LISTENER_ADDRESS` | Yes | Address of the record keeper contract | - |
| `RPC_URL` | No | Filecoin RPC endpoint | `https://api.calibration.node.glif.io/rpc/v1` |
| `PIECE_CID` | No | Piece CID to add to the proof set | - |

## Example Output

```
2026/01/27 20:00:00 Using default RPC URL: https://api.calibration.node.glif.io/rpc/v1
2026/01/27 20:00:00 Connected to calibration (Chain ID: 314159)
2026/01/27 20:00:00 Using address: 0xYourAddress
2026/01/27 20:00:00 Creating proof set...
2026/01/27 20:00:05 ✓ Proof set created successfully!
2026/01/27 20:00:05   Proof Set ID: 123
2026/01/27 20:00:05   Transaction Hash: 0xabc...
2026/01/27 20:00:05   Block Number: 1234567
2026/01/27 20:00:05   Gas Used: 250000

2026/01/27 20:00:05 Querying proof set details...
2026/01/27 20:00:05 ✓ Proof Set Details:
2026/01/27 20:00:05   ID: 123
2026/01/27 20:00:05   Live: true
2026/01/27 20:00:05   Listener: 0xListenerAddress
2026/01/27 20:00:05   Storage Provider: 0x000...
2026/01/27 20:00:05   Active Pieces: 0
2026/01/27 20:00:05   Leaf Count: 0
2026/01/27 20:00:05   Next Piece ID: 0

2026/01/27 20:00:05 ✓ Complete!
```

## What This Example Does

1. **Connects to Filecoin Network**: Establishes connection to RPC endpoint
2. **Detects Network**: Automatically identifies Mainnet or Calibration
3. **Creates Proof Set**: Deploys a new proof set on-chain
4. **Queries Details**: Retrieves and displays proof set information
5. **Adds Pieces (Optional)**: If PIECE_CID is provided, adds the piece to the proof set
6. **Shows Transaction Details**: Displays transaction hashes, block numbers, and gas usage

## Notes

- Creating a proof set requires gas fees (FIL)
- The listener address should be a deployed record keeper contract
- On Calibration testnet, you can get test FIL from the [faucet](https://faucet.calibration.fildev.network/)
- Transaction confirmation may take 30-60 seconds depending on network congestion
