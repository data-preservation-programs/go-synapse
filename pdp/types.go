package pdp

import (
	"github.com/ethereum/go-ethereum/common"
	"github.com/ipfs/go-cid"
)

type MetadataEntry struct {
	Key   string `json:"key"`
	Value string `json:"value"`
}

type AuthSignature struct {
	Signature  []byte
	V          uint8
	R          [32]byte
	S          [32]byte
	SignedData common.Hash
}

type CreateDataSetRequest struct {
	RecordKeeper string `json:"recordKeeper"`
	ExtraData    string `json:"extraData"`
}

type CreateDataSetResponse struct {
	TxHash    string `json:"txHash"`
	StatusURL string `json:"statusUrl"`
}

type DataSetCreationStatus struct {
	CreateMessageHash string `json:"createMessageHash"`
	DataSetCreated    bool   `json:"dataSetCreated"`
	Service           string `json:"service"`
	TxStatus          string `json:"txStatus"`
	OK                *bool  `json:"ok"`
	DataSetID         *int   `json:"dataSetId,omitempty"`
}

type AddPiecesRequest struct {
	Pieces    []PieceData `json:"pieces"`
	ExtraData string      `json:"extraData"`
}

type PieceData struct {
	PieceCID  string         `json:"pieceCid"`
	SubPieces []SubPieceData `json:"subPieces"`
}

type SubPieceData struct {
	SubPieceCID string `json:"subPieceCid"`
}

type AddPiecesResponse struct {
	Message   string `json:"message"`
	TxHash    string `json:"txHash"`
	StatusURL string `json:"statusUrl"`
}

type PieceAdditionStatus struct {
	TxHash            string `json:"txHash"`
	TxStatus          string `json:"txStatus"`
	DataSetID         int    `json:"dataSetId"`
	PieceCount        int    `json:"pieceCount"`
	AddMessageOK      *bool  `json:"addMessageOk"`
	ConfirmedPieceIDs []int  `json:"confirmedPieceIds,omitempty"`
}

type UploadPieceResponse struct {
	PieceCID cid.Cid
	Size     int64
}

type FindPieceResponse struct {
	PieceCID cid.Cid
}

type DataSetData struct {
	ID                 int         `json:"id"`
	Pieces             []PieceInfo `json:"pieces"`
	NextChallengeEpoch int64       `json:"nextChallengeEpoch"`
}

type PieceInfo struct {
	PieceID        int     `json:"pieceId"`
	PieceCID       cid.Cid `json:"pieceCid"`
	SubPieceCID    cid.Cid `json:"subPieceCid"`
	SubPieceOffset int64   `json:"subPieceOffset"`
}

type PieceStatus struct {
	PieceCID    string  `json:"pieceCid"`
	Status      string  `json:"status"`
	Indexed     bool    `json:"indexed"`
	Advertised  bool    `json:"advertised"`
	Retrieved   bool    `json:"retrieved"`
	RetrievedAt *string `json:"retrievedAt,omitempty"`
}

type CreateAndAddRequest struct {
	RecordKeeper string      `json:"recordKeeper"`
	Pieces       []PieceData `json:"pieces"`
	ExtraData    string      `json:"extraData"`
}

type UploadStartResponse struct {
	UploadUUID string `json:"uploadUuid"`
}

type UploadCompleteResponse struct {
	PieceCID string `json:"pieceCid"`
	Size     int64  `json:"size"`
}
