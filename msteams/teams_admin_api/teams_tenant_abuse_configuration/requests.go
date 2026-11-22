package teamstenantabuseconfiguration

import (
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// ScalarSettings distinguishes omission, explicit null and bound zero values.
type ScalarSettings struct {
	LowSeatLimit                         policy.Field[uint64] `json:"LowSeatLimit,omitzero"`
	TrialTenantValidation                policy.Field[string] `json:"TrialTenantValidation,omitzero"`
	SkipValidation                       policy.Field[string] `json:"SkipValidation,omitzero"`
	CreateThreadThresholdPerSeat         policy.Field[uint64] `json:"CreateThreadThresholdPerSeat,omitzero"`
	AddMemberThresholdPerSeat            policy.Field[uint64] `json:"AddMemberThresholdPerSeat,omitzero"`
	SendMessageThresholdPerSeat          policy.Field[uint64] `json:"SendMessageThresholdPerSeat,omitzero"`
	CreateThreadThresholdForTrialTenants policy.Field[uint64] `json:"CreateThreadThresholdForTrialTenants,omitzero"`
	AddMemberThresholdForTrialTenants    policy.Field[uint64] `json:"AddMemberThresholdForTrialTenants,omitzero"`
	SendMessageThresholdForTrialTenants  policy.Field[uint64] `json:"SendMessageThresholdForTrialTenants,omitzero"`
}

// Settings sends a complete replacement array when BlockedTenants is nonnil.
type Settings struct {
	ScalarSettings
	BlockedTenants *policy.StringList `json:"BlockedTenants,omitempty"`
}
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}
type UpdateRequest struct {
	ScalarSettings
	BlockedTenants *policy.ListEdit
	policy.Controls
}
type UpdateResult struct {
	Before, After              *Configuration
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "TeamsTenantAbuseConfiguration update result (redacted)" }
func (v UpdateResult) GoString() string { return v.String() }
