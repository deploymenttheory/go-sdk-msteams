package teamsmigrationconfiguration

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

type Service struct{ client client.Client }

// NewService constructs local discovery/routing dependencies without I/O.
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
	p := []string{"Skype.Policy", "configurations", "TeamsMigrationConfiguration"}
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

// GetByIdentity reads one explicit configuration, escaping the identity as one
// segment. The documented Global-only contract is not an invented local enum.
func (s *Service) GetByIdentity(ctx context.Context, identity string) (*Configuration, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !validIdentity(identity) {
		return nil, nil, ErrInput
	}
	var value Configuration
	response, err := s.do(ctx, "Get-CsTeamsMigrationConfiguration", client.Request{Method: http.MethodGet, Path: path(identity), Replay: client.ReplaySafe}, http.StatusOK, &value)
	if err != nil {
		return nil, response, err
	}
	if value.RawFields == nil {
		return nil, response, client.ErrDecode
	}
	return &value, response, nil
}

// List reads the bare array, retaining stable Global-first ordering. The pinned
// command has no pagination/nextLink contract or server Filter parameter.
func (s *Service) List(ctx context.Context) ([]Configuration, *resty.Response, error) {
	var raw json.RawMessage
	response, err := s.do(ctx, "Get-CsTeamsMigrationConfiguration", client.Request{Method: http.MethodGet, Path: path(""), Replay: client.ReplaySafe}, http.StatusOK, &raw)
	if err != nil {
		return nil, response, err
	}
	var values []Configuration
	if json.Unmarshal(raw, &values) != nil || values == nil {
		return nil, response, client.ErrDecode
	}
	slices.SortStableFunc(values, func(a, b Configuration) int {
		left, right := a.Identity != nil && strings.EqualFold(*a.Identity, "Global"), b.Identity != nil && strings.EqualFold(*b.Identity, "Global")
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

// UpdateByIdentity sends one sparse PUT including Identity, requiring HTTP 204.
// It never substitutes PATCH, replays a mutation or inserts setting defaults.
func (s *Service) UpdateByIdentity(ctx context.Context, identity string, input UpdateRequest) (*resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, err
	}
	if !validIdentity(identity) {
		return nil, ErrInput
	}
	body := map[string]any{"Identity": identity}
	if input.EnableLegacyClientInterop != nil {
		body["EnableLegacyClientInterop"] = *input.EnableLegacyClientInterop
	}
	return s.do(ctx, "Set-CsTeamsMigrationConfiguration", client.Request{Method: http.MethodPut, Path: path(identity), Body: body, Replay: client.ReplayNever}, http.StatusNoContent, nil)
}

// UpdateGlobal explicitly applies the pinned omitted-Identity default.
func (s *Service) UpdateGlobal(ctx context.Context, input UpdateRequest) (*resty.Response, error) {
	return s.UpdateByIdentity(ctx, "Global", input)
}
