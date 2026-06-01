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
		return nil, fmt.Errorf("изображение слишком большое (макс. 10 МБ)")
	}
	data, err := io.ReadAll(io.LimitReader(file, maxUploadImageBytes+1))
	if err != nil {
		return nil, fmt.Errorf("не удалось прочитать изображение: %w", err)
	}
	if len(data) > maxUploadImageBytes {
		return nil, fmt.Errorf("изображение слишком большое (макс. 10 МБ)")
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
		return nil, fmt.Errorf("не удалось открыть файл изображения")
	}
	defer f.Close()
	return readUploadImage(f, fh.Size)
}
