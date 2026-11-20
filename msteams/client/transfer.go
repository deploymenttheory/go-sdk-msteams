package client

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"mime"
	"net/http"
	"reflect"
	"sync"
	"time"

	"github.com/go-resty/resty/v2"
)

var (
	ErrUploadTooLarge = errors.New("upload exceeds the configured byte limit")
	ErrUploadRead     = errors.New("upload source failed")
	ErrStreamClose    = errors.New("transfer stream could not be closed")
	ErrDownloadWrite  = errors.New("download destination failed")
)

// TransferClient adds explicit stream operations to the JSON request boundary.
// Resource services select content types, size limits and evidenced routes.
type TransferClient interface {
	Upload(context.Context, Request, io.ReadCloser, string, int64, any) (*resty.Response, error)
	Download(context.Context, Request, io.Writer, int64) (*resty.Response, int64, error)
}

var _ TransferClient = (*Scope)(nil)

// Upload takes ownership of source immediately, including on validation failure.
// It closes source exactly once. Close must unblock concurrent Read; cancellation
// invokes it to release blocked sources. Bytes are sent without JSON/base64 encoding
// or full buffering. maxBytes must be positive. Oversize/failed uploads can leave
// partial remote state; the caller decides recovery. Neither redirects nor retries
// are followed, even for seekable sources. Request.Body must be nil, Replay must be
// ReplayNever, and RefreshOnUnauthorized must be false. result decodes buffered JSON.
func (s *Scope) Upload(ctx context.Context, request Request, source io.ReadCloser, contentType string, maxBytes int64, result any) (response *resty.Response, err error) {
	if nilStream(source) {
		return nil, ErrRequest
	}
	owned := &ownedStream{source: source}
	defer func() {
		if owned.Close() != nil {
			err = errors.Join(err, ErrStreamClose)
		}
	}()
	if maxBytes <= 0 || request.Body != nil || request.Replay != ReplayNever || request.RefreshOnUnauthorized || (request.Method != http.MethodPost && request.Method != http.MethodPut && request.Method != http.MethodPatch) {
		return nil, ErrRequest
	}
	if _, _, parseErr := mime.ParseMediaType(contentType); parseErr != nil || contentType == "" {
		return nil, ErrRequest
	}
	headers, err := requestHeaders(request.Header)
	if err != nil {
		return nil, err
	}
	if existing := headers.Get("Content-Type"); existing != "" && existing != contentType {
		return nil, ErrRequest
	}
	headers.Set("Content-Type", contentType)
	request.Header = headers
	bounded := &uploadReader{source: owned, remaining: maxBytes}
	response, _, err = s.transfer(ctx, request, bounded, owned, nil, 0)
	_ = owned.Close() // Wait for in-flight source reads before inspecting their result.
	if sourceErr := bounded.failure(); sourceErr != nil {
		err = errors.Join(err, sourceErr)
	}
	if err == nil && result != nil && len(response.Body()) != 0 {
		if json.Unmarshal(response.Body(), result) != nil {
			return response, ErrDecode
		}
		response.Request.Result = result
	}
	return response, err
}

// Download copies at most maxBytes to a borrowed destination and closes the HTTP
// body before returning. It never closes destination. A positive limit is required.
// The returned count includes bytes successfully written even on failure. Partial
// bytes are not rolled back and the request is never automatically retried. Non-2xx
// bodies are buffered for APIError classification, never written to destination.
// Request must be a GET without Body, ReplaySafe or RefreshOnUnauthorized.
// Destination.Write must return promptly; arbitrary blocked writers cannot be canceled.
func (s *Scope) Download(ctx context.Context, request Request, destination io.Writer, maxBytes int64) (*resty.Response, int64, error) {
	if nilStream(destination) || maxBytes <= 0 || request.Method != http.MethodGet || request.Body != nil || request.Replay != ReplayNever || request.RefreshOnUnauthorized {
		return nil, 0, ErrRequest
	}
	return s.transfer(ctx, request, nil, nil, destination, maxBytes)
}

// transfer uses net/http for stream bodies: the pinned Resty middleware otherwise
// materializes io.Reader into GetBody. Responses retain the SDK's Resty metadata
// shape. The same destination guard, deadline and concurrency budget still apply.
func (s *Scope) transfer(ctx context.Context, request Request, body io.Reader, source io.Closer, sink io.Writer, maxBytes int64) (*resty.Response, int64, error) {
	if s == nil || s.session == nil {
		return nil, 0, ErrClosed
	}
	operation, cancel, err := s.session.Begin(ctx)
	if err != nil {
		return nil, 0, err
	}
	defer cancel()
	controls := s.session.requests
	operation, stop := context.WithTimeout(operation, controls.policy.Timeout)
	defer stop()
	if source != nil {
		stopClose := context.AfterFunc(operation, func() { _ = source.Close() })
		defer stopClose()
	}
	if request.MinimumRetryDelay < 0 {
		return nil, 0, ErrRequest
	}
	profile, err := s.profile()
	if err != nil {
		return nil, 0, err
	}
	destination, err := profile.destination(request)
	if err != nil {
		return nil, 0, err
	}
	headers, err := requestHeaders(request.Header)
	if err != nil {
		return nil, 0, err
	}
	token, err := s.session.AcquireToken(operation, profile.audience)
	if err != nil {
		return nil, 0, err
	}
	for {
		if err := controls.acquire(operation); err != nil {
			return nil, 0, err
		}
		if token.ExpiresAt.After(time.Now()) {
			break
		}
		controls.release()
		token, err = s.session.AcquireToken(operation, profile.audience)
		if err != nil {
			return nil, 0, err
		}
	}
	defer controls.release()
	httpClient, err := s.session.HTTPClient()
	if err != nil {
		return nil, 0, err
	}
	base := httpClient.Transport
	if base == nil {
		base = http.DefaultTransport
	}
	guard := &authenticatedTransport{base: base, profile: profile, token: token.AccessToken}
	defer guard.clearToken()
	httpClient.Transport = guard
	httpClient.Jar = nil
	httpClient.CheckRedirect = func(*http.Request, []*http.Request) error { return http.ErrUseLastResponse }
	rawRequest, err := http.NewRequestWithContext(operation, request.Method, destination.String(), body)
	if err != nil {
		return nil, 0, ErrRequest
	}
	rawRequest.Header = headers
	rawRequest.GetBody = nil
	raw, err := httpClient.Do(rawRequest)
	if err != nil {
		if operation.Err() != nil {
			return nil, 0, operation.Err()
		}
		if errors.Is(err, context.DeadlineExceeded) {
			return nil, 0, errors.Join(ErrTransport, context.DeadlineExceeded)
		}
		return nil, 0, ErrTransport
	}
	response := &resty.Response{RawResponse: raw, Request: &resty.Request{Method: request.Method, URL: destination.String(), Header: headers.Clone(), RawRequest: rawRequest}}
	var count int64
	if response.IsSuccess() && sink != nil {
		count, err = copyDownload(operation, sink, raw.Body, maxBytes)
	} else {
		var content []byte
		content, err = io.ReadAll(io.LimitReader(raw.Body, MaxResponseBytes+1))
		if len(content) > MaxResponseBytes {
			content = content[:MaxResponseBytes]
			err = ErrResponseTooLarge
		} else if err != nil {
			err = transferReadError(err)
		}
		response.SetBody(content)
		if !response.IsSuccess() {
			err = errors.Join(err, normalizeError(s.backend, response))
		}
	}
	if raw.Body.Close() != nil {
		err = errors.Join(err, ErrStreamClose)
	}
	if operation.Err() != nil {
		err = errors.Join(err, operation.Err())
	}
	return response, count, err
}

func copyDownload(ctx context.Context, destination io.Writer, source io.Reader, limit int64) (int64, error) {
	buffer := make([]byte, 32<<10)
	var written int64
	for {
		if err := ctx.Err(); err != nil {
			return written, err
		}
		// A one-byte probe detects overflow without writing beyond the caller's cap.
		readSize := min(int64(len(buffer)), limit-written)
		if readSize == 0 {
			readSize = 1
		}
		n, readErr := source.Read(buffer[:int(readSize)])
		if written == limit && n > 0 {
			return written, ErrResponseTooLarge
		}
		if n > 0 {
			count, writeErr := destination.Write(buffer[:n])
			written += int64(max(0, min(count, n)))
			if writeErr != nil || count != n {
				return written, ErrDownloadWrite
			}
		}
		if readErr == io.EOF {
			return written, nil
		}
		if readErr != nil {
			return written, transferReadError(readErr)
		}
	}
}

func transferReadError(err error) error {
	if errors.Is(err, context.DeadlineExceeded) {
		return errors.Join(ErrTransport, context.DeadlineExceeded)
	}
	if errors.Is(err, context.Canceled) {
		return context.Canceled
	}
	return ErrTransport
}

type ownedStream struct {
	source io.ReadCloser
	once   sync.Once
	err    error
	readMu sync.Mutex
	closed bool
}

func (s *ownedStream) Read(p []byte) (int, error) {
	s.readMu.Lock()
	defer s.readMu.Unlock()
	if s.closed {
		return 0, io.ErrClosedPipe
	}
	return s.source.Read(p)
}
func (s *ownedStream) Close() error {
	s.once.Do(func() { s.err = s.source.Close(); s.readMu.Lock(); s.closed = true; s.readMu.Unlock() })
	return s.err
}
func nilStream(value any) bool {
	if value == nil {
		return true
	}
	v := reflect.ValueOf(value)
	switch v.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return v.IsNil()
	default:
		return false
	}
}

type uploadReader struct {
	source    io.Reader
	remaining int64
	mu        sync.Mutex
	err       error
}

func (r *uploadReader) Read(p []byte) (int, error) {
	r.mu.Lock()
	defer r.mu.Unlock()
	if len(p) == 0 {
		return 0, nil
	}
	var n int
	var err error
	if r.remaining == 0 {
		var probe [1]byte
		n, err = r.source.Read(probe[:])
		if n > 0 {
			n, err = 0, ErrUploadTooLarge
		}
	} else {
		n, err = r.source.Read(p[:min(int64(len(p)), r.remaining)])
		r.remaining -= int64(n)
	}
	if err != nil && err != io.EOF {
		if !errors.Is(err, ErrUploadTooLarge) {
			err = ErrUploadRead
		}
		r.err = err
	}
	return n, err
}
func (r *uploadReader) failure() error { r.mu.Lock(); defer r.mu.Unlock(); return r.err }
