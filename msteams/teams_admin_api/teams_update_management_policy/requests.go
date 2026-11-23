package teamsupdatemanagementpolicy

import (
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// ScalarSettings distinguishes omission, explicit null and bound zero values.
type ScalarSettings struct {
	Description              policy.Field[string]   `json:"Description,omitzero"`
	AllowManagedUpdates      policy.Field[bool]     `json:"AllowManagedUpdates,omitzero"`
	AllowPreview             policy.Field[bool]     `json:"AllowPreview,omitzero"`
	UpdateDayOfWeek          policy.Field[int64]    `json:"UpdateDayOfWeek,omitzero"`
	UpdateTime               policy.Field[string]   `json:"UpdateTime,omitzero"`
	UpdateTimeOfDay          policy.Field[DateTime] `json:"UpdateTimeOfDay,omitzero"`
	AllowPublicPreview       policy.Field[string]   `json:"AllowPublicPreview,omitzero"`
	UseNewTeamsClient        policy.Field[string]   `json:"UseNewTeamsClient,omitzero"`
	BlockLegacyAuthorization policy.Field[bool]     `json:"BlockLegacyAuthorization,omitzero"`
}

// Settings sends a complete replacement array when DisabledInProductMessages is nonnil.
type Settings struct {
	ScalarSettings
	DisabledInProductMessages *policy.StringList `json:"DisabledInProductMessages,omitempty"`
}
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}
type UpdateRequest struct {
	ScalarSettings
	DisabledInProductMessages *policy.ListEdit
	policy.Controls
}
type UpdateResult struct {
	Before, After              *Policy
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "TeamsUpdateManagementPolicy update result (redacted)" }
func (v UpdateResult) GoString() string { return v.String() }

// CreateRequest applies bound list edits to a fresh empty list.
type CreateRequest = UpdateRequest
