// Package signer provides dual-protocol signing for Filecoin and Ethereum/FEVM.
//
// A secp256k1 private key can sign both native Filecoin messages (blake2b)
// and Ethereum transactions (keccak256). This package exposes both capabilities
// through composable interfaces, eliminating the need for adapter glue in
// every tool that touches both protocols.
//
// BLS keys can only sign Filecoin messages. Attempting to use a BLS key
// for EVM operations returns an error.
package signer

import (
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/crypto"
)

// Signer signs native Filecoin messages. Every key type can do this.
type Signer interface {
	FilecoinAddress() address.Address
	Sign(msg []byte) (*crypto.Signature, error)
}

// EVMSigner signs Ethereum/FEVM transactions and EIP-712 typed data.
// Only secp256k1 keys can do this.
//
// SignDigest produces a 65-byte recoverable secp256k1 signature over a
// 32-byte keccak digest, in [R || S || V] form with V = 0 or 1 (the
// go-ethereum crypto.Sign convention). Callers that need on-chain
// ECDSA recovery (e.g. PDP extraData) must normalize V to 27/28
// themselves; the digest-signer interface keeps the raw recovery ID.
type EVMSigner interface {
	Signer
	EVMAddress() common.Address
	Transactor(chainID *big.Int) (*bind.TransactOpts, error)
	SignDigest(digest []byte) ([]byte, error)
}
