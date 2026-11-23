package provisioningdiagnostics

import (
	"context"
	"encoding/json"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// Move-CsInternalHelper dispatches each ActionType to a HostedMigration
// operation. The module sends them with X-MS-Target-Uri set to the discovered
// forest plus "HostedMigration" and with the tenant in X-MS-Tenant-Id.
const (
	migrationRoute  = "Teams.HostedMigrationService"
	migrationTarget = "HostedMigration"
	moveCommand     = "Move-CsInternalHelper"
)

func (m UserMove) header() (http.Header, bool) {
	tenant, ok := guid(m.TenantID)
	if !ok || !validText(m.UserSipURI) {
		return nil, false
	}
	h := headers(moveCommand)
	h.Set("X-MS-Tenant-Id", tenant)
	return h, true
}

func (s *Service) move(ctx context.Context, m UserMove, operation string, body any, replay client.ReplaySafety) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	h, ok := m.header()
	if !ok {
		return nil, ErrInput
	}
	response, err := s.client.DoTarget(ctx, migrationTarget, client.Request{Method: http.MethodPost, Path: []string{migrationRoute, operation}, Body: body, Header: h, Replay: replay}, nil)
	if err != nil {
		return response, err
	}
	if !accepted(response, []int{http.StatusOK}) {
		return response, ErrResponse
	}
	return response, nil
}

type sipBody struct {
	UserSipURI string `json:"userSipUri"`
}

type versionBody struct {
	UserSipURI   string `json:"userSipUri"`
	MajorVersion string `json:"majorVersion"`
}

// envelope decodes a required object property from a HostedMigration response.
func envelope(raw []byte, name string, out any) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return client.ErrDecode
	}
	value, ok := fields[name]
	if !ok || json.Unmarshal(value, out) != nil {
		return client.ErrDecode
	}
	return nil
}

// ValidateUserMove checks whether a user can move between on-premises and
// online (Move-CsInternalHelper -ActionType Validateuser) and returns the
// user's current deployment. It changes nothing and may be replayed.
func (s *Service) ValidateUserMove(ctx context.Context, request ValidateUserMoveRequest) (*DeploymentInfo, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validText(request.CmdletVersion) || !validText(request.LocalDeployment.PresenceFqdn) || !validText(request.LocalDeployment.RegistrarFqdn) {
		return nil, nil, ErrInput
	}
	body := struct {
		LocalDeploymentInfo LocalDeployment `json:"localDeploymentInfo"`
		TeamsData           TeamsData       `json:"teamsData"`
		UserSipURI          string          `json:"userSipUri"`
		MoveToCloud         bool            `json:"moveToCloud"`
		Force               bool            `json:"force"`
		CmdletVersion       string          `json:"cmdletVersion"`
	}{request.LocalDeployment, request.TeamsData, request.UserSipURI, request.MoveToCloud, request.Force, request.CmdletVersion}
	response, err := s.move(ctx, request.UserMove, "ValidateUser", body, client.ReplaySafe)
	if err != nil {
		return nil, response, err
	}
	var value *DeploymentInfo
	if envelope(response.Body(), "ValidateUserResult", &value) != nil || value == nil {
		return nil, response, client.ErrDecode
	}
	return value, response, nil
}

// BeginUserMove starts a user move (Move-CsInternalHelper -ActionType
// Beginmove) and returns the opaque move state. It is never replayed.
func (s *Service) BeginUserMove(ctx context.Context, request UserMoveVersionRequest) (*MigrationData, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validText(request.MajorVersion) {
		return nil, nil, ErrInput
	}
	response, err := s.move(ctx, request.UserMove, "BeginMove", versionBody{request.UserSipURI, request.MajorVersion}, client.ReplayNever)
	if err != nil {
		return nil, response, err
	}
	var value *MigrationData
	if envelope(response.Body(), "BeginMoveResult", &value) != nil || value == nil {
		return nil, response, client.ErrDecode
	}
	return value, response, nil
}

// CompleteUserMove completes a user move (Move-CsInternalHelper -ActionType
// Completemove). It is never replayed.
func (s *Service) CompleteUserMove(ctx context.Context, request UserMoveVersionRequest) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if !validText(request.MajorVersion) {
		return nil, ErrInput
	}
	return s.move(ctx, request.UserMove, "CompleteMove", versionBody{request.UserSipURI, request.MajorVersion}, client.ReplayNever)
}

// GetUserMovePolicies reads the policies a moving user carries
// (Move-CsInternalHelper -ActionType Getpolicy). It changes nothing and may be
// replayed.
func (s *Service) GetUserMovePolicies(ctx context.Context, user UserMove) ([]*KeyValue, *resty.Response, error) {
	response, err := s.move(ctx, user, "GetPolicies", sipBody{user.UserSipURI}, client.ReplaySafe)
	if err != nil {
		return nil, response, err
	}
	var values []*KeyValue
	if envelope(response.Body(), "GetPoliciesResult", &values) != nil || values == nil {
		return nil, response, client.ErrDecode
	}
	for _, value := range values {
		if value == nil {
			return nil, response, client.ErrDecode
		}
	}
	return values, response, nil
}

// RollbackUserMove rolls back a user move (Move-CsInternalHelper -ActionType
// Rollback). It is never replayed.
func (s *Service) RollbackUserMove(ctx context.Context, user UserMove) (*resty.Response, error) {
	return s.move(ctx, user, "Rollback", sipBody{user.UserSipURI}, client.ReplayNever)
}

// RehomeUserMove rehomes a moving user (Move-CsInternalHelper -ActionType
// Rehomeuser). It is never replayed.
func (s *Service) RehomeUserMove(ctx context.Context, request RehomeUserMoveRequest) (*resty.Response, error) {
	body := struct {
		TeamsData   TeamsData `json:"teamsData"`
		UserSipURI  string    `json:"userSipUri"`
		MoveToCloud bool      `json:"moveToCloud"`
	}{request.TeamsData, request.UserSipURI, request.MoveToCloud}
	return s.move(ctx, request.UserMove, "RehomeUser", body, client.ReplayNever)
}

// SetMovedResourceData stores a moved user's resource data
// (Move-CsInternalHelper -ActionType Setmovedresourcedata). It is never
// replayed.
func (s *Service) SetMovedResourceData(ctx context.Context, request MovedResourceDataRequest) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if !validText(request.ResourceData) || !validText(request.MajorVersion) {
		return nil, ErrInput
	}
	body := struct {
		ResourceData struct {
			Datastr string `json:"Datastr"`
		} `json:"resourceData"`
		TeamsData    TeamsData `json:"teamsData"`
		UserSipURI   string    `json:"userSipUri"`
		MajorVersion string    `json:"majorVersion"`
	}{TeamsData: request.TeamsData, UserSipURI: request.UserSipURI, MajorVersion: request.MajorVersion}
	body.ResourceData.Datastr = request.ResourceData
	return s.move(ctx, request.UserMove, "SetMovedResourceData", body, client.ReplayNever)
}
