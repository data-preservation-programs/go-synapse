package signer

import (
	"encoding/hex"
	"encoding/json"
	"math/big"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	ethcrypto "github.com/ethereum/go-ethereum/crypto"
	"github.com/filecoin-project/go-address"
	blst "github.com/supranational/blst/bindings/go"
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

	// SignDigest should produce a 65-byte signature recoverable to the same address
	digest := ethcrypto.Keccak256([]byte("test digest input"))
	sigBytes, err := s.SignDigest(digest)
	if err != nil {
		t.Fatalf("SignDigest: %v", err)
	}
	if len(sigBytes) != 65 {
		t.Errorf("SignDigest length = %d, want 65", len(sigBytes))
	}
	if sigBytes[64] != 0 && sigBytes[64] != 1 {
		t.Errorf("SignDigest V = %d, want 0 or 1", sigBytes[64])
	}
	recovered, err := ethcrypto.SigToPub(digest, sigBytes)
	if err != nil {
		t.Fatalf("SigToPub: %v", err)
	}
	if ethcrypto.PubkeyToAddress(*recovered) != expectedEth {
		t.Errorf("recovered address %s != signer %s", ethcrypto.PubkeyToAddress(*recovered), expectedEth)
	}

	// SignDigest rejects non-32-byte input
	if _, err := s.SignDigest([]byte("short")); err == nil {
		t.Error("SignDigest should reject non-32-byte input")
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

func TestBLSSigner_Sign(t *testing.T) {
	// generate a valid BLS secret key via blst
	var ikm [32]byte
	// deterministic seed for reproducible test
	copy(ikm[:], []byte("test-bls-key-seed-for-unit-test!"))

	sk := blst.KeyGen(ikm[:])
	if sk == nil {
		t.Fatal("failed to generate BLS key")
	}
	raw := sk.Serialize()

	s, err := NewBLSSigner(raw)
	if err != nil {
		t.Fatal(err)
	}

	if s.FilecoinAddress().Protocol() != address.BLS {
		t.Errorf("expected BLS address, got protocol %d", s.FilecoinAddress().Protocol())
	}

	msg := []byte("test message")
	sig, err := s.Sign(msg)
	if err != nil {
		t.Fatal(err)
	}
	if sig.Type != 2 { // SigTypeBLS
		t.Errorf("signature type = %d, want 2", sig.Type)
	}
	if len(sig.Data) != 96 { // compressed G2 point
		t.Errorf("signature length = %d, want 96", len(sig.Data))
	}
}

func TestBLSSigner_NotEVM(t *testing.T) {
	var ikm [32]byte
	copy(ikm[:], []byte("test-bls-key-seed-for-unit-test!"))
	sk := blst.KeyGen(ikm[:])
	raw := sk.Serialize()

	s, err := NewBLSSigner(raw)
	if err != nil {
		t.Fatal(err)
	}

	_, ok := AsEVM(s)
	if ok {
		t.Error("BLS signer should not satisfy EVMSigner")
	}
}

func TestBLSSigner_FromLotusExport(t *testing.T) {
	var ikm [32]byte
	copy(ikm[:], []byte("test-bls-key-seed-for-unit-test!"))
	sk := blst.KeyGen(ikm[:])
	raw := sk.Serialize()

	exported := makeTestLotusExport("bls", raw)
	s, err := NewBLSSignerFromLotusExport(exported)
	if err != nil {
		t.Fatal(err)
	}

	if s.FilecoinAddress().Protocol() != address.BLS {
		t.Errorf("expected BLS address, got protocol %d", s.FilecoinAddress().Protocol())
	}
}

func TestBLSSigner_RejectsWrongType(t *testing.T) {
	exported := makeTestLotusExport("secp256k1", []byte("dummy"))
	_, err := NewBLSSignerFromLotusExport(exported)
	if err == nil {
		t.Error("expected error for secp256k1 key, got nil")
	}
}
