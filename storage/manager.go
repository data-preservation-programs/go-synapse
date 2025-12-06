package storage

import (
	"bytes"
	"context"
	"crypto/rand"
	"fmt"
	"io"
	"math/big"
	"time"

	"github.com/data-preservation-programs/go-synapse/pdp"
	"github.com/data-preservation-programs/go-synapse/warmstorage"
	"github.com/ethereum/go-ethereum/common"
	"github.com/filecoin-project/go-commp-utils/v2/writer"
	"github.com/ipfs/go-cid"
)

const (
	pieceParkingTimeout    = 7 * time.Minute
	pieceAdditionTimeout   = 7 * time.Minute
	dataSetCreationTimeout = 7 * time.Minute
)

type DataSetInfoFetcher interface {
	GetDataSet(ctx context.Context, dataSetID int) (*warmstorage.DataSetInfo, error)
}

type Manager struct {
	clientAddress      common.Address
	warmStorageAddress common.Address
	authHelper         *pdp.AuthHelper
	pdpServer          *pdp.Server
	dataSetID          int
	clientDataSetID    *big.Int
	dataSetInfoFetcher DataSetInfoFetcher
	clientDataSetIDLoaded bool
}

type ManagerOption func(*Manager)

func WithDataSetInfoFetcher(fetcher DataSetInfoFetcher) ManagerOption {
	return func(m *Manager) {
		m.dataSetInfoFetcher = fetcher
	}
}

func WithClientDataSetID(clientDataSetID *big.Int) ManagerOption {
	return func(m *Manager) {
		m.clientDataSetID = clientDataSetID
		m.clientDataSetIDLoaded = true
	}
}

func NewManager(
	clientAddress common.Address,
	warmStorageAddress common.Address,
	authHelper *pdp.AuthHelper,
	pdpServer *pdp.Server,
	dataSetID int,
	opts ...ManagerOption,
) *Manager {
	m := &Manager{
		clientAddress:      clientAddress,
		warmStorageAddress: warmStorageAddress,
		authHelper:         authHelper,
		pdpServer:          pdpServer,
		dataSetID:          dataSetID,
		clientDataSetID:    big.NewInt(0),
	}
	for _, opt := range opts {
		opt(m)
	}
	return m
}

func (m *Manager) Upload(ctx context.Context, data io.Reader, opts *UploadOptions) (*UploadResult, error) {
	if opts == nil {
		opts = &UploadOptions{}
	}

	if opts.PieceCID != cid.Undef && opts.Size > 0 {
		return m.uploadStream(ctx, data, opts)
	}

	dataBytes, err := io.ReadAll(data)
	if err != nil {
		return nil, fmt.Errorf("failed to read data: %w", err)
	}

	return m.UploadBytes(ctx, dataBytes, opts)
}

func (m *Manager) UploadBytes(ctx context.Context, data []byte, opts *UploadOptions) (*UploadResult, error) {
	if opts == nil {
		opts = &UploadOptions{}
	}

	pieceCID := opts.PieceCID
	if pieceCID == cid.Undef {
		var err error
		pieceCID, err = CalculatePieceCID(data)
		if err != nil {
			return nil, fmt.Errorf("failed to calculate PieceCID: %w", err)
		}
	}

	if err := m.ensureDataSet(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure data set: %w", err)
	}

	_, err := m.pdpServer.UploadPiece(ctx, bytes.NewReader(data), int64(len(data)), pieceCID)
	if err != nil {
		return nil, fmt.Errorf("failed to upload piece: %w", err)
	}

	if err := m.pdpServer.WaitForPiece(ctx, pieceCID, pieceParkingTimeout); err != nil {
		return nil, fmt.Errorf("failed waiting for piece: %w", err)
	}

	pieceID, err := m.addPieceToDataSet(ctx, pieceCID, opts.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to add piece to data set: %w", err)
	}

	return &UploadResult{
		PieceCID:  pieceCID,
		Size:      int64(len(data)),
		PieceID:   pieceID,
		DataSetID: m.dataSetID,
	}, nil
}

func (m *Manager) uploadStream(ctx context.Context, data io.Reader, opts *UploadOptions) (*UploadResult, error) {
	if err := m.ensureDataSet(ctx); err != nil {
		return nil, fmt.Errorf("failed to ensure data set: %w", err)
	}

	_, err := m.pdpServer.UploadPiece(ctx, data, opts.Size, opts.PieceCID)
	if err != nil {
		return nil, fmt.Errorf("failed to upload piece: %w", err)
	}

	if err := m.pdpServer.WaitForPiece(ctx, opts.PieceCID, pieceParkingTimeout); err != nil {
		return nil, fmt.Errorf("failed waiting for piece: %w", err)
	}

	pieceID, err := m.addPieceToDataSet(ctx, opts.PieceCID, opts.Metadata)
	if err != nil {
		return nil, fmt.Errorf("failed to add piece to data set: %w", err)
	}

	return &UploadResult{
		PieceCID:  opts.PieceCID,
		Size:      opts.Size,
		PieceID:   pieceID,
		DataSetID: m.dataSetID,
	}, nil
}

func (m *Manager) Download(ctx context.Context, pieceCID cid.Cid, opts *DownloadOptions) ([]byte, error) {
	return m.pdpServer.DownloadPiece(ctx, pieceCID)
}

func (m *Manager) DataSetID() int {
	return m.dataSetID
}

func (m *Manager) ensureDataSet(ctx context.Context) error {
	if m.dataSetID != 0 {
		return m.ensureClientDataSetID(ctx)
	}

	m.clientDataSetID = randomBigInt()
	m.clientDataSetIDLoaded = true
	metadata := []pdp.MetadataEntry{}

	authSig, err := m.authHelper.SignCreateDataSet(m.clientDataSetID, m.authHelper.Address(), metadata)
	if err != nil {
		return fmt.Errorf("failed to sign create data set: %w", err)
	}

	extraData, err := pdp.EncodeDataSetCreateData(
		m.clientAddress,
		m.clientDataSetID,
		metadata,
		authSig.Signature,
	)
	if err != nil {
		return fmt.Errorf("failed to encode extra data: %w", err)
	}

	createResp, err := m.pdpServer.CreateDataSet(ctx, m.warmStorageAddress.Hex(), extraData)
	if err != nil {
		return fmt.Errorf("failed to create data set: %w", err)
	}

	status, err := m.pdpServer.WaitForDataSetCreation(ctx, createResp.TxHash, dataSetCreationTimeout)
	if err != nil {
		return fmt.Errorf("failed waiting for data set creation: %w", err)
	}

	if status.DataSetID == nil {
		return fmt.Errorf("data set created but no ID returned")
	}

	m.dataSetID = *status.DataSetID
	return nil
}

func (m *Manager) ensureClientDataSetID(ctx context.Context) error {
	if m.clientDataSetIDLoaded {
		return nil
	}

	if m.dataSetInfoFetcher == nil {
		return fmt.Errorf("cannot add pieces to existing dataset %d: no DataSetInfoFetcher configured (use WithDataSetInfoFetcher option)", m.dataSetID)
	}

	info, err := m.dataSetInfoFetcher.GetDataSet(ctx, m.dataSetID)
	if err != nil {
		return fmt.Errorf("failed to fetch dataset info for dataset %d: %w", m.dataSetID, err)
	}

	m.clientDataSetID = info.ClientDataSetID
	m.clientDataSetIDLoaded = true
	return nil
}

func (m *Manager) addPieceToDataSet(ctx context.Context, pieceCID cid.Cid, metadata map[string]string) (int, error) {
	var pieceMetadata []pdp.MetadataEntry
	for k, v := range metadata {
		pieceMetadata = append(pieceMetadata, pdp.MetadataEntry{Key: k, Value: v})
	}
	allMetadata := [][]pdp.MetadataEntry{pieceMetadata}

	nonce := randomBigInt()

	authSig, err := m.authHelper.SignAddPieces(m.clientDataSetID, nonce, []cid.Cid{pieceCID}, allMetadata)
	if err != nil {
		return 0, fmt.Errorf("failed to sign add pieces: %w", err)
	}

	extraData, err := pdp.EncodeAddPiecesExtraData(nonce, allMetadata, authSig.Signature)
	if err != nil {
		return 0, fmt.Errorf("failed to encode extra data: %w", err)
	}

	addResp, err := m.pdpServer.AddPieces(ctx, m.dataSetID, []cid.Cid{pieceCID}, extraData)
	if err != nil {
		return 0, fmt.Errorf("failed to add pieces: %w", err)
	}

	status, err := m.pdpServer.WaitForPieceAddition(ctx, m.dataSetID, addResp.TxHash, pieceAdditionTimeout)
	if err != nil {
		return 0, fmt.Errorf("failed waiting for piece addition: %w", err)
	}

	if len(status.ConfirmedPieceIDs) == 0 {
		return 0, fmt.Errorf("no piece IDs returned")
	}

	return status.ConfirmedPieceIDs[0], nil
}

func CalculatePieceCID(data []byte) (cid.Cid, error) {
	w := &writer.Writer{}

	_, err := w.Write(data)
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to write to CommP calculator: %w", err)
	}

	result, err := w.Sum()
	if err != nil {
		return cid.Undef, fmt.Errorf("failed to calculate CommP: %w", err)
	}

	return result.PieceCID, nil
}

func randomBigInt() *big.Int {
	b := make([]byte, 32)
	rand.Read(b)
	return new(big.Int).SetBytes(b)
}
