package oauth

import (
	"context"
	"crypto/tls"
	"encoding/hex"
	"fmt"
	"math/rand"
	"net"
	"net/http"
	"net/url"

	"github.com/bachelor-thesis-hown3d/chat-api-server/pkg/oauth"
	"github.com/coreos/go-oidc"
	"golang.org/x/oauth2"
)

var (
	conf       oauth2.Config
	oauthState string
	//Channel to recieve IDTokens from
	IDToken *string
)

func newOAuth2Config(ctx context.Context, issuerURL, redirectURL *url.URL, clientID, clientSecret string) error {
	provider, err := oauth.NewOAuth2Provider(ctx, issuerURL)
	if err != nil {
		return err
	}

	conf = oauth2.Config{
		RedirectURL:  redirectURL.String(),
		ClientID:     clientID,
		ClientSecret: clientSecret,
		Scopes:       []string{oidc.ScopeOpenID},
		Endpoint:     provider.Endpoint(),
	}

	oauth.NewOAuth2Verifier(provider, clientID)
	return nil
}

func randomHex(n int) (string, error) {
	rand.Seed(42)
	bytes := make([]byte, n)
	if _, err := rand.Read(bytes); err != nil {
		return "", err
	}
	return hex.EncodeToString(bytes), nil
}

func handleAuthLogin(w http.ResponseWriter, r *http.Request) {
	oauthState, _ = randomHex(16)
	url := conf.AuthCodeURL(oauthState)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	token, err := getToken(oauth.Verifer, r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		fmt.Printf("Error: %v", err)
		return
	}
	IDToken = &token
}

// getToken returns
func getToken(verifer *oidc.IDTokenVerifier, state, code string) (string, error) {
	if state != oauthState {
		return "", fmt.Errorf("invalid oauth state")
	}

	// since we use a bad ssl certificate, embed a Insecure HTTP Client for oauth to use
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	ctx := context.Background()
	ctx = context.WithValue(ctx, oauth2.HTTPClient, http.DefaultClient)

	token, err := conf.Exchange(ctx, code)
	if err != nil {
		return "", fmt.Errorf("code exchange failed: %s", err.Error())
	}

	rawIDToken, ok := token.Extra("id_token").(string)
	if !ok {
		return "", fmt.Errorf("id_token was missing from oauth2 response")
	}
	_, err = verifer.Verify(context.Background(), rawIDToken)
	if err != nil {
		return "", fmt.Errorf("Can't verify idToken: %v", err)
	}
	return rawIDToken, nil
	// claims := &Claims{}
	// if err := idToken.Claims(&claims); err != nil {
	// 	return TokenWithClaims{}, fmt.Errorf("Claims were missing from id token")
	// }
	// return TokenWithClaims{IDToken: idToken, Claims: claims}, nil
}

func NewServer(ctx context.Context, clientID, clientSecret string, redirectUrl, issuerUrl *url.URL) (*http.Server, net.Listener, error) {

	s := &http.Server{
		Addr: redirectUrl.Host,
	}
	// set up a listener on the redirect port
	port := fmt.Sprintf(":%v", redirectUrl.Port())
	l, err := net.Listen("tcp", port)
	if err != nil {
		return nil, nil, fmt.Errorf("can't listen to port %s: %s\n", port, err)
	}

	callbackPath := "/callback"
	redirectUrl.Path = callbackPath
	err = newOAuth2Config(ctx, issuerUrl, redirectUrl, clientID, clientSecret)
	if err != nil {
		return nil, nil, fmt.Errorf("Can't create oauth2 config: %v", err)
	}
	http.HandleFunc("/auth", handleAuthLogin)
	http.HandleFunc(callbackPath, handleAuthCallback)
	return s, l, nil
}
func StartWebServer(s *http.Server, l net.Listener) error {
	fmt.Printf("Starting oauth listener on %v\n", l.Addr())
	err := s.Serve(l)
	return err
}

func StopWebServer(s *http.Server) {
	s.Close()
}
