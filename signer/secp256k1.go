package signer

import (
	"crypto/ecdsa"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/crypto"

	blake2b "github.com/minio/blake2b-simd"

	dcrdecdsa "github.com/decred/dcrd/dcrec/secp256k1/v4/ecdsa"
	dcrdsecp "github.com/decred/dcrd/dcrec/secp256k1/v4"
)

// Secp256k1Signer implements EVMSigner backed by a secp256k1 private key.
// It can sign both Filecoin messages and Ethereum transactions.
type Secp256k1Signer struct {
	raw     []byte          // raw 32-byte scalar
	ecdsaKey *ecdsa.PrivateKey
	filAddr  address.Address
	ethAddr  common.Address
}

// NewSecp256k1Signer creates a dual-protocol signer from raw key bytes.
// The input is left-padded to 32 bytes if shorter (e.g. from big.Int.Bytes()).
func NewSecp256k1Signer(raw []byte) (*Secp256k1Signer, error) {
	if len(raw) == 0 || len(raw) > 32 {
		return nil, fmt.Errorf("invalid key length: %d", len(raw))
	}

	// left-pad to 32 bytes — big.Int.Bytes() drops leading zeros
	var padded [32]byte
	copy(padded[32-len(raw):], raw)

	ecdsaKey, err := ethcrypto.ToECDSA(padded[:])
	if err != nil {
		return nil, fmt.Errorf("invalid secp256k1 key: %w", err)
	}

	return newFromECDSA(ecdsaKey, padded[:])
}

// NewSecp256k1SignerFromECDSA creates a dual-protocol signer from a go-ethereum
// ECDSA private key. This is the preferred constructor when you already have
// a parsed key (e.g. from crypto.GenerateKey or crypto.HexToECDSA).
func NewSecp256k1SignerFromECDSA(key *ecdsa.PrivateKey) (*Secp256k1Signer, error) {
	if key == nil {
		return nil, fmt.Errorf("nil private key")
	}
	raw := ethcrypto.FromECDSA(key) // always 32 bytes, zero-padded
	return newFromECDSA(key, raw)
}

func newFromECDSA(ecdsaKey *ecdsa.PrivateKey, raw []byte) (*Secp256k1Signer, error) {
	dcrdKey := dcrdsecp.PrivKeyFromBytes(raw)
	uncompressed := dcrdKey.PubKey().SerializeUncompressed()
	filAddr, err := address.NewSecp256k1Address(uncompressed)
	if err != nil {
		return nil, fmt.Errorf("deriving filecoin address: %w", err)
	}

	ethAddr := ethcrypto.PubkeyToAddress(ecdsaKey.PublicKey)

	return &Secp256k1Signer{
		raw:      raw,
		ecdsaKey: ecdsaKey,
		filAddr:  filAddr,
		ethAddr:  ethAddr,
	}, nil
}

// NewSecp256k1SignerFromLotusExport creates a signer from a lotus-exported
// private key (hex-encoded JSON with Type and PrivateKey fields).
// This is the format produced by `lotus wallet export`.
func NewSecp256k1SignerFromLotusExport(exported string) (*Secp256k1Signer, error) {
	ki, err := decodeLotusKey(exported)
	if err != nil {
		return nil, err
	}
	if ki.Type != "secp256k1" {
		return nil, fmt.Errorf("expected secp256k1 key, got %s", ki.Type)
	}
	return NewSecp256k1Signer(ki.PrivateKey)
}

func (s *Secp256k1Signer) FilecoinAddress() address.Address {
	return s.filAddr
}

// Sign produces a Filecoin-native signature (blake2b-256 hash, R|S|V format).
func (s *Secp256k1Signer) Sign(msg []byte) (*crypto.Signature, error) {
	hash := blake2b.Sum256(msg)
	dcrdKey := dcrdsecp.PrivKeyFromBytes(s.raw)
	sig := dcrdecdsa.SignCompact(dcrdKey, hash[:], false)

	// rotate: go from V|R|S to R|S|V, adjust recovery ID
	recoveryID := sig[0]
	copy(sig, sig[1:])
	sig[64] = recoveryID - 27

	return &crypto.Signature{
		Type: crypto.SigTypeSecp256k1,
		Data: sig,
	}, nil
}

func (s *Secp256k1Signer) EVMAddress() common.Address {
	return s.ethAddr
}

// Transactor returns bind.TransactOpts for signing Ethereum/FEVM transactions.
func (s *Secp256k1Signer) Transactor(chainID *big.Int) (*bind.TransactOpts, error) {
	return bind.NewKeyedTransactorWithChainID(s.ecdsaKey, chainID)
}

// lotusKeyInfo mirrors the JSON structure of a lotus wallet export.
type lotusKeyInfo struct {
	Type       string `json:"Type"`
	PrivateKey []byte `json:"PrivateKey"`
}

func decodeLotusKey(exported string) (*lotusKeyInfo, error) {
	raw, err := hex.DecodeString(exported)
	if err != nil {
		return nil, fmt.Errorf("decoding hex: %w", err)
	}
	var ki lotusKeyInfo
	if err := json.Unmarshal(raw, &ki); err != nil {
		return nil, fmt.Errorf("unmarshaling key: %w", err)
	}
	return &ki, nil
}
