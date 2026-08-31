package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"path/filepath"
	"regexp"
	"strings"
	"time"

	"github.com/gin-gonic/gin"

	"grounded_llm_server/internal/audit"
	"grounded_llm_server/internal/domain"
	"grounded_llm_server/internal/store"
	"grounded_llm_server/internal/tenant"
)

var safeFilename = regexp.MustCompile(`^[a-zA-Z0-9._-]+\.(txt|pdf|docx)$`)

const maxKnowledgeFileBytes = 10 * 1024 * 1024

func handleStatus(c *gin.Context) {
	cCfg := cfg()
	dataDir := ""
	if cCfg != nil {
		dataDir = cCfg.DataDir
	}
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"data_dir": dataDir,
		"domains":  len(domain.Catalog().Domains),
		"roles":    RolesFromContext(c),
	})
}

type articleInfo struct {
	Filename  string `json:"filename"`
	SizeBytes int64  `json:"size_bytes"`
	Modified  string `json:"modified"`
	Chunks    int    `json:"chunks"`
}

func handleListArticles(c *gin.Context) {
	domainID, err := domain.NormalizeID(domain.IDFromQuery(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	tid := tenant.AdminID(c)
	s := st()
	if s == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Store not ready"})
		return
	}
	rows, err := s.ListKBArticles(c.Request.Context(), tid, domainID)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	chunkCounts, _ := fetchPythonIndexStats(tid, domainID)
	var articles []articleInfo
	for _, row := range rows {
		a := articleInfo{
			Filename:  row.LogicalKey,
			SizeBytes: row.SizeBytes,
			Modified:  row.UpdatedAt,
			Chunks:    chunkCounts[row.LogicalKey],
		}
		articles = append(articles, a)
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "domain_id": domainID, "articles": articles})
}

func handleDeleteArticle(c *gin.Context) {
	domainID, err := domain.NormalizeID(domain.IDFromQuery(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	name := filepath.Base(strings.TrimSpace(c.Query("filename")))
	if !safeFilename.MatchString(name) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid filename"})
		return
	}
	tid := tenant.AdminID(c)
	s := st()
	if s == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "Store not ready"})
		return
	}
	ok, err := s.MarkKBDocumentDeleted(c.Request.Context(), tid, domainID, name)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if !ok {
		c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Document not found"})
		return
	}
	log.Printf("Admin delete: tenant=%s domain=%s file=%s", tid, domainID, name)
	audit.Record(c, audit.Opts{
		Action:   store.AuditActionKBDelete,
		TenantID: tid,
		DomainID: domainID,
		Resource: name,
		Success:  true,
	})
	c.JSON(http.StatusOK, gin.H{"success": true, "domain_id": domainID, "filename": name, "reindex_recommended": true})
}

func handleUpload(c *gin.Context) {
	domainID, err := domain.NormalizeID(domain.IDFromForm(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	fh, err := c.FormFile("file")
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "A .txt, .pdf, or .docx file is required"})
		return
	}
	name := filepath.Base(fh.Filename)
	if !safeFilename.MatchString(name) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Filename: Latin letters, digits, extension .txt/.pdf/.docx"})
		return
	}
	if fh.Size > maxKnowledgeFileBytes {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Max file size is 10 MB"})
		return
	}
	tid := tenant.AdminID(c)
	s := st()
	if s == nil || kbBlob() == nil {
		c.JSON(http.StatusServiceUnavailable, gin.H{"success": false, "error": "KB registry not ready"})
		return
	}
	if err := tenant.CheckDomainQuota(c.Request.Context(), tid, domainID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error(), "code": "quota_exceeded"})
		return
	}
	if err := tenant.CheckStorageQuota(c.Request.Context(), tid, fh.Size); err != nil {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"success": false, "error": err.Error(), "code": "quota_exceeded"})
		return
	}
	src, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	defer func() { _ = src.Close() }()
	data, err := io.ReadAll(io.LimitReader(src, maxKnowledgeFileBytes+1))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if len(data) > maxKnowledgeFileBytes {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Max file size is 10 MB"})
		return
	}
	if err := validateKnowledgeFileBytes(data, name); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}

	doc, ver, regErr := registerKBUpload(c.Request.Context(), s, tid, domainID, name, audit.ActorFromContext(c), data)
	if regErr != nil {
		log.Printf("KB registry upload: %v", regErr)
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": "Failed to register document"})
		return
	}
	docID := doc.ID
	versionID := ver.ID

	log.Printf("Admin upload: %s -> registry (tenant=%s domain=%s)", name, tid, domainID)
	audit.Record(c, audit.Opts{
		Action:   store.AuditActionKBUpload,
		TenantID: tid,
		DomainID: domainID,
		Resource: name,
		Success:  true,
		Details:  map[string]any{"size_bytes": fh.Size},
	})
	c.JSON(http.StatusOK, gin.H{
		"success":     true,
		"domain_id":   domainID,
		"filename":    name,
		"document_id": docID,
		"version_id":  versionID,
		"reindex_recommended": true,
	})
}

type pythonIndexStatsResponse struct {
	Success bool `json:"success"`
	Files   []struct {
		Filename string `json:"filename"`
		Chunks   int    `json:"chunks"`
	} `json:"files"`
}

func fetchPythonIndexStats(tenantID, domainID string) (map[string]int, error) {
	c := cfg()
	if c == nil || c.AdminSecret == "" {
		return nil, fmt.Errorf("ADMIN_SECRET is not set")
	}
	url := strings.TrimRight(c.PythonBaseURL, "/") +
		"/admin/index-stats?domain_id=" + domainID + "&tenant_id=" + tenantID
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Admin-Secret", c.AdminSecret)
	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := client.Do(req)
	if err != nil {
		return nil, err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("index-stats HTTP %d", resp.StatusCode)
	}
	var out pythonIndexStatsResponse
	if err := json.Unmarshal(body, &out); err != nil {
		return nil, err
	}
	counts := make(map[string]int)
	for _, f := range out.Files {
		counts[f.Filename] = f.Chunks
	}
	return counts, nil
}

// SafeFilename matches allowed knowledge-base upload names (for tests).
func SafeFilename() *regexp.Regexp {
	return safeFilename
}
