package shiftsconnections

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// The module's DateTime wire format ("yyyy-MM-ddTHH:mm:ss.fffffffK").
const scheduleDateLayout = "2006-01-02T15:04:05.0000000Z07:00"

func pathV1(parts ...string) []string { return append([]string{root, "api", "v1"}, parts...) }

// decodeEnvelope decodes the service's {"<key>":[...]} collection envelope.
func decodeEnvelope[T any](response *resty.Response, key string) ([]*T, error) {
	var envelope map[string]json.RawMessage
	if json.Unmarshal(response.Body(), &envelope) != nil || envelope == nil {
		return nil, client.ErrDecode
	}
	var values []*T
	if json.Unmarshal(envelope[key], &values) != nil || values == nil {
		return nil, client.ErrDecode
	}
	for _, value := range values {
		if value == nil {
			return nil, client.ErrDecode
		}
	}
	return values, nil
}

func (s *Service) get(ctx context.Context, command string, path []string, query url.Values) (*resty.Response, error) {
	return s.do(ctx, client.Request{Method: http.MethodGet, Path: path, Query: query, Header: headers(command, nil), Replay: client.ReplaySafe}, http.StatusOK)
}

func readList[T any](ctx context.Context, s *Service, command, key string, path []string, query url.Values) ([]*T, *resty.Response, error) {
	response, err := s.get(ctx, command, path, query)
	if err != nil {
		return nil, response, err
	}
	values, err := decodeEnvelope[T](response, key)
	if err != nil {
		return nil, response, err
	}
	return values, response, nil
}

func readOne[T any](ctx context.Context, s *Service, command string, path []string) (*T, *resty.Response, error) {
	response, err := s.get(ctx, command, path, nil)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeOne[T](response)
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

func (s *Service) checkParts(ctx context.Context, parts ...string) error {
	if err := s.check(ctx); err != nil {
		return err
	}
	for _, part := range parts {
		if !validPart(part) {
			return ErrInput
		}
	}
	return nil
}

// ListErrorReports reads Shifts connector synchronization errors
// (Get-CsTeamsShiftsConnectionErrorReport). Nil options read every report.
func (s *Service) ListErrorReports(ctx context.Context, options *ErrorReportListOptions) ([]*ErrorReport, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	query := url.Values{}
	if options != nil {
		for key, value := range map[string]*string{
			"activeness": options.Activeness, "connectorInstanceId": options.ConnectorInstanceID, "connectionId": options.ConnectionID,
			"teamId": options.TeamID, "operation": options.Operation, "procedure": options.Procedure, "code": options.Code,
			"before": options.Before, "after": options.After,
		} {
			if value == nil {
				continue
			}
			if !validText(*value) {
				return nil, nil, ErrInput
			}
			query.Set(key, *value)
		}
	}
	if len(query) == 0 {
		query = nil
	}
	return readList[ErrorReport](ctx, s, "Get-CsTeamsShiftsConnectionErrorReport", "errorReports", path("reports", "errors"), query)
}

// GetErrorReportByID reads one synchronization error
// (Get-CsTeamsShiftsConnectionErrorReport -ErrorReportId).
func (s *Service) GetErrorReportByID(ctx context.Context, id string) (*ErrorReport, *resty.Response, error) {
	if err := s.checkParts(ctx, id); err != nil {
		return nil, nil, err
	}
	return readOne[ErrorReport](ctx, s, "Get-CsTeamsShiftsConnectionErrorReport", path("reports", "errors", id))
}

// GetOperationByID reads an asynchronous connector operation, such as a
// batch team mapping (Get-CsTeamsShiftsConnectionOperation).
func (s *Service) GetOperationByID(ctx context.Context, id string) (*Operation, *resty.Response, error) {
	if err := s.checkParts(ctx, id); err != nil {
		return nil, nil, err
	}
	return readOne[Operation](ctx, s, "Get-CsTeamsShiftsConnectionOperation", pathV1("operations", id))
}

// GetSyncResult reads the user synchronization result of one team mapped by
// a connector instance (Get-CsTeamsShiftsConnectionSyncResult).
func (s *Service) GetSyncResult(ctx context.Context, instanceID, teamID string) (*SyncResult, *resty.Response, error) {
	if err := s.checkParts(ctx, instanceID, teamID); err != nil {
		return nil, nil, err
	}
	return readOne[SyncResult](ctx, s, "Get-CsTeamsShiftsConnectionSyncResult", pathV1(instances, instanceID, "usersync", teamID))
}

// ListTeamMaps reads the team mappings of a connector instance
// (Get-CsTeamsShiftsConnectionTeamMap).
func (s *Service) ListTeamMaps(ctx context.Context, instanceID string) ([]*TeamMap, *resty.Response, error) {
	if err := s.checkParts(ctx, instanceID); err != nil {
		return nil, nil, err
	}
	return readList[TeamMap](ctx, s, "Get-CsTeamsShiftsConnectionTeamMap", "teamConnects", pathV1(instances, instanceID, "teamconnects"), nil)
}

// ListInstanceWfmTeams reads the WFM teams a connector instance can map
// (Get-CsTeamsShiftsConnectionWfmTeam -ConnectorInstanceId).
func (s *Service) ListInstanceWfmTeams(ctx context.Context, instanceID string) ([]*WfmTeam, *resty.Response, error) {
	if err := s.checkParts(ctx, instanceID); err != nil {
		return nil, nil, err
	}
	return readList[WfmTeam](ctx, s, "Get-CsTeamsShiftsConnectionWfmTeam", "wfmTeams", pathV1(instances, instanceID, "wfmteams"), nil)
}

// ListConnectionWfmTeams reads the WFM teams of a connection
// (Get-CsTeamsShiftsConnectionWfmTeam -ConnectionId).
func (s *Service) ListConnectionWfmTeams(ctx context.Context, connectionID string) ([]*WfmTeam, *resty.Response, error) {
	if err := s.checkParts(ctx, connectionID); err != nil {
		return nil, nil, err
	}
	return readList[WfmTeam](ctx, s, "Get-CsTeamsShiftsConnectionWfmTeam", "wfmTeams", path(connections, connectionID, "wfmteams"), nil)
}

// ListWfmUsers reads the users of a WFM team through a connector instance
// (Get-CsTeamsShiftsConnectionWfmUser).
func (s *Service) ListWfmUsers(ctx context.Context, instanceID, wfmTeamID string) ([]*User, *resty.Response, error) {
	if err := s.checkParts(ctx, instanceID, wfmTeamID); err != nil {
		return nil, nil, err
	}
	return readList[User](ctx, s, "Get-CsTeamsShiftsConnectionWfmUser", "wfmUsers", pathV1(instances, instanceID, "wfmteams", wfmTeamID), nil)
}

// MapTeams maps Teams teams to WFM teams for a connector instance
// (New-CsTeamsShiftsConnectionBatchTeamMap, PUT). The service accepts the
// batch with HTTP 202 and runs it asynchronously; poll the returned operation
// with GetOperationByID. It is never replayed automatically.
func (s *Service) MapTeams(ctx context.Context, instanceID string, mappings []TeamMapping) (*TeamMapOperation, *resty.Response, error) {
	if err := s.checkParts(ctx, instanceID); err != nil {
		return nil, nil, err
	}
	if len(mappings) == 0 {
		return nil, nil, ErrInput
	}
	for _, mapping := range mappings {
		if !validText(mapping.TeamID) || !validText(mapping.WfmTeamID) || !optionalText(mapping.TimeZone) {
			return nil, nil, ErrInput
		}
	}
	body := map[string]any{"teamMappings": mappings}
	response, err := s.do(ctx, client.Request{Method: http.MethodPut, Path: pathV1(instances, instanceID, "teamconnects"),
		Header: headers("New-CsTeamsShiftsConnectionBatchTeamMap", nil), Body: body}, http.StatusAccepted)
	if err != nil {
		return nil, response, err
	}
	value, err := decodeOne[TeamMapOperation](response)
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// DeleteTeamMap removes the mapping of one Teams team from a connector
// instance (Remove-CsTeamsShiftsConnectionTeamMap). The service accepts the
// removal with HTTP 202.
func (s *Service) DeleteTeamMap(ctx context.Context, instanceID, teamID string) (*resty.Response, error) {
	if err := s.checkParts(ctx, instanceID, teamID); err != nil {
		return nil, err
	}
	return s.do(ctx, client.Request{Method: http.MethodDelete, Path: pathV1(instances, instanceID, "teamconnects", teamID),
		Header: headers("Remove-CsTeamsShiftsConnectionTeamMap", nil)}, http.StatusAccepted)
}

// ClearSchedule permanently deletes Shifts schedule records of the given
// entity types from a team (Remove-CsTeamsShiftsScheduleRecord, POST
// clearschedule/clear). It cannot be undone and is never replayed
// automatically.
func (s *Service) ClearSchedule(ctx context.Context, request ClearScheduleRequest) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if !validText(request.TeamID) || len(request.EntityTypes) == 0 || !optionalText(request.DesignatedActorID) || !optionalText(request.TimeZone) {
		return nil, ErrInput
	}
	for _, entity := range request.EntityTypes {
		if !validText(entity) {
			return nil, ErrInput
		}
	}
	body := map[string]any{"teamId": request.TeamID, "clearSchedulingGroup": request.ClearSchedulingGroup, "entityTypes": request.EntityTypes}
	if request.StartDate != nil || request.EndDate != nil {
		dates := map[string]string{}
		if request.StartDate != nil {
			dates["startDate"] = request.StartDate.Format(scheduleDateLayout)
		}
		if request.EndDate != nil {
			dates["endDate"] = request.EndDate.Format(scheduleDateLayout)
		}
		body["dateRange"] = dates
	}
	if request.DesignatedActorID != nil {
		body["designatedActorId"] = *request.DesignatedActorID
	}
	if request.TimeZone != nil {
		body["timeZone"] = *request.TimeZone
	}
	return s.do(ctx, client.Request{Method: http.MethodPost, Path: pathV1("clearschedule", "clear"),
		Header: headers("Remove-CsTeamsShiftsScheduleRecord", nil), Body: body}, http.StatusOK)
}
