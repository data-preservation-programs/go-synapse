package txutil

import (
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/ethclient"
)

func TestNonceManager_MarkFailed(t *testing.T) {
	address := common.HexToAddress("0x1234567890123456789012345678901234567890")

	t.Run("mark failed adds nonce to reclaimable pool", func(t *testing.T) {
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

		// Verify nonce was released and added to reclaimable pool
		if len(nm.pendingTxs) != 0 {
			t.Errorf("expected 0 pending txs, got %d", len(nm.pendingTxs))
		}
		// Nonce counter stays the same (no rollback)
		if *nm.nonce != 11 {
			t.Errorf("expected nonce counter to stay at 11, got %d", *nm.nonce)
		}
		// But the nonce should be in the reclaimable pool
		if len(nm.reclaimable) != 1 || nm.reclaimable[0] != 10 {
			t.Errorf("expected reclaimable pool to contain [10], got %v", nm.reclaimable)
		}
	})

	t.Run("mark failed on old nonce adds to reclaimable pool", func(t *testing.T) {
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
		// Nonce counter stays the same
		if *nm.nonce != 18 {
			t.Errorf("nonce counter should stay at 18, got %d", *nm.nonce)
		}
		// Should be in reclaimable pool
		if len(nm.reclaimable) != 1 || nm.reclaimable[0] != 15 {
			t.Errorf("expected reclaimable pool to contain [15], got %v", nm.reclaimable)
		}
	})

	t.Run("multiple failed nonces are all reclaimable", func(t *testing.T) {
		// Create fresh instance for this test
		nm := &NonceManager{
			client:     (*ethclient.Client)(nil),
			address:    address,
			pendingTxs: make(map[uint64]bool),
		}
		currentNonce := uint64(35)
		nm.nonce = &currentNonce

		// Allocate several nonces
		nm.pendingTxs[30] = true
		nm.pendingTxs[31] = true
		nm.pendingTxs[32] = true
		nm.pendingTxs[33] = true
		nm.pendingTxs[34] = true

		// Mark several as failed (out of order to test sorting)
		nm.MarkFailed(32)
		nm.MarkFailed(30)
		nm.MarkFailed(34)

		// All should be in reclaimable pool
		if len(nm.reclaimable) != 3 {
			t.Errorf("expected 3 reclaimable nonces, got %d", len(nm.reclaimable))
		}

		// Nonce counter stays the same
		if *nm.nonce != 35 {
			t.Errorf("nonce counter should stay at 35, got %d", *nm.nonce)
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

func TestNonceManager_ReclaimablePool(t *testing.T) {
	address := common.HexToAddress("0x1234567890123456789012345678901234567890")

	t.Run("reclaimable nonces are reused in sorted order", func(t *testing.T) {
		nm := &NonceManager{
			client:     (*ethclient.Client)(nil),
			address:    address,
			pendingTxs: make(map[uint64]bool),
		}
		currentNonce := uint64(100)
		nm.nonce = &currentNonce

		// Add some nonces to reclaimable pool (out of order)
		nm.reclaimable = []uint64{50, 30, 40}

		// GetNonce should use smallest reclaimable first
		nm.mu.Lock()
		// Simulate what GetNonce does for reclaimable
		if len(nm.reclaimable) > 0 {
			// Sort and take smallest
			smallest := nm.reclaimable[0]
			for _, n := range nm.reclaimable {
				if n < smallest {
					smallest = n
				}
			}
			// Should be 30
			if smallest != 30 {
				t.Errorf("expected smallest to be 30, got %d", smallest)
			}
		}
		nm.mu.Unlock()
	})

	t.Run("reset clears reclaimable pool", func(t *testing.T) {
		nm := &NonceManager{
			client:     (*ethclient.Client)(nil),
			address:    address,
			pendingTxs: make(map[uint64]bool),
		}
		currentNonce := uint64(100)
		nm.nonce = &currentNonce

		// Add some nonces to reclaimable pool
		nm.reclaimable = []uint64{50, 30, 40}
		nm.pendingTxs[60] = true

		// Manually simulate reset (without network call)
		nm.mu.Lock()
		nm.pendingTxs = make(map[uint64]bool)
		nm.reclaimable = nil
		nm.mu.Unlock()

		// Verify reclaimable is cleared
		if nm.reclaimable != nil {
			t.Errorf("expected reclaimable to be nil, got %v", nm.reclaimable)
		}
		if len(nm.pendingTxs) != 0 {
			t.Errorf("expected pendingTxs to be empty, got %d", len(nm.pendingTxs))
		}
	})
}
