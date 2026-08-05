package httpapi

import (
	"fmt"
	"io"
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

// MaxUploadImageBytes is the maximum allowed image upload size for /message multipart.
const MaxUploadImageBytes = 10 * 1024 * 1024

// ReadUploadImage reads and size-checks an uploaded image file.
func ReadUploadImage(file multipart.File, size int64) ([]byte, error) {
	if size > MaxUploadImageBytes {
		return nil, fmt.Errorf("image too large (max 10 MB)")
	}
	data, err := io.ReadAll(io.LimitReader(file, MaxUploadImageBytes+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}
	if len(data) > MaxUploadImageBytes {
		return nil, fmt.Errorf("image too large (max 10 MB)")
	}
	return data, nil
}

// ReadImageFromForm reads an optional image field from a parsed multipart form.
func ReadImageFromForm(c *gin.Context, field string) ([]byte, error) {
	fh, err := c.FormFile(field)
	if err != nil || fh == nil {
		return nil, nil
	}
	f, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open image file")
	}
	defer func() { _ = f.Close() }()
	return ReadUploadImage(f, fh.Size)
}
