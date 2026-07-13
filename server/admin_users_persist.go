package main

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"

	"golang.org/x/crypto/bcrypt"
)

func adminUsersFilePath() string {
	return strings.TrimSpace(os.Getenv("ADMIN_USERS_FILE"))
}

func saasProvisionAdmin() bool {
	return adminUsersFilePath() != ""
}

func signupAdminUsername(tenantID string) string {
	return normalizeTenantID(tenantID) + "-admin"
}

func generateAdminPassword() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

func provisionSignupAdminUser(tenantID string) (username, password string, err error) {
	path := adminUsersFilePath()
	if path == "" {
		return "", "", nil
	}

	username = signupAdminUsername(tenantID)
	password, err = generateAdminPassword()
	if err != nil {
		return "", "", err
	}
	hash, err := bcrypt.GenerateFromPassword([]byte(password), bcrypt.DefaultCost)
	if err != nil {
		return "", "", err
	}

	tenantRegistryMu.Lock()
	defer tenantRegistryMu.Unlock()

	var entries []adminUserRecord
	if body, readErr := os.ReadFile(path); readErr == nil {
		_ = json.Unmarshal(body, &entries)
	}
	for _, e := range entries {
		if strings.TrimSpace(e.Username) == username {
			return "", "", fmt.Errorf("admin user already exists for tenant")
		}
	}
	entries = append(entries, adminUserRecord{
		Username:       username,
		PasswordBcrypt: string(hash),
		Roles:          []string{RoleAdmin},
	})

	body, err := json.MarshalIndent(entries, "", "  ")
	if err != nil {
		return "", "", err
	}
	body = append(body, '\n')
	tmp := path + ".tmp"
	if err := os.WriteFile(tmp, body, 0o600); err != nil {
		return "", "", err
	}
	if err := os.Rename(tmp, path); err != nil {
		return "", "", err
	}

	adminUserRegistry[username] = adminUserRecord{
		Username:       username,
		PasswordBcrypt: string(hash),
		Roles:          []string{RoleAdmin},
	}
	return username, password, nil
}
