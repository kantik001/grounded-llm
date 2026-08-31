package admin

import (
	"bytes"
	"context"
	"io"
	"mime"
	"os"
	"path/filepath"
	"strings"

	"grounded_llm_server/internal/kb/blobstore"
	"grounded_llm_server/internal/store"
)

var kbBlobStore blobstore.Store

// InitKBBlobStore initializes the KB blob backend from environment.
func InitKBBlobStore(dataDir string) error {
	s, err := blobstore.NewFromEnv(dataDir)
	if err != nil {
		return err
	}
	kbBlobStore = s
	return nil
}

func kbBlob() blobstore.Store {
	return kbBlobStore
}

func mimeForFilename(name string) string {
	ext := strings.ToLower(filepath.Ext(name))
	if ext == "" {
		return "application/octet-stream"
	}
	if mt := mime.TypeByExtension(ext); mt != "" {
		return mt
	}
	switch ext {
	case ".txt":
		return "text/plain"
	case ".pdf":
		return "application/pdf"
	case ".docx":
		return "application/vnd.openxmlformats-officedocument.wordprocessingml.document"
	default:
		return "application/octet-stream"
	}
}

// registerKBUpload stores blob bytes and upserts Postgres registry metadata.
func registerKBUpload(
	ctx context.Context,
	st *store.ChatStore,
	tenantID, domainID, filename, actor string,
	data []byte,
) (store.KBDocument, store.KBDocumentVersion, error) {
	var zeroDoc store.KBDocument
	var zeroVer store.KBDocumentVersion
	if st == nil {
		return zeroDoc, zeroVer, io.ErrClosedPipe
	}
	blob := kbBlob()
	if blob == nil {
		return zeroDoc, zeroVer, io.ErrUnexpectedEOF
	}

	sha := blobstore.ContentSHA256(data)
	doc, ver, err := st.UpsertKBDocument(ctx, store.UpsertKBDocumentInput{
		TenantID:      tenantID,
		DomainID:      domainID,
		LogicalKey:    filename,
		Title:         filename,
		MimeType:      mimeForFilename(filename),
		ContentSHA256: sha,
		SizeBytes:     int64(len(data)),
		Source:        "upload",
		SourceRef:     map[string]any{"actor": actor},
		CreatedBy:     actor,
	})
	if err != nil {
		return zeroDoc, zeroVer, err
	}

	if _, err := blob.Put(ctx, ver.StorageKey, bytes.NewReader(data), int64(len(data)), mimeForFilename(filename)); err != nil {
		return zeroDoc, zeroVer, err
	}

	backend := strings.ToLower(strings.TrimSpace(os.Getenv("VECTOR_STORE")))
	if backend == "" {
		backend = "chroma"
	}
	model := strings.TrimSpace(os.Getenv("EMBEDDING_MODEL"))
	if model == "" {
		model = "intfloat/multilingual-e5-small"
	}
	_, _ = st.EnsureActiveIndexRun(ctx, tenantID, domainID, backend, model)

	return doc, ver, nil
}
