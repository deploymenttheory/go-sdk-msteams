package client

import (
	"context"
	"encoding/json"
	"errors"
	"mime"
	"net/http"
	"net/url"
	"strings"
	"sync"

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
}

// MaxResponseBytes bounds buffered API responses, including decompressed JSON.
const MaxResponseBytes = 8 << 20

var (
	ErrTransport        = errors.New("backend request failed")
	ErrResponseTooLarge = errors.New("backend response exceeds the buffering limit")
	ErrDecode           = errors.New("backend response is not valid for the result model")
)

var _ Client = (*Scope)(nil)

// Do sends one request without automatic retries or redirect following. It closes
// the response body and returns buffered Resty metadata even for HTTP failures.
// result, when non-nil, receives JSON from nonempty successful responses. Raw
// responses may contain tenant data; callers must not log them indiscriminately.
// Errors deliberately omit response messages, URLs, bodies and credentials.
func (s *Scope) Do(ctx context.Context, request Request, result any) (*resty.Response, error) {
	if s == nil || s.session == nil {
		return nil, ErrClosed
	}
	operation, cancel, err := s.session.Begin(ctx)
	if err != nil {
		return nil, err
	}
	defer cancel()
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
	headers, err := requestHeaders(request.Header)
	if err != nil {
		return nil, err
	}
	var body []byte
	if request.Body != nil {
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
	httpClient, err := s.session.HTTPClient()
	if err != nil {
		return nil, err
	}
	token, err := s.session.AcquireToken(operation, profile.audience)
	if err != nil {
		return nil, err
	}
	base := httpClient.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	transport := &authenticatedTransport{base: base, profile: profile, token: token.AccessToken}
	defer transport.clearToken()
	httpClient.Transport = transport
	httpClient.Jar = nil
	httpClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	// Per-operation Resty state prevents headers, payloads and result pointers from
	// leaking between concurrent calls. The underlying connection pool is shared.
	restyClient := resty.NewWithClient(&httpClient).SetLogger(quietLogger{}).
		SetRetryCount(0).SetResponseBodyLimit(MaxResponseBytes).SetAllowGetMethodPayload(true)
	r := restyClient.R().SetContext(operation)
	r.Header = headers
	if request.Body != nil {
		r.SetBody(body)
	}
	response, err := r.Execute(request.Method, destination.String())
	if operation.Err() != nil {
		return response, operation.Err()
	}
	if err != nil {
		if errors.Is(err, resty.ErrResponseBodyTooLarge) {
			return response, ErrResponseTooLarge
		}
		// net/http errors embed the URL and may embed an arbitrary transport's
		// diagnostic text. Do not retain those errors as printable or unwrap causes.
		if errors.Is(err, context.DeadlineExceeded) {
			return response, context.DeadlineExceeded
		}
		if errors.Is(err, context.Canceled) {
			return response, context.Canceled
		}
		return response, ErrTransport
	}
	if !response.IsSuccess() {
		return response, normalizeError(s.backend, response)
	}
	if result != nil && len(response.Body()) != 0 {
		if err := json.Unmarshal(response.Body(), result); err != nil {
			return response, ErrDecode
		}
		r.Result = result
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
