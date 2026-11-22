package teamssharedcallingroutingpolicy

import (
	"context"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// Service binds shared calling routing policy operations to the direct compiled protocol.
type Service struct {
	resource *policyrp.Resource[Policy]
}

// NewService constructs a service without discovery, credentials or network I/O.
func NewService(scope *client.Scope) *Service {
	return &Service{resource: policyrp.New(scope, "TeamsSharedCallingRoutingPolicy", func(v *Policy) *string { return v.Identity })}
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
func (s *Service) GetByIdentity(ctx context.Context, id string) (*Policy, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.resource.Get(ctx, id, "Get-CsTeamsSharedCallingRoutingPolicy")
}

// List returns the bare collection with exact Global identities first.
func (s *Service) List(ctx context.Context) ([]*Policy, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.resource.List(ctx, "Get-CsTeamsSharedCallingRoutingPolicy")
}

// ListFiltered applies the pinned local Identity/Tag wildcard selection.
func (s *Service) ListFiltered(ctx context.Context, pattern string) ([]*Policy, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if pattern == "" {
		return nil, nil, ErrInput
	}
	return s.resource.ListFiltered(ctx, pattern, "Get-CsTeamsSharedCallingRoutingPolicy")
}

// PatchByIdentity sends only bound settings, without an implicit preparation read.
func (s *Service) PatchByIdentity(ctx context.Context, id string, input PatchRequest) (*Policy, *resty.Response, error) {
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
	return s.resource.Write(ctx, id, "Set-CsTeamsSharedCallingRoutingPolicy", http.MethodPatch, fields)
}

// UpdateByIdentity composes Set's read/modify/PATCH, retaining read preparation
// and HTTP acceptance independently from persisted/effective tenant state.
func (s *Service) UpdateByIdentity(ctx context.Context, id string, input UpdateRequest) (*UpdateResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if input.EmergencyNumbers != nil {
		if err := input.EmergencyNumbers.Validate(); err != nil {
			return nil, nil, err
		}
	}
	fields, err := policyrp.Fields(input.ScalarSettings)
	if err != nil {
		return nil, nil, err
	}
	if len(fields) == 0 && input.EmergencyNumbers == nil {
		return nil, nil, ErrInput
	}
	before, read, err := s.resource.Get(ctx, id, "Set-CsTeamsSharedCallingRoutingPolicy")
	result := &UpdateResult{Before: before, ReadResponse: read}
	if err != nil {
		return result, read, err
	}
	if before == nil {
		return result, read, client.ErrDecode
	}
	settings := Settings{ScalarSettings: input.ScalarSettings}
	if input.EmergencyNumbers != nil {
		settings.EmergencyNumbers, err = input.EmergencyNumbers.Apply(before.EmergencyNumbers)
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

// CreateByIdentity applies New's bound list edit to a fresh empty list, then POSTs once.
func (s *Service) CreateByIdentity(ctx context.Context, id string, input CreateRequest) (*Policy, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	settings := Settings{ScalarSettings: input.ScalarSettings}
	if input.EmergencyNumbers != nil {
		var e error
		settings.EmergencyNumbers, e = input.EmergencyNumbers.Apply(nil)
		if e != nil {
			return nil, nil, e
		}
	}
	body, e := policyrp.Fields(settings)
	if e != nil {
		return nil, nil, e
	}
	if e := policyrp.Identity(body, id, policy.Field[string]{}, input.Controls); e != nil {
		return nil, nil, e
	}
	return s.resource.Write(ctx, id, "New-CsTeamsSharedCallingRoutingPolicy", http.MethodPost, body)
}

// DeleteByIdentity sends the compiled Identity and optional Force body once.
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
	return s.resource.Write(ctx, id, "Remove-CsTeamsSharedCallingRoutingPolicy", http.MethodDelete, body)
}
