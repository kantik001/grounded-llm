package admin

import (
	"encoding/json"
	"fmt"
	"io"
	"log"
	"net/http"
	"os"
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

var knowledgeFileExtensions = map[string]bool{
	".txt":  true,
	".pdf":  true,
	".docx": true,
}

func isKnowledgeFile(name string) bool {
	return knowledgeFileExtensions[strings.ToLower(filepath.Ext(name))]
}

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
	dir := tenant.KBDataDir(tid, domainID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusOK, gin.H{"success": true, "domain_id": domainID, "articles": []articleInfo{}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	chunkCounts, _ := fetchPythonIndexStats(tid, domainID)
	var articles []articleInfo
	for _, e := range entries {
		if e.IsDir() || !isKnowledgeFile(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		a := articleInfo{
			Filename:  e.Name(),
			SizeBytes: info.Size(),
			Modified:  info.ModTime().UTC().Format(time.RFC3339),
			Chunks:    chunkCounts[e.Name()],
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
	path := filepath.Join(tenant.KBDataDir(tid, domainID), name)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "File not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	log.Printf("Admin delete: %s", path)
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
	dir := tenant.KBDataDir(tid, domainID)
	if err := tenant.CheckDomainQuota(tid, domainID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error(), "code": "quota_exceeded"})
		return
	}
	if err := tenant.CheckStorageQuota(tid, fh.Size); err != nil {
		c.JSON(http.StatusRequestEntityTooLarge, gin.H{"success": false, "error": err.Error(), "code": "quota_exceeded"})
		return
	}
	if err := os.MkdirAll(dir, 0o755); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	dst := filepath.Join(dir, name)
	src, err := fh.Open()
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	defer func() { _ = src.Close() }()
	out, err := os.Create(dst)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	defer func() { _ = out.Close() }()
	if _, err := io.Copy(out, io.LimitReader(src, maxKnowledgeFileBytes+1)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if err := validateKnowledgeFileContent(dst, name); err != nil {
		_ = os.Remove(dst)
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	log.Printf("Admin upload: %s -> %s", name, dst)
	audit.Record(c, audit.Opts{
		Action:   store.AuditActionKBUpload,
		TenantID: tid,
		DomainID: domainID,
		Resource: name,
		Success:  true,
		Details:  map[string]any{"size_bytes": fh.Size},
	})
	c.JSON(http.StatusOK, gin.H{"success": true, "domain_id": domainID, "filename": name, "path": dst})
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
