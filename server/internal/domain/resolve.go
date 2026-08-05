package domain

import (
	"strings"

	"github.com/gin-gonic/gin"
)

// IDFromQuery reads domain_id from the query string.
func IDFromQuery(c *gin.Context) string {
	return strings.TrimSpace(c.Query("domain_id"))
}

// IDFromForm reads domain_id from form data.
func IDFromForm(c *gin.Context) string {
	return strings.TrimSpace(c.PostForm("domain_id"))
}
