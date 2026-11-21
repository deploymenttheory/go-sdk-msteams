package auth

import (
	"context"
	"errors"
	"sync"
	"time"
)

// RefreshFunc obtains a token for the exact request binding. It must return when
// ctx is cancelled and must not retain credentials in diagnostics.
type RefreshFunc func(context.Context, TokenRequest) (Token, error)

var (
	ErrInvalidRefresh   = errors.New("refresh callback and nonnegative refresh skew are required")
	ErrTokenExpiresSoon = errors.New("acquired token expires inside the refresh window")
)

// AcquisitionError keeps callback diagnostics out of its printable representation.
// Unwrap preserves explicit errors.Is/As inspection by callers who own the callback.
type AcquisitionError struct{ cause error }

func (AcquisitionError) Error() string      { return "token acquisition failed" }
func (e AcquisitionError) GoString() string { return e.Error() }
func (e AcquisitionError) Unwrap() error    { return e.cause }

// RefreshOptions controls caching. Zero RefreshSkew accepts any unexpired token;
// choose a positive skew to refresh before expiration. A nil Clock uses time.Now.
// An injected Clock must be safe for concurrent calls.
type RefreshOptions struct {
	RefreshSkew time.Duration
	Clock       func() time.Time
}

type flight struct {
	done      chan struct{}
	cancel    context.CancelFunc
	waiters   int
	abandoned bool
	finished  bool
	token     Token
	err       error
}

// RefreshProvider caches independently by tenant, cloud, identity, mode, audience and scope.
// One cancelled waiter does not cancel work still needed by another waiter. When all
// waiters leave, cancellation reaches the callback; a replacement waits for it to exit.
type RefreshProvider struct {
	mu       sync.Mutex
	binding  Binding
	refresh  RefreshFunc
	now      func() time.Time
	skew     time.Duration
	tokens   map[TokenRequest]Token
	flights  map[TokenRequest]*flight
	closed   chan struct{}
	isClosed bool
	workers  sync.WaitGroup
}

// NewRefreshProvider binds an explicit refresh callback; it does not call it.
func NewRefreshProvider(binding Binding, refresh RefreshFunc, options RefreshOptions) (*RefreshProvider, error) {
	if err := binding.Validate(); err != nil {
		return nil, err
	}
	if refresh == nil || options.RefreshSkew < 0 {
		return nil, ErrInvalidRefresh
	}
	now := options.Clock
	if now == nil {
		now = time.Now
	}
	return &RefreshProvider{binding: binding, refresh: refresh, now: now, skew: options.RefreshSkew,
		tokens: map[TokenRequest]Token{}, flights: map[TokenRequest]*flight{}, closed: make(chan struct{})}, nil
}

// Binding returns the immutable identity this provider serves.
func (p *RefreshProvider) Binding() Binding { return p.binding }

// String keeps cached tokens and callback state out of formatted diagnostics.
func (*RefreshProvider) String() string { return "msteams refresh provider (redacted)" }

// GoString also redacts Go-syntax formatted diagnostics.
func (p *RefreshProvider) GoString() string { return p.String() }

// AcquireToken returns a valid cached token or joins one refresh for the complete key.
func (p *RefreshProvider) AcquireToken(ctx context.Context, request TokenRequest) (Token, error) {
	if err := validateRequest(ctx, p.binding, request); err != nil {
		return Token{}, err
	}
	for {
		if err := ctx.Err(); err != nil {
			return Token{}, err
		}
		p.mu.Lock()
		if p.isClosed || p.closed == nil {
			p.mu.Unlock()
			return Token{}, ErrProviderClosed
		}
		if token, ok := p.tokens[request]; ok && token.ExpiresAt.After(p.now().Add(p.skew)) {
			p.mu.Unlock()
			return token, nil
		}
		delete(p.tokens, request)
		active := p.flights[request]
		if active != nil && active.abandoned {
			p.mu.Unlock()
			select {
			case <-ctx.Done():
				return Token{}, ctx.Err()
			case <-p.closed:
				return Token{}, ErrProviderClosed
			case <-active.done:
				continue
			}
		}
		if active == nil {
			// Shared work keeps the first waiter's values but has cancellation owned
			// by the waiter group, rather than by whichever caller arrived first.
			shared, cancel := context.WithCancel(context.WithoutCancel(ctx))
			active = &flight{done: make(chan struct{}), cancel: cancel}
			p.flights[request] = active
			p.workers.Add(1)
			go p.runRefresh(shared, request, active)
		}
		active.waiters++
		p.mu.Unlock()
		token, err := p.wait(ctx, active)
		p.release(active)
		return token, err
	}
}

func (p *RefreshProvider) wait(ctx context.Context, active *flight) (Token, error) {
	select {
	case <-ctx.Done():
		return Token{}, ctx.Err()
	case <-p.closed:
		return Token{}, ErrProviderClosed
	case <-active.done:
		if err := ctx.Err(); err != nil {
			return Token{}, err
		}
		p.mu.Lock()
		defer p.mu.Unlock()
		if p.isClosed {
			return Token{}, ErrProviderClosed
		}
		return active.token, active.err
	}
}

func (p *RefreshProvider) release(active *flight) {
	p.mu.Lock()
	defer p.mu.Unlock()
	active.waiters--
	if active.waiters == 0 && !active.finished {
		active.abandoned = true
		active.cancel()
	}
}

func (p *RefreshProvider) runRefresh(ctx context.Context, request TokenRequest, active *flight) {
	defer p.workers.Done()
	defer active.cancel()
	token, err := p.refresh(ctx, request)
	if ctx.Err() != nil {
		err = ctx.Err()
	} else if err != nil {
		err = AcquisitionError{cause: err}
	} else {
		if token.AccessToken == "" || token.ExpiresAt.IsZero() {
			err = ErrInvalidToken
		} else {
			now := p.now()
			if !token.ExpiresAt.After(now) {
				err = ErrTokenExpired
			} else if !token.ExpiresAt.After(now.Add(p.skew)) {
				err = ErrTokenExpiresSoon
			}
		}
	}
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isClosed {
		err = ErrProviderClosed
	} else if active.abandoned {
		err = context.Canceled
	}
	if err == nil {
		p.tokens[request] = token
	} else {
		token = Token{}
	}
	active.token, active.err, active.finished = token, err, true
	delete(p.flights, request)
	close(active.done)
}

// Close cancels outstanding refreshes and clears this provider's owned cache.
// Callback functions must honor cancellation; Close does not wait for foreign code.
func (p *RefreshProvider) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()
	if p.isClosed {
		return
	}
	p.isClosed = true
	clear(p.tokens)
	if p.closed != nil {
		close(p.closed)
	}
	for _, active := range p.flights {
		active.abandoned = true
		active.cancel()
	}
}

// Wait waits for callbacks to exit after Close. Close must be called first so no
// new refresh can start. It is intended for callers that own callback lifecycle.
func (p *RefreshProvider) Wait() { p.workers.Wait() }
