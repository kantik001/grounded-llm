package admin

import (
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"

	"golang.org/x/crypto/bcrypt"

	"grounded_llm_server/internal/tenant"
)

func adminUsersFilePath() string {
	return strings.TrimSpace(os.Getenv("ADMIN_USERS_FILE"))
}

// SaaSProvisionAdmin reports whether signup should auto-provision admin users.
func SaaSProvisionAdmin() bool {
	return adminUsersFilePath() != ""
}

func signupAdminUsername(tenantID string) string {
	return tenant.NormalizeTenantID(tenantID) + "-admin"
}

func generateAdminPassword() (string, error) {
	b := make([]byte, 12)
	if _, err := rand.Read(b); err != nil {
		return "", err
	}
	return hex.EncodeToString(b), nil
}

var adminUsersMu sync.Mutex

// ProvisionSignupAdminUser appends a tenant-scoped admin user to ADMIN_USERS_FILE.
func ProvisionSignupAdminUser(tenantID string) (username, password string, err error) {
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

	adminUsersMu.Lock()
	defer adminUsersMu.Unlock()

	var entries []userRecord
	if body, readErr := os.ReadFile(path); readErr == nil {
		_ = json.Unmarshal(body, &entries)
	}
	for _, e := range entries {
		if strings.TrimSpace(e.Username) == username {
			return "", "", fmt.Errorf("admin user already exists for tenant")
		}
	}
	entries = append(entries, userRecord{
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

	userRegistry[username] = userRecord{
		Username:       username,
		PasswordBcrypt: string(hash),
		Roles:          []string{RoleAdmin},
	}
	return username, password, nil
}
