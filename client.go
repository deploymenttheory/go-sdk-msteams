package msteams

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"reflect"
	"time"

	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/config"
	graphapi "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api"
	teamsadminapi "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api"
)

// AuthConfig selects an explicit portable authentication method.
type AuthConfig = auth.Config

// Client owns one locally configured authentication session and typed backend groups.
// Resource operations are added by their implementation tickets. Construction never
// obtains tokens, reads credential files or discovers remote service routes.
type Client struct {
	GraphAPI      *graphapi.Services
	TeamsAdminAPI *teamsadminapi.Services
	session       *client.Session
	logger        *zap.Logger
	tracer        trace.TracerProvider
}

// ErrInvalidOption identifies an invalid client option without exposing its value.
var ErrInvalidOption = errors.New("invalid msteams client option")

type options struct {
	auth       *auth.Config
	tokensJSON string
	timeout    string
	http       *http.Client
	logger     *zap.Logger
	tracer     trace.TracerProvider
}

// ClientOption changes constructor configuration before local validation.
type ClientOption func(*options) error

// WithAuthConfig replaces authentication configuration, including environment defaults.
func WithAuthConfig(cfg *AuthConfig) ClientOption {
	return func(o *options) error {
		o.auth = cfg.Clone()
		o.tokensJSON = ""
		return nil
	}
}

// WithTokenProvider replaces authentication with a provider borrowed from the
// caller and one explicit identity binding. Client Close never closes the provider.
func WithTokenProvider(binding auth.Binding, provider auth.Provider) ClientOption {
	return WithAuthConfig(&AuthConfig{Method: auth.ExternalProvider, TenantID: binding.TenantID,
		Cloud: binding.Cloud, IdentityID: binding.IdentityID, TokenMode: binding.Mode, Provider: provider})
}

// WithHTTPTimeout sets a positive total HTTP request timeout.
func WithHTTPTimeout(timeout time.Duration) ClientOption {
	return func(o *options) error {
		o.timeout = timeout.String()
		return nil
	}
}

// WithHTTPClient borrows a round tripper but copies the client value. The SDK does
// not close the injected transport or mutate the caller's client configuration.
func WithHTTPClient(httpClient *http.Client) ClientOption {
	return func(o *options) error {
		if httpClient == nil || (httpClient.Transport != nil && nilValue(httpClient.Transport)) {
			return fmt.Errorf("http client and a non-nil transport value are required: %w", ErrInvalidOption)
		}
		cloned := *httpClient
		o.http = &cloned
		return nil
	}
}

// WithLogger uses a caller-owned logger. The default is a no-op logger.
func WithLogger(logger *zap.Logger) ClientOption {
	return func(o *options) error {
		if logger == nil {
			return fmt.Errorf("logger is required: %w", ErrInvalidOption)
		}
		o.logger = logger
		return nil
	}
}

// WithTracerProvider opts into a caller-owned tracer provider. The SDK does not
// initialize global telemetry or shut down the provider on client close.
func WithTracerProvider(provider trace.TracerProvider) ClientOption {
	return func(o *options) error {
		if nilValue(provider) {
			return fmt.Errorf("tracer provider is required: %w", ErrInvalidOption)
		}
		o.tracer = provider
		return nil
	}
}

// NewClient constructs a client from explicit configuration, without reading MSTEAMS
// environment variables. Later options override earlier constructor values.
func NewClient(authConfig *AuthConfig, opts ...ClientOption) (*Client, error) {
	return newClient(options{auth: authConfig.Clone()}, opts)
}

// NewClientFromEnv reads documented MSTEAMS variables, then applies explicit options.
// Parsing and validation occur after options, allowing replacement of invalid defaults.
func NewClientFromEnv(opts ...ClientOption) (*Client, error) {
	env := config.FromEnv()
	return newClient(options{auth: env.Auth, tokensJSON: env.TokensJSON, timeout: env.Timeout}, opts)
}

func newClient(o options, opts []ClientOption) (*Client, error) {
	o.logger = zap.NewNop()
	for _, option := range opts {
		if option == nil {
			return nil, fmt.Errorf("nil option: %w", ErrInvalidOption)
		}
		if err := option(&o); err != nil {
			return nil, err
		}
	}
	if o.tokensJSON != "" {
		if err := json.Unmarshal([]byte(o.tokensJSON), &o.auth.AccessTokens); err != nil {
			return nil, &auth.ConfigError{Field: "access_tokens", Reason: "must be an audience-to-token JSON object"}
		}
	}
	if err := o.auth.Validate(); err != nil {
		return nil, err
	}
	if o.http != nil && o.http.Timeout < 0 && o.timeout == "" {
		return nil, fmt.Errorf("injected http timeout must not be negative: %w", ErrInvalidOption)
	}
	timeout := 30 * time.Second
	if o.http != nil && o.http.Timeout > 0 {
		timeout = o.http.Timeout
	}
	if o.timeout != "" {
		parsed, err := time.ParseDuration(o.timeout)
		if err != nil || parsed <= 0 {
			return nil, fmt.Errorf("http timeout must be a positive duration: %w", ErrInvalidOption)
		}
		timeout = parsed
	}
	var cleanup func()
	if o.http == nil || o.http.Transport == nil {
		transport := &http.Transport{
			ForceAttemptHTTP2: true, MaxIdleConns: 100,
			IdleConnTimeout: 90 * time.Second, TLSHandshakeTimeout: 10 * time.Second,
			ExpectContinueTimeout: time.Second,
		}
		if o.http == nil {
			o.http = &http.Client{}
		}
		o.http.Transport = transport
		cleanup = transport.CloseIdleConnections
	}
	o.http.Timeout = timeout
	session, err := client.NewSession(o.auth, o.http, cleanup)
	if err != nil {
		if cleanup != nil {
			cleanup()
		}
		return nil, err
	}
	graph, err := graphapi.NewServices(session)
	if err != nil {
		session.Close()
		return nil, err
	}
	admin, err := teamsadminapi.NewServices(session)
	if err != nil {
		session.Close()
		return nil, err
	}
	return &Client{GraphAPI: graph, TeamsAdminAPI: admin, session: session, logger: o.logger, tracer: o.tracer}, nil
}

// Logger returns the configured logger without transferring lifecycle ownership.
func (c *Client) Logger() *zap.Logger { return c.logger }

// TracerProvider returns the opt-in provider, or nil when tracing is not configured.
func (c *Client) TracerProvider() trace.TracerProvider { return c.tracer }

// Close cancels local work and clears owned session material. It is idempotent and
// does not revoke Entra tokens or shut down caller-owned logging/telemetry resources.
func (c *Client) Close() error {
	if c != nil && c.session != nil {
		c.session.Close()
	}
	return nil
}

// nilValue also rejects interfaces carrying typed-nil injected dependencies.
func nilValue(value any) bool {
	if value == nil {
		return true
	}
	reflected := reflect.ValueOf(value)
	switch reflected.Kind() {
	case reflect.Chan, reflect.Func, reflect.Interface, reflect.Map, reflect.Pointer, reflect.Slice:
		return reflected.IsNil()
	default:
		return false
	}
}
