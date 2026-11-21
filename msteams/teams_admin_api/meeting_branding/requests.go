package meetingbranding

import (
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// Fields distinguishes unbound, explicit null, false and empty scalar values.
type Fields struct {
	DefaultTheme                                 policy.Field[string] `json:"DefaultTheme,omitzero"`
	EnableMeetingOptionsThemeOverride            policy.Field[bool]   `json:"EnableMeetingOptionsThemeOverride,omitzero"`
	EnableMeetingOptionsBrandedReactionsOverride policy.Field[bool]   `json:"EnableMeetingOptionsBrandedReactionsOverride,omitzero"`
	EnableNdiAssuranceSlate                      policy.Field[bool]   `json:"EnableNdiAssuranceSlate,omitzero"`
	EnableMeetingBackgroundImages                policy.Field[bool]   `json:"EnableMeetingBackgroundImages,omitzero"`
	RequireBackgroundEffect                      policy.Field[bool]   `json:"RequireBackgroundEffect,omitzero"`
}

// Settings sends full object arrays. A nil list omits the property; Assets[T]()
// sends []. Nested elements use known public properties, including explicit nulls.
type Settings struct {
	Fields
	NDIAssuranceSlateImages *AssetList[AssuranceSlate]  `json:"NdiAssuranceSlateImages,omitempty"`
	MeetingBackgroundImages *AssetList[BackgroundImage] `json:"MeetingBackgroundImages,omitempty"`
	MeetingBrandingThemes   *AssetList[BrandingTheme]   `json:"MeetingBrandingThemes,omitempty"`
}
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}
type CreateRequest struct {
	Identity string
	Settings
	policy.Controls
}

// UpdateRequest composes Get, bound local edits and one sparse PATCH. It has
// no atomic list operators, ETag claim, automatic replay or rollback.
type UpdateRequest struct {
	Fields
	policy.Controls
	NDIAssuranceSlateImages *AssetEdit[AssuranceSlate]
	MeetingBackgroundImages *AssetEdit[BackgroundImage]
	MeetingBrandingThemes   *AssetEdit[BrandingTheme]
}
type UpdateResult struct {
	Before       *Policy
	After        *Policy
	ReadResponse *resty.Response
	PatchInvoked bool
	HTTPAccepted bool
}
type CreateResult struct {
	Comparison         *Policy
	ComparisonResponse *resty.Response
	ComparisonError    error
	Created            *Policy
	CreateInvoked      bool
	HTTPAccepted       bool
}

func (UpdateResult) String() string     { return "branding update result (redacted)" }
func (v UpdateResult) GoString() string { return v.String() }
func (CreateResult) String() string     { return "branding create result (redacted)" }
func (v CreateResult) GoString() string { return v.String() }
