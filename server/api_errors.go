package main

import (
	"log"
	"strings"

	"github.com/gin-gonic/gin"
)

// publicAPIError возвращает безопасное сообщение для клиента; детали — в лог.
func publicAPIError(err error) string {
	if err == nil {
		return "Ошибка сервера"
	}
	s := strings.TrimSpace(err.Error())
	if s == "" {
		return "Ошибка сервера"
	}

	// Уже по-русски из normalizeDomainID и domain guards.
	if strings.Contains(s, "культура") ||
		strings.Contains(s, "изображение") ||
		strings.Contains(s, "файл") ||
		strings.Contains(s, "сессия") ||
		strings.Contains(s, "помощник") ||
		strings.Contains(s, "распознавание") ||
		strings.Contains(s, "Пустой вопрос") ||
		strings.Contains(s, "LLM_API_KEY") ||
		strings.Contains(s, "классификации") {
		return s
	}

	lower := strings.ToLower(s)
	switch {
	case strings.Contains(lower, "connection refused"),
		strings.Contains(lower, "timeout"),
		strings.Contains(lower, "no such host"),
		strings.Contains(lower, "rag request failed"),
		strings.Contains(lower, "classifier"):
		return "Сервис анализа временно недоступен. Попробуйте позже."
	case strings.Contains(lower, "unauthorized"),
		strings.Contains(lower, "telegram"):
		return "Ошибка авторизации. Откройте приложение из бота Telegram."
	default:
		log.Printf("publicAPIError (скрыта деталь): %v", err)
		return "Ошибка сервера"
	}
}

func jsonError(c *gin.Context, code int, err error) {
	if err != nil {
		log.Printf("%s %s: %v", c.Request.Method, c.Request.URL.Path, err)
	}
	c.JSON(code, gin.H{"success": false, "error": publicAPIError(err)})
}
