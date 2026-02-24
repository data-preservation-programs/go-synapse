package signer

import (
	"encoding/hex"
	"encoding/json"
	"fmt"
)

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

// FromLotusExport creates a Signer from a lotus-exported private key string.
// The key type (secp256k1 or bls) is detected automatically.
// For secp256k1 keys, the returned Signer also implements EVMSigner.
func FromLotusExport(exported string) (Signer, error) {
	ki, err := decodeLotusKey(exported)
	if err != nil {
		return nil, err
	}
	switch ki.Type {
	case "secp256k1":
		return NewSecp256k1Signer(ki.PrivateKey)
	case "bls":
		return NewBLSSigner(ki.PrivateKey)
	default:
		return nil, fmt.Errorf("unsupported key type: %s", ki.Type)
	}
}

// AsEVM checks whether a Signer can sign EVM transactions.
// Returns nil, false for BLS keys.
func AsEVM(s Signer) (EVMSigner, bool) {
	e, ok := s.(EVMSigner)
	return e, ok
}
