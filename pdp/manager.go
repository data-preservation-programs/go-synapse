package pdp

import (
	"context"
	"errors"
	"fmt"
	"math/big"

	"github.com/data-preservation-programs/go-synapse/constants"
	"github.com/data-preservation-programs/go-synapse/contracts"
	"github.com/data-preservation-programs/go-synapse/pkg/txutil"
	"github.com/ethereum/go-ethereum/accounts/abi/bind"
	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/core/types"
	"github.com/ethereum/go-ethereum/ethclient"
	"github.com/ipfs/go-cid"
)

// ProofSetManager provides high-level operations for managing PDP proof sets
type ProofSetManager interface {
	// CreateProofSet creates a new proof set on-chain
	CreateProofSet(ctx context.Context, opts CreateProofSetOptions) (*ProofSetResult, error)

	// GetProofSet retrieves proof set details
	GetProofSet(ctx context.Context, proofSetID *big.Int) (*ProofSet, error)

	// AddRoots adds data roots to an existing proof set
	AddRoots(ctx context.Context, proofSetID *big.Int, roots []Root) (*AddRootsResult, error)

	// GetRoots retrieves roots from a proof set with pagination
	GetRoots(ctx context.Context, proofSetID *big.Int, offset, limit uint64) ([]Root, bool, error)

	// DeleteProofSet removes a proof set
	DeleteProofSet(ctx context.Context, proofSetID *big.Int, extraData []byte) error

	// GetNextChallengeEpoch gets the next challenge epoch for a proof set
	GetNextChallengeEpoch(ctx context.Context, proofSetID *big.Int) (uint64, error)

	// DataSetLive checks if a proof set is live
	DataSetLive(ctx context.Context, proofSetID *big.Int) (bool, error)
}

// CreateProofSetOptions options for creating a proof set
type CreateProofSetOptions struct {
	Listener  common.Address
	ExtraData []byte
	Value     *big.Int // Optional payment value
}

// ProofSetResult result of creating a proof set
type ProofSetResult struct {
	ProofSetID      *big.Int
	TransactionHash common.Hash
	Receipt         *types.Receipt
}

// ProofSet represents a proof set's details
type ProofSet struct {
	ID              *big.Int
	Listener        common.Address
	StorageProvider common.Address
	LeafCount       uint64
	ActivePieces    uint64
	NextPieceID     uint64
	Live            bool
}

// Root represents a data root
type Root struct {
	PieceCID cid.Cid
	PieceID  uint64
}

// AddRootsResult result of adding roots
type AddRootsResult struct {
	TransactionHash common.Hash
	Receipt         *types.Receipt
	RootsAdded      int
	PieceIDs        []uint64
}

// Manager implements ProofSetManager.
type Manager struct {
	client       *ethclient.Client
	signer       Signer
	address      common.Address
	contract     *contracts.PDPVerifier
	contractAddr common.Address
	chainID      *big.Int
	nonceManager *txutil.NonceManager
	config       ManagerConfig
}

// NewManagerWithContext creates a new ProofSetManager with context support and default configuration.
func NewManagerWithContext(ctx context.Context, client *ethclient.Client, signer Signer, network constants.Network) (*Manager, error) {
	return NewManagerWithConfig(ctx, client, signer, network, nil)
}

// NewManagerWithConfig creates a new ProofSetManager with custom configuration.
// If config is nil, default configuration will be used.
func NewManagerWithConfig(ctx context.Context, client *ethclient.Client, signer Signer, network constants.Network, config *ManagerConfig) (*Manager, error) {
	if signer == nil {
		return nil, errors.New("signer is required")
	}

	// Validate chain ID matches expected network
	expectedChainID, ok := constants.ExpectedChainID(network)
	if !ok {
		return nil, fmt.Errorf("unknown network: %v", network)
	}
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}

	if chainID.Int64() != expectedChainID {
		return nil, fmt.Errorf("chain ID mismatch: RPC returned %d but network %s expects %d", chainID.Int64(), network, expectedChainID)
	}

	// Use default config if none provided
	if config == nil {
		cfg := DefaultManagerConfig()
		config = &cfg
	}

	// Validate configuration
	if config.GasBufferPercent < 0 || config.GasBufferPercent > 100 {
		return nil, fmt.Errorf("gas buffer percent must be between 0 and 100, got %d", config.GasBufferPercent)
	}

	contractAddr := config.ContractAddress
	if contractAddr == (common.Address{}) {
		contractAddr = constants.GetPDPVerifierAddress(network)
		if contractAddr == (common.Address{}) {
			return nil, fmt.Errorf("no PDPVerifier address for network %v", network)
		}
	}

	contract, err := contracts.NewPDPVerifier(contractAddr, client)
	if err != nil {
		return nil, fmt.Errorf("failed to create contract instance: %w", err)
	}

	address := signer.Address()
	nonceManager := txutil.NewNonceManager(client, address)

	return &Manager{
		client:       client,
		signer:       signer,
		address:      address,
		contract:     contract,
		contractAddr: contractAddr,
		chainID:      chainID,
		nonceManager: nonceManager,
		config:       *config,
	}, nil
}

func (m *Manager) newTransactor(ctx context.Context, nonce uint64, value *big.Int) (*bind.TransactOpts, error) {
	signerFn, err := m.signer.SignerFunc(m.chainID)
	if err != nil {
		return nil, fmt.Errorf("failed to create signer: %w", err)
	}

	auth := &bind.TransactOpts{
		From:    m.address,
		Signer:  signerFn,
		Nonce:   big.NewInt(int64(nonce)),
		Context: ctx,
	}
	if value != nil {
		auth.Value = value
	}
	return auth, nil
}

// CreateProofSet creates a new proof set on-chain
func (m *Manager) CreateProofSet(ctx context.Context, opts CreateProofSetOptions) (*ProofSetResult, error) {
	nonce, err := m.nonceManager.GetNonce(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Track whether transaction was actually sent to the network
	txSent := false
	defer func() {
		if !txSent {
			// Local failure before sending - release nonce immediately
			m.nonceManager.MarkFailed(nonce)
		}
	}()

	auth, err := m.newTransactor(ctx, nonce, opts.Value)
	if err != nil {
		return nil, err
	}

	// Estimate gas
	auth.NoSend = true
	tx, err := m.contract.CreateDataSet(auth, opts.Listener, opts.ExtraData)
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas for createDataSet: %w", err)
	}
	// Apply configurable gas buffer
	bufferMultiplier := 1.0 + (float64(m.config.GasBufferPercent) / 100.0)
	auth.GasLimit = uint64(float64(tx.Gas()) * bufferMultiplier)
	auth.NoSend = false

	tx, err = m.contract.CreateDataSet(auth, opts.Listener, opts.ExtraData)
	if err != nil {
		// txSent is still false - defer will call MarkFailed
		return nil, fmt.Errorf("failed to create data set: %w", err)
	}
	// Mark as sent only after successful contract call
	txSent = true

	receipt, err := txutil.WaitForReceipt(ctx, m.client, tx.Hash(), txutil.DefaultRetryConfig().MaxBackoff*3)
	if err != nil {
		// Error waiting for receipt - transaction may be pending, don't release nonce
		return nil, fmt.Errorf("failed to wait for receipt: %w", err)
	}

	m.nonceManager.MarkConfirmed(nonce)

	// Extract proof set ID from logs
	proofSetID, err := m.extractProofSetIDFromReceipt(receipt)
	if err != nil {
		return nil, fmt.Errorf("failed to extract proof set ID: %w", err)
	}

	return &ProofSetResult{
		ProofSetID:      proofSetID,
		TransactionHash: tx.Hash(),
		Receipt:         receipt,
	}, nil
}

// GetProofSet retrieves proof set details
func (m *Manager) GetProofSet(ctx context.Context, proofSetID *big.Int) (*ProofSet, error) {
	opts := &bind.CallOpts{Context: ctx}

	live, err := m.contract.DataSetLive(opts, proofSetID)
	if err != nil {
		return nil, fmt.Errorf("failed to check if data set is live: %w", err)
	}

	listener, err := m.contract.GetDataSetListener(opts, proofSetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get listener: %w", err)
	}

	sp, _, err := m.contract.GetDataSetStorageProvider(opts, proofSetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get storage provider: %w", err)
	}

	leafCount, err := m.contract.GetDataSetLeafCount(opts, proofSetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get leaf count: %w", err)
	}

	activePieces, err := m.contract.GetActivePieceCount(opts, proofSetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get active piece count: %w", err)
	}

	nextPieceID, err := m.contract.GetNextPieceId(opts, proofSetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get next piece ID: %w", err)
	}

	return &ProofSet{
		ID:              proofSetID,
		Listener:        listener,
		StorageProvider: sp,
		LeafCount:       leafCount.Uint64(),
		ActivePieces:    activePieces.Uint64(),
		NextPieceID:     nextPieceID.Uint64(),
		Live:            live,
	}, nil
}

// AddRoots adds data roots to an existing proof set
func (m *Manager) AddRoots(ctx context.Context, proofSetID *big.Int, roots []Root) (*AddRootsResult, error) {
	if len(roots) == 0 {
		return nil, errors.New("no roots provided")
	}

	// Get the proof set's listener address
	proofSet, err := m.GetProofSet(ctx, proofSetID)
	if err != nil {
		return nil, fmt.Errorf("failed to get proof set: %w", err)
	}
	listenerAddr := proofSet.Listener

	// Convert roots to contract format
	pieceData := make([]contracts.CidsCid, len(roots))
	for i, root := range roots {
		pieceData[i] = contracts.CidsCid{
			Data: root.PieceCID.Bytes(),
		}
	}

	nonce, err := m.nonceManager.GetNonce(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get nonce: %w", err)
	}

	// Track whether transaction was actually sent to the network
	txSent := false
	defer func() {
		if !txSent {
			// Local failure before sending - release nonce immediately
			m.nonceManager.MarkFailed(nonce)
		}
	}()

	auth, err := m.newTransactor(ctx, nonce, nil)
	if err != nil {
		return nil, err
	}

	// Estimate gas
	auth.NoSend = true
	tx, err := m.contract.AddPieces(auth, proofSetID, listenerAddr, pieceData, []byte{})
	if err != nil {
		return nil, fmt.Errorf("failed to estimate gas for addPieces: %w", err)
	}
	// Apply configurable gas buffer
	bufferMultiplier := 1.0 + (float64(m.config.GasBufferPercent) / 100.0)
	auth.GasLimit = uint64(float64(tx.Gas()) * bufferMultiplier)
	auth.NoSend = false

	tx, err = m.contract.AddPieces(auth, proofSetID, listenerAddr, pieceData, []byte{})
	if err != nil {
		// txSent is still false - defer will call MarkFailed
		return nil, fmt.Errorf("failed to add pieces: %w", err)
	}
	// Mark as sent only after successful contract call
	txSent = true

	receipt, err := txutil.WaitForReceipt(ctx, m.client, tx.Hash(), txutil.DefaultRetryConfig().MaxBackoff*3)
	if err != nil {
		// Error waiting for receipt - transaction may be pending, don't release nonce
		return nil, fmt.Errorf("failed to wait for receipt: %w", err)
	}

	m.nonceManager.MarkConfirmed(nonce)

	// Extract piece IDs from logs
	pieceIDs, err := m.extractPieceIDsFromReceipt(receipt)
	if err != nil {
		return nil, fmt.Errorf("failed to extract piece IDs: %w", err)
	}

	return &AddRootsResult{
		TransactionHash: tx.Hash(),
		Receipt:         receipt,
		RootsAdded:      len(roots),
		PieceIDs:        pieceIDs,
	}, nil
}

// GetRoots retrieves roots from a proof set with pagination
func (m *Manager) GetRoots(ctx context.Context, proofSetID *big.Int, offset, limit uint64) ([]Root, bool, error) {
	opts := &bind.CallOpts{Context: ctx}

	result, err := m.contract.GetActivePieces(opts, proofSetID, big.NewInt(int64(offset)), big.NewInt(int64(limit)))
	if err != nil {
		return nil, false, fmt.Errorf("failed to get active pieces: %w", err)
	}

	roots := make([]Root, len(result.Pieces))
	for i, piece := range result.Pieces {
		c, err := cid.Cast(piece.Data)
		if err != nil {
			return nil, false, fmt.Errorf("failed to parse piece CID at index %d: %w", i, err)
		}

		var pieceID uint64
		if i < len(result.PieceIds) {
			pieceID = result.PieceIds[i].Uint64()
		}

		roots[i] = Root{
			PieceCID: c,
			PieceID:  pieceID,
		}
	}

	return roots, result.HasMore, nil
}

// DeleteProofSet removes a proof set
func (m *Manager) DeleteProofSet(ctx context.Context, proofSetID *big.Int, extraData []byte) error {
	nonce, err := m.nonceManager.GetNonce(ctx)
	if err != nil {
		return fmt.Errorf("failed to get nonce: %w", err)
	}

	// Track whether transaction was actually sent to the network
	txSent := false
	defer func() {
		if !txSent {
			// Local failure before sending - release nonce immediately
			m.nonceManager.MarkFailed(nonce)
		}
	}()

	auth, err := m.newTransactor(ctx, nonce, nil)
	if err != nil {
		return err
	}

	tx, err := m.contract.DeleteDataSet(auth, proofSetID, extraData)
	if err != nil {
		// txSent is still false - defer will call MarkFailed
		return fmt.Errorf("failed to delete data set: %w", err)
	}
	// Mark as sent only after successful contract call
	txSent = true

	_, err = txutil.WaitForReceipt(ctx, m.client, tx.Hash(), txutil.DefaultRetryConfig().MaxBackoff*3)
	if err != nil {
		// Error waiting for receipt - transaction may be pending, don't release nonce
		return fmt.Errorf("failed to wait for receipt: %w", err)
	}

	m.nonceManager.MarkConfirmed(nonce)
	return nil
}

// GetNextChallengeEpoch gets the next challenge epoch for a proof set
func (m *Manager) GetNextChallengeEpoch(ctx context.Context, proofSetID *big.Int) (uint64, error) {
	opts := &bind.CallOpts{Context: ctx}

	epoch, err := m.contract.GetNextChallengeEpoch(opts, proofSetID)
	if err != nil {
		return 0, fmt.Errorf("failed to get next challenge epoch: %w", err)
	}

	return epoch.Uint64(), nil
}

// DataSetLive checks if a proof set is live
func (m *Manager) DataSetLive(ctx context.Context, proofSetID *big.Int) (bool, error) {
	opts := &bind.CallOpts{Context: ctx}

	live, err := m.contract.DataSetLive(opts, proofSetID)
	if err != nil {
		return false, fmt.Errorf("failed to check if data set is live: %w", err)
	}

	return live, nil
}

// extractProofSetIDFromReceipt extracts the proof set ID from transaction receipt logs
func (m *Manager) extractProofSetIDFromReceipt(receipt *types.Receipt) (*big.Int, error) {
	for _, log := range receipt.Logs {
		event, err := m.contract.ParseDataSetCreated(*log)
		if err == nil && event != nil {
			return event.SetId, nil
		}
	}
	return nil, errors.New("DataSetCreated event not found in receipt")
}

// extractPieceIDsFromReceipt extracts piece IDs from transaction receipt logs
func (m *Manager) extractPieceIDsFromReceipt(receipt *types.Receipt) ([]uint64, error) {
	for _, log := range receipt.Logs {
		event, err := m.contract.ParsePiecesAdded(*log)
		if err == nil && event != nil {
			pieceIDs := make([]uint64, len(event.PieceIds))
			for i, id := range event.PieceIds {
				pieceIDs[i] = id.Uint64()
			}
			return pieceIDs, nil
		}
	}
	return nil, errors.New("PiecesAdded event not found in receipt")
}
