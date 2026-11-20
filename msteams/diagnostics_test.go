package msteams_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"strings"
	"sync"
	"testing"
	"time"

	msteams "github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.opentelemetry.io/otel/trace/noop"
	"go.uber.org/zap"
	"go.uber.org/zap/zaptest/observer"
)

type diagnosticTransport func(*http.Request) (*http.Response, error)

func (f diagnosticTransport) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

type recordedTraces struct {
	trace.TracerProvider
	mu    sync.Mutex
	spans []*recordedSpan
}
type recordedTracer struct {
	trace.Tracer
	p *recordedTraces
}
type recordedSpan struct {
	trace.Span
	mu         sync.Mutex
	name       string
	attributes []attribute.KeyValue
	events     []string
	errors     []string
	status     string
	ends       int
}

func newRecordedTraces() *recordedTraces {
	return &recordedTraces{TracerProvider: noop.NewTracerProvider()}
}
func (p *recordedTraces) Tracer(name string, options ...trace.TracerOption) trace.Tracer {
	return &recordedTracer{Tracer: p.TracerProvider.Tracer(name, options...), p: p}
}
func (t *recordedTracer) Start(ctx context.Context, name string, options ...trace.SpanStartOption) (context.Context, trace.Span) {
	ctx, base := t.Tracer.Start(ctx, name, options...)
	configuration := trace.NewSpanStartConfig(options...)
	span := &recordedSpan{Span: base, name: name, attributes: configuration.Attributes()}
	t.p.mu.Lock()
	t.p.spans = append(t.p.spans, span)
	t.p.mu.Unlock()
	return trace.ContextWithSpan(ctx, span), span
}
func (s *recordedSpan) SetAttributes(values ...attribute.KeyValue) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.attributes = append(s.attributes, values...)
}
func (s *recordedSpan) AddEvent(name string, options ...trace.EventOption) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.events = append(s.events, name)
	configuration := trace.NewEventConfig(options...)
	s.attributes = append(s.attributes, configuration.Attributes()...)
}
func (s *recordedSpan) RecordError(err error, _ ...trace.EventOption) {
	s.mu.Lock()
	defer s.mu.Unlock()
	if err != nil {
		s.errors = append(s.errors, err.Error())
	}
}
func (s *recordedSpan) SetStatus(_ codes.Code, description string) {
	s.mu.Lock()
	defer s.mu.Unlock()
	s.status = description
}
func (s *recordedSpan) End(...trace.SpanEndOption) { s.mu.Lock(); defer s.mu.Unlock(); s.ends++ }

func TestRequestDiagnosticsExcludeSecretsAcrossOutcomes(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"json", "forbidden", "provider", "transport", "retry", "upload", "download", "cancelled", "invalid-method"} {
		t.Run(scenario, func(t *testing.T) {
			secret := "fixture-private-" + scenario
			core, logs := observer.New(zap.DebugLevel)
			traces := newRecordedTraces()
			calls := 0
			provider := &testProvider{acquire: func(ctx context.Context, _ auth.TokenRequest) (auth.Token, error) {
				if scenario == "provider" {
					return auth.Token{}, errors.New(secret + " provider assertion password key")
				}
				if _, ok := trace.SpanFromContext(ctx).(*recordedSpan); !ok {
					t.Error("span context did not reach authentication")
				}
				return auth.Token{AccessToken: secret + "-token", ExpiresAt: time.Now().Add(time.Hour)}, nil
			}}
			transport := diagnosticTransport(func(r *http.Request) (*http.Response, error) {
				calls++
				if _, ok := trace.SpanFromContext(r.Context()).(*recordedSpan); !ok {
					t.Error("span context did not reach HTTP")
				}
				if r.Header.Get("Authorization") != "Bearer "+secret+"-token" {
					t.Error("credential was not applied")
				}
				if scenario == "transport" {
					return nil, errors.New(secret + " transport URL")
				}
				if scenario == "upload" {
					if _, err := io.Copy(io.Discard, r.Body); err != nil {
						return nil, err
					}
				}
				status := 200
				if scenario == "forbidden" {
					status = 403
				}
				if scenario == "retry" && calls == 1 {
					status = 503
				}
				return &http.Response{StatusCode: status, Header: http.Header{"Request-Id": {secret + "-request"}, "Retry-After": {"0"}}, Body: io.NopCloser(strings.NewReader(`{"error":{"code":"` + secret + `","message":"` + secret + `"}}`)), Request: r}, nil
			})
			cfg := &auth.Config{Method: auth.ExternalProvider, TenantID: secret + "-tenant", IdentityID: secret + "-identity", TokenMode: auth.Application, Provider: provider}
			sdk, err := msteams.NewClient(cfg, msteams.WithTransport(transport), msteams.WithLogger(zap.New(core)), msteams.WithTracerProvider(traces), msteams.WithDebug(), msteams.WithRequestPolicy(client.RequestPolicy{MaxAttempts: 2, BaseDelay: time.Nanosecond, MaxDelay: time.Nanosecond}))
			if err != nil {
				t.Fatal(err)
			}
			closeClient(t, sdk)
			ctx := t.Context()
			if scenario == "cancelled" {
				cancelled, cancel := context.WithCancel(ctx)
				cancel()
				ctx = cancelled
			}
			request := client.Request{Method: http.MethodPost, Path: []string{"v1.0", secret}, Header: http.Header{"X-Fixture": {secret}}, Body: map[string]string{"password": secret, "assertion": secret, "private_key": secret}}
			scope := sdk.GraphAPI.Scope()
			switch scenario {
			case "upload":
				request.Body = nil
				_, err = scope.Upload(ctx, request, io.NopCloser(strings.NewReader(secret)), "application/octet-stream", 1024, nil)
			case "download":
				request.Body = nil
				request.Method = http.MethodGet
				_, _, err = scope.Download(ctx, request, io.Discard, 1024)
			default:
				if scenario == "retry" {
					request.Replay = client.ReplaySafe
				}
				if scenario == "invalid-method" {
					request.Method = secret
				}
				_, err = scope.Do(ctx, request, nil)
			}
			wantSuccess := scenario == "json" || scenario == "retry" || scenario == "upload" || scenario == "download"
			if (err == nil) != wantSuccess {
				t.Fatalf("scenario result changed: %v", err)
			}
			if len(traces.spans) != 1 || traces.spans[0].ends != 1 {
				t.Fatal("span was not started/ended exactly once")
			}
			span := traces.spans[0]
			var rendered []map[string]any
			for _, entry := range logs.All() {
				rendered = append(rendered, map[string]any{"message": entry.Message, "fields": entry.ContextMap()})
			}
			captured, _ := json.Marshal(map[string]any{"logs": rendered, "name": span.name, "attributes": span.attributes, "events": span.events, "errors": span.errors, "status": span.status})
			if strings.Contains(string(captured), secret) {
				t.Fatal("SDK diagnostics exposed credential/tenant/request/response data")
			}
			if logs.FilterMessage("msteams request completed").Len() != 1 || logs.FilterMessage("msteams request attempt").Len() != calls || len(span.events) != calls {
				t.Fatal("logical request/attempt observations missing")
			}
			if scenario == "retry" && calls != 2 {
				t.Fatal("retry fixture did not exercise both attempts")
			}
		})
	}
}

func TestDiagnosticsRemainOptInAndBorrowed(t *testing.T) {
	t.Parallel()
	core, logs := observer.New(zap.DebugLevel)
	logger := zap.New(core)
	transport := diagnosticTransport(func(r *http.Request) (*http.Response, error) {
		return &http.Response{StatusCode: 204, Header: http.Header{}, Body: http.NoBody, Request: r}, nil
	})
	cfg := &auth.Config{Method: auth.SuppliedTokens, TenantID: "fixture", IdentityID: "fixture", TokenMode: auth.Application, AccessTokens: map[string]auth.Token{auth.GraphAudience: {AccessToken: "fixture", ExpiresAt: time.Now().Add(time.Hour)}}}
	sdk, err := msteams.NewClient(cfg, msteams.WithTransport(transport), msteams.WithLogger(logger))
	if err != nil {
		t.Fatal(err)
	}
	if _, err := sdk.GraphAPI.Scope().Do(t.Context(), client.Request{Method: http.MethodGet, Path: []string{"v1.0", "fixture"}}, nil); err != nil {
		t.Fatal(err)
	}
	if logs.FilterMessage("msteams request attempt").Len() != 0 || logs.FilterMessage("msteams request completed").Len() != 1 {
		t.Fatal("debug was implicitly enabled")
	}
	if err := sdk.Close(); err != nil {
		t.Fatal(err)
	}
	logger.Info("caller still owns logger")
	if logs.FilterMessage("caller still owns logger").Len() != 1 {
		t.Fatal("client shut down borrowed logger")
	}
}
