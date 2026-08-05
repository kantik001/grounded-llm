package domain

import "fmt"

// RequireRAGEnabled returns an error when text assistant is disabled for the domain.
func RequireRAGEnabled(domainID string) error {
	info, ok := Lookup(domainID)
	if !ok {
		return fmt.Errorf("unknown domain: %s", domainID)
	}
	if !info.RAGEnabled {
		return fmt.Errorf("text assistant is not available for domain %q", DisplayName(info, defaultLocale()))
	}
	return nil
}
