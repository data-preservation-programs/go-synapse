package pdp

import (
	"encoding/hex"
	"fmt"
	"math/big"
	"strings"

	"github.com/ethereum/go-ethereum/accounts/abi"
	"github.com/ethereum/go-ethereum/common"
)



var (
	addressType, _   = abi.NewType("address", "", nil)
	uint256Type, _   = abi.NewType("uint256", "", nil)
	stringArrayType, _ = abi.NewType("string[]", "", nil)
	stringArray2DType, _ = abi.NewType("string[][]", "", nil)
	bytesType, _     = abi.NewType("bytes", "", nil)
)


func EncodeDataSetCreateData(payer common.Address, clientDataSetID *big.Int, metadata []MetadataEntry, signature []byte) (string, error) {
	keys := make([]string, len(metadata))
	values := make([]string, len(metadata))
	for i, m := range metadata {
		keys[i] = m.Key
		values[i] = m.Value
	}

	args := abi.Arguments{
		{Type: addressType},
		{Type: uint256Type},
		{Type: stringArrayType},
		{Type: stringArrayType},
		{Type: bytesType},
	}

	encoded, err := args.Pack(payer, clientDataSetID, keys, values, signature)
	if err != nil {
		return "", fmt.Errorf("failed to encode data set create data: %w", err)
	}

	return "0x" + common.Bytes2Hex(encoded), nil
}


func EncodeAddPiecesExtraData(nonce *big.Int, metadata [][]MetadataEntry, signature []byte) (string, error) {
	keys := make([][]string, len(metadata))
	values := make([][]string, len(metadata))
	for i, pieceMetadata := range metadata {
		keys[i] = make([]string, len(pieceMetadata))
		values[i] = make([]string, len(pieceMetadata))
		for j, m := range pieceMetadata {
			keys[i][j] = m.Key
			values[i][j] = m.Value
		}
	}

	args := abi.Arguments{
		{Type: uint256Type},
		{Type: stringArray2DType},
		{Type: stringArray2DType},
		{Type: bytesType},
	}

	encoded, err := args.Pack(nonce, keys, values, signature)
	if err != nil {
		return "", fmt.Errorf("failed to encode add pieces extra data: %w", err)
	}

	return "0x" + common.Bytes2Hex(encoded), nil
}


func EncodeScheduleRemovalsExtraData(signature []byte) (string, error) {
	args := abi.Arguments{
		{Type: bytesType},
	}

	encoded, err := args.Pack(signature)
	if err != nil {
		return "", fmt.Errorf("failed to encode schedule removals extra data: %w", err)
	}

	return "0x" + common.Bytes2Hex(encoded), nil
}


// EncodeCreateDataSetAndAddPiecesExtraData wraps the two extraData blobs
// (from EncodeDataSetCreateData and EncodeAddPiecesExtraData) into the
// combined abi.encode(bytes,bytes) form Curio's /pdp/piece/pull expects
// when atomically creating a new data set and adding pieces in one shot.
// Inputs are hex strings (with or without 0x prefix), as produced by the
// sibling encoders in this file.
func EncodeCreateDataSetAndAddPiecesExtraData(createDataSetExtraHex, addPiecesExtraHex string) (string, error) {
	createBytes, err := decodeHex(createDataSetExtraHex)
	if err != nil {
		return "", fmt.Errorf("invalid createDataSet extra data: %w", err)
	}
	addPiecesBytes, err := decodeHex(addPiecesExtraHex)
	if err != nil {
		return "", fmt.Errorf("invalid addPieces extra data: %w", err)
	}

	args := abi.Arguments{
		{Type: bytesType},
		{Type: bytesType},
	}

	encoded, err := args.Pack(createBytes, addPiecesBytes)
	if err != nil {
		return "", fmt.Errorf("failed to encode create-and-add extra data: %w", err)
	}

	return "0x" + common.Bytes2Hex(encoded), nil
}

func decodeHex(s string) ([]byte, error) {
	return hex.DecodeString(strings.TrimPrefix(s, "0x"))
}
