// Package channelusermanagement composes channel and membership resources for
// name/user-oriented administration. Native resource CRUD remains independent.
package channelusermanagement

import (
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	member "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_member"
	"github.com/go-resty/resty/v2"
)

// Role is a projected category or action selector, not a native role-array item.
type Role string

const (
	Owner  Role = "owner"
	Member Role = "member"
	Guest  Role = "guest"
)

type Action string

const (
	ResolveChannel    Action = "resolve-channel"
	ResolveUser       Action = "resolve-user"
	ResolveMembership Action = "resolve-membership"
	ReadMembers       Action = "read-members"
	EnrichUsers       Action = "enrich-users"
	CreateMember      Action = "create-member"
	PromoteOwner      Action = "promote-owner"
	DemoteOwner       Action = "demote-owner"
	DeleteMember      Action = "delete-member"
)

type State string

const (
	Succeeded   State = "succeeded"
	Unconfirmed State = "unconfirmed"
)

// Stage records observed completion. Unconfirmed never proves that a mutation
// was absent or rolled back. Responses remain available for explicit inspection.
type Stage struct {
	Action   Action
	State    State
	Response *resty.Response
}

func (Stage) String() string     { return "channel-user stage (redacted)" }
func (s Stage) GoString() string { return s.String() }

// ChangeResult retains resolved identities and completed stages on any later
// failure. Exactly one native mutation is selected; no rollback is attempted.
type ChangeResult struct {
	ChannelID    string
	UserID       member.DirectoryUserID
	MembershipID member.MembershipID
	Member       *member.Member
	Stages       []Stage
	Response     *resty.Response
}

func (ChangeResult) String() string     { return "channel-user change result (redacted)" }
func (r ChangeResult) GoString() string { return r.String() }

// ChangeOptions selects ordinary membership or an owner action. Empty Role adds
// or deletes membership; Owner promotes or demotes an existing membership. No
// other write role is accepted. TenantID is an optional explicit UUID; it is
// forwarded on creation and disambiguates reads for owner/removal actions.
type ChangeOptions struct {
	Role            Role
	TenantID        *string
	Traversal       client.PagingOptions
	ResolveMaxPages int
}

// UserSummary preserves native identities and the raw membership behind the
// projected User/Name/Role fields. Missing tenant/user fields are not inferred.
type UserSummary struct {
	MembershipID member.MembershipID
	UserID       *member.DirectoryUserID
	User         *string
	Name         *string
	Role         Role
	TenantID     *string
	Membership   member.Member
}

type ListOptions struct {
	Role               Role
	Traversal          client.PagingOptions
	UserLookupMaxPages int
}

// ListResult retains native partial members and projection stages. Complete is
// true only after bounded channel/member traversal and all user enrichment.
type ListResult struct {
	ChannelID string
	Items     []UserSummary
	Members   []member.Member
	Stages    []Stage
	Response  *resty.Response
	Complete  bool
}

func (ListResult) String() string     { return "channel-user list result (redacted)" }
func (r ListResult) GoString() string { return r.String() }

func record(stages *[]Stage, last **resty.Response, action Action, response *resty.Response, err error) {
	state := Succeeded
	if err != nil {
		state = Unconfirmed
	}
	*stages = append(*stages, Stage{Action: action, State: state, Response: response})
	if response != nil {
		*last = response
	}
}
