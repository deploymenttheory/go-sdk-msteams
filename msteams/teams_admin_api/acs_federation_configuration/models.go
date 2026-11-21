// Package acsfederationconfiguration manages the compiled TeamsAcsFederationConfiguration ConfigAPI resource.
package acsfederationconfiguration

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput

// Configuration retains nullable/omitted/future wire fields. RawFields is authoritative;
// native reads do not materialize defaults from the PowerShell CLR constructor.
type Configuration struct {
	Identity                       *string                    `json:"Identity"`
	LabelForAllowedAcsUsers        *string                    `json:"LabelForAllowedAcsUsers"`
	EnableAcsUsers                 *bool                      `json:"EnableAcsUsers"`
	RequireAcsFederationForMeeting *bool                      `json:"RequireAcsFederationForMeeting"`
	HideBannerForAllowedAcsUsers   *bool                      `json:"HideBannerForAllowedAcsUsers"`
	AllowedAcsResources            []string                   `json:"AllowedAcsResources"`
	RawFields                      map[string]json.RawMessage `json:"-"`
}

func (v *Configuration) UnmarshalJSON(data []byte) error {
	type snapshot Configuration
	var value snapshot
	if json.Unmarshal(data, &value) != nil || json.Unmarshal(data, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = Configuration(value)
	return nil
}
func (v Configuration) MarshalJSON() ([]byte, error) {
	if v.RawFields != nil {
		return json.Marshal(v.RawFields)
	}
	type snapshot Configuration
	return json.Marshal(snapshot(v))
}
func (Configuration) String() string {
	return "compiled TeamsAcsFederationConfiguration snapshot (redacted)"
}
func (v Configuration) GoString() string { return v.String() }

// Fields preserves omitted/null/value scalar binding for sparse request bodies.
type Fields struct {
	LabelForAllowedAcsUsers        policy.Field[string] `json:"LabelForAllowedAcsUsers,omitzero"`
	EnableAcsUsers                 policy.Field[bool]   `json:"EnableAcsUsers,omitzero"`
	RequireAcsFederationForMeeting policy.Field[bool]   `json:"RequireAcsFederationForMeeting,omitzero"`
	HideBannerForAllowedAcsUsers   policy.Field[bool]   `json:"HideBannerForAllowedAcsUsers,omitzero"`
}

// Settings contains fully materialized request arrays. A nil list pointer omits
// the property; policy.Strings() sends an explicit empty array.
type Settings struct {
	Fields
	AllowedAcsResources *policy.StringList `json:"AllowedAcsResources,omitempty"`
}

// PatchRequest is one native sparse PATCH. BodyIdentity defaults to the route
// identity; compositions bind the preceding read model's identity explicitly.
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}

// UpdateRequest composes a read, local list edits and one sparse PATCH. No ETag,
// atomic Add/Remove, automatic retry or rollback is inferred.
type UpdateRequest struct {
	Fields
	policy.Controls
	AllowedAcsResources *policy.ListEdit
}

// UpdateResult preserves preparation and HTTP acceptance if the later step fails.
// PatchInvoked records the SDK call, not proof of a network send or propagation.
type UpdateResult struct {
	Before       *Configuration
	After        *Configuration
	ReadResponse *resty.Response
	PatchInvoked bool
	HTTPAccepted bool
}

func (UpdateResult) String() string     { return "compiled policy update result (redacted)" }
func (v UpdateResult) GoString() string { return v.String() }
