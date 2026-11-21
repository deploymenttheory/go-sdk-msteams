package recordingmessages

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// Settings is a sparse native write. A bound localization slice is a full array;
// nil slices become [], while an explicit null field fails the model's list rule.
type Settings struct {
	Description   policy.Field[string]
	Localizations policy.Field[[]*Localization]
}
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
	// BodyID is required by native PatchByID. It comes from the read model in
	// UpdateByID; the service never guesses it from a possibly different route ID.
	BodyID policy.Field[string]
}
type UpdateRequest struct {
	Description   policy.Field[string]
	Localizations *LocalizationEdit
	policy.Controls
}

// CreateRequest applies local edits to a fresh empty list. No setting is required
// by the pinned New base; Set/Update has a separate nonempty-settings rule.
type CreateRequest UpdateRequest
type UpdateResult struct {
	Before, After              *Message
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

// CreateResult retains the selected generated ID even when submission fails.
type CreateResult struct {
	ID, Identity string
	Value        *Message
	HTTPAccepted bool
}

func (UpdateResult) String() string     { return "recording message update result (redacted)" }
func (v UpdateResult) GoString() string { return v.String() }
func (CreateResult) String() string     { return "recording message create result (redacted)" }
func (v CreateResult) GoString() string { return v.String() }

func fields(input Settings) (map[string]json.RawMessage, error) {
	body := map[string]json.RawMessage{}
	if input.Description.IsSet() {
		raw, err := json.Marshal(input.Description)
		if err != nil {
			return nil, ErrInput
		}
		body["Description"] = raw
	}
	if input.Localizations.IsSet() {
		if input.Localizations.IsNull() {
			return nil, ErrInput
		}
		list, _ := input.Localizations.Value()
		raw, err := json.Marshal(localizationDTO(list))
		if err != nil {
			return nil, ErrInput
		}
		body["RecordingAndTranscriptionLocalizationCustomMessage"] = raw
	}
	return body, nil
}
func resolve(input UpdateRequest, current []*Localization) (Settings, error) {
	result := Settings{Description: input.Description}
	if input.Localizations != nil {
		values, err := input.Localizations.Apply(current)
		if err != nil {
			return Settings{}, err
		}
		result.Localizations = policy.Set(values)
	}
	return result, nil
}
func valueField(v *string) policy.Field[string] {
	if v == nil {
		return policy.Null[string]()
	}
	return policy.Set(*v)
}
func put(body map[string]json.RawMessage, key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return ErrInput
	}
	body[key] = raw
	return nil
}
func force(body map[string]json.RawMessage, controls policy.Controls) error {
	if controls.Force == nil {
		return nil
	}
	value := "False"
	if *controls.Force {
		value = "True"
	}
	return put(body, "Force", value)
}
