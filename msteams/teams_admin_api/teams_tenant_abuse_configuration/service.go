package teamstenantabuseconfiguration

import (
	"context"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// Service binds TeamsTenantAbuseConfiguration operations to the direct compiled protocol.
type Service struct {
	resource *policyrp.Resource[Configuration]
}

// NewService constructs a service without discovery, credentials or network I/O.
func NewService(scope *client.Scope) *Service {
	return &Service{resource: policyrp.New(scope, "TeamsTenantAbuseConfiguration", func(v *Configuration) *string { return v.Identity })}
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

// GetByIdentity reads the supplied identity, including an explicitly selected Global.
func (s *Service) GetByIdentity(ctx context.Context, id string) (*Configuration, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.resource.Get(ctx, id, "Get-CsTeamsTenantAbuseConfiguration")
}

// List returns the bare collection with exact Global identities first.
func (s *Service) List(ctx context.Context) ([]*Configuration, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.resource.List(ctx, "Get-CsTeamsTenantAbuseConfiguration")
}

// ListFiltered applies the pinned local Identity/Tag wildcard selection.
func (s *Service) ListFiltered(ctx context.Context, pattern string) ([]*Configuration, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if pattern == "" {
		return nil, nil, ErrInput
	}
	return s.resource.ListFiltered(ctx, pattern, "Get-CsTeamsTenantAbuseConfiguration")
}

// PatchByIdentity sends only bound settings, without an implicit preparation read.
func (s *Service) PatchByIdentity(ctx context.Context, id string, input PatchRequest) (*Configuration, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	fields, err := policyrp.Fields(input.Settings)
	if err != nil {
		return nil, nil, err
	}
	if len(fields) == 0 {
		return nil, nil, ErrInput
	}
	if err := policyrp.Identity(fields, id, input.BodyIdentity, input.Controls); err != nil {
		return nil, nil, err
	}
	return s.resource.Write(ctx, id, "Set-CsTeamsTenantAbuseConfiguration", http.MethodPatch, fields)
}

// UpdateByIdentity composes Set's read/modify/PATCH, retaining read preparation
// and HTTP acceptance independently from persisted/effective tenant state.
func (s *Service) UpdateByIdentity(ctx context.Context, id string, input UpdateRequest) (*UpdateResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if input.BlockedTenants != nil {
		if err := input.BlockedTenants.Validate(); err != nil {
			return nil, nil, err
		}
	}
	fields, err := policyrp.Fields(input.ScalarSettings)
	if err != nil {
		return nil, nil, err
	}
	if len(fields) == 0 && input.BlockedTenants == nil {
		return nil, nil, ErrInput
	}
	before, read, err := s.resource.Get(ctx, id, "Set-CsTeamsTenantAbuseConfiguration")
	result := &UpdateResult{Before: before, ReadResponse: read}
	if err != nil {
		return result, read, err
	}
	if before == nil {
		return result, read, client.ErrDecode
	}
	settings := Settings{ScalarSettings: input.ScalarSettings}
	if input.BlockedTenants != nil {
		settings.BlockedTenants, err = input.BlockedTenants.ApplyNullable(before.BlockedTenants)
		if err != nil {
			return result, read, err
		}
	}
	patch := PatchRequest{Settings: settings, Controls: input.Controls, BodyIdentity: policy.Null[string]()}
	if before.Identity != nil {
		patch.BodyIdentity = policy.Set(*before.Identity)
	}
	result.PatchInvoked = true
	result.After, read, err = s.PatchByIdentity(ctx, id, patch)
	result.HTTPAccepted = read != nil && read.StatusCode() >= 200 && read.StatusCode() < 300
	return result, read, err
}
