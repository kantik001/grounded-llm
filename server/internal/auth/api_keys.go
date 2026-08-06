package auth

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/hex"
	"encoding/json"
	"log"
	"os"
	"strings"

	"grounded_llm_server/internal/config"
)

const (
	CtxKeyAPIKeyLabel = "api_key_label"
	CtxKeyAPIActorID  = "api_actor_id"
	HeaderAPIKey      = "X-API-Key"

	// hashedKeyPrefix marks a registry entry that is already a SHA-256 digest.
	hashedKeyPrefix = "sha256:"
)

type KeyRecord struct {
	Label string
	Roles []string
	// Tenant the key is bound to. Empty = default tenant; "*" = any allowlisted tenant.
	Tenant string
}

type apiKeyFileEntry struct {
	Key    string   `json:"key"`
	Label  string   `json:"label"`
	Roles  []string `json:"roles"`
	Tenant string   `json:"tenant"`
}

// Registry maps SHA-256 hex digests of API keys to their records.
// Plaintext keys are hashed at load time and never kept in memory.
var Registry map[string]KeyRecord

// HashKey returns the lowercase SHA-256 hex digest used as the registry key.
func HashKey(key string) string {
	sum := sha256.Sum256([]byte(strings.TrimSpace(key)))
	return hex.EncodeToString(sum[:])
}

// registryKeyFor accepts a plaintext key or a "sha256:<hex>" digest entry.
func registryKeyFor(configured string) string {
	configured = strings.TrimSpace(configured)
	if rest, ok := strings.CutPrefix(configured, hashedKeyPrefix); ok {
		return strings.ToLower(strings.TrimSpace(rest))
	}
	return HashKey(configured)
}

func LoadAPIKeys(cfg *config.Config) {
	_ = cfg
	Registry = make(map[string]KeyRecord)
	if path := strings.TrimSpace(os.Getenv("API_KEYS_FILE")); path != "" {
		body, err := os.ReadFile(path)
		if err != nil {
			log.Printf("API_KEYS_FILE read error: %v", err)
			return
		}
		var entries []apiKeyFileEntry
		if err := json.Unmarshal(body, &entries); err != nil {
			log.Printf("API_KEYS_FILE parse error: %v", err)
			return
		}
		for _, e := range entries {
			k := strings.TrimSpace(e.Key)
			if k == "" {
				continue
			}
			roles := normalizeRoles(e.Roles)
			if len(roles) == 0 {
				roles = defaultAPIKeyRoles()
			}
			Registry[registryKeyFor(k)] = KeyRecord{
				Label:  strings.TrimSpace(e.Label),
				Roles:  roles,
				Tenant: strings.TrimSpace(e.Tenant),
			}
		}
		return
	}
	raw := strings.TrimSpace(os.Getenv("API_KEYS"))
	if raw == "" {
		return
	}
	// API_KEYS format: key[:label[:tenant]],... — tenant "*" allows any allowlisted tenant.
	// Keys may be pre-hashed as sha256:<hex>.
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		hashedPrefix := ""
		if strings.HasPrefix(part, hashedKeyPrefix) {
			hashedPrefix = hashedKeyPrefix
			part = strings.TrimPrefix(part, hashedKeyPrefix)
		}
		fields := strings.SplitN(part, ":", 3)
		key := hashedPrefix + strings.TrimSpace(fields[0])
		if strings.TrimSpace(fields[0]) == "" {
			continue
		}
		label := key
		tenantID := ""
		if len(fields) > 1 && strings.TrimSpace(fields[1]) != "" {
			label = strings.TrimSpace(fields[1])
		}
		if len(fields) > 2 {
			tenantID = strings.TrimSpace(fields[2])
		}
		Registry[registryKeyFor(key)] = KeyRecord{Label: label, Roles: defaultAPIKeyRoles(), Tenant: tenantID}
	}
}

// Lookup finds a record by the presented plaintext key (compared via SHA-256 digest).
func Lookup(key string) (KeyRecord, bool) {
	rec, ok := Registry[HashKey(key)]
	return rec, ok
}

func ActorID(key string) int64 {
	sum := sha256.Sum256([]byte(key))
	n := binary.BigEndian.Uint64(sum[:8]) & 0x7FFFFFFFFFFFFFFF
	return -int64(n)
}
