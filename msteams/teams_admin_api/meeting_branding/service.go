package meetingbranding

import (
	"context"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/go-resty/resty/v2"
)

// Service uses the direct compiled policy protocol, without generated discovery.
type Service struct{ resource *policyrp.Resource[Policy] }

// NewService binds the direct compiled protocol to an injected Teams scope without I/O.
func NewService(scope *client.Scope) *Service {
	return &Service{resource: policyrp.New(scope, "TeamsMeetingBrandingPolicy", func(v *Policy) *string { return v.Identity })}
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
func (s *Service) GetByIdentity(ctx context.Context, id string) (*Policy, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.resource.Get(ctx, id, "Get-CsTeamsMeetingBrandingPolicy")
}

// List reads one bare collection, with exact Global identities first. A successful
// empty/null response remains nil; no collection or continuation is fabricated.
func (s *Service) List(ctx context.Context) ([]*Policy, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.resource.List(ctx, "Get-CsTeamsMeetingBrandingPolicy")
}

// ListFiltered applies the pinned local Identity/Tag wildcard selection.
func (s *Service) ListFiltered(ctx context.Context, pattern string) ([]*Policy, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.resource.ListFiltered(ctx, pattern, "Get-CsTeamsMeetingBrandingPolicy")
}

// PatchByIdentity performs one native sparse PATCH, without a preceding read.
// At least one family setting must be bound; Force alone is not a setting.
func (s *Service) PatchByIdentity(ctx context.Context, id string, input PatchRequest) (*Policy, *resty.Response, error) {
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
	return s.resource.Write(ctx, id, "Set-CsTeamsMeetingBrandingPolicy", http.MethodPatch, body)
}
