package txutil

import (
	"context"
	"fmt"
	"sync"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

// NonceManager allocates and tracks transaction nonces for a single sender.
type NonceManager struct {
	client     *ethclient.Client
	address    common.Address
	mu         sync.Mutex
	nonce      *uint64
	pendingTxs map[uint64]bool
}

func NewNonceManager(client *ethclient.Client, address common.Address) *NonceManager {
	return &NonceManager{
		client:     client,
		address:    address,
		pendingTxs: make(map[uint64]bool),
	}
}

// GetNonce returns the next available nonce, fetching from the network on
// first call (or after MarkFailed clears the cache).
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

func (nm *NonceManager) MarkConfirmed(nonce uint64) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	delete(nm.pendingTxs, nonce)
}

// MarkFailed releases a nonce that was never successfully sent to the network.
// Call only for local failures before send (gas estimation, signing); not for
// network errors after sending, since those tx may still be pending in the
// mempool. The cached nonce is cleared so the next GetNonce refreshes from
// the network.
func (nm *NonceManager) MarkFailed(nonce uint64) {
	nm.mu.Lock()
	defer nm.mu.Unlock()
	delete(nm.pendingTxs, nonce)
	nm.nonce = nil
}
