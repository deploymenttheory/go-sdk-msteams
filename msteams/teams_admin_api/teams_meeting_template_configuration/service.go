package teamsmeetingtemplateconfiguration

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/go-resty/resty/v2"
)

const resourceName = "TeamsMeetingTemplateConfiguration"
const getCommand = "Get-CsTeamsMeetingTemplateConfiguration"

// Service uses the compiled policy provider with an injected Teams scope.
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

// GetByIdentity reads one compiled policy with an explicit full identity.
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
