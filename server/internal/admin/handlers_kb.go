package admin

import (
	"net/http"
	"os"
	"strings"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/domain"
	"grounded_llm_server/internal/store"
	"grounded_llm_server/internal/tenant"
)

func handleListKBDocuments(c *gin.Context) {
	s := st()
	if s == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Database unavailable"})
		return
	}
	domainID, err := domain.NormalizeID(domain.IDFromQuery(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	tenantID := tenant.AdminID(c)
	docs, err := s.ListKBDocuments(c.Request.Context(), tenantID, domainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "domain_id": domainID, "tenant_id": tenantID, "documents": docs})
}

type rebuildIndexRequest struct {
	Backend         string `json:"backend"`
	EmbeddingModel  string `json:"embedding_model"`
	Activate        bool   `json:"activate"`
}

func handleRebuildIndexRun(c *gin.Context) {
	s := st()
	if s == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Database unavailable"})
		return
	}
	domainID, err := domain.NormalizeID(domain.IDFromQuery(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	tenantID := tenant.AdminID(c)
	var body rebuildIndexRequest
	_ = c.ShouldBindJSON(&body)
	backend := strings.TrimSpace(body.Backend)
	if backend == "" {
		backend = strings.ToLower(strings.TrimSpace(osGetenv("VECTOR_STORE", "chroma")))
	}
	model := strings.TrimSpace(body.EmbeddingModel)
	if model == "" {
		model = osGetenv("EMBEDDING_MODEL", "intfloat/multilingual-e5-small")
	}
	run, err := s.CreateIndexRun(c.Request.Context(), tenantID, domainID, backend, model, store.DefaultChunkSchema())
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if body.Activate {
		if err := s.ActivateIndexRun(c.Request.Context(), tenantID, domainID, run.ID); err != nil {
			c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
			return
		}
		run.Status = store.IndexRunStatusActive
	}
	c.JSON(http.StatusAccepted, gin.H{
		"success":     true,
		"index_run":   run,
		"message":     "Index run created (building). Run full ingest with index_run_id, then activate (or set activate_on_complete on ingest).",
	})
}

func osGetenv(key, def string) string {
	if v := strings.TrimSpace(os.Getenv(key)); v != "" {
		return v
	}
	return def
}
