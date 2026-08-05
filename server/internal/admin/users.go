package admin

import (
	"crypto/subtle"
	"encoding/json"
	"log"
	"os"
	"sort"
	"strings"

	"golang.org/x/crypto/bcrypt"

	cfgpkg "grounded_llm_server/internal/config"
)

type userRecord struct {
	Username       string   `json:"username"`
	Password       string   `json:"password"`
	PasswordBcrypt string   `json:"password_bcrypt"`
	Roles          []string `json:"roles"`
}

type userSummary struct {
	Username string   `json:"username"`
	Roles    []string `json:"roles"`
}

var userRegistry map[string]userRecord

// LoadUsers reads ADMIN_USERS_FILE into the in-memory registry.
func LoadUsers(cfg *cfgpkg.Config) {
	_ = cfg
	userRegistry = make(map[string]userRecord)
	path := strings.TrimSpace(os.Getenv("ADMIN_USERS_FILE"))
	if path == "" {
		return
	}
	body, err := os.ReadFile(path)
	if err != nil {
		log.Printf("ADMIN_USERS_FILE read error: %v", err)
		return
	}
	var entries []userRecord
	if err := json.Unmarshal(body, &entries); err != nil {
		log.Printf("ADMIN_USERS_FILE parse error: %v", err)
		return
	}
	for _, e := range entries {
		user := strings.TrimSpace(e.Username)
		if user == "" {
			continue
		}
		roles := NormalizeRoles(e.Roles)
		if len(roles) == 0 {
			log.Printf("ADMIN_USERS_FILE: user %q has no valid roles, skipping", user)
			continue
		}
		userRegistry[user] = userRecord{
			Username:       user,
			Password:       e.Password,
			PasswordBcrypt: strings.TrimSpace(e.PasswordBcrypt),
			Roles:          roles,
		}
	}
}

// UserCount returns the number of users loaded from ADMIN_USERS_FILE.
func UserCount() int {
	return len(userRegistry)
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
		out = append(out, userSummary{Username: user, Roles: rec.Roles})
	}
	c := cfg()
	if len(out) == 0 && c != nil && c.AdminPassword != "" {
		out = append(out, userSummary{Username: c.AdminUser, Roles: []string{RoleAdmin}})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Username < out[j].Username })
	return out
}
