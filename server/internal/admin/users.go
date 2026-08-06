package admin

import (
	"context"
	"crypto/subtle"
	"encoding/json"
	"log"
	"os"
	"sort"
	"strings"
	"time"

	"golang.org/x/crypto/bcrypt"

	cfgpkg "grounded_llm_server/internal/config"
	"grounded_llm_server/internal/store"
)

type userRecord struct {
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	PasswordBcrypt string   `json:"password_bcrypt"`
	Roles          []string `json:"roles"`
	TenantID       string   `json:"tenant_id,omitempty"`
}

type userSummary struct {
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
	TenantID string   `json:"tenant_id,omitempty"`
}

var userRegistry map[string]userRecord

// UsePostgresBackend reports whether admin users prefer Postgres.
func UsePostgresBackend() bool {
	mode := strings.ToLower(strings.TrimSpace(os.Getenv("ADMIN_USERS_STORE")))
	if mode == "file" || mode == "json" {
		return false
	}
	return st() != nil
}

// LoadUsers reads ADMIN_USERS_FILE and/or Postgres into the in-memory registry.
func LoadUsers(cfg *cfgpkg.Config) {
	_ = cfg
	userRegistry = make(map[string]userRecord)
	path := strings.TrimSpace(os.Getenv("ADMIN_USERS_FILE"))
	if path != "" {
		body, err := os.ReadFile(path)
		if err != nil {
			if !os.IsNotExist(err) {
				log.Printf("ADMIN_USERS_FILE read error: %v", err)
			}
		} else {
			var entries []userRecord
			if err := json.Unmarshal(body, &entries); err != nil {
				log.Printf("ADMIN_USERS_FILE parse error: %v", err)
			} else {
				for _, e := range entries {
					ingestUser(e)
				}
			}
		}
	}

	if UsePostgresBackend() {
		ctx, cancel := context.WithTimeout(context.Background(), 10*time.Second)
		defer cancel()
		rows, err := st().ListAdminUsers(ctx)
		if err != nil {
			log.Printf("admin_users load error: %v", err)
		} else {
			for _, u := range rows {
				ingestUser(userRecord{
					Username:       u.Username,
					PasswordBcrypt: u.PasswordBcrypt,
					Roles:          u.Roles,
					TenantID:       u.TenantID,
				})
			}
			if len(rows) > 0 {
				log.Printf("Admin users: %d from Postgres (+ optional JSON seed)", UserCount())
			}
		}
	}
}

func ingestUser(e userRecord) {
	user := strings.TrimSpace(e.Username)
	if user == "" {
		return
	}
	roles := NormalizeRoles(e.Roles)
	if len(roles) == 0 {
		log.Printf("admin user %q has no valid roles, skipping", user)
		return
	}
	userRegistry[user] = userRecord{
		Username:       user,
		Password:       e.Password,
		PasswordBcrypt: strings.TrimSpace(e.PasswordBcrypt),
		Roles:          roles,
		TenantID:       strings.TrimSpace(e.TenantID),
	}
}

// UserCount returns the number of users in the in-memory registry.
func UserCount() int {
	return len(userRegistry)
}

// PlaintextPasswordCount reports users still using a plaintext password
// instead of password_bcrypt (rejected by production validation).
func PlaintextPasswordCount() int {
	n := 0
	for _, rec := range userRegistry {
		if rec.PasswordBcrypt == "" && rec.Password != "" {
			n++
		}
	}
	return n
}

func authenticateUser(username, password string) ([]string, bool) {
	username = strings.TrimSpace(username)
	if username == "" || password == "" {
		return nil, false
	}
	if rec, ok := userRegistry[username]; ok {
		if verifyPassword(rec, password) {
			return rec.Roles, true
		}
		return nil, false
	}
	c := cfg()
	if c != nil && c.AdminPassword != "" &&
		username == c.AdminUser && password == c.AdminPassword {
		return []string{RoleAdmin}, true
	}
	return nil, false
}

func verifyPassword(rec userRecord, password string) bool {
	if rec.PasswordBcrypt != "" {
		return bcrypt.CompareHashAndPassword([]byte(rec.PasswordBcrypt), []byte(password)) == nil
	}
	if rec.Password != "" {
		return subtle.ConstantTimeCompare([]byte(rec.Password), []byte(password)) == 1
	}
	return false
}

func listUserSummaries() []userSummary {
	var out []userSummary
	for user, rec := range userRegistry {
		out = append(out, userSummary{Username: user, Roles: rec.Roles, TenantID: rec.TenantID})
	}
	c := cfg()
	if len(out) == 0 && c != nil && c.AdminPassword != "" {
		out = append(out, userSummary{Username: c.AdminUser, Roles: []string{RoleAdmin}})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Username < out[j].Username })
	return out
}

func persistAdminUserPG(rec userRecord) error {
	s := st()
	if s == nil {
		return nil
	}
	return s.UpsertAdminUser(context.Background(), store.AdminUser{
		Username:       rec.Username,
		PasswordBcrypt: rec.PasswordBcrypt,
		Roles:          rec.Roles,
		TenantID:       rec.TenantID,
	})
}
