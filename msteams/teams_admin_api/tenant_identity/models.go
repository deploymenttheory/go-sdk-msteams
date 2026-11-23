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

// Group is one Find-CsGroup match.
type Group struct {
	ID          *string                    `json:"Id"`
	DisplayName *string                    `json:"DisplayName"`
	RawFields   map[string]json.RawMessage `json:"-"`
}

func (v *Group) UnmarshalJSON(raw []byte) error {
	type plain Group
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = Group(value)
	return nil
}

// GroupSearchRequest is Find-CsGroup. SearchQuery is required; the service
// rejects queries shorter than three characters. Nil options are omitted and a
// non-nil boolean is sent explicitly as True or False.
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
	RawFields         map[string]json.RawMessage `json:"-"`
}

func (v *AssignedPlan) UnmarshalJSON(raw []byte) error {
	type plain AssignedPlan
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = AssignedPlan(value)
	return nil
}

// ProvisionedPlan is one provisioned tenant service.
type ProvisionedPlan struct {
	CapabilityStatus   *string                    `json:"capabilityStatus"`
	ProvisioningStatus *string                    `json:"provisioningStatus"`
	Service            *string                    `json:"service"`
	RawFields          map[string]json.RawMessage `json:"-"`
}

func (v *ProvisionedPlan) UnmarshalJSON(raw []byte) error {
	type plain ProvisionedPlan
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = ProvisionedPlan(value)
	return nil
}

// VerifiedDomain is a tenant domain and its verification status.
type VerifiedDomain struct {
	Name      *string                    `json:"name"`
	Status    *string                    `json:"status"`
	RawFields map[string]json.RawMessage `json:"-"`
}

func (v *VerifiedDomain) UnmarshalJSON(raw []byte) error {
	type plain VerifiedDomain
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = VerifiedDomain(value)
	return nil
}

// Tenant is the Get-CsTenant extended tenant record. Timestamps keep their
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
	SIPDomains                       []string                   `json:"sipDomains"`
	TeamsUpgradeEffectiveMode        *string                    `json:"teamsUpgradeEffectiveMode"`
	TeamsUpgradeNotificationsEnabled *bool                      `json:"teamsUpgradeNotificationsEnabled"`
	TeamsUpgradeOverridePolicy       *string                    `json:"teamsUpgradeOverridePolicy"`
	TeamsUpgradePolicyIsReadOnly     *string                    `json:"teamsUpgradePolicyIsReadOnly"`
	LastProvisionTimeStamps          map[string]string          `json:"lastProvisionTimeStamps"`
	LastPublishTimeStamps            map[string]string          `json:"lastPublishTimeStamps"`
	RawFields                        map[string]json.RawMessage `json:"-"`
}

func (v *Tenant) UnmarshalJSON(raw []byte) error {
	type plain Tenant
	var value plain
	if err := decodeFields(raw, &value, &value.RawFields); err != nil {
		return err
	}
	*v = Tenant(value)
	return nil
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
