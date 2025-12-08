package pdp

import (
	"context"
	"encoding/hex"
	"math/big"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/ethereum/go-ethereum/common"
	"github.com/ethereum/go-ethereum/crypto"
)

func testAuthHelper(t *testing.T) *AuthHelper {
	privateKeyBytes, _ := hex.DecodeString("1234567890123456789012345678901234567890123456789012345678901234")
	privateKey, _ := crypto.ToECDSA(privateKeyBytes)
	contractAddr := common.HexToAddress("0x5615dEB798BB3E4dFa0139dFa1b3D433Cc23b72f")
	chainID := big.NewInt(31337)

	return NewAuthHelper(privateKey, contractAddr, chainID)
}

func setupMockServer(t *testing.T, handler http.Handler) (*Server, *httptest.Server) {
	authHelper := testAuthHelper(t)
	mockServer := httptest.NewServer(handler)
	t.Cleanup(mockServer.Close)
	return NewServer(mockServer.URL, authHelper), mockServer
}

func TestServer_NewServer(t *testing.T) {
	authHelper := testAuthHelper(t)

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
			server := NewServer(tt.baseURL, authHelper)
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
