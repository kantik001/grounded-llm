package admin

import (
	"bytes"
	"fmt"
	"os"
	"path/filepath"
	"strings"
	"unicode/utf8"
)

func validateKnowledgeFileBytes(data []byte, filename string) error {
	ext := strings.ToLower(strings.TrimPrefix(filepath.Ext(filename), "."))
	if len(data) == 0 {
		return fmt.Errorf("file is empty")
	}
	switch ext {
	case "pdf":
		if !bytes.HasPrefix(data, []byte("%PDF")) {
			return fmt.Errorf("file content is not a valid PDF")
		}
	case "docx":
		if !bytes.HasPrefix(data, []byte("PK")) {
			return fmt.Errorf("file content is not a valid DOCX (ZIP)")
		}
	case "txt":
		if bytes.Contains(data, []byte{0}) {
			return fmt.Errorf("TXT file contains binary null bytes")
		}
		if !utf8.Valid(data) {
			return fmt.Errorf("TXT file is not valid UTF-8")
		}
		if len(strings.TrimSpace(string(data))) == 0 {
			return fmt.Errorf("TXT file is empty")
		}
	default:
		return fmt.Errorf("unsupported extension")
	}
	return nil
}

// validateKnowledgeFileContent checks magic bytes / content shape after upload.
func validateKnowledgeFileContent(path, filename string) error {
	data, err := os.ReadFile(path)
	if err != nil {
		return err
	}
	return validateKnowledgeFileBytes(data, filename)
}

// ValidateKnowledgeFileContent is exported for tests.
func ValidateKnowledgeFileContent(path, filename string) error {
	return validateKnowledgeFileContent(path, filename)
}

// ValidateKnowledgeFileBytes is exported for tests.
func ValidateKnowledgeFileBytes(data []byte, filename string) error {
	return validateKnowledgeFileBytes(data, filename)
}
