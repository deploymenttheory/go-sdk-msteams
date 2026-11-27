// Package teamssipdevicesconfiguration manages compiled Teams SIP devices configuration.
package teamssipdevicesconfiguration

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput

// Configuration preserves nullable known settings and an owned complete snapshot.
// All source setting strings remain opaque and nullable.
type Configuration struct {
	Identity   *string                    `json:"Identity"`
	BulkSignIn *string                    `json:"BulkSignIn"`
	RawFields  map[string]json.RawMessage `json:"-"`
	RawJSON    json.RawMessage            `json:"-"`
}

func (p *Configuration) UnmarshalJSON(raw []byte) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return client.ErrDecode
	}
	var wire struct {
		Identity   policyrp.ReadString `json:"Identity"`
		BulkSignIn policyrp.ReadString `json:"BulkSignIn"`
	}
	if json.Unmarshal(raw, &wire) != nil {
		return client.ErrDecode
	}
	*p = Configuration{
		Identity:   wire.Identity.Value,
		BulkSignIn: wire.BulkSignIn.Value,
		RawFields:  fields, RawJSON: append(json.RawMessage(nil), raw...),
	}
	return nil
}
func (p Configuration) MarshalJSON() ([]byte, error) {
	if p.RawJSON != nil {
		return append([]byte(nil), p.RawJSON...), nil
	}
	type snapshot Configuration
	return json.Marshal(snapshot(p))
}
func (Configuration) String() string     { return "SIP devices configuration (redacted)" }
func (p Configuration) GoString() string { return p.String() }

// Settings distinguishes omission, explicit null and bound empty/false values.
// The pinned setters do not impose enums, list parsing or cross-field rules.
type Settings struct {
	BulkSignIn policy.Field[string] `json:"BulkSignIn,omitzero"`
}

// PatchRequest selects settings for a single PATCH, with optional body Identity.
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}

// UpdateRequest selects settings for the explicit Set read-then-PATCH composition.
type UpdateRequest struct {
	Settings
	policy.Controls
}

// UpdateResult retains preparation and HTTP acceptance on partial failure.
// Acceptance does not establish persisted or effective tenant settings.
type UpdateResult struct {
	Before, After              *Configuration
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string {
	return "SIP devices configuration update result (redacted)"
}
func (p UpdateResult) GoString() string { return p.String() }

func (Settings) String() string     { return "SIP devices configuration settings (redacted)" }
func (p Settings) GoString() string { return p.String() }
