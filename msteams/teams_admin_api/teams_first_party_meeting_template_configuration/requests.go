package teamsfirstpartymeetingtemplateconfiguration

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// ScalarSettings distinguishes an omitted, explicit null and bound Description.
type ScalarSettings struct {
	Description policy.Field[string] `json:"Description,omitzero"`
}

// Settings is a sparse native write. A bound template slice is sent as a complete
// replacement array; a bound nil slice sends `[]`.
type Settings struct {
	ScalarSettings
	TeamsMeetingTemplates policy.Field[[]*Template]
}
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}

// UpdateRequest applies local Name-keyed template edits to the preparation read.
type UpdateRequest struct {
	ScalarSettings
	TeamsMeetingTemplates *TemplateEdit
	policy.Controls
}
type UpdateResult struct {
	Before, After              *Configuration
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "first-party meeting template update result (redacted)" }
func (v UpdateResult) GoString() string { return v.String() }

func fields(input Settings) (map[string]json.RawMessage, error) {
	body, err := policyrp.Fields(input.ScalarSettings)
	if err != nil {
		return nil, err
	}
	if input.TeamsMeetingTemplates.IsSet() {
		if input.TeamsMeetingTemplates.IsNull() {
			return nil, ErrInput
		}
		items, _ := input.TeamsMeetingTemplates.Value()
		raw, err := json.Marshal(templatesDTO(items))
		if err != nil {
			return nil, ErrInput
		}
		body["TeamsMeetingTemplates"] = raw
	}
	return body, nil
}
func validateEdits(input UpdateRequest) error {
	if input.TeamsMeetingTemplates != nil {
		return input.TeamsMeetingTemplates.Validate()
	}
	return nil
}

// resolve mirrors the compiled serializer: binding TeamsMeetingTemplates expands
// the nested template property names, which include Description, so the
// prepared root Description is serialized even when it was not bound.
func resolve(input UpdateRequest, before *Configuration) (Settings, error) {
	if err := validateEdits(input); err != nil {
		return Settings{}, err
	}
	out := Settings{ScalarSettings: input.ScalarSettings}
	if input.TeamsMeetingTemplates != nil {
		var current []*Template
		if before != nil {
			current = before.TeamsMeetingTemplates
		}
		value, err := input.TeamsMeetingTemplates.Apply(current)
		if err != nil {
			return Settings{}, err
		}
		out.TeamsMeetingTemplates = policy.Set(value)
		if !out.Description.IsSet() {
			out.Description = policy.Null[string]()
			if before != nil && before.Description != nil {
				out.Description = policy.Set(*before.Description)
			}
		}
	}
	return out, nil
}
