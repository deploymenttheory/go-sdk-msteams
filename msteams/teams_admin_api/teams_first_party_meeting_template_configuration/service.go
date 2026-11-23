package teamsfirstpartymeetingtemplateconfiguration

import (
	"context"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

const resourceName = "TeamsFirstPartyMeetingTemplateConfiguration"
const getCommand = "Get-CsTeamsFirstPartyMeetingTemplateConfiguration"
const setCommand = "Set-CsTeamsFirstPartyMeetingTemplateConfiguration"

// Service binds first-party meeting-template configuration operations to the
// direct compiled PolicyRP protocol with an injected Teams scope.
type Service struct {
	resource *policyrp.Resource[Configuration]
}

// NewService constructs a service without discovery, credentials or network I/O.
func NewService(scope *client.Scope) *Service {
	return &Service{resource: policyrp.New(scope, resourceName, func(v *Configuration) *string { return v.Identity })}
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
	return s.resource.Get(ctx, id, getCommand)
}

// List returns the bare collection with exact Global identities first.
func (s *Service) List(ctx context.Context) ([]*Configuration, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.resource.List(ctx, getCommand)
}

// ListFiltered applies the pinned local Identity/Tag wildcard selection.
func (s *Service) ListFiltered(ctx context.Context, pattern string) ([]*Configuration, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if pattern == "" {
		return nil, nil, ErrInput
	}
	return s.resource.ListFiltered(ctx, pattern, getCommand)
}

// PatchByIdentity sends only bound settings, without an implicit preparation read.
// Force alone is not a setting.
func (s *Service) PatchByIdentity(ctx context.Context, id string, input PatchRequest) (*Configuration, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	body, err := fields(input.Settings)
	if err != nil {
		return nil, nil, err
	}
	if len(body) == 0 {
		return nil, nil, ErrInput
	}
	if err := policyrp.Identity(body, id, input.BodyIdentity, input.Controls); err != nil {
		return nil, nil, err
	}
	return s.resource.Write(ctx, id, setCommand, http.MethodPatch, body)
}

// UpdateByIdentity composes Set's read/modify/PATCH. Template edits apply locally
// to the preparation read and are sent as a complete replacement array; binding
// the template list also sends the read Description, as the pinned serializer
// does. Preparation and HTTP acceptance are retained separately.
func (s *Service) UpdateByIdentity(ctx context.Context, id string, input UpdateRequest) (*UpdateResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if err := validateEdits(input); err != nil {
		return nil, nil, err
	}
	scalars, err := policyrp.Fields(input.ScalarSettings)
	if err != nil {
		return nil, nil, err
	}
	if len(scalars) == 0 && input.TeamsMeetingTemplates == nil {
		return nil, nil, ErrInput
	}
	before, read, err := s.resource.Get(ctx, id, setCommand)
	result := &UpdateResult{Before: before, ReadResponse: read}
	if err != nil {
		return result, read, err
	}
	if before == nil {
		return result, read, client.ErrDecode
	}
	settings, err := resolve(input, before)
	if err != nil {
		return result, read, err
	}
	identity := policy.Null[string]()
	if before.Identity != nil {
		identity = policy.Set(*before.Identity)
	}
	result.PatchInvoked = true
	result.After, read, err = s.PatchByIdentity(ctx, id, PatchRequest{Settings: settings, Controls: input.Controls, BodyIdentity: identity})
	result.HTTPAccepted = read != nil && read.StatusCode() >= 200 && read.StatusCode() < 300
	return result, read, err
}
