package unifiedapps

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// Assignment and installation types accepted by Update. Values are compared
// case-insensitively and sent exactly as supplied, as the module does.
const (
	TypeUsersAndGroups = "UsersAndGroups"
	TypeEveryone       = "Everyone"
	TypeNoone          = "Noone"
)

// Membership operations for UsersAndGroups assignments and tenant settings.
const (
	OperationAdd    = "add"
	OperationRemove = "remove"
)

// Review statuses accepted by ReviewCustomPending.
const (
	ReviewPublished = "Published"
	ReviewRejected  = "Rejected"
)

// Tenant setting names and values accepted by the pinned module (case-sensitive).
const (
	SettingDefaultApp                 = "DefaultApp"
	SettingGlobalApp                  = "GlobalApp"
	SettingPrivateApp                 = "PrivateApp"
	SettingEnableCopilotExtensibility = "EnableCopilotExtensibility"

	SettingValueAll  = "All"
	SettingValueNone = "None"
	SettingValueSome = "Some"
)

// ErrInput identifies a request rejected locally before credentials or I/O.
var ErrInput = errors.New("unified app input is invalid")

// ErrResponse identifies an undeclared successful HTTP status.
var ErrResponse = errors.New("unexpected unified app response")

// ErrPendingAppNotFound means the supplied ID is not in the pending custom app list,
// so no review was sent.
var ErrPendingAppNotFound = errors.New("custom pending app was not found")

// ErrMissingIteration means the pending app details carried no staged app
// definition etag, so no review was sent.
var ErrMissingIteration = errors.New("custom pending app has no staged app definition etag")

// decodeFields unmarshals raw into value (a pointer to a plain alias type) and
// keeps every service field in fields, so unknown properties survive.
func decodeFields(raw []byte, value any, fields *map[string]json.RawMessage) error {
	if json.Unmarshal(raw, value) != nil || json.Unmarshal(raw, fields) != nil || *fields == nil {
		return client.ErrDecode
	}
	return nil
}

// App is one catalog app from Get-AllM365TeamsApps or Get-M365TeamsApp. Absent
// and null fields stay nil; timestamps keep their wire text. RawFields keeps the
// complete service object, including any expanded metadata.
type App struct {
	ID           *string                    `json:"id"`
	IsBlocked    *bool                      `json:"isBlocked"`
	AppType      *string                    `json:"appType"`
	AppName      *string                    `json:"appName"`
	AvailableTo  *Availability              `json:"availableTo"`
	InstalledFor *Installation              `json:"deployedTo"`
	RawFields    map[string]json.RawMessage `json:"-"`
}

func (v *App) UnmarshalJSON(raw []byte) error {
	type plain App
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = App(value)
	return nil
}

// Availability says who may use an app. Users and Groups are populated only by
// the single-app read.
type Availability struct {
	AssignmentType  *string                    `json:"availableToType"`
	Users           []Assignment               `json:"users"`
	Groups          []Assignment               `json:"groups"`
	LastUpdatedDate *string                    `json:"lastUpdatedDate"`
	AssignedBy      *string                    `json:"assignedBy"`
	RawFields       map[string]json.RawMessage `json:"-"`
}

func (v *Availability) UnmarshalJSON(raw []byte) error {
	type plain Availability
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = Availability(value)
	return nil
}

// Assignment is one user or group availability entry. The pinned model types
// id and assignedBy as string arrays.
type Assignment struct {
	ID              []string                   `json:"id"`
	LastUpdatedDate *string                    `json:"lastUpdatedDate"`
	AssignedBy      []string                   `json:"assignedBy"`
	RawFields       map[string]json.RawMessage `json:"-"`
}

func (v *Assignment) UnmarshalJSON(raw []byte) error {
	type plain Assignment
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = Assignment(value)
	return nil
}

// Installation says for whom an app is installed (wire name deployedTo).
type Installation struct {
	InstallType     *string                    `json:"deployedToType"`
	LastUpdatedDate *string                    `json:"lastUpdatedDate"`
	InstalledBy     *string                    `json:"deployedBy"`
	InstalledSource *string                    `json:"deployReason"`
	Version         *string                    `json:"version"`
	Users           []Installed                `json:"users"`
	Groups          []Installed                `json:"groups"`
	RawFields       map[string]json.RawMessage `json:"-"`
}

func (v *Installation) UnmarshalJSON(raw []byte) error {
	type plain Installation
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = Installation(value)
	return nil
}

// Installed is one user or group installation entry.
type Installed struct {
	ID              []string                   `json:"id"`
	LastUpdatedDate *string                    `json:"lastUpdatedDate"`
	InstalledBy     []string                   `json:"deployedBy"`
	RawFields       map[string]json.RawMessage `json:"-"`
}

func (v *Installed) UnmarshalJSON(raw []byte) error {
	type plain Installed
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = Installed(value)
	return nil
}

// PendingApp is a custom app awaiting admin review. Iteration is the staged app
// definition etag that a review must quote.
type PendingApp struct {
	ID                 *string                    `json:"id"`
	ExternalID         *string                    `json:"externalId"`
	Iteration          *string                    `json:"iteration"`
	CreatedBy          *string                    `json:"createdBy"`
	LastUpdateDateTime *string                    `json:"lastUpdateDateTime"`
	ReviewStatus       *string                    `json:"reviewStatus"`
	Metadata           *Metadata                  `json:"metadata"`
	RawFields          map[string]json.RawMessage `json:"-"`
}

func (v *PendingApp) UnmarshalJSON(raw []byte) error {
	type plain PendingApp
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = PendingApp(value)
	return nil
}

// Metadata describes a pending app's manifest.
type Metadata struct {
	AppName          *string                    `json:"appName"`
	TermsOfUseURL    *string                    `json:"termsOfUseUrl"`
	LongDescription  *string                    `json:"longDescription"`
	ShortDescription *string                    `json:"shortDescription"`
	LargeImageURL    *string                    `json:"largeImageUrl"`
	SmallImageURL    *string                    `json:"smallImageUrl"`
	DeveloperName    *string                    `json:"developerName"`
	DeveloperURL     *string                    `json:"developerUrl"`
	ManifestVersion  *string                    `json:"manifestVersion"`
	AppVersion       *string                    `json:"appVersion"`
	CopilotEnabled   *bool                      `json:"copilotEnabled"`
	CompatibleHubs   []string                   `json:"compatibleHubs"`
	PublishDateTime  *string                    `json:"publishDateTime"`
	RawFields        map[string]json.RawMessage `json:"-"`
}

func (v *Metadata) UnmarshalJSON(raw []byte) error {
	type plain Metadata
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = Metadata(value)
	return nil
}

// TenantSetting is one unified app tenant setting.
type TenantSetting struct {
	SettingName  *string                    `json:"settingName"`
	SettingValue *string                    `json:"settingValue"`
	Users        []string                   `json:"users"`
	Groups       []string                   `json:"groups"`
	RawFields    map[string]json.RawMessage `json:"-"`
}

func (v *TenantSetting) UnmarshalJSON(raw []byte) error {
	type plain TenantSetting
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = TenantSetting(value)
	return nil
}

// AppUpdate mirrors Update-M365TeamsApp. At least one of IsBlocked,
// AssignmentType or InstallType is required.
//
// AssignmentType UsersAndGroups needs Operation and at least one of Users or
// Groups; Everyone and Noone forbid Users and Groups and always send operation
// "add". The Install* fields follow the same rules for installation.
type AppUpdate struct {
	IsBlocked *bool

	AssignmentType string
	Operation      string
	Users          []string
	Groups         []string

	InstallType      string
	InstallOperation string
	InstallUsers     []string
	InstallGroups    []string
	InstallVersion   string
}

// TenantSettingsUpdate mirrors Update-M365UnifiedTenantSettings. DefaultApp,
// GlobalApp and PrivateApp accept only All or None and no Users, Groups or
// Operation.
type TenantSettingsUpdate struct {
	SettingName  string
	SettingValue string
	Users        []string
	Groups       []string
	Operation    string
}

type membership struct {
	Type      string   `json:"type,omitempty"`
	Users     []string `json:"users,omitempty"`
	Groups    []string `json:"groups,omitempty"`
	Operation string   `json:"operation,omitempty"`
	Version   string   `json:"version,omitempty"`
}

type appUpdateBody struct {
	ID          string      `json:"id"`
	IsBlocked   *bool       `json:"isBlocked,omitempty"`
	AvailableTo *membership `json:"availableTo,omitempty"`
	DeployTo    *membership `json:"deployTo,omitempty"`
}

type settingsBody struct {
	SettingName  string   `json:"settingName,omitempty"`
	SettingValue string   `json:"settingValue,omitempty"`
	Users        []string `json:"users,omitempty"`
	Groups       []string `json:"groups,omitempty"`
	Operation    string   `json:"operation,omitempty"`
}

type reviewBody struct {
	ID                      string `json:"Id"`
	ReviewStatus            string `json:"ReviewStatus"`
	StagedAppDefinitionEtag string `json:"StagedAppDefinitionEtag"`
}

type pendingDetails struct {
	StagedAppEntitlement *struct {
		Iteration *string `json:"iteration"`
	} `json:"stagedAppEntitlement"`
}
