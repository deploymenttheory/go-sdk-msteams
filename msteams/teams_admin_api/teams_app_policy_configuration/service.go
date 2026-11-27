package teamsapppolicyconfiguration

import (
	"context"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

const resourceName = "TeamsAppPolicyConfiguration"
const getCommand = "Get-CsTeamsAppPolicyConfiguration"
const setCommand = "Set-CsTeamsAppPolicyConfiguration"

// Service uses the conditional compiled configuration provider with an injected Teams scope.
type Service struct {
	resource *policyrp.Resource[Configuration]
}

// NewService constructs the resource without authentication, discovery or HTTP.
func NewService(scope *client.Scope) *Service {
	return &Service{resource: policyrp.New(scope, resourceName, func(p *Configuration) *string { return p.Identity })}
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

// GetByIdentity reads one conditional compiled configuration with an explicit full identity.
func (s *Service) GetByIdentity(ctx context.Context, id string) (*Configuration, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	return s.resource.Get(ctx, id, getCommand)
}

// List reads a bare collection and places exact Global identities first.
func (s *Service) List(ctx context.Context) ([]*Configuration, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	return s.resource.List(ctx, getCommand)
}

// ListFiltered applies compiled Identity/Tag wildcards; whitespace stays literal.
func (s *Service) ListFiltered(ctx context.Context, pattern string) ([]*Configuration, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	if pattern == "" {
		return nil, nil, ErrInput
	}
	return s.resource.ListFiltered(ctx, pattern, getCommand)
}

// PatchByIdentity sends one sparse compiled PATCH. Force alone is not a setting.
func (s *Service) PatchByIdentity(ctx context.Context, id string, input PatchRequest) (*Configuration, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	if e := validate(input.Settings); e != nil {
		return nil, nil, e
	}
	body, e := policyrp.Fields(input.Settings)
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

// UpdateByIdentity explicitly composes Set-metadata Get and sparse PATCH,
// retaining preparation and HTTP acceptance independently of decoding failures.
func (s *Service) UpdateByIdentity(ctx context.Context, id string, input UpdateRequest) (*UpdateResult, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	if e := validate(input.Settings); e != nil {
		return nil, nil, e
	}
	body, e := policyrp.Fields(input.Settings)
	if e != nil {
		return nil, nil, e
	}
	if len(body) == 0 {
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
	identity := policy.Null[string]()
	if before.Identity != nil {
		identity = policy.Set(*before.Identity)
	}
	result.PatchInvoked = true
	result.After, read, e = s.PatchByIdentity(ctx, id, PatchRequest{Settings: input.Settings, Controls: input.Controls, BodyIdentity: identity})
	result.HTTPAccepted = read != nil && read.StatusCode() >= 200 && read.StatusCode() < 300
	return result, read, e
}
