package signer

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/filecoin-project/go-address"
)

func makeTestLotusExport(keyType string, raw []byte) string {
	ki := lotusKeyInfo{Type: keyType, PrivateKey: raw}
	j, _ := json.Marshal(ki)
	return hex.EncodeToString(j)
}

func TestSecp256k1Signer_DualProtocol(t *testing.T) {
	key, err := ethcrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	s, err := NewSecp256k1SignerFromECDSA(key)
	if err != nil {
		t.Fatal(err)
	}

	// filecoin address should be secp256k1 protocol
	filAddr := s.FilecoinAddress()
	if filAddr.Protocol() != address.SECP256K1 {
		t.Errorf("expected secp256k1 address, got protocol %d", filAddr.Protocol())
	}

	// evm address should match go-ethereum derivation
	expectedEth := ethcrypto.PubkeyToAddress(key.PublicKey)
	if s.EVMAddress() != expectedEth {
		t.Errorf("EVMAddress() = %s, want %s", s.EVMAddress(), expectedEth)
	}

	// sign a filecoin message
	msg := []byte("test message")
	sig, err := s.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}
	if sig.Type != 1 { // SigTypeSecp256k1
		t.Errorf("signature type = %d, want 1", sig.Type)
	}
	if len(sig.Data) != 65 {
		t.Errorf("signature length = %d, want 65", len(sig.Data))
	}

	// create an evm transactor
	opts, err := s.Transactor(big.NewInt(314159))
	if err != nil {
		t.Fatal(err)
	}
	if opts.From != expectedEth {
		t.Errorf("Transactor.From = %s, want %s", opts.From, expectedEth)
	}
}

func TestSecp256k1Signer_FromLotusExport(t *testing.T) {
	key, err := ethcrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	exported := makeTestLotusExport("secp256k1", ethcrypto.FromECDSA(key))

	s, err := NewSecp256k1SignerFromLotusExport(exported)
	if err != nil {
		t.Fatal(err)
	}

	expectedEth := ethcrypto.PubkeyToAddress(key.PublicKey)
	if s.EVMAddress() != expectedEth {
		t.Errorf("EVMAddress() = %s, want %s", s.EVMAddress(), expectedEth)
	}
}

func TestSecp256k1Signer_RejectsWrongType(t *testing.T) {
	exported := makeTestLotusExport("bls", []byte("dummy"))
	_, err := NewSecp256k1SignerFromLotusExport(exported)
	if err == nil {
		t.Error("expected error for bls key, got nil")
	}
}

func TestSecp256k1Signer_PadsShortKeys(t *testing.T) {
	// simulate big.Int.Bytes() dropping a leading zero
	key, err := ethcrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	raw := ethcrypto.FromECDSA(key)

	// construct from canonical 32-byte form
	s1, err := NewSecp256k1Signer(raw)
	if err != nil {
		t.Fatal(err)
	}

	// construct from potentially short D.Bytes()
	s2, err := NewSecp256k1Signer(key.D.Bytes())
	if err != nil {
		t.Fatal(err)
	}

	if s1.EVMAddress() != s2.EVMAddress() {
		t.Errorf("addresses differ: %s vs %s", s1.EVMAddress(), s2.EVMAddress())
	}
	if s1.FilecoinAddress() != s2.FilecoinAddress() {
		t.Errorf("filecoin addresses differ: %s vs %s", s1.FilecoinAddress(), s2.FilecoinAddress())
	}
}

func TestFromLotusExport_AutoDetect(t *testing.T) {
	key, err := ethcrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}
	exported := makeTestLotusExport("secp256k1", ethcrypto.FromECDSA(key))

	s, err := FromLotusExport(exported)
	if err != nil {
		t.Fatal(err)
	}

	if s.FilecoinAddress().Protocol() != address.SECP256K1 {
		t.Error("expected secp256k1 signer")
	}

	evm, ok := AsEVM(s)
	if !ok {
		t.Fatal("secp256k1 signer should satisfy EVMSigner")
	}
	if evm.EVMAddress() == (common.Address{}) {
		t.Error("EVMAddress should not be zero")
	}
}

func TestFromLotusExport_UnsupportedType(t *testing.T) {
	exported := makeTestLotusExport("ed25519", []byte("dummy"))
	_, err := FromLotusExport(exported)
	if err == nil {
		t.Error("expected error for unsupported key type")
	}
}

func TestAsEVM(t *testing.T) {
	key, err := ethcrypto.GenerateKey()
	if err != nil {
		t.Fatal(err)
	}

	s, err := NewSecp256k1SignerFromECDSA(key)
	if err != nil {
		t.Fatal(err)
	}

	evm, ok := AsEVM(s)
	if !ok {
		t.Fatal("secp256k1 signer should be EVMSigner")
	}
	if evm.EVMAddress() == (common.Address{}) {
		t.Error("EVMAddress should not be zero")
	}
}
