package pdp

import (
	"crypto/ecdsa"

	"github.com/data-preservation-programs/go-synapse/signer"
)

// Signer is kept for backward compatibility. New code should use signer.EVMSigner.
type Signer = signer.EVMSigner

// PrivateKeySigner is kept for backward compatibility.
type PrivateKeySigner = signer.Secp256k1Signer

// NewPrivateKeySigner creates a dual-protocol signer from an ECDSA private key.
// Kept for backward compatibility — new code should use
// signer.NewSecp256k1SignerFromECDSA or signer.NewSecp256k1Signer.
func NewPrivateKeySigner(privateKey *ecdsa.PrivateKey) *PrivateKeySigner {
	s, err := signer.NewSecp256k1SignerFromECDSA(privateKey)
	if err != nil {
		// should never happen for a valid *ecdsa.PrivateKey
		panic("NewPrivateKeySigner: " + err.Error())
	}
	return s
}
