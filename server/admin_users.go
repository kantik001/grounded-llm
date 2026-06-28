package main

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"os"
	"sort"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

type adminUserRecord struct {
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	PasswordBcrypt string   `json:"password_bcrypt"`
	Roles          []string `json:"roles"`
}

type adminUserSummary struct {
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

var adminUserRegistry map[string]adminUserRecord

func loadAdminUsers(cfg *Config) {
	adminUserRegistry = make(map[string]adminUserRecord)
	path := strings.TrimSpace(os.Getenv("ADMIN_USERS_FILE"))
	if path == "" {
		return
	}
	body, err := os.ReadFile(path)
	if err != nil {
		log.Printf("ADMIN_USERS_FILE read error: %v", err)
		return
	}
	var entries []adminUserRecord
	if err := json.Unmarshal(body, &entries); err != nil {
		log.Printf("ADMIN_USERS_FILE parse error: %v", err)
		return
	}
	for _, e := range entries {
		user := strings.TrimSpace(e.Username)
		if user == "" {
			continue
		}
		roles := normalizeRoles(e.Roles)
		if len(roles) == 0 {
			log.Printf("ADMIN_USERS_FILE: user %q has no valid roles, skipping", user)
			continue
		}
		adminUserRegistry[user] = adminUserRecord{
			Username:       user,
			Password:       e.Password,
			PasswordBcrypt: strings.TrimSpace(e.PasswordBcrypt),
			Roles:          roles,
		}
	}
}

func authenticateAdminUser(username, password string) ([]string, bool) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, false
	}
	if rec, ok := adminUserRegistry[username]; ok {
		if verifyAdminPassword(rec, password) {
			return rec.Roles, true
		}
		return nil, false
	}
	if config != nil && config.AdminPassword != "" &&
		username == config.AdminUser && password == config.AdminPassword {
		return []string{RoleAdmin}, true
	}
	return nil, false
}

func verifyAdminPassword(rec adminUserRecord, password string) bool {
	if rec.PasswordBcrypt != "" {
		return bcrypt.CompareHashAndPassword([]byte(rec.PasswordBcrypt), []byte(password)) == nil
	}
	if rec.Password != "" {
		return subtle.ConstantTimeCompare([]byte(rec.Password), []byte(password)) == 1
	}
	return false
}

func listAdminUserSummaries() []adminUserSummary {
	var out []adminUserSummary
	for user, rec := range adminUserRegistry {
		out = append(out, adminUserSummary{Username: user, Roles: rec.Roles})
	}
	if len(out) == 0 && config != nil && config.AdminPassword != "" {
		out = append(out, adminUserSummary{Username: config.AdminUser, Roles: []string{RoleAdmin}})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Username < out[j].Username })
	return out
}
