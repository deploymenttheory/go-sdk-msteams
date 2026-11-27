package conversationrolessetting

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/go-resty/resty/v2"
)

const resourceName = "ConversationRolesSetting"
const getCommand = "Get-CsConversationRolesSetting"
const setCommand = "Set-CsConversationRolesSetting"

type Service struct {
	resource *policyrp.Resource[Configuration]
}

func NewService(scope *client.Scope) *Service {
	return &Service{resource: policyrp.New(scope, resourceName, func(v *Configuration) *string { return v.Identity })}
}
func (s *Service) ready(ctx context.Context) error {
	if s == nil {
		if ctx == nil {
			return client.ErrNilContext
		}
		if e := ctx.Err(); e != nil {
			return e
		}
		return client.ErrClosed
	}
	return s.resource.Ready(ctx)
}
func (s *Service) GetByIdentity(ctx context.Context, identity string) (*Configuration, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	return s.resource.Get(ctx, identity, getCommand)
}
func (s *Service) List(ctx context.Context) ([]*Configuration, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	return s.resource.List(ctx, getCommand)
}
func (s *Service) ListFiltered(ctx context.Context, pattern string) ([]*Configuration, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	if pattern == "" {
		return nil, nil, ErrInput
	}
	return s.resource.ListFiltered(ctx, pattern, getCommand)
}

// PrepareUpdateByIdentity reads only, applies the pinned local transition, and
// freezes its exact projected body for inspection. No PATCH occurs here.
func (s *Service) PrepareUpdateByIdentity(ctx context.Context, identity string, input UpdateParameters) (*UpdateRequest, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	if !policyrp.ValidIdentity(identity) {
		return nil, nil, ErrInput
	}
	request := &UpdateRequest{owner: s, identity: identity}
	if !input.ConversationRoles.IsSet() {
		return nil, nil, ErrInput
	}
	before, response, e := s.resource.Get(ctx, identity, setCommand)
	request.before, request.read = before, response
	if e != nil {
		return request, response, e
	}
	if before == nil {
		return request, response, client.ErrDecode
	}
	if e = ctx.Err(); e != nil {
		return request, response, e
	}
	request.projection, e = project(before, identity, input)
	request.prepared = e == nil
	return request, response, e
}

// UpdateByIdentity applies a frozen plan only to the exact preparing service and
// resource identity. Projection copies, inputs and exposed before snapshots do
// not permit cross-client transfer or retargeting of the prepared wire request.
func (s *Service) UpdateByIdentity(ctx context.Context, identity string, input UpdateRequest) (*UpdateResult, *resty.Response, error) {
	result := &UpdateResult{Before: input.Before(), ReadResponse: input.read, Projection: input.Projection()}
	if e := s.ready(ctx); e != nil {
		return result, nil, e
	}
	if !input.prepared || input.owner != s || input.identity != identity {
		return result, nil, ErrInput
	}
	var body map[string]json.RawMessage
	if json.Unmarshal(input.projection.Body, &body) != nil {
		return result, nil, ErrInput
	}
	result.PatchInvoked = true
	after, response, e := s.resource.Write(ctx, input.identity, setCommand, http.MethodPatch, body)
	result.After = after
	result.HTTPAccepted = response != nil && response.StatusCode() >= 200 && response.StatusCode() < 300
	return result, response, e
}
