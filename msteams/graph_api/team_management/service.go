// Package teammanagement composes native group/team resources for PowerShell
// management behavior. Individual CRUD/actions remain available separately.
package teammanagement

import (
	"context"
	"regexp"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
	"github.com/go-resty/resty/v2"
)

var directoryID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

type Action string

const (
	ReadGroup            Action = "read-group"
	ReadTeam             Action = "read-team"
	UpdateGroup          Action = "update-group"
	UpdateTeam           Action = "update-team"
	CheckArchivedState   Action = "check-archived-state"
	ArchiveTeam          Action = "archive-team"
	UnarchiveTeam        Action = "unarchive-team"
	WaitArchivedState    Action = "wait-archived-state"
	ConfirmArchivedState Action = "confirm-archived-state"
)

type State string

const (
	Succeeded   State = "succeeded"
	Unconfirmed State = "unconfirmed"
	Accepted    State = "accepted"
	Blocked     State = "blocked"
)

type Stage struct {
	Action   Action
	State    State
	Response *resty.Response
}

func (Stage) String() string     { return "team management stage (redacted)" }
func (s Stage) GoString() string { return s.String() }

// ManagedTeam projects backing-group properties and native team state without
// discarding either snapshot. Nil fields remain unknown, not fabricated defaults.
type ManagedTeam struct {
	ID string
	group.Properties
	Settings      team.Settings
	InternalID    *string
	IsArchived    *bool
	GroupSnapshot *group.Group
	TeamSnapshot  *team.Team
}

// Result retains stage metadata and the known group/team ID on failure. Value
// is populated only after both paired reads complete; no transaction is implied.
type Result struct {
	ID       string
	Value    *ManagedTeam
	Stages   []Stage
	Response *resty.Response
}

func (Result) String() string     { return "team management result (redacted)" }
func (r Result) GoString() string { return r.String() }

func (r *Result) record(action Action, response *resty.Response, err error) {
	state := Succeeded
	if err != nil {
		state = Unconfirmed
	}
	r.Stages = append(r.Stages, Stage{Action: action, State: state, Response: response})
	if response != nil {
		r.Response = response
	}
}

// UpdateRequest separates backing-group changes from native team settings.
type UpdateRequest struct {
	Properties group.UpdateRequest
	Settings   team.Settings
}

// Service composes resources on one Graph tenant/cloud/identity binding.
type Service struct {
	scope  *client.Scope
	groups *group.Service
	teams  *team.Service
}

// NewService constructs local dependencies on the same Graph scope.
func NewService(scope *client.Scope) *Service {
	return &Service{scope: scope, groups: group.NewService(scope), teams: team.NewService(scope)}
}

func (s *Service) begin(ctx context.Context, id string) (context.Context, context.CancelFunc, error) {
	if ctx == nil {
		return nil, nil, client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if s == nil || s.scope == nil || s.scope.Session() == nil {
		return nil, nil, client.ErrClosed
	}
	if s.scope.Backend() != client.Graph {
		return nil, nil, client.ErrDestination
	}
	if !directoryID.MatchString(id) {
		return nil, nil, client.ErrRequest
	}
	return s.scope.Session().Begin(ctx)
}

func (s *Service) read(ctx context.Context, result *Result) error {
	groupData, response, err := s.groups.GetByID(ctx, result.ID, group.ReadOptions{})
	if err == nil && !strings.EqualFold(groupData.ID, result.ID) {
		err = client.ErrDecode
	}
	result.record(ReadGroup, response, err)
	if err != nil {
		return err
	}
	teamData, response, err := s.teams.GetByID(ctx, result.ID)
	if err == nil && !strings.EqualFold(teamData.ID, result.ID) {
		err = client.ErrDecode
	}
	result.record(ReadTeam, response, err)
	if err != nil {
		return err
	}
	result.Value = &ManagedTeam{ID: result.ID, Properties: groupData.Properties, Settings: teamData.Settings, InternalID: teamData.InternalID, IsArchived: teamData.IsArchived, GroupSnapshot: groupData, TeamSnapshot: teamData}
	return nil
}

// GetByID performs the pinned paired group/team read and projects group-owned
// display/description/visibility/mailNickname/classification over team fields.
// It implements the identity branch, not the separate tenant/user list branches.
func (s *Service) GetByID(ctx context.Context, id string) (Result, error) {
	result := Result{ID: id}
	ctx, stop, err := s.begin(ctx, id)
	if err != nil {
		return result, err
	}
	defer stop()
	err = s.read(ctx, &result)
	return result, err
}

func completeSettings(value team.Settings) team.Settings {
	if value.MemberSettings == nil {
		value.MemberSettings = &team.MemberSettings{}
	}
	if value.GuestSettings == nil {
		value.GuestSettings = &team.GuestSettings{}
	}
	if value.MessagingSettings == nil {
		value.MessagingSettings = &team.MessagingSettings{}
	}
	if value.FunSettings == nil {
		value.FunSettings = &team.FunSettings{}
	}
	if value.DiscoverySettings == nil {
		value.DiscoverySettings = &team.DiscoverySettings{}
	}
	return value
}

// UpdateByID patches group properties first, then all five settings containers,
// then performs paired readback. Omitted containers are explicitly empty as in
// the pinned Set-Team command; native Teams.UpdateByID remains sparse. Failure
// retains completed stages and does not roll back or replay either mutation.
func (s *Service) UpdateByID(ctx context.Context, id string, value UpdateRequest) (Result, error) {
	result := Result{ID: id}
	ctx, stop, err := s.begin(ctx, id)
	if err != nil {
		return result, err
	}
	defer stop()
	response, err := s.groups.UpdateByID(ctx, id, value.Properties)
	result.record(UpdateGroup, response, err)
	if err != nil {
		return result, err
	}
	response, err = s.teams.UpdateByID(ctx, id, completeSettings(value.Settings))
	result.record(UpdateTeam, response, err)
	if err != nil {
		return result, err
	}
	err = s.read(ctx, &result)
	return result, err
}
