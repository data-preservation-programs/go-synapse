# Contract ABIs

This directory contains the ABI (Application Binary Interface) files for Synapse smart contracts.

## PDPVerifier Contract

The PDPVerifier contract implements the Provable Data Possession (PDP) protocol for verifying data storage on Filecoin.

### Contract Addresses

- **Filecoin Mainnet (Chain ID: 314)**: `0xBADd0B92C1c71d02E7d520f64c0876538fa2557F`
  - [View on Filfox](https://filfox.info/en/address/0xBADd0B92C1c71d02E7d520f64c0876538fa2557F)

- **Filecoin Calibration Testnet (Chain ID: 314159)**: `0x85e366Cf9DD2c0aE37E963d9556F5f4718d6417C`
  - [View on Filscan](https://calibration.filscan.io/address/0x85e366Cf9DD2c0aE37E963d9556F5f4718d6417C)

### Key Functions

- `createDataSet`: Create a new data set for PDP verification
- `addPieces`: Add pieces to an existing data set
- `provePossession`: Submit proofs of possession for challenged pieces
- `schedulePieceDeletions`: Schedule pieces for deletion
- `deleteDataSet`: Delete an entire data set
- `getActivePieces`: Retrieve active pieces in a data set
- `getNextChallengeEpoch`: Get the next epoch when a challenge will be issued

### Generating Go Bindings

To generate Go bindings from this ABI:

```bash
abigen --abi=contracts/abi/PDPVerifier.json --pkg=contracts --type=PDPVerifier --out=contracts/pdpverifier.go
```
