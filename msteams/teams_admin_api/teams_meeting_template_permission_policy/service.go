package teamsmeetingtemplatepermissionpolicy

import (
	"context"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

const getCommand = "Get-CsTeamsMeetingTemplatePermissionPolicy"
const newCommand = "New-CsTeamsMeetingTemplatePermissionPolicy"
const setCommand = "Set-CsTeamsMeetingTemplatePermissionPolicy"
const removeCommand = "Remove-CsTeamsMeetingTemplatePermissionPolicy"
const grantCommand = "Grant-CsTeamsMeetingTemplatePermissionPolicy"
const resourceName = "TeamsMeetingTemplatePermissionPolicy"

// Service uses the direct compiled policy protocol and an injected Teams scope.
type Service struct{ resource *policyrp.Resource[Policy] }

// NewService constructs the resource without credentials, discovery or HTTP.
func NewService(scope *client.Scope) *Service {
	return &Service{resource: policyrp.New(scope, resourceName, func(v *Policy) *string { return v.Identity })}
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

// GetByIdentity reads an explicit full configuration identity.
func (s *Service) GetByIdentity(ctx context.Context, id string) (*Policy, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	return s.resource.Get(ctx, id, getCommand)
}

// List returns a bare collection, with exact Global identities first.
func (s *Service) List(ctx context.Context) ([]*Policy, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	return s.resource.List(ctx, getCommand)
}

// ListFiltered applies the pinned local Identity/Tag wildcard selection.
func (s *Service) ListFiltered(ctx context.Context, pattern string) ([]*Policy, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	if pattern == "" {
		return nil, nil, ErrInput
	}
	return s.resource.ListFiltered(ctx, pattern, getCommand)
}

// CreateByIdentity applies edits to fresh empty lists and POSTs without an
// implicit Global comparison read, which is source ShouldProcess presentation.
func (s *Service) CreateByIdentity(ctx context.Context, id string, input CreateRequest) (*Policy, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	settings, e := resolve(UpdateRequest(input), nil)
	if e != nil {
		return nil, nil, e
	}
	body, e := fields(settings)
	if e != nil {
		return nil, nil, e
	}
	if e := policyrp.Identity(body, id, policy.Field[string]{}, input.Controls); e != nil {
		return nil, nil, e
	}
	return s.resource.Write(ctx, id, newCommand, http.MethodPost, body)
}

// PatchByIdentity sends one sparse native PATCH. Force alone is not a setting.
func (s *Service) PatchByIdentity(ctx context.Context, id string, input PatchRequest) (*Policy, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	body, e := fields(input.Settings)
	if e != nil {
		return nil, nil, e
	}
	if len(body) == 0 {
		return nil, nil, ErrInput
	}
	if e := policyrp.Identity(body, id, input.BodyIdentity, input.Controls); e != nil {
		return nil, nil, e
	}
	return s.resource.Write(ctx, id, setCommand, http.MethodPatch, body)
}

// UpdateByIdentity explicitly reads, copies and edits current lists, then PATCHes
// bound fields once. Preparation and HTTP acceptance survive partial failures.
func (s *Service) UpdateByIdentity(ctx context.Context, id string, input UpdateRequest) (*UpdateResult, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	if e := validateEdits(input); e != nil {
		return nil, nil, e
	}
	scalars, e := policyrp.Fields(input.ScalarSettings)
	if e != nil {
		return nil, nil, e
	}
	if len(scalars) == 0 && input.HiddenMeetingTemplates == nil {
		return nil, nil, ErrInput
	}
	before, read, e := s.resource.Get(ctx, id, setCommand)
	result := &UpdateResult{Before: before, ReadResponse: read}
	if e != nil {
		return result, read, e
	}
	if before == nil {
		return result, read, client.ErrDecode
	}
	settings, e := resolve(input, before)
	if e != nil {
		return result, read, e
	}
	identity := policy.Null[string]()
	if before.Identity != nil {
		identity = policy.Set(*before.Identity)
	}
	result.PatchInvoked = true
	result.After, read, e = s.PatchByIdentity(ctx, id, PatchRequest{Settings: settings, Controls: input.Controls, BodyIdentity: identity})
	result.HTTPAccepted = read != nil && read.StatusCode() >= 200 && read.StatusCode() < 300
	return result, read, e
}

// DeleteByIdentity sends the bound Identity and optional Force body once.
func (s *Service) DeleteByIdentity(ctx context.Context, id string, controls policy.Controls) (*Policy, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	body, e := policyrp.Fields(struct{}{})
	if e != nil {
		return nil, nil, e
	}
	if e := policyrp.Identity(body, id, policy.Field[string]{}, controls); e != nil {
		return nil, nil, e
	}
	return s.resource.Write(ctx, id, removeCommand, http.MethodDelete, body)
}
