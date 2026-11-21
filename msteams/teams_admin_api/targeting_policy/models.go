// Package targetingpolicy manages the native ConfigAPI tag-setting policy
// resource. Targeting hierarchy uploads use separate services and audiences.
package targetingpolicy

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var (
	ErrInput    = errors.New("invalid targeting policy input")
	ErrResponse = errors.New("targeting policy returned an unexpected successful HTTP status")
)

// Policy retains optional known settings and the complete received object.
// Strings remain open: baseline metadata is not an exhaustive service enum.
type Policy struct {
	Identity                           string                     `json:"Identity"`
	CustomTagsMode                     *string                    `json:"CustomTagsMode,omitempty"`
	Description                        *string                    `json:"Description,omitempty"`
	ManageTagsPermissionMode           *string                    `json:"ManageTagsPermissionMode,omitempty"`
	ShiftBackedTagsMode                *string                    `json:"ShiftBackedTagsMode,omitempty"`
	SuggestedPresetTags                *string                    `json:"SuggestedPresetTags,omitempty"`
	TeamOwnersEditWhoCanManageTagsMode *string                    `json:"TeamOwnersEditWhoCanManageTagsMode,omitempty"`
	RawFields                          map[string]json.RawMessage `json:"-"`
}

func (p *Policy) UnmarshalJSON(data []byte) error {
	type snapshot Policy
	var value snapshot
	if json.Unmarshal(data, &value) != nil || strings.TrimSpace(value.Identity) == "" || json.Unmarshal(data, &value.RawFields) != nil {
		return client.ErrDecode
	}
	*p = Policy(value)
	return nil
}

// MarshalJSON preserves received omission/null/future fields. A read snapshot
// is not accepted as a mutation request by this service.
func (p Policy) MarshalJSON() ([]byte, error) {
	if p.RawFields != nil {
		return json.Marshal(p.RawFields)
	}
	type snapshot Policy
	return json.Marshal(snapshot(p))
}

// UpdateRequest contains only bound settings. Nil omits a field (the pinned
// mapper also drops explicit null); a pointer to an empty string sends "".
// No setting defaults or closed metadata enums are inferred locally.
type UpdateRequest struct {
	CustomTagsMode                     *string `json:"CustomTagsMode,omitempty"`
	Description                        *string `json:"Description,omitempty"`
	ManageTagsPermissionMode           *string `json:"ManageTagsPermissionMode,omitempty"`
	ShiftBackedTagsMode                *string `json:"ShiftBackedTagsMode,omitempty"`
	SuggestedPresetTags                *string `json:"SuggestedPresetTags,omitempty"`
	TeamOwnersEditWhoCanManageTagsMode *string `json:"TeamOwnersEditWhoCanManageTagsMode,omitempty"`
}
