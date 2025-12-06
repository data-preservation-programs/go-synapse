package pdp

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ipfs/go-cid"
)

var fixtures = struct {
	PrivateKey      string
	SignerAddress   string
	ContractAddress string
	ChainID         int64
	Signatures      struct {
		CreateDataSet struct {
			Signature       string
			ClientDataSetID int64
			Payee           string
			Metadata        []MetadataEntry
		}
		AddPieces struct {
			Signature       string
			ClientDataSetID int64
			Nonce           int64
			PieceCIDs       []string
			Metadata        [][]MetadataEntry
		}
		SchedulePieceRemovals struct {
			Signature       string
			ClientDataSetID int64
			PieceIDs        []int64
		}
		DeleteDataSet struct {
			Signature       string
			ClientDataSetID int64
		}
	}
}{
	PrivateKey:      "1234567890123456789012345678901234567890123456789012345678901234",
	SignerAddress:   "0x2e988A386a799F506693793c6A5AF6B54dfAaBfB",
	ContractAddress: "0x5615dEB798BB3E4dFa0139dFa1b3D433Cc23b72f",
	ChainID:         31337,
	Signatures: struct {
		CreateDataSet struct {
			Signature       string
			ClientDataSetID int64
			Payee           string
			Metadata        []MetadataEntry
		}
		AddPieces struct {
			Signature       string
			ClientDataSetID int64
			Nonce           int64
			PieceCIDs       []string
			Metadata        [][]MetadataEntry
		}
		SchedulePieceRemovals struct {
			Signature       string
			ClientDataSetID int64
			PieceIDs        []int64
		}
		DeleteDataSet struct {
			Signature       string
			ClientDataSetID int64
		}
	}{
		CreateDataSet: struct {
			Signature       string
			ClientDataSetID int64
			Payee           string
			Metadata        []MetadataEntry
		}{
			Signature:       "c77965e2b6efd594629c44eb61127bc3133b65d08c25f8aa33e3021e7f46435845ab67ffbac96afc4b4671ecbd32d4869ca7fe1c0eaa5affa942d0abbfd98d601b",
			ClientDataSetID: 12345,
			Payee:           "0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
			Metadata:        []MetadataEntry{{Key: "title", Value: "TestDataSet"}},
		},
		AddPieces: struct {
			Signature       string
			ClientDataSetID int64
			Nonce           int64
			PieceCIDs       []string
			Metadata        [][]MetadataEntry
		}{
			Signature:       "1f09427806dc1e4c073a9fd7345fdd1919973abe3f3021594964134887c964d82e7b242019c79b21a8fa40331d14b59219b431846e4cdc08adb2e9470e7273161c",
			ClientDataSetID: 12345,
			Nonce:           1,
			PieceCIDs: []string{
				"bafkzcibcauan42av3szurbbscwuu3zjssvfwbpsvbjf6y3tukvlgl2nf5rha6pa",
				"bafkzcibcpybwiktap34inmaex4wbs6cghlq5i2j2yd2bb2zndn5ep7ralzphkdy",
			},
			Metadata: [][]MetadataEntry{{}, {}},
		},
		SchedulePieceRemovals: struct {
			Signature       string
			ClientDataSetID int64
			PieceIDs        []int64
		}{
			Signature:       "cb8e645f2894fde89de54d4a54eb1e0d9871901c6fa1c2ee8a0390dc3a29e6cb2244d0561e3eca6452fa59efaab3d4b18a0b5b59ab52e233b3469422556ae9c61c",
			ClientDataSetID: 12345,
			PieceIDs:        []int64{1, 3, 5},
		},
		DeleteDataSet: struct {
			Signature       string
			ClientDataSetID int64
		}{
			Signature:       "94e366bd2f9bfc933a87575126715bccf128b77d9c6937e194023e13b54272eb7a74b7e6e26acf4341d9c56e141ff7ba154c37ea03e9c35b126fff1efe1a0c831c",
			ClientDataSetID: 12345,
		},
	},
}

func setupAuthHelper(t *testing.T) *AuthHelper {
	privateKeyBytes, err := hex.DecodeString(fixtures.PrivateKey)
	if err != nil {
		t.Fatalf("Failed to decode private key: %v", err)
	}

	privateKey, err := crypto.ToECDSA(privateKeyBytes)
	if err != nil {
		t.Fatalf("Failed to parse private key: %v", err)
	}

	address := crypto.PubkeyToAddress(privateKey.PublicKey)
	expectedAddress := common.HexToAddress(fixtures.SignerAddress)
	if address != expectedAddress {
		t.Fatalf("Derived address %s does not match expected %s", address.Hex(), expectedAddress.Hex())
	}

	contractAddr := common.HexToAddress(fixtures.ContractAddress)
	chainID := big.NewInt(fixtures.ChainID)

	return NewAuthHelper(privateKey, contractAddr, chainID)
}

func TestAuthHelper_SignCreateDataSet(t *testing.T) {
	authHelper := setupAuthHelper(t)

	result, err := authHelper.SignCreateDataSet(
		big.NewInt(fixtures.Signatures.CreateDataSet.ClientDataSetID),
		common.HexToAddress(fixtures.Signatures.CreateDataSet.Payee),
		fixtures.Signatures.CreateDataSet.Metadata,
	)
	if err != nil {
		t.Fatalf("SignCreateDataSet failed: %v", err)
	}

	expectedSig := fixtures.Signatures.CreateDataSet.Signature
	actualSig := hex.EncodeToString(result.Signature)

	if actualSig != expectedSig {
		t.Errorf("Signature mismatch:\nExpected: %s\nActual:   %s", expectedSig, actualSig)
	}

	sigForRecovery := make([]byte, len(result.Signature))
	copy(sigForRecovery, result.Signature)
	if sigForRecovery[64] >= 27 {
		sigForRecovery[64] -= 27
	}
	pubKey, err := crypto.SigToPub(result.SignedData.Bytes(), sigForRecovery)
	if err != nil {
		t.Fatalf("Failed to recover public key: %v", err)
	}
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	expectedAddr := common.HexToAddress(fixtures.SignerAddress)
	if !strings.EqualFold(recoveredAddr.Hex(), expectedAddr.Hex()) {
		t.Errorf("Recovered address %s does not match expected %s", recoveredAddr.Hex(), expectedAddr.Hex())
	}
}

func TestAuthHelper_SignSchedulePieceRemovals(t *testing.T) {
	authHelper := setupAuthHelper(t)

	pieceIDs := make([]*big.Int, len(fixtures.Signatures.SchedulePieceRemovals.PieceIDs))
	for i, id := range fixtures.Signatures.SchedulePieceRemovals.PieceIDs {
		pieceIDs[i] = big.NewInt(id)
	}

	result, err := authHelper.SignSchedulePieceRemovals(
		big.NewInt(fixtures.Signatures.SchedulePieceRemovals.ClientDataSetID),
		pieceIDs,
	)
	if err != nil {
		t.Fatalf("SignSchedulePieceRemovals failed: %v", err)
	}

	expectedSig := fixtures.Signatures.SchedulePieceRemovals.Signature
	actualSig := hex.EncodeToString(result.Signature)

	if actualSig != expectedSig {
		t.Errorf("Signature mismatch:\nExpected: %s\nActual:   %s", expectedSig, actualSig)
	}

	sigForRecovery := make([]byte, len(result.Signature))
	copy(sigForRecovery, result.Signature)
	if sigForRecovery[64] >= 27 {
		sigForRecovery[64] -= 27
	}
	pubKey, err := crypto.SigToPub(result.SignedData.Bytes(), sigForRecovery)
	if err != nil {
		t.Fatalf("Failed to recover public key: %v", err)
	}
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	expectedAddr := common.HexToAddress(fixtures.SignerAddress)
	if !strings.EqualFold(recoveredAddr.Hex(), expectedAddr.Hex()) {
		t.Errorf("Recovered address %s does not match expected %s", recoveredAddr.Hex(), expectedAddr.Hex())
	}
}

func TestAuthHelper_SignDeleteDataSet(t *testing.T) {
	authHelper := setupAuthHelper(t)

	result, err := authHelper.SignDeleteDataSet(
		big.NewInt(fixtures.Signatures.DeleteDataSet.ClientDataSetID),
	)
	if err != nil {
		t.Fatalf("SignDeleteDataSet failed: %v", err)
	}

	expectedSig := fixtures.Signatures.DeleteDataSet.Signature
	actualSig := hex.EncodeToString(result.Signature)

	if actualSig != expectedSig {
		t.Errorf("Signature mismatch:\nExpected: %s\nActual:   %s", expectedSig, actualSig)
	}

	sigForRecovery := make([]byte, len(result.Signature))
	copy(sigForRecovery, result.Signature)
	if sigForRecovery[64] >= 27 {
		sigForRecovery[64] -= 27
	}
	pubKey, err := crypto.SigToPub(result.SignedData.Bytes(), sigForRecovery)
	if err != nil {
		t.Fatalf("Failed to recover public key: %v", err)
	}
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	expectedAddr := common.HexToAddress(fixtures.SignerAddress)
	if !strings.EqualFold(recoveredAddr.Hex(), expectedAddr.Hex()) {
		t.Errorf("Recovered address %s does not match expected %s", recoveredAddr.Hex(), expectedAddr.Hex())
	}
}

func TestAuthHelper_SignAddPieces(t *testing.T) {
	authHelper := setupAuthHelper(t)

	pieceCIDs := make([]cid.Cid, len(fixtures.Signatures.AddPieces.PieceCIDs))
	for i, cidStr := range fixtures.Signatures.AddPieces.PieceCIDs {
		c, err := cid.Decode(cidStr)
		if err != nil {
			t.Fatalf("Failed to parse PieceCID %s: %v", cidStr, err)
		}
		pieceCIDs[i] = c
	}

	result, err := authHelper.SignAddPieces(
		big.NewInt(fixtures.Signatures.AddPieces.ClientDataSetID),
		big.NewInt(fixtures.Signatures.AddPieces.Nonce),
		pieceCIDs,
		fixtures.Signatures.AddPieces.Metadata,
	)
	if err != nil {
		t.Fatalf("SignAddPieces failed: %v", err)
	}

	expectedSig := fixtures.Signatures.AddPieces.Signature
	actualSig := hex.EncodeToString(result.Signature)

	if actualSig != expectedSig {
		t.Errorf("Signature mismatch:\nExpected: %s\nActual:   %s", expectedSig, actualSig)
	}

	sigForRecovery := make([]byte, len(result.Signature))
	copy(sigForRecovery, result.Signature)
	if sigForRecovery[64] >= 27 {
		sigForRecovery[64] -= 27
	}
	pubKey, err := crypto.SigToPub(result.SignedData.Bytes(), sigForRecovery)
	if err != nil {
		t.Fatalf("Failed to recover public key: %v", err)
	}
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	expectedAddr := common.HexToAddress(fixtures.SignerAddress)
	if !strings.EqualFold(recoveredAddr.Hex(), expectedAddr.Hex()) {
		t.Errorf("Recovered address %s does not match expected %s", recoveredAddr.Hex(), expectedAddr.Hex())
	}
}

func TestAuthHelper_ConsistentSignatures(t *testing.T) {
	authHelper := setupAuthHelper(t)

	sig1, err := authHelper.SignCreateDataSet(
		big.NewInt(fixtures.Signatures.CreateDataSet.ClientDataSetID),
		common.HexToAddress(fixtures.Signatures.CreateDataSet.Payee),
		fixtures.Signatures.CreateDataSet.Metadata,
	)
	if err != nil {
		t.Fatalf("First SignCreateDataSet failed: %v", err)
	}

	sig2, err := authHelper.SignCreateDataSet(
		big.NewInt(fixtures.Signatures.CreateDataSet.ClientDataSetID),
		common.HexToAddress(fixtures.Signatures.CreateDataSet.Payee),
		fixtures.Signatures.CreateDataSet.Metadata,
	)
	if err != nil {
		t.Fatalf("Second SignCreateDataSet failed: %v", err)
	}

	if hex.EncodeToString(sig1.Signature) != hex.EncodeToString(sig2.Signature) {
		t.Error("Signatures are not deterministic")
	}

	if sig1.SignedData != sig2.SignedData {
		t.Error("Signed data is not deterministic")
	}
}

func TestAuthHelper_EmptyPieceData(t *testing.T) {
	authHelper := setupAuthHelper(t)

	result, err := authHelper.SignAddPieces(
		big.NewInt(fixtures.Signatures.AddPieces.ClientDataSetID),
		big.NewInt(fixtures.Signatures.AddPieces.Nonce),
		[]cid.Cid{}, // empty array
		[][]MetadataEntry{},
	)
	if err != nil {
		t.Fatalf("SignAddPieces with empty array failed: %v", err)
	}

	if len(result.Signature) != 65 {
		t.Errorf("Expected 65 byte signature, got %d bytes", len(result.Signature))
	}

	sigForRecovery := make([]byte, len(result.Signature))
	copy(sigForRecovery, result.Signature)
	if sigForRecovery[64] >= 27 {
		sigForRecovery[64] -= 27
	}
	pubKey, err := crypto.SigToPub(result.SignedData.Bytes(), sigForRecovery)
	if err != nil {
		t.Fatalf("Failed to recover public key: %v", err)
	}
	recoveredAddr := crypto.PubkeyToAddress(*pubKey)
	expectedAddr := common.HexToAddress(fixtures.SignerAddress)
	if !strings.EqualFold(recoveredAddr.Hex(), expectedAddr.Hex()) {
		t.Errorf("Recovered address %s does not match expected %s", recoveredAddr.Hex(), expectedAddr.Hex())
	}
}

func TestAuthHelper_Address(t *testing.T) {
	authHelper := setupAuthHelper(t)

	expectedAddr := common.HexToAddress(fixtures.SignerAddress)
	if authHelper.Address() != expectedAddr {
		t.Errorf("Address() returned %s, expected %s", authHelper.Address().Hex(), expectedAddr.Hex())
	}
}
