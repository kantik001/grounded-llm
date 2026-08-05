package auth

import "strings"

// Role names used by API keys (subset of app RBAC).
const (
	RoleChatOnly = "chat_only"
)

func normalizeRoles(in []string) []string {
	if len(in) == 0 {
		return nil
	}
	seen := make(map[string]struct{}, len(in))
	var out []string
	for _, raw := range in {
		r := normalizeRoleName(raw)
		if r == "" {
			continue
		}
		if _, ok := seen[r]; ok {
			continue
		}
		seen[r] = struct{}{}
		out = append(out, r)
	}
	return out
}

func normalizeRoleName(raw string) string {
	switch strings.ToLower(strings.TrimSpace(raw)) {
	case RoleChatOnly, "chat", "chat-only":
		return RoleChatOnly
	case "kb_editor", "kb", "editor", "kb-editor":
		return "kb_editor"
	case "admin":
		return "admin"
	case "api_manager", "api-manager":
		return "api_manager"
	default:
		return ""
	}
}

func defaultAPIKeyRoles() []string {
	return []string{RoleChatOnly}
}

// CanUseChatAPI reports whether API key roles may call chat endpoints.
func CanUseChatAPI(apiRoles []string) bool {
	for _, r := range apiRoles {
		switch r {
		case RoleChatOnly, "kb_editor", "admin", "api_manager":
			return true
		}
	}
	return false
}
