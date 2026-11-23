// Package teamsbyodanddeskspolicy manages compiled Teams BYOD and desks policy.
package teamsbyodanddeskspolicy

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput

// Policy preserves nullable known settings and an owned complete snapshot.
// The source DeviceDataCollection string remains opaque and nullable.
type Policy struct {
	Identity             *string                    `json:"Identity"`
	DeviceDataCollection *string                    `json:"DeviceDataCollection"`
	RawFields            map[string]json.RawMessage `json:"-"`
	RawJSON              json.RawMessage            `json:"-"`
}

func (p *Policy) UnmarshalJSON(raw []byte) error {
	type snapshot Policy
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*p = Policy(next)
	return nil
}
func (p Policy) MarshalJSON() ([]byte, error) {
	if p.RawJSON != nil {
		return append([]byte(nil), p.RawJSON...), nil
	}
	type snapshot Policy
	return json.Marshal(snapshot(p))
}
func (Policy) String() string     { return "BYOD and desks policy (redacted)" }
func (p Policy) GoString() string { return p.String() }

// Settings distinguishes omission, explicit null and bound empty values.
// The pinned setter binds DeviceDataCollection as an opaque string with no enum.
type Settings struct {
	DeviceDataCollection policy.Field[string] `json:"DeviceDataCollection,omitzero"`
}

// CreateRequest carries optional settings for one compiled New POST.
type CreateRequest struct {
	Settings
	policy.Controls
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
	Before, After              *Policy
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "BYOD and desks policy update result (redacted)" }
func (p UpdateResult) GoString() string { return p.String() }

func (Settings) String() string     { return "BYOD and desks policy settings (redacted)" }
func (p Settings) GoString() string { return p.String() }
