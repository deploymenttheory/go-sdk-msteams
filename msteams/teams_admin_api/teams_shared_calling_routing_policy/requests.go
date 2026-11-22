package teamssharedcallingroutingpolicy

import (
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// ScalarSettings distinguishes omission, explicit null and bound zero values.
type ScalarSettings struct {
	Description     policy.Field[string] `json:"Description,omitzero"`
	ResourceAccount policy.Field[string] `json:"ResourceAccount,omitzero"`
}

// Settings sends a complete replacement array when EmergencyNumbers is nonnil.
type Settings struct {
	ScalarSettings
	EmergencyNumbers *policy.StringList `json:"EmergencyNumbers,omitempty"`
}
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}
type UpdateRequest struct {
	ScalarSettings
	EmergencyNumbers *policy.ListEdit
	policy.Controls
}
type UpdateResult struct {
	Before, After              *Policy
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "shared calling routing policy update result (redacted)" }
func (v UpdateResult) GoString() string { return v.String() }

// CreateRequest applies local edits to an initially empty emergency-number list.
type CreateRequest UpdateRequest
