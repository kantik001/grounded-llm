package httpapi

import (
	_ "embed"
	"net/http"

	"github.com/gin-gonic/gin"
)

//go:embed openapi.v1.json
var openAPIV1 []byte

// OpenAPI serves the embedded OpenAPI v1 document.
func OpenAPI(c *gin.Context) {
	c.Data(http.StatusOK, "application/json; charset=utf-8", openAPIV1)
}
