package blobstore

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// Store persists KB blob bytes (local filesystem or S3-compatible object storage).
type Store interface {
	Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (storageKey string, err error)
	Get(ctx context.Context, key string) ([]byte, error)
	Delete(ctx context.Context, key string) error
	BuildKey(tenantID, domainID, documentID string, version int, sha256Hex, ext string) string
}

// ContentSHA256 returns hex-encoded SHA-256 of data.
func ContentSHA256(data []byte) string {
	sum := sha256.Sum256(data)
	return hex.EncodeToString(sum[:])
}

// NewFromEnv selects backend via KB_BLOB_BACKEND (local | s3). Default: local.
func NewFromEnv(dataDir string) (Store, error) {
	backend := strings.ToLower(strings.TrimSpace(os.Getenv("KB_BLOB_BACKEND")))
	if backend == "" || backend == "local" {
		root := strings.TrimSpace(os.Getenv("KB_BLOB_DIR"))
		if root == "" {
			root = filepath.Join(dataDir, "blobs")
		}
		return NewLocal(root), nil
	}
	if backend == "s3" {
		return NewS3FromEnv()
	}
	return nil, fmt.Errorf("unknown KB_BLOB_BACKEND=%q (supported: local, s3)", backend)
}
