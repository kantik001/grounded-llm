package main

import (
	"net/http"
	"sort"

	"github.com/gin-gonic/gin"
)

type apiKeySummary struct {
	Label string   `json:"label"`
	Roles []string `json:"roles"`
}

// GET /admin/api-keys — labels and roles only (never secret key material).
func handleAdminAPIKeys(c *gin.Context) {
	keys := listAPIKeySummaries()
	c.JSON(http.StatusOK, gin.H{
		"success": true,
		"keys":    keys,
		"users":   listAdminUserSummaries(),
		"roles":   allRoles,
	})
}

func listAPIKeySummaries() []apiKeySummary {
	out := make([]apiKeySummary, 0, len(apiKeyRegistry))
	for _, rec := range apiKeyRegistry {
		roles := rec.Roles
		if len(roles) == 0 {
			roles = defaultAPIKeyRoles()
		}
		out = append(out, apiKeySummary{Label: rec.Label, Roles: roles})
	}
	sort.Slice(out, func(i, j int) bool { return out[i].Label < out[j].Label })
	return out
}
