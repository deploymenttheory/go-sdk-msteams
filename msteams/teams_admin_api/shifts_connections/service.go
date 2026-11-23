// Package shiftsconnections manages Teams Shifts connector connections to
// workforce management (WFM) systems and the connector instances that
// synchronize them with Shifts, through the Teams.ShiftsConnector service.
package shiftsconnections

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/go-resty/resty/v2"
)

const (
	root        = "Teams.ShiftsConnector"
	connections = "connections"
	instances   = "connectorinstances"
)

// Service composes the bound ConfigAPI backend with lazy forest discovery.
type Service struct{ client client.Client }

// NewService performs no I/O. Pass the backend bound to ConfigAPI's audience.
func NewService(base client.Client) *Service { return &Service{client: configapi.New(base)} }

func (s *Service) check(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil || s.client == nil {
		return client.ErrClosed
	}
	return nil
}

func validText(value string) bool {
	return strings.TrimSpace(value) != "" && !strings.ContainsFunc(value, unicode.IsControl)
}

func validPart(value string) bool {
	return validText(value) && value != "." && value != ".." && !strings.ContainsAny(value, `/\`)
}

func optionalText(value *string) bool {
	return value == nil || !strings.ContainsFunc(*value, unicode.IsControl)
}

func headers(command string, ifMatch *string) http.Header {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}}
	if ifMatch != nil {
		h.Set("If-Match", *ifMatch)
	}
	return h
}

func path(parts ...string) []string { return append([]string{root, "api", "v2"}, parts...) }

func (s *Service) do(ctx context.Context, request client.Request, statuses ...int) (*resty.Response, error) {
	response, err := s.client.Do(ctx, request, nil)
	if err != nil {
		return response, err
	}
	for _, status := range statuses {
		if response != nil && response.StatusCode() == status {
			return response, nil
		}
	}
	return response, ErrResponse
}

// decodeOne decodes a non-null JSON object into a new T.
func decodeOne[T any](response *resty.Response) (*T, error) {
	var value *T
	if json.Unmarshal(response.Body(), &value) != nil || value == nil {
		return nil, client.ErrDecode
	}
	return value, nil
}

// decodeList decodes the service's {"value":[...]} collection envelope.
func decodeList[T any](response *resty.Response) ([]*T, error) {
	var envelope struct {
		Value []*T `json:"value"`
	}
	if json.Unmarshal(response.Body(), &envelope) != nil || envelope.Value == nil {
		return nil, client.ErrDecode
	}
	for _, value := range envelope.Value {
		if value == nil {
			return nil, client.ErrDecode
		}
	}
	return envelope.Value, nil
}

func (s *Service) read(ctx context.Context, command string, parts ...string) (*resty.Response, error) {
	return s.do(ctx, client.Request{Method: http.MethodGet, Path: path(parts...), Header: headers(command, nil), Replay: client.ReplaySafe}, http.StatusOK)
}

// ListConnectors reads the supported WFM connectors
// (Get-CsTeamsShiftsConnectionConnector).
func (s *Service) ListConnectors(ctx context.Context) ([]*Connector, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	response, err := s.read(ctx, "Get-CsTeamsShiftsConnectionConnector", "connectors")
	if err != nil {
		return nil, response, err
	}
	values, err := decodeList[Connector](response)
	if err != nil {
		return nil, response, err
	}
	return values, response, nil
}

// ListConnections reads every WFM connection (Get-CsTeamsShiftsConnection).
func (s *Service) ListConnections(ctx context.Context) ([]*Connection, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	response, err := s.read(ctx, "Get-CsTeamsShiftsConnection", connections)
	if err != nil {
		return nil, response, err
	}
	values, err := decodeList[Connection](response)
	if err != nil {
		return nil, response, err
	}
	return values, response, nil
}

// GetConnectionByID reads one WFM connection
// (Get-CsTeamsShiftsConnection -ConnectionId).
func (s *Service) GetConnectionByID(ctx context.Context, id string) (*Connection, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(id) {
		return nil, nil, ErrInput
	}
	response, err := s.read(ctx, "Get-CsTeamsShiftsConnection", connections, id)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeOne[Connection](response)
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

func connectionBody(name, connectorID, state *string, etag *string, settings Settings) map[string]any {
	body := map[string]any{}
	for key, value := range map[string]*string{"name": name, "connectorId": connectorID, "state": state, "_etag": etag} {
		if value != nil {
			body[key] = *value
		}
	}
	if settings != nil {
		body["connectorSpecificSettings"] = map[string]any(settings)
	}
	return body
}

// CreateConnection creates a WFM connection (New-CsTeamsShiftsConnection).
// It is never replayed automatically.
func (s *Service) CreateConnection(ctx context.Context, request CreateConnectionRequest) (*Connection, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validText(request.Name) || !validText(request.ConnectorID) || !validText(request.State) || request.ConnectorSpecificSettings == nil {
		return nil, nil, ErrInput
	}
	body := connectionBody(&request.Name, &request.ConnectorID, &request.State, nil, request.ConnectorSpecificSettings)
	response, err := s.do(ctx, client.Request{Method: http.MethodPost, Path: path(connections), Header: headers("New-CsTeamsShiftsConnection", nil), Body: body},
		http.StatusOK, http.StatusCreated)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeOne[Connection](response)
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

func (s *Service) writeConnection(ctx context.Context, method, command, id string, update ConnectionUpdate) (*Connection, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(id) || !optionalText(update.IfMatch) || !optionalText(update.Etag) || !optionalText(update.Name) || !optionalText(update.ConnectorID) || !optionalText(update.State) {
		return nil, nil, ErrInput
	}
	body := connectionBody(update.Name, update.ConnectorID, update.State, update.Etag, update.ConnectorSpecificSettings)
	response, err := s.do(ctx, client.Request{Method: method, Path: path(connections, id), Header: headers(command, update.IfMatch), Body: body}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeOne[Connection](response)
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// ReplaceConnectionByID replaces a WFM connection (Set-CsTeamsShiftsConnection, PUT).
func (s *Service) ReplaceConnectionByID(ctx context.Context, id string, update ConnectionUpdate) (*Connection, *resty.Response, error) {
	return s.writeConnection(ctx, http.MethodPut, "Set-CsTeamsShiftsConnection", id, update)
}

// UpdateConnectionByID changes fields of a WFM connection
// (Update-CsTeamsShiftsConnection, PATCH).
func (s *Service) UpdateConnectionByID(ctx context.Context, id string, update ConnectionUpdate) (*Connection, *resty.Response, error) {
	return s.writeConnection(ctx, http.MethodPatch, "Update-CsTeamsShiftsConnection", id, update)
}

func (s *Service) remove(ctx context.Context, command, collection, id string) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if !validPart(id) {
		return nil, ErrInput
	}
	return s.do(ctx, client.Request{Method: http.MethodDelete, Path: path(collection, id), Header: headers(command, nil)}, http.StatusOK, http.StatusNoContent)
}

// DeleteConnectionByID removes a WFM connection (Remove-CsTeamsShiftsConnection).
func (s *Service) DeleteConnectionByID(ctx context.Context, id string) (*resty.Response, error) {
	return s.remove(ctx, "Remove-CsTeamsShiftsConnection", connections, id)
}

// ListInstances reads every connector instance (Get-CsTeamsShiftsConnectionInstance).
func (s *Service) ListInstances(ctx context.Context) ([]*Instance, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	response, err := s.read(ctx, "Get-CsTeamsShiftsConnectionInstance", instances)
	if err != nil {
		return nil, response, err
	}
	values, err := decodeList[Instance](response)
	if err != nil {
		return nil, response, err
	}
	return values, response, nil
}

// GetInstanceByID reads one connector instance
// (Get-CsTeamsShiftsConnectionInstance -ConnectorInstanceId).
func (s *Service) GetInstanceByID(ctx context.Context, id string) (*Instance, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	if !validPart(id) {
		return nil, nil, ErrInput
	}
	response, err := s.read(ctx, "Get-CsTeamsShiftsConnectionInstance", instances, id)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeOne[Instance](response)
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

func instanceBody(request InstanceRequest, etag *string) (map[string]any, bool) {
	body := map[string]any{}
	for key, value := range map[string]*string{"connectionId": request.ConnectionID, "designatedActorId": request.DesignatedActorID, "name": request.Name, "state": request.State, "_etag": etag} {
		if value != nil {
			if strings.ContainsFunc(*value, unicode.IsControl) {
				return nil, false
			}
			body[key] = *value
		}
	}
	if request.ConnectorAdminEmails != nil {
		for _, email := range request.ConnectorAdminEmails {
			if !validText(email) {
				return nil, false
			}
		}
		body["connectorAdminEmails"] = request.ConnectorAdminEmails
	}
	if request.SyncFrequencyInMin != nil {
		body["syncFrequencyInMin"] = *request.SyncFrequencyInMin
	}
	if request.SyncScenarios != nil {
		body["syncScenarios"] = request.SyncScenarios
	}
	return body, true
}

// CreateInstance creates a connector instance
// (New-CsTeamsShiftsConnectionInstance). It is never replayed automatically.
func (s *Service) CreateInstance(ctx context.Context, request InstanceRequest) (*Instance, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	body, ok := instanceBody(request, nil)
	if !ok {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: http.MethodPost, Path: path(instances), Header: headers("New-CsTeamsShiftsConnectionInstance", nil), Body: body},
		http.StatusOK, http.StatusCreated)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeOne[Instance](response)
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

func (s *Service) writeInstance(ctx context.Context, method, command, id string, update InstanceUpdate) (*Instance, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	body, ok := instanceBody(update.InstanceRequest, update.Etag)
	if !ok || !validPart(id) || !validText(update.IfMatch) {
		return nil, nil, ErrInput
	}
	response, err := s.do(ctx, client.Request{Method: method, Path: path(instances, id), Header: headers(command, &update.IfMatch), Body: body}, http.StatusOK)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeOne[Instance](response)
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// ReplaceInstanceByID replaces a connector instance
// (Set-CsTeamsShiftsConnectionInstance, PUT). IfMatch is required.
func (s *Service) ReplaceInstanceByID(ctx context.Context, id string, update InstanceUpdate) (*Instance, *resty.Response, error) {
	return s.writeInstance(ctx, http.MethodPut, "Set-CsTeamsShiftsConnectionInstance", id, update)
}

// UpdateInstanceByID changes fields of a connector instance
// (Update-CsTeamsShiftsConnectionInstance, PATCH). IfMatch is required.
func (s *Service) UpdateInstanceByID(ctx context.Context, id string, update InstanceUpdate) (*Instance, *resty.Response, error) {
	return s.writeInstance(ctx, http.MethodPatch, "Update-CsTeamsShiftsConnectionInstance", id, update)
}

// DeleteInstanceByID removes a connector instance
// (Remove-CsTeamsShiftsConnectionInstance).
func (s *Service) DeleteInstanceByID(ctx context.Context, id string) (*resty.Response, error) {
	return s.remove(ctx, "Remove-CsTeamsShiftsConnectionInstance", instances, id)
}

// Validate asks the service to sign in to the WFM system with the given
// settings without saving them (Test-CsTeamsShiftsConnectionValidate). It
// sends credentials to the WFM system and is never replayed automatically.
func (s *Service) Validate(ctx context.Context, request ValidateRequest) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if !validText(request.Name) || !validText(request.ConnectorID) || request.ConnectorSpecificSettings == nil {
		return nil, ErrInput
	}
	body := connectionBody(&request.Name, &request.ConnectorID, nil, nil, request.ConnectorSpecificSettings)
	return s.do(ctx, client.Request{Method: http.MethodPost, Path: []string{root, "api", "v1", "connect"}, Header: headers("Test-CsTeamsShiftsConnectionValidate", nil), Body: body},
		http.StatusOK, http.StatusNoContent)
}
