package targetingpolicy

import (
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/go-resty/resty/v2"
)

// Service binds the evidenced modern ConfigAPI policy resource to an injected
// backend. It never starts a PowerShell/TRPS fallback or assumes flight eligibility.
type Service struct{ client client.Client }

// NewService constructs local routing/discovery dependencies without I/O.
func NewService(base client.Client) *Service { return &Service{client: configapi.New(base)} }

func (s *Service) ready(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil || s.client == nil {
		return client.ErrClosed
	}
	return nil
}

func validIdentity(identity string) bool {
	return strings.TrimSpace(identity) != "" && identity != "." && identity != ".." && !strings.ContainsRune(identity, '\\') && !strings.ContainsFunc(identity, unicode.IsControl)
}

func path(identity string) []string {
	p := []string{"Skype.Policy", "configurations", "TeamsTargetingPolicy"}
	if identity != "" {
		p = append(p, "configuration", identity)
	}
	return p
}

func (s *Service) do(ctx context.Context, command string, request client.Request, status int, result any) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	request.Header = http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}
	response, err := s.client.Do(ctx, request, result)
	if err == nil && (response == nil || response.StatusCode() != status) {
		err = ErrResponse
	}
	return response, err
}

// GetByIdentity reads one explicit configuration identity, escaped as one segment.
func (s *Service) GetByIdentity(ctx context.Context, identity string) (*Policy, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !validIdentity(identity) {
		return nil, nil, ErrInput
	}
	var value Policy
	response, err := s.do(ctx, "Get-CsTeamsTargetingPolicy", client.Request{Method: http.MethodGet, Path: path(identity), Replay: client.ReplaySafe}, http.StatusOK, &value)
	if err != nil {
		return nil, response, err
	}
	if value.RawFields == nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

// List reads the bare array and stably places Global first. No unevidenced
// pagination or server Filter query is added.
func (s *Service) List(ctx context.Context) ([]Policy, *resty.Response, error) {
	var raw json.RawMessage
	response, err := s.do(ctx, "Get-CsTeamsTargetingPolicy", client.Request{Method: http.MethodGet, Path: path(""), Replay: client.ReplaySafe}, http.StatusOK, &raw)
	if err != nil {
		return nil, response, err
	}
	var values []Policy
	if json.Unmarshal(raw, &values) != nil || values == nil {
		return nil, response, client.ErrDecode
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
	return values, response, nil
}

// UpdateByIdentity sends the evidenced sparse PUT and requires HTTP 204. It
// never substitutes PATCH, retries a mutation or fabricates default properties.
func (s *Service) UpdateByIdentity(ctx context.Context, identity string, input UpdateRequest) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if !validIdentity(identity) {
		return nil, ErrInput
	}
	body := struct {
		Identity string `json:"Identity"`
		UpdateRequest
	}{Identity: identity, UpdateRequest: input}
	return s.do(ctx, "Set-CsTeamsTargetingPolicy", client.Request{Method: http.MethodPut, Path: path(identity), Body: body}, http.StatusNoContent, nil)
}

// UpdateGlobal names the pinned omitted-Identity default explicitly. Native
// UpdateByIdentity requires an identity instead of silently selecting Global.
func (s *Service) UpdateGlobal(ctx context.Context, input UpdateRequest) (*resty.Response, error) {
	return s.UpdateByIdentity(ctx, "Global", input)
}

// DeleteByIdentity sends one native Delete. The documented Global reset and
// assignment fallback are service behavior; no guessed reset body is constructed.
func (s *Service) DeleteByIdentity(ctx context.Context, identity string) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if !validIdentity(identity) {
		return nil, ErrInput
	}
	return s.do(ctx, "Remove-CsTeamsTargetingPolicy", client.Request{Method: http.MethodDelete, Path: path(identity)}, http.StatusNoContent, nil)
}
