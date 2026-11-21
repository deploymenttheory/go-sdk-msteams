package client_test

import (
	"bytes"
	"context"
	"errors"
	"io"
	"mime"
	"mime/multipart"
	"net/http"
	"strings"
	"sync"
	"sync/atomic"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

type countedStream struct {
	io.Reader
	closes   atomic.Int32
	closeErr error
}

func (s *countedStream) Close() error { s.closes.Add(1); return s.closeErr }

type failingReader struct{}

func (failingReader) Read([]byte) (int, error) { return 0, errors.New("private-source-diagnostic") }

type failedWriter struct{}

func (failedWriter) Write(p []byte) (int, error) {
	return min(2, len(p)), errors.New("private-writer-diagnostic")
}

type deadlineReader struct{}

func (deadlineReader) Read([]byte) (int, error) { return 0, context.DeadlineExceeded }

func TestUnitDownloadPreservesBodyDeadline(t *testing.T) {
	t.Parallel()
	body := &countedStream{Reader: deadlineReader{}}
	scope, _, _ := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 200, Header: http.Header{}, Body: body, Request: r}, nil
	})}, "")
	_, count, err := scope.Download(t.Context(), client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}}, io.Discard, 1024)
	if !errors.Is(err, context.DeadlineExceeded) || count != 0 || body.closes.Load() != 1 {
		t.Fatalf("body deadline lost: %v", err)
	}
}

func TestUnitBinaryTransferPreservesBytesAndOwnership(t *testing.T) {
	t.Parallel()
	want := []byte{0, 255, 1, 0, 128, '\r', '\n', '{', '}'}
	for _, backend := range []client.Backend{client.Graph, client.TeamsAdmin} {
		t.Run(string(backend), func(t *testing.T) {
			var calls int
			responseBody := &countedStream{Reader: bytes.NewReader(want)}
			httpClient := &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				if r.Header.Get("Authorization") == "" || r.GetBody != nil {
					t.Error("stream auth/replay guard missing")
				}
				if r.Method == http.MethodPut {
					if r.Header.Get("Content-Type") != "application/zip" {
						t.Error("upload content type changed")
					}
					got, err := io.ReadAll(r.Body)
					if err != nil || !bytes.Equal(got, want) {
						t.Errorf("upload bytes changed: %v", err)
					}
					return &http.Response{StatusCode: 200, Header: http.Header{}, Body: io.NopCloser(strings.NewReader(`{"id":"fixture"}`)), Request: r}, nil
				}
				return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/octet-stream"}}, Body: responseBody, Request: r}, nil
			})}
			scope, _, _ := transportScope(t, backend, httpClient, "")
			source := &countedStream{Reader: bytes.NewReader(want)}
			var result struct {
				ID string `json:"id"`
			}
			response, err := scope.Upload(t.Context(), client.Request{Method: http.MethodPut, Path: []string{"v1.0", "fixture"}}, source, "application/zip", int64(len(want)), &result)
			if err != nil || result.ID != "fixture" || source.closes.Load() != 1 {
				t.Fatalf("upload result/ownership failed: %v", err)
			}
			if response.Request.Header.Get("Authorization") != "" || response.RawResponse.Request.Header.Get("Authorization") != "" {
				t.Fatal("returned metadata retained bearer credential")
			}
			var sink bytes.Buffer
			response, count, err := scope.Download(t.Context(), client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}}, &sink, int64(len(want)))
			if err != nil || count != int64(len(want)) || !bytes.Equal(sink.Bytes(), want) || responseBody.closes.Load() != 1 || response.Header().Get("Content-Type") != "application/octet-stream" || calls != 2 {
				t.Fatalf("download bytes/ownership failed: %v", err)
			}
		})
	}
}

func TestUnitUploadFailuresCloseSourceAndNeverReplay(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		source io.Reader
		limit  int64
		status int
		replay client.ReplaySafety
		target string
		want   error
	}{
		{"oversize", strings.NewReader("four"), 3, 200, client.ReplayNever, "", client.ErrUploadTooLarge},
		{"read", failingReader{}, 8, 200, client.ReplayNever, "", client.ErrUploadRead},
		{"throttled", strings.NewReader("body"), 8, 429, client.ReplayNever, "", nil},
		{"unsafe-replay", strings.NewReader("body"), 8, 200, client.ReplaySafe, "", client.ErrRequest},
		{"foreign", strings.NewReader("body"), 8, 200, client.ReplayNever, "https://foreign.test/v1.0/fixture", client.ErrDestination},
	} {
		t.Run(tc.name, func(t *testing.T) {
			calls := 0
			scope, _, provider := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				_, err := io.Copy(io.Discard, r.Body)
				if err != nil {
					return nil, err
				}
				return &http.Response{StatusCode: tc.status, Header: http.Header{"Retry-After": {"0"}}, Body: io.NopCloser(strings.NewReader(`{"error":{"code":"fixture"}}`)), Request: r}, nil
			})}, "")
			request := client.Request{Method: http.MethodPost, Path: []string{"v1.0", "fixture"}, Replay: tc.replay}
			if tc.target != "" {
				request.Path = nil
				request.URL = tc.target
			}
			source := &countedStream{Reader: tc.source}
			response, err := scope.Upload(t.Context(), request, source, "application/octet-stream", tc.limit, nil)
			if tc.want != nil && !errors.Is(err, tc.want) {
				t.Fatalf("wrong upload error: %v", err)
			}
			if tc.status == 429 && (err == nil || response.StatusCode() != 429) {
				t.Fatal("throttle status lost")
			}
			if source.closes.Load() != 1 || calls > 1 {
				t.Fatal("source leaked or upload replayed")
			}
			if strings.Contains(fmtError(err), "private-") {
				t.Fatal("stream error leaked diagnostic")
			}
			if (tc.name == "foreign" || tc.name == "unsafe-replay") && (calls != 0 || len(provider.requests) != 0) {
				t.Fatal("invalid upload reached authentication")
			}
		})
	}
}
func fmtError(err error) string {
	if err == nil {
		return ""
	}
	return err.Error()
}

type truncatedStream struct {
	sent   bool
	closes int
}

func (s *truncatedStream) Read(p []byte) (int, error) {
	if s.sent {
		return 0, io.ErrUnexpectedEOF
	}
	s.sent = true
	return copy(p, "part"), nil
}
func (s *truncatedStream) Close() error { s.closes++; return nil }

func TestUnitDownloadLimitsTruncationAndPartialWrites(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"limit", "truncated", "writer", "forbidden", "close"} {
		t.Run(name, func(t *testing.T) {
			body := &countedStream{Reader: strings.NewReader("abcdef")}
			var source io.ReadCloser = body
			var sink io.Writer = &bytes.Buffer{}
			limit := int64(10)
			status := 200
			want := client.ErrTransport
			wantCount := int64(4)
			switch name {
			case "limit":
				limit = 3
				want = client.ErrResponseTooLarge
				wantCount = 3
			case "truncated":
				source = &truncatedStream{}
			case "writer":
				sink = failedWriter{}
				want = client.ErrDownloadWrite
				wantCount = 2
			case "forbidden":
				status = 403
				want = nil
				wantCount = 0
			case "close":
				body.closeErr = errors.New("private-close-error")
				want = client.ErrStreamClose
				wantCount = 6
			}
			calls := 0
			scope, _, _ := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				return &http.Response{StatusCode: status, Header: http.Header{}, Body: source, Request: r}, nil
			})}, "")
			response, count, err := scope.Download(t.Context(), client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}}, sink, limit)
			if err == nil || want != nil && !errors.Is(err, want) || count != wantCount || calls != 1 {
				t.Fatalf("partial download wrong: count=%d err=%v", count, err)
			}
			if name == "truncated" {
				if source.(*truncatedStream).closes != 1 {
					t.Fatal("truncated body leaked")
				}
			} else if body.closes.Load() != 1 {
				t.Fatal("response body leaked")
			}
			if name == "forbidden" && len(response.Body()) == 0 {
				t.Fatal("failure metadata lost")
			}
			if strings.Contains(err.Error(), "private-") {
				t.Fatal("raw I/O error leaked")
			}
		})
	}
}

type blockingUpload struct {
	entered chan struct{}
	closed  chan struct{}
	once    sync.Once
	closes  atomic.Int32
}

func (s *blockingUpload) Read([]byte) (int, error) {
	s.once.Do(func() { close(s.entered) })
	<-s.closed
	return 0, io.ErrClosedPipe
}
func (s *blockingUpload) Close() error { s.closes.Add(1); close(s.closed); return nil }

func TestUnitUploadCancellationClosesBlockedSource(t *testing.T) {
	t.Parallel()
	source := &blockingUpload{entered: make(chan struct{}), closed: make(chan struct{})}
	scope, _, _ := transportScope(t, client.Graph, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) { _, err := io.Copy(io.Discard, r.Body); return nil, err })}, "")
	ctx, cancel := context.WithCancel(t.Context())
	defer cancel()
	done := make(chan error, 1)
	go func() {
		_, err := scope.Upload(ctx, client.Request{Method: http.MethodPost, Path: []string{"v1.0", "fixture"}}, source, "application/octet-stream", 1024, nil)
		done <- err
	}()
	select {
	case <-source.entered:
	case <-time.After(time.Second):
		t.Fatal("source never read")
	}
	cancel()
	select {
	case err := <-done:
		if !errors.Is(err, context.Canceled) {
			t.Fatalf("cancellation lost: %v", err)
		}
	case <-time.After(time.Second):
		t.Fatal("cancellation leaked blocked upload")
	}
	if source.closes.Load() != 1 {
		t.Fatal("cancellation double-closed source")
	}
}

func TestUnitMultipartFramingBytesAndFailureCleanup(t *testing.T) {
	t.Parallel()
	for _, invalid := range []bool{false, true} {
		t.Run(map[bool]string{false: "valid", true: "invalid"}[invalid], func(t *testing.T) {
			file := &countedStream{Reader: bytes.NewReader([]byte{0, 255, 3})}
			field := &countedStream{Reader: strings.NewReader("field-value")}
			parts := []client.MultipartPart{{Name: "audio", Filename: `quoted"file.wav`, ContentType: "audio/wav", Body: file}, {Name: "metadata", Body: field}}
			if invalid {
				parts[1].Name = "bad\r\nheader"
			}
			calls := 0
			scope, _, provider := transportScope(t, client.TeamsAdmin, &http.Client{Transport: transportFunc(func(r *http.Request) (*http.Response, error) {
				calls++
				kind, params, err := mime.ParseMediaType(r.Header.Get("Content-Type"))
				if err != nil || kind != "multipart/form-data" || params["boundary"] == "" {
					t.Fatal("multipart boundary missing")
				}
				reader := multipart.NewReader(r.Body, params["boundary"])
				for i := range 2 {
					part, err := reader.NextPart()
					if err != nil {
						t.Fatal(err)
					}
					body, err := io.ReadAll(part)
					if err != nil {
						t.Fatal(err)
					}
					if i == 0 && (!bytes.Equal(body, []byte{0, 255, 3}) || part.FileName() != parts[0].Filename || part.Header.Get("Content-Type") != "audio/wav") {
						t.Fatal("file part changed")
					}
					if i == 1 && (string(body) != "field-value" || part.FormName() != "metadata") {
						t.Fatal("form field changed")
					}
				}
				if _, err := reader.NextPart(); err != io.EOF {
					t.Fatal("multipart final boundary missing")
				}
				return &http.Response{StatusCode: 204, Header: http.Header{}, Body: http.NoBody, Request: r}, nil
			})}, "")
			_, err := scope.UploadMultipart(t.Context(), client.Request{Method: http.MethodPost, Path: []string{"fixture", "audio"}}, parts, 4096, nil)
			if invalid && (!errors.Is(err, client.ErrRequest) || calls != 0 || len(provider.requests) != 0) {
				t.Fatal("invalid multipart reached transport")
			}
			if !invalid && (err != nil || calls != 1) {
				t.Fatalf("multipart failed: %v", err)
			}
			if file.closes.Load() != 1 || field.closes.Load() != 1 {
				t.Fatal("multipart ownership leaked")
			}
		})
	}
}

func TestUnitMultipartRejectsRepeatedSourceWithoutDoubleClose(t *testing.T) {
	t.Parallel()
	source := &countedStream{Reader: strings.NewReader("fixture")}
	var scope *client.Scope
	_, err := scope.UploadMultipart(t.Context(), client.Request{}, []client.MultipartPart{{Name: "a", Body: source}, {Name: "b", Body: source}}, 1024, nil)
	if !errors.Is(err, client.ErrRequest) || source.closes.Load() != 1 {
		t.Fatal("reused source was not rejected and closed once")
	}
}

func TestUnitUploadReachesHTTPServerBeforeSourceCompletes(t *testing.T) {
	t.Parallel()
	firstChunk := make(chan struct{})
	httpClient := localBackend(t, func(w http.ResponseWriter, r *http.Request) {
		prefix := make([]byte, 3)
		if _, err := io.ReadFull(r.Body, prefix); err != nil || string(prefix) != "one" {
			t.Error("first chunk changed")
			w.WriteHeader(400)
			return
		}
		close(firstChunk)
		tail, err := io.ReadAll(r.Body)
		if err != nil || string(tail) != "two" {
			t.Error("tail changed")
			w.WriteHeader(400)
			return
		}
		w.WriteHeader(204)
	}, nil)
	scope, _, _ := transportScope(t, client.Graph, httpClient, "")
	reader, writer := io.Pipe()
	t.Cleanup(func() { _ = reader.Close(); _ = writer.Close() })
	done := make(chan error, 1)
	go func() {
		_, err := scope.Upload(t.Context(), client.Request{Method: http.MethodPut, Path: []string{"v1.0", "fixture"}}, reader, "application/octet-stream", 6, nil)
		done <- err
	}()
	if _, err := writer.Write([]byte("one")); err != nil {
		t.Fatal(err)
	}
	select {
	case <-firstChunk:
	case <-time.After(2 * time.Second):
		t.Fatal("upload buffered entire source before HTTP")
	}
	if _, err := writer.Write([]byte("two")); err != nil {
		t.Fatal(err)
	}
	if err := writer.Close(); err != nil {
		t.Fatal(err)
	}
	select {
	case err := <-done:
		if err != nil {
			t.Fatal(err)
		}
	case <-time.After(2 * time.Second):
		t.Fatal("upload failed to finish")
	}
}
