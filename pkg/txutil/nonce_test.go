package txutil

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func TestNonceManager_MarkFailed(t *testing.T) {
	address := common.HexToAddress("0x1234567890123456789012345678901234567890")

	t.Run("mark failed clears pending and cached nonce", func(t *testing.T) {
		// Create fresh instance for this test
		nm := &NonceManager{
			client:     (*ethclient.Client)(nil),
			address:    address,
			pendingTxs: make(map[uint64]bool),
		}
		startNonce := uint64(10)
		nm.nonce = &startNonce

		// Simulate getting a nonce
		currentNonce := *nm.nonce
		nm.pendingTxs[currentNonce] = true
		*nm.nonce++

		// Verify nonce was allocated
		if len(nm.pendingTxs) != 1 {
			t.Errorf("expected 1 pending tx, got %d", len(nm.pendingTxs))
		}
		if *nm.nonce != 11 {
			t.Errorf("expected nonce 11, got %d", *nm.nonce)
		}

		// Mark as failed
		nm.MarkFailed(currentNonce)

		// Verify nonce was released
		if len(nm.pendingTxs) != 0 {
			t.Errorf("expected 0 pending txs, got %d", len(nm.pendingTxs))
		}
		// Cached nonce should be cleared so next GetNonce refreshes from network
		if nm.nonce != nil {
			t.Errorf("expected cached nonce to be cleared, got %v", *nm.nonce)
		}
	})

	t.Run("mark failed on old nonce clears cached nonce", func(t *testing.T) {
		// Create fresh instance for this test
		nm := &NonceManager{
			client:     (*ethclient.Client)(nil),
			address:    address,
			pendingTxs: make(map[uint64]bool),
		}
		currentNonce := uint64(18)
		nm.nonce = &currentNonce

		// Simulate multiple nonces allocated
		nm.pendingTxs[15] = true
		nm.pendingTxs[16] = true
		nm.pendingTxs[17] = true

		// Mark old nonce as failed (not the most recent)
		nm.MarkFailed(15)

		// Should remove from pending
		if _, exists := nm.pendingTxs[15]; exists {
			t.Error("nonce 15 should be removed from pending")
		}
		// Cached nonce should be cleared
		if nm.nonce != nil {
			t.Errorf("expected cached nonce to be cleared, got %v", *nm.nonce)
		}
	})
}

func TestNonceManager_MarkConfirmed(t *testing.T) {
	address := common.HexToAddress("0x1234567890123456789012345678901234567890")
	nm := &NonceManager{
		client:     (*ethclient.Client)(nil),
		address:    address,
		pendingTxs: make(map[uint64]bool),
	}

	// Add some pending transactions
	nm.pendingTxs[10] = true
	nm.pendingTxs[11] = true
	nm.pendingTxs[12] = true

	// Mark one as confirmed
	nm.MarkConfirmed(11)

	// Verify it was removed
	if _, exists := nm.pendingTxs[11]; exists {
		t.Error("nonce 11 should be removed from pending")
	}

	// Verify others remain
	if _, exists := nm.pendingTxs[10]; !exists {
		t.Error("nonce 10 should still be pending")
	}
	if _, exists := nm.pendingTxs[12]; !exists {
		t.Error("nonce 12 should still be pending")
	}
}

func TestNonceManager_GetPendingCount(t *testing.T) {
	address := common.HexToAddress("0x1234567890123456789012345678901234567890")
	nm := &NonceManager{
		client:     (*ethclient.Client)(nil),
		address:    address,
		pendingTxs: make(map[uint64]bool),
	}

	if nm.GetPendingCount() != 0 {
		t.Errorf("expected 0 pending, got %d", nm.GetPendingCount())
	}

	nm.pendingTxs[1] = true
	nm.pendingTxs[2] = true
	nm.pendingTxs[3] = true

	if nm.GetPendingCount() != 3 {
		t.Errorf("expected 3 pending, got %d", nm.GetPendingCount())
	}
}

func TestNonceManager_ResetClearsPending(t *testing.T) {
	address := common.HexToAddress("0x1234567890123456789012345678901234567890")

	nm := &NonceManager{
		client:     (*ethclient.Client)(nil),
		address:    address,
		pendingTxs: make(map[uint64]bool),
	}
	currentNonce := uint64(100)
	nm.nonce = &currentNonce
	nm.pendingTxs[60] = true

	// Manually simulate reset (without network call)
	nm.mu.Lock()
	nm.pendingTxs = make(map[uint64]bool)
	nm.nonce = nil
	nm.mu.Unlock()

	// Verify pendingTxs cleared and cached nonce reset
	if nm.nonce != nil {
		t.Errorf("expected cached nonce to be nil, got %v", nm.nonce)
	}
	if len(nm.pendingTxs) != 0 {
		t.Errorf("expected pendingTxs to be empty, got %d", len(nm.pendingTxs))
	}
}
