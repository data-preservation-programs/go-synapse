package storage

import (
	"github.com/ipfs/go-cid"
)

type UploadResult struct {
	PieceCID  cid.Cid
	Size      int64
	PieceID   int
	DataSetID int
}

type UploadOptions struct {
	Metadata map[string]string
	PieceCID cid.Cid // Required for streaming uploads, optional for byte uploads
	Size     int64   // Required for streaming uploads
}

type DownloadOptions struct {
	WithCDN bool
}
