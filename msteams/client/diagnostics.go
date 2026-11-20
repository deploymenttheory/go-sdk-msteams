package client

import (
	"context"
	"errors"
	"net/http"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/go-resty/resty/v2"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/codes"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var ErrSessionOption = errors.New("invalid session configuration option")

type sessionOptions struct {
	policy  RequestPolicy
	logger  *zap.Logger
	tracer  trace.TracerProvider
	debug   bool
	headers http.Header
}

func defaultSessionOptions() sessionOptions { return sessionOptions{logger: zap.NewNop()} }

// WithDiagnostics borrows observability dependencies. Nil logger disables logs;
// nil tracer disables tracing. Debug emits sanitized attempt events, never raw
// Resty debug output. Caller-added logger fields/exporter behavior remain caller-owned.
func WithDiagnostics(logger *zap.Logger, tracer trace.TracerProvider, debug bool) SessionOption {
	if logger == nil {
		logger = zap.NewNop()
	}
	return func(o *sessionOptions) error {
		if tracer != nil && nilStream(tracer) {
			return ErrSessionOption
		}
		o.logger, o.tracer, o.debug = logger, tracer, debug
		return nil
	}
}

// WithDefaultHeaders copies management-request defaults. They never apply to
// token exchanges. Per-request values replace defaults, subject to the same guard.
func WithDefaultHeaders(headers http.Header) SessionOption {
	copy := headers.Clone()
	return func(o *sessionOptions) error {
		validated, err := requestHeaders(copy)
		if err != nil {
			return ErrSessionOption
		}
		o.headers = validated
		return nil
	}
}

func (s *Session) requestHeaders(headers http.Header) (http.Header, error) {
	validated, err := requestHeaders(headers)
	if err != nil {
		return nil, err
	}
	combined := s.requests.settings.headers.Clone()
	if combined == nil {
		combined = make(http.Header)
	}
	for key, values := range validated {
		combined[key] = values
	}
	if combined.Get("Accept") == "" {
		combined.Set("Accept", "application/json")
	}
	return combined, nil
}

type observation struct {
	logger          *zap.Logger
	span            trace.Span
	started         time.Time
	backend, method string
	debug           bool
	attempts        int
}

func (s *Scope) observe(ctx context.Context, method string) (context.Context, *observation) {
	if ctx == nil || s == nil || s.session == nil || s.session.requests == nil {
		return ctx, nil
	}
	settings := s.session.requests.settings
	backend := "unknown"
	if s.backend == Graph || s.backend == TeamsAdmin {
		backend = string(s.backend)
	}
	switch method {
	case http.MethodGet, http.MethodHead, http.MethodPost, http.MethodPut, http.MethodPatch, http.MethodDelete, http.MethodOptions:
	default:
		method = "unknown"
	}
	o := &observation{logger: settings.logger, started: time.Now(), backend: backend, method: method, debug: settings.debug}
	if settings.tracer != nil {
		ctx, o.span = settings.tracer.Tracer("github.com/deploymenttheory/go-sdk-msteams").Start(ctx, "msteams."+backend+".request", trace.WithSpanKind(trace.SpanKindClient), trace.WithAttributes(attribute.String("msteams.backend", backend), attribute.String("http.request.method", method)))
	}
	return ctx, o
}

func (o *observation) attempt(response *resty.Response, err error) {
	if o == nil {
		return
	}
	o.attempts++
	status := responseStatus(response)
	if o.debug {
		o.logger.Debug("msteams request attempt", zap.String("backend", o.backend), zap.String("method", o.method), zap.Int("attempt", o.attempts), zap.Int("status", status), zap.String("outcome", safeOutcome(err)))
	}
	if o.span != nil {
		o.span.AddEvent("http.attempt", trace.WithAttributes(attribute.Int("attempt", o.attempts), attribute.Int("http.response.status_code", status), attribute.String("outcome", safeOutcome(err))))
	}
}

func (o *observation) finish(response *resty.Response, err error) {
	if o == nil {
		return
	}
	outcome := safeOutcome(err)
	o.logger.Info("msteams request completed", zap.String("backend", o.backend), zap.String("method", o.method), zap.Int("status", responseStatus(response)), zap.Int("attempts", o.attempts), zap.Duration("elapsed", time.Since(o.started)), zap.String("outcome", outcome))
	if o.span != nil {
		o.span.SetAttributes(attribute.Int("http.response.status_code", responseStatus(response)), attribute.Int("msteams.attempts", o.attempts), attribute.String("msteams.outcome", outcome))
		if err != nil {
			o.span.SetStatus(codes.Error, outcome)
		}
		o.span.End()
	}
}

func responseStatus(response *resty.Response) int {
	if response == nil {
		return 0
	}
	return response.StatusCode()
}

// Never format an error, unwrap it into diagnostics, or emit service codes/IDs.
func safeOutcome(err error) string {
	switch {
	case err == nil:
		return "success"
	case errors.Is(err, context.Canceled):
		return "cancelled"
	case errors.Is(err, context.DeadlineExceeded):
		return "deadline"
	case errors.Is(err, auth.ErrUnauthenticated):
		return "unauthenticated"
	case errors.Is(err, auth.ErrForbidden):
		return "forbidden"
	case errors.Is(err, ErrDestination):
		return "destination_rejected"
	case errors.Is(err, ErrRequest), errors.Is(err, ErrSessionOption):
		return "invalid_request"
	case errors.Is(err, ErrClosed):
		return "closed"
	case errors.Is(err, ErrRetryBudget):
		return "retry_budget"
	case errors.Is(err, ErrDecode):
		return "decode"
	case errors.Is(err, ErrUploadTooLarge), errors.Is(err, ErrResponseTooLarge):
		return "size_limit"
	case errors.Is(err, ErrUploadRead), errors.Is(err, ErrDownloadWrite), errors.Is(err, ErrStreamClose):
		return "stream"
	case errors.Is(err, ErrTransport):
		return "transport"
	default:
		return "failed"
	}
}
