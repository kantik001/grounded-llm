package admin

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"os"
	"strings"
	"sync"
	"time"

	"golang.org/x/crypto/bcrypt"

	"grounded_llm_server/internal/tenant"
)

func adminUsersFilePath() string {
	return strings.TrimSpace(os.Getenv("ADMIN_USERS_FILE"))
}

// SaaSProvisionAdmin reports whether signup should auto-provision admin users.
func SaaSProvisionAdmin() bool {
	return adminUsersFilePath() != "" || UsePostgresBackend()
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

// ProvisionSignupAdminUser creates a tenant-scoped admin user in Postgres and/or ADMIN_USERS_FILE.
func ProvisionSignupAdminUser(tenantID string) (username, password string, err error) {
	path := adminUsersFilePath()
	pg := UsePostgresBackend()
	if path == "" && !pg {
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
	rec := userRecord{
		Username:       username,
		PasswordBcrypt: string(hash),
		Roles:          []string{RoleAdmin},
		TenantID:       tenant.NormalizeTenantID(tenantID),
	}

	adminUsersMu.Lock()
	defer adminUsersMu.Unlock()

	if _, exists := userRegistry[username]; exists {
		return "", "", fmt.Errorf("admin user already exists for tenant")
	}
	if pg {
		ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
		defer cancel()
		ok, err := st().AdminUserExists(ctx, username)
		if err != nil {
			return "", "", err
		}
		if ok {
			return "", "", fmt.Errorf("admin user already exists for tenant")
		}
		if err := persistAdminUserPG(rec); err != nil {
			return "", "", err
		}
	}

	if path != "" {
		var entries []userRecord
		if body, readErr := os.ReadFile(path); readErr == nil {
			_ = json.Unmarshal(body, &entries)
		}
		for _, e := range entries {
			if strings.TrimSpace(e.Username) == username {
				return "", "", fmt.Errorf("admin user already exists for tenant")
			}
		}
		entries = append(entries, rec)
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
	}

	if userRegistry == nil {
		userRegistry = make(map[string]userRecord)
	}
	userRegistry[username] = rec
	return username, password, nil
}
