package blobstore

import (
	"context"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
)

// LocalStore stores blobs on disk under root (content-addressed layout optional).
type LocalStore struct {
	root string
}

// NewLocal returns a filesystem blob store.
func NewLocal(root string) *LocalStore {
	return &LocalStore{root: root}
}

func (s *LocalStore) absKey(key string) string {
	clean := strings.TrimPrefix(strings.ReplaceAll(key, "\\", "/"), "/")
	return filepath.Join(s.root, filepath.FromSlash(clean))
}

// BuildKey builds a versioned storage key for a document version.
func (s *LocalStore) BuildKey(tenantID, domainID, documentID string, version int, sha256Hex, ext string) string {
	ext = strings.TrimPrefix(ext, ".")
	if ext != "" {
		ext = "." + ext
	}
	return fmt.Sprintf(
		"tenants/%s/domains/%s/docs/%s/v%d/%s%s",
		tenantID, domainID, documentID, version, sha256Hex, ext,
	)
}

func (s *LocalStore) Put(_ context.Context, key string, r io.Reader, _ int64, _ string) (string, error) {
	path := s.absKey(key)
	if err := os.MkdirAll(filepath.Dir(path), 0o755); err != nil {
		return "", err
	}
	tmp := path + ".tmp"
	out, err := os.Create(tmp)
	if err != nil {
		return "", err
	}
	if _, err := io.Copy(out, r); err != nil {
		_ = out.Close()
		_ = os.Remove(tmp)
		return "", err
	}
	if err := out.Close(); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	if err := os.Rename(tmp, path); err != nil {
		_ = os.Remove(tmp)
		return "", err
	}
	return key, nil
}

func (s *LocalStore) Get(_ context.Context, key string) ([]byte, error) {
	return os.ReadFile(s.absKey(key))
}

func (s *LocalStore) Delete(_ context.Context, key string) error {
	err := os.Remove(s.absKey(key))
	if os.IsNotExist(err) {
		return nil
	}
	return err
}
