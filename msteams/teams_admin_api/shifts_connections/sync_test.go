package shiftsconnections_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	sc "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/shifts_connections"
	"github.com/go-resty/resty/v2"
)

const (
	v1   = "/Teams.ShiftsConnector/api/v1"
	team = "00000000-0000-4000-8000-000000000002"
)

// noValue reports whether a returned value is absent (nil interface or nil
// pointer or slice), whatever its type.
func noValue(v any) bool {
	if v == nil {
		return true
	}
	value := reflect.ValueOf(v)
	switch value.Kind() {
	case reflect.Pointer, reflect.Slice, reflect.Map:
		return value.IsNil()
	}
	return false
}

func TestUnitShiftsSyncAuthorizationDenialUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	reports := readFile(t, "testdata/list_error_reports_403.json")
	maps := readFile(t, "testdata/list_team_maps_403.json")
	mapped := readFile(t, "testdata/map_teams_403.json")
	zone := "America/Los_Angeles"
	start, end := time.Date(2026, 1, 1, 0, 0, 0, 0, time.UTC), time.Date(2026, 1, 2, 0, 0, 0, 0, time.UTC)
	cases := []struct {
		name, method, path, command, body, fixture string
		reply                                      []byte
		mappings                                   []contracttest.Mapping
		invoke                                     invocation
	}{
		{"error-reports", "GET", v2 + "/reports/errors", "Get-CsTeamsShiftsConnectionErrorReport", "", "list_error_reports_403", reports,
			mapping("Get-CsTeamsShiftsConnectionErrorReport", "Get"),
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.ListErrorReports(ctx, nil)
			}},
		{"error-reports-filtered", "GET", v2 + "/reports/errors?activeness=Active&connectorInstanceId=" + probe + "&teamId=" + team,
			"Get-CsTeamsShiftsConnectionErrorReport", "", "list_error_reports_403", reports, mapping("Get-CsTeamsShiftsConnectionErrorReport", "Get"),
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.ListErrorReports(ctx, &sc.ErrorReportListOptions{Activeness: ptr("Active"), ConnectorInstanceID: ptr(probe), TeamID: ptr(team)})
			}},
		{"error-report", "GET", v2 + "/reports/errors/" + probe, "Get-CsTeamsShiftsConnectionErrorReport", "", "list_error_reports_403", reports,
			mapping("Get-CsTeamsShiftsConnectionErrorReport", "Get1", "GetViaIdentity"),
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.GetErrorReportByID(ctx, probe)
			}},
		{"operation", "GET", v1 + "/operations/" + probe, "Get-CsTeamsShiftsConnectionOperation", "", "list_error_reports_403", reports,
			mapping("Get-CsTeamsShiftsConnectionOperation", "Get", "GetViaIdentity"),
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.GetOperationByID(ctx, probe)
			}},
		{"sync-result", "GET", v1 + "/connectorinstances/" + probe + "/usersync/" + team, "Get-CsTeamsShiftsConnectionSyncResult", "", "list_error_reports_403", reports,
			mapping("Get-CsTeamsShiftsConnectionSyncResult", "Get", "GetViaIdentity"),
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.GetSyncResult(ctx, probe, team)
			}},
		{"team-maps", "GET", v1 + "/connectorinstances/" + probe + "/teamconnects", "Get-CsTeamsShiftsConnectionTeamMap", "", "list_team_maps_403", maps,
			mapping("Get-CsTeamsShiftsConnectionTeamMap", "Get", "GetViaIdentity"),
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.ListTeamMaps(ctx, probe)
			}},
		{"instance-wfm-teams", "GET", v1 + "/connectorinstances/" + probe + "/wfmteams", "Get-CsTeamsShiftsConnectionWfmTeam", "", "list_team_maps_403", maps,
			mapping("Get-CsTeamsShiftsConnectionWfmTeam", "Get", "GetViaIdentity"),
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.ListInstanceWfmTeams(ctx, probe)
			}},
		{"connection-wfm-teams", "GET", v2 + "/connections/" + probe + "/wfmteams", "Get-CsTeamsShiftsConnectionWfmTeam", "", "list_team_maps_403", maps,
			mapping("Get-CsTeamsShiftsConnectionWfmTeam", "Get1", "GetViaIdentity1"),
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.ListConnectionWfmTeams(ctx, probe)
			}},
		{"wfm-users", "GET", v1 + "/connectorinstances/" + probe + "/wfmteams/msteams-sdk-probe", "Get-CsTeamsShiftsConnectionWfmUser", "", "list_team_maps_403", maps,
			mapping("Get-CsTeamsShiftsConnectionWfmUser", "Get", "GetViaIdentity"),
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.ListWfmUsers(ctx, probe, "msteams-sdk-probe")
			}},
		{"map-teams", "PUT", v1 + "/connectorinstances/" + probe + "/teamconnects", "New-CsTeamsShiftsConnectionBatchTeamMap",
			`{"teamMappings":[{"teamId":"` + team + `","wfmTeamId":"msteams-sdk-probe","timeZone":"America/Los_Angeles"}]}`, "map_teams_403", mapped,
			mapping("New-CsTeamsShiftsConnectionBatchTeamMap", "Create", "CreateExpanded", "CreateViaIdentity", "CreateViaIdentityExpanded"),
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.MapTeams(ctx, probe, []sc.TeamMapping{{TeamID: team, WfmTeamID: "msteams-sdk-probe", TimeZone: &zone}})
			}},
		{"delete-team-map", "DELETE", v1 + "/connectorinstances/" + probe + "/teamconnects/" + team, "Remove-CsTeamsShiftsConnectionTeamMap", "", "map_teams_403", mapped,
			mapping("Remove-CsTeamsShiftsConnectionTeamMap", "Remove", "RemoveViaIdentity"),
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				response, err := s.DeleteTeamMap(ctx, probe, team)
				return nil, response, err
			}},
		{"clear-schedule", "POST", v1 + "/clearschedule/clear", "Remove-CsTeamsShiftsScheduleRecord",
			`{"clearSchedulingGroup":false,"dateRange":{"endDate":"2026-01-02T00:00:00.0000000Z","startDate":"2026-01-01T00:00:00.0000000Z"},"entityTypes":["shift"],"teamId":"` + team + `","timeZone":"America/Los_Angeles"}`,
			"map_teams_403", mapped, mapping("Remove-CsTeamsShiftsScheduleRecord", "Remove", "RemoveExpanded"),
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				response, err := s.ClearSchedule(ctx, sc.ClearScheduleRequest{TeamID: team, EntityTypes: []string{"shift"}, StartDate: &start, EndDate: &end, TimeZone: &zone})
				return nil, response, err
			}},
	}
	for _, c := range cases {
		var body json.RawMessage
		if c.body != "" {
			body = json.RawMessage(c.body)
		}
		run(t, contracttest.Case{Name: "actual-denied-" + c.name, Source: "msteams/teams_admin_api/shifts_connections/testdata/" + c.fixture + ".json", Mappings: c.mappings,
			Steps: []contracttest.Exchange{discovery(t), wire(c.method, c.path, c.command, "", body, 403, c.reply)},
			Want:  contracttest.Expected{Status: 403, Body: c.reply, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 403, Codes: []string{"Forbidden"}}},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.ShiftsConnections)
				if !errors.Is(err, auth.ErrForbidden) || !noValue(value) {
					t.Fatal("denial was not a typed forbidden error without a value")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
}

const (
	errorReportJSON = `{"ttl":3600,"id":"` + probe + `","tenantId":null,"wfmConnectorInstanceId":"` + probe + `","connectionId":"` + probe + `","teamId":"` + team + `",` +
		`"operation":"Sync","procedure":"ShiftSync","code":"WfmUserNotFound","message":"Fixture message","parameters":["a","b"],"referenceLinks":{"help":"https://example.invalid/help"},` +
		`"revisitIntervalInMinutes":15,"errorNotificationSent":true,"resolvedAt":null,"resolvedNotificationSentOn":null,"revisitedAt":"2026-01-01T00:00:00.0000000Z",` +
		`"totalIncidents":3,"intermediateIncidents":1,"createdAt":"2026-01-01T00:00:00.0000000Z","culture":"en-US","scheduleSequenceNumber":9007199254740993,` +
		`"severity":"Error","errorType":"User","future":1}`
	operationJSON = `{"Id":"` + probe + `","type":"TeamMapping","tenantId":null,"wfmConnectorInstanceId":"` + probe + `","createdDateTime":"2026-01-01T00:00:00Z",` +
		`"lastActionDateTime":null,"status":"Completed","details":[{"status":"Failed","error":"WfmTeamNotFound","properties":{"teamId":"` + team + `","attempt":2}}]}`
	syncResultJSON = `{"successfulUsers":[{"userId":"` + probe + `","userName":"user@example.invalid","wfmUserId":"w1","wfmUserName":"WFM user"}],` +
		`"failedWfmUsers":[{"id":"w2","name":"Unmatched"}],"failedAadUsers":[]}`
	teamMapsJSON  = `{"teamConnects":[{"teamId":"` + team + `","teamName":"Store","wfmTeamId":"w-team","wfmTeamName":"WFM store","timeZone":"America/Los_Angeles","state":"Active"}]}`
	wfmTeamsJSON  = `{"wfmTeams":[{"id":"w-team","name":"WFM store"}]}`
	connTeamsJSON = `{"wfmTeams":[{"id":"w-team","name":"WFM store","connectorInstanceId":"` + probe + `","teamId":null}]}`
	wfmUsersJSON  = `{"wfmUsers":[{"id":"w1","name":"WFM user"}]}`
	teamMapOpJSON = `{"operationId":"` + probe + `","createdDateTime":"2026-01-01T00:00:00.0000000Z","lastActionDateTime":null,"status":"NotStarted"}`
)

func TestUnitShiftsSyncSuccessShapesPreserveFields(t *testing.T) {
	t.Parallel()
	reports := []byte(`{"errorReports":[` + errorReportJSON + `]}`)
	query := "?activeness=All&after=2026-01-01T00%3A00%3A00Z&before=2026-01-02T00%3A00%3A00Z&code=c&connectionId=" + probe +
		"&connectorInstanceId=" + probe + "&operation=o&procedure=p&teamId=" + team
	run(t, contracttest.Case{Name: "success-reads",
		Mappings: slices.Concat(mapping("Get-CsTeamsShiftsConnectionErrorReport", "Get", "Get1"),
			mapping("Get-CsTeamsShiftsConnectionOperation", "Get"), mapping("Get-CsTeamsShiftsConnectionSyncResult", "Get"),
			mapping("Get-CsTeamsShiftsConnectionTeamMap", "Get"), mapping("Get-CsTeamsShiftsConnectionWfmTeam", "Get", "Get1"),
			mapping("Get-CsTeamsShiftsConnectionWfmUser", "Get")),
		Steps: []contracttest.Exchange{discovery(t),
			wire("GET", v2+"/reports/errors"+query, "Get-CsTeamsShiftsConnectionErrorReport", "", nil, 200, reports),
			wire("GET", v2+"/reports/errors/"+probe, "Get-CsTeamsShiftsConnectionErrorReport", "", nil, 200, []byte(errorReportJSON)),
			wire("GET", v1+"/operations/"+probe, "Get-CsTeamsShiftsConnectionOperation", "", nil, 200, []byte(operationJSON)),
			wire("GET", v1+"/connectorinstances/"+probe+"/usersync/"+team, "Get-CsTeamsShiftsConnectionSyncResult", "", nil, 200, []byte(syncResultJSON)),
			wire("GET", v1+"/connectorinstances/"+probe+"/teamconnects", "Get-CsTeamsShiftsConnectionTeamMap", "", nil, 200, []byte(teamMapsJSON)),
			wire("GET", v1+"/connectorinstances/"+probe+"/wfmteams", "Get-CsTeamsShiftsConnectionWfmTeam", "", nil, 200, []byte(wfmTeamsJSON)),
			wire("GET", v2+"/connections/"+probe+"/wfmteams", "Get-CsTeamsShiftsConnectionWfmTeam", "", nil, 200, []byte(connTeamsJSON)),
			wire("GET", v1+"/connectorinstances/"+probe+"/wfmteams/w-team", "Get-CsTeamsShiftsConnectionWfmUser", "", nil, 200, []byte(wfmUsersJSON))},
		Want: contracttest.Expected{Status: 200, Body: []byte(wfmUsersJSON)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.ShiftsConnections
			listed, _, err := s.ListErrorReports(ctx, &sc.ErrorReportListOptions{Activeness: ptr("All"), ConnectorInstanceID: ptr(probe), ConnectionID: ptr(probe),
				TeamID: ptr(team), Operation: ptr("o"), Procedure: ptr("p"), Code: ptr("c"), Before: ptr("2026-01-02T00:00:00Z"), After: ptr("2026-01-01T00:00:00Z")})
			if err != nil || len(listed) != 1 || listed[0].RawFields["future"] == nil || listed[0].TenantID != nil || *listed[0].ScheduleSequenceNumber != 9007199254740993 {
				t.Fatal("error report list lost null, unknown or 64-bit fields")
			}
			report, _, err := s.GetErrorReportByID(ctx, probe)
			if err != nil || *report.TotalIncidents != 3 || !*report.ErrorNotificationSent || report.Parameters[1] != "b" || *report.ReferenceLinks["help"] != "https://example.invalid/help" || report.ResolvedAt != nil {
				t.Fatal("error report detail was not decoded")
			}
			operation, _, err := s.GetOperationByID(ctx, probe)
			if err != nil || *operation.ID != probe || *operation.Status != "Completed" || *operation.Details[0].Error != "WfmTeamNotFound" || string(operation.Details[0].Properties["attempt"]) != "2" {
				t.Fatal("operation was not decoded")
			}
			result, _, err := s.GetSyncResult(ctx, probe, team)
			if err != nil || *result.SuccessfulUsers[0].WfmUserID != "w1" || *result.FailedWfmUsers[0].Name != "Unmatched" || result.FailedAadUsers == nil || len(result.FailedAadUsers) != 0 {
				t.Fatal("sync result was not decoded")
			}
			maps, _, err := s.ListTeamMaps(ctx, probe)
			if err != nil || len(maps) != 1 || *maps[0].WfmTeamName != "WFM store" || *maps[0].State != "Active" {
				t.Fatal("team maps were not decoded")
			}
			teams, _, err := s.ListInstanceWfmTeams(ctx, probe)
			if err != nil || len(teams) != 1 || *teams[0].ID != "w-team" || teams[0].ConnectorInstanceID != nil {
				t.Fatal("instance WFM teams were not decoded")
			}
			teams, _, err = s.ListConnectionWfmTeams(ctx, probe)
			if err != nil || len(teams) != 1 || *teams[0].ConnectorInstanceID != probe || teams[0].TeamID != nil {
				t.Fatal("connection WFM teams were not decoded")
			}
			users, response, err := s.ListWfmUsers(ctx, probe, "w-team")
			if err != nil || len(users) != 1 || *users[0].Name != "WFM user" {
				t.Fatal("WFM users were not decoded")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})

	mapBody := json.RawMessage(`{"teamMappings":[{"teamId":"` + team + `","wfmTeamId":"w-team","timeZone":"UTC"},{"teamId":"` + probe + `","wfmTeamId":"w-team-2"}]}`)
	clearBody := json.RawMessage(`{"teamId":"` + team + `","clearSchedulingGroup":true,"entityTypes":["openShift","timeOff"],"designatedActorId":"` + probe + `"}`)
	startOnly := json.RawMessage(`{"teamId":"` + team + `","clearSchedulingGroup":false,"entityTypes":["shift"],"dateRange":{"startDate":"2026-01-01T08:30:00.0000000-08:00"}}`)
	run(t, contracttest.Case{Name: "success-writes",
		Mappings: slices.Concat(mapping("New-CsTeamsShiftsConnectionBatchTeamMap", "CreateExpanded"), mapping("Remove-CsTeamsShiftsConnectionTeamMap", "Remove"),
			mapping("Remove-CsTeamsShiftsScheduleRecord", "RemoveExpanded")),
		Steps: []contracttest.Exchange{discovery(t),
			wire("PUT", v1+"/connectorinstances/"+probe+"/teamconnects", "New-CsTeamsShiftsConnectionBatchTeamMap", "", mapBody, 202, []byte(teamMapOpJSON)),
			wire("DELETE", v1+"/connectorinstances/"+probe+"/teamconnects/"+team, "Remove-CsTeamsShiftsConnectionTeamMap", "", nil, 202, nil),
			wire("POST", v1+"/clearschedule/clear", "Remove-CsTeamsShiftsScheduleRecord", "", clearBody, 200, nil),
			wire("POST", v1+"/clearschedule/clear", "Remove-CsTeamsShiftsScheduleRecord", "", startOnly, 200, nil)},
		Want: contracttest.Expected{Status: 200},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.ShiftsConnections
			operation, response, err := s.MapTeams(ctx, probe, []sc.TeamMapping{{TeamID: team, WfmTeamID: "w-team", TimeZone: ptr("UTC")}, {TeamID: probe, WfmTeamID: "w-team-2"}})
			if err != nil || response.StatusCode() != 202 || *operation.OperationID != probe || *operation.Status != "NotStarted" || operation.LastActionDateTime != nil {
				t.Fatal("batch team map receipt was not decoded")
			}
			if response, err := s.DeleteTeamMap(ctx, probe, team); err != nil || response.StatusCode() != 202 {
				t.Fatal("team map removal was not accepted")
			}
			if _, err := s.ClearSchedule(ctx, sc.ClearScheduleRequest{TeamID: team, EntityTypes: []string{"openShift", "timeOff"}, ClearSchedulingGroup: true, DesignatedActorID: ptr(probe)}); err != nil {
				t.Fatal("schedule clear without a date range failed")
			}
			start := time.Date(2026, 1, 1, 8, 30, 0, 0, time.FixedZone("PST", -8*3600))
			response, err = s.ClearSchedule(ctx, sc.ClearScheduleRequest{TeamID: team, EntityTypes: []string{"shift"}, StartDate: &start})
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitShiftsSyncMalformedResponsesFailDecode(t *testing.T) {
	t.Parallel()
	for name, c := range map[string]struct {
		path, command, body string
		invoke              invocation
	}{
		"error-reports-value-envelope": {v2 + "/reports/errors", "Get-CsTeamsShiftsConnectionErrorReport", `{"value":[]}`,
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.ListErrorReports(ctx, nil)
			}},
		"team-maps-null": {v1 + "/connectorinstances/x/teamconnects", "Get-CsTeamsShiftsConnectionTeamMap", `{"teamConnects":null}`,
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.ListTeamMaps(ctx, "x")
			}},
		"wfm-teams-nil-item": {v1 + "/connectorinstances/x/wfmteams", "Get-CsTeamsShiftsConnectionWfmTeam", `{"wfmTeams":[null]}`,
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.ListInstanceWfmTeams(ctx, "x")
			}},
		"wfm-users-array": {v1 + "/connectorinstances/x/wfmteams/w", "Get-CsTeamsShiftsConnectionWfmUser", `[]`,
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.ListWfmUsers(ctx, "x", "w")
			}},
		"operation-null": {v1 + "/operations/x", "Get-CsTeamsShiftsConnectionOperation", `null`,
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.GetOperationByID(ctx, "x")
			}},
		"report-bad-type": {v2 + "/reports/errors/x", "Get-CsTeamsShiftsConnectionErrorReport", `{"totalIncidents":"three"}`,
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.GetErrorReportByID(ctx, "x")
			}},
		"sync-result-bad-users": {v1 + "/connectorinstances/x/usersync/t", "Get-CsTeamsShiftsConnectionSyncResult", `{"successfulUsers":{}}`,
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.GetSyncResult(ctx, "x", "t")
			}},
	} {
		run(t, contracttest.Case{Name: name, Steps: []contracttest.Exchange{discovery(t), wire("GET", c.path, c.command, "", nil, 200, []byte(c.body))},
			Want: contracttest.Expected{Status: 200, Error: client.ErrDecode},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.ShiftsConnections)
				if !noValue(value) {
					t.Fatal("malformed response returned a value")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
	body := json.RawMessage(`{"teamMappings":[{"teamId":"t","wfmTeamId":"w"}]}`)
	run(t, contracttest.Case{Name: "map-teams-ok-is-undeclared",
		Steps: []contracttest.Exchange{discovery(t), wire("PUT", v1+"/connectorinstances/x/teamconnects", "New-CsTeamsShiftsConnectionBatchTeamMap", "", body, 200, []byte(teamMapOpJSON))},
		Want:  contracttest.Expected{Status: 200, Error: sc.ErrResponse},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.ShiftsConnections.MapTeams(ctx, "x", []sc.TeamMapping{{TeamID: "t", WfmTeamID: "w"}})
			if value != nil {
				t.Fatal("undeclared status returned a value")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
	run(t, contracttest.Case{Name: "map-teams-empty-receipt",
		Steps: []contracttest.Exchange{discovery(t), wire("PUT", v1+"/connectorinstances/x/teamconnects", "New-CsTeamsShiftsConnectionBatchTeamMap", "", body, 202, []byte(`null`))},
		Want:  contracttest.Expected{Status: 202, Error: client.ErrDecode},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, response, err := sdk.TeamsAdminAPI.ShiftsConnections.MapTeams(ctx, "x", []sc.TeamMapping{{TeamID: "t", WfmTeamID: "w"}})
			return contracttest.Observation{Response: response, Err: err}
		}})
	run(t, contracttest.Case{Name: "delete-team-map-no-content-is-undeclared",
		Steps: []contracttest.Exchange{discovery(t), wire("DELETE", v1+"/connectorinstances/x/teamconnects/t", "Remove-CsTeamsShiftsConnectionTeamMap", "", nil, 204, nil)},
		Want:  contracttest.Expected{Status: 204, Error: sc.ErrResponse},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.ShiftsConnections.DeleteTeamMap(ctx, "x", "t")
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitShiftsSyncInvalidInputFailsBeforeIO(t *testing.T) {
	t.Parallel()
	for name, call := range map[string]func(context.Context, *sc.Service) error{
		"error-report-blank": func(ctx context.Context, s *sc.Service) error { _, _, e := s.GetErrorReportByID(ctx, " "); return e },
		"error-reports-control-filter": func(ctx context.Context, s *sc.Service) error {
			_, _, e := s.ListErrorReports(ctx, &sc.ErrorReportListOptions{Code: ptr("a\nb")})
			return e
		},
		"error-reports-empty-filter": func(ctx context.Context, s *sc.Service) error {
			_, _, e := s.ListErrorReports(ctx, &sc.ErrorReportListOptions{TeamID: ptr("")})
			return e
		},
		"operation-slash":     func(ctx context.Context, s *sc.Service) error { _, _, e := s.GetOperationByID(ctx, "a/b"); return e },
		"sync-result-no-team": func(ctx context.Context, s *sc.Service) error { _, _, e := s.GetSyncResult(ctx, "x", ""); return e },
		"team-maps-traversal": func(ctx context.Context, s *sc.Service) error { _, _, e := s.ListTeamMaps(ctx, ".."); return e },
		"instance-teams-dot":  func(ctx context.Context, s *sc.Service) error { _, _, e := s.ListInstanceWfmTeams(ctx, "."); return e },
		"connection-teams-bad": func(ctx context.Context, s *sc.Service) error {
			_, _, e := s.ListConnectionWfmTeams(ctx, `a\b`)
			return e
		},
		"wfm-users-no-team": func(ctx context.Context, s *sc.Service) error { _, _, e := s.ListWfmUsers(ctx, "x", " "); return e },
		"map-teams-empty":   func(ctx context.Context, s *sc.Service) error { _, _, e := s.MapTeams(ctx, "x", nil); return e },
		"map-teams-no-wfm-team": func(ctx context.Context, s *sc.Service) error {
			_, _, e := s.MapTeams(ctx, "x", []sc.TeamMapping{{TeamID: "t"}})
			return e
		},
		"map-teams-control-zone": func(ctx context.Context, s *sc.Service) error {
			_, _, e := s.MapTeams(ctx, "x", []sc.TeamMapping{{TeamID: "t", WfmTeamID: "w", TimeZone: ptr("a\tb")}})
			return e
		},
		"delete-team-map-blank": func(ctx context.Context, s *sc.Service) error { _, e := s.DeleteTeamMap(ctx, "x", ""); return e },
		"clear-no-team": func(ctx context.Context, s *sc.Service) error {
			_, e := s.ClearSchedule(ctx, sc.ClearScheduleRequest{EntityTypes: []string{"shift"}})
			return e
		},
		"clear-no-entities": func(ctx context.Context, s *sc.Service) error {
			_, e := s.ClearSchedule(ctx, sc.ClearScheduleRequest{TeamID: "t"})
			return e
		},
		"clear-blank-entity": func(ctx context.Context, s *sc.Service) error {
			_, e := s.ClearSchedule(ctx, sc.ClearScheduleRequest{TeamID: "t", EntityTypes: []string{"shift", " "}})
			return e
		},
	} {
		run(t, contracttest.Case{Name: name, Scope: "local", Want: contracttest.Expected{Error: sc.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return contracttest.Observation{Err: call(ctx, sdk.TeamsAdminAPI.ShiftsConnections)}
			}})
	}
}

func TestUnitShiftsSyncNilContextAndServiceAreGuarded(t *testing.T) {
	t.Parallel()
	var s *sc.Service
	//nolint:staticcheck // A nil context is the guarded input.
	if _, _, err := s.ListErrorReports(nil, nil); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, err := s.ClearSchedule(t.Context(), sc.ClearScheduleRequest{TeamID: "t", EntityTypes: []string{"shift"}}); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, _, err := sc.NewService(nil).MapTeams(ctx, "x", []sc.TeamMapping{{TeamID: "t", WfmTeamID: "w"}}); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled context accepted")
	}
}
