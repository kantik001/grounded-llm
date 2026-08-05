package rag

import "net/http"

func setPythonServiceHeaders(req *http.Request) {
	c := cfg()
	if c == nil {
		return
	}
	if c.RAGServiceToken != "" {
		req.Header.Set("X-RAG-Service-Token", c.RAGServiceToken)
	}
}
