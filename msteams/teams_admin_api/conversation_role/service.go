package conversationrole

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

const resourceName = "ConversationRole"
const getCommand = "Get-CsConversationRole"
const setCommand = "Set-CsConversationRole"
const deleteCommand = "Remove-CsConversationRole"

// Service uses the conditional compiled provider. It does not install exports,
// authenticate in constructors, or claim full permission-write qualification.
type Service struct{ resource *policyrp.Resource[Role] }

func NewService(scope *client.Scope) *Service {
	return &Service{resource: policyrp.New(scope, resourceName, func(v *Role) *string { return v.Identity })}
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

// GetByIdentity reads one explicitly identified role.
func (s *Service) GetByIdentity(ctx context.Context, identity string) (*Role, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	return s.resource.Get(ctx, identity, getCommand)
}

// GetByID composes the source Global/relative-ID selector without conflating it
// with the role's separate Id response property.
func (s *Service) GetByID(ctx context.Context, id string) (*Role, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	if !policyrp.ValidIdentity(id) {
		return nil, nil, ErrInput
	}
	return s.resource.Get(ctx, "Global/"+id, getCommand)
}
func (s *Service) List(ctx context.Context) ([]*Role, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	return s.resource.List(ctx, getCommand)
}

// ListFiltered preserves literal whitespace and the explicit nonempty Filter
// declaration; omission uses List.
func (s *Service) ListFiltered(ctx context.Context, pattern string) ([]*Role, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	if pattern == "" {
		return nil, nil, ErrInput
	}
	return s.resource.ListFiltered(ctx, pattern, getCommand)
}

// Create applies a previously inspected source projection. The request accepts
// upstream DTO flags, but its Projection reports what the pinned wire omits.
func (s *Service) Create(ctx context.Context, input CreateRequest) (*Role, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	if input.identity == "" || input.projection.Command != "New-CsConversationRole" || len(input.projection.Body) == 0 {
		return nil, nil, ErrInput
	}
	var body map[string]json.RawMessage
	if json.Unmarshal(input.projection.Body, &body) != nil {
		return nil, nil, ErrInput
	}
	return s.resource.Write(ctx, input.identity, input.projection.Command, http.MethodPost, body)
}

// PrepareUpdateByIdentity performs only the preparation read and local source
// projection. Inspect its owned Projection before calling UpdateByIdentity.
func (s *Service) PrepareUpdateByIdentity(ctx context.Context, identity string, input UpdateParameters) (*UpdateRequest, *resty.Response, error) {
	return s.prepare(ctx, identity, identity, IdentitySet, input)
}

// PrepareUpdateByID retains the separate relative selector/model-Id behavior.
func (s *Service) PrepareUpdateByID(ctx context.Context, id string, input UpdateParameters) (*UpdateRequest, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	if !policyrp.ValidIdentity(id) {
		return nil, nil, ErrInput
	}
	return s.prepare(ctx, "Global/"+id, id, RelativeIDSet, input)
}
func (s *Service) prepare(ctx context.Context, identity, selector, set string, input UpdateParameters) (*UpdateRequest, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	if !policyrp.ValidIdentity(identity) {
		return nil, nil, ErrInput
	}
	request := &UpdateRequest{owner: s, identity: identity, selector: selector, set: set}
	if !input.specified() {
		request.prepared = true
		request.projection = WriteProjection{Command: setCommand, ParameterSet: set, NoOp: true}
		return request, nil, nil
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
	request.projection, e = projectUpdate(before, identity, set, input)
	request.prepared = e == nil
	return request, response, e
}

// UpdateByIdentity applies only the frozen PATCH prepared by this service for
// exactly this selector. It never repeats the preparation or silently reprojects.
func (s *Service) UpdateByIdentity(ctx context.Context, identity string, input UpdateRequest) (*UpdateResult, *resty.Response, error) {
	return s.update(ctx, identity, IdentitySet, input)
}

// UpdateByID applies only a matching relative-ID prepared request.
func (s *Service) UpdateByID(ctx context.Context, id string, input UpdateRequest) (*UpdateResult, *resty.Response, error) {
	return s.update(ctx, id, RelativeIDSet, input)
}
func (s *Service) update(ctx context.Context, selector, set string, input UpdateRequest) (*UpdateResult, *resty.Response, error) {
	result := &UpdateResult{Before: input.Before(), ReadResponse: input.read, Projection: input.Projection(), NoOp: input.projection.NoOp}
	if e := s.ready(ctx); e != nil {
		return result, nil, e
	}
	if !input.prepared || input.owner != s || input.selector != selector || input.set != set {
		return result, nil, ErrInput
	}
	if input.projection.NoOp {
		return result, nil, nil
	}
	var body map[string]json.RawMessage
	if json.Unmarshal(input.projection.Body, &body) != nil {
		return result, nil, ErrInput
	}
	result.PatchInvoked = true
	value, response, e := s.resource.Write(ctx, input.identity, setCommand, http.MethodPatch, body)
	result.After = value
	result.HTTPAccepted = response != nil && response.StatusCode() >= 200 && response.StatusCode() < 300
	return result, response, e
}

// DeleteByIdentity removes one exact role using the compiled scope metadata.
func (s *Service) DeleteByIdentity(ctx context.Context, identity string, controls policy.Controls) (*Role, *resty.Response, error) {
	return s.delete(ctx, identity, IdentitySet, "", controls)
}

// DeleteByID preserves the relative selector as a separate bound Id parameter.
func (s *Service) DeleteByID(ctx context.Context, id string, controls policy.Controls) (*Role, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	if !policyrp.ValidIdentity(id) {
		return nil, nil, ErrInput
	}
	return s.delete(ctx, "Global/"+id, RelativeIDSet, id, controls)
}
func (s *Service) delete(ctx context.Context, identity, set, id string, controls policy.Controls) (*Role, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	body, e := deleteBody(identity, set, id, controls)
	if e != nil {
		return nil, nil, e
	}
	return s.resource.Write(ctx, identity, deleteCommand, http.MethodDelete, body)
}
