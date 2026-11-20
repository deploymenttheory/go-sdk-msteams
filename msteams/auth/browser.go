package auth

import (
	"context"
	"crypto/rand"
	"crypto/sha256"
	"encoding/base64"
	"net"
	"net/http"
	"net/url"
	"strings"
	"time"

	"github.com/AzureAD/microsoft-authentication-library-for-go/apps/public"
)

type browserResult struct {
	code string
	err  error
}

// MSAL builds the authorize URL and redeems the code. The SDK owns the loopback
// receiver so cancellation forcibly closes active connections, including a slow
// POST body. MSAL v1.10.1's interactive receiver uses an unbounded Shutdown wait.
func browserToken(ctx context.Context, msal public.Client, authority *delegatedHTTP, cfg *Config, scopes []string) (public.AuthResult, error) {
	var secret [32]byte
	if _, err := rand.Read(secret[:]); err != nil {
		return public.AuthResult{}, ErrDelegatedAuthentication
	}
	verifier := base64.RawURLEncoding.EncodeToString(secret[:])
	challenge := sha256.Sum256([]byte(verifier))
	state := rand.Text()
	var listen net.ListenConfig
	listener, err := listen.Listen(ctx, "tcp", "localhost:0")
	if err != nil {
		return public.AuthResult{}, ErrDelegatedAuthentication
	}
	_, port, err := net.SplitHostPort(listener.Addr().String())
	if err != nil {
		_ = listener.Close()
		return public.AuthResult{}, ErrDelegatedAuthentication
	}
	redirect := "http://localhost:" + port
	results := make(chan browserResult, 1)
	server := &http.Server{ReadHeaderTimeout: 5 * time.Second, ReadTimeout: 10 * time.Second, WriteTimeout: 5 * time.Second,
		Handler: browserCallback("localhost:"+port, state, results)}
	done := make(chan struct{})
	go func() { defer close(done); _ = server.Serve(listener) }()
	defer func() { _ = server.Close(); <-done }()
	authorize, err := msal.AuthCodeURL(ctx, cfg.ClientID, redirect, append(append([]string{}, scopes...), "openid", "offline_access", "profile"), public.WithLoginHint(cfg.LoginHint))
	if err != nil {
		return public.AuthResult{}, err
	}
	u, err := url.Parse(authorize)
	if err != nil || !authority.validAuthorizationURL(u) {
		return public.AuthResult{}, ErrAuthorityDestination
	}
	query := u.Query()
	query.Set("state", state)
	query.Set("code_challenge", base64.RawURLEncoding.EncodeToString(challenge[:]))
	query.Set("code_challenge_method", "S256")
	query.Set("response_mode", "form_post")
	u.RawQuery = query.Encode()
	if err := cfg.OpenBrowser(ctx, u.String()); err != nil {
		return public.AuthResult{}, loginCallbackError(ctx, err)
	}
	select {
	case <-ctx.Done():
		return public.AuthResult{}, ctx.Err()
	case result := <-results:
		if result.err != nil {
			return public.AuthResult{}, result.err
		}
		return msal.AcquireTokenByAuthCode(ctx, result.code, redirect, scopes, public.WithChallenge(verifier))
	}
}

func browserCallback(host, state string, results chan<- browserResult) http.Handler {
	return http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Cache-Control", "no-store")
		w.Header().Set("Content-Type", "text/plain; charset=utf-8")
		if r.Host != host || r.URL.Path != "/" || r.Method != http.MethodPost {
			http.Error(w, "Invalid sign-in callback", http.StatusBadRequest)
			return
		}
		r.Body = http.MaxBytesReader(w, r.Body, 16<<10)
		defer func() { _ = r.Body.Close() }()
		if err := r.ParseForm(); err != nil {
			http.Error(w, "Invalid sign-in callback", http.StatusBadRequest)
			return
		}
		// Validate state before accepting either a code or an OAuth error. Do not
		// reflect descriptions or incoming values into the local browser response.
		if values := r.PostForm["state"]; len(values) != 1 || values[0] != state {
			http.Error(w, "Invalid sign-in state", http.StatusBadRequest)
			return
		}
		var result browserResult
		if values := r.PostForm["error"]; len(values) > 0 && values[0] != "" {
			result.err = ErrAuthenticationDenied
		} else {
			codes := r.PostForm["code"]
			if len(codes) != 1 || strings.TrimSpace(codes[0]) == "" {
				http.Error(w, "Missing authorization code", http.StatusBadRequest)
				return
			}
			result.code = codes[0]
		}
		select {
		case results <- result:
		default:
		}
		_, _ = w.Write([]byte("Sign-in response received. You may close this window."))
	})
}
