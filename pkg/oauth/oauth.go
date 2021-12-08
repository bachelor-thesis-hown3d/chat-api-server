// Package oauth provides utility for oauth authentication
package oauth

import (
	"context"
	"fmt"
	"net/url"

	"github.com/coreos/go-oidc"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
)

type Claims struct {
	Email string `json:"email"`
	Name  string `json:"preferred_username"`
}

type claimKey struct{}

var (
	// EmailClaim is the type to use for extracting the email of the user from the context
	EmailClaim claimKey
	// NameClaim is the type to use for extracting the name of the user from the context
	NameClaim claimKey
	Verifer   *oidc.IDTokenVerifier
)

// OAuthMiddleware is used by a middleware to authenticate requests
func OAuthMiddleware(ctx context.Context) (context.Context, error) {
	rawToken, err := GetAuthTokenFromContext(ctx)
	if err != nil {
		return nil, err
	}

	token, err := Verifer.Verify(ctx, rawToken)
	if err != nil {
		return ctx, fmt.Errorf("Can't verify idToken: %v", err)
	}

	claims := &Claims{}
	if err := token.Claims(&claims); err != nil {
		return ctx, fmt.Errorf("Claims were missing from id token")
	}

	ctx = context.WithValue(ctx, EmailClaim, claims.Email)
	ctx = context.WithValue(ctx, NameClaim, claims.Name)
	return ctx, nil
}

func GetAuthTokenFromContext(ctx context.Context) (string, error) {
	return grpc_auth.AuthFromMD(ctx, "bearer")
}

func NewOAuth2Provider(ctx context.Context, issuerURL *url.URL) (*oidc.Provider, error) {
	return oidc.NewProvider(ctx, issuerURL.String())
}

func NewOAuth2Verifier(provider *oidc.Provider, clientID string) {
	Verifer = provider.Verifier(&oidc.Config{ClientID: clientID})
}
