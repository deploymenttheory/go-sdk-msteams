package audioconferencingprompts

import (
	"context"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

const getCommand = "Get-CsTeamsAudioConferencingCustomPromptsConfiguration"
const setCommand = "Set-CsTeamsAudioConferencingCustomPromptsConfiguration"
const resourceName = "TeamsAudioConferencingCustomPromptsConfiguration"

// Service uses the pinned direct compiled policy protocol and an injected scope.
type Service struct {
	resource *policyrp.Resource[Configuration]
}

// NewService constructs the resource without credentials, discovery or HTTP.
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

// GetByIdentity reads one full configuration identity.
func (s *Service) GetByIdentity(ctx context.Context, id string) (*Configuration, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	return s.resource.Get(ctx, id, getCommand)
}

// List returns the bare collection with exact Global identities first.
func (s *Service) List(ctx context.Context) ([]*Configuration, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	return s.resource.List(ctx, getCommand)
}

// ListFiltered applies local Identity/Tag wildcard selection after a collection read.
func (s *Service) ListFiltered(ctx context.Context, pattern string) ([]*Configuration, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	if pattern == "" {
		return nil, nil, ErrInput
	}
	return s.resource.ListFiltered(ctx, pattern, getCommand)
}

// PatchByIdentity sends complete bound arrays in one sparse native PATCH.
// Force alone is not a setting; mutation requests are never replayed.
func (s *Service) PatchByIdentity(ctx context.Context, id string, input PatchRequest) (*Configuration, *resty.Response, error) {
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

// UpdateByIdentity explicitly reads with Set metadata, edits owned copies, then
// PATCHes bound arrays once. Read preparation and HTTP acceptance remain visible
// on failure. A package removal can fail locally after the read, before PATCH.
func (s *Service) UpdateByIdentity(ctx context.Context, id string, input UpdateRequest) (*UpdateResult, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	if e := validateEdits(input); e != nil {
		return nil, nil, e
	}
	if input.Prompts == nil && input.Packages == nil {
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
