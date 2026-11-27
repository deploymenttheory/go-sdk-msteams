package teamsexternalaccessconfiguration

import (
	"context"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// Service uses the direct compiled policy protocol, without generated discovery.
type Service struct {
	resource *policyrp.Resource[Configuration]
}

// NewService binds the direct compiled protocol to an injected Teams scope without I/O.
func NewService(scope *client.Scope) *Service {
	return &Service{resource: policyrp.New(scope, "TeamsExternalAccessConfiguration", func(v *Configuration) *string { return v.Identity })}
}
func (s *Service) ready(ctx context.Context) error {
	if s == nil {
		if ctx == nil {
			return client.ErrNilContext
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		return client.ErrClosed
	}
	return s.resource.Ready(ctx)
}

// GetByIdentity reads one explicit identity; it does not silently select Global.
func (s *Service) GetByIdentity(ctx context.Context, id string) (*Configuration, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.resource.Get(ctx, id, "Get-CsTeamsExternalAccessConfiguration")
}

// List reads one bare collection, with exact Global identities first. A successful
// empty/null response remains nil; no collection or continuation is fabricated.
func (s *Service) List(ctx context.Context) ([]*Configuration, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.resource.List(ctx, "Get-CsTeamsExternalAccessConfiguration")
}

// ListFiltered applies the pinned local Identity/Tag wildcard selection.
// An explicit empty Filter is invalid; use List to omit the filter.
func (s *Service) ListFiltered(ctx context.Context, pattern string) ([]*Configuration, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if pattern == "" {
		return nil, nil, ErrInput
	}
	return s.resource.ListFiltered(ctx, pattern, "Get-CsTeamsExternalAccessConfiguration")
}

// PatchByIdentity performs one native sparse PATCH, without a preceding read.
// At least one family setting must be bound; Force alone is not a setting.
func (s *Service) PatchByIdentity(ctx context.Context, id string, input PatchRequest) (*Configuration, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	body, err := policyrp.Fields(input.Settings)
	if err != nil {
		return nil, nil, err
	}
	if len(body) == 0 {
		return nil, nil, ErrInput
	}
	if err := policyrp.Identity(body, id, input.BodyIdentity, input.Controls); err != nil {
		return nil, nil, err
	}
	return s.resource.Write(ctx, id, "Set-CsTeamsExternalAccessConfiguration", http.MethodPatch, body)
}

// UpdateByIdentity mirrors the pinned Set composition: read current, apply only
// bound fields/list edits, then PATCH once. Concurrent changes may be overwritten.
func (s *Service) UpdateByIdentity(ctx context.Context, id string, input UpdateRequest) (*UpdateResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	fields, err := policyrp.Fields(input.Fields)
	if err != nil {
		return nil, nil, err
	}
	if len(fields) == 0 && input.BlockedUsers == nil {
		return nil, nil, ErrInput
	}
	if input.BlockedUsers != nil {
		if err := input.BlockedUsers.Validate(); err != nil {
			return nil, nil, err
		}
	}
	before, read, err := s.resource.Get(ctx, id, "Set-CsTeamsExternalAccessConfiguration")
	result := &UpdateResult{Before: before, ReadResponse: read}
	if err != nil {
		return result, read, err
	}
	if before == nil {
		return result, read, client.ErrDecode
	}
	patch := PatchRequest{Settings: Settings{Fields: input.Fields}, Controls: input.Controls, BodyIdentity: policy.Null[string]()}
	if before.Identity != nil {
		patch.BodyIdentity = policy.Set(*before.Identity)
	}
	if input.BlockedUsers != nil {
		patch.BlockedUsers, err = input.BlockedUsers.Apply(before.BlockedUsers)
		if err != nil {
			return result, read, err
		}
	}
	result.PatchInvoked = true
	after, response, err := s.PatchByIdentity(ctx, id, patch)
	result.After = after
	result.HTTPAccepted = response != nil && response.StatusCode() >= 200 && response.StatusCode() < 300
	return result, response, err
}
