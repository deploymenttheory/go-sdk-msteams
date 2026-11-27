package customappsettings

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

var (
	// ErrResponse identifies an undeclared successful HTTP status.
	ErrResponse = errors.New("unexpected custom app settings response")
	// ErrInput identifies a request the SDK rejects before sending it.
	ErrInput = errors.New("invalid custom app settings request")
)

// Settings is the tenant-wide Teams apps settings document read from
// /Teams.MiddletierService/tenantWideAppsSettingsGlobal. The service sends
// camelCase names (lobTextColor), and every field is a pointer so an absent or
// null value stays distinguishable from false/empty. RawFields keeps the
// complete object, including appSettingsList.
type Settings struct {
	IsAppsEnabled                      *bool                      `json:"isAppsEnabled"`
	IsAppsPurchaseEnabled              *bool                      `json:"isAppsPurchaseEnabled"`
	IsExternalAppsEnabledByDefault     *bool                      `json:"isExternalAppsEnabledByDefault"`
	IsLicenseBasedPinnedAppsEnabled    *bool                      `json:"isLicenseBasedPinnedAppsEnabled"`
	IsSideloadedAppsInteractionEnabled *bool                      `json:"isSideloadedAppsInteractionEnabled"`
	IsTenantWideAutoInstallEnabled     *bool                      `json:"isTenantWideAutoInstallEnabled"`
	LobTextColor                       *string                    `json:"lobTextColor"`
	LobBackground                      *string                    `json:"lobBackground"`
	LobLogo                            *string                    `json:"lobLogo"`
	LobLogomark                        *string                    `json:"lobLogomark"`
	AppSettingsList                    []*AppSetting              `json:"appSettingsList"`
	AppAccessRequestConfig             *AppAccessRequestConfig    `json:"appAccessRequestConfig"`
	RawFields                          map[string]json.RawMessage `json:"-"`
}

// SideloadedAppsInteractionEnabled is the Get-CsTeamsSettingsCustomApp
// projection: the cmdlet reports a missing/null value as false.
func (v *Settings) SideloadedAppsInteractionEnabled() bool {
	if v == nil {
		return false
	}
	if v.RawFields != nil {
		return projectSideloadedFlag(v.RawFields)
	}
	return v.IsSideloadedAppsInteractionEnabled != nil && *v.IsSideloadedAppsInteractionEnabled
}

func projectSideloadedFlag(fields map[string]json.RawMessage) bool {
	raw, exists := fields["isSideloadedAppsInteractionEnabled"]
	if !exists {
		return false
	}
	var flag *bool
	return json.Unmarshal(raw, &flag) == nil && flag != nil && *flag
}

func (v *Settings) UnmarshalJSON(raw []byte) error {
	type plain Settings
	var value plain
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	for _, item := range value.AppSettingsList {
		if item == nil {
			return client.ErrDecode
		}
	}
	*v = Settings(value)
	return nil
}

// AppSetting is one per-app entry of appSettingsList. The module's generated
// response model does not declare the list; the service returns it.
type AppSetting struct {
	ID                           *string                    `json:"id"`
	IsEnabled                    *bool                      `json:"isEnabled"`
	IsAppLevelAutoInstallEnabled *bool                      `json:"isAppLevelAutoInstallEnabled"`
	Context                      *string                    `json:"context"`
	RawFields                    map[string]json.RawMessage `json:"-"`
}

func (v *AppSetting) UnmarshalJSON(raw []byte) error {
	type plain AppSetting
	var value plain
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = AppSetting(value)
	return nil
}

// AppAccessRequestConfig controls how users request access to blocked apps.
type AppAccessRequestConfig struct {
	AdminInstructionMessage *string                    `json:"adminInstructionMessage,omitempty"`
	ApprovalPortalURL       *string                    `json:"approvalPortalUrl,omitempty"`
	RawFields               map[string]json.RawMessage `json:"-"`
}

func (v *AppAccessRequestConfig) UnmarshalJSON(raw []byte) error {
	type plain AppAccessRequestConfig
	var value plain
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = AppAccessRequestConfig(value)
	return nil
}

// MarshalJSON writes the declared fields and keeps unknown properties.
func (v AppAccessRequestConfig) MarshalJSON() ([]byte, error) {
	out := make(map[string]any, len(v.RawFields)+2)
	for key, value := range v.RawFields {
		out[key] = value
	}
	if v.AdminInstructionMessage != nil {
		out["adminInstructionMessage"] = *v.AdminInstructionMessage
	}
	if v.ApprovalPortalURL != nil {
		out["approvalPortalUrl"] = *v.ApprovalPortalURL
	}
	return json.Marshal(out)
}

// UpdateRequest is the PUT body (the module's UpdateCustomAppSettingRequest).
// Nil fields are omitted. The service replaces the document, so build it from
// a fresh Get. AppSettingsList is sent as given; an explicit empty list is what
// Set-CsTeamsSettingsCustomApp sends, and the lab showed it leaves the stored
// per-app entries unchanged. Lob* keys use the module's PascalCase names.
type UpdateRequest struct {
	IsAppsEnabled                      *bool                   `json:"isAppsEnabled,omitempty"`
	IsAppsPurchaseEnabled              *bool                   `json:"isAppsPurchaseEnabled,omitempty"`
	IsExternalAppsEnabledByDefault     *bool                   `json:"isExternalAppsEnabledByDefault,omitempty"`
	IsLicenseBasedPinnedAppsEnabled    *bool                   `json:"isLicenseBasedPinnedAppsEnabled,omitempty"`
	IsSideloadedAppsInteractionEnabled *bool                   `json:"isSideloadedAppsInteractionEnabled,omitempty"`
	IsTenantWideAutoInstallEnabled     *bool                   `json:"isTenantWideAutoInstallEnabled,omitempty"`
	LobTextColor                       *string                 `json:"LobTextColor,omitempty"`
	LobBackground                      *string                 `json:"LobBackground,omitempty"`
	LobLogo                            *string                 `json:"LobLogo,omitempty"`
	LobLogomark                        *string                 `json:"LobLogomark,omitempty"`
	AppSettingsList                    []AppSettingUpdate      `json:"appSettingsList"`
	AppAccessRequestConfig             *AppAccessRequestConfig `json:"appAccessRequestConfig,omitempty"`
}

// AppSettingUpdate is one appSettingsList entry in an UpdateRequest.
type AppSettingUpdate struct {
	ID                           string  `json:"id"`
	IsEnabled                    *bool   `json:"isEnabled,omitempty"`
	IsAppLevelAutoInstallEnabled *bool   `json:"isAppLevelAutoInstallEnabled,omitempty"`
	Context                      *string `json:"context,omitempty"`
}
