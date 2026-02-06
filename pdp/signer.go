package pdp

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

// Signer provides transaction signing for the Manager without exposing key material.
type Signer interface {
	Address() common.Address
	SignerFunc(chainID *big.Int) (bind.SignerFn, error)
}

// PrivateKeySigner is a simple signer backed by a local ECDSA private key.
type PrivateKeySigner struct {
	privateKey *ecdsa.PrivateKey
	address    common.Address
}

// NewPrivateKeySigner creates a signer backed by the provided private key.
func NewPrivateKeySigner(privateKey *ecdsa.PrivateKey) *PrivateKeySigner {
	return &PrivateKeySigner{
		privateKey: privateKey,
		address:    crypto.PubkeyToAddress(privateKey.PublicKey),
	}
}

// Address returns the signer address.
func (s *PrivateKeySigner) Address() common.Address {
	return s.address
}

// SignerFunc returns a bind.SignerFn using the provided chain ID.
func (s *PrivateKeySigner) SignerFunc(chainID *big.Int) (bind.SignerFn, error) {
	auth, err := bind.NewKeyedTransactorWithChainID(s.privateKey, chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create transactor: %w", err)
	}
	return auth.Signer, nil
}
