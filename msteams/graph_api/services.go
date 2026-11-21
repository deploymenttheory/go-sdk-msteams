// Package graphapi composes Graph resource services for a client.
package graphapi

import (
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel"
	channelmember "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_member"
	channelusermanagement "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_user_management"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
	grouprelationship "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group_relationship"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
	teammanagement "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_management"
	teammember "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_member"
	teamoperation "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_operation"
	teamusermanagement "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_user_management"
)

// Services is the typed Graph service collection. Resource fields are added with
// their evidenced implementations; no placeholder management methods are exposed.
type Services struct {
	scope                 *client.Scope
	Teams                 *team.Service
	Channels              *channel.Service
	Groups                *group.Service
	TeamMembers           *teammember.Service
	TeamOperations        *teamoperation.Service
	ChannelMembers        *channelmember.Service
	ChannelUserManagement *channelusermanagement.Service
	GroupMembers          *grouprelationship.MemberService
	GroupOwners           *grouprelationship.OwnerService
	UserResolver          *directoryuser.Resolver
	TeamUserManagement    *teamusermanagement.Service
	TeamManagement        *teammanagement.Service
}

// NewServices binds the collection to its client's session.
func NewServices(session *client.Session) (*Services, error) {
	scope, err := client.NewScope(client.Graph, session)
	if err != nil {
		return nil, err
	}
	return &Services{
		scope: scope, Teams: team.NewService(scope), Groups: group.NewService(scope),
		Channels: channel.NewService(scope), ChannelMembers: channelmember.NewService(scope),
		TeamOperations:        teamoperation.NewService(scope),
		ChannelUserManagement: channelusermanagement.NewService(scope),
		TeamMembers:           teammember.NewService(scope), GroupMembers: grouprelationship.NewMemberService(scope),
		GroupOwners: grouprelationship.NewOwnerService(scope), UserResolver: directoryuser.NewResolver(scope),
		TeamUserManagement: teamusermanagement.NewService(scope),
		TeamManagement:     teammanagement.NewService(scope),
	}, nil
}

// Scope returns the backend dependency shared by this collection's resources.
func (s *Services) Scope() *client.Scope { return s.scope }
