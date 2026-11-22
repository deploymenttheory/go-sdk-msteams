package teamstemplatepermissionpolicy

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// ScalarSettings preserves omitted, null and opaque string values.
type ScalarSettings struct {
	Description policy.Field[string] `json:"Description,omitzero"`
}

// Settings is a sparse native write. A bound nil slice clears the array.
type Settings struct {
	ScalarSettings
	HiddenTemplates policy.Field[[]*HiddenTemplate]
}
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}

// UpdateRequest applies local Id-keyed list edits to the preparation read.
type UpdateRequest struct {
	ScalarSettings
	HiddenTemplates *HiddenTemplateEdit
	policy.Controls
}
type CreateRequest UpdateRequest
type UpdateResult struct {
	Before, After              *Policy
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "template permission update result (redacted)" }
func (v UpdateResult) GoString() string { return v.String() }
func fields(input Settings) (map[string]json.RawMessage, error) {
	body, err := policyrp.Fields(input.ScalarSettings)
	if err != nil {
		return nil, err
	}
	if input.HiddenTemplates.IsSet() {
		if input.HiddenTemplates.IsNull() {
			return nil, ErrInput
		}
		items, _ := input.HiddenTemplates.Value()
		raw, err := json.Marshal(templateDTO(items))
		if err != nil {
			return nil, ErrInput
		}
		body["HiddenTemplates"] = raw
	}
	return body, nil
}
func validateEdits(input UpdateRequest) error {
	if input.HiddenTemplates != nil {
		return input.HiddenTemplates.Validate()
	}
	return nil
}
func resolve(input UpdateRequest, before *Policy) (Settings, error) {
	if err := validateEdits(input); err != nil {
		return Settings{}, err
	}
	out := Settings{ScalarSettings: input.ScalarSettings}
	if input.HiddenTemplates != nil {
		var current []*HiddenTemplate
		if before != nil {
			current = before.HiddenTemplates
		}
		value, err := input.HiddenTemplates.Apply(current)
		if err != nil {
			return Settings{}, err
		}
		out.HiddenTemplates = policy.Set(value)
	}
	return out, nil
}
