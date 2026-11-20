package graphapi

import (
	"encoding/json"
	"net/http"
	"strings"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// NewTeamOperation interprets the Location from a Graph Teams 202 response. Graph
// documents both absolute locations and unversioned /teams(...) locations. Relative
// locations are anchored to the commercial v1.0 endpoint, never a response-supplied
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
	if strings.HasPrefix(location, "/teams(") || strings.HasPrefix(location, "/teams/") {
		location = "https://graph.microsoft.com/v1.0" + location
	} else if strings.HasPrefix(location, "/v1.0/") {
		location = "https://graph.microsoft.com" + location
	}
	return client.NewOperation(scope, client.Request{Method: http.MethodGet, URL: location, Replay: client.ReplaySafe}, "", DecodeTeamOperation, 31*time.Second, accepted)
}

// DecodeTeamOperation maps the Teams async-operation envelope. Unknown future
// statuses fail explicitly instead of being treated as completion. The target
// resource location is returned as opaque metadata and is never followed here.
func DecodeTeamOperation(body []byte) (client.OperationStatus, error) {
	var wire struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Target string `json:"targetResourceLocation"`
	}
	if json.Unmarshal(body, &wire) != nil || wire.ID == "" {
		return client.OperationStatus{}, client.ErrDecode
	}
	status := client.OperationStatus{ID: wire.ID, TargetLocation: wire.Target}
	switch wire.Status {
	case "notStarted":
		status.State = client.OperationAccepted
	case "inProgress":
		status.State = client.OperationRunning
	case "succeeded":
		status.State = client.OperationSucceeded
	case "failed":
		status.State = client.OperationFailed
	default:
		return client.OperationStatus{}, client.ErrOperationState
	}
	return status, nil
}
