package audioconferencingprompts

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// Settings contains complete typed arrays for sparse PATCH writes.
// Nil bound slices emit []; explicit null fields are not source DTO arrays.
type Settings struct {
	Prompts  policy.Field[[]*Prompt]
	Packages policy.Field[[]*PromptPackage]
}
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}
type UpdateRequest struct {
	Prompts  *PromptEdit
	Packages *PackageEdit
	policy.Controls
}
type UpdateResult struct {
	Before, After              *Configuration
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "audio conferencing prompts update result (redacted)" }
func (v UpdateResult) GoString() string { return v.String() }

func promptDTO(x *Prompt) map[string]any {
	if x == nil {
		return nil
	}
	return map[string]any{"Identity": x.Identity, "Id": x.ID, "Name": x.Name, "AudioPrompt": x.AudioPrompt, "TextPrompt": x.TextPrompt, "Type": x.Type, "Locale": x.Locale}
}
func packageDTO(x *PromptPackage) map[string]any {
	if x == nil {
		return nil
	}
	return map[string]any{
		"Identity": x.Identity, "Id": x.ID, "Name": x.Name, "DefaultLocale": x.DefaultLocale,
		"InboundStartRecordingPrompt":  copyItems(x.InboundStartRecordingPrompt, cloneString),
		"InboundEndRecordingPrompt":    copyItems(x.InboundEndRecordingPrompt, cloneString),
		"OutboundStartRecordingPrompt": copyItems(x.OutboundStartRecordingPrompt, cloneString),
		"OutboundEndRecordingPrompt":   copyItems(x.OutboundEndRecordingPrompt, cloneString),
		"MeetingStartRecordingPrompt":  copyItems(x.MeetingStartRecordingPrompt, cloneString),
		"MeetingEndRecordingPrompt":    copyItems(x.MeetingEndRecordingPrompt, cloneString),
	}
}
func project[T any](values []*T, dto func(*T) map[string]any) []map[string]any {
	out := make([]map[string]any, len(values))
	for i, v := range values {
		out[i] = dto(v)
	}
	return out
}
func arrayField[T any](body map[string]json.RawMessage, key string, value policy.Field[[]*T], dto func(*T) map[string]any, valid func(*T) bool) error {
	if !value.IsSet() {
		return nil
	}
	if value.IsNull() {
		return ErrInput
	}
	items, _ := value.Value()
	for _, v := range items {
		if !valid(v) {
			return ErrInput
		}
	}
	raw, e := json.Marshal(project(items, dto))
	if e != nil {
		return ErrInput
	}
	body[key] = raw
	return nil
}
func fields(v Settings) (map[string]json.RawMessage, error) {
	out := make(map[string]json.RawMessage)
	if e := arrayField(out, "Prompts", v.Prompts, promptDTO, prompts().validItem); e != nil {
		return nil, e
	}
	if e := arrayField(out, "Packages", v.Packages, packageDTO, packages().validItem); e != nil {
		return nil, e
	}
	return out, nil
}
func validateEdits(v UpdateRequest) error {
	if v.Prompts != nil {
		if e := v.Prompts.Validate(); e != nil {
			return e
		}
	}
	if v.Packages != nil {
		return v.Packages.Validate()
	}
	return nil
}
func resolve(v UpdateRequest, before *Configuration) (Settings, error) {
	var out Settings
	if v.Prompts != nil {
		items, e := v.Prompts.Apply(before.Prompts)
		if e != nil {
			return out, e
		}
		out.Prompts = policy.Set(items)
	}
	if v.Packages != nil {
		items, e := v.Packages.Apply(before.Packages)
		if e != nil {
			return out, e
		}
		out.Packages = policy.Set(items)
	}
	return out, nil
}
