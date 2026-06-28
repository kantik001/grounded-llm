package main

import (
	"context"
	"fmt"
	"sync"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var (
	oidcProvider     *oidc.Provider
	oidcVerifier     *oidc.IDTokenVerifier
	oidcOAuth2Config *oauth2.Config
	oidcProviderErr  error
	oidcProviderMu   sync.Mutex
)

func resetOIDCProvider() {
	oidcProviderMu.Lock()
	defer oidcProviderMu.Unlock()
	oidcProvider = nil
	oidcVerifier = nil
	oidcOAuth2Config = nil
	oidcProviderErr = nil
}

func ensureOIDCProvider(ctx context.Context) (*oidc.Provider, *oauth2.Config, *oidc.IDTokenVerifier, error) {
	if !oidcConfigured() {
		return nil, nil, nil, fmt.Errorf("oidc not configured")
	}
	oidcProviderMu.Lock()
	defer oidcProviderMu.Unlock()
	if oidcProvider != nil && oidcOAuth2Config != nil && oidcVerifier != nil {
		return oidcProvider, oidcOAuth2Config, oidcVerifier, nil
	}
	if oidcProviderErr != nil {
		return nil, nil, nil, oidcProviderErr
	}
	provider, err := oidc.NewProvider(ctx, oidcCfg.Issuer)
	if err != nil {
		oidcProviderErr = err
		return nil, nil, nil, err
	}
	oauth2Config := &oauth2.Config{
		ClientID:     oidcCfg.ClientID,
		ClientSecret: oidcCfg.ClientSecret,
		RedirectURL:  oidcCfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       oidcCfg.Scopes,
	}
	oidcProvider = provider
	oidcOAuth2Config = oauth2Config
	oidcVerifier = provider.Verifier(&oidc.Config{ClientID: oidcCfg.ClientID})
	return oidcProvider, oidcOAuth2Config, oidcVerifier, nil
}
