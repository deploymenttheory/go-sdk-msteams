package tenantidentity

import (
	"bytes"
	"encoding/json"
	"strconv"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// The own Tenant constructors use the same pinned JsonNode parser as Group,
// including exact recursive duplicate rejection. Model conversion remains
// separate: there is no Group GUID projection or public-helper date conversion.
func tenantWireFields(raw []byte) (map[string]json.RawMessage, error) {
	fields, err := groupNode(raw)
	if err != nil || fields == nil {
		return nil, client.ErrDecode
	}
	return fields, nil
}
func tenantWireBool(raw json.RawMessage) *bool {
	raw = bytes.TrimSpace(raw)
	if !bytes.Equal(raw, []byte("true")) && !bytes.Equal(raw, []byte("false")) {
		return nil
	}
	value := bytes.Equal(raw, []byte("true"))
	return &value
}
func tenantWireInt32(raw json.RawMessage) (*int32, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || (raw[0] != '-' && (raw[0] < '0' || raw[0] > '9')) {
		return nil, nil
	}
	value, err := strconv.ParseInt(string(raw), 10, 32)
	if err != nil {
		return nil, client.ErrDecode
	}
	result := int32(value)
	return &result, nil
}
func tenantWireArray(raw json.RawMessage) ([]json.RawMessage, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '[' {
		return nil, nil
	}
	var items []json.RawMessage
	if json.Unmarshal(raw, &items) != nil {
		return nil, client.ErrDecode
	}
	return items, nil
}
func tenantWireStrings(raw json.RawMessage) ([]*string, error) {
	items, err := tenantWireArray(raw)
	if err != nil || items == nil {
		return nil, err
	}
	result := make([]*string, len(items))
	for i, item := range items {
		result[i] = groupString(item)
	}
	return result, nil
}
func tenantWireObjects[T any](raw json.RawMessage) ([]*T, error) {
	items, err := tenantWireArray(raw)
	if err != nil || items == nil {
		return nil, err
	}
	result := make([]*T, len(items))
	for i, item := range items {
		item = bytes.TrimSpace(item)
		if len(item) == 0 || item[0] != '{' {
			continue
		}
		var value T
		if json.Unmarshal(item, &value) != nil {
			return nil, client.ErrDecode
		}
		result[i] = &value
	}
	return result, nil
}
func tenantWireMap(raw json.RawMessage) (map[string]json.RawMessage, error) {
	raw = bytes.TrimSpace(raw)
	if len(raw) == 0 || raw[0] != '{' {
		return nil, nil
	}
	var value map[string]json.RawMessage
	if json.Unmarshal(raw, &value) != nil {
		return nil, client.ErrDecode
	}
	return value, nil
}

func (v *CompanyPartnership) UnmarshalJSON(raw []byte) error {
	fields, err := tenantWireFields(raw)
	if err != nil {
		return err
	}
	*v = CompanyPartnership{
		LoggingEnabled:   tenantWireBool(fields["loggingEnabled"]),
		PartnerContextID: groupString(fields["partnerContextId"]),
		PartnerType:      groupString(fields["partnerType"]),
		SupportPartner:   tenantWireBool(fields["supportPartner"]),
		RawFields:        fields}
	return nil
}

func (v *SyncInLyncADInfo) UnmarshalJSON(raw []byte) error {
	fields, err := tenantWireFields(raw)
	if err != nil {
		return err
	}
	*v = SyncInLyncADInfo{
		IsSyncDisabledAtTenantCreation:  tenantWireBool(fields["isSyncDisabledAtTenantCreation"]),
		IsUserSyncDisabled:              tenantWireBool(fields["isUserSyncDisabled"]),
		IsUserSyncStateChanging:         tenantWireBool(fields["isUserSyncStateChanging"]),
		StopSyncRevertCompleteTimestamp: groupString(fields["stopSyncRevertCompleteTimestamp"]),
		StopSyncRevertTimestamp:         groupString(fields["stopSyncRevertTimestamp"]),
		StopSyncTimestamp:               groupString(fields["stopSyncTimestamp"]),
		RawFields:                       fields}
	return nil
}

func (v *AssignedPlan) UnmarshalJSON(raw []byte) error {
	fields, err := tenantWireFields(raw)
	if err != nil {
		return err
	}
	*v = AssignedPlan{
		AssignedTimestamp: groupString(fields["assignedTimestamp"]),
		Capability:        groupString(fields["capability"]),
		CapabilityStatus:  groupString(fields["capabilityStatus"]),
		IsInGracePeriod:   tenantWireBool(fields["isInGracePeriod"]),
		ServicePlanID:     groupString(fields["servicePlanId"]),
		SubscribedPlanID:  groupString(fields["subscribedPlanId"]),
		ServiceInstance:   groupString(fields["serviceInstance"]),
		RawFields:         fields}
	return nil
}

func (v *ProvisionedPlan) UnmarshalJSON(raw []byte) error {
	fields, err := tenantWireFields(raw)
	if err != nil {
		return err
	}
	*v = ProvisionedPlan{
		CapabilityStatus:   groupString(fields["capabilityStatus"]),
		ProvisioningStatus: groupString(fields["provisioningStatus"]),
		Service:            groupString(fields["service"]),
		RawFields:          fields}
	return nil
}

func (v *VerifiedDomain) UnmarshalJSON(raw []byte) error {
	fields, err := tenantWireFields(raw)
	if err != nil {
		return err
	}
	*v = VerifiedDomain{
		Name:      groupString(fields["name"]),
		Status:    groupString(fields["status"]),
		RawFields: fields}
	return nil
}

func (v *Tenant) UnmarshalJSON(raw []byte) error {
	fields, err := tenantWireFields(raw)
	if err != nil {
		return err
	}
	var value Tenant
	value.ObjectID = groupString(fields["objectId"])
	value.DisplayName = groupString(fields["displayName"])
	value.City = groupString(fields["city"])
	value.Street = groupString(fields["street"])
	value.PostalCode = groupString(fields["postalCode"])
	value.CountryLetterCode = groupString(fields["countryLetterCode"])
	value.PreferredLanguage = groupString(fields["preferredLanguage"])
	value.CreatedDateTime = groupString(fields["createdDateTime"])
	value.WhenCreated = groupString(fields["whenCreated"])
	value.LastSyncTimeStamp = groupString(fields["lastSyncTimeStamp"])
	value.ServiceInstance = groupString(fields["serviceInstance"])
	value.ServiceInfo = groupString(fields["serviceInfo"])
	value.AssignedPlans, err = tenantWireObjects[AssignedPlan](fields["assignedPlans"])
	if err != nil {
		return err
	}
	value.ProvisionedPlans, err = tenantWireObjects[ProvisionedPlan](fields["provisionedPlans"])
	if err != nil {
		return err
	}
	value.VerifiedDomains, err = tenantWireObjects[VerifiedDomain](fields["verifiedDomains"])
	if err != nil {
		return err
	}
	value.SIPDomains, err = tenantWireStrings(fields["sipDomains"])
	if err != nil {
		return err
	}
	value.TeamsUpgradeEffectiveMode = groupString(fields["teamsUpgradeEffectiveMode"])
	value.TeamsUpgradeNotificationsEnabled = tenantWireBool(fields["teamsUpgradeNotificationsEnabled"])
	value.TeamsUpgradeOverridePolicy = groupString(fields["teamsUpgradeOverridePolicy"])
	value.TeamsUpgradePolicyIsReadOnly = groupString(fields["teamsUpgradePolicyIsReadOnly"])
	value.LastProvisionTimeStamps, err = tenantWireMap(fields["lastProvisionTimeStamps"])
	if err != nil {
		return err
	}
	value.LastPublishTimeStamps, err = tenantWireMap(fields["lastPublishTimeStamps"])
	if err != nil {
		return err
	}
	value.AnnouncementsDisabled = tenantWireBool(fields["announcementsDisabled"])
	value.CompanyPartnerships, err = tenantWireObjects[CompanyPartnership](fields["companyPartnership"])
	if err != nil {
		return err
	}
	value.CompanyTags, err = tenantWireStrings(fields["companyTags"])
	if err != nil {
		return err
	}
	value.Country = groupString(fields["country"])
	value.DefaultPoolFQDN = groupString(fields["defaultPoolFqdn"])
	value.DirSyncEnabled = tenantWireBool(fields["dirSyncEnabled"])
	value.NameRecordingDisabled = tenantWireBool(fields["nameRecordingDisabled"])
	value.Pools, err = tenantWireStrings(fields["pools"])
	if err != nil {
		return err
	}
	value.ServiceNumberCount, err = tenantWireInt32(fields["serviceNumberCount"])
	if err != nil {
		return err
	}
	value.State = groupString(fields["state"])
	value.SubscriberNumberCount, err = tenantWireInt32(fields["subscriberNumberCount"])
	if err != nil {
		return err
	}
	if data := bytes.TrimSpace(fields["syncInLyncADInfo"]); len(data) > 0 && data[0] == '{' {
		if json.Unmarshal(data, &value.SyncInLyncADInfo) != nil {
			return client.ErrDecode
		}
	}
	value.TNMAccountID = groupString(fields["tnmAccountId"])
	value.WhenChanged = groupString(fields["whenChanged"])
	value.RawFields = fields
	*v = value
	return nil
}
