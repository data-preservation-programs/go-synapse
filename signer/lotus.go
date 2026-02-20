package signer

import "fmt"

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
