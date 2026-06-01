package main

import (
	"crypto/hmac"
	"crypto/sha256"
	"encoding/hex"
	"net/url"
	"sort"
	"strconv"
	"strings"
	"testing"
	"time"
)

// buildTestInitData собирает подписанный initData для unit-тестов (как Telegram).
func buildTestInitData(botToken string, authDate int64, userJSON string) string {
	vals := url.Values{}
	vals.Set("auth_date", strconv.FormatInt(authDate, 10))
	vals.Set("user", userJSON)

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
	vals.Set("hash", hex.EncodeToString(signMAC.Sum(nil)))

	return vals.Encode()
}

func TestValidateTelegramInitData_OK(t *testing.T) {
	token := "123456:ABC-DEF1234ghIkl-zyx57W2v1u123ew11"
	user := `{"id":42,"first_name":"Test"}`
	initData := buildTestInitData(token, time.Now().Unix(), user)

	u, err := validateTelegramInitData(initData, token, 24*time.Hour)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if u.ID != 42 {
		t.Fatalf("expected user id 42, got %d", u.ID)
	}
}

func TestValidateTelegramInitData_BadHash(t *testing.T) {
	_, err := validateTelegramInitData("auth_date=1&hash=deadbeef&user=%7B%22id%22%3A1%7D", "token", time.Hour)
	if err == nil {
		t.Fatal("expected error for bad hash")
	}
}
