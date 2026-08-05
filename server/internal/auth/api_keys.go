package auth

import (
	"crypto/sha256"
	"encoding/binary"
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
)

type KeyRecord struct {
	Label string
	Roles []string
}

type apiKeyFileEntry struct {
	Key   string   `json:"key"`
	Label string   `json:"label"`
	Roles []string `json:"roles"`
}

var Registry map[string]KeyRecord

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
			Registry[k] = KeyRecord{
				Label: strings.TrimSpace(e.Label),
				Roles: roles,
			}
		}
		return
	}
	raw := strings.TrimSpace(os.Getenv("API_KEYS"))
	if raw == "" {
		return
	}
	for _, part := range strings.Split(raw, ",") {
		part = strings.TrimSpace(part)
		if part == "" {
			continue
		}
		label := part
		key := part
		if i := strings.Index(part, ":"); i > 0 {
			key = strings.TrimSpace(part[:i])
			label = strings.TrimSpace(part[i+1:])
		}
		if key != "" {
			Registry[key] = KeyRecord{Label: label, Roles: defaultAPIKeyRoles()}
		}
	}
}

func Lookup(key string) (KeyRecord, bool) {
	rec, ok := Registry[strings.TrimSpace(key)]
	return rec, ok
}

func ActorID(key string) int64 {
	sum := sha256.Sum256([]byte(key))
	n := binary.BigEndian.Uint64(sum[:8]) & 0x7FFFFFFFFFFFFFFF
	return -int64(n)
}
