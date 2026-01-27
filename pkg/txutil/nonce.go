package txutil

import (
	"context"
	"fmt"
	"math/big"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// NonceManager manages nonces for transaction sending
type NonceManager struct {
	client     *ethclient.Client
	address    common.Address
	mu         sync.Mutex
	nonce      *uint64
	pendingTxs map[uint64]bool
}

// NewNonceManager creates a new nonce manager
func NewNonceManager(client *ethclient.Client, address common.Address) *NonceManager {
	return &NonceManager{
		client:     client,
		address:    address,
		pendingTxs: make(map[uint64]bool),
	}
}

// GetNonce returns the next available nonce
func (nm *NonceManager) GetNonce(ctx context.Context) (uint64, error) {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	if nm.nonce == nil {
		nonce, err := nm.client.PendingNonceAt(ctx, nm.address)
		if err != nil {
			return 0, fmt.Errorf("failed to get pending nonce: %w", err)
		}
		nm.nonce = &nonce
	}

	currentNonce := *nm.nonce
	nm.pendingTxs[currentNonce] = true
	*nm.nonce++

	return currentNonce, nil
}

// MarkConfirmed marks a nonce as confirmed (transaction mined)
func (nm *NonceManager) MarkConfirmed(nonce uint64) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	delete(nm.pendingTxs, nonce)
}

// Reset resets the nonce manager (fetches fresh nonce from network)
func (nm *NonceManager) Reset(ctx context.Context) error {
	nm.mu.Lock()
	defer nm.mu.Unlock()

	nonce, err := nm.client.PendingNonceAt(ctx, nm.address)
	if err != nil {
		return fmt.Errorf("failed to reset nonce: %w", err)
	}

	nm.nonce = &nonce
	nm.pendingTxs = make(map[uint64]bool)
	return nil
}

// GetPendingCount returns the number of pending transactions
func (nm *NonceManager) GetPendingCount() int {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	return len(nm.pendingTxs)
}

// GetFreshNonce gets a fresh nonce from the network without caching
func GetFreshNonce(ctx context.Context, client *ethclient.Client, address common.Address) (uint64, error) {
	nonce, err := client.PendingNonceAt(ctx, address)
	if err != nil {
		return 0, fmt.Errorf("failed to get nonce: %w", err)
	}
	return nonce, nil
}

// GetChainID returns the chain ID from the client
func GetChainID(ctx context.Context, client *ethclient.Client) (*big.Int, error) {
	chainID, err := client.ChainID(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to get chain ID: %w", err)
	}
	return chainID, nil
}
