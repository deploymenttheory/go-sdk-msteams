package recordingmessages

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

const getCommand = "Get-CsTeamsRecordingAndTranscriptionCustomMessage"
const newCommand = "New-CsTeamsRecordingAndTranscriptionCustomMessage"
const setCommand = "Set-CsTeamsRecordingAndTranscriptionCustomMessage"
const removeCommand = "Remove-CsTeamsRecordingAndTranscriptionCustomMessage"

// Service manages the direct compiled recording-message resource.
type Service struct{ resource *policyrp.Resource[Message] }

// NewService constructs a service without credentials, discovery or network I/O.
func NewService(scope *client.Scope) *Service {
	return &Service{resource: policyrp.New(scope, "TeamsRecordingAndTranscriptionCustomMessage", func(v *Message) *string { return v.Identity })}
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
func relativeIdentity(id string) (string, error) {
	if id == "" || !policyrp.ValidIdentity("Global/"+id) {
		return "", ErrInput
	}
	return "Global/" + id, nil
}

// GetByIdentity reads an explicit full identity without normalizing it to an Id.
func (s *Service) GetByIdentity(ctx context.Context, identity string) (*Message, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.resource.Get(ctx, identity, getCommand)
}

// GetByID reads the source's Global/{Id} route.
func (s *Service) GetByID(ctx context.Context, id string) (*Message, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	identity, err := relativeIdentity(id)
	if err != nil {
		return nil, nil, err
	}
	return s.resource.Get(ctx, identity, getCommand)
}

// List returns the bare collection, with exact Global identities first.
func (s *Service) List(ctx context.Context) ([]*Message, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.resource.List(ctx, getCommand)
}

// ListFiltered applies local Identity/Tag wildcard matching; empty is invalid.
func (s *Service) ListFiltered(ctx context.Context, pattern string) ([]*Message, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if pattern == "" {
		return nil, nil, ErrInput
	}
	return s.resource.ListFiltered(ctx, pattern, getCommand)
}

// CreateByIdentity preserves the source's extracted Id metadata in the POST body.
func (s *Service) CreateByIdentity(ctx context.Context, identity string, input CreateRequest) (*Message, *resty.Response, error) {
	return s.create(ctx, identity, "identity", input)
}

// CreateByID binds Id and preserves the new DTO's null Id body value.
func (s *Service) CreateByID(ctx context.Context, id string, input CreateRequest) (*Message, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	identity, err := relativeIdentity(id)
	if err != nil {
		return nil, nil, err
	}
	return s.create(ctx, identity, "id", input)
}

// Create generates one UUID, retains it on failure, and submits without binding
// the Id parameter. That body differs from CreateByID's explicitly bound null Id.
func (s *Service) Create(ctx context.Context, input CreateRequest) (*CreateResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if _, err := resolve(UpdateRequest(input), nil); err != nil {
		return nil, nil, err
	}
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return nil, nil, err
	}
	raw[6] = raw[6]&0x0f | 0x40
	raw[8] = raw[8]&0x3f | 0x80
	id := fmt.Sprintf("%x-%x-%x-%x-%x", raw[:4], raw[4:6], raw[6:8], raw[8:10], raw[10:])
	result := &CreateResult{ID: id, Identity: "Global/" + id}
	v, r, err := s.create(ctx, result.Identity, "auto", input)
	result.Value = v
	result.HTTPAccepted = r != nil && r.StatusCode() >= 200 && r.StatusCode() < 300
	return result, r, err
}
func (s *Service) create(ctx context.Context, identity, selector string, input CreateRequest) (*Message, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	settings, err := resolve(UpdateRequest(input), nil)
	if err != nil {
		return nil, nil, err
	}
	body, err := fields(settings)
	if err != nil {
		return nil, nil, err
	}
	if err := put(body, "Identity", identity); err != nil {
		return nil, nil, err
	}
	if selector == "id" {
		if err := put(body, "Id", nil); err != nil {
			return nil, nil, err
		}
	} else if selector == "identity" && strings.Contains(identity, "/") {
		if err := put(body, "Id", strings.Split(identity, "/")[1]); err != nil {
			return nil, nil, err
		}
	}
	if err := force(body, input.Controls); err != nil {
		return nil, nil, err
	}
	return s.resource.Write(ctx, identity, newCommand, http.MethodPost, body)
}

// PatchByIdentity sends one native sparse PATCH. UpdateByIdentity adds the
// explicit Set preparation read and local list-modifier composition.
func (s *Service) PatchByIdentity(ctx context.Context, identity string, input PatchRequest) (*Message, *resty.Response, error) {
	return s.patch(ctx, identity, false, input)
}

// PatchByID sends one sparse PATCH with an explicit read-model BodyID value/null.
func (s *Service) PatchByID(ctx context.Context, id string, input PatchRequest) (*Message, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	identity, err := relativeIdentity(id)
	if err != nil {
		return nil, nil, err
	}
	return s.patch(ctx, identity, true, input)
}
func (s *Service) patch(ctx context.Context, identity string, byID bool, input PatchRequest) (*Message, *resty.Response, error) {
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
	if byID {
		if !input.BodyID.IsSet() {
			return nil, nil, ErrInput
		}
		if id, ok := input.BodyID.Value(); ok && !validModelID(id) {
			return nil, nil, ErrInput
		}
		if err := put(body, "Id", input.BodyID); err != nil {
			return nil, nil, err
		}
	} else if input.BodyID.IsSet() {
		return nil, nil, ErrInput
	}
	if !byID || input.Localizations.IsSet() {
		value := input.BodyIdentity
		if !value.IsSet() {
			value = policy.Set(identity)
		}
		if err := put(body, "Identity", value); err != nil {
			return nil, nil, err
		}
	}
	if err := force(body, input.Controls); err != nil {
		return nil, nil, err
	}
	return s.resource.Write(ctx, identity, setCommand, http.MethodPatch, body)
}

// DeleteByIdentity sends the bound Identity and optional Force body.
func (s *Service) DeleteByIdentity(ctx context.Context, identity string, controls policy.Controls) (*Message, *resty.Response, error) {
	return s.remove(ctx, identity, "Identity", identity, controls)
}

// DeleteByID uses Global/{Id} and an Id body, without a fabricated Identity field.
func (s *Service) DeleteByID(ctx context.Context, id string, controls policy.Controls) (*Message, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	identity, err := relativeIdentity(id)
	if err != nil {
		return nil, nil, err
	}
	return s.remove(ctx, identity, "Id", id, controls)
}
func (s *Service) remove(ctx context.Context, identity, key, value string, controls policy.Controls) (*Message, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	body := map[string]json.RawMessage{}
	if err := put(body, key, value); err != nil {
		return nil, nil, err
	}
	if err := force(body, controls); err != nil {
		return nil, nil, err
	}
	return s.resource.Write(ctx, identity, removeCommand, http.MethodDelete, body)
}
