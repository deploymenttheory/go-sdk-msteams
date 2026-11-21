package graphapi

import (
	"net/http"
	"strings"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
	"github.com/go-resty/resty/v2"
)

// NewTeamOperation interprets the Location from a Graph Teams 202 response. Graph
// documents both absolute locations and unversioned /teams(...) locations. Relative
// locations are anchored to the selected Graph v1.0 endpoint, never a response-supplied
// authority. Construction validates the resulting route before token acquisition.
// Polling respects Graph's documented interval of more than 30 seconds.
func NewTeamOperation(scope *client.Scope, accepted *resty.Response) (*client.Operation, error) {
	if scope == nil || scope.Backend() != client.Graph || accepted == nil || accepted.StatusCode() != http.StatusAccepted {
		return nil, client.ErrOperation
	}
	location := accepted.Header().Get("Location")
	if location == "" {
		return nil, client.ErrOperation
	}
	environment, err := scope.Session().Environment()
	if err != nil {
		return nil, err
	}
	if strings.HasPrefix(location, "/teams(") || strings.HasPrefix(location, "/teams/") {
		location = environment.Endpoints.Graph + "/v1.0" + location
	} else if strings.HasPrefix(location, "/v1.0/") {
		location = environment.Endpoints.Graph + location
	}
	return client.NewOperation(scope, client.Request{Method: http.MethodGet, URL: location, Replay: client.ReplaySafe}, "", DecodeTeamOperation, 31*time.Second, accepted)
}

// DecodeTeamOperation maps the Teams async-operation envelope. Unknown future
// statuses fail explicitly instead of being treated as completion. The target
// resource location is returned as opaque metadata and is never followed here.
func DecodeTeamOperation(body []byte) (client.OperationStatus, error) {
	return team.DecodeOperation(body)
}
