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

// PullStatus is the wire-level status enum for a pull request or an
// individual piece within one. Aggregate (response-level) status reflects
// the worst case across all pieces: failed > retrying > inProgress >
// pending > complete.
type PullStatus string

const (
	PullStatusPending    PullStatus = "pending"
	PullStatusInProgress PullStatus = "inProgress"
	PullStatusRetrying   PullStatus = "retrying"
	PullStatusComplete   PullStatus = "complete"
	PullStatusFailed     PullStatus = "failed"
)

// PullPieceInput names a piece and the source URL Curio should fetch it
// from. SourceURL must be HTTPS and end in /piece/{pieceCid}.
type PullPieceInput struct {
	PieceCID  string `json:"pieceCid"`
	SourceURL string `json:"sourceUrl"`
}

// PullPieceStatus is per-piece status in a PullPiecesResponse.
type PullPieceStatus struct {
	PieceCID string     `json:"pieceCid"`
	Status   PullStatus `json:"status"`
}

// PullPiecesRequest is the JSON body Curio expects on POST /pdp/piece/pull.
// DataSetID is omitted from the wire when nil; callers signal
// "create a new set" by passing nil and supplying a combined
// CreateDataSet+AddPieces extraData (see EncodeCreateDataSetAndAddPiecesExtraData).
type PullPiecesRequest struct {
	ExtraData    string           `json:"extraData"`
	RecordKeeper string           `json:"recordKeeper"`
	Pieces       []PullPieceInput `json:"pieces"`
	DataSetID    *uint64          `json:"dataSetId,omitempty"`
}

// PullPiecesResponse is the JSON body Curio returns from POST /pdp/piece/pull.
type PullPiecesResponse struct {
	Status PullStatus        `json:"status"`
	Pieces []PullPieceStatus `json:"pieces"`
}

// PullPiecesOptions is the higher-level argument for Server.PullPieces and
// Server.WaitForPullPieces. DataSetID == 0 signals "create a new set"
// (the request will omit the field on the wire).
type PullPiecesOptions struct {
	// RecordKeeper is the FWSS contract address (hex). Required.
	RecordKeeper string
	// Pieces lists the pieces to pull and their source URLs.
	Pieces []PullPieceInput
	// ExtraData is the EIP-712-signed authorization blob. For new sets,
	// build it with EncodeCreateDataSetAndAddPiecesExtraData. For
	// existing sets, EncodeAddPiecesExtraData alone.
	ExtraData string
	// DataSetID is the target set ID, or 0 to create a new set atomically.
	DataSetID uint64
}

// ManagerConfig holds configuration options for the Manager
type ManagerConfig struct {
	// GasBufferPercent is the percentage buffer to add to gas estimates (0-100)
	// For example, 10 means add 10% to the estimated gas limit.
	// Ignored when DefaultGasLimit is set.
	GasBufferPercent int
	// DefaultGasLimit, when non-zero, is used for all transactions instead
	// of estimating gas. FEVM gas estimation is unreliable for payable
	// calls and for contracts that do cross-actor calls, so callers should
	// set this when targeting FEVM.
	DefaultGasLimit uint64
	// ContractAddress overrides the default PDPVerifier contract address for the network.
	// Leave zero to use the network default.
	ContractAddress common.Address
}

// DefaultManagerConfig returns the default configuration for Manager
func DefaultManagerConfig() ManagerConfig {
	return ManagerConfig{
		GasBufferPercent: 10, // Default 10% buffer
	}
}
