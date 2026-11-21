// Package meetingbranding models compiled meeting branding policies and their
// local nested assets. Assets contain URI values; they do not upload image bytes.
package meetingbranding

import (
	"encoding/json"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// BackgroundImage retains received fields without manufacturing PowerShell model defaults.
type BackgroundImage struct {
	Identity   *string                    `json:"Identity"`
	ID         *string                    `json:"Id"`
	Order      *int64                     `json:"Order"`
	Name       *string                    `json:"Name"`
	ImageURI   *string                    `json:"ImageUri"`
	IsRequired *bool                      `json:"IsRequired"`
	IsHidden   *bool                      `json:"IsHidden"`
	RawFields  map[string]json.RawMessage `json:"-"`
}

func (v *BackgroundImage) UnmarshalJSON(data []byte) error {
	type snapshot BackgroundImage
	var value snapshot
	if json.Unmarshal(data, &value) != nil || json.Unmarshal(data, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = BackgroundImage(value)
	return nil
}
func (v BackgroundImage) MarshalJSON() ([]byte, error) {
	if v.RawFields != nil {
		return json.Marshal(v.RawFields)
	}
	type snapshot BackgroundImage
	return json.Marshal(snapshot(v))
}
func (BackgroundImage) String() string     { return "meeting branding snapshot (redacted)" }
func (v BackgroundImage) GoString() string { return v.String() }

// AssuranceSlate retains received fields without manufacturing PowerShell model defaults.
type AssuranceSlate struct {
	Identity    *string                    `json:"Identity"`
	ID          *string                    `json:"Id"`
	Name        *string                    `json:"Name"`
	NDIImageURI *string                    `json:"NdiImageUri"`
	IsDefault   *bool                      `json:"IsDefault"`
	RawFields   map[string]json.RawMessage `json:"-"`
}

func (v *AssuranceSlate) UnmarshalJSON(data []byte) error {
	type snapshot AssuranceSlate
	var value snapshot
	if json.Unmarshal(data, &value) != nil || json.Unmarshal(data, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = AssuranceSlate(value)
	return nil
}
func (v AssuranceSlate) MarshalJSON() ([]byte, error) {
	if v.RawFields != nil {
		return json.Marshal(v.RawFields)
	}
	type snapshot AssuranceSlate
	return json.Marshal(snapshot(v))
}
func (AssuranceSlate) String() string     { return "meeting branding snapshot (redacted)" }
func (v AssuranceSlate) GoString() string { return v.String() }

// Reaction retains received fields without manufacturing PowerShell model defaults.
type Reaction struct {
	Identity               *string                    `json:"Identity"`
	ID                     *string                    `json:"Id"`
	ReactionImageURL       *string                    `json:"ReactionImageUrl"`
	ReactionSpriteImageURL *string                    `json:"ReactionSpriteImageUrl"`
	ReactionType           *string                    `json:"ReactionType"`
	BrandedDisplayName     *string                    `json:"BrandedDisplayName"`
	RawFields              map[string]json.RawMessage `json:"-"`
}

func (v *Reaction) UnmarshalJSON(data []byte) error {
	type snapshot Reaction
	var value snapshot
	if json.Unmarshal(data, &value) != nil || json.Unmarshal(data, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = Reaction(value)
	return nil
}
func (v Reaction) MarshalJSON() ([]byte, error) {
	if v.RawFields != nil {
		return json.Marshal(v.RawFields)
	}
	type snapshot Reaction
	return json.Marshal(snapshot(v))
}
func (Reaction) String() string     { return "meeting branding snapshot (redacted)" }
func (v Reaction) GoString() string { return v.String() }

// BrandingTheme retains received fields without manufacturing PowerShell model defaults.
type BrandingTheme struct {
	Identity                              *string                    `json:"Identity"`
	DisplayName                           *string                    `json:"DisplayName"`
	LogoImageLightURI                     *string                    `json:"LogoImageLightUri"`
	LogoImageDarkURI                      *string                    `json:"LogoImageDarkUri"`
	BackgroundImageLightURI               *string                    `json:"BackgroundImageLightUri"`
	BackgroundImageDarkURI                *string                    `json:"BackgroundImageDarkUri"`
	ChatBackgroundImageLightURI           *string                    `json:"ChatBackgroundImageLightUri"`
	ChatBackgroundImageDarkURI            *string                    `json:"ChatBackgroundImageDarkUri"`
	ChatBackgroundImageAlignment          *string                    `json:"ChatBackgroundImageAlignment"`
	LogoImageLightPreAuthURI              *string                    `json:"LogoImageLightPreAuthUri"`
	LogoImageDarkPreAuthURI               *string                    `json:"LogoImageDarkPreAuthUri"`
	MeetingInviteLogoImageLightPreAuthURI *string                    `json:"MeetingInviteLogoImageLightPreAuthUri"`
	MeetingInviteLogoImageDarkPreAuthURI  *string                    `json:"MeetingInviteLogoImageDarkPreAuthUri"`
	BackgroundImageLightPreAuthURI        *string                    `json:"BackgroundImageLightPreAuthUri"`
	BackgroundImageDarkPreAuthURI         *string                    `json:"BackgroundImageDarkPreAuthUri"`
	BrandAccentColor                      *string                    `json:"BrandAccentColor"`
	EnableChatBackground                  *bool                      `json:"EnableChatBackground"`
	DisableOriginalReactions              *bool                      `json:"DisableOriginalReactions"`
	Enabled                               *bool                      `json:"Enabled"`
	MeetingReactions                      []*Reaction                `json:"MeetingReactions"`
	RawFields                             map[string]json.RawMessage `json:"-"`
}

func (v *BrandingTheme) UnmarshalJSON(data []byte) error {
	type snapshot BrandingTheme
	var value snapshot
	if json.Unmarshal(data, &value) != nil || json.Unmarshal(data, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = BrandingTheme(value)
	return nil
}
func (v BrandingTheme) MarshalJSON() ([]byte, error) {
	if v.RawFields != nil {
		return json.Marshal(v.RawFields)
	}
	type snapshot BrandingTheme
	return json.Marshal(snapshot(v))
}
func (BrandingTheme) String() string     { return "meeting branding snapshot (redacted)" }
func (v BrandingTheme) GoString() string { return v.String() }

// Policy retains received fields without manufacturing PowerShell model defaults.
type Policy struct {
	Identity                                     *string                    `json:"Identity"`
	DefaultTheme                                 *string                    `json:"DefaultTheme"`
	EnableMeetingOptionsThemeOverride            *bool                      `json:"EnableMeetingOptionsThemeOverride"`
	EnableMeetingOptionsBrandedReactionsOverride *bool                      `json:"EnableMeetingOptionsBrandedReactionsOverride"`
	EnableNdiAssuranceSlate                      *bool                      `json:"EnableNdiAssuranceSlate"`
	EnableMeetingBackgroundImages                *bool                      `json:"EnableMeetingBackgroundImages"`
	RequireBackgroundEffect                      *bool                      `json:"RequireBackgroundEffect"`
	NDIAssuranceSlateImages                      []*AssuranceSlate          `json:"NdiAssuranceSlateImages"`
	MeetingBackgroundImages                      []*BackgroundImage         `json:"MeetingBackgroundImages"`
	MeetingBrandingThemes                        []*BrandingTheme           `json:"MeetingBrandingThemes"`
	RawFields                                    map[string]json.RawMessage `json:"-"`
}

func (v *Policy) UnmarshalJSON(data []byte) error {
	type snapshot Policy
	var value snapshot
	if json.Unmarshal(data, &value) != nil || json.Unmarshal(data, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = Policy(value)
	return nil
}
func (v Policy) MarshalJSON() ([]byte, error) {
	if v.RawFields != nil {
		return json.Marshal(v.RawFields)
	}
	type snapshot Policy
	return json.Marshal(snapshot(v))
}
func (Policy) String() string     { return "meeting branding snapshot (redacted)" }
func (v Policy) GoString() string { return v.String() }
