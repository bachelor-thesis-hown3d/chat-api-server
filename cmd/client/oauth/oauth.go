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

	"golang.org/x/oauth2"
)

var (
	keycloakConf *oauth2.Config
	oauthState   string
	Token        *string
)

func init() {
	keycloakConf = &oauth2.Config{
		RedirectURL:  "http://localhost:7070/kubernetes/callback",
		ClientID:     "kubernetes",
		ClientSecret: "25930b78-1656-47b1-93aa-a40c17754ac9",
		Endpoint: oauth2.Endpoint{
			AuthURL:   "https://localhost:8443/auth/realms/kubernetes/protocol/openid-connect/auth",
			TokenURL:  "https://localhost:8443/auth/realms/kubernetes/protocol/openid-connect/token",
			AuthStyle: oauth2.AuthStyleInHeader,
		},
	}
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
	url := keycloakConf.AuthCodeURL(oauthState)
	http.Redirect(w, r, url, http.StatusTemporaryRedirect)
}

func handleAuthCallback(w http.ResponseWriter, r *http.Request) {
	token, err := getToken(r.FormValue("state"), r.FormValue("code"))
	if err != nil {
		fmt.Printf("Error: %v", err)
	}
	// tokenData, err := json.Marshal(token)
	// if err != nil {
	// 	fmt.Printf("Error: %v", err)
	// }
	//fmt.Fprintf(w, string(tokenData))
	Token = &token.AccessToken
}

func getToken(state, code string) (*oauth2.Token, error) {
	if state != oauthState {
		return nil, fmt.Errorf("invalid oauth state")
	}

	// since we use a bad ssl certificate, embed a Insecure HTTP Client for oauth to use
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	ctx := context.Background()
	ctx = context.WithValue(ctx, oauth2.HTTPClient, http.DefaultClient)

	token, err := keycloakConf.Exchange(ctx, code)
	if err != nil {
		return nil, fmt.Errorf("code exchange failed: %s", err.Error())
	}
	return token, nil
}

var ()

func NewServer(u *url.URL) (*http.Server, net.Listener, error) {
	s := &http.Server{
		Addr: u.Host,
	}
	// set up a listener on the redirect port
	fmt.Println(u.Port())
	port := fmt.Sprintf(":%v", u.Port())
	l, err := net.Listen("tcp", port)
	if err != nil {
		return nil, nil, fmt.Errorf("can't listen to port %s: %s\n", port, err)
	}

	return s, l, nil
}
func StartWebServer(s *http.Server, l net.Listener) {
	http.DefaultTransport.(*http.Transport).TLSClientConfig = &tls.Config{InsecureSkipVerify: true}
	http.HandleFunc("/auth", handleAuthLogin)
	http.HandleFunc("/kubernetes/callback", handleAuthCallback)
	fmt.Printf("Starting oauth listener on %v\n", l.Addr())

	go s.Serve(l)
}

func StopWebServer(s *http.Server) {
	s.Close()
}
