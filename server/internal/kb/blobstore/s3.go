package blobstore

import (
	"context"
	"fmt"
	"io"
	"os"
	"strings"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

// S3Store stores blobs in S3-compatible object storage (AWS S3, MinIO, etc.).
type S3Store struct {
	client *minio.Client
	bucket string
	prefix string
}

// NewS3FromEnv configures S3 from KB_S3_* env vars.
func NewS3FromEnv() (*S3Store, error) {
	endpoint := strings.TrimSpace(os.Getenv("KB_S3_ENDPOINT"))
	accessKey := strings.TrimSpace(os.Getenv("KB_S3_ACCESS_KEY"))
	secretKey := strings.TrimSpace(os.Getenv("KB_S3_SECRET_KEY"))
	bucket := strings.TrimSpace(os.Getenv("KB_S3_BUCKET"))
	if bucket == "" {
		bucket = "grounded-kb"
	}
	prefix := strings.Trim(strings.TrimSpace(os.Getenv("KB_S3_PREFIX")), "/")
	useSSL := strings.EqualFold(os.Getenv("KB_S3_USE_SSL"), "true")
	region := strings.TrimSpace(os.Getenv("KB_S3_REGION"))
	if region == "" {
		region = "us-east-1"
	}
	if endpoint == "" {
		return nil, fmt.Errorf("KB_S3_ENDPOINT is required when KB_BLOB_BACKEND=s3")
	}
	if accessKey == "" || secretKey == "" {
		return nil, fmt.Errorf("KB_S3_ACCESS_KEY and KB_S3_SECRET_KEY are required")
	}
	client, err := minio.New(endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(accessKey, secretKey, ""),
		Secure: useSSL,
		Region: region,
	})
	if err != nil {
		return nil, err
	}
	return &S3Store{client: client, bucket: bucket, prefix: prefix}, nil
}

func (s *S3Store) fullKey(key string) string {
	key = strings.TrimPrefix(strings.ReplaceAll(key, "\\", "/"), "/")
	if s.prefix == "" {
		return key
	}
	return s.prefix + "/" + key
}

func (s *S3Store) BuildKey(tenantID, domainID, documentID string, version int, sha256Hex, ext string) string {
	ext = strings.TrimPrefix(ext, ".")
	if ext != "" {
		ext = "." + ext
	}
	return fmt.Sprintf(
		"tenants/%s/domains/%s/docs/%s/v%d/%s%s",
		tenantID, domainID, documentID, version, sha256Hex, ext,
	)
}

func (s *S3Store) Put(ctx context.Context, key string, r io.Reader, size int64, contentType string) (string, error) {
	if contentType == "" {
		contentType = "application/octet-stream"
	}
	_, err := s.client.PutObject(ctx, s.bucket, s.fullKey(key), r, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	if err != nil {
		return "", err
	}
	return key, nil
}

func (s *S3Store) Get(ctx context.Context, key string) ([]byte, error) {
	obj, err := s.client.GetObject(ctx, s.bucket, s.fullKey(key), minio.GetObjectOptions{})
	if err != nil {
		return nil, err
	}
	defer obj.Close()
	return io.ReadAll(obj)
}

func (s *S3Store) Delete(ctx context.Context, key string) error {
	return s.client.RemoveObject(ctx, s.bucket, s.fullKey(key), minio.RemoveObjectOptions{})
}
