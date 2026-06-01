package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"time"
)

// TelegramUser — данные пользователя из поля user в initData.
type TelegramUser struct {
	ID           int64  `json:"id"`
	FirstName    string `json:"first_name"`
	LastName     string `json:"last_name,omitempty"`
	Username     string `json:"username,omitempty"`
	LanguageCode string `json:"language_code,omitempty"`
}

// Проверяет HMAC-подпись и срок действия Telegram Web App initData.
func validateTelegramInitData(initData, botToken string, maxAge time.Duration) (*TelegramUser, error) {
	initData = strings.TrimSpace(initData)
	if initData == "" {
		return nil, fmt.Errorf("пустой initData")
	}
	if botToken == "" {
		return nil, fmt.Errorf("не задан TELEGRAM_BOT_TOKEN")
	}

	vals, err := url.ParseQuery(initData)
	if err != nil {
		return nil, fmt.Errorf("некорректный initData: %w", err)
	}

	receivedHash := vals.Get("hash")
	if receivedHash == "" {
		return nil, fmt.Errorf("в initData нет hash")
	}
	vals.Del("hash")

	var pairs []string
	for key := range vals {
		pairs = append(pairs, key+"="+vals.Get(key))
	}
	sort.Strings(pairs)
	dataCheckString := strings.Join(pairs, "\n")

	secretMAC := hmac.New(sha256.New, []byte("WebAppData"))
	secretMAC.Write([]byte(botToken))
	secretKey := secretMAC.Sum(nil)

	signMAC := hmac.New(sha256.New, secretKey)
	signMAC.Write([]byte(dataCheckString))
	calculatedHash := hex.EncodeToString(signMAC.Sum(nil))

	if !hmac.Equal([]byte(calculatedHash), []byte(receivedHash)) {
		return nil, fmt.Errorf("неверная подпись initData")
	}

	authDateStr := vals.Get("auth_date")
	if authDateStr == "" {
		return nil, fmt.Errorf("в initData нет auth_date")
	}
	authUnix, err := strconv.ParseInt(authDateStr, 10, 64)
	if err != nil {
		return nil, fmt.Errorf("некорректный auth_date: %w", err)
	}
	authTime := time.Unix(authUnix, 0)
	if maxAge > 0 && time.Since(authTime) > maxAge {
		return nil, fmt.Errorf("initData устарел")
	}

	userJSON := vals.Get("user")
	if userJSON == "" {
		return nil, fmt.Errorf("в initData нет user")
	}
	var user TelegramUser
	if err := json.Unmarshal([]byte(userJSON), &user); err != nil {
		return nil, fmt.Errorf("не удалось разобрать user: %w", err)
	}
	if user.ID == 0 {
		return nil, fmt.Errorf("некорректный user.id")
	}

	return &user, nil
}
