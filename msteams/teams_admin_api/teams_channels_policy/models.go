// Package teamschannelspolicy manages channel policies and scoped assignments.
package teamschannelspolicy

import (
	"encoding/json"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

var ErrInput = policy.ErrInput
var ErrResponse = errors.New("unexpected channels policy assignment response")

// Policy preserves nullable known fields and an owned complete received snapshot.
// Boolean permissions and string creation modes are independently represented.
type Policy struct {
	Identity                                      *string                    `json:"Identity"`
	Description                                   *string                    `json:"Description"`
	AllowOrgWideTeamCreation                      *bool                      `json:"AllowOrgWideTeamCreation"`
	EnablePrivateTeamDiscovery                    *bool                      `json:"EnablePrivateTeamDiscovery"`
	AllowPrivateChannelCreation                   *bool                      `json:"AllowPrivateChannelCreation"`
	AllowSharedChannelCreation                    *bool                      `json:"AllowSharedChannelCreation"`
	AllowChannelSharingToExternalUser             *bool                      `json:"AllowChannelSharingToExternalUser"`
	AllowUserToParticipateInExternalSharedChannel *bool                      `json:"AllowUserToParticipateInExternalSharedChannel"`
	ThreadedChannelCreation                       *string                    `json:"ThreadedChannelCreation"`
	AllowCreateChannel                            *string                    `json:"AllowCreateChannel"`
	AllowCreateClassicChannel                     *string                    `json:"AllowCreateClassicChannel"`
	AllowCreatePrivateChannel                     *string                    `json:"AllowCreatePrivateChannel"`
	AllowCreateSharedChannel                      *string                    `json:"AllowCreateSharedChannel"`
	CreateSharedChannelsByDefault                 *string                    `json:"CreateSharedChannelsByDefault"`
	AllowUsersFromOutsideTeam                     *string                    `json:"AllowUsersFromOutsideTeam"`
	AllowGuestsFromOutsideTeam                    *string                    `json:"AllowGuestsFromOutsideTeam"`
	AllowUsersFromOutsideTeamInPrivateChannel     *string                    `json:"AllowUsersFromOutsideTeamInPrivateChannel"`
	AllowGuestsFromOutsideTeamInPrivateChannel    *string                    `json:"AllowGuestsFromOutsideTeamInPrivateChannel"`
	AllowSharingWithTeamInOrg                     *string                    `json:"AllowSharingWithTeamInOrg"`
	AllowSharingPrivateChannelWithTeamInOrg       *string                    `json:"AllowSharingPrivateChannelWithTeamInOrg"`
	DefaultChannelTypeOnCreation                  *string                    `json:"DefaultChannelTypeOnCreation"`
	RawFields                                     map[string]json.RawMessage `json:"-"`
	RawJSON                                       json.RawMessage            `json:"-"`
}

func (p *Policy) UnmarshalJSON(raw []byte) error {
	type snapshot Policy
	var next snapshot
	if json.Unmarshal(raw, &next) != nil || json.Unmarshal(raw, &next.RawFields) != nil || next.RawFields == nil {
		return client.ErrDecode
	}
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*p = Policy(next)
	return nil
}
func (p Policy) MarshalJSON() ([]byte, error) {
	if p.RawJSON != nil {
		return append([]byte(nil), p.RawJSON...), nil
	}
	type snapshot Policy
	return json.Marshal(snapshot(p))
}
func (Policy) String() string     { return "channels policy (redacted)" }
func (p Policy) GoString() string { return p.String() }

// Settings distinguishes omission, explicit null and bound empty/false values.
// The pinned setters do not impose string enums or cross-field restrictions.
type Settings struct {
	Description                                   policy.Field[string] `json:"Description,omitzero"`
	AllowOrgWideTeamCreation                      policy.Field[bool]   `json:"AllowOrgWideTeamCreation,omitzero"`
	EnablePrivateTeamDiscovery                    policy.Field[bool]   `json:"EnablePrivateTeamDiscovery,omitzero"`
	AllowPrivateChannelCreation                   policy.Field[bool]   `json:"AllowPrivateChannelCreation,omitzero"`
	AllowSharedChannelCreation                    policy.Field[bool]   `json:"AllowSharedChannelCreation,omitzero"`
	AllowChannelSharingToExternalUser             policy.Field[bool]   `json:"AllowChannelSharingToExternalUser,omitzero"`
	AllowUserToParticipateInExternalSharedChannel policy.Field[bool]   `json:"AllowUserToParticipateInExternalSharedChannel,omitzero"`
	ThreadedChannelCreation                       policy.Field[string] `json:"ThreadedChannelCreation,omitzero"`
	AllowCreateChannel                            policy.Field[string] `json:"AllowCreateChannel,omitzero"`
	AllowCreateClassicChannel                     policy.Field[string] `json:"AllowCreateClassicChannel,omitzero"`
	AllowCreatePrivateChannel                     policy.Field[string] `json:"AllowCreatePrivateChannel,omitzero"`
	AllowCreateSharedChannel                      policy.Field[string] `json:"AllowCreateSharedChannel,omitzero"`
	CreateSharedChannelsByDefault                 policy.Field[string] `json:"CreateSharedChannelsByDefault,omitzero"`
	AllowUsersFromOutsideTeam                     policy.Field[string] `json:"AllowUsersFromOutsideTeam,omitzero"`
	AllowGuestsFromOutsideTeam                    policy.Field[string] `json:"AllowGuestsFromOutsideTeam,omitzero"`
	AllowUsersFromOutsideTeamInPrivateChannel     policy.Field[string] `json:"AllowUsersFromOutsideTeamInPrivateChannel,omitzero"`
	AllowGuestsFromOutsideTeamInPrivateChannel    policy.Field[string] `json:"AllowGuestsFromOutsideTeamInPrivateChannel,omitzero"`
	AllowSharingWithTeamInOrg                     policy.Field[string] `json:"AllowSharingWithTeamInOrg,omitzero"`
	AllowSharingPrivateChannelWithTeamInOrg       policy.Field[string] `json:"AllowSharingPrivateChannelWithTeamInOrg,omitzero"`
	DefaultChannelTypeOnCreation                  policy.Field[string] `json:"DefaultChannelTypeOnCreation,omitzero"`
}
type CreateRequest struct {
	Settings
	policy.Controls
}
type PatchRequest struct {
	Settings
	policy.Controls
	BodyIdentity policy.Field[string]
}
type UpdateRequest struct {
	Settings
	policy.Controls
}

// UpdateResult retains preparation and HTTP acceptance independently on failure.
type UpdateResult struct {
	Before, After              *Policy
	ReadResponse               *resty.Response
	PatchInvoked, HTTPAccepted bool
}

func (UpdateResult) String() string     { return "channels policy update result (redacted)" }
func (p UpdateResult) GoString() string { return p.String() }

// GroupAssignment retains HTTP acceptance and future operation payload fields.
type GroupAssignment struct {
	OperationID string          `json:"OperationId"`
	Accepted    bool            `json:"-"`
	RawJSON     json.RawMessage `json:"-"`
}

func (p GroupAssignment) MarshalJSON() ([]byte, error) {
	if p.RawJSON != nil {
		return append([]byte(nil), p.RawJSON...), nil
	}
	type snapshot GroupAssignment
	return json.Marshal(snapshot(p))
}
func (GroupAssignment) String() string     { return "channels policy group assignment (redacted)" }
func (p GroupAssignment) GoString() string { return p.String() }

type UserAssignment struct{ Identity, PolicyName string }
type TenantAssignment struct{ PolicyName string }
type GroupGrant struct {
	ID, PolicyName string
	Rank           *int32
}
type Assignment struct {
	User   *UserAssignment
	Tenant *TenantAssignment
	Group  *GroupGrant
}
