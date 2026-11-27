package channel

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"regexp"
	"strings"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
	"github.com/go-resty/resty/v2"
)

// CreateResult distinguishes a completed standard/private channel from an
// accepted shared-channel operation. Response survives extraction failures and
// retains the last preparation response if creation produces no HTTP response.
type CreateResult struct {
	Channel   *Channel
	Operation *client.Operation
	Response  *resty.Response

	// CreateWithOwner retains the resolver result and its last HTTP response
	// independently of the POST result. Primitive Create leaves these nil.
	OwnerResolution *directoryuser.Resolution
	OwnerResponse   *resty.Response
}

func (CreateResult) String() string     { return "channel creation result (redacted)" }
func (r CreateResult) GoString() string { return r.String() }

type ownerPayload struct {
	Type  string   `json:"@odata.type"`
	User  string   `json:"user@odata.bind"`
	Roles []string `json:"roles"`
}

type createPayload struct {
	DisplayName     string               `json:"displayName"`
	Description     *string              `json:"description,omitempty"`
	MembershipType  MembershipType       `json:"membershipType"`
	Members         []ownerPayload       `json:"members,omitempty"`
	SharedWithTeams *[]InitialSharedTeam `json:"sharedWithTeams,omitempty"`
}

func validateCreate(value CreateRequest) error {
	if strings.TrimSpace(value.DisplayName) == "" {
		return client.ErrRequest
	}
	typeName := value.MembershipType
	if typeName == "" {
		typeName = Standard
	}
	if !validType(typeName) || value.OwnerUserID != nil && (typeName == Standard || !directoryID.MatchString(*value.OwnerUserID)) {
		return client.ErrRequest
	}
	if value.SharedWithTeams != nil {
		for _, target := range *value.SharedWithTeams {
			if !directoryID.MatchString(target.ID) {
				return client.ErrRequest
			}
		}
	}
	return nil
}

// Create sends one native POST. It uses documented lowercase membership types;
// the pinned PowerShell serializer's uppercase values remain separate evidence.
// A missing owner is left to server validation. It does not poll or replay.
func (s *Service) Create(ctx context.Context, teamID string, value CreateRequest) (CreateResult, error) {
	if err := s.ready(ctx); err != nil {
		return CreateResult{}, err
	}
	req, err := request(http.MethodPost, teamID, "", Hosted)
	if err != nil {
		return CreateResult{}, err
	}
	if err := validateCreate(value); err != nil {
		return CreateResult{}, err
	}
	if value.MembershipType == "" {
		value.MembershipType = Standard
	}
	body := createPayload{DisplayName: value.DisplayName, Description: value.Description, MembershipType: value.MembershipType, SharedWithTeams: value.SharedWithTeams}
	if value.OwnerUserID != nil {
		env, err := s.scope.Session().Environment()
		if err != nil {
			return CreateResult{}, err
		}
		body.Members = []ownerPayload{{Type: "#microsoft.graph.aadUserConversationMember", Roles: []string{"owner"},
			User: env.Endpoints.Graph + "/beta/users('" + *value.OwnerUserID + "')"}}
	}
	req.Body = body
	expected := http.StatusCreated
	if value.MembershipType == Shared {
		expected = http.StatusAccepted
	}
	response, err := s.do(ctx, req, expected)
	result := CreateResult{Response: response}
	if err != nil {
		return result, err
	}
	if value.MembershipType == Shared {
		location := response.Header().Get("Location")
		if location == "" {
			// The documented shared-channel 202 response uses Content-Location.
			// Both headers still pass the same team/origin operation validation.
			location = response.Header().Get("Content-Location")
		}
		result.Operation, err = s.operation(teamID, location, response)
		return result, err
	}
	var created Channel
	if json.Unmarshal(response.Body(), &created) != nil {
		return result, client.ErrDecode
	}
	result.Channel = &created
	return result, nil
}

// CreateWithOwner explicitly composes directory resolution and channel creation.
// UUID selectors remain local; UPN lookup and a unique complete mail fallback
// use the shared resolver. Conflicting owner inputs fail before either lookup.
// This uses the canonical /users fallback, not the pinned historical alias.
func (s *Service) CreateWithOwner(ctx context.Context, teamID string, value CreateRequest, owner string, options directoryuser.ResolveOptions) (CreateResult, error) {
	ctx, stop, err := s.begin(ctx)
	if err != nil {
		return CreateResult{}, err
	}
	defer stop()
	if !directoryID.MatchString(teamID) || value.OwnerUserID != nil || value.MembershipType != Private && value.MembershipType != Shared {
		return CreateResult{}, client.ErrRequest
	}
	if err := validateCreate(value); err != nil {
		return CreateResult{}, err
	}
	resolved, response, err := directoryuser.NewResolver(s.scope).Resolve(ctx, owner, options)
	if err != nil {
		return CreateResult{Response: response, OwnerResolution: &resolved, OwnerResponse: response}, err
	}
	value.OwnerUserID = &resolved.ID
	result, err := s.Create(ctx, teamID, value)
	result.OwnerResolution, result.OwnerResponse = &resolved, response
	if result.Response == nil {
		result.Response = response
	}
	return result, err
}

var (
	operationPath      = regexp.MustCompile(`^/beta/teams/([^/]+)/operations/([a-zA-Z0-9_-]+)$`)
	odataOperationPath = regexp.MustCompile(`^/beta/teams\('([^']+)'\)/operations\('([a-zA-Z0-9_-]+)'\)$`)
)

// ResumeOperation reopens only this team's Graph beta operation route. It is
// local, retains an accepted state and never replays the initiating POST.
func (s *Service) ResumeOperation(teamID, location string) (*client.Operation, error) {
	if err := s.local(); err != nil {
		return nil, err
	}
	return s.operation(teamID, location, nil)
}

func (s *Service) operation(teamID, location string, response *resty.Response) (*client.Operation, error) {
	if !directoryID.MatchString(teamID) {
		return nil, client.ErrOperation
	}
	u, err := url.Parse(location)
	if err != nil || location == "" || u.User != nil || u.Opaque != "" || u.RawQuery != "" || u.ForceQuery || u.Fragment != "" {
		return nil, client.ErrOperation
	}
	env, err := s.scope.Session().Environment()
	if err != nil {
		return nil, err
	}
	if u.Scheme == "" {
		if u.Host != "" || !strings.HasPrefix(u.Path, "/") || strings.HasPrefix(u.Path, "//") {
			return nil, client.ErrOperation
		}
		if strings.HasPrefix(u.Path, "/teams(") || strings.HasPrefix(u.Path, "/teams/") {
			u.Path = "/beta" + u.Path
			u.RawPath = ""
		}
		origin, _ := url.Parse(env.Endpoints.Graph)
		u.Scheme, u.Host = origin.Scheme, origin.Host
	}
	if u.Scheme+"://"+u.Host != env.Endpoints.Graph {
		return nil, client.ErrOperation
	}
	match := operationPath.FindStringSubmatch(u.Path)
	if match == nil {
		match = odataOperationPath.FindStringSubmatch(u.Path)
	}
	if match == nil || !strings.EqualFold(match[1], teamID) {
		return nil, client.ErrOperation
	}
	req := client.Request{Method: http.MethodGet, URL: u.String(), Header: http.Header{"Accept": {"application/json"}}, Replay: client.ReplaySafe}
	return client.NewOperation(s.scope, req, match[2], team.DecodeOperation, 31*time.Second, response)
}
