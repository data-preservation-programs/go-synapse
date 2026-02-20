package signer

import (
	"fmt"

	"github.com/filecoin-project/go-address"
	"github.com/filecoin-project/go-state-types/crypto"
	blst "github.com/supranational/blst/bindings/go"
)

// BLSSigner implements Signer (but not EVMSigner) backed by a BLS private key.
type BLSSigner struct {
	raw     []byte
	sk      *blst.SecretKey
	filAddr address.Address
}

// NewBLSSigner creates a Filecoin-only signer from raw BLS key bytes.
func NewBLSSigner(raw []byte) (*BLSSigner, error) {
	sk := new(blst.SecretKey).Deserialize(raw)
	if sk == nil {
		return nil, fmt.Errorf("invalid BLS secret key")
	}

	pk := new(blst.P1Affine).From(sk).Compress()
	filAddr, err := address.NewBLSAddress(pk)
	if err != nil {
		return nil, fmt.Errorf("deriving BLS address: %w", err)
	}

	return &BLSSigner{
		raw:     raw,
		sk:      sk,
		filAddr: filAddr,
	}, nil
}

// NewBLSSignerFromLotusExport creates a signer from a lotus-exported BLS key.
func NewBLSSignerFromLotusExport(exported string) (*BLSSigner, error) {
	ki, err := decodeLotusKey(exported)
	if err != nil {
		return nil, err
	}
	if ki.Type != "bls" {
		return nil, fmt.Errorf("expected bls key, got %s", ki.Type)
	}
	return NewBLSSigner(ki.PrivateKey)
}

func (s *BLSSigner) FilecoinAddress() address.Address {
	return s.filAddr
}

// Sign produces a BLS signature over the raw message bytes (no prehash).
func (s *BLSSigner) Sign(msg []byte) (*crypto.Signature, error) {
	sig := new(blst.P2Affine).Sign(s.sk, msg, []byte("BLS_SIG_BLS12381G2_XMD:SHA-256_SSWU_RO_NUL_"))
	return &crypto.Signature{
		Type: crypto.SigTypeBLS,
		Data: sig.Compress(),
	}, nil
}
