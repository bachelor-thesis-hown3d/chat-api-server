// Package oauth provides utility for oauth authentication
package oauth

import (
	"context"
	"crypto/tls"
	"fmt"
	"net/http"
	"net/url"

	"github.com/coreos/go-oidc"
	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
	"golang.org/x/oauth2"
)

type Claims struct {
	Email string `json:"email"`
	Name  string `json:"preferred_username"`
}

type claimKey string

const (
	// EmailClaimKey is the type to use for extracting the email of the user from the context
	EmailClaimKey = claimKey("email")
	// NameClaimKey is the type to use for extracting the name of the user from the context
	NameClaimKey = claimKey("name")
)

var (
	Verifer *oidc.IDTokenVerifier
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

	ctx = context.WithValue(ctx, EmailClaimKey, claims.Email)
	ctx = context.WithValue(ctx, NameClaimKey, claims.Name)
	return ctx, nil
}

func GetAuthTokenFromContext(ctx context.Context) (string, error) {
	return grpc_auth.AuthFromMD(ctx, "bearer")
}

func NewOAuth2Provider(ctx context.Context, issuerURL *url.URL) (*oidc.Provider, error) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	ctx = context.WithValue(ctx, oauth2.HTTPClient, http.DefaultClient)
	return oidc.NewProvider(ctx, issuerURL.String())
}

func NewOAuth2Verifier(provider *oidc.Provider, clientID string) {
	Verifer = provider.Verifier(&oidc.Config{ClientID: clientID})
}
