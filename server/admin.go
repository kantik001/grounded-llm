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

// Basic Auth для маршрутов /admin (ADMIN_USER / ADMIN_PASSWORD).
func adminBasicAuth(cfg *Config) gin.HandlerFunc {
	return func(c *gin.Context) {
		if cfg.AdminPassword == "" {
			c.AbortWithStatusJSON(http.StatusServiceUnavailable, gin.H{
				"success": false,
				"error":   "Админка отключена: задайте ADMIN_PASSWORD в .env",
			})
			return
		}
		user, pass, ok := c.Request.BasicAuth()
		if !ok || user != cfg.AdminUser || pass != cfg.AdminPassword {
			c.Header("WWW-Authenticate", `Basic realm="Grounded LLM Admin"`)
			c.AbortWithStatus(http.StatusUnauthorized)
			return
		}
		c.Next()
	}
}

// Регистрирует админские маршруты: upload, reindex RAG.
func registerAdminRoutes(router *gin.Engine, cfg *Config) {
	auth := adminBasicAuth(cfg)
	registerAdminRouteGroup(router.Group("/admin"), auth)
	registerAdminRouteGroup(router.Group("/api/admin"), auth)
}

func registerAdminRouteGroup(g *gin.RouterGroup, auth gin.HandlerFunc) {
	g.Use(auth)
	g.GET("/status", handleAdminStatus)
	g.GET("/articles", handleAdminListArticles)
	g.DELETE("/articles", handleAdminDeleteArticle)
	g.POST("/upload", handleAdminUpload)
	g.POST("/reindex", handleAdminReindex)
}

// GET /admin/status: data_dir и число доменов.
func handleAdminStatus(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{
		"success":  true,
		"data_dir": config.DataDir,
		"domains":  len(domainCatalog.Domains),
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
	dir := filepath.Join(config.DataDir, domainID)
	entries, err := os.ReadDir(dir)
	if err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusOK, gin.H{"success": true, "domain_id": domainID, "articles": []adminArticleInfo{}})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	chunkCounts, _ := fetchPythonIndexStats(domainID)
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
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Некорректное имя файла"})
		return
	}
	path := filepath.Join(config.DataDir, domainID, name)
	if err := os.Remove(path); err != nil {
		if os.IsNotExist(err) {
			c.JSON(http.StatusNotFound, gin.H{"success": false, "error": "Файл не найден"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"success": false, "error": err.Error()})
		return
	}
	log.Printf("Admin delete: %s", path)
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
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Нужен файл .txt, .pdf или .docx"})
		return
	}
	name := filepath.Base(fh.Filename)
	if !safeFilename.MatchString(name) {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Имя файла: латиница, цифры, расширение .txt/.pdf/.docx"})
		return
	}
	if fh.Size > maxKnowledgeFileBytes {
		c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "Макс. размер файла 10 МБ"})
		return
	}
	dir := filepath.Join(config.DataDir, domainID)
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
			c.JSON(http.StatusBadRequest, gin.H{"success": false, "error": "TXT-файл пустой"})
			return
		}
	}
	log.Printf("Admin upload: %s -> %s", name, dst)
	c.JSON(http.StatusOK, gin.H{"success": true, "domain_id": domainID, "filename": name, "path": dst})
}

// POST /admin/reindex: запуск переиндексации Chroma в Python.
func handleAdminReindex(c *gin.Context) {
	if err := triggerRAGReindex(); err != nil {
		log.Printf("Admin reindex: %v", err)
		c.JSON(http.StatusBadGateway, gin.H{"success": false, "error": err.Error()})
		return
	}
	c.JSON(http.StatusOK, gin.H{"success": true, "message": "Переиндексация RAG запущена"})
}

// Вызывает POST /admin/reindex на Python-сервисе с X-Admin-Secret.
func triggerRAGReindex() error {
	if config.AdminSecret == "" {
		return fmt.Errorf("ADMIN_SECRET не задан")
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

func fetchPythonIndexStats(domainID string) (map[string]int, error) {
	if config.AdminSecret == "" {
		return nil, fmt.Errorf("ADMIN_SECRET не задан")
	}
	url := strings.TrimRight(config.PythonBaseURL, "/") + "/admin/index-stats?domain_id=" + domainID
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
