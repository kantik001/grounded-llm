package main

import (
	"crypto/sha256"
	"encoding/binary"
	"encoding/json"
	"log"
	"os"
	"strings"
)

const (
	ctxKeyAPIKeyLabel = "api_key_label"
	ctxKeyAPIActorID  = "api_actor_id"
	headerAPIKey      = "X-API-Key"
)

type apiKeyEntry struct {
	Key   string `json:"key"`
	Label string `json:"label"`
}

var apiKeyRegistry map[string]string // key -> label

func loadAPIKeys(cfg *Config) {
	apiKeyRegistry = make(map[string]string)
	if path := strings.TrimSpace(os.Getenv("API_KEYS_FILE")); path != "" {
		body, err := os.ReadFile(path)
		if err != nil {
			log.Printf("API_KEYS_FILE read error: %v", err)
			return
		}
		var entries []apiKeyEntry
		if err := json.Unmarshal(body, &entries); err != nil {
			log.Printf("API_KEYS_FILE parse error: %v", err)
			return
		}
		for _, e := range entries {
			k := strings.TrimSpace(e.Key)
			if k != "" {
				apiKeyRegistry[k] = strings.TrimSpace(e.Label)
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
			apiKeyRegistry[key] = label
		}
	}
}

func lookupAPIKey(key string) (label string, ok bool) {
	label, ok = apiKeyRegistry[strings.TrimSpace(key)]
	return label, ok
}

func apiKeyActorID(key string) int64 {
	sum := sha256.Sum256([]byte(key))
	n := binary.BigEndian.Uint64(sum[:8]) & 0x7FFFFFFFFFFFFFFF
	return -int64(n)
}
