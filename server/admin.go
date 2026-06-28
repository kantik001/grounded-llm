package main

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

// Admin auth: OIDC session cookie and/or HTTP Basic (ADMIN_USER / ADMIN_PASSWORD).
func registerAdminRoutes(router *gin.Engine, cfg *Config) {
	registerOIDCAuthRoutes(router.Group("/admin/auth"))
	registerOIDCAuthRoutes(router.Group("/api/admin/auth"))
	auth := adminAuthMiddleware(cfg)
	registerAdminRouteGroup(router.Group("/admin"), auth)
	registerAdminRouteGroup(router.Group("/api/admin"), auth)
}

func registerAdminRouteGroup(g *gin.RouterGroup, auth gin.HandlerFunc) {
	g.Use(auth)

	g.GET("/status", handleAdminStatus)

	kb := g.Group("")
	kb.Use(requireAdminRoles(RoleKBEditor))
	kb.GET("/articles", handleAdminListArticles)
	kb.DELETE("/articles", handleAdminDeleteArticle)
	kb.POST("/upload", handleAdminUpload)
	kb.POST("/reindex", handleAdminReindex)
	kb.GET("/quotas", handleAdminQuotas)

	adminOnly := g.Group("")
	adminOnly.Use(requireAdminRoles(RoleAdmin))
	adminOnly.GET("/feedback", handleAdminFeedbackSummary)
	adminOnly.GET("/audit-log", handleAdminAuditLog)

	apiMgr := g.Group("")
	apiMgr.Use(requireAdminRoles(RoleAPIManager))
	apiMgr.GET("/api-keys", handleAdminAPIKeys)
}

// GET /admin/status: data_dir и число доменов.
func handleAdminStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"data_dir": config.DataDir,
		"domains":  len(domainCatalog.Domains),
		"roles":    adminRolesFromContext(c),
	})
}

type adminArticleInfo struct {
	Filename  string `json:"filename"`
	SizeBytes int64  `json:"size_bytes"`
	Modified  string `json:"modified"`
	Chunks    int    `json:"chunks"`
}

// GET /admin/articles: список документов с размером, датой и числом chunks в индексе.
func handleAdminListArticles(c *gin.Context) {
	domainID, err := normalizeDomainID(domainIDFromQuery(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	dir := kbDataDir(adminTenantID(c), domainID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusOK, gin.H{"success": true, "domain_id": domainID, "articles": []adminArticleInfo{}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	chunkCounts, _ := fetchPythonIndexStats(adminTenantID(c), domainID)
	var articles []adminArticleInfo
	for _, e := range entries {
		if e.IsDir() || !isKnowledgeFile(e.Name()) {
			continue
		}
		info, err := e.Info()
		if err != nil {
			continue
		}
		a := adminArticleInfo{
			Filename:  e.Name(),
			SizeBytes: info.Size(),
			Modified:  info.ModTime().UTC().Format(time.RFC3339),
			Chunks:    chunkCounts[e.Name()],
		}
		articles = append(articles, a)
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "domain_id": domainID, "articles": articles})
}

// DELETE /admin/articles?domain_id=&filename=
func handleAdminDeleteArticle(c *gin.Context) {
	domainID, err := normalizeDomainID(domainIDFromQuery(c))
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error()})
		return
	}
	name := filepath.Base(strings.TrimSpace(c.Query("filename")))
	if !safeFilename.MatchString(name) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Invalid filename"})
		return
	}
	path := filepath.Join(kbDataDir(adminTenantID(c), domainID), name)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "File not found"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	log.Printf("Admin delete: %s", path)
	recordAdminAudit(c, auditOpts{
		Action:   auditActionKBDelete,
		TenantID: adminTenantID(c),
		DomainID: domainID,
		Resource: name,
		Success:  true,
	})
	c.JSON(http.StatusOK, gin.H{"success": true, "domain_id": domainID, "filename": name, "reindex_recommended": true})
}

func handleAdminUpload(c *gin.Context) {
	domainID, err := normalizeDomainID(domainIDFromForm(c))
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
	dir := kbDataDir(adminTenantID(c), domainID)
	if err := checkDomainQuota(adminTenantID(c), domainID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": err.Error(), "code": "quota_exceeded"})
		return
	}
	if err := checkStorageQuota(adminTenantID(c), fh.Size); err != nil {
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
	defer src.Close()
	out, err := os.Create(dst)
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	defer out.Close()
	if _, err := io.Copy(out, io.LimitReader(src, maxKnowledgeFileBytes+1)); err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	if strings.EqualFold(filepath.Ext(name), ".txt") {
		body, err := os.ReadFile(dst)
		if err != nil || len(strings.TrimSpace(string(body))) == 0 {
			_ = os.Remove(dst)
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "TXT file is empty"})
			return
		}
	}
	log.Printf("Admin upload: %s -> %s", name, dst)
	recordAdminAudit(c, auditOpts{
		Action:   auditActionKBUpload,
		TenantID: adminTenantID(c),
		DomainID: domainID,
		Resource: name,
		Success:  true,
		Details:  map[string]any{"size_bytes": fh.Size},
	})
	c.JSON(http.StatusOK, gin.H{"success": true, "domain_id": domainID, "filename": name, "path": dst})
}

// POST /admin/reindex: запуск переиндексации Chroma в Python.
func handleAdminReindex(c *gin.Context) {
	if err := triggerRAGReindex(); err != nil {
		log.Printf("Admin reindex: %v", err)
		recordAdminAudit(c, auditOpts{
			Action:  auditActionKBReindex,
			Success: false,
			Details: map[string]any{"error": err.Error()},
		})
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}
	recordAdminAudit(c, auditOpts{
		Action:  auditActionKBReindex,
		Success: true,
	})
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "RAG reindex started"})
}

// Вызывает POST /admin/reindex на Python-сервисе с X-Admin-Secret.
func triggerRAGReindex() error {
	if config.AdminSecret == "" {
		return fmt.Errorf("ADMIN_SECRET is not set")
	}
	url := strings.TrimRight(config.PythonBaseURL, "/") + "/admin/reindex"
	req, err := http.NewRequest(http.MethodPost, url, nil)
	if err != nil {
		return err
	}
	req.Header.Set("X-Admin-Secret", config.AdminSecret)
	client := &http.Client{Timeout: 10 * time.Minute}
	resp, err := client.Do(req)
	if err != nil {
		return err
	}
	defer resp.Body.Close()
	body, _ := io.ReadAll(resp.Body)
	if resp.StatusCode != http.StatusOK {
		return fmt.Errorf("python reindex HTTP %d: %s", resp.StatusCode, string(body))
	}
	return nil
}

type pythonIndexStatsResponse struct {
	Success bool `json:"success"`
	Files   []struct {
		Filename string `json:"filename"`
		Chunks   int    `json:"chunks"`
	} `json:"files"`
}

func fetchPythonIndexStats(tenantID, domainID string) (map[string]int, error) {
	if config.AdminSecret == "" {
		return nil, fmt.Errorf("ADMIN_SECRET is not set")
	}
	url := strings.TrimRight(config.PythonBaseURL, "/") +
		"/admin/index-stats?domain_id=" + domainID + "&tenant_id=" + tenantID
	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return nil, err
	}
	req.Header.Set("X-Admin-Secret", config.AdminSecret)
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
