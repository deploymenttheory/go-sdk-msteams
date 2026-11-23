// Package teamsremotelogcollectiondevice manages the compiled Teams remote log
// collection device records. Each record asks a user's Teams client to upload
// diagnostic logs until its expiry.
package teamsremotelogcollectiondevice

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// ErrInput identifies local validation failures raised before any I/O.
var ErrInput = policy.ErrInput

// Device is one remote log collection request. Identity is Global/{Id}; the
// service requires DeviceId and UserId to be GUIDs and ExpireAfter to be an
// ISO 8601 UTC time. RawFields keeps every received property.
type Device struct {
	Identity    *string                    `json:"Identity"`
	ID          *string                    `json:"Id"`
	DeviceID    *string                    `json:"DeviceId"`
	UserID      *string                    `json:"UserId"`
	ExpireAfter *string                    `json:"ExpireAfter"`
	RawFields   map[string]json.RawMessage `json:"-"`
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

// String redacts the user and device identifiers.
func (Device) String() string     { return "remote log collection device (redacted)" }
func (d Device) GoString() string { return d.String() }

// Settings preserves omitted, explicit null and empty values for Set.
type Settings struct {
	DeviceID    policy.Field[string] `json:"DeviceId,omitzero"`
	UserID      policy.Field[string] `json:"UserId,omitzero"`
	ExpireAfter policy.Field[string] `json:"ExpireAfter,omitzero"`
}

// CreateRequest carries New's three mandatory values. The service, not the
// SDK, validates the GUID and date formats and rejects them with 400.
type CreateRequest struct {
	DeviceID, UserID, ExpireAfter string
	policy.Controls
}

// PatchRequest is one sparse Set PATCH. PatchByID needs the read model's Id
// (value or null) in BodyID; PatchByIdentity defaults BodyIdentity to the route.
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
	BodyID       policy.Field[string]
}

// UpdateRequest is Set's read-then-PATCH composition.
type UpdateRequest struct {
	Settings
	policy.Controls
}

// UpdateResult keeps the preparation read and whether the PATCH was sent and
// accepted, also on failure.
type UpdateResult struct {
	Before, After              *Device
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

// CreateResult keeps the generated Id and Identity even when submission fails.
type CreateResult struct {
	ID, Identity string
	Value        *Device
	HTTPAccepted bool
}

func (UpdateResult) String() string     { return "remote log collection device update (redacted)" }
func (r UpdateResult) GoString() string { return r.String() }
func (CreateResult) String() string     { return "remote log collection device create (redacted)" }
func (r CreateResult) GoString() string { return r.String() }
