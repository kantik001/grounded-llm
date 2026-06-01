package main

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

// publicAPIError returns a safe client-facing message; details go to the log.
func publicAPIError(err error) string {
	if err == nil {
		return "Server error"
	}
	s := strings.TrimSpace(err.Error())
	if s == "" {
		return "Server error"
	}

	// Known user-safe messages from domain guards, RAG, auth, uploads, etc.
	if isPublicClientMessage(s) {
		return s
	}

	lower := strings.ToLower(s)
	switch {
	case strings.Contains(lower, "connection refused"),
		strings.Contains(lower, "timeout"),
		strings.Contains(lower, "no such host"),
		strings.Contains(lower, "rag request failed"),
		strings.Contains(lower, "python reindex"):
		return "Analysis service is temporarily unavailable. Please try again later."
	case strings.Contains(lower, "unauthorized"),
		strings.Contains(lower, "telegram"):
		return "Authentication failed. Open the app from the Telegram bot."
	default:
		log.Printf("publicAPIError (detail hidden): %v", err)
		return "Server error"
	}
}

func isPublicClientMessage(s string) bool {
	lower := strings.ToLower(s)
	prefixes := []string{
		"unknown domain", "text assistant is not available", "unknown tenant",
		"empty question", "llm_api_key", "image upload", "image too large",
		"invalid", "required", "not found", "too many requests",
		"telegram", "api key", "admin", "session", "rating", "file",
		"domain", "tenant", "multipart", "text required", "storage",
	}
	for _, p := range prefixes {
		if strings.Contains(lower, p) {
			return true
		}
	}
	return strings.Contains(s, "LLM_API_KEY")
}

func jsonError(c *gin.Context, code int, err error) {
	if err != nil {
		log.Printf("%s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	}
	c.JSON(code, gin.H{"success": false, "error": publicAPIError(err)})
}
