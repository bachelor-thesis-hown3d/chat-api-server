// Package oauth provides utility for oauth authentication
package oauth

import (
	"context"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
)

// OAuthMiddleware is used by a middleware to authenticate requests
func OAuthMiddleware(ctx context.Context) (context.Context, error) {
	_, err := GetAuthTokenFromContext(ctx)
	if err != nil {
		return nil, err
	}

	return ctx, nil
}

func GetAuthTokenFromContext(ctx context.Context) (string, error) {
	return grpc_auth.AuthFromMD(ctx, "bearer")
}
