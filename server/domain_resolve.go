package main

import (
	"strings"

	"github.com/gin-gonic/gin"
)

func domainIDFromQuery(c *gin.Context) string {
	return strings.TrimSpace(c.Query("domain_id"))
}

func domainIDFromForm(c *gin.Context) string {
	return strings.TrimSpace(c.PostForm("domain_id"))
}
