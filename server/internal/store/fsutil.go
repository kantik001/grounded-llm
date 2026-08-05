package store

import (
	"os"
	"path/filepath"
	"strings"
)

func isKnowledgeFile(name string) bool {
	n := strings.ToLower(name)
	return strings.HasSuffix(n, ".txt") || strings.HasSuffix(n, ".pdf") || strings.HasSuffix(n, ".docx")
}

func countDataFiles(dir string) int {
	n := 0
	_ = filepath.WalkDir(dir, func(path string, d os.DirEntry, err error) error {
		if err != nil || d.IsDir() {
			return nil
		}
		if isKnowledgeFile(d.Name()) {
			n++
		}
		return nil
	})
	return n
}
