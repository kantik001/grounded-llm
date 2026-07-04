package main

import (
	"net/http"
	"strings"
)

// pythonServiceBaseURL returns PYTHON_BASE_URL without trailing slash.
func pythonServiceBaseURL() string {
	return strings.TrimRight(config.PythonBaseURL, "/")
}

// setPythonServiceHeaders attaches optional internal auth for Go → Python calls.
func setPythonServiceHeaders(req *http.Request) {
	if config.RAGServiceToken != "" {
		req.Header.Set("X-RAG-Service-Token", config.RAGServiceToken)
	}
}
