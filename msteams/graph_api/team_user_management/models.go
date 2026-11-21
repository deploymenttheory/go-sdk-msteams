// Package teamusermanagement composes backing-group reference operations for
// the pinned team-user commands. Native conversation membership CRUD is separate.
package teamusermanagement

import (
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
	relation "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group_relationship"
	"github.com/go-resty/resty/v2"
)

// Role is the command composition's category, not a conversation-member qualifier.
type Role string

const (
	Owner  Role = "owner"
	Member Role = "member"
	Guest  Role = "guest"
)

type Action string

const (
	ResolveUser           Action = "resolve-user"
	ReadMembers           Action = "read-members"
	ReadOwners            Action = "read-owners"
	AddMemberReference    Action = "add-member-reference"
	AddOwnerReference     Action = "add-owner-reference"
	CheckLastOwner        Action = "check-last-owner"
	RemoveOwnerReference  Action = "remove-owner-reference"
	RemoveMemberReference Action = "remove-member-reference"
)

// State records observed stages. Unconfirmed never means a mutation was absent
// or rolled back; retain identities and reconcile ambiguous remote outcomes.
type State string

const (
	Succeeded   State = "succeeded"
	Skipped     State = "skipped"
	Unconfirmed State = "unconfirmed"
	Blocked     State = "blocked"
)

type Stage struct {
	Action   Action
	State    State
	Response *resty.Response
}

func (Stage) String() string     { return "team-user stage (redacted)" }
func (s Stage) GoString() string { return s.String() }

// ChangeResult survives a failed later stage. Succeeded mutation stages are
// confirmed writes; skipped stages mean the prerequisite/read selected no write.
// A non-nil error prevents treating the whole workflow as successful.
type ChangeResult struct {
	UserID           relation.DirectoryObjectID
	ResolutionSource directoryuser.Source
	Stages           []Stage
	Response         *resty.Response
}

func (ChangeResult) String() string     { return "team-user change result (redacted)" }
func (r ChangeResult) GoString() string { return r.String() }

func (r *ChangeResult) record(action Action, state State, response *resty.Response) {
	r.Stages = append(r.Stages, Stage{Action: action, State: state, Response: response})
	if response != nil {
		r.Response = response
	}
}

// ChangeOptions applies to Add/Remove. Omitted role means Member; Owner adds
// ownership after membership or removes only ownership. Traversal bounds apply
// independently to each relationship read. ResolveMaxPages bounds mail lookup.
type ChangeOptions struct {
	Role            Role
	Traversal       client.PagingOptions
	ResolveMaxPages int
}

// UserSummary reproduces the pinned directory projection. Raw directory data
// remains available: nil userType is not reinterpreted in the native read model,
// even though this compatibility projection categorizes it as member.
type UserSummary struct {
	UserID          relation.DirectoryObjectID
	User            *string
	Name            *string
	Role            Role
	DirectoryObject relation.DirectoryObject
}

type ListOptions struct {
	Role      Role
	Traversal client.PagingOptions
}

// ListResult preserves a partial projection on a read failure. Complete is true
// only after every required relationship has traversed successfully.
type ListResult struct {
	Items    []UserSummary
	Stages   []Stage
	Response *resty.Response
	Complete bool
}

func (ListResult) String() string     { return "team-user list result (redacted)" }
func (r ListResult) GoString() string { return r.String() }
