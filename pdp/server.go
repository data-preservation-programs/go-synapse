package pdp

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/data-preservation-programs/go-synapse/internal/retry"
	"github.com/ipfs/go-cid"
)

const (
	defaultTimeout = 5 * time.Minute
)

type Server struct {
	baseURL         string
	authHelper      *AuthHelper
	httpClient      *http.Client
	uploadClientMu  sync.Mutex
	uploadClientVal *http.Client
}


func NewServer(baseURL string, authHelper *AuthHelper) *Server {
	baseURL = strings.TrimSuffix(baseURL, "/")

	return &Server{
		baseURL:    baseURL,
		authHelper: authHelper,
		httpClient: &http.Client{
			Timeout: defaultTimeout,
		},
	}
}

func (s *Server) uploadClient() *http.Client {
	s.uploadClientMu.Lock()
	defer s.uploadClientMu.Unlock()
	if s.uploadClientVal == nil {
		s.uploadClientVal = &http.Client{}
	}
	return s.uploadClientVal
}


func (s *Server) BaseURL() string {
	return s.baseURL
}


func (s *Server) CreateDataSet(ctx context.Context, recordKeeper string, extraData string) (*CreateDataSetResponse, error) {
	reqBody := map[string]string{
		"recordKeeper": recordKeeper,
		"extraData":    extraData,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	req, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/pdp/data-sets", bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return nil, fmt.Errorf("missing Location header")
	}

	parts := strings.Split(location, "/")
	txHash := parts[len(parts)-1]
	if !strings.HasPrefix(txHash, "0x") {
		return nil, fmt.Errorf("invalid txHash in Location header: %s", txHash)
	}

	statusURL := s.baseURL + location

	return &CreateDataSetResponse{
		TxHash:    txHash,
		StatusURL: statusURL,
	}, nil
}


func (s *Server) GetDataSetCreationStatus(ctx context.Context, txHash string) (*DataSetCreationStatus, error) {
	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/pdp/data-sets/created/"+txHash, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("data set creation not found for txHash: %s", txHash)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var status DataSetCreationStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &status, nil
}


func (s *Server) WaitForDataSetCreation(ctx context.Context, txHash string, timeout time.Duration) (*DataSetCreationStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var status *DataSetCreationStatus
	err := retry.Poll(ctx, 4*time.Second, timeout, func() (bool, error) {
		var err error
		status, err = s.GetDataSetCreationStatus(ctx, txHash)
		if err != nil {
			return false, err
		}
		return status.DataSetCreated, nil
	})
	if err != nil {
		return nil, err
	}
	return status, nil
}


func (s *Server) AddPieces(ctx context.Context, dataSetID int, pieceCIDs []cid.Cid, extraData string) (*AddPiecesResponse, error) {
	pieces := make([]PieceData, len(pieceCIDs))
	for i, c := range pieceCIDs {
		cidStr := c.String()
		pieces[i] = PieceData{
			PieceCID: cidStr,
			SubPieces: []SubPieceData{
				{SubPieceCID: cidStr},
			},
		}
	}

	reqBody := AddPiecesRequest{
		Pieces:    pieces,
		ExtraData: extraData,
	}

	body, err := json.Marshal(reqBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request: %w", err)
	}

	url := fmt.Sprintf("%s/pdp/data-sets/%d/pieces", s.baseURL, dataSetID)
	req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(body))
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusCreated && resp.StatusCode != http.StatusAccepted {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	location := resp.Header.Get("Location")
	if location == "" {
		return nil, fmt.Errorf("missing Location header")
	}

	parts := strings.Split(location, "/")
	txHash := parts[len(parts)-1]

	statusURL := s.baseURL + location

	return &AddPiecesResponse{
		Message:   fmt.Sprintf("Pieces added to data set ID %d", dataSetID),
		TxHash:    txHash,
		StatusURL: statusURL,
	}, nil
}


func (s *Server) GetPieceAdditionStatus(ctx context.Context, dataSetID int, txHash string) (*PieceAdditionStatus, error) {
	url := fmt.Sprintf("%s/pdp/data-sets/%d/pieces/added/%s", s.baseURL, dataSetID, txHash)
	req, err := http.NewRequestWithContext(ctx, "GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("piece addition not found for txHash: %s", txHash)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var status PieceAdditionStatus
	if err := json.NewDecoder(resp.Body).Decode(&status); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &status, nil
}


func (s *Server) WaitForPieceAddition(ctx context.Context, dataSetID int, txHash string, timeout time.Duration) (*PieceAdditionStatus, error) {
	ctx, cancel := context.WithTimeout(ctx, timeout)
	defer cancel()

	var status *PieceAdditionStatus
	err := retry.Poll(ctx, time.Second, timeout, func() (bool, error) {
		var err error
		status, err = s.GetPieceAdditionStatus(ctx, dataSetID, txHash)
		if err != nil {
			return false, err
		}
		return status.AddMessageOK != nil && *status.AddMessageOK, nil
	})
	if err != nil {
		return nil, err
	}
	return status, nil
}


func (s *Server) UploadPiece(ctx context.Context, data io.Reader, size int64, pieceCID cid.Cid) (*UploadPieceResponse, error) {
	createReq, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/pdp/piece/uploads", nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create session request: %w", err)
	}

	createResp, err := s.httpClient.Do(createReq)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload session: %w", err)
	}
	defer createResp.Body.Close()

	if createResp.StatusCode != http.StatusCreated {
		respBody, _ := io.ReadAll(createResp.Body)
		return nil, fmt.Errorf("failed to create upload session: status %d: %s", createResp.StatusCode, string(respBody))
	}

	location := createResp.Header.Get("Location")
	if location == "" {
		return nil, fmt.Errorf("missing Location header in upload session response")
	}

	uuidRegex := regexp.MustCompile(`/pdp/piece/uploads/([a-fA-F0-9-]+)`)
	matches := uuidRegex.FindStringSubmatch(location)
	if len(matches) < 2 {
		return nil, fmt.Errorf("invalid Location header format: %s", location)
	}
	uploadUUID := matches[1]

	uploadReq, err := http.NewRequestWithContext(ctx, "PUT", s.baseURL+"/pdp/piece/uploads/"+uploadUUID, data)
	if err != nil {
		return nil, fmt.Errorf("failed to create upload request: %w", err)
	}
	uploadReq.Header.Set("Content-Type", "application/octet-stream")
	if size > 0 {
		uploadReq.ContentLength = size
	}

	uploadResp, err := s.uploadClient().Do(uploadReq)
	if err != nil {
		return nil, fmt.Errorf("upload failed: %w", err)
	}
	defer uploadResp.Body.Close()

	if uploadResp.StatusCode != http.StatusNoContent {
		respBody, _ := io.ReadAll(uploadResp.Body)
		return nil, fmt.Errorf("upload failed: status %d: %s", uploadResp.StatusCode, string(respBody))
	}

	finalizeBody, err := json.Marshal(map[string]string{
		"pieceCid": pieceCID.String(),
	})
	if err != nil {
		return nil, fmt.Errorf("failed to marshal finalize request: %w", err)
	}

	finalizeReq, err := http.NewRequestWithContext(ctx, "POST", s.baseURL+"/pdp/piece/uploads/"+uploadUUID, bytes.NewReader(finalizeBody))
	if err != nil {
		return nil, fmt.Errorf("failed to create finalize request: %w", err)
	}
	finalizeReq.Header.Set("Content-Type", "application/json")

	finalizeResp, err := s.httpClient.Do(finalizeReq)
	if err != nil {
		return nil, fmt.Errorf("finalize failed: %w", err)
	}
	defer finalizeResp.Body.Close()

	if finalizeResp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(finalizeResp.Body)
		return nil, fmt.Errorf("finalize failed: status %d: %s", finalizeResp.StatusCode, string(respBody))
	}

	return &UploadPieceResponse{
		PieceCID: pieceCID,
		Size:     size,
	}, nil
}


func (s *Server) FindPiece(ctx context.Context, pieceCID cid.Cid) error {
	params := url.Values{}
	params.Set("pieceCid", pieceCID.String())

	reqURL := fmt.Sprintf("%s/pdp/piece?%s", s.baseURL, params.Encode())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return fmt.Errorf("piece not found: %s", pieceCID.String())
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	return nil
}


func (s *Server) WaitForPiece(ctx context.Context, pieceCID cid.Cid, timeout time.Duration) error {
	return retry.Poll(ctx, 5*time.Second, timeout, func() (bool, error) {
		err := s.FindPiece(ctx, pieceCID)
		if err != nil {
			if strings.Contains(err.Error(), "piece not found") {
				return false, nil
			}
			return false, err
		}
		return true, nil
	})
}


func (s *Server) DownloadPiece(ctx context.Context, pieceCID cid.Cid) ([]byte, error) {
	reqURL := fmt.Sprintf("%s/pdp/piece/%s", s.baseURL, pieceCID.String())
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("piece not found: %s", pieceCID.String())
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	return io.ReadAll(resp.Body)
}


func (s *Server) GetDataSet(ctx context.Context, dataSetID int) (*DataSetData, error) {
	reqURL := fmt.Sprintf("%s/pdp/data-sets/%d", s.baseURL, dataSetID)
	req, err := http.NewRequestWithContext(ctx, "GET", reqURL, nil)
	if err != nil {
		return nil, fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("request failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode == http.StatusNotFound {
		return nil, fmt.Errorf("data set not found: %d", dataSetID)
	}

	if resp.StatusCode != http.StatusOK {
		respBody, _ := io.ReadAll(resp.Body)
		return nil, fmt.Errorf("unexpected status %d: %s", resp.StatusCode, string(respBody))
	}

	var data DataSetData
	if err := json.NewDecoder(resp.Body).Decode(&data); err != nil {
		return nil, fmt.Errorf("failed to decode response: %w", err)
	}

	return &data, nil
}


func (s *Server) Ping(ctx context.Context) error {
	req, err := http.NewRequestWithContext(ctx, "GET", s.baseURL+"/pdp/ping", nil)
	if err != nil {
		return fmt.Errorf("failed to create request: %w", err)
	}

	resp, err := s.httpClient.Do(req)
	if err != nil {
		return fmt.Errorf("ping failed: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("ping failed: status %d", resp.StatusCode)
	}

	return nil
}
