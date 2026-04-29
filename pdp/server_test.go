package pdp

import (
	"context"
	"encoding/hex"
	"encoding/json"
	"io"
	"math/big"
	"net/http"
	"net/http/httptest"
	"sync/atomic"
	"testing"
	"time"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
	"github.com/ipfs/go-cid"
)

func testAuthHelper(t *testing.T) *AuthHelper {
	privateKeyBytes, _ := hex.DecodeString("1234567890123456789012345678901234567890123456789012345678901234")
	privateKey, _ := crypto.ToECDSA(privateKeyBytes)
	contractAddr := common.HexToAddress("0x5615dEB798BB3E4dFa0139dFa1b3D433Cc23b72f")
	chainID := big.NewInt(31337)

	return NewAuthHelperFromKey(privateKey, contractAddr, chainID)
}

func setupMockServer(t *testing.T, handler http.Handler) (*Server, *httptest.Server) {
	mockServer := httptest.NewServer(handler)
	t.Cleanup(mockServer.Close)
	return NewServer(mockServer.URL), mockServer
}

func TestServer_NewServer(t *testing.T) {
	tests := []struct {
		name        string
		baseURL     string
		expectedURL string
	}{
		{
			name:        "regular URL",
			baseURL:     "https://example.com/pdp",
			expectedURL: "https://example.com/pdp",
		},
		{
			name:        "URL with trailing slash",
			baseURL:     "https://example.com/pdp/",
			expectedURL: "https://example.com/pdp",
		},
		{
			name:        "URL with multiple trailing slashes",
			baseURL:     "https://example.com/pdp///",
			expectedURL: "https://example.com/pdp//",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := NewServer(tt.baseURL)
			if server.BaseURL() != tt.expectedURL {
				t.Errorf("BaseURL() = %s, want %s", server.BaseURL(), tt.expectedURL)
			}
		})
	}
}

func TestServer_Ping(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful ping",
			statusCode: http.StatusOK,
			wantErr:    false,
		},
		{
			name:       "server error",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
		{
			name:       "not found",
			statusCode: http.StatusNotFound,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server, _ := setupMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.URL.Path != "/pdp/ping" {
					t.Errorf("Expected path /pdp/ping, got %s", r.URL.Path)
				}
				w.WriteHeader(tt.statusCode)
			}))

			err := server.Ping(context.Background())
			if (err != nil) != tt.wantErr {
				t.Errorf("Ping() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestServer_CreateDataSet(t *testing.T) {
	t.Run("successful creation", func(t *testing.T) {
		expectedTxHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

		server, _ := setupMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/pdp/data-sets" {
				t.Errorf("Expected path /pdp/data-sets, got %s", r.URL.Path)
			}
			w.Header().Set("Location", "/pdp/data-sets/created/"+expectedTxHash)
			w.WriteHeader(http.StatusCreated)
		}))

		result, err := server.CreateDataSet(
			context.Background(),
			"0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
			"0xextraData",
		)
		if err != nil {
			t.Fatalf("CreateDataSet failed: %v", err)
		}

		if result.TxHash != expectedTxHash {
			t.Errorf("TxHash = %s, want %s", result.TxHash, expectedTxHash)
		}
	})

	t.Run("missing Location header", func(t *testing.T) {
		server, _ := setupMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}))

		_, err := server.CreateDataSet(
			context.Background(),
			"0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
			"0xextraData",
		)
		if err == nil {
			t.Error("Expected error for missing Location header, got nil")
		}
	})

	t.Run("server error", func(t *testing.T) {
		server, _ := setupMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusInternalServerError)
			_, _ = w.Write([]byte("Internal server error"))
		}))

		_, err := server.CreateDataSet(
			context.Background(),
			"0x70997970C51812dc3A010C7d01b50e0d17dc79C8",
			"0xextraData",
		)
		if err == nil {
			t.Error("Expected error for server error, got nil")
		}
	})
}

func TestServer_CreateDataSetAndAddPieces(t *testing.T) {
	pieceCID := mustCID(t, "baga6ea4seaqao7s73y24kcutaosvacpdjgfe5pw76ooefnyqw4ynr3d2y6x2mpq")
	recordKeeper := "0x02925630df557F957f70E112bA06e50965417CA0"
	extraData := "0xdeadbeef"

	t.Run("successful creation", func(t *testing.T) {
		expectedTxHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"
		var seen CreateAndAddRequest

		server, _ := setupMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/pdp/data-sets/create-and-add" {
				t.Errorf("Expected path /pdp/data-sets/create-and-add, got %s", r.URL.Path)
			}
			body, _ := io.ReadAll(r.Body)
			if err := json.Unmarshal(body, &seen); err != nil {
				t.Fatalf("decode body: %v", err)
			}
			w.Header().Set("Location", "/pdp/data-sets/created/"+expectedTxHash)
			w.WriteHeader(http.StatusCreated)
		}))

		result, err := server.CreateDataSetAndAddPieces(
			context.Background(),
			recordKeeper,
			[]cid.Cid{pieceCID},
			extraData,
		)
		if err != nil {
			t.Fatalf("CreateDataSetAndAddPieces failed: %v", err)
		}
		if result.TxHash != expectedTxHash {
			t.Errorf("TxHash = %s, want %s", result.TxHash, expectedTxHash)
		}
		if seen.RecordKeeper != recordKeeper {
			t.Errorf("RecordKeeper = %s, want %s", seen.RecordKeeper, recordKeeper)
		}
		if seen.ExtraData != extraData {
			t.Errorf("ExtraData = %s, want %s", seen.ExtraData, extraData)
		}
		if len(seen.Pieces) != 1 {
			t.Fatalf("len(Pieces) = %d, want 1", len(seen.Pieces))
		}
		if seen.Pieces[0].PieceCID != pieceCID.String() {
			t.Errorf("Pieces[0].PieceCID = %s, want %s", seen.Pieces[0].PieceCID, pieceCID.String())
		}
		if len(seen.Pieces[0].SubPieces) != 1 || seen.Pieces[0].SubPieces[0].SubPieceCID != pieceCID.String() {
			t.Errorf("Pieces[0].SubPieces shape mismatch: %+v", seen.Pieces[0].SubPieces)
		}
	})

	t.Run("missing Location header", func(t *testing.T) {
		server, _ := setupMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusCreated)
		}))

		_, err := server.CreateDataSetAndAddPieces(
			context.Background(),
			recordKeeper,
			[]cid.Cid{pieceCID},
			extraData,
		)
		if err == nil {
			t.Error("Expected error for missing Location header, got nil")
		}
	})

	t.Run("server error is surfaced", func(t *testing.T) {
		server, _ := setupMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("extraData validation failed"))
		}))

		_, err := server.CreateDataSetAndAddPieces(
			context.Background(),
			recordKeeper,
			[]cid.Cid{pieceCID},
			extraData,
		)
		if err == nil {
			t.Error("Expected error for 400 response, got nil")
		}
	})
}

func mustCID(t *testing.T, s string) cid.Cid {
	t.Helper()
	c, err := cid.Decode(s)
	if err != nil {
		t.Fatalf("decode CID %q: %v", s, err)
	}
	return c
}

func TestServer_GetDataSetCreationStatus(t *testing.T) {
	t.Run("successful status check", func(t *testing.T) {
		txHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

		server, _ := setupMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			expectedPath := "/pdp/data-sets/created/" + txHash
			if r.URL.Path != expectedPath {
				t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"createMessageHash": "` + txHash + `",
				"dataSetCreated": true,
				"txStatus": "confirmed",
				"dataSetId": 123
			}`))
		}))

		status, err := server.GetDataSetCreationStatus(context.Background(), txHash)
		if err != nil {
			t.Fatalf("GetDataSetCreationStatus failed: %v", err)
		}

		if !status.DataSetCreated {
			t.Error("Expected DataSetCreated to be true")
		}
		if status.DataSetID == nil || *status.DataSetID != 123 {
			t.Error("Expected DataSetID to be 123")
		}
	})

	t.Run("not found", func(t *testing.T) {
		txHash := "0x1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcdef"

		server, _ := setupMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))

		_, err := server.GetDataSetCreationStatus(context.Background(), txHash)
		if err == nil {
			t.Error("Expected error for not found, got nil")
		}
	})
}

func TestServer_GetDataSet(t *testing.T) {
	t.Run("successful fetch", func(t *testing.T) {
		server, _ := setupMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.URL.Path != "/pdp/data-sets/123" {
				t.Errorf("Expected path /pdp/data-sets/123, got %s", r.URL.Path)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{
				"id": 123,
				"pieces": [],
				"nextChallengeEpoch": 1500
			}`))
		}))

		data, err := server.GetDataSet(context.Background(), 123)
		if err != nil {
			t.Fatalf("GetDataSet failed: %v", err)
		}

		if data.ID != 123 {
			t.Errorf("ID = %d, want 123", data.ID)
		}
	})

	t.Run("not found", func(t *testing.T) {
		server, _ := setupMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusNotFound)
		}))

		_, err := server.GetDataSet(context.Background(), 999)
		if err == nil {
			t.Error("Expected error for not found, got nil")
		}
	})
}

func TestServer_PullPieces(t *testing.T) {
	pieces := []PullPieceInput{
		{PieceCID: "bafkz...A", SourceURL: "https://example.com/piece/bafkz...A"},
		{PieceCID: "bafkz...B", SourceURL: "https://example.com/piece/bafkz...B"},
	}
	recordKeeper := "0x02925630df557F957f70E112bA06e50965417CA0"
	extraData := "0xdeadbeef"

	t.Run("create-new omits dataSetId on the wire", func(t *testing.T) {
		var seen PullPiecesRequest
		server, _ := setupMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			if r.Method != "POST" {
				t.Errorf("Expected POST, got %s", r.Method)
			}
			if r.URL.Path != "/pdp/piece/pull" {
				t.Errorf("Expected path /pdp/piece/pull, got %s", r.URL.Path)
			}
			body, _ := io.ReadAll(r.Body)
			// re-decode into a map to assert dataSetId truly absent
			var raw map[string]any
			if err := json.Unmarshal(body, &raw); err != nil {
				t.Fatalf("decode raw body: %v", err)
			}
			if _, present := raw["dataSetId"]; present {
				t.Errorf("expected dataSetId omitted for new set, body=%s", string(body))
			}
			if err := json.Unmarshal(body, &seen); err != nil {
				t.Fatalf("decode typed body: %v", err)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"pending","pieces":[{"pieceCid":"bafkz...A","status":"pending"},{"pieceCid":"bafkz...B","status":"pending"}]}`))
		}))

		resp, err := server.PullPieces(context.Background(), PullPiecesOptions{
			RecordKeeper: recordKeeper,
			Pieces:       pieces,
			ExtraData:    extraData,
			DataSetID:    0,
		})
		if err != nil {
			t.Fatalf("PullPieces failed: %v", err)
		}
		if resp.Status != PullStatusPending {
			t.Errorf("Status = %s, want pending", resp.Status)
		}
		if len(resp.Pieces) != 2 {
			t.Errorf("len(Pieces) = %d, want 2", len(resp.Pieces))
		}
		if seen.RecordKeeper != recordKeeper {
			t.Errorf("RecordKeeper = %s, want %s", seen.RecordKeeper, recordKeeper)
		}
		if seen.ExtraData != extraData {
			t.Errorf("ExtraData = %s, want %s", seen.ExtraData, extraData)
		}
		if len(seen.Pieces) != 2 || seen.Pieces[0].PieceCID != pieces[0].PieceCID {
			t.Errorf("Pieces mismatch: %+v", seen.Pieces)
		}
	})

	t.Run("existing set sends dataSetId on the wire", func(t *testing.T) {
		server, _ := setupMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			body, _ := io.ReadAll(r.Body)
			var raw map[string]any
			if err := json.Unmarshal(body, &raw); err != nil {
				t.Fatalf("decode raw body: %v", err)
			}
			id, ok := raw["dataSetId"].(float64)
			if !ok {
				t.Errorf("expected numeric dataSetId, body=%s", string(body))
			}
			if uint64(id) != 13245 {
				t.Errorf("dataSetId = %v, want 13245", id)
			}
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"complete","pieces":[{"pieceCid":"bafkz...A","status":"complete"}]}`))
		}))

		resp, err := server.PullPieces(context.Background(), PullPiecesOptions{
			RecordKeeper: recordKeeper,
			Pieces:       pieces[:1],
			ExtraData:    extraData,
			DataSetID:    13245,
		})
		if err != nil {
			t.Fatalf("PullPieces failed: %v", err)
		}
		if resp.Status != PullStatusComplete {
			t.Errorf("Status = %s, want complete", resp.Status)
		}
	})

	t.Run("server error is surfaced", func(t *testing.T) {
		server, _ := setupMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.WriteHeader(http.StatusBadRequest)
			_, _ = w.Write([]byte("extraData validation failed"))
		}))

		_, err := server.PullPieces(context.Background(), PullPiecesOptions{
			RecordKeeper: recordKeeper,
			Pieces:       pieces,
			ExtraData:    extraData,
		})
		if err == nil {
			t.Error("Expected error from 400 response, got nil")
		}
	})
}

func TestServer_WaitForPullPieces(t *testing.T) {
	pieces := []PullPieceInput{
		{PieceCID: "bafkz...A", SourceURL: "https://example.com/piece/bafkz...A"},
	}
	opts := PullPiecesOptions{
		RecordKeeper: "0x02925630df557F957f70E112bA06e50965417CA0",
		Pieces:       pieces,
		ExtraData:    "0xdeadbeef",
	}

	t.Run("polls until complete", func(t *testing.T) {
		var hits int32
		server, _ := setupMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			n := atomic.AddInt32(&hits, 1)
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			if n < 2 {
				_, _ = w.Write([]byte(`{"status":"inProgress","pieces":[{"pieceCid":"bafkz...A","status":"inProgress"}]}`))
				return
			}
			_, _ = w.Write([]byte(`{"status":"complete","pieces":[{"pieceCid":"bafkz...A","status":"complete"}]}`))
		}))

		// shrink the poll interval inside retry.Poll: not configurable, but
		// the default 4s window is fine if we give a generous timeout.
		resp, err := server.WaitForPullPieces(context.Background(), opts, 30*time.Second)
		if err != nil {
			t.Fatalf("WaitForPullPieces failed: %v", err)
		}
		if resp.Status != PullStatusComplete {
			t.Errorf("Status = %s, want complete", resp.Status)
		}
		if atomic.LoadInt32(&hits) < 2 {
			t.Errorf("expected at least 2 polls, got %d", hits)
		}
	})

	t.Run("returns failed status without error", func(t *testing.T) {
		server, _ := setupMockServer(t, http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			_, _ = w.Write([]byte(`{"status":"failed","pieces":[{"pieceCid":"bafkz...A","status":"failed"}]}`))
		}))

		resp, err := server.WaitForPullPieces(context.Background(), opts, 5*time.Second)
		if err != nil {
			t.Fatalf("WaitForPullPieces unexpectedly errored on failed status: %v", err)
		}
		if resp.Status != PullStatusFailed {
			t.Errorf("Status = %s, want failed", resp.Status)
		}
	})
}
