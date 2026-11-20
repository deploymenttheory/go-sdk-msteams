// Package client owns per-client session state shared by backend services.
package client

import (
	"context"
	"errors"
	"net/http"
	"sync"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

// ErrClosed means the local SDK session has been closed, not that Entra revoked a token.
var ErrClosed = errors.New("msteams client is closed")

// ErrNilContext identifies missing operation cancellation context.
var ErrNilContext = errors.New("operation context is required")

// ErrProviderUnavailable means this configuration's acquisition adapter is not installed.
var ErrProviderUnavailable = errors.New("authentication acquisition adapter is unavailable")

// Session owns copied credentials and cancellation for one client. The lifetime
// context represents the session itself; operation contexts are never stored here.
type Session struct {
	mu            sync.RWMutex
	auth          *auth.Config
	http          http.Client
	closed        bool
	lifetime      context.Context
	cancel        context.CancelFunc
	cleanup       func()
	provider      auth.Provider
	binding       auth.Binding
	closeProvider func()
}

// NewSession copies configuration. cleanup is called once when this owned session
// closes; callers must not transfer cleanup ownership of borrowed resources.
func NewSession(cfg *auth.Config, httpClient *http.Client, cleanup func()) (*Session, error) {
	if err := cfg.Validate(); err != nil {
		return nil, err
	}
	if httpClient == nil {
		return nil, errors.New("http client is required")
	}
	var provider auth.Provider
	var closeProvider func()
	binding := cfg.TokenBinding()
	switch cfg.Method {
	case auth.Browser, auth.DeviceCode, auth.Password:
		delegated, err := auth.NewDelegatedProvider(cfg, httpClient)
		if err != nil {
			return nil, err
		}
		provider, closeProvider = delegated, delegated.Close
	case auth.Certificate, auth.ClientSecret:
		application, err := auth.NewApplicationProvider(cfg, httpClient)
		if err != nil {
			return nil, err
		}
		provider, closeProvider = application, application.Close
	case auth.SuppliedTokens:
		supplied, err := auth.NewSuppliedTokenProvider(binding, cfg.AccessTokens)
		if err != nil {
			return nil, err
		}
		provider, closeProvider = supplied, supplied.Close
	case auth.ExternalProvider:
		borrowed, err := auth.NewBoundProvider(binding, cfg.Provider)
		if err != nil {
			return nil, err
		}
		provider = borrowed
	}
	lifetime, cancel := context.WithCancel(context.Background())
	return &Session{auth: cfg.Clone(), http: *httpClient, lifetime: lifetime, cancel: cancel, cleanup: cleanup,
		provider: provider, binding: binding, closeProvider: closeProvider}, nil
}

// AcquireToken obtains the exact audience requested by a backend adapter. The
// session supplies immutable identity fields and cancellation; it never substitutes
// a token from another audience. Endpoint-to-audience routing belongs to transport.
func (s *Session) AcquireToken(ctx context.Context, audience string) (auth.Token, error) {
	operation, cancel, err := s.Begin(ctx)
	if err != nil {
		return auth.Token{}, err
	}
	defer cancel()
	s.mu.RLock()
	provider, binding, closed := s.provider, s.binding, s.closed
	s.mu.RUnlock()
	if closed {
		return auth.Token{}, ErrClosed
	}
	if provider == nil {
		return auth.Token{}, ErrProviderUnavailable
	}
	token, err := provider.AcquireToken(operation, binding.Request(audience))
	s.mu.RLock()
	closed = s.closed
	s.mu.RUnlock()
	if closed {
		return auth.Token{}, ErrClosed
	}
	if operation.Err() != nil {
		return auth.Token{}, operation.Err()
	}
	return token, err
}

// AuthSnapshot returns independently mutable credential material to an adapter.
// It must never be logged or retained after the adapter's operation needs it.
func (s *Session) AuthSnapshot() (*auth.Config, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed || s.auth == nil {
		return nil, ErrClosed
	}
	return s.auth.Clone(), nil
}

// HTTPClient returns a client value; the configured round tripper remains shared
// within the session (or caller-owned when explicitly injected).
func (s *Session) HTTPClient() (http.Client, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed || s.auth == nil {
		return http.Client{}, ErrClosed
	}
	return s.http, nil
}

// Begin derives an operation from the caller's context and local session lifetime.
// The returned cancel function must be called when the operation finishes.
func (s *Session) Begin(ctx context.Context) (context.Context, context.CancelFunc, error) {
	if ctx == nil {
		return nil, nil, ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	s.mu.RLock()
	defer s.mu.RUnlock()
	if s.closed || s.auth == nil {
		return nil, nil, ErrClosed
	}
	operation, cancel := context.WithCancel(ctx)
	stop := context.AfterFunc(s.lifetime, cancel)
	return operation, func() { stop(); cancel() }, nil
}

// Close cancels local work, clears owned credential references and closes owned
// idle connections. It neither revokes Entra tokens nor clears caller-owned caches.
func (s *Session) Close() {
	s.mu.Lock()
	if s.closed || s.auth == nil {
		s.mu.Unlock()
		return
	}
	s.closed = true
	if s.cancel != nil {
		s.cancel()
	}
	clear(s.auth.Certificate)
	clear(s.auth.PrivateKey)
	clear(s.auth.PFX)
	clear(s.auth.PFXPassword)
	clear(s.auth.AccessTokens)
	s.auth = nil
	s.provider = nil
	closeProvider := s.closeProvider
	s.closeProvider = nil
	cleanup := s.cleanup
	s.cleanup = nil
	s.mu.Unlock()
	if closeProvider != nil {
		closeProvider()
	}
	if cleanup != nil {
		cleanup()
	}
}

// Backend selects a service family independently of its eventual API route.
type Backend string

const (
	Graph      Backend = "graph"
	TeamsAdmin Backend = "teams-admin"
)

// Scope binds a backend service collection to its owning session.
type Scope struct {
	backend Backend
	session *Session
}

// NewScope binds service dependencies; no authentication or discovery occurs.
func NewScope(backend Backend, session *Session) (*Scope, error) {
	if session == nil || (backend != Graph && backend != TeamsAdmin) {
		return nil, errors.New("valid backend and session are required")
	}
	return &Scope{backend: backend, session: session}, nil
}

// Backend returns the immutable service family.
func (s *Scope) Backend() Backend { return s.backend }

// Session supplies the shared session to resource adapters in this service family.
func (s *Scope) Session() *Session { return s.session }
