package main

import (
	"fmt"
	"io"
	"mime/multipart"

	"github.com/gin-gonic/gin"
)

const maxUploadImageBytes = 10 * 1024 * 1024

func readUploadImage(file multipart.File, size int64) ([]byte, error) {
	if size > maxUploadImageBytes {
		return nil, fmt.Errorf("image too large (max 10 MB)")
	}
	data, err := io.ReadAll(io.LimitReader(file, maxUploadImageBytes+1))
	if err != nil {
		return nil, fmt.Errorf("failed to read image: %w", err)
	}
	if len(data) > maxUploadImageBytes {
		return nil, fmt.Errorf("image too large (max 10 MB)")
	}
	return data, nil
}

func readImageFromFormFile(c *gin.Context, field string) ([]byte, error) {
	fh, err := c.FormFile(field)
	if err != nil || fh == nil {
		return nil, nil
	}
	f, err := fh.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open image file")
	}
	defer func() { _ = f.Close() }()
	return readUploadImage(f, fh.Size)
}
