package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"sync"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"

	"github.com/go-resty/resty/v2"
)

// Client is the request boundary injected into typed resource services.
// Implementations bind a backend and choose its token audience independently of requests.
type Client interface {
	Do(context.Context, Request, any) (*resty.Response, error)
}

// Request describes one JSON API operation. Path contains unescaped segments,
// including Graph's explicit v1.0 or beta version. Query is encoded separately.
// URL accepts an absolute discovery/continuation destination instead of Path and
// Query; it must pass the same backend allowlist. Body is marshaled as JSON.
// Streaming and binary payloads are separate contracts, not implicit JSON bodies.
type Request struct {
	Method string
	Path   []string
	Query  url.Values
	URL    string
	Header http.Header
	Body   any
	Replay ReplaySafety
	// MinimumRetryDelay is an operation-specific floor between retry attempts.
	// It can only increase shared backoff/Retry-After delays. Zero adds no floor.
	MinimumRetryDelay time.Duration
	// RefreshOnUnauthorized permits one noninteractive token replacement and replay
	// on 401. It also requires ReplaySafe and a capable provider. Claims challenges
	// and interactive consent are left to the caller.
	RefreshOnUnauthorized bool
}

// MaxResponseBytes bounds buffered API responses, including decompressed JSON.
const MaxResponseBytes = 8 << 20

var (
	ErrTransport        = errors.New("backend request failed")
	ErrResponseTooLarge = errors.New("backend response exceeds the buffering limit")
	ErrDecode           = errors.New("backend response is not valid for the result model")
)

var _ Client = (*Scope)(nil)

// Do applies the shared request policy to explicitly replay-safe operations. It closes
// the response body and returns buffered Resty metadata even for HTTP failures.
// result, when non-nil, receives JSON from nonempty successful responses. Raw
// responses may contain tenant data; callers must not log them indiscriminately.
// Errors deliberately omit response messages, URLs, bodies and credentials.
func (s *Scope) Do(ctx context.Context, request Request, result any) (response *resty.Response, err error) {
	ctx, observed := s.observe(ctx, request.Method)
	defer func() { observed.finish(response, err) }()
	if s == nil || s.session == nil {
		return nil, ErrClosed
	}
	operation, cancel, err := s.session.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
	controls := s.session.requests
	operation, stop := context.WithTimeout(operation, controls.policy.Timeout)
	defer stop()
	expires := controls.now().Add(controls.policy.Timeout)
	if (request.Replay != ReplayNever && request.Replay != ReplaySafe) || request.MinimumRetryDelay < 0 {
		return nil, ErrRequest
	}
	profile, err := s.profile()
	if err != nil {
		return nil, err
	}
	destination, err := profile.destination(request)
	if err != nil {
		return nil, err
	}
	switch request.Method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions:
	default:
		return nil, ErrRequest
	}
	headers, err := s.session.requestHeaders(request.Header)
	if err != nil {
		return nil, err
	}
	var body []byte
	if request.Body != nil {
		if _, stream := request.Body.(io.Reader); stream {
			return nil, ErrRequest
		}
		if request.Method == http.MethodHead || request.Method == http.MethodOptions {
			return nil, ErrRequest
		}
		body, err = json.Marshal(request.Body)
		if err != nil {
			return nil, ErrRequest
		}
		contentType := headers.Get("Content-Type")
		if contentType == "" {
			headers.Set("Content-Type", "application/json")
		} else {
			mediaType, _, parseErr := mime.ParseMediaType(contentType)
			if parseErr != nil || mediaType != "application/json" &&
				(!strings.HasPrefix(mediaType, "application/") || !strings.HasSuffix(mediaType, "+json")) {
				return nil, ErrRequest
			}
		}
	}

	token, err := s.session.AcquireTokenScope(operation, profile.audience, profile.scope)
	if err != nil {
		return nil, err
	}
	refreshed := false
	for attempt := 1; ; attempt++ {
		if err = operation.Err(); err != nil {
			return response, err
		}
		if !controls.now().Before(expires) {
			return response, ErrRetryBudget
		}
		for {
			if err = controls.acquire(operation); err != nil {
				return response, err
			}
			if token.ExpiresAt.After(time.Now()) {
				break
			}
			controls.release()
			token, err = s.session.AcquireTokenScope(operation, profile.audience, profile.scope)
			if err != nil {
				return response, err
			}
		}
		response, err = s.attempt(operation, profile, request.Method, destination.String(), headers, body, request.Body != nil, token.AccessToken)
		controls.release()
		observed.attempt(response, err)
		if operation.Err() != nil {
			return response, operation.Err()
		}
		if err == nil {
			if result != nil && len(response.Body()) != 0 {
				if json.Unmarshal(response.Body(), result) != nil {
					return response, ErrDecode
				}
				response.Request.Result = result
			}
			return response, nil
		}
		if request.Replay != ReplaySafe {
			return response, err
		}
		unauthorized := response != nil && response.StatusCode() == http.StatusUnauthorized && request.RefreshOnUnauthorized && !refreshed
		transient := errors.Is(err, ErrTransport) || response != nil && retryableStatus(response.StatusCode()) && !errors.Is(err, ErrResponseTooLarge)
		if !unauthorized && !transient {
			return response, err
		}
		if attempt >= controls.policy.MaxAttempts {
			return response, errors.Join(ErrRetryBudget, err)
		}
		if unauthorized {
			refreshed = true
			replacement, refreshErr := s.session.RefreshAfterRejectionScope(operation, profile.audience, profile.scope, token)
			if refreshErr != nil {
				if errors.Is(refreshErr, auth.ErrRefreshUnsupported) {
					return response, err
				}
				return response, refreshErr
			}
			if replacement.AccessToken == token.AccessToken {
				return response, err
			}
			token = replacement
		}
		retryHeader := ""
		if response != nil {
			retryHeader = response.Header().Get("Retry-After")
		}
		delay := controls.delay(retryHeader, attempt)
		if unauthorized {
			// Token replacement has no ordinary transient backoff, but still
			// respects this operation's floor and an explicit server delay.
			delay = 0
			if server, valid := retryAfter(retryHeader, controls.now()); valid {
				delay = server
			}
		}
		delay = max(delay, request.MinimumRetryDelay)
		remaining := expires.Sub(controls.now())
		if deadline, ok := operation.Deadline(); ok {
			remaining = min(remaining, time.Until(deadline))
		}
		if delay >= remaining {
			return response, errors.Join(ErrRetryBudget, err)
		}
		if waitErr := controls.sleep(operation, delay); waitErr != nil {
			return response, waitErr
		}
	}
}

// attempt owns a fresh Resty request and closes/buffers its response before the
// concurrency permit is released. The JSON body was serialized once before retry.
func (s *Scope) attempt(operation context.Context, profile routeProfile, method, destination string, headers http.Header, body []byte, hasBody bool, token string) (*resty.Response, error) {
	httpClient, err := s.session.HTTPClient()
	if err != nil {
		return nil, err
	}
	base := httpClient.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	transport := &authenticatedTransport{base: base, profile: profile, token: token}
	defer transport.clearToken()
	httpClient.Transport = transport
	httpClient.Jar = nil
	httpClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	restyClient := resty.NewWithClient(&httpClient).SetLogger(quietLogger{}).
		SetRetryCount(0).SetResponseBodyLimit(MaxResponseBytes).SetAllowGetMethodPayload(true)
	r := restyClient.R().SetContext(operation)
	r.Header = headers.Clone()
	if hasBody {
		r.SetBody(body)
	}
	response, err := r.Execute(method, destination)
	if operation.Err() != nil {
		return response, operation.Err()
	}
	if err != nil {
		if errors.Is(err, resty.ErrResponseBodyTooLarge) {
			return response, ErrResponseTooLarge
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return response, errors.Join(ErrTransport, context.DeadlineExceeded)
		}
		if errors.Is(err, context.Canceled) {
			return response, context.Canceled
		}
		return response, ErrTransport
	}
	if !response.IsSuccess() {
		return response, normalizeError(s.backend, response)
	}
	return response, nil
}

// Credentials are attached only to the transport's private request clone. Resty's
// request metadata and the returned RawResponse.Request never retain the token.
type authenticatedTransport struct {
	mu      sync.Mutex
	base    http.RoundTripper
	profile routeProfile
	token   string
}

func (t *authenticatedTransport) clearToken() {
	t.mu.Lock()
	t.token = ""
	t.mu.Unlock()
}

func (t *authenticatedTransport) RoundTrip(request *http.Request) (*http.Response, error) {
	if !t.profile.allows(request.URL) || request.Host != "" && request.Host != t.profile.host {
		closeRequestBody(request)
		return nil, ErrDestination
	}
	if err := request.Context().Err(); err != nil {
		closeRequestBody(request)
		return nil, err
	}
	private := request.Clone(request.Context())
	// SDK retries create fresh buffered requests. Do not let the standard HTTP
	// transport independently rewind a bodied operation after an ambiguous failure.
	private.GetBody = nil
	t.mu.Lock()
	token := t.token
	t.mu.Unlock()
	if token == "" {
		closeRequestBody(request)
		return nil, ErrClosed
	}
	private.Header.Set("Authorization", "Bearer "+token)
	response, err := t.base.RoundTrip(private)
	if response != nil {
		response.Request = request
		if err != nil && response.Body != nil {
			// net/http discards a response returned alongside a transport error.
			_ = response.Body.Close()
		}
	}
	return response, err
}

func closeRequestBody(request *http.Request) {
	if request.Body != nil {
		_ = request.Body.Close()
	}
}

// Resty diagnostics can contain server bodies. Observability is provided at the
// SDK boundary, never by enabling its unredacted debug logger.
type quietLogger struct{}

func (quietLogger) Errorf(string, ...any) {}
func (quietLogger) Warnf(string, ...any)  {}
func (quietLogger) Debugf(string, ...any) {}
