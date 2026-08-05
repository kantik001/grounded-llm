package oidc

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

// ResetProvider clears cached OIDC provider state (e.g. after config reload).
func ResetProvider() {
	oidcProviderMu.Lock()
	defer oidcProviderMu.Unlock()
	oidcProvider = nil
	oidcVerifier = nil
	oidcOAuth2Config = nil
	oidcProviderErr = nil
}

func ensureProvider(ctx context.Context) (*oidc.Provider, *oauth2.Config, *oidc.IDTokenVerifier, error) {
	if !Configured() {
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
	provider, err := oidc.NewProvider(ctx, envCfg.Issuer)
	if err != nil {
		oidcProviderErr = err
		return nil, nil, nil, err
	}
	oauth2Config := &oauth2.Config{
		ClientID:     envCfg.ClientID,
		ClientSecret: envCfg.ClientSecret,
		RedirectURL:  envCfg.RedirectURL,
		Endpoint:     provider.Endpoint(),
		Scopes:       envCfg.Scopes,
	}
	oidcProvider = provider
	oidcOAuth2Config = oauth2Config
	oidcVerifier = provider.Verifier(&oidc.Config{ClientID: envCfg.ClientID})
	return oidcProvider, oidcOAuth2Config, oidcVerifier, nil
}
