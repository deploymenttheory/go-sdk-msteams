// Package graphapi composes Graph resource services for a client.
package graphapi

import "github.com/deploymenttheory/go-sdk-msteams/msteams/client"

// Services is the typed Graph service collection. Resource fields are added with
// their evidenced implementations; no placeholder management methods are exposed.
type Services struct {
	scope *client.Scope
}

// NewServices binds the collection to its client's session.
func NewServices(session *client.Session) (*Services, error) {
	scope, err := client.NewScope(client.Graph, session)
	if err != nil {
		return nil, err
	}
	return &Services{scope: scope}, nil
}

// Scope returns the backend dependency shared by this collection's resources.
func (s *Services) Scope() *client.Scope { return s.scope }
