package migration_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/migration"
	"github.com/go-resty/resty/v2"
)

const (
	configOrigin = "https://api.interfaces.records.teams.microsoft.com"
	targetForest = "https://admin.example.invalid/"
	source       = "msteams/teams_admin_api/migration/service_test.go"
	probe        = "00000000-0000-4000-8000-000000000001"
	partition    = probe + "_msteams-sdk-probe"
	jsonType     = "application/json; charset=utf-8"
)

func readFile(t *testing.T, path string) []byte {
	t.Helper()
	root, err := os.OpenRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	raw, err := root.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func discovery(t *testing.T) contracttest.Exchange {
	t.Helper()
	body, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: http.MethodGet, URL: configOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience},
		Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body}}
}

func wire(method, path, command string, body json.RawMessage, status int, reply []byte) contracttest.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {targetForest}}
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	contentType := jsonType
	if len(reply) == 0 {
		contentType, reply = "", nil
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: configOrigin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body},
		Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {contentType}}, Body: reply}}
}

func run(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family = "migration"
	if c.Source == "" {
		c.Source = source
	}
	if c.Scope == "" {
		c.Scope = "service"
	}
	result, err := contracttest.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Fatalf("migration contract %s differed: %+v", c.Name, result.Differences())
	}
}

func mapping(command, set string) []contracttest.Mapping {
	return []contracttest.Mapping{{Command: command, ParameterSet: set}}
}

// apiError is the normalized error for a fixture: ConfigAPI reports the code
// at the top level or inside an "error" object.
func apiError(status int, body []byte) *client.APIError {
	var envelope struct {
		Code  string `json:"code"`
		Error struct {
			Code string `json:"code"`
		} `json:"error"`
	}
	_ = json.Unmarshal(body, &envelope)
	result := &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}
	for _, code := range []string{envelope.Code, envelope.Error.Code} {
		if code != "" {
			result.Codes = append(result.Codes, code)
		}
	}
	return result
}

func ptr[T any](v T) *T { return &v }

type invocation func(context.Context, *migration.Service) (any, *resty.Response, error)

// TestUnitMigrationDeniedOperationsUseAcceptanceResponses replays every
// response the app-only lab identity received: HTTP 401 with no body from the
// meeting migration service and HTTP 403 from the tenant and AVS services.
func TestUnitMigrationDeniedOperationsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	avsBody := `{"BasePartitionKey":"` + probe + `","ContainerName":"msteams-sdk-probe","NumberOfDocuments":0,"SourcePartitionKey":"` + partition +
		`-source","TargetPartitionKey":"` + partition + `-target","Workload":"msteams-sdk-probe"}`
	for _, c := range []struct {
		name, method, path, command, set, fixture, body string
		status                                          int
		invoke                                          invocation
	}{
		{"status", "GET", "/Teams.Mms/GetMeetingMigrationStatus", "Get-CsMeetingMigrationStatus", "__AllParameterSets", "status_401.body", "", 401,
			func(ctx context.Context, s *migration.Service) (any, *resty.Response, error) {
				return s.ListMeetingMigrationStatuses(ctx, nil)
			}},
		{"summary", "GET", "/Teams.Mms/GetMeetingMigrationStatusSummary", "Get-CsMeetingMigrationStatus", "__AllParameterSets", "summary_401.body", "", 401,
			func(ctx context.Context, s *migration.Service) (any, *resty.Response, error) {
				return s.GetMeetingMigrationStatusSummary(ctx, nil)
			}},
		{"history", "GET", "/Teams.Mms/GetMeetingMigrationTransactionHistory?userIdentity=" + probe, "Get-CsMeetingMigrationTransactionHistory", "__AllParameterSets", "history_403.json", "", 403,
			func(ctx context.Context, s *migration.Service) (any, *resty.Response, error) {
				return s.ListMeetingMigrationTransactionHistory(ctx, probe, nil)
			}},
		{"task-status", "GET", "/Teams.Tenant/TenantMigrationDetail", "Get-CsMoveTenantServiceInstanceTaskStatus", "__AllParameterSets", "task_status_403.json", "", 403,
			func(ctx context.Context, s *migration.Service) (any, *resty.Response, error) {
				return s.ListTenantMoveTaskStatuses(ctx)
			}},
		{"start", "POST", "/Teams.Mms/StartMeetingMigration", "Start-CsExMeetingMigration", "UserId", "start_meeting_migration_401.body", `{"Identity":"` + probe + `"}`, 401,
			func(ctx context.Context, s *migration.Service) (any, *resty.Response, error) {
				return s.StartMeetingMigration(ctx, migration.StartMeetingMigrationRequest{Identity: probe})
			}},
		{"avs-move", "POST", "/Teams.AVS/MovePartition", "Move-CsAvsTenantPartition", "MoveExpanded", "move_avs_partition_403.json", avsBody, 403,
			func(ctx context.Context, s *migration.Service) (any, *resty.Response, error) {
				return s.MoveAVSTenantPartition(ctx, migration.PartitionMoveRequest{SourcePartitionKey: partition + "-source", TargetPartitionKey: partition + "-target",
					Workload: ptr("msteams-sdk-probe"), ContainerName: ptr("msteams-sdk-probe"), BasePartitionKey: ptr(probe), NumberOfDocuments: ptr(int64(0))})
			}},
	} {
		reply := readFile(t, "testdata/"+c.fixture)
		var body json.RawMessage
		if c.body != "" {
			body = json.RawMessage(c.body)
		}
		want := contracttest.Expected{Status: c.status, APIError: apiError(c.status, reply)}
		if len(reply) != 0 {
			want.Body = reply
		}
		// Empty-body fixtures (.body) cannot be a contract source; the test is.
		fixtureSource := source
		if len(reply) != 0 {
			fixtureSource = "msteams/teams_admin_api/migration/testdata/" + c.fixture
		}
		run(t, contracttest.Case{Name: "actual-" + c.name, Source: fixtureSource, Mappings: mapping(c.command, c.set),
			Steps: []contracttest.Exchange{discovery(t), wire(c.method, c.path, c.command, body, c.status, reply)},
			Want:  want,
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				_, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.Migration)
				wantClass := auth.ErrForbidden
				if c.status == 401 {
					wantClass = auth.ErrUnauthenticated
				}
				if !errors.Is(err, wantClass) {
					t.Fatal("denial was not classified")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
}

func TestUnitMigrationSuccessShapesPreserveFields(t *testing.T) {
	t.Parallel()
	statuses := []byte(`[{"UserId":"` + probe + `","UserPrincipalName":"user@example.invalid","State":"Failed","MigrationType":"AllToTeams","CreateDate":"2026-11-01T10:00:00Z",` +
		`"ModifiedDate":null,"RetryCount":2,"CorrelationId":"c","TotalMeetings":4,"SucceededMeetings":3,"FailedMeetings":1,"InvitesUpdates":3,"LastMessage":"m","EnqueueReason":"Manual","Future":true}]`)
	summary := []byte(`{"MigrationType":"AllToTeams","SummaryOutput":[{"State":"Succeeded","UserCount":5},{"State":"Failed","UserCount":0,"Extra":1}]}`)
	history := []byte(`[{"UserId":"` + probe + `","CorrelationId":"c","Operation":"Migrate","TriggerSource":"Admin","State":"Succeeded","TotalMeetings":1,"InvitesUpdated":1,"LastErrorMessage":null}]`)
	enqueued := []byte(`{"OperationId":"op-1"}`)
	tasks := []byte(`[{"taskName":"PrepForMove","status":"Completed"}]`)
	moved := []byte(`{"TenantId":"` + probe + `","JobName":"job","OrderId":"order"}`)
	avs := []byte(`[{"documentId":"d1","sourcePartitionKey":"s","targetPartitionKey":"t","isSuccessful":false,"lastCompletedStage":"Read","exception":{"ErrorId":7,"CorrelationId":"c","Message":"m"}},` +
		`{"documentId":"d2","isSuccessful":true,"exception":null}]`)
	start := time.Date(2026, 11, 1, 10, 0, 0, 0, time.UTC)
	end := time.Date(2026, 11, 20, 9, 5, 7, 0, time.UTC)
	run(t, contracttest.Case{Name: "success-shapes", Mappings: append(mapping("Get-CsMeetingMigrationStatus", "__AllParameterSets"),
		contracttest.Mapping{Command: "Get-CsMeetingMigrationTransactionHistory", ParameterSet: "__AllParameterSets"},
		contracttest.Mapping{Command: "Start-CsExMeetingMigration", ParameterSet: "UserId"},
		contracttest.Mapping{Command: "Get-CsMoveTenantServiceInstanceTaskStatus", ParameterSet: "__AllParameterSets"},
		contracttest.Mapping{Command: "Move-CsTenantCrossRegion", ParameterSet: "__AllParameterSets"},
		contracttest.Mapping{Command: "Move-CsTenantServiceInstance", ParameterSet: "__AllParameterSets"},
		contracttest.Mapping{Command: "Move-CsAvsTenantPartition", ParameterSet: "MoveExpanded"}),
		Steps: []contracttest.Exchange{discovery(t),
			wire("GET", "/Teams.Mms/GetMeetingMigrationStatus?endTime=11%2F20%2F2026+09%3A05%3A07&identity=user%40example.invalid&migrationType=AllToTeams&startTime=11%2F01%2F2026+10%3A00%3A00&state=Failed",
				"Get-CsMeetingMigrationStatus", nil, 200, statuses),
			wire("GET", "/Teams.Mms/GetMeetingMigrationStatusSummary?state=Failed", "Get-CsMeetingMigrationStatus", nil, 200, summary),
			wire("GET", "/Teams.Mms/GetMeetingMigrationTransactionHistory?CorrelationId=c&startTime=11%2F01%2F2026+10%3A00%3A00&userIdentity=user%40example.invalid",
				"Get-CsMeetingMigrationTransactionHistory", nil, 200, history),
			wire("POST", "/Teams.Mms/StartMeetingMigration", "Start-CsExMeetingMigration",
				json.RawMessage(`{"Identity":"user@example.invalid","SourceMeetingType":"All","TargetMeetingType":"Teams"}`), 201, enqueued),
			wire("POST", "/Teams.Mms/StartMeetingMigration", "Start-CsExMeetingMigration", json.RawMessage(`{"Identity":"user@example.invalid"}`), 204, nil),
			wire("GET", "/Teams.Tenant/TenantMigrationDetail", "Get-CsMoveTenantServiceInstanceTaskStatus", nil, 200, tasks),
			wire("POST", "/Teams.Tenant/CrossRegionMoveTask", "Move-CsTenantCrossRegion", json.RawMessage(`{}`), 200, moved),
			wire("POST", "/Teams.Tenant/CrossRegionMoveTask", "Move-CsTenantServiceInstance", json.RawMessage(`{"MoveOption":"PrepForMove","TargetServiceInstance":""}`), 200, moved),
			wire("POST", "/Teams.Tenant/CrossRegionMoveTask", "Move-CsTenantServiceInstance",
				json.RawMessage(`{"MoveOption":"Finalize","TargetServiceInstance":"MicrosoftCommunicationsOnline/EMEA-01-01"}`), 200, moved),
			wire("POST", "/Teams.AVS/MovePartition", "Move-CsAvsTenantPartition",
				json.RawMessage(`{"BatchSize":10,"PercentageOfPartition":50,"SourcePartitionKey":"s","TargetPartitionKey":"t"}`), 207, avs)},
		Want: contracttest.Expected{Status: 207, Body: avs},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.Migration
			values, _, err := s.ListMeetingMigrationStatuses(ctx, &migration.StatusOptions{Identity: ptr("user@example.invalid"), State: ptr(migration.StateFailed),
				MigrationType: ptr(migration.MigrationTypeAllToTeams), StartTime: &start, EndTime: &end})
			if err != nil || len(values) != 1 || *values[0].RetryCount != 2 || values[0].ModifiedDate != nil || *values[0].InvitesUpdates != 3 || values[0].RawFields["Future"] == nil {
				t.Fatal("status records lost fields")
			}
			counts, _, err := s.GetMeetingMigrationStatusSummary(ctx, &migration.StatusOptions{State: ptr(migration.StateFailed)})
			if err != nil || len(counts.SummaryOutput) != 2 || *counts.SummaryOutput[0].UserCount != 5 || counts.SummaryOutput[1].RawFields["Extra"] == nil {
				t.Fatal("summary was not decoded")
			}
			entries, _, err := s.ListMeetingMigrationTransactionHistory(ctx, "user@example.invalid", &migration.TransactionHistoryOptions{CorrelationID: ptr("c"), StartTime: &start})
			if err != nil || len(entries) != 1 || *entries[0].InvitesUpdated != 1 || entries[0].LastErrorMessage != nil {
				t.Fatal("history was not decoded")
			}
			receipt, response, err := s.StartMeetingMigration(ctx, migration.StartMeetingMigrationRequest{Identity: "user@example.invalid", SourceMeetingType: ptr("All"), TargetMeetingType: ptr("Teams")})
			if err != nil || response.StatusCode() != 201 || *receipt.OperationID != "op-1" {
				t.Fatal("enqueue receipt was not decoded")
			}
			receipt, response, err = s.StartMeetingMigration(ctx, migration.StartMeetingMigrationRequest{Identity: "user@example.invalid"})
			if err != nil || response.StatusCode() != 204 || receipt != nil {
				t.Fatal("no-content start was not a nil receipt")
			}
			statuses, _, err := s.ListTenantMoveTaskStatuses(ctx)
			if err != nil || len(statuses) != 1 || string(statuses[0].Fields["status"]) != `"Completed"` {
				t.Fatal("free-form task status lost fields")
			}
			result, _, err := s.MoveTenantCrossRegion(ctx)
			if err != nil || *result.JobName != "job" {
				t.Fatal("cross-region receipt was not decoded")
			}
			if _, _, err = s.MoveTenantServiceInstance(ctx, migration.MoveTenantServiceInstanceRequest{MoveOption: migration.MoveOptionPrepForMove}); err != nil {
				t.Fatal("omitted target was not sent as an empty string")
			}
			if _, _, err = s.MoveTenantServiceInstance(ctx, migration.MoveTenantServiceInstanceRequest{MoveOption: migration.MoveOptionFinalize,
				TargetServiceInstance: ptr("MicrosoftCommunicationsOnline/EMEA-01-01")}); err != nil {
				t.Fatal("service instance move was not sent")
			}
			results, response, err := s.MoveAVSTenantPartition(ctx, migration.PartitionMoveRequest{SourcePartitionKey: "s", TargetPartitionKey: "t",
				PercentageOfPartition: ptr(int64(50)), BatchSize: ptr(int64(10))})
			if err != nil || len(results) != 2 || *results[0].IsSuccessful || *results[0].Exception.ErrorID != 7 || results[1].Exception != nil || !*results[1].IsSuccessful {
				t.Fatal("multi-status results were not decoded")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitMigrationMalformedResponsesFailDecode(t *testing.T) {
	t.Parallel()
	for name, c := range map[string]struct {
		method, path, command, body, reply string
		status                             int
		invoke                             invocation
	}{
		"status-object": {"GET", "/Teams.Mms/GetMeetingMigrationStatus", "Get-CsMeetingMigrationStatus", "", `{}`, 200,
			func(ctx context.Context, s *migration.Service) (any, *resty.Response, error) {
				return s.ListMeetingMigrationStatuses(ctx, nil)
			}},
		"status-null-item": {"GET", "/Teams.Mms/GetMeetingMigrationStatus", "Get-CsMeetingMigrationStatus", "", `[null]`, 200,
			func(ctx context.Context, s *migration.Service) (any, *resty.Response, error) {
				return s.ListMeetingMigrationStatuses(ctx, nil)
			}},
		"summary-null": {"GET", "/Teams.Mms/GetMeetingMigrationStatusSummary", "Get-CsMeetingMigrationStatus", "", `null`, 200,
			func(ctx context.Context, s *migration.Service) (any, *resty.Response, error) {
				return s.GetMeetingMigrationStatusSummary(ctx, nil)
			}},
		"summary-null-item": {"GET", "/Teams.Mms/GetMeetingMigrationStatusSummary", "Get-CsMeetingMigrationStatus", "", `{"SummaryOutput":[null]}`, 200,
			func(ctx context.Context, s *migration.Service) (any, *resty.Response, error) {
				return s.GetMeetingMigrationStatusSummary(ctx, nil)
			}},
		"history-bad-type": {"GET", "/Teams.Mms/GetMeetingMigrationTransactionHistory?userIdentity=x", "Get-CsMeetingMigrationTransactionHistory", "", `[{"TotalMeetings":"x"}]`, 200,
			func(ctx context.Context, s *migration.Service) (any, *resty.Response, error) {
				return s.ListMeetingMigrationTransactionHistory(ctx, "x", nil)
			}},
		"task-scalar": {"GET", "/Teams.Tenant/TenantMigrationDetail", "Get-CsMoveTenantServiceInstanceTaskStatus", "", `[1]`, 200,
			func(ctx context.Context, s *migration.Service) (any, *resty.Response, error) {
				return s.ListTenantMoveTaskStatuses(ctx)
			}},
		"cross-region-array": {"POST", "/Teams.Tenant/CrossRegionMoveTask", "Move-CsTenantCrossRegion", `{}`, `[]`, 200,
			func(ctx context.Context, s *migration.Service) (any, *resty.Response, error) {
				return s.MoveTenantCrossRegion(ctx)
			}},
		"start-array": {"POST", "/Teams.Mms/StartMeetingMigration", "Start-CsExMeetingMigration", `{"Identity":"x"}`, `[]`, 201,
			func(ctx context.Context, s *migration.Service) (any, *resty.Response, error) {
				return s.StartMeetingMigration(ctx, migration.StartMeetingMigrationRequest{Identity: "x"})
			}},
		"avs-object": {"POST", "/Teams.AVS/MovePartition", "Move-CsAvsTenantPartition", `{"SourcePartitionKey":"s","TargetPartitionKey":"t"}`, `{}`, 207,
			func(ctx context.Context, s *migration.Service) (any, *resty.Response, error) {
				return s.MoveAVSTenantPartition(ctx, migration.PartitionMoveRequest{SourcePartitionKey: "s", TargetPartitionKey: "t"})
			}},
	} {
		var body json.RawMessage
		if c.body != "" {
			body = json.RawMessage(c.body)
		}
		run(t, contracttest.Case{Name: name, Steps: []contracttest.Exchange{discovery(t), wire(c.method, c.path, c.command, body, c.status, []byte(c.reply))},
			Want: contracttest.Expected{Status: c.status, Error: client.ErrDecode},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				_, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.Migration)
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
	// The AVS move declares only HTTP 207; a plain 200 is undeclared.
	run(t, contracttest.Case{Name: "avs-undeclared-status", Steps: []contracttest.Exchange{discovery(t),
		wire("POST", "/Teams.AVS/MovePartition", "Move-CsAvsTenantPartition", json.RawMessage(`{"SourcePartitionKey":"s","TargetPartitionKey":"t"}`), 200, []byte(`[]`))},
		Want: contracttest.Expected{Status: 200, Error: migration.ErrResponse},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, response, err := sdk.TeamsAdminAPI.Migration.MoveAVSTenantPartition(ctx, migration.PartitionMoveRequest{SourcePartitionKey: "s", TargetPartitionKey: "t"})
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitMigrationInvalidInputFailsBeforeIO(t *testing.T) {
	t.Parallel()
	for name, call := range map[string]func(context.Context, *migration.Service) error{
		"status-blank-identity": func(ctx context.Context, s *migration.Service) error {
			_, _, e := s.ListMeetingMigrationStatuses(ctx, &migration.StatusOptions{Identity: ptr(" ")})
			return e
		},
		"summary-control-state": func(ctx context.Context, s *migration.Service) error {
			_, _, e := s.GetMeetingMigrationStatusSummary(ctx, &migration.StatusOptions{State: ptr("a\nb")})
			return e
		},
		"history-blank": func(ctx context.Context, s *migration.Service) error {
			_, _, e := s.ListMeetingMigrationTransactionHistory(ctx, "", nil)
			return e
		},
		"history-blank-correlation": func(ctx context.Context, s *migration.Service) error {
			_, _, e := s.ListMeetingMigrationTransactionHistory(ctx, "x", &migration.TransactionHistoryOptions{CorrelationID: ptr("")})
			return e
		},
		"start-blank": func(ctx context.Context, s *migration.Service) error {
			_, _, e := s.StartMeetingMigration(ctx, migration.StartMeetingMigrationRequest{})
			return e
		},
		"start-blank-type": func(ctx context.Context, s *migration.Service) error {
			_, _, e := s.StartMeetingMigration(ctx, migration.StartMeetingMigrationRequest{Identity: "x", TargetMeetingType: ptr("")})
			return e
		},
		"service-instance-no-option": func(ctx context.Context, s *migration.Service) error {
			_, _, e := s.MoveTenantServiceInstance(ctx, migration.MoveTenantServiceInstanceRequest{})
			return e
		},
		"service-instance-control": func(ctx context.Context, s *migration.Service) error {
			_, _, e := s.MoveTenantServiceInstance(ctx, migration.MoveTenantServiceInstanceRequest{MoveOption: "Finalize", TargetServiceInstance: ptr("a\tb")})
			return e
		},
		"avs-no-source": func(ctx context.Context, s *migration.Service) error {
			_, _, e := s.MoveAVSTenantPartition(ctx, migration.PartitionMoveRequest{TargetPartitionKey: "t"})
			return e
		},
		"avs-blank-workload": func(ctx context.Context, s *migration.Service) error {
			_, _, e := s.MoveAVSTenantPartition(ctx, migration.PartitionMoveRequest{SourcePartitionKey: "s", TargetPartitionKey: "t", Workload: ptr("")})
			return e
		},
	} {
		run(t, contracttest.Case{Name: name, Scope: "local", Want: contracttest.Expected{Error: migration.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return contracttest.Observation{Err: call(ctx, sdk.TeamsAdminAPI.Migration)}
			}})
	}
}

func TestUnitMigrationNilContextAndServiceAreRejected(t *testing.T) {
	t.Parallel()
	var s *migration.Service
	//nolint:staticcheck // A nil context is the guarded input.
	if _, _, err := s.ListMeetingMigrationStatuses(nil, nil); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, _, err := s.MoveTenantCrossRegion(t.Context()); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, _, err := migration.NewService(nil).StartMeetingMigration(ctx, migration.StartMeetingMigrationRequest{Identity: probe}); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled context accepted")
	}
}
