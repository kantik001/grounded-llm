package main

import "fmt"

func requireRAGEnabled(domainID string) error {
	info, ok := domainInfo(domainID)
	if !ok {
		return fmt.Errorf("неизвестный домен: %s", domainID)
	}
	if !info.RAGEnabled {
		return fmt.Errorf("для домена «%s» текстовый помощник пока недоступен", domainDisplayName(info))
	}
	return nil
}
