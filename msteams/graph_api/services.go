// Package graphapi composes Graph resource services for a client.
package graphapi

import (
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	grouprelationship "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group_relationship"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
	teammember "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_member"
)

// Services is the typed Graph service collection. Resource fields are added with
// their evidenced implementations; no placeholder management methods are exposed.
type Services struct {
	scope        *client.Scope
	Teams        *team.Service
	TeamMembers  *teammember.Service
	GroupMembers *grouprelationship.MemberService
	GroupOwners  *grouprelationship.OwnerService
}

// NewServices binds the collection to its client's session.
func NewServices(session *client.Session) (*Services, error) {
	scope, err := client.NewScope(client.Graph, session)
	if err != nil {
		return nil, err
	}
	return &Services{scope: scope, Teams: team.NewService(scope), TeamMembers: teammember.NewService(scope),
		GroupMembers: grouprelationship.NewMemberService(scope), GroupOwners: grouprelationship.NewOwnerService(scope)}, nil
}

// Scope returns the backend dependency shared by this collection's resources.
func (s *Services) Scope() *client.Scope { return s.scope }
