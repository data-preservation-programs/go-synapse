package pdp

import (
	"encoding/hex"
	"math/big"
	"strings"
	"testing"

	"github.com/ethereum/go-ethereum/accounts/abi"
)

func TestEncodeCreateDataSetAndAddPiecesExtraData(t *testing.T) {
	t.Run("round-trips through abi.Unpack", func(t *testing.T) {
		create := []byte{0xde, 0xad, 0xbe, 0xef}
		add := []byte{0xfe, 0xed, 0xfa, 0xce, 0xca, 0xfe}

		out, err := EncodeCreateDataSetAndAddPiecesExtraData(
			"0x"+hex.EncodeToString(create),
			"0x"+hex.EncodeToString(add),
		)
		if err != nil {
			t.Fatalf("encode failed: %v", err)
		}
		if !strings.HasPrefix(out, "0x") {
			t.Fatalf("output missing 0x prefix: %s", out)
		}

		raw, err := hex.DecodeString(out[2:])
		if err != nil {
			t.Fatalf("decode output: %v", err)
		}

		args := abi.Arguments{
			{Type: bytesType},
			{Type: bytesType},
		}
		unpacked, err := args.Unpack(raw)
		if err != nil {
			t.Fatalf("unpack: %v", err)
		}
		if len(unpacked) != 2 {
			t.Fatalf("expected 2 fields, got %d", len(unpacked))
		}
		gotCreate, ok := unpacked[0].([]byte)
		if !ok {
			t.Fatalf("first field not []byte: %T", unpacked[0])
		}
		gotAdd, ok := unpacked[1].([]byte)
		if !ok {
			t.Fatalf("second field not []byte: %T", unpacked[1])
		}
		if string(gotCreate) != string(create) {
			t.Errorf("createDataSet round-trip mismatch: got %x, want %x", gotCreate, create)
		}
		if string(gotAdd) != string(add) {
			t.Errorf("addPieces round-trip mismatch: got %x, want %x", gotAdd, add)
		}
	})

	t.Run("accepts inputs without 0x prefix", func(t *testing.T) {
		_, err := EncodeCreateDataSetAndAddPiecesExtraData("deadbeef", "feedface")
		if err != nil {
			t.Fatalf("expected no-prefix inputs to work, got %v", err)
		}
	})

	t.Run("rejects non-hex input", func(t *testing.T) {
		_, err := EncodeCreateDataSetAndAddPiecesExtraData("0xnothex!", "0xdeadbeef")
		if err == nil {
			t.Error("expected error on non-hex createDataSet input")
		}
		_, err = EncodeCreateDataSetAndAddPiecesExtraData("0xdeadbeef", "0xnothex!")
		if err == nil {
			t.Error("expected error on non-hex addPieces input")
		}
	})

	t.Run("round-trips real CreateDataSet+AddPieces extras", func(t *testing.T) {
		// produce a CreateDataSet extraData and an AddPieces extraData via
		// the sibling encoders, then wrap. Verifies that the canonical caller
		// path (sign -> encode each -> wrap combined) round-trips cleanly.
		auth := testAuthHelper(t)
		clientDataSetID := big.NewInt(1)
		payee := auth.Address()

		createSig, err := auth.SignCreateDataSet(clientDataSetID, payee, nil)
		if err != nil {
			t.Fatalf("sign create: %v", err)
		}
		createExtra, err := EncodeDataSetCreateData(payee, clientDataSetID, nil, createSig.Signature)
		if err != nil {
			t.Fatalf("encode create extra: %v", err)
		}

		nonce := big.NewInt(42)
		addSig, err := auth.SignAddPieces(clientDataSetID, nonce, nil, nil)
		if err != nil {
			t.Fatalf("sign add: %v", err)
		}
		addExtra, err := EncodeAddPiecesExtraData(nonce, nil, addSig.Signature)
		if err != nil {
			t.Fatalf("encode add extra: %v", err)
		}

		combined, err := EncodeCreateDataSetAndAddPiecesExtraData(createExtra, addExtra)
		if err != nil {
			t.Fatalf("combine: %v", err)
		}
		if !strings.HasPrefix(combined, "0x") {
			t.Fatalf("combined missing 0x prefix: %s", combined)
		}

		raw, err := hex.DecodeString(combined[2:])
		if err != nil {
			t.Fatalf("decode combined: %v", err)
		}
		args := abi.Arguments{
			{Type: bytesType},
			{Type: bytesType},
		}
		unpacked, err := args.Unpack(raw)
		if err != nil {
			t.Fatalf("unpack combined: %v", err)
		}
		gotCreateHex := "0x" + hex.EncodeToString(unpacked[0].([]byte))
		gotAddHex := "0x" + hex.EncodeToString(unpacked[1].([]byte))
		if gotCreateHex != createExtra {
			t.Errorf("createDataSet extra mismatch")
		}
		if gotAddHex != addExtra {
			t.Errorf("addPieces extra mismatch")
		}
	})
}
