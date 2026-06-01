package main

import "fmt"

func requireRAGEnabled(domainID string) error {
	info, ok := domainInfo(domainID)
	if !ok {
		return fmt.Errorf("unknown domain: %s", domainID)
	}
	if !info.RAGEnabled {
		return fmt.Errorf("text assistant is not available for domain %q", domainDisplayName(info, appDefaultLocale))
	}
	return nil
}
