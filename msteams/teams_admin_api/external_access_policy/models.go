// Package externalaccesspolicy manages the compiled ExternalAccessPolicy ConfigAPI resource.
package externalaccesspolicy

import (
	"bytes"
	"encoding/json"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput

// CommunicationWithExternalOrgs preserves the pinned numeric JSON contract.
// Unknown Int32 values remain representable; no closed server enum is inferred.
type CommunicationWithExternalOrgs int32

const (
	OrganizationDefault CommunicationWithExternalOrgs = iota
	AllowAllExternalDomains
	AllowSpecificExternalDomains
	BlockSpecificExternalDomains
	BlockAllExternalDomains
)

// UnmarshalJSON accepts the named live response form as well as the pinned
// serializer's numeric form. Request serialization stays numeric; Policy keeps
// the original response representation in RawFields.
func (v *CommunicationWithExternalOrgs) UnmarshalJSON(data []byte) error {
	var number int32
	if json.Unmarshal(data, &number) == nil {
		*v = CommunicationWithExternalOrgs(number)
		return nil
	}
	var name string
	if json.Unmarshal(data, &name) != nil {
		return client.ErrDecode
	}
	var mode CommunicationWithExternalOrgs
	switch strings.ToLower(name) {
	case "organizationdefault":
		mode = OrganizationDefault
	case "allowallexternaldomains":
		mode = AllowAllExternalDomains
	case "allowspecificexternaldomains":
		mode = AllowSpecificExternalDomains
	case "blockspecificexternaldomains":
		mode = BlockSpecificExternalDomains
	case "blockallexternaldomains":
		mode = BlockAllExternalDomains
	default:
		return client.ErrDecode
	}
	*v = mode
	return nil
}

// Policy retains nullable/omitted/future wire fields. RawFields is authoritative;
// native reads do not materialize defaults from the PowerShell CLR constructor.
type Policy struct {
	Identity                                          *string                        `json:"Identity"`
	Description                                       *string                        `json:"Description"`
	EnableFederationAccess                            *bool                          `json:"EnableFederationAccess"`
	EnableXmppAccess                                  *bool                          `json:"EnableXmppAccess"`
	EnablePublicCloudAudioVideoAccess                 *bool                          `json:"EnablePublicCloudAudioVideoAccess"`
	EnableTeamsSmsAccess                              *bool                          `json:"EnableTeamsSmsAccess"`
	EnableOutsideAccess                               *bool                          `json:"EnableOutsideAccess"`
	EnableAcsFederationAccess                         *bool                          `json:"EnableAcsFederationAccess"`
	EnableTeamsConsumerAccess                         *bool                          `json:"EnableTeamsConsumerAccess"`
	EnableTeamsConsumerInbound                        *bool                          `json:"EnableTeamsConsumerInbound"`
	RestrictTeamsConsumerAccessToExternalUserProfiles *bool                          `json:"RestrictTeamsConsumerAccessToExternalUserProfiles"`
	FederatedBilateralChats                           *bool                          `json:"FederatedBilateralChats"`
	AllowedExternalDomains                            []string                       `json:"AllowedExternalDomains"`
	BlockedExternalDomains                            []string                       `json:"BlockedExternalDomains"`
	CommunicationWithExternalOrgs                     *CommunicationWithExternalOrgs `json:"CommunicationWithExternalOrgs"`
	RawFields                                         map[string]json.RawMessage     `json:"-"`
}

func (v *Policy) UnmarshalJSON(data []byte) error {
	type snapshot Policy
	var value snapshot
	if json.Unmarshal(data, &value) != nil || json.Unmarshal(data, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	if bytes.Equal(bytes.TrimSpace(value.RawFields["AllowedExternalDomains"]), []byte("null")) {
		return client.ErrDecode
	}
	if bytes.Equal(bytes.TrimSpace(value.RawFields["BlockedExternalDomains"]), []byte("null")) {
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
func (Policy) String() string     { return "compiled ExternalAccessPolicy snapshot (redacted)" }
func (v Policy) GoString() string { return v.String() }

// Fields preserves omitted/null/value scalar binding for sparse request bodies.
type Fields struct {
	Description                                       policy.Field[string]                        `json:"Description,omitzero"`
	EnableFederationAccess                            policy.Field[bool]                          `json:"EnableFederationAccess,omitzero"`
	EnableXmppAccess                                  policy.Field[bool]                          `json:"EnableXmppAccess,omitzero"`
	EnablePublicCloudAudioVideoAccess                 policy.Field[bool]                          `json:"EnablePublicCloudAudioVideoAccess,omitzero"`
	EnableTeamsSmsAccess                              policy.Field[bool]                          `json:"EnableTeamsSmsAccess,omitzero"`
	EnableOutsideAccess                               policy.Field[bool]                          `json:"EnableOutsideAccess,omitzero"`
	EnableAcsFederationAccess                         policy.Field[bool]                          `json:"EnableAcsFederationAccess,omitzero"`
	EnableTeamsConsumerAccess                         policy.Field[bool]                          `json:"EnableTeamsConsumerAccess,omitzero"`
	EnableTeamsConsumerInbound                        policy.Field[bool]                          `json:"EnableTeamsConsumerInbound,omitzero"`
	RestrictTeamsConsumerAccessToExternalUserProfiles policy.Field[bool]                          `json:"RestrictTeamsConsumerAccessToExternalUserProfiles,omitzero"`
	FederatedBilateralChats                           policy.Field[bool]                          `json:"FederatedBilateralChats,omitzero"`
	CommunicationWithExternalOrgs                     policy.Field[CommunicationWithExternalOrgs] `json:"CommunicationWithExternalOrgs,omitzero"`
}

// Settings contains fully materialized request arrays. A nil list pointer omits
// the property; policy.Strings() sends an explicit empty array.
type Settings struct {
	Fields
	AllowedExternalDomains *policy.StringList `json:"AllowedExternalDomains,omitempty"`
	BlockedExternalDomains *policy.StringList `json:"BlockedExternalDomains,omitempty"`
}

// PatchRequest is one native sparse PATCH. BodyIdentity defaults to the route
// identity; compositions bind the preceding read model's identity explicitly.
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}

// UpdateRequest composes a read, local list edits and one sparse PATCH. No ETag,
// atomic Add/Remove, automatic retry or rollback is inferred.
type UpdateRequest struct {
	Fields
	policy.Controls
	AllowedExternalDomains *policy.ListEdit
	BlockedExternalDomains *policy.ListEdit
}

// UpdateResult preserves preparation and HTTP acceptance if the later step fails.
// PatchInvoked records the SDK call, not proof of a network send or propagation.
type UpdateResult struct {
	Before       *Policy
	After        *Policy
	ReadResponse *resty.Response
	PatchInvoked bool
	HTTPAccepted bool
}

func (UpdateResult) String() string     { return "compiled policy update result (redacted)" }
func (v UpdateResult) GoString() string { return v.String() }

// CreateRequest describes one native creation; no comparison/preflight read is implicit.
type CreateRequest struct {
	Identity string
	Settings
	policy.Controls
}

// CreateResult separates the optional comparison observation from submission.
// Comparison errors are retained but do not prevent a later explicitly requested Create.
type CreateResult struct {
	Comparison         *Policy
	ComparisonResponse *resty.Response
	ComparisonError    error
	Created            *Policy
	CreateInvoked      bool
	HTTPAccepted       bool
}

func (CreateResult) String() string     { return "compiled policy create result (redacted)" }
func (v CreateResult) GoString() string { return v.String() }
