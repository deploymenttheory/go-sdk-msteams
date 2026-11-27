package shiftsconnections_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	sc "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/shifts_connections"
	"github.com/go-resty/resty/v2"
)

const (
	configOrigin = "https://api.interfaces.records.teams.microsoft.com"
	targetForest = "https://admin.example.invalid/"

	probe     = "00000000-0000-4000-8000-000000000001"
	connector = "00000000-0000-4000-8000-000000000003"
	jsonType  = "application/json; charset=utf-8"
	v2        = "/Teams.ShiftsConnector/api/v2"
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

func discovery(t *testing.T) testutil.Exchange {
	t.Helper()
	body, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: http.MethodGet, URL: configOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience},
		Reply: testutil.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body}}
}

func wire(method, path, command, ifMatch string, body json.RawMessage, status int, reply []byte) testutil.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {targetForest}}
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	if ifMatch != "" {
		h.Set("If-Match", ifMatch)
	}
	contentType := jsonType
	if reply == nil {
		contentType = ""
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: configOrigin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body},
		Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {contentType}}, Body: reply}}
}

func run(t *testing.T, c testutil.Case) {
	t.Helper()

	result, err := testutil.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Fatalf("shifts connection contract %s differed: %+v", c.Name, result.Differences())
	}
}

func ptr[T any](v T) *T { return &v }

type invocation func(context.Context, *sc.Service) (any, *resty.Response, error)

func TestUnitShiftsConnectionsAuthorizationDenialUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	denied := readFile(t, "testdata/list_connectors_403.json")
	created := readFile(t, "testdata/create_connection_403.json")
	settings := sc.Settings{"apiUrl": "https://example.invalid"}

	cases := []struct {
		name, method, path, command, ifMatch, body string
		reply                                      []byte
		invoke                                     invocation
	}{
		{"connectors", "GET", v2 + "/connectors", "Get-CsTeamsShiftsConnectionConnector", "", "", denied, func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) { return s.ListConnectors(ctx) }},
		{"connections", "GET", v2 + "/connections", "Get-CsTeamsShiftsConnection", "", "", denied, func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) { return s.ListConnections(ctx) }},
		{"connection", "GET", v2 + "/connections/" + probe, "Get-CsTeamsShiftsConnection", "", "", denied, func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
			return s.GetConnectionByID(ctx, probe)
		}},
		{"instances", "GET", v2 + "/connectorinstances", "Get-CsTeamsShiftsConnectionInstance", "", "", denied, func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) { return s.ListInstances(ctx) }},
		{"instance", "GET", v2 + "/connectorinstances/" + probe, "Get-CsTeamsShiftsConnectionInstance", "", "", denied, func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
			return s.GetInstanceByID(ctx, probe)
		}},
		{"create-connection", "POST", v2 + "/connections", "New-CsTeamsShiftsConnection", "",
			`{"connectorId":"` + connector + `","connectorSpecificSettings":{"apiUrl":"https://example.invalid"},"name":"msteams-sdk-probe","state":"Inactive"}`, created,
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.CreateConnection(ctx, sc.CreateConnectionRequest{Name: "msteams-sdk-probe", ConnectorID: connector, State: "Inactive", ConnectorSpecificSettings: settings})
			}},
		{"replace-connection", "PUT", v2 + "/connections/" + probe, "Set-CsTeamsShiftsConnection", "probe-etag", `{"name":"msteams-sdk-probe"}`, denied,
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.ReplaceConnectionByID(ctx, probe, sc.ConnectionUpdate{IfMatch: ptr("probe-etag"), Name: ptr("msteams-sdk-probe")})
			}},
		{"update-connection", "PATCH", v2 + "/connections/" + probe, "Update-CsTeamsShiftsConnection", "probe-etag", `{"name":"msteams-sdk-probe"}`, denied,
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.UpdateConnectionByID(ctx, probe, sc.ConnectionUpdate{IfMatch: ptr("probe-etag"), Name: ptr("msteams-sdk-probe")})
			}},
		{"delete-connection", "DELETE", v2 + "/connections/" + probe, "Remove-CsTeamsShiftsConnection", "", "", denied, func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
			response, err := s.DeleteConnectionByID(ctx, probe)
			return nil, response, err
		}},
		{"validate", "POST", "/Teams.ShiftsConnector/api/v1/connect", "Test-CsTeamsShiftsConnectionValidate", "",
			`{"connectorId":"` + connector + `","connectorSpecificSettings":{"apiUrl":"https://example.invalid"},"name":"msteams-sdk-probe"}`, denied,
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				response, err := s.Validate(ctx, sc.ValidateRequest{Name: "msteams-sdk-probe", ConnectorID: connector, ConnectorSpecificSettings: settings})
				return nil, response, err
			}},
		{"create-instance", "POST", v2 + "/connectorinstances", "New-CsTeamsShiftsConnectionInstance", "",
			`{"connectionId":"` + probe + `","name":"msteams-sdk-probe","state":"Inactive"}`, denied, func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.CreateInstance(ctx, sc.InstanceRequest{ConnectionID: ptr(probe), Name: ptr("msteams-sdk-probe"), State: ptr("Inactive")})
			}},
		{"replace-instance", "PUT", v2 + "/connectorinstances/" + probe, "Set-CsTeamsShiftsConnectionInstance", "probe-etag", `{"name":"msteams-sdk-probe"}`, denied,
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.ReplaceInstanceByID(ctx, probe, sc.InstanceUpdate{IfMatch: "probe-etag", InstanceRequest: sc.InstanceRequest{Name: ptr("msteams-sdk-probe")}})
			}},
		{"update-instance", "PATCH", v2 + "/connectorinstances/" + probe, "Update-CsTeamsShiftsConnectionInstance", "probe-etag", `{"name":"msteams-sdk-probe"}`, denied,
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				return s.UpdateInstanceByID(ctx, probe, sc.InstanceUpdate{IfMatch: "probe-etag", InstanceRequest: sc.InstanceRequest{Name: ptr("msteams-sdk-probe")}})
			}},
		{"delete-instance", "DELETE", v2 + "/connectorinstances/" + probe, "Remove-CsTeamsShiftsConnectionInstance", "", "", denied,
			func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
				response, err := s.DeleteInstanceByID(ctx, probe)
				return nil, response, err
			}},
	}
	for _, c := range cases {
		var body json.RawMessage
		if c.body != "" {
			body = json.RawMessage(c.body)
		}
		run(t, testutil.Case{Name: "actual-denied-" + c.name,
			Steps: []testutil.Exchange{discovery(t), wire(c.method, c.path, c.command, c.ifMatch, body, 403, c.reply)},
			Want:  testutil.Expected{Status: 403, Body: c.reply, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 403, Codes: []string{"Forbidden"}}},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				value, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.ShiftsConnections)
				if !errors.Is(err, auth.ErrForbidden) || !isNil(value) {
					t.Fatal("denial was not a typed forbidden error without a value")
				}
				return testutil.Observation{Response: response, Err: err}
			}})
	}
}

func isNil(v any) bool {
	switch x := v.(type) {
	case nil:
		return true
	case []*sc.Connector:
		return x == nil
	case []*sc.Connection:
		return x == nil
	case []*sc.Instance:
		return x == nil
	case *sc.Connection:
		return x == nil
	case *sc.Instance:
		return x == nil
	}
	return false
}

const (
	connectionJSON = `{"connectorSpecificSettings":{"apiUrl":"https://wfm.example.invalid","ssoUrl":null},"id":"` + probe + `","tenantId":null,` +
		`"connectorId":"` + connector + `","name":"Store","etag":"\"1\"","createdDateTime":"2026-01-01T00:00:00.0000000Z","lastModifiedDateTime":null,"state":"Active","future":1}`
	instanceJSON = `{"syncScenarios":{"shift":"TwoWay","openShift":null},"id":"` + probe + `","tenantId":null,"connectionId":"` + probe + `",` +
		`"connectorAdminEmails":["admin@example.invalid"],"connectorId":"` + connector + `","designatedActorId":"actor","name":"Store sync",` +
		`"syncFrequencyInMin":10,"workforceIntegrationId":"wi","etag":"\"2\"","createdDateTime":null,"lastModifiedDateTime":null,"state":"Active"}`
	connectorJSON = `{"supportedSyncScenarios":{"shift":["Disabled","TwoWay"]},"id":"` + connector + `","name":"UKG Dimensions","version":"1.0"}`
)

func TestUnitShiftsConnectionsSuccessShapesPreserveFields(t *testing.T) {
	t.Parallel()
	connectors := []byte(`{"value":[` + connectorJSON + `]}`)
	connectionsList := []byte(`{"value":[` + connectionJSON + `]}`)
	instances := []byte(`{"value":[` + instanceJSON + `]}`)
	run(t, testutil.Case{Name: "success-reads",
		Steps: []testutil.Exchange{discovery(t),
			wire("GET", v2+"/connectors", "Get-CsTeamsShiftsConnectionConnector", "", nil, 200, connectors),
			wire("GET", v2+"/connections", "Get-CsTeamsShiftsConnection", "", nil, 200, connectionsList),
			wire("GET", v2+"/connections/"+probe, "Get-CsTeamsShiftsConnection", "", nil, 200, []byte(connectionJSON)),
			wire("GET", v2+"/connectorinstances", "Get-CsTeamsShiftsConnectionInstance", "", nil, 200, instances),
			wire("GET", v2+"/connectorinstances/"+probe, "Get-CsTeamsShiftsConnectionInstance", "", nil, 200, []byte(instanceJSON))},
		Want: testutil.Expected{Status: 200, Body: []byte(instanceJSON)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.ShiftsConnections
			found, _, err := s.ListConnectors(ctx)
			if err != nil || len(found) != 1 || *found[0].Name != "UKG Dimensions" || string(found[0].SupportedSyncScenarios["shift"]) != `["Disabled","TwoWay"]` {
				t.Fatal("connector list was not decoded")
			}
			listed, _, err := s.ListConnections(ctx)
			if err != nil || len(listed) != 1 || listed[0].TenantID != nil || listed[0].RawFields["future"] == nil || string(listed[0].ConnectorSpecificSettings["ssoUrl"]) != "null" {
				t.Fatal("connection list lost null or unknown fields")
			}
			connection, _, err := s.GetConnectionByID(ctx, probe)
			if err != nil || *connection.Etag != `"1"` || connection.LastModifiedDateTime != nil {
				t.Fatal("connection detail was not decoded")
			}
			all, _, err := s.ListInstances(ctx)
			if err != nil || len(all) != 1 {
				t.Fatal("instance list was not decoded")
			}
			instance, response, err := s.GetInstanceByID(ctx, probe)
			if err != nil || *instance.SyncFrequencyInMin != 10 || *instance.SyncScenarios.Shift != "TwoWay" || instance.SyncScenarios.OpenShift != nil ||
				instance.ConnectorAdminEmails[0] != "admin@example.invalid" || *instance.WorkforceIntegrationID != "wi" {
				t.Fatal("instance detail was not decoded")
			}
			return testutil.Observation{Response: response, Err: err}
		}})

	createConnection := json.RawMessage(`{"connectorId":"` + connector + `","connectorSpecificSettings":{"apiUrl":"https://wfm.example.invalid","clientSecret":"s","appKey":"k"},"name":"Store","state":"Active"}`)
	updateConnection := json.RawMessage(`{"_etag":"\"1\"","connectorSpecificSettings":{"apiUrl":"https://wfm2.example.invalid"},"state":"Inactive"}`)
	createInstance := json.RawMessage(`{"connectionId":"` + probe + `","connectorAdminEmails":[],"designatedActorId":"actor","name":"Store sync","state":"Active","syncFrequencyInMin":0,"syncScenarios":{"shift":"TwoWay"}}`)
	updateInstance := json.RawMessage(`{"_etag":"\"2\"","syncFrequencyInMin":15}`)
	run(t, testutil.Case{Name: "success-writes",

		Steps: []testutil.Exchange{discovery(t),
			wire("POST", v2+"/connections", "New-CsTeamsShiftsConnection", "", createConnection, 201, []byte(connectionJSON)),
			wire("PUT", v2+"/connections/"+probe, "Set-CsTeamsShiftsConnection", "", updateConnection, 200, []byte(connectionJSON)),
			wire("PATCH", v2+"/connections/"+probe, "Update-CsTeamsShiftsConnection", `"1"`, updateConnection, 200, []byte(connectionJSON)),
			wire("POST", v2+"/connectorinstances", "New-CsTeamsShiftsConnectionInstance", "", createInstance, 200, []byte(instanceJSON)),
			wire("PUT", v2+"/connectorinstances/"+probe, "Set-CsTeamsShiftsConnectionInstance", `"2"`, updateInstance, 200, []byte(instanceJSON)),
			wire("PATCH", v2+"/connectorinstances/"+probe, "Update-CsTeamsShiftsConnectionInstance", `"2"`, updateInstance, 200, []byte(instanceJSON)),
			wire("POST", "/Teams.ShiftsConnector/api/v1/connect", "Test-CsTeamsShiftsConnectionValidate", "", json.RawMessage(`{"connectorId":"c","connectorSpecificSettings":{},"name":"n"}`), 204, nil),
			wire("DELETE", v2+"/connectorinstances/"+probe, "Remove-CsTeamsShiftsConnectionInstance", "", nil, 204, nil),
			wire("DELETE", v2+"/connections/"+probe, "Remove-CsTeamsShiftsConnection", "", nil, 200, nil)},
		Want: testutil.Expected{Status: 200},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.ShiftsConnections
			created, response, err := s.CreateConnection(ctx, sc.CreateConnectionRequest{Name: "Store", ConnectorID: connector, State: "Active",
				ConnectorSpecificSettings: sc.Settings{"apiUrl": "https://wfm.example.invalid", "clientSecret": "s", "appKey": "k"}})
			if err != nil || response.StatusCode() != 201 || *created.ID != probe {
				t.Fatal("connection create receipt was not decoded")
			}
			update := sc.ConnectionUpdate{Etag: ptr(`"1"`), State: ptr("Inactive"), ConnectorSpecificSettings: sc.Settings{"apiUrl": "https://wfm2.example.invalid"}}
			if _, _, err := s.ReplaceConnectionByID(ctx, probe, update); err != nil {
				t.Fatal("connection replace without If-Match failed")
			}
			update.IfMatch = ptr(`"1"`)
			if _, _, err := s.UpdateConnectionByID(ctx, probe, update); err != nil {
				t.Fatal("connection update with If-Match failed")
			}
			if _, _, err := s.CreateInstance(ctx, sc.InstanceRequest{ConnectionID: ptr(probe), ConnectorAdminEmails: []string{}, DesignatedActorID: ptr("actor"),
				Name: ptr("Store sync"), State: ptr("Active"), SyncFrequencyInMin: ptr(int32(0)), SyncScenarios: &sc.SyncScenarios{Shift: ptr("TwoWay")}}); err != nil {
				t.Fatal("instance create did not keep empty and zero values")
			}
			instanceUpdate := sc.InstanceUpdate{IfMatch: `"2"`, Etag: ptr(`"2"`), InstanceRequest: sc.InstanceRequest{SyncFrequencyInMin: ptr(int32(15))}}
			if _, _, err := s.ReplaceInstanceByID(ctx, probe, instanceUpdate); err != nil {
				t.Fatal("instance replace failed")
			}
			if _, _, err := s.UpdateInstanceByID(ctx, probe, instanceUpdate); err != nil {
				t.Fatal("instance update failed")
			}
			if response, err := s.Validate(ctx, sc.ValidateRequest{Name: "n", ConnectorID: "c", ConnectorSpecificSettings: sc.Settings{}}); err != nil || response.StatusCode() != 204 {
				t.Fatal("validation was not accepted")
			}
			if response, err := s.DeleteInstanceByID(ctx, probe); err != nil || response.StatusCode() != 204 {
				t.Fatal("instance delete was not accepted")
			}
			response, err = s.DeleteConnectionByID(ctx, probe)
			return testutil.Observation{Response: response, Err: err}
		}})
}

func TestUnitShiftsConnectionsMalformedResponsesFailDecode(t *testing.T) {
	t.Parallel()
	for name, c := range map[string]struct {
		path, command, body string
		invoke              invocation
	}{
		"connectors-array":   {v2 + "/connectors", "Get-CsTeamsShiftsConnectionConnector", `[]`, func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) { return s.ListConnectors(ctx) }},
		"connections-null":   {v2 + "/connections", "Get-CsTeamsShiftsConnection", `{"value":null}`, func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) { return s.ListConnections(ctx) }},
		"instances-nil-item": {v2 + "/connectorinstances", "Get-CsTeamsShiftsConnectionInstance", `{"value":[null]}`, func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) { return s.ListInstances(ctx) }},
		"connection-null": {v2 + "/connections/x", "Get-CsTeamsShiftsConnection", `null`, func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
			return s.GetConnectionByID(ctx, "x")
		}},
		"instance-bad-type": {v2 + "/connectorinstances/x", "Get-CsTeamsShiftsConnectionInstance", `{"syncFrequencyInMin":"ten"}`, func(ctx context.Context, s *sc.Service) (any, *resty.Response, error) {
			return s.GetInstanceByID(ctx, "x")
		}},
	} {
		run(t, testutil.Case{Name: name, Steps: []testutil.Exchange{discovery(t), wire("GET", c.path, c.command, "", nil, 200, []byte(c.body))},
			Want: testutil.Expected{Status: 200, Error: client.ErrDecode},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				value, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.ShiftsConnections)
				if !isNil(value) {
					t.Fatal("malformed response returned a value")
				}
				return testutil.Observation{Response: response, Err: err}
			}})
	}
	run(t, testutil.Case{Name: "undeclared-status", Steps: []testutil.Exchange{discovery(t), wire("DELETE", v2+"/connections/x", "Remove-CsTeamsShiftsConnection", "", nil, 202, nil)},
		Want: testutil.Expected{Status: 202, Error: sc.ErrResponse},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			response, err := sdk.TeamsAdminAPI.ShiftsConnections.DeleteConnectionByID(ctx, "x")
			return testutil.Observation{Response: response, Err: err}
		}})
}

func TestUnitShiftsConnectionsInvalidInputFailsBeforeIO(t *testing.T) {
	t.Parallel()
	settings := sc.Settings{"apiUrl": "u"}
	for name, call := range map[string]func(context.Context, *sc.Service) error{
		"get-connection-blank": func(ctx context.Context, s *sc.Service) error { _, _, e := s.GetConnectionByID(ctx, " "); return e },
		"get-instance-slash":   func(ctx context.Context, s *sc.Service) error { _, _, e := s.GetInstanceByID(ctx, "a/b"); return e },
		"create-no-settings": func(ctx context.Context, s *sc.Service) error {
			_, _, e := s.CreateConnection(ctx, sc.CreateConnectionRequest{Name: "n", ConnectorID: "c", State: "s"})
			return e
		},
		"create-no-state": func(ctx context.Context, s *sc.Service) error {
			_, _, e := s.CreateConnection(ctx, sc.CreateConnectionRequest{Name: "n", ConnectorID: "c", ConnectorSpecificSettings: settings})
			return e
		},
		"replace-traversal": func(ctx context.Context, s *sc.Service) error {
			_, _, e := s.ReplaceConnectionByID(ctx, "..", sc.ConnectionUpdate{})
			return e
		},
		"update-control-etag": func(ctx context.Context, s *sc.Service) error {
			_, _, e := s.UpdateConnectionByID(ctx, "x", sc.ConnectionUpdate{IfMatch: ptr("a\nb")})
			return e
		},
		"delete-connection-dot": func(ctx context.Context, s *sc.Service) error { _, e := s.DeleteConnectionByID(ctx, "."); return e },
		"create-instance-email": func(ctx context.Context, s *sc.Service) error {
			_, _, e := s.CreateInstance(ctx, sc.InstanceRequest{ConnectorAdminEmails: []string{""}})
			return e
		},
		"replace-instance-no-if-match": func(ctx context.Context, s *sc.Service) error {
			_, _, e := s.ReplaceInstanceByID(ctx, "x", sc.InstanceUpdate{})
			return e
		},
		"update-instance-no-if-match": func(ctx context.Context, s *sc.Service) error {
			_, _, e := s.UpdateInstanceByID(ctx, "x", sc.InstanceUpdate{IfMatch: " "})
			return e
		},
		"delete-instance-blank": func(ctx context.Context, s *sc.Service) error { _, e := s.DeleteInstanceByID(ctx, ""); return e },
		"validate-no-name": func(ctx context.Context, s *sc.Service) error {
			_, e := s.Validate(ctx, sc.ValidateRequest{ConnectorID: "c", ConnectorSpecificSettings: settings})
			return e
		},
	} {
		run(t, testutil.Case{Name: name, Want: testutil.Expected{Error: sc.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				return testutil.Observation{Err: call(ctx, sdk.TeamsAdminAPI.ShiftsConnections)}
			}})
	}
}

func TestUnitShiftsConnectionsNilContextServiceAndSecretsAreGuarded(t *testing.T) {
	t.Parallel()
	var s *sc.Service
	//nolint:staticcheck // A nil context is the guarded input.
	if _, _, err := s.ListConnectors(nil); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, err := s.DeleteInstanceByID(t.Context(), "x"); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := sc.NewService(nil).Validate(ctx, sc.ValidateRequest{}); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled context accepted")
	}
	secret := sc.Settings{"clientSecret": "do-not-print"}
	for _, value := range []any{
		sc.CreateConnectionRequest{ConnectorSpecificSettings: secret},
		sc.ConnectionUpdate{ConnectorSpecificSettings: secret},
		sc.ValidateRequest{ConnectorSpecificSettings: secret},
	} {
		for _, format := range []string{"%v", "%+v", "%#v", "%s"} {
			if text := fmt.Sprintf(format, value); text == "" || strings.Contains(text, "do-not-print") {
				t.Fatal("request settings were printed")
			}
		}
	}
}
