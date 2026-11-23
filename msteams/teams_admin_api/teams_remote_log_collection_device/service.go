package teamsremotelogcollectiondevice

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

const (
	getCommand    = "Get-CsTeamsRemoteLogCollectionDevice"
	newCommand    = "New-CsTeamsRemoteLogCollectionDevice"
	setCommand    = "Set-CsTeamsRemoteLogCollectionDevice"
	removeCommand = "Remove-CsTeamsRemoteLogCollectionDevice"
)

// Service manages the compiled TeamsRemoteLogCollectionDevice resource.
type Service struct{ resource *policyrp.Resource[Device] }

// NewService constructs a service without credentials, discovery or network I/O.
func NewService(scope *client.Scope) *Service {
	return &Service{resource: policyrp.New(scope, "TeamsRemoteLogCollectionDevice", func(v *Device) *string { return v.Identity })}
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
	if id == "" || strings.Contains(id, "/") || !policyrp.ValidIdentity("Global/"+id) {
		return "", ErrInput
	}
	return "Global/" + id, nil
}

// GetByIdentity reads one device by its full identity, such as Global/{Id}
// (Get-CsTeamsRemoteLogCollectionDevice -Identity).
func (s *Service) GetByIdentity(ctx context.Context, identity string) (*Device, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.resource.Get(ctx, identity, getCommand)
}

// GetByID reads Global/{Id} (Get-CsTeamsRemoteLogCollectionDevice -Id).
func (s *Service) GetByID(ctx context.Context, id string) (*Device, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	identity, err := relativeIdentity(id)
	if err != nil {
		return nil, nil, err
	}
	return s.resource.Get(ctx, identity, getCommand)
}

// List returns the bare collection without invented paging.
func (s *Service) List(ctx context.Context) ([]*Device, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.resource.List(ctx, getCommand)
}

// ListFiltered matches Identity wildcards locally (-Filter); empty is invalid.
func (s *Service) ListFiltered(ctx context.Context, pattern string) ([]*Device, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if pattern == "" {
		return nil, nil, ErrInput
	}
	return s.resource.ListFiltered(ctx, pattern, getCommand)
}

// CreateByIdentity posts New's Identity set, adding the identity's second
// component as Id metadata.
func (s *Service) CreateByIdentity(ctx context.Context, identity string, input CreateRequest) (*Device, *resty.Response, error) {
	return s.create(ctx, identity, "identity", input)
}

// CreateByID binds Id and sends the source DTO's null Id body value.
func (s *Service) CreateByID(ctx context.Context, id string, input CreateRequest) (*Device, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	identity, err := relativeIdentity(id)
	if err != nil {
		return nil, nil, err
	}
	return s.create(ctx, identity, "id", input)
}

// Create generates one UUID Id, as New does when Id is not bound. The result
// keeps the Id even if submission fails; the POST is never replayed.
func (s *Service) Create(ctx context.Context, input CreateRequest) (*CreateResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !input.complete() {
		return nil, nil, ErrInput
	}
	var raw [16]byte
	if _, err := rand.Read(raw[:]); err != nil {
		return nil, nil, err
	}
	raw[6] = raw[6]&0x0f | 0x40
	raw[8] = raw[8]&0x3f | 0x80
	id := fmt.Sprintf("%x-%x-%x-%x-%x", raw[:4], raw[4:6], raw[6:8], raw[8:10], raw[10:])
	result := &CreateResult{ID: id, Identity: "Global/" + id}
	value, response, err := s.create(ctx, result.Identity, "auto", input)
	result.Value = value
	result.HTTPAccepted = response != nil && response.StatusCode() >= 200 && response.StatusCode() < 300
	return result, response, err
}

func (r CreateRequest) complete() bool {
	return r.DeviceID != "" && r.UserID != "" && r.ExpireAfter != ""
}

func (s *Service) create(ctx context.Context, identity, selector string, input CreateRequest) (*Device, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if !input.complete() || !policyrp.ValidIdentity(identity) {
		return nil, nil, ErrInput
	}
	body, err := policyrp.Fields(Settings{DeviceID: policy.Set(input.DeviceID), UserID: policy.Set(input.UserID), ExpireAfter: policy.Set(input.ExpireAfter)})
	if err != nil {
		return nil, nil, err
	}
	put(body, "Identity", identity)
	switch {
	case selector == "id":
		put(body, "Id", nil)
	case selector == "identity" && strings.Contains(identity, "/"):
		put(body, "Id", strings.SplitN(identity, "/", 2)[1])
	}
	force(body, input.Controls)
	return s.resource.Write(ctx, identity, newCommand, http.MethodPost, body)
}

// PatchByIdentity sends one sparse PATCH; UpdateByIdentity adds Set's read.
func (s *Service) PatchByIdentity(ctx context.Context, identity string, input PatchRequest) (*Device, *resty.Response, error) {
	return s.patch(ctx, identity, false, input)
}

// PatchByID sends one sparse PATCH with an explicit read-model BodyID.
func (s *Service) PatchByID(ctx context.Context, id string, input PatchRequest) (*Device, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	identity, err := relativeIdentity(id)
	if err != nil {
		return nil, nil, err
	}
	return s.patch(ctx, identity, true, input)
}

func (s *Service) patch(ctx context.Context, identity string, byID bool, input PatchRequest) (*Device, *resty.Response, error) {
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
	if byID {
		if input.BodyIdentity.IsSet() || !input.BodyID.IsSet() {
			return nil, nil, ErrInput
		}
		put(body, "Id", input.BodyID)
	} else {
		if input.BodyID.IsSet() {
			return nil, nil, ErrInput
		}
		value := input.BodyIdentity
		if !value.IsSet() {
			value = policy.Set(identity)
		}
		put(body, "Identity", value)
	}
	force(body, input.Controls)
	return s.resource.Write(ctx, identity, setCommand, http.MethodPatch, body)
}

// UpdateByIdentity reads with Set's command metadata, then PATCHes once.
func (s *Service) UpdateByIdentity(ctx context.Context, identity string, input UpdateRequest) (*UpdateResult, *resty.Response, error) {
	return s.update(ctx, identity, false, input)
}

// UpdateByID reads Global/{Id}, then PATCHes with the read model's Id.
func (s *Service) UpdateByID(ctx context.Context, id string, input UpdateRequest) (*UpdateResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	identity, err := relativeIdentity(id)
	if err != nil {
		return nil, nil, err
	}
	return s.update(ctx, identity, true, input)
}

func (s *Service) update(ctx context.Context, identity string, byID bool, input UpdateRequest) (*UpdateResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if body, err := policyrp.Fields(input.Settings); err != nil || len(body) == 0 {
		return nil, nil, ErrInput
	}
	before, read, err := s.resource.Get(ctx, identity, setCommand)
	result := &UpdateResult{Before: before, ReadResponse: read}
	if err != nil {
		return result, read, err
	}
	if before == nil {
		return result, read, client.ErrDecode
	}
	patch := PatchRequest{Settings: input.Settings, Controls: input.Controls}
	if byID {
		patch.BodyID = field(before.ID)
	} else {
		patch.BodyIdentity = field(before.Identity)
	}
	result.PatchInvoked = true
	result.After, read, err = s.patch(ctx, identity, byID, patch)
	result.HTTPAccepted = read != nil && read.StatusCode() >= 200 && read.StatusCode() < 300
	return result, read, err
}

// DeleteByIdentity sends the bound Identity and optional Force body.
func (s *Service) DeleteByIdentity(ctx context.Context, identity string, controls policy.Controls) (*Device, *resty.Response, error) {
	return s.remove(ctx, identity, "Identity", identity, controls)
}

// DeleteByID deletes Global/{Id} with an Id body and no Identity field.
func (s *Service) DeleteByID(ctx context.Context, id string, controls policy.Controls) (*Device, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	identity, err := relativeIdentity(id)
	if err != nil {
		return nil, nil, err
	}
	return s.remove(ctx, identity, "Id", id, controls)
}

func (s *Service) remove(ctx context.Context, identity, key, value string, controls policy.Controls) (*Device, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	body := map[string]json.RawMessage{}
	put(body, key, value)
	force(body, controls)
	return s.resource.Write(ctx, identity, removeCommand, http.MethodDelete, body)
}

func field(value *string) policy.Field[string] {
	if value == nil {
		return policy.Null[string]()
	}
	return policy.Set(*value)
}

// put marshals strings, nil and policy fields, none of which can fail.
func put(body map[string]json.RawMessage, key string, value any) {
	body[key], _ = json.Marshal(value)
}

func force(body map[string]json.RawMessage, controls policy.Controls) {
	if controls.Force == nil {
		return
	}
	value := "False"
	if *controls.Force {
		value = "True"
	}
	put(body, "Force", value)
}
