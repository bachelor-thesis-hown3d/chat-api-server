// Package oauth provides utility for oauth authentication
package oauth

import (
	"context"

	grpc_auth "github.com/grpc-ecosystem/go-grpc-middleware/auth"
)

type key string

const (
	// TokenKey is the type to use for extracting the token of the user from the context
	TokenKey = key("token")
)

// Middleware is used to authenticate requests
func Middleware(ctx context.Context) (context.Context, error) {
	rawToken, err := GetAuthTokenFromContext(ctx)
	if err != nil {
		return nil, err
	}

	ctx = context.WithValue(ctx, TokenKey, rawToken)
	return ctx, nil
}

func GetAuthTokenFromContext(ctx context.Context) (string, error) {
	return grpc_auth.AuthFromMD(ctx, "bearer")
}
