// Package teamsclientconfiguration manages compiled Teams client configuration.
package teamsclientconfiguration

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput

// Configuration preserves nullable known settings and an owned complete snapshot.
// RestrictedSenderList and the other source string settings remain opaque text.
type Configuration struct {
	Identity                         *string                    `json:"Identity"`
	AllowEmailIntoChannel            *bool                      `json:"AllowEmailIntoChannel"`
	RestrictedSenderList             *string                    `json:"RestrictedSenderList"`
	AllowDropBox                     *bool                      `json:"AllowDropBox"`
	AllowBox                         *bool                      `json:"AllowBox"`
	AllowGoogleDrive                 *bool                      `json:"AllowGoogleDrive"`
	AllowShareFile                   *bool                      `json:"AllowShareFile"`
	AllowEgnyte                      *bool                      `json:"AllowEgnyte"`
	AllowOrganizationTab             *bool                      `json:"AllowOrganizationTab"`
	AllowSkypeBusinessInterop        *bool                      `json:"AllowSkypeBusinessInterop"`
	ContentPin                       *string                    `json:"ContentPin"`
	AllowResourceAccountSendMessage  *bool                      `json:"AllowResourceAccountSendMessage"`
	ResourceAccountContentAccess     *string                    `json:"ResourceAccountContentAccess"`
	AllowGuestUser                   *bool                      `json:"AllowGuestUser"`
	AllowScopedPeopleSearchAndAccess *bool                      `json:"AllowScopedPeopleSearchandAccess"`
	AllowRoleBasedChatPermissions    *bool                      `json:"AllowRoleBasedChatPermissions"`
	ExtendedWorkInfoInPeopleSearch   *bool                      `json:"ExtendedWorkInfoInPeopleSearch"`
	UseUnifiedDomain                 *string                    `json:"UseUnifiedDomain"`
	RawFields                        map[string]json.RawMessage `json:"-"`
	RawJSON                          json.RawMessage            `json:"-"`
}

func (p *Configuration) UnmarshalJSON(raw []byte) error {
	type snapshot Configuration
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*p = Configuration(next)
	return nil
}
func (p Configuration) MarshalJSON() ([]byte, error) {
	if p.RawJSON != nil {
		return append([]byte(nil), p.RawJSON...), nil
	}
	type snapshot Configuration
	return json.Marshal(snapshot(p))
}
func (Configuration) String() string     { return "client configuration (redacted)" }
func (p Configuration) GoString() string { return p.String() }

// Settings distinguishes omission, explicit null and bound empty/false values.
// The pinned setters do not impose enums, list parsing or cross-field rules.
type Settings struct {
	AllowEmailIntoChannel            policy.Field[bool]   `json:"AllowEmailIntoChannel,omitzero"`
	RestrictedSenderList             policy.Field[string] `json:"RestrictedSenderList,omitzero"`
	AllowDropBox                     policy.Field[bool]   `json:"AllowDropBox,omitzero"`
	AllowBox                         policy.Field[bool]   `json:"AllowBox,omitzero"`
	AllowGoogleDrive                 policy.Field[bool]   `json:"AllowGoogleDrive,omitzero"`
	AllowShareFile                   policy.Field[bool]   `json:"AllowShareFile,omitzero"`
	AllowEgnyte                      policy.Field[bool]   `json:"AllowEgnyte,omitzero"`
	AllowOrganizationTab             policy.Field[bool]   `json:"AllowOrganizationTab,omitzero"`
	AllowSkypeBusinessInterop        policy.Field[bool]   `json:"AllowSkypeBusinessInterop,omitzero"`
	ContentPin                       policy.Field[string] `json:"ContentPin,omitzero"`
	AllowResourceAccountSendMessage  policy.Field[bool]   `json:"AllowResourceAccountSendMessage,omitzero"`
	ResourceAccountContentAccess     policy.Field[string] `json:"ResourceAccountContentAccess,omitzero"`
	AllowGuestUser                   policy.Field[bool]   `json:"AllowGuestUser,omitzero"`
	AllowScopedPeopleSearchAndAccess policy.Field[bool]   `json:"AllowScopedPeopleSearchandAccess,omitzero"`
	AllowRoleBasedChatPermissions    policy.Field[bool]   `json:"AllowRoleBasedChatPermissions,omitzero"`
	ExtendedWorkInfoInPeopleSearch   policy.Field[bool]   `json:"ExtendedWorkInfoInPeopleSearch,omitzero"`
	UseUnifiedDomain                 policy.Field[string] `json:"UseUnifiedDomain,omitzero"`
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

func (UpdateResult) String() string     { return "client configuration update result (redacted)" }
func (p UpdateResult) GoString() string { return p.String() }
