// Package teamsremotelogcollectionconfiguration reads the compiled Teams
// remote log collection configuration, which lists the tenant's pending
// remote log collection devices. The module has no cmdlet to change it
// directly; manage devices with the teams_remote_log_collection_device package.
package teamsremotelogcollectionconfiguration

import (
	"context"
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

const getCommand = "Get-CsTeamsRemoteLogCollectionConfiguration"

// ErrInput identifies local validation failures raised before any I/O.
var ErrInput = policy.ErrInput

// Configuration is one configuration instance; the tenant has Global.
// RawFields keeps every received property, including ConfigMetadata.
type Configuration struct {
	Identity  *string                    `json:"Identity"`
	Devices   []*Device                  `json:"Devices"`
	RawFields map[string]json.RawMessage `json:"-"`
}

// Device is a device entry embedded in the configuration.
type Device struct {
	Identity    *string                    `json:"Identity"`
	ID          *string                    `json:"Id"`
	DeviceID    *string                    `json:"DeviceId"`
	UserID      *string                    `json:"UserId"`
	ExpireAfter *string                    `json:"ExpireAfter"`
	RawFields   map[string]json.RawMessage `json:"-"`
}

func (c *Configuration) UnmarshalJSON(raw []byte) error {
	type plain Configuration
	var value plain
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*c = Configuration(value)
	return nil
}

func (d *Device) UnmarshalJSON(raw []byte) error {
	type plain Device
	var value plain
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*d = Device(value)
	return nil
}

// String redacts the embedded user and device identifiers.
func (Configuration) String() string     { return "remote log collection configuration (redacted)" }
func (c Configuration) GoString() string { return c.String() }
func (Device) String() string            { return "remote log collection device (redacted)" }
func (d Device) GoString() string        { return d.String() }

// Service reads the compiled TeamsRemoteLogCollectionConfiguration resource.
type Service struct {
	resource *policyrp.Resource[Configuration]
}

// NewService constructs a service without credentials, discovery or network I/O.
func NewService(scope *client.Scope) *Service {
	return &Service{resource: policyrp.New(scope, "TeamsRemoteLogCollectionConfiguration", func(v *Configuration) *string { return v.Identity })}
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

// GetByIdentity reads one instance, such as Global
// (Get-CsTeamsRemoteLogCollectionConfiguration -Identity).
func (s *Service) GetByIdentity(ctx context.Context, identity string) (*Configuration, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.resource.Get(ctx, identity, getCommand)
}

// List returns every instance, Global first, without invented paging.
func (s *Service) List(ctx context.Context) ([]*Configuration, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.resource.List(ctx, getCommand)
}

// ListFiltered matches Identity wildcards locally (-Filter); empty is invalid.
func (s *Service) ListFiltered(ctx context.Context, pattern string) ([]*Configuration, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if pattern == "" {
		return nil, nil, ErrInput
	}
	return s.resource.ListFiltered(ctx, pattern, getCommand)
}
