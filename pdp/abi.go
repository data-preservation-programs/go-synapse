package pdp

import (
	"fmt"
	"math/big"

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
