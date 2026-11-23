package teamsmedialoggingpolicy

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyfilter"
	"github.com/go-resty/resty/v2"
)

const resourceName = "TeamsMediaLoggingPolicy"
const getCommand = "Get-CsTeamsMediaLoggingPolicy"
const grantCommand = "Grant-CsTeamsMediaLoggingPolicy"

// Service uses the injected Teams ConfigAPI client and its forest discovery.
type Service struct{ client client.Client }

// NewService performs no I/O. The backend must use the Teams ConfigAPI audience.
// The root client supplies its shared, identity-bound ConfigAPI backend.
func NewService(base client.Client) *Service { return &Service{client: configapi.New(base)} }
func (s *Service) ready(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if e := ctx.Err(); e != nil {
		return e
	}
	if s == nil || s.client == nil {
		return client.ErrClosed
	}
	return nil
}
func validIdentity(v string) bool {
	return strings.TrimSpace(v) != "" && v != "." && v != ".." && !strings.ContainsRune(v, '\\') && !strings.ContainsFunc(v, unicode.IsControl)
}
func path(id string) []string {
	p := []string{"Skype.Policy", "configurations", resourceName}
	if id != "" {
		p = append(p, "configuration", id)
	}
	return p
}
func (s *Service) do(ctx context.Context, command string, request client.Request, result any, statuses ...int) (*resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, e
	}
	request.Header = http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}
	r, e := s.client.Do(ctx, request, result)
	if e != nil {
		return r, e
	}
	for _, status := range statuses {
		if r != nil && r.StatusCode() == status {
			return r, nil
		}
	}
	return r, ErrResponse
}

// GetByIdentity reads one explicit identity, escaped as one whole route segment.
func (s *Service) GetByIdentity(ctx context.Context, id string) (*Policy, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	if !validIdentity(id) {
		return nil, nil, ErrInput
	}
	var v Policy
	r, e := s.do(ctx, getCommand, client.Request{Method: http.MethodGet, Path: path(id), Replay: client.ReplaySafe}, &v, http.StatusOK)
	if e != nil {
		return nil, r, e
	}
	if v.RawFields == nil {
		return nil, r, client.ErrDecode
	}
	return &v, r, nil
}

// List reads a bare collection with a stable, case-insensitive Global-first partition.
func (s *Service) List(ctx context.Context) ([]Policy, *resty.Response, error) {
	var raw json.RawMessage
	r, e := s.do(ctx, getCommand, client.Request{Method: http.MethodGet, Path: path(""), Replay: client.ReplaySafe}, &raw, http.StatusOK)
	if e != nil {
		return nil, r, e
	}
	var values []Policy
	if json.Unmarshal(raw, &values) != nil || values == nil {
		return nil, r, client.ErrDecode
	}
	slices.SortStableFunc(values, func(a, b Policy) int {
		left, right := strings.EqualFold(a.Identity, "Global"), strings.EqualFold(b.Identity, "Global")
		if left == right {
			return 0
		}
		if left {
			return -1
		}
		return 1
	})
	return values, r, nil
}

// ListFiltered applies local Identity/Tag wildcard selection. Whitespace selects
// the complete collection as in the mapped modern helper.
func (s *Service) ListFiltered(ctx context.Context, pattern string) ([]Policy, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	if strings.TrimSpace(pattern) == "" {
		return s.List(ctx)
	}
	plain, e := policyfilter.Compile(pattern)
	if e != nil {
		return nil, nil, ErrInput
	}
	tagged, e := policyfilter.Compile("Tag:" + pattern)
	if e != nil {
		return nil, nil, ErrInput
	}
	values, r, e := s.List(ctx)
	if e != nil {
		return nil, r, e
	}
	out := make([]Policy, 0)
	for _, v := range values {
		match, e := plain.Match(ctx, v.Identity)
		if e != nil {
			return nil, r, e
		}
		if !match {
			match, e = tagged.Match(ctx, v.Identity)
			if e != nil {
				return nil, r, e
			}
		}
		if match {
			out = append(out, v)
		}
	}
	return out, r, nil
}
