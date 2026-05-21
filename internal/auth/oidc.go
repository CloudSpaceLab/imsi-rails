package auth

import (
	"context"
	"errors"

	"github.com/coreos/go-oidc/v3/oidc"
	"golang.org/x/oauth2"
)

var ErrOIDCConfigMissing = errors.New("google oidc configuration is missing")

type OIDCRedirect struct {
	URL   string `json:"url"`
	State string `json:"state"`
}

type OIDCProfile struct {
	Subject     string `json:"subject"`
	Email       string `json:"email"`
	DisplayName string `json:"display_name"`
}

type OIDCAuthenticator interface {
	Start(ctx context.Context, provider AuthProviderConfig, returnTo string) (OIDCRedirect, error)
	Complete(ctx context.Context, provider AuthProviderConfig, state, code string) (OIDCProfile, error)
}

type GoogleOIDCAuthenticator struct {
	states map[string]string
}

func NewGoogleOIDCAuthenticator() *GoogleOIDCAuthenticator {
	return &GoogleOIDCAuthenticator{states: map[string]string{}}
}

func (a *GoogleOIDCAuthenticator) Start(ctx context.Context, provider AuthProviderConfig, returnTo string) (OIDCRedirect, error) {
	if provider.GoogleOIDC == nil {
		return OIDCRedirect{}, ErrOIDCConfigMissing
	}
	state, err := randomToken(24)
	if err != nil {
		return OIDCRedirect{}, err
	}
	config := oauth2.Config{
		ClientID:     provider.GoogleOIDC.ClientID,
		ClientSecret: provider.GoogleOIDC.ClientSecret,
		RedirectURL:  provider.GoogleOIDC.RedirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
		Endpoint:     googleEndpoint(provider.GoogleOIDC.Issuer),
	}
	a.states[state] = returnTo
	return OIDCRedirect{URL: config.AuthCodeURL(state), State: state}, nil
}

func (a *GoogleOIDCAuthenticator) Complete(ctx context.Context, provider AuthProviderConfig, state, code string) (OIDCProfile, error) {
	if provider.GoogleOIDC == nil {
		return OIDCProfile{}, ErrOIDCConfigMissing
	}
	if _, ok := a.states[state]; !ok {
		return OIDCProfile{}, ErrInvalidCredentials
	}
	delete(a.states, state)

	oidcProvider, err := oidc.NewProvider(ctx, provider.GoogleOIDC.Issuer)
	if err != nil {
		return OIDCProfile{}, err
	}
	config := oauth2.Config{
		ClientID:     provider.GoogleOIDC.ClientID,
		ClientSecret: provider.GoogleOIDC.ClientSecret,
		RedirectURL:  provider.GoogleOIDC.RedirectURL,
		Scopes:       []string{oidc.ScopeOpenID, "email", "profile"},
		Endpoint:     oidcProvider.Endpoint(),
	}
	token, err := config.Exchange(ctx, code)
	if err != nil {
		return OIDCProfile{}, err
	}
	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return OIDCProfile{}, ErrInvalidCredentials
	}
	verifier := oidcProvider.Verifier(&oidc.Config{ClientID: provider.GoogleOIDC.ClientID})
	idToken, err := verifier.Verify(ctx, rawIDToken)
	if err != nil {
		return OIDCProfile{}, err
	}
	var claims struct {
		Email         string `json:"email"`
		Name          string `json:"name"`
		EmailVerified bool   `json:"email_verified"`
	}
	if err := idToken.Claims(&claims); err != nil {
		return OIDCProfile{}, err
	}
	if !claims.EmailVerified {
		return OIDCProfile{}, ErrInvalidCredentials
	}
	return OIDCProfile{Subject: idToken.Subject, Email: claims.Email, DisplayName: claims.Name}, nil
}

func googleEndpoint(issuer string) oauth2.Endpoint {
	if issuer == "" || issuer == "https://accounts.google.com" {
		return oauth2.Endpoint{
			AuthURL:  "https://accounts.google.com/o/oauth2/v2/auth",
			TokenURL: "https://oauth2.googleapis.com/token",
		}
	}
	return oauth2.Endpoint{
		AuthURL:  issuer + "/auth",
		TokenURL: issuer + "/token",
	}
}
