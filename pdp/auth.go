package pdp

import (
	"crypto/ecdsa"
	"fmt"
	"math/big"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/common/math"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ethereum/go-ethereum/signer/core/apitypes"
	"github.com/ipfs/go-cid"
)

type AuthHelper struct {
	privateKey         *ecdsa.PrivateKey
	address            common.Address
	warmStorageAddress common.Address
	chainID            *big.Int
	domain             apitypes.TypedDataDomain
}

func NewAuthHelper(privateKey *ecdsa.PrivateKey, warmStorageAddr common.Address, chainID *big.Int) *AuthHelper {
	address := crypto.PubkeyToAddress(privateKey.PublicKey)

	return &AuthHelper{
		privateKey:         privateKey,
		address:            address,
		warmStorageAddress: warmStorageAddr,
		chainID:            chainID,
		domain: apitypes.TypedDataDomain{
			Name:              "FilecoinWarmStorageService",
			Version:           "1",
			ChainId:           (*math.HexOrDecimal256)(chainID),
			VerifyingContract: warmStorageAddr.Hex(),
		},
	}
}

func (a *AuthHelper) Address() common.Address {
	return a.address
}

var eip712Types = apitypes.Types{
	"EIP712Domain": {
		{Name: "name", Type: "string"},
		{Name: "version", Type: "string"},
		{Name: "chainId", Type: "uint256"},
		{Name: "verifyingContract", Type: "address"},
	},
	"MetadataEntry": {
		{Name: "key", Type: "string"},
		{Name: "value", Type: "string"},
	},
	"CreateDataSet": {
		{Name: "clientDataSetId", Type: "uint256"},
		{Name: "payee", Type: "address"},
		{Name: "metadata", Type: "MetadataEntry[]"},
	},
	"Cid": {
		{Name: "data", Type: "bytes"},
	},
	"PieceMetadata": {
		{Name: "pieceIndex", Type: "uint256"},
		{Name: "metadata", Type: "MetadataEntry[]"},
	},
	"AddPieces": {
		{Name: "clientDataSetId", Type: "uint256"},
		{Name: "nonce", Type: "uint256"},
		{Name: "pieceData", Type: "Cid[]"},
		{Name: "pieceMetadata", Type: "PieceMetadata[]"},
	},
	"SchedulePieceRemovals": {
		{Name: "clientDataSetId", Type: "uint256"},
		{Name: "pieceIds", Type: "uint256[]"},
	},
	"DeleteDataSet": {
		{Name: "clientDataSetId", Type: "uint256"},
	},
}

func (a *AuthHelper) SignCreateDataSet(clientDataSetID *big.Int, payee common.Address, metadata []MetadataEntry) (*AuthSignature, error) {
	metadataArray := make([]interface{}, len(metadata))
	for i, m := range metadata {
		metadataArray[i] = map[string]interface{}{
			"key":   m.Key,
			"value": m.Value,
		}
	}

	message := apitypes.TypedDataMessage{
		"clientDataSetId": (*math.HexOrDecimal256)(clientDataSetID),
		"payee":           payee.Hex(),
		"metadata":        metadataArray,
	}

	return a.signTypedData("CreateDataSet", message)
}

func (a *AuthHelper) SignAddPieces(clientDataSetID, nonce *big.Int, pieceCIDs []cid.Cid, metadata [][]MetadataEntry) (*AuthSignature, error) {
	if len(metadata) == 0 {
		metadata = make([][]MetadataEntry, len(pieceCIDs))
		for i := range metadata {
			metadata[i] = []MetadataEntry{}
		}
	}
	if len(metadata) != len(pieceCIDs) {
		return nil, fmt.Errorf("metadata length (%d) must match pieceCIDs length (%d)", len(metadata), len(pieceCIDs))
	}

	pieceData := make([]interface{}, len(pieceCIDs))
	for i, c := range pieceCIDs {
		pieceData[i] = map[string]interface{}{
			"data": c.Bytes(),
		}
	}

	pieceMetadata := make([]interface{}, len(pieceCIDs))
	for i, meta := range metadata {
		metadataArray := make([]interface{}, len(meta))
		for j, m := range meta {
			metadataArray[j] = map[string]interface{}{
				"key":   m.Key,
				"value": m.Value,
			}
		}
		pieceMetadata[i] = map[string]interface{}{
			"pieceIndex": (*math.HexOrDecimal256)(big.NewInt(int64(i))),
			"metadata":   metadataArray,
		}
	}

	message := apitypes.TypedDataMessage{
		"clientDataSetId": (*math.HexOrDecimal256)(clientDataSetID),
		"nonce":           (*math.HexOrDecimal256)(nonce),
		"pieceData":       pieceData,
		"pieceMetadata":   pieceMetadata,
	}

	return a.signTypedData("AddPieces", message)
}

func (a *AuthHelper) SignSchedulePieceRemovals(clientDataSetID *big.Int, pieceIDs []*big.Int) (*AuthSignature, error) {
	pieceIDsArray := make([]interface{}, len(pieceIDs))
	for i, id := range pieceIDs {
		pieceIDsArray[i] = (*math.HexOrDecimal256)(id)
	}

	message := apitypes.TypedDataMessage{
		"clientDataSetId": (*math.HexOrDecimal256)(clientDataSetID),
		"pieceIds":        pieceIDsArray,
	}

	return a.signTypedData("SchedulePieceRemovals", message)
}

func (a *AuthHelper) SignDeleteDataSet(clientDataSetID *big.Int) (*AuthSignature, error) {
	message := apitypes.TypedDataMessage{
		"clientDataSetId": (*math.HexOrDecimal256)(clientDataSetID),
	}

	return a.signTypedData("DeleteDataSet", message)
}

func (a *AuthHelper) signTypedData(primaryType string, message apitypes.TypedDataMessage) (*AuthSignature, error) {
	typedData := apitypes.TypedData{
		Types:       eip712Types,
		PrimaryType: primaryType,
		Domain:      a.domain,
		Message:     message,
	}

	domainSeparator, err := typedData.HashStruct("EIP712Domain", typedData.Domain.Map())
	if err != nil {
		return nil, fmt.Errorf("failed to hash domain: %w", err)
	}

	messageHash, err := typedData.HashStruct(primaryType, message)
	if err != nil {
		return nil, fmt.Errorf("failed to hash message: %w", err)
	}

	rawData := []byte{0x19, 0x01}
	rawData = append(rawData, domainSeparator...)
	rawData = append(rawData, messageHash...)
	signedData := crypto.Keccak256Hash(rawData)

	signature, err := crypto.Sign(signedData.Bytes(), a.privateKey)
	if err != nil {
		return nil, fmt.Errorf("failed to sign: %w", err)
	}

	if signature[64] < 27 {
		signature[64] += 27
	}

	var r, s [32]byte
	copy(r[:], signature[:32])
	copy(s[:], signature[32:64])

	return &AuthSignature{
		Signature:  signature,
		V:          signature[64],
		R:          r,
		S:          s,
		SignedData: signedData,
	}, nil
}
