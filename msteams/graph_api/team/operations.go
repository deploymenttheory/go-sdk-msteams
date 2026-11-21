package team

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// Graph requires more than 30 seconds between teams-operation status checks.
const operationInterval = 31 * time.Second

var (
	operationPath      = regexp.MustCompile(`^/v1\.0/teams/([^/]+)/operations/([a-zA-Z0-9_-]+)$`)
	odataOperationPath = regexp.MustCompile(`^/v1\.0/teams\('([^']+)'\)/operations\('([a-zA-Z0-9_-]+)'\)$`)
)

// AcceptedOperation retains the HTTP response even if extracting a safe polling
// handle fails. A non-nil Operation starts Accepted, never Succeeded. TeamID is
// populated only from a validated team-operation route. These identifiers and
// response bodies may contain tenant data; inspect them explicitly.
type AcceptedOperation struct {
	Operation *client.Operation
	TeamID    string
	Response  *resty.Response
}

func (AcceptedOperation) String() string     { return "accepted team operation (redacted)" }
func (a AcceptedOperation) GoString() string { return a.String() }

type memberPayload struct {
	Type  string   `json:"@odata.type"`
	Roles []string `json:"roles"`
	User  string   `json:"user@odata.bind"`
}

type createPayload struct {
	Settings
	Template         string          `json:"template@odata.bind"`
	DisplayName      string          `json:"displayName"`
	Description      *string         `json:"description,omitempty"`
	Visibility       *Visibility     `json:"visibility,omitempty"`
	FirstChannelName *string         `json:"firstChannelName,omitempty"`
	Members          []memberPayload `json:"members,omitempty"`
}

// Create submits a native team create exactly once. An application request must
// specify an initial user. It does not resolve users, create a group first, wait,
// or automatically delete anything after an ambiguous response.
func (s *Service) Create(ctx context.Context, value CreateRequest) (AcceptedOperation, error) {
	if err := s.ready(ctx); err != nil {
		return AcceptedOperation{}, err
	}
	if strings.TrimSpace(value.DisplayName) == "" {
		return AcceptedOperation{}, client.ErrRequest
	}
	if value.TemplateID == "" {
		value.TemplateID = "standard"
	}
	if !bindingID.MatchString(value.TemplateID) {
		return AcceptedOperation{}, client.ErrRequest
	}
	mode, err := s.mode()
	if err != nil {
		return AcceptedOperation{}, err
	}
	if mode == auth.Application && len(value.Members) == 0 {
		return AcceptedOperation{}, client.ErrRequest
	}
	env, err := s.scope.Session().Environment()
	if err != nil {
		return AcceptedOperation{}, err
	}
	origin := env.Endpoints.Graph
	body := createPayload{Settings: value.Settings,
		Template:    origin + "/v1.0/teamsTemplates('" + value.TemplateID + "')",
		DisplayName: value.DisplayName, Description: value.Description,
		Visibility: value.Visibility, FirstChannelName: value.FirstChannelName}
	for _, member := range value.Members {
		if !directoryID.MatchString(member.UserID) {
			return AcceptedOperation{}, client.ErrRequest
		}
		roles := []string{}
		if member.Owner {
			roles = []string{"owner"}
		}
		body.Members = append(body.Members, memberPayload{Type: "#microsoft.graph.aadUserConversationMember",
			Roles: roles, User: origin + "/v1.0/users('" + member.UserID + "')"})
	}
	req := request(http.MethodPost, "teams")
	req.Body = body
	return s.submit(ctx, req, "")
}

// Archive submits the archive operation. Omitting the option sends no body;
// this intentionally follows Graph rather than PowerShell's JSON-null artifact.
func (s *Service) Archive(ctx context.Context, id string, options ArchiveOptions) (AcceptedOperation, error) {
	if err := s.ready(ctx); err != nil {
		return AcceptedOperation{}, err
	}
	if !directoryID.MatchString(id) {
		return AcceptedOperation{}, client.ErrRequest
	}
	if options.ShouldSetSpoSiteReadOnlyForMembers != nil {
		mode, err := s.mode()
		if err != nil {
			return AcceptedOperation{}, err
		}
		if mode == auth.Application {
			return AcceptedOperation{}, ErrArchiveOption
		}
	}
	req := request(http.MethodPost, "teams", id, "archive")
	if options.ShouldSetSpoSiteReadOnlyForMembers != nil {
		req.Body = options
	}
	return s.submit(ctx, req, id)
}

// Unarchive sends no body, as documented by Graph. It does not send the pinned
// PowerShell helper's empty JSON string or pre-read the archive state.
func (s *Service) Unarchive(ctx context.Context, id string) (AcceptedOperation, error) {
	if err := s.ready(ctx); err != nil {
		return AcceptedOperation{}, err
	}
	if !directoryID.MatchString(id) {
		return AcceptedOperation{}, client.ErrRequest
	}
	return s.submit(ctx, request(http.MethodPost, "teams", id, "unarchive"), id)
}

func (s *Service) submit(ctx context.Context, req client.Request, expectedTeam string) (AcceptedOperation, error) {
	response, err := s.do(ctx, req, http.StatusAccepted)
	result := AcceptedOperation{Response: response}
	if err != nil {
		return result, err
	}
	teamID, operationID, location, err := s.operationLocation(response.Header().Get("Location"))
	if err != nil {
		return result, err
	}
	if expectedTeam != "" && !strings.EqualFold(teamID, expectedTeam) {
		return result, client.ErrOperation
	}
	statusRequest := client.Request{Method: http.MethodGet, URL: location,
		Header: http.Header{"Accept": {"application/json"}}, Replay: client.ReplaySafe}
	op, err := client.NewOperation(s.scope, statusRequest, operationID, DecodeOperation, operationInterval, response)
	if err != nil {
		return result, err
	}
	result.Operation, result.TeamID = op, teamID
	return result, nil
}

// operationLocation resolves only the documented team-operation route. Graph
// examples omit the version in root-relative OData locations, so that exact
// family is resolved against v1.0 explicitly. Arbitrary same-origin routes,
// network-path references, queries, alternate versions and destinations fail.
func (s *Service) operationLocation(location string) (string, string, string, error) {
	u, err := url.Parse(location)
	if err != nil || location == "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" || u.User != nil || u.Opaque != "" {
		return "", "", "", client.ErrOperation
	}
	if u.Scheme == "" {
		if u.Host != "" || !strings.HasPrefix(u.Path, "/") || strings.HasPrefix(u.Path, "//") {
			return "", "", "", client.ErrOperation
		}
		env, envErr := s.scope.Session().Environment()
		if envErr != nil {
			return "", "", "", envErr
		}
		origin, _ := url.Parse(env.Endpoints.Graph) // Validated by session construction.
		u.Scheme, u.Host = origin.Scheme, origin.Host
		if strings.HasPrefix(u.Path, "/teams('") {
			u.Path, u.RawPath = "/v1.0"+u.Path, ""
		}
	}
	if err := s.scope.ValidateRequest(client.Request{Method: http.MethodGet, URL: u.String()}); err != nil {
		return "", "", "", err
	}
	parts := operationPath.FindStringSubmatch(u.Path)
	if parts == nil {
		parts = odataOperationPath.FindStringSubmatch(u.Path)
	}
	if len(parts) != 3 || !directoryID.MatchString(parts[1]) {
		return "", "", "", client.ErrOperation
	}
	return parts[1], parts[2], u.String(), nil
}

// ResumeOperation reconstructs a native operation under this client's tenant
// and cloud. It never resubmits the initiating mutation or follows a target URL.
// The first read is immediate; subsequent checks observe the Graph cadence.
func (s *Service) ResumeOperation(teamID, operationID string) (*client.Operation, error) {
	if err := s.ready(context.Background()); err != nil {
		return nil, err
	}
	if !directoryID.MatchString(teamID) || !bindingID.MatchString(operationID) {
		return nil, client.ErrOperation
	}
	req := request(http.MethodGet, "teams", teamID, "operations", operationID)
	req.Replay = client.ReplaySafe
	return client.NewOperation(s.scope, req, operationID, DecodeOperation, operationInterval, nil)
}

// DecodeOperation interprets a native teamsAsyncOperation status envelope.
// Unknown statuses fail explicitly; targetResourceLocation is opaque metadata.
func DecodeOperation(body []byte) (client.OperationStatus, error) {
	var value struct {
		ID     string `json:"id"`
		Status string `json:"status"`
		Target string `json:"targetResourceLocation"`
	}
	if json.Unmarshal(body, &value) != nil || value.ID == "" {
		return client.OperationStatus{}, client.ErrDecode
	}
	status := client.OperationStatus{ID: value.ID, TargetLocation: value.Target}
	switch value.Status {
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
