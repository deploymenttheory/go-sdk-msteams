package tenantfederationconfiguration

import (
	"encoding/json"
	"errors"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/listchange"
)

var ErrResponse = errors.New("tenant federation configuration returned an unexpected successful HTTP status")

// Settings contains nullable scalar settings. Request nil omits a field; a
// pointer to false or an empty string sends that bound value without defaults.
type Settings struct {
	AllowFederatedUsers                                 *bool   `json:"AllowFederatedUsers,omitempty"`
	AllowTeamsConsumer                                  *bool   `json:"AllowTeamsConsumer,omitempty"`
	AllowTeamsConsumerInbound                           *bool   `json:"AllowTeamsConsumerInbound,omitempty"`
	RestrictTeamsConsumerToExternalUserProfiles         *bool   `json:"RestrictTeamsConsumerToExternalUserProfiles,omitempty"`
	SharedSipAddressSpace                               *bool   `json:"SharedSipAddressSpace,omitempty"`
	TreatDiscoveredPartnersAsUnverified                 *bool   `json:"TreatDiscoveredPartnersAsUnverified,omitempty"`
	BlockAllSubdomains                                  *bool   `json:"BlockAllSubdomains,omitempty"`
	EnableExternalAccessRestrictionsForChatParticipants *bool   `json:"EnableExternalAccessRestrictionsForChatParticipants,omitempty"`
	EnableMutualFederationForChatParticipants           *bool   `json:"EnableMutualFederationForChatParticipants,omitempty"`
	ExternalAccessWithTrialTenants                      *string `json:"ExternalAccessWithTrialTenants,omitempty"`
	SecurityTeamAllowBlockListDelegation                *string `json:"SecurityTeamAllowBlockListDelegation,omitempty"`
}

// Configuration retains the raw domain-choice properties because the service's
// response discriminators/null defaults are not established by presentation
// fixups. RawFields preserves every received field, omission and explicit null.
type Configuration struct {
	Identity string `json:"Identity"`
	Settings
	AllowedDomains            json.RawMessage            `json:"AllowedDomains,omitempty"`
	BlockedDomains            json.RawMessage            `json:"BlockedDomains,omitempty"`
	AllowedTrialTenantDomains json.RawMessage            `json:"AllowedTrialTenantDomains,omitempty"`
	RawFields                 map[string]json.RawMessage `json:"-"`
}

func (v *Configuration) UnmarshalJSON(data []byte) error {
	type snapshot Configuration
	var value snapshot
	if json.Unmarshal(data, &value) != nil || strings.TrimSpace(value.Identity) == "" || json.Unmarshal(data, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = Configuration(value)
	return nil
}

func (v Configuration) MarshalJSON() ([]byte, error) {
	if v.RawFields != nil {
		return json.Marshal(v.RawFields)
	}
	type snapshot Configuration
	return json.Marshal(snapshot(v))
}
func (Configuration) String() string     { return "tenant federation configuration (redacted)" }
func (v Configuration) GoString() string { return v.String() }

// ListChange describes the modern helper's replacement or Add/Remove request.
// It is distinct from compiled policy providers' read/modify/write list edits.
type ListChange struct{ change *listchange.Change }

func ReplaceDomains(domains []string) *ListChange {
	return &ListChange{change: listchange.ReplaceStrings(domains)}
}
func ChangeDomains(add, remove []string) *ListChange {
	return &ListChange{change: listchange.DeltaStrings(add, remove)}
}

// ReplaceDomainValues accepts a mixed array of strings/legacy objects, removing
// only known serialization metadata from objects. Local normalization does not
// establish backend acceptance of an arbitrary object as a domain identifier.
func ReplaceDomainValues(values []json.RawMessage) (*ListChange, error) {
	v, err := listchange.ReplaceValues(values)
	if err != nil {
		return nil, ErrInput
	}
	return &ListChange{change: v}, nil
}

func ChangeDomainValues(add, remove []json.RawMessage) (*ListChange, error) {
	v, err := listchange.DeltaValues(add, remove)
	if err != nil {
		return nil, ErrInput
	}
	return &ListChange{change: v}, nil
}

func (v ListChange) MarshalJSON() ([]byte, error) {
	if v.change == nil {
		return nil, ErrInput
	}
	return v.change.MarshalJSON()
}

// UpdateRequest contains only bound properties. AllowedDomains and
// AllowedDomainsAsAList are mutually exclusive; the latter's wire key is also
// AllowedDomains. Nil fields omit values, rather than inventing a null alias.
type UpdateRequest struct {
	Settings
	AllowedDomains            *AllowedDomains
	AllowedDomainsAsAList     *ListChange
	BlockedDomains            *ListChange
	AllowedTrialTenantDomains *ListChange
}

func writeBody(identity string, input UpdateRequest) (map[string]json.RawMessage, error) {
	if input.AllowedDomains != nil && input.AllowedDomainsAsAList != nil {
		return nil, ErrInput
	}
	raw, err := json.Marshal(input.Settings)
	if err != nil {
		return nil, ErrInput
	}
	var body map[string]json.RawMessage
	if json.Unmarshal(raw, &body) != nil {
		return nil, ErrInput
	}
	body["Identity"], err = json.Marshal(identity)
	if err != nil {
		return nil, ErrInput
	}
	if input.AllowedDomains != nil {
		value, err := input.AllowedDomains.MarshalJSON()
		if err != nil {
			return nil, err
		}
		body["AllowedDomains"] = value
	}
	for _, entry := range []struct {
		name  string
		value *ListChange
	}{{"AllowedDomains", input.AllowedDomainsAsAList}, {"BlockedDomains", input.BlockedDomains}, {"AllowedTrialTenantDomains", input.AllowedTrialTenantDomains}} {
		if entry.value == nil {
			continue
		}
		value, err := entry.value.MarshalJSON()
		if err != nil {
			return nil, err
		}
		body[entry.name] = value
	}
	return body, nil
}
