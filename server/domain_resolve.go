package main

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// coalesceDomainID returns the first non-empty domain/crop id (legacy API alias).
func coalesceDomainID(values ...string) string {
	for _, v := range values {
		if s := strings.TrimSpace(v); s != "" {
			return s
		}
	}
	return ""
}

func domainIDFromQuery(c *gin.Context) string {
	return coalesceDomainID(c.Query("domain_id"), c.Query("crop_id"))
}

func domainIDFromForm(c *gin.Context) string {
	return coalesceDomainID(c.PostForm("domain_id"), c.PostForm("crop_id"))
}
