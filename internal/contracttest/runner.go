package contracttest

import (
	"context"
	"crypto/sha256"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"net/url"
	"path"
	"reflect"
	"regexp"
	"slices"
	"sort"
	"strconv"
	"strings"
	"sync"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

var (
	ErrFixture = errors.New("contracttest: invalid synthetic fixture")
	ErrSetup   = errors.New("contracttest: sdk setup or cleanup failed")
	ErrWire    = errors.New("contracttest: unexpected offline request")
	label      = regexp.MustCompile(`^[a-z][a-z0-9_-]{0,95}$`)
	command    = regexp.MustCompile(`^[A-Za-z]+-[A-Za-z][A-Za-z0-9]*$`)
	setName    = regexp.MustCompile(`^[A-Za-z_][A-Za-z0-9_]{0,95}$`)
)

// Mapping is an exact PowerShell command and observed parameter-set name.
// A mapping names the exercised scope; it does not assert full command parity.
type Mapping struct {
	Command      string `json:"command"`
	ParameterSet string `json:"parameter_set"`
}

// WireRequest keeps the fully escaped path and repeated query values exact.
// Headers are explicit assertions; transport-owned headers can remain unspecified.
// A nil JSON body means no body, which differs from JSON null, {}, or [].
// Binary asserts exact bytes instead of JSON structure. JSON and Binary are
// mutually exclusive; nil for both requires an absent or zero-length body.
type WireRequest struct {
	Method   string
	URL      string
	Audience string
	Headers  http.Header
	JSON     json.RawMessage
	Binary   []byte
}

// Reply is a synthetic HTTP response. Bodies may deliberately be malformed.
type Reply struct {
	Status  int
	Headers http.Header
	Body    []byte
}

// Exchange describes one ordered HTTP attempt, including discovery or read-back.
type Exchange struct {
	Request WireRequest
	Reply   Reply
}

// Observation adapts any typed service result to the runner without logging it.
type Observation struct {
	Value    any
	Response *resty.Response
	Err      error
}

func (Observation) String() string   { return "offline contract observation (redacted)" }
func (Observation) GoString() string { return "offline contract observation (redacted)" }

// Expected compares only the explicitly selected result fields. Status zero
// requires no response. Error uses errors.Is; APIError compares every normalized
// diagnostic field. With neither error assertion, the operation must succeed.
// JSON nil skips value comparison; Body nil skips raw response comparison.
type Expected struct {
	JSON     json.RawMessage
	Body     []byte
	Status   int
	Headers  http.Header
	Error    error
	APIError *client.APIError
}

// Case runs a callback against a real root client with synthetic credentials and
// an in-memory transport. Scope is "service", "primitive", or "local". Local
// cases must have no exchanges and are checked for zero credential acquisition.
// Source is a repository-relative contract reference, optionally a JSON pointer.
// Callbacks must finish their goroutines before returning; cases own no background work.
type Case struct {
	Name     string
	Family   string
	Scope    string
	Source   string
	Mappings []Mapping
	Cloud    string
	Mode     auth.Mode
	Steps    []Exchange
	Invoke   func(context.Context, *msteams.Client) Observation
	Want     Expected
}

func (Case) String() string   { return "synthetic offline contract case" }
func (Case) GoString() string { return "synthetic offline contract case" }

// Result retains structural evidence only, never observations, URLs or tokens.
type Result struct {
	name, family, scope, source string
	cloud, mode                 string
	mappings                    []Mapping
	differences                 []Difference
	httpCalls, tokenCalls       int
}

func (r Result) Passed() bool { return r.name != "" && len(r.differences) == 0 }

// Differences returns an independent list; paths have no server-provided names.
func (r Result) Differences() []Difference { return slices.Clone(r.differences) }

func (r Result) String() string   { return "offline contract result (structural evidence only)" }
func (r Result) GoString() string { return r.String() }

// Suite registers resource-owned cases. Registration precedes execution; the
// zero value works. Run sorts by family/name independently of registration order.
// A Suite must not be changed concurrently with Run.
type Suite struct{ cases map[string]Case }

func (s *Suite) Register(c Case) error {
	if s == nil || !validCase(c) {
		return ErrFixture
	}
	key := c.Family + "/" + c.Name
	if _, exists := s.cases[key]; exists {
		return ErrFixture
	}
	if s.cases == nil {
		s.cases = map[string]Case{}
	}
	s.cases[key] = c
	return nil
}

func (s *Suite) Run(ctx context.Context) ([]Result, error) {
	if s == nil || ctx == nil {
		return nil, ErrFixture
	}
	keys := make([]string, 0, len(s.cases))
	for key := range s.cases {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	results := make([]Result, 0, len(keys))
	for _, key := range keys {
		result, err := Run(ctx, s.cases[key])
		if err != nil {
			return results, err
		}
		results = append(results, result)
	}
	return results, nil
}

// Run exercises a single case. Comparison failures are in Result, while invalid
// fixture/setup errors are returned separately and cannot become passing evidence.
func Run(ctx context.Context, c Case) (Result, error) {
	if ctx == nil || !validCase(c) {
		return Result{}, ErrFixture
	}
	environment, err := auth.ResolveEnvironment(c.Cloud, nil)
	if err != nil {
		return Result{}, ErrFixture
	}
	mode := c.Mode
	if mode == "" {
		mode = auth.Application
	}
	h := &harness{steps: c.Steps, differences: []Difference{},
		binding: auth.Binding{TenantID: "fixture-tenant", IdentityID: "fixture-identity", Cloud: environment.BindingCloud, Mode: mode}}
	sdk, err := msteams.NewClient(&auth.Config{Method: auth.ExternalProvider,
		TenantID: h.binding.TenantID, IdentityID: h.binding.IdentityID, Cloud: c.Cloud,
		TokenMode: h.binding.Mode, Provider: h}, msteams.WithHTTPClient(&http.Client{Transport: h}))
	if err != nil {
		return Result{}, ErrSetup
	}
	// Also release the client if the test callback panics. Close is idempotent.
	defer func() { _ = sdk.Close() }()
	observed := c.Invoke(ctx, sdk)
	if err := sdk.Close(); err != nil {
		return Result{}, ErrSetup
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	if h.httpCalls < len(c.Steps) {
		h.add("requests", "missing-attempts")
	}
	if len(c.Steps) == 0 && h.tokenCalls != 0 {
		h.add("tokens", "unexpected-acquisition")
	}
	return Result{name: c.Name, family: c.Family, scope: c.Scope, source: c.Source,
		cloud: environment.Name, mode: string(mode),
		mappings: slices.Clone(c.Mappings), httpCalls: h.httpCalls, tokenCalls: h.tokenCalls,
		differences: append(slices.Clone(h.differences), compareObservation(c.Want, observed)...)}, nil
}

func validCase(c Case) bool {
	validLabels := label.MatchString(c.Name) && label.MatchString(c.Family)
	validScope := c.Scope == "service" || c.Scope == "primitive" || c.Scope == "local"
	if !validLabels || !validScope || c.Invoke == nil || !validSource(c.Source) {
		return false
	}
	if c.Mode != "" && c.Mode != auth.Application && c.Mode != auth.Delegated {
		return false
	}
	if c.Scope == "local" && len(c.Steps) != 0 {
		return false
	}
	for _, mapping := range c.Mappings {
		if !command.MatchString(mapping.Command) || !setName.MatchString(mapping.ParameterSet) {
			return false
		}
	}
	for _, step := range c.Steps {
		if step.Request.JSON != nil && step.Request.Binary != nil {
			return false
		}
		u, err := url.Parse(step.Request.URL)
		if err != nil || u.Scheme != "https" || u.Host == "" || u.User != nil || u.Fragment != "" {
			return false
		}
		if _, err := url.ParseQuery(u.RawQuery); err != nil {
			return false
		}
		if step.Request.Method == "" || step.Request.Audience == "" || step.Reply.Status < 200 || step.Reply.Status > 599 {
			return false
		}
		if step.Request.JSON != nil {
			if _, ok := parseJSON(step.Request.JSON); !ok {
				return false
			}
		}
	}
	if c.Want.JSON != nil {
		_, ok := parseJSON(c.Want.JSON)
		return ok
	}
	return true
}

func validSource(source string) bool {
	file, pointer, _ := strings.Cut(source, "#")
	validFile := strings.HasPrefix(file, "contracts/") && path.Clean(file) == file && strings.HasSuffix(file, ".json")
	validPointer := pointer == "" || strings.HasPrefix(pointer, "/")
	return validFile && validPointer && !strings.ContainsAny(source, "\\\r\n\x00")
}

type harness struct {
	mu                    sync.Mutex
	steps                 []Exchange
	binding               auth.Binding
	differences           []Difference
	httpCalls, tokenCalls int
}

func (h *harness) add(path, kind string) {
	h.differences = append(h.differences, Difference{Path: path, Kind: kind})
}

func fixtureToken(audience string) string {
	digest := sha256.Sum256([]byte(audience))
	return "offline-fixture-" + hex.EncodeToString(digest[:])
}

func (h *harness) AcquireToken(ctx context.Context, request auth.TokenRequest) (auth.Token, error) {
	if err := ctx.Err(); err != nil {
		return auth.Token{}, err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	h.tokenCalls++
	if request != h.binding.Request(request.Audience) {
		h.add("token.binding", "different-value")
	}
	for _, step := range h.steps {
		if step.Request.Audience == request.Audience {
			return auth.Token{AccessToken: fixtureToken(request.Audience), ExpiresAt: time.Now().Add(time.Hour)}, nil
		}
	}
	h.add("token.audience", "unexpected-value")
	return auth.Token{}, auth.ErrTokenMissing
}

func (h *harness) RoundTrip(request *http.Request) (*http.Response, error) {
	if err := request.Context().Err(); err != nil {
		return nil, err
	}
	h.mu.Lock()
	defer h.mu.Unlock()
	index := h.httpCalls
	h.httpCalls++
	if index >= len(h.steps) {
		h.add("requests", "unexpected-attempt")
		return nil, ErrWire
	}
	step := h.steps[index]
	prefix := "request[" + strconv.Itoa(index) + "]"
	h.differences = append(h.differences, compareRequest(step.Request, request, prefix)...)
	return &http.Response{StatusCode: step.Reply.Status, Header: step.Reply.Headers.Clone(),
		Body: io.NopCloser(strings.NewReader(string(step.Reply.Body))), Request: request}, nil
}

func compareRequest(want WireRequest, got *http.Request, prefix string) []Difference {
	differences := []Difference{}
	expected, _ := url.Parse(want.URL) // Already validated before any SDK call.
	actualQuery, queryErr := url.ParseQuery(got.URL.RawQuery)
	for _, field := range []struct {
		name string
		ok   bool
	}{
		{name: "method", ok: got.Method == want.Method},
		{name: "origin", ok: got.URL.Scheme == expected.Scheme && got.URL.Host == expected.Host && got.URL.User == nil},
		{name: "path", ok: got.URL.EscapedPath() == expected.EscapedPath()},
		{name: "query", ok: queryErr == nil && reflect.DeepEqual(actualQuery, expected.Query())},
		{name: "fragment", ok: got.URL.Fragment == ""},
		{name: "audience", ok: got.Header.Get("Authorization") == "Bearer "+fixtureToken(want.Audience)},
	} {
		if !field.ok {
			differences = append(differences, Difference{Path: prefix + "." + field.name, Kind: "different-value"})
		}
	}
	differences = append(differences, compareHeaders(want.Headers, got.Header, prefix+".headers")...)
	raw := []byte{}
	if got.Body != nil {
		defer func() { _ = got.Body.Close() }()
		var err error
		raw, err = io.ReadAll(io.LimitReader(got.Body, client.MaxResponseBytes+1))
		if err != nil || len(raw) > client.MaxResponseBytes {
			return append(differences, Difference{Path: prefix + ".body", Kind: "read-failed"})
		}
	}
	if want.Binary != nil {
		if !slices.Equal(want.Binary, raw) {
			differences = append(differences, Difference{Path: prefix + ".body", Kind: "different-value"})
		}
		return differences
	}
	if want.JSON == nil {
		if len(raw) != 0 {
			differences = append(differences, Difference{Path: prefix + ".body", Kind: "unexpected-body"})
		}
		return differences
	}
	return append(differences, compareJSON(want.JSON, raw, prefix+".body")...)
}

func compareHeaders(want, got http.Header, prefix string) []Difference {
	keys := make([]string, 0, len(want))
	for key := range want {
		keys = append(keys, key)
	}
	sort.Strings(keys)
	differences := []Difference{}
	for index, key := range keys {
		if !slices.Equal(want[key], got.Values(key)) {
			differences = append(differences, Difference{Path: prefix + "[" + strconv.Itoa(index) + "]", Kind: "different-value"})
		}
	}
	return differences
}

func compareObservation(want Expected, got Observation) []Difference {
	differences := []Difference{}
	if want.Error != nil {
		if !errors.Is(got.Err, want.Error) {
			differences = append(differences, Difference{Path: "result.error", Kind: "different-classification"})
		}
	} else if want.APIError == nil && got.Err != nil {
		differences = append(differences, Difference{Path: "result.error", Kind: "unexpected-error"})
	}
	if want.APIError != nil {
		var actual *client.APIError
		if !errors.As(got.Err, &actual) || !reflect.DeepEqual(want.APIError, actual) {
			differences = append(differences, Difference{Path: "result.api_error", Kind: "different-diagnostics"})
		}
	}
	status := 0
	if got.Response != nil {
		status = got.Response.StatusCode()
		differences = append(differences, compareHeaders(want.Headers, got.Response.Header(), "result.headers")...)
		if want.Body != nil && !slices.Equal(want.Body, got.Response.Body()) {
			differences = append(differences, Difference{Path: "result.body", Kind: "different-value"})
		}
	} else if want.Body != nil || len(want.Headers) != 0 {
		differences = append(differences, Difference{Path: "result.response", Kind: "missing-response"})
	}
	if status != want.Status {
		differences = append(differences, Difference{Path: "result.status", Kind: "different-value"})
	}
	if want.JSON != nil {
		raw, err := json.Marshal(got.Value)
		if err != nil {
			differences = append(differences, Difference{Path: "result.value", Kind: "encoding-failed"})
		} else {
			differences = append(differences, compareJSON(want.JSON, raw, "result.value")...)
		}
	}
	return differences
}
