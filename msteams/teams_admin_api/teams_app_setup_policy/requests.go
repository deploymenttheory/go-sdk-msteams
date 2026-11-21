package teamsappsetuppolicy

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

type ScalarSettings struct {
	Description      policy.Field[string] `json:"Description,omitzero"`
	AllowSideLoading policy.Field[bool]   `json:"AllowSideLoading,omitzero"`
	AllowUserPinning policy.Field[bool]   `json:"AllowUserPinning,omitzero"`
}

// Settings uses full arrays for sparse native writes. Nil slices encode []; an
// explicit null array field is not a source DTO shape. Edits belong to UpdateRequest.
type Settings struct {
	ScalarSettings
	AppPresetList               policy.Field[[]*App]
	PinnedAppBarApps            policy.Field[[]*PinnedApp]
	PinnedMessageBarApps        policy.Field[[]*PinnedApp]
	AppPresetMeetingList        policy.Field[[]*App]
	AdditionalCustomizationApps policy.Field[[]*CustomizationApp]
	PinnedCallingBarApps        policy.Field[[]*App]
}
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}
type UpdateRequest struct {
	ScalarSettings
	AppPresetList               *AppEdit
	PinnedAppBarApps            *PinEdit
	PinnedMessageBarApps        *PinEdit
	AppPresetMeetingList        *AppEdit
	AdditionalCustomizationApps *CustomizationEdit
	PinnedCallingBarApps        *AppEdit
	policy.Controls
}
type CreateRequest UpdateRequest
type UpdateResult struct {
	Before, After              *Policy
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "app setup update result (redacted)" }
func (v UpdateResult) GoString() string { return v.String() }

func appDTO(v []*App) []map[string]any {
	out := make([]map[string]any, len(v))
	for i, x := range v {
		if x != nil {
			out[i] = map[string]any{"Identity": x.Identity, "Id": x.ID}
		}
	}
	return out
}
func pinDTO(v []*PinnedApp) []map[string]any {
	out := make([]map[string]any, len(v))
	for i, x := range v {
		if x != nil {
			out[i] = map[string]any{"Identity": x.Identity, "Id": x.ID, "Order": x.Order}
		}
	}
	return out
}
func customizationDTO(v []*CustomizationApp) []map[string]any {
	out := make([]map[string]any, len(v))
	for i, x := range v {
		if x != nil {
			out[i] = map[string]any{"Identity": x.Identity, "Id": x.ID, "AdditionalCustomizationId": x.AdditionalCustomizationID}
		}
	}
	return out
}
func arrayField[T any](body map[string]json.RawMessage, key string, value policy.Field[[]*T], project func([]*T) []map[string]any) error {
	if !value.IsSet() {
		return nil
	}
	if value.IsNull() {
		return ErrInput
	}
	items, _ := value.Value()
	raw, e := json.Marshal(project(items))
	if e != nil {
		return ErrInput
	}
	body[key] = raw
	return nil
}
func fields(v Settings) (map[string]json.RawMessage, error) {
	body, e := policyrp.Fields(v.ScalarSettings)
	if e != nil {
		return nil, e
	}
	for _, call := range []func() error{
		func() error { return arrayField(body, "AppPresetList", v.AppPresetList, appDTO) }, func() error { return arrayField(body, "PinnedAppBarApps", v.PinnedAppBarApps, pinDTO) }, func() error { return arrayField(body, "PinnedMessageBarApps", v.PinnedMessageBarApps, pinDTO) }, func() error { return arrayField(body, "AppPresetMeetingList", v.AppPresetMeetingList, appDTO) }, func() error {
			return arrayField(body, "AdditionalCustomizationApps", v.AdditionalCustomizationApps, customizationDTO)
		}, func() error { return arrayField(body, "PinnedCallingBarApps", v.PinnedCallingBarApps, appDTO) },
	} {
		if e := call(); e != nil {
			return nil, e
		}
	}
	return body, nil
}
func validateEdits(v UpdateRequest) error {
	for _, e := range []*AppEdit{v.AppPresetList, v.AppPresetMeetingList, v.PinnedCallingBarApps} {
		if e != nil {
			if err := e.Validate(); err != nil {
				return err
			}
		}
	}
	for _, e := range []*PinEdit{v.PinnedAppBarApps, v.PinnedMessageBarApps} {
		if e != nil {
			if err := e.Validate(); err != nil {
				return err
			}
		}
	}
	if v.AdditionalCustomizationApps != nil {
		return v.AdditionalCustomizationApps.Validate()
	}
	return nil
}
func boundEdits(v UpdateRequest) bool {
	return v.AppPresetList != nil || v.PinnedAppBarApps != nil || v.PinnedMessageBarApps != nil || v.AppPresetMeetingList != nil || v.AdditionalCustomizationApps != nil || v.PinnedCallingBarApps != nil
}
func resolve(v UpdateRequest, before *Policy) (Settings, error) {
	if e := validateEdits(v); e != nil {
		return Settings{}, e
	}
	if before == nil {
		before = &Policy{}
	}
	out := Settings{ScalarSettings: v.ScalarSettings}
	for _, c := range []struct {
		edit    *AppEdit
		current []*App
		field   *policy.Field[[]*App]
	}{{v.AppPresetList, before.AppPresetList, &out.AppPresetList}, {v.AppPresetMeetingList, before.AppPresetMeetingList, &out.AppPresetMeetingList}, {v.PinnedCallingBarApps, before.PinnedCallingBarApps, &out.PinnedCallingBarApps}} {
		if c.edit != nil {
			items, e := c.edit.Apply(c.current)
			if e != nil {
				return Settings{}, e
			}
			*c.field = policy.Set(items)
		}
	}
	for _, c := range []struct {
		edit    *PinEdit
		current []*PinnedApp
		field   *policy.Field[[]*PinnedApp]
	}{{v.PinnedAppBarApps, before.PinnedAppBarApps, &out.PinnedAppBarApps}, {v.PinnedMessageBarApps, before.PinnedMessageBarApps, &out.PinnedMessageBarApps}} {
		if c.edit != nil {
			items, e := c.edit.Apply(c.current)
			if e != nil {
				return Settings{}, e
			}
			*c.field = policy.Set(items)
		}
	}
	if v.AdditionalCustomizationApps != nil {
		items, e := v.AdditionalCustomizationApps.Apply(before.AdditionalCustomizationApps)
		if e != nil {
			return Settings{}, e
		}
		out.AdditionalCustomizationApps = policy.Set(items)
	}
	return out, nil
}
