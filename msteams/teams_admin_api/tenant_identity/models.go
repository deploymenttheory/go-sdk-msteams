package tenantidentity

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// ErrInput identifies a request rejected locally before discovery or credentials.
var ErrInput = errors.New("tenant identity input is invalid")

// ErrResponse identifies an undeclared successful HTTP status.
var ErrResponse = errors.New("unexpected tenant identity response")

// ErrOnlineDirectoryTenantDeprecated is the module's local outcome for
// Get-CsOnlineDirectoryTenant, which dispatches to Invoke-CsDeprecatedError and
// never contacts the service.
var ErrOnlineDirectoryTenantDeprecated = errors.New("Get-CsOnlineDirectoryTenant is no longer supported; use GetTenant and the dial-in conferencing bridge instead")

func decodeFields(raw []byte, value any, fields *map[string]json.RawMessage) error {
	if json.Unmarshal(raw, value) != nil || json.Unmarshal(raw, fields) != nil || *fields == nil {
		return client.ErrDecode
	}
	return nil
}

// Group is one source-projected Find-CsGroup match. ID uses canonical D-format
// GUID text; RawFields retains the original wire properties independently.
type Group struct {
	ID          *string                    `json:"Id"`
	DisplayName *string                    `json:"DisplayName"`
	RawFields   map[string]json.RawMessage `json:"-"`
}

func (v *Group) UnmarshalJSON(raw []byte) error {
	fields, err := groupNode(raw)
	if err != nil || fields == nil {
		return client.ErrDecode
	}
	id := groupString(fields["Id"])
	if id == nil {
		return client.ErrDecode
	}
	canonical, ok := groupGUID(*id)
	if !ok {
		return client.ErrDecode
	}
	*v = Group{ID: &canonical, DisplayName: groupString(fields["DisplayName"]), RawFields: fields}
	return nil
}

// GroupSearchRequest is Find-CsGroup. SearchQuery is required; the service
// rejects queries shorter than three characters. MaxResults above MaxInt32
// cannot bind to the forwarded Int32 parameter and fails locally. Nil options
// are omitted; non-nil zero/false values are sent explicitly.
type GroupSearchRequest struct {
	SearchQuery     string
	MaxResults      *uint32
	ExactMatchOnly  *bool
	MailEnabledOnly *bool
}

// AssignedPlan is one tenant service plan assignment.
type AssignedPlan struct {
	AssignedTimestamp *string                    `json:"assignedTimestamp"`
	Capability        *string                    `json:"capability"`
	CapabilityStatus  *string                    `json:"capabilityStatus"`
	IsInGracePeriod   *bool                      `json:"isInGracePeriod"`
	ServicePlanID     *string                    `json:"servicePlanId"`
	SubscribedPlanID  *string                    `json:"subscribedPlanId"`
	ServiceInstance   *string                    `json:"serviceInstance"`
	RawFields         map[string]json.RawMessage `json:"-"`
}

// ProvisionedPlan is one provisioned tenant service.
type ProvisionedPlan struct {
	CapabilityStatus   *string                    `json:"capabilityStatus"`
	ProvisioningStatus *string                    `json:"provisioningStatus"`
	Service            *string                    `json:"service"`
	RawFields          map[string]json.RawMessage `json:"-"`
}

// VerifiedDomain is a tenant domain and its verification status.
type VerifiedDomain struct {
	Name      *string                    `json:"name"`
	Status    *string                    `json:"status"`
	RawFields map[string]json.RawMessage `json:"-"`
}

// CompanyPartnership is the pinned tenant partnership object.
type CompanyPartnership struct {
	LoggingEnabled   *bool                      `json:"loggingEnabled"`
	PartnerContextID *string                    `json:"partnerContextId"`
	PartnerType      *string                    `json:"partnerType"`
	SupportPartner   *bool                      `json:"supportPartner"`
	RawFields        map[string]json.RawMessage `json:"-"`
}

// SyncInLyncADInfo keeps known settings and original associative properties.
// Timestamp strings remain the wire view, not CLR DateTime projections.
type SyncInLyncADInfo struct {
	IsSyncDisabledAtTenantCreation  *bool                      `json:"isSyncDisabledAtTenantCreation"`
	IsUserSyncDisabled              *bool                      `json:"isUserSyncDisabled"`
	IsUserSyncStateChanging         *bool                      `json:"isUserSyncStateChanging"`
	StopSyncRevertCompleteTimestamp *string                    `json:"stopSyncRevertCompleteTimestamp"`
	StopSyncRevertTimestamp         *string                    `json:"stopSyncRevertTimestamp"`
	StopSyncTimestamp               *string                    `json:"stopSyncTimestamp"`
	RawFields                       map[string]json.RawMessage `json:"-"`
}

// Tenant is the native full-wire view of the extended tenant record. Timestamps keep their
// service text; LastProvisionTimeStamps and LastPublishTimeStamps are keyed by
// processor name. RawFields keeps every service property, including ones the
// PowerShell projection renames or drops.
type Tenant struct {
	ObjectID                         *string                    `json:"objectId"`
	DisplayName                      *string                    `json:"displayName"`
	City                             *string                    `json:"city"`
	Street                           *string                    `json:"street"`
	PostalCode                       *string                    `json:"postalCode"`
	CountryLetterCode                *string                    `json:"countryLetterCode"`
	PreferredLanguage                *string                    `json:"preferredLanguage"`
	CreatedDateTime                  *string                    `json:"createdDateTime"`
	WhenCreated                      *string                    `json:"whenCreated"`
	LastSyncTimeStamp                *string                    `json:"lastSyncTimeStamp"`
	ServiceInstance                  *string                    `json:"serviceInstance"`
	ServiceInfo                      *string                    `json:"serviceInfo"`
	AssignedPlans                    []*AssignedPlan            `json:"assignedPlans"`
	ProvisionedPlans                 []*ProvisionedPlan         `json:"provisionedPlans"`
	VerifiedDomains                  []*VerifiedDomain          `json:"verifiedDomains"`
	SIPDomains                       []*string                  `json:"sipDomains"`
	TeamsUpgradeEffectiveMode        *string                    `json:"teamsUpgradeEffectiveMode"`
	TeamsUpgradeNotificationsEnabled *bool                      `json:"teamsUpgradeNotificationsEnabled"`
	TeamsUpgradeOverridePolicy       *string                    `json:"teamsUpgradeOverridePolicy"`
	TeamsUpgradePolicyIsReadOnly     *string                    `json:"teamsUpgradePolicyIsReadOnly"`
	LastProvisionTimeStamps          map[string]json.RawMessage `json:"lastProvisionTimeStamps"`
	LastPublishTimeStamps            map[string]json.RawMessage `json:"lastPublishTimeStamps"`
	AnnouncementsDisabled            *bool                      `json:"announcementsDisabled"`
	CompanyPartnerships              []*CompanyPartnership      `json:"companyPartnership"`
	CompanyTags                      []*string                  `json:"companyTags"`
	Country                          *string                    `json:"country"`
	DefaultPoolFQDN                  *string                    `json:"defaultPoolFqdn"`
	DirSyncEnabled                   *bool                      `json:"dirSyncEnabled"`
	NameRecordingDisabled            *bool                      `json:"nameRecordingDisabled"`
	Pools                            []*string                  `json:"pools"`
	ServiceNumberCount               *int32                     `json:"serviceNumberCount"`
	State                            *string                    `json:"state"`
	SubscriberNumberCount            *int32                     `json:"subscriberNumberCount"`
	SyncInLyncADInfo                 *SyncInLyncADInfo          `json:"syncInLyncADInfo"`
	TNMAccountID                     *string                    `json:"tnmAccountId"`
	WhenChanged                      *string                    `json:"whenChanged"`
	RawFields                        map[string]json.RawMessage `json:"-"`
}

// DirectoryObject is a directory tenant or user record returned by the
// support cmdlets (Get-CsAadTenant/User, Get-CsCloudTenant/User). The module
// returns their properties without a fixed model, so only common identifiers
// are typed; RawFields keeps the complete object.
type DirectoryObject struct {
	ObjectID    *string                    `json:"objectId"`
	DisplayName *string                    `json:"displayName"`
	RawFields   map[string]json.RawMessage `json:"-"`
}

func (v *DirectoryObject) UnmarshalJSON(raw []byte) error {
	type plain DirectoryObject
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = DirectoryObject(value)
	return nil
}

// Configuration is one hosting provider or tenant licensing configuration as
// returned by the modern configuration read. Identity is typed; RawFields
// keeps every setting.
type Configuration struct {
	Identity  *string                    `json:"Identity"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (v *Configuration) UnmarshalJSON(raw []byte) error {
	type plain Configuration
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = Configuration(value)
	return nil
}

// GracePeriodRemoval is the Remove-CsUserLicenseGracePeriod body. Action is
// sent as given when non-nil. Capabilities nil omits the property; a non-nil
// empty slice sends [] (the module allows an empty collection).
type GracePeriodRemoval struct {
	Action       *string
	Capabilities []string
}
