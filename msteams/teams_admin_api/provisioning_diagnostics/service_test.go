package provisioningdiagnostics_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	pd "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/provisioning_diagnostics"
	"github.com/go-resty/resty/v2"
)

const (
	configOrigin = "https://api.interfaces.records.teams.microsoft.com"
	targetForest = "https://admin.example.invalid/"
	source       = "msteams/teams_admin_api/provisioning_diagnostics/service_test.go"
	probe        = "00000000-0000-4000-8000-000000000026"
	tenant       = "00000000-0000-4000-8000-0000000000aa"
	sip          = "sip:msteams-sdk-probe@example.invalid"
	instance     = "MicrosoftCommunicationsOnline/msteams-sdk-probe"
)

func discovery(t *testing.T) contracttest.Exchange {
	t.Helper()
	body, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: http.MethodGet, URL: configOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience},
		Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body}}
}

func ptr[T any](value T) *T { return &value }

// operation is one SDK method with the exact request it must send.
type operation struct {
	name, command, set string
	method, path       string
	kind               string // "config", "move" or "oce"
	body               string // expected JSON request body; empty means none
	fixture            string // live denial fixture
	invoke             func(context.Context, *pd.Service) (any, *resty.Response, error)
}

func (o operation) request() contracttest.WireRequest {
	h := http.Header{}
	switch o.kind {
	case "config":
		h = http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {o.command}, "X-MS-Target-Uri": {targetForest}}
	case "move":
		h = http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {"Move-CsInternalHelper"}, "X-MS-Target-Uri": {targetForest + "HostedMigration"}, "X-MS-Tenant-Id": {tenant}}
	case "oce":
		h = http.Header{"X-MS-Tenant-Id": {tenant}}
	}
	w := contracttest.WireRequest{Method: o.method, URL: configOrigin + o.path, Audience: auth.TeamsAdminAudience, Headers: h}
	if o.body != "" {
		w.JSON = json.RawMessage(o.body)
	}
	return w
}

func (o operation) steps(t *testing.T, status int, reply []byte) []contracttest.Exchange {
	t.Helper()
	exchange := contracttest.Exchange{Request: o.request(), Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json; charset=utf-8"}}, Body: reply}}
	if o.kind == "oce" {
		return []contracttest.Exchange{exchange}
	}
	return []contracttest.Exchange{discovery(t), exchange}
}

func run(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family = "provisioning_diagnostics"
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
		t.Fatalf("provisioning diagnostics contract %s differed: %+v", c.Name, result.Differences())
	}
}

func three[T any](value T, response *resty.Response, err error) (any, *resty.Response, error) {
	return value, response, err
}

func two(response *resty.Response, err error) (any, *resty.Response, error) {
	return nil, response, err
}

var (
	move    = pd.UserMove{TenantID: tenant, UserSipURI: sip}
	teams   = pd.TeamsData{MoveToTeams: true, CheckCpc: true}
	payload = json.RawMessage(`{"value": 1}`)
)

func operations() []operation {
	return []operation{
		{name: "directory_diagnostic_data", command: "Get-CsBusinessVoiceDirectoryDiagnosticData", set: "__AllParameterSets", method: "GET", kind: "config",
			path: "/Teams.Cpc/bvdDiagnosticData/Users/NOAM/" + probe + "?resultSize=5&rowKey=" + probe, fixture: "directory_diagnostic_data_403.json",
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return three(s.GetDirectoryDiagnosticData(ctx, pd.DirectoryDiagnosticQuery{Table: "Users", Region: "NOAM", PartitionKey: probe, RowKey: ptr(probe), ResultSize: ptr(int32(5))}))
			}},
		{name: "object_changelog", command: "Get-CsMasObjectChangelog", set: "Get", method: "GET", kind: "config",
			path: "/Teams.Tenant/GetMASObjectChangelog?Identity=" + probe + "&SchemaNames=UserAuthoredProps&Versions=3", fixture: "object_changelog_403.json",
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return three(s.ListObjectChangelog(ctx, pd.ObjectChangelogQuery{Identity: probe, SchemaName: "UserAuthoredProps", Version: "3"}))
			}},
		{name: "versioned_schema_data", command: "Get-CsMasVersionedSchemaData", set: "__AllParameterSets", method: "GET", kind: "config",
			path: "/Teams.Tenant/GetSchemaDataFromMAS?Identity=" + probe + "&SchemaName=User+Authored&Versions=3", fixture: "versioned_schema_data_403.json",
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return three(s.ListVersionedSchemaData(ctx, pd.VersionedSchemaQuery{SchemaName: "User Authored", Identity: probe, Version: "3"}))
			}},
		{name: "user_setting", command: "Get-CsUssUserSettings", set: "Get", method: "GET", kind: "config",
			path: "/Teams.UserSettings/api/v1/tenant/" + tenant + "/user/" + probe + "/setting?name=msteams+sdk%26probe", fixture: "user_setting_403.json",
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return three(s.GetUserSetting(ctx, tenant, probe, "msteams sdk&probe"))
			}},
		{name: "set_user_setting", command: "Set-CsUssUserSettings", set: "Set", method: "PATCH", kind: "config", body: `{"value":1}`,
			path: "/Teams.UserSettings/api/v1/tenant/" + tenant + "/user/" + probe + "/setting/msteams-sdk-probe", fixture: "set_user_setting_403.json",
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return three(s.SetUserSetting(ctx, tenant, probe, "msteams-sdk-probe", payload))
			}},
		{name: "custom_handler_callback", command: "Invoke-CsCustomHandlerCallBackNgtprov", set: "__AllParameterSets", method: "POST", kind: "config",
			path: "/Teams.Tenant/CustomHandlerCallback?eventname=msteams-sdk-probe&id=" + probe + "&operation=StatusQueryGetUri", fixture: "custom_handler_callback_401.json",
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return three(s.InvokeCustomHandlerCallback(ctx, pd.CustomHandlerCallbackRequest{ID: probe, Operation: pd.StatusQueryGetURI, EventName: "msteams-sdk-probe"}))
			}},
		{name: "custom_handler", command: "Invoke-CsCustomHandlerNgtprov", set: "CustomExpanded", method: "POST", kind: "config",
			path: "/Teams.Tenant/CustomHandler", body: `{"HandlerFullyQualifiedName":"msteams-sdk-invalid","Payload":"{}"}`, fixture: "custom_handler_401.json",
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return three(s.InvokeCustomHandler(ctx, pd.CustomHandlerPayload{HandlerFullyQualifiedName: ptr("msteams-sdk-invalid"), Payload: ptr("{}")}))
			}},
		{name: "direct_object_sync", command: "Invoke-CsDirectObjectSync", set: "PostExpanded", method: "POST", kind: "config",
			path: "/Teams.InternalSupport/ProvisioningOperations/DsRequest", fixture: "direct_object_sync_403.json",
			body: `{"ServiceInstance":"x","ObjectId":"` + probe + `","ScenariosToSuppress":"s","ReSyncOptions":2,"DeploymentName":"EastUS","ObjectClass":"User","IsValidationRequest":true,"ObjectIds":["` + probe + `"]}`,
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return two(s.SyncDirectObjects(ctx, pd.DirectObjectSyncRequest{ServiceInstance: ptr("x"), ObjectID: ptr(probe), ScenariosToSuppress: ptr("s"), ReSyncOptions: ptr(int64(2)),
					DeploymentName: ptr("EastUS"), ObjectClass: ptr("User"), IsValidationRequest: ptr(true), ObjectIDs: []string{probe}}))
			}},
		{name: "msods_sync", command: "Invoke-CsMsodsSync", set: "PostExpanded", method: "POST", kind: "config",
			path: "/Teams.Tenant/MsodsSync", fixture: "msods_sync_403.json",
			body: `{"ServiceInstance":"x","TenantId":"` + probe + `","ScenariosToSuppress":"s","SynchronizeTenantWithAllObjects":true,"ReSyncOptions":2,"ObjectClass":"User","IsValidationRequest":true,"ObjectIds":["` + probe + `"]}`,
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return two(s.SyncMsodsObjects(ctx, pd.MsodsSyncRequest{ServiceInstance: ptr("x"), TenantID: probe, ScenariosToSuppress: ptr("s"), SynchronizeTenantWithAllObjects: ptr(true),
					ReSyncOptions: ptr(int64(2)), ObjectClass: "User", IsValidationRequest: ptr(true), ObjectIDs: []string{probe}}))
			}},
		{name: "rehome_user", command: "Invoke-CsRehomeuser", set: "Post", method: "POST", kind: "config", body: `{"value":1}`,
			path: "/Teams.User/RehomeUser/" + probe + "/moveUser", fixture: "rehome_user_403.json",
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return two(s.RehomeUser(ctx, probe, payload))
			}},
		{name: "tenant_user_backfill", command: "Set-CsTenantUserBackfill", set: "Set", method: "POST", kind: "config", body: `{"value":1}`,
			path: "/Teams.InternalSupport/SetBackfillToPubsub?operation=msteams-sdk+invalid", fixture: "tenant_user_backfill_403.json",
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return three(s.SetTenantUserBackfill(ctx, "msteams-sdk invalid", payload))
			}},
		{name: "validate_user_move", set: "Validate", method: "POST", kind: "move", path: "/Teams.HostedMigrationService/ValidateUser", fixture: "validate_user_move_403.json",
			body: `{"localDeploymentInfo":{"PresenceFqdn":"p.example.invalid","RegistrarFqdn":"r.example.invalid","HostingProviderFqdn":"h.example.invalid","MajorVersion":6},
				"teamsData":{"MoveToTeams":true,"CheckEnterpriseVoice":false,"CheckCpc":true},"userSipUri":"` + sip + `","moveToCloud":true,"force":true,"cmdletVersion":"1.0"}`,
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return three(s.ValidateUserMove(ctx, pd.ValidateUserMoveRequest{UserMove: move, CmdletVersion: "1.0", TeamsData: teams, MoveToCloud: true, Force: true,
					LocalDeployment: pd.LocalDeployment{PresenceFqdn: "p.example.invalid", RegistrarFqdn: "r.example.invalid", HostingProviderFqdn: "h.example.invalid", MajorVersion: 6}}))
			}},
		{name: "begin_user_move", set: "BeginAndCompleteMove", method: "POST", kind: "move", path: "/Teams.HostedMigrationService/BeginMove", fixture: "begin_user_move_403.json",
			body: `{"userSipUri":"` + sip + `","majorVersion":"6"}`,
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return three(s.BeginUserMove(ctx, pd.UserMoveVersionRequest{UserMove: move, MajorVersion: "6"}))
			}},
		{name: "complete_user_move", set: "BeginAndCompleteMove", method: "POST", kind: "move", path: "/Teams.HostedMigrationService/CompleteMove", fixture: "complete_user_move_403.json",
			body: `{"userSipUri":"` + sip + `","majorVersion":"6"}`,
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return two(s.CompleteUserMove(ctx, pd.UserMoveVersionRequest{UserMove: move, MajorVersion: "6"}))
			}},
		{name: "user_move_policies", set: "Rehome", method: "POST", kind: "move", path: "/Teams.HostedMigrationService/GetPolicies", fixture: "user_move_policies_403.json",
			body: `{"userSipUri":"` + sip + `"}`,
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return three(s.GetUserMovePolicies(ctx, move))
			}},
		{name: "rollback_user_move", set: "Rehome", method: "POST", kind: "move", path: "/Teams.HostedMigrationService/Rollback", fixture: "rollback_user_move_403.json",
			body: `{"userSipUri":"` + sip + `"}`,
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return two(s.RollbackUserMove(ctx, move))
			}},
		{name: "rehome_user_move", set: "Rehome", method: "POST", kind: "move", path: "/Teams.HostedMigrationService/RehomeUser", fixture: "rehome_user_move_403.json",
			body: `{"teamsData":{"MoveToTeams":true,"CheckEnterpriseVoice":false,"CheckCpc":true},"userSipUri":"` + sip + `","moveToCloud":true}`,
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return two(s.RehomeUserMove(ctx, pd.RehomeUserMoveRequest{UserMove: move, TeamsData: teams, MoveToCloud: true}))
			}},
		{name: "moved_resource_data", set: "MoveResourcedata", method: "POST", kind: "move", path: "/Teams.HostedMigrationService/SetMovedResourceData", fixture: "moved_resource_data_403.json",
			body: `{"resourceData":{"Datastr":"msteams-sdk-probe"},"teamsData":{"MoveToTeams":true,"CheckEnterpriseVoice":false,"CheckCpc":true},"userSipUri":"` + sip + `","majorVersion":"6"}`,
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return two(s.SetMovedResourceData(ctx, pd.MovedResourceDataRequest{UserMove: move, ResourceData: "msteams-sdk-probe", MajorVersion: "6", TeamsData: teams}))
			}},
		{name: "user_provisioning_history", command: "Get-CsUserProvHistory", set: "__AllParameterSets", method: "GET", kind: "oce",
			path: "/Teams.InternalSupport/AdminStoreOperations/userHistory/" + probe + "/EMEA", fixture: "user_provisioning_history_403.json",
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return three(s.GetUserProvisioningHistory(ctx, tenant, strings.ToUpper(probe), pd.RegionEMEA))
			}},
		{name: "ngtprov_handler", command: "Invoke-CsGenericNgtProvCommand", set: "__AllParameterSets", method: "POST", kind: "oce", body: `{"value":1}`,
			path:    "/Teams.InternalSupport/ProvisioningOperations/runHandler/EastUS?handler=msteams-sdk-invalid%3Ftenant%3D" + tenant + "%3Ftarget%3D" + probe + "%3FserviceInstance%3DMicrosoftCommunicationsOnline%2Fmsteams-sdk-probe",
			fixture: "ngtprov_handler_403.json",
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return two(s.InvokeNgtProvHandler(ctx, pd.NgtProvHandlerRequest{TenantID: tenant, Region: pd.DeploymentEastUS, HandlerName: "msteams-sdk-invalid", Target: probe, ServiceInstance: instance, Payload: payload}))
			}},
		{name: "directory_object_sync", command: "Invoke-CsDirectoryObjectSync", set: "__AllParameterSets", method: "POST", kind: "oce",
			path: "/Teams.InternalSupport/ProvisioningOperations/DsRequest", fixture: "directory_object_sync_403.json",
			body: `{"ServiceInstance":"` + instance + `","DeploymentName":3,"ObjectId":"` + probe + `","ObjectIds":null,"ObjectClass":3,"SynchronizeTenantWithAllObjects":false,"IsValidationRequest":true,"ScenariosToSuppress":null}`,
			invoke: func(ctx context.Context, s *pd.Service) (any, *resty.Response, error) {
				return two(s.SyncDirectoryObject(ctx, pd.DirectoryObjectSyncRequest{TenantID: tenant, Deployment: pd.DeploymentEastUS, ServiceInstance: instance, ObjectClass: pd.ObjectClassUser,
					ObjectID: "{" + strings.ToUpper(probe) + "}", IsValidationRequest: true}))
			}},
	}
}

func (o operation) mapping() []contracttest.Mapping {
	command := o.command
	if o.kind == "move" {
		command = "Move-CsInternalHelper"
	}
	return []contracttest.Mapping{{Command: command, ParameterSet: o.set}}
}

// fixtureError rebuilds the normalized API error from a live denial body.
func fixtureError(t *testing.T, status int, body []byte) *client.APIError {
	t.Helper()
	var fields struct {
		Code string `json:"code"`
	}
	if json.Unmarshal(body, &fields) != nil {
		t.Fatal("invalid denial fixture")
	}
	want := &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}
	if fields.Code != "" {
		want.Codes = []string{fields.Code}
	}
	return want
}

func TestUnitProvisioningDiagnosticsDenialsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	for _, o := range operations() {
		body, err := os.ReadFile("testdata/" + o.fixture)
		if err != nil {
			t.Fatal(err)
		}
		status, want := 403, auth.ErrForbidden
		if strings.HasSuffix(o.fixture, "_401.json") {
			status, want = 401, auth.ErrUnauthenticated
		}
		run(t, contracttest.Case{Name: "actual-" + strings.ReplaceAll(o.name, "_", "-"), Source: "msteams/teams_admin_api/provisioning_diagnostics/testdata/" + o.fixture, Mappings: o.mapping(),
			Steps: o.steps(t, status, body),
			Want:  contracttest.Expected{Status: status, Body: body, APIError: fixtureError(t, status, body)},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := o.invoke(ctx, sdk.TeamsAdminAPI.ProvisioningDiagnostics)
				if !errors.Is(err, want) || !isZero(value) {
					t.Fatalf("%s denial was not classified or returned a value", o.name)
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
}

func isZero(value any) bool {
	switch v := value.(type) {
	case nil:
		return true
	case string:
		return v == ""
	case map[string]json.RawMessage:
		return v == nil
	case []map[string]json.RawMessage:
		return v == nil
	case []*pd.DirectoryDiagnosticEntity:
		return v == nil
	case []*pd.ChangelogItem:
		return v == nil
	case []*pd.KeyValue:
		return v == nil
	case *pd.CustomHandlerResult:
		return v == nil
	case *pd.OperationResult:
		return v == nil
	case *pd.DeploymentInfo:
		return v == nil
	case *pd.MigrationData:
		return v == nil
	case *pd.UserHistory:
		return v == nil
	}
	return false
}

// successes pairs each operation with a declared success reply and checks the
// decoded value keeps typed and unknown fields.
func TestUnitProvisioningDiagnosticsSuccessShapesPreserveFields(t *testing.T) {
	t.Parallel()
	type success struct {
		status int
		body   string
		check  func(any) bool
	}
	cases := map[string]success{
		"directory_diagnostic_data": {200, `[{"Properties":[{"Name":"Tenant","Type":"String","Value":"x","Extra":1}],"Future":true}]`, func(v any) bool {
			e := v.([]*pd.DirectoryDiagnosticEntity)
			return len(e) == 1 && *e[0].Properties[0].Name == "Tenant" && e[0].Properties[0].RawFields["Extra"] != nil && e[0].RawFields["Future"] != nil
		}},
		"object_changelog": {200, `[{"SchemaName":"UserAuthoredProps","Version":3,"VersionCreatedTime":"2026-11-23T01:02:03.1234567Z","ModifiedValues":[{"Path":"a","x":1}],"Orders":["o"],"Future":1}]`, func(v any) bool {
			c := v.([]*pd.ChangelogItem)
			return len(c) == 1 && *c[0].Version == 3 && len(c[0].ModifiedValues) == 1 && c[0].Orders[0] == "o" && c[0].RawFields["Future"] != nil
		}},
		"versioned_schema_data": {200, `[{"id":"a","Version":1,"Anything":{"x":1}}]`, func(v any) bool {
			s := v.([]map[string]json.RawMessage)
			return len(s) == 1 && string(s[0]["Anything"]) == `{"x":1}`
		}},
		"user_setting":     {200, `{"setting":"on"}`, func(v any) bool { return v.(string) == `{"setting":"on"}` }},
		"set_user_setting": {201, `{"setting":"on"}`, func(v any) bool { return v.(string) == `{"setting":"on"}` }},
		"custom_handler_callback": {200, `{"runtimeStatus":"Running"}`, func(v any) bool {
			return string(v.(map[string]json.RawMessage)["runtimeStatus"]) == `"Running"`
		}},
		"custom_handler": {200, `{"id":"instance","Future":1}`, func(v any) bool {
			r := v.(*pd.CustomHandlerResult)
			return *r.ID == "instance" && r.RawFields["Future"] != nil
		}},
		"direct_object_sync":   {200, ``, func(v any) bool { return v == nil }},
		"msods_sync":           {200, ``, func(v any) bool { return v == nil }},
		"rehome_user":          {204, ``, func(v any) bool { return v == nil }},
		"tenant_user_backfill": {200, `{"message":"queued"}`, func(v any) bool { return *v.(*pd.OperationResult).Message == "queued" }},
		"validate_user_move": {200, `{"ValidateUserResult":{"PresenceFqdn":"p","RegistrarFqdn":"r","HostingProviderFqdn":"h","MajorVersion":6,"Future":1}}`, func(v any) bool {
			d := v.(*pd.DeploymentInfo)
			return *d.MajorVersion == 6 && *d.HostingProviderFqdn == "h" && d.RawFields["Future"] != nil
		}},
		"begin_user_move":    {200, `{"BeginMoveResult":{"Datastr":"state"}}`, func(v any) bool { return *v.(*pd.MigrationData).Datastr == "state" }},
		"complete_user_move": {200, ``, func(v any) bool { return v == nil }},
		"user_move_policies": {200, `{"GetPoliciesResult":[{"Key":"VoicePolicy","Value":"Tag:x"}]}`, func(v any) bool {
			p := v.([]*pd.KeyValue)
			return len(p) == 1 && *p[0].Key == "VoicePolicy" && *p[0].Value == "Tag:x"
		}},
		"rollback_user_move":  {200, ``, func(v any) bool { return v == nil }},
		"rehome_user_move":    {200, ``, func(v any) bool { return v == nil }},
		"moved_resource_data": {200, ``, func(v any) bool { return v == nil }},
		"user_provisioning_history": {200, `{"UserId":"u","TenantId":"t","VersionHistory":[{"Version":2,"VersionCreatedTime":"2026-11-23T01:02:03Z","OrderHistory":[{"OrderId":"o","Faulted":null,"AttemptCount":1,"JarvisUri":"j"}]}]}`, func(v any) bool {
			h := v.(*pd.UserHistory)
			o := h.VersionHistory[0].OrderHistory[0]
			return *h.VersionHistory[0].Version == 2 && *o.OrderID == "o" && o.Faulted == nil && *o.AttemptCount == 1 && *o.JarvisURI == "j"
		}},
		"ngtprov_handler":       {202, ``, func(v any) bool { return v == nil }},
		"directory_object_sync": {200, `{"accepted":true}`, func(v any) bool { return v == nil }},
	}
	for _, o := range operations() {
		c, ok := cases[o.name]
		if !ok {
			t.Fatalf("%s has no success shape", o.name)
		}
		var body []byte
		if c.body != "" {
			body = []byte(c.body)
		}
		run(t, contracttest.Case{Name: "success-" + strings.ReplaceAll(o.name, "_", "-"), Mappings: o.mapping(), Steps: o.steps(t, c.status, body),
			Want: contracttest.Expected{Status: c.status},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := o.invoke(ctx, sdk.TeamsAdminAPI.ProvisioningDiagnostics)
				if err != nil || !c.check(value) {
					t.Fatalf("%s success value differs: %v", o.name, err)
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
}

func TestUnitProvisioningDiagnosticsFailuresStayTyped(t *testing.T) {
	t.Parallel()
	byName := map[string]operation{}
	for _, o := range operations() {
		byName[o.name] = o
	}
	for index, c := range []struct {
		name   string
		status int
		body   string
		want   error
	}{
		{"directory_diagnostic_data", 200, `null`, client.ErrDecode},
		{"directory_diagnostic_data", 200, `[null]`, client.ErrDecode},
		{"object_changelog", 200, `{}`, client.ErrDecode},
		{"versioned_schema_data", 200, `[1]`, client.ErrDecode},
		{"versioned_schema_data", 200, `[null]`, client.ErrDecode},
		{"custom_handler_callback", 200, `[]`, client.ErrDecode},
		{"custom_handler", 200, `null`, client.ErrDecode},
		{"tenant_user_backfill", 200, ``, client.ErrDecode},
		{"validate_user_move", 200, `{"Other":{}}`, client.ErrDecode},
		{"validate_user_move", 200, `{"ValidateUserResult":null}`, client.ErrDecode},
		{"begin_user_move", 200, `{"BeginMoveResult":[]}`, client.ErrDecode},
		{"user_move_policies", 200, `{"GetPoliciesResult":[null]}`, client.ErrDecode},
		{"user_provisioning_history", 200, `[]`, client.ErrDecode},
		{"user_setting", 202, `{}`, pd.ErrResponse},
		{"set_user_setting", 204, ``, pd.ErrResponse},
		{"rehome_user", 200, ``, pd.ErrResponse},
		{"msods_sync", 201, ``, pd.ErrResponse},
		{"complete_user_move", 204, ``, pd.ErrResponse},
		{"ngtprov_handler", 206, ``, pd.ErrResponse},
	} {
		o := byName[c.name]
		var body []byte
		if c.body != "" {
			body = []byte(c.body)
		}
		run(t, contracttest.Case{Name: fmt.Sprintf("failure-%d", index), Mappings: o.mapping(), Steps: o.steps(t, c.status, body),
			Want: contracttest.Expected{Status: c.status, Error: c.want},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := o.invoke(ctx, sdk.TeamsAdminAPI.ProvisioningDiagnostics)
				if !isZero(value) {
					t.Fatalf("%s failure returned a value", c.name)
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
}

// Each request below is rejected before discovery or credential acquisition.
func TestUnitProvisioningDiagnosticsRejectsInvalidInputLocally(t *testing.T) {
	t.Parallel()
	valid := pd.DirectoryObjectSyncRequest{TenantID: tenant, Deployment: pd.DeploymentEastUS, ServiceInstance: instance, ObjectClass: pd.ObjectClassUser}
	invalid := map[string]func(context.Context, *pd.Service) error{
		"bvd-table-slash": func(ctx context.Context, s *pd.Service) error {
			_, _, err := s.GetDirectoryDiagnosticData(ctx, pd.DirectoryDiagnosticQuery{Table: "a/b", Region: "NOAM", PartitionKey: "p"})
			return err
		},
		"bvd-empty-row": func(ctx context.Context, s *pd.Service) error {
			_, _, err := s.GetDirectoryDiagnosticData(ctx, pd.DirectoryDiagnosticQuery{Table: "a", Region: "NOAM", PartitionKey: "p", RowKey: ptr(" ")})
			return err
		},
		"changelog-control": func(ctx context.Context, s *pd.Service) error {
			_, _, err := s.ListObjectChangelog(ctx, pd.ObjectChangelogQuery{Identity: "a\nb"})
			return err
		},
		"schema-name-required": func(ctx context.Context, s *pd.Service) error {
			_, _, err := s.ListVersionedSchemaData(ctx, pd.VersionedSchemaQuery{Identity: probe})
			return err
		},
		"setting-dot-user": func(ctx context.Context, s *pd.Service) error {
			_, _, err := s.GetUserSetting(ctx, tenant, "..", "n")
			return err
		},
		"setting-no-name": func(ctx context.Context, s *pd.Service) error {
			_, _, err := s.GetUserSetting(ctx, tenant, probe, "")
			return err
		},
		"set-setting-bad-json": func(ctx context.Context, s *pd.Service) error {
			_, _, err := s.SetUserSetting(ctx, tenant, probe, "n", json.RawMessage(`{`))
			return err
		},
		"set-setting-slash-name": func(ctx context.Context, s *pd.Service) error {
			_, _, err := s.SetUserSetting(ctx, tenant, probe, "a/b", payload)
			return err
		},
		"callback-no-id": func(ctx context.Context, s *pd.Service) error {
			_, _, err := s.InvokeCustomHandlerCallback(ctx, pd.CustomHandlerCallbackRequest{Operation: pd.RestartPostURI})
			return err
		},
		"msods-no-class": func(ctx context.Context, s *pd.Service) error {
			_, err := s.SyncMsodsObjects(ctx, pd.MsodsSyncRequest{TenantID: probe})
			return err
		},
		"rehome-empty-body": func(ctx context.Context, s *pd.Service) error { _, err := s.RehomeUser(ctx, probe, nil); return err },
		"backfill-no-operation": func(ctx context.Context, s *pd.Service) error {
			_, _, err := s.SetTenantUserBackfill(ctx, "", payload)
			return err
		},
		"move-bad-tenant": func(ctx context.Context, s *pd.Service) error {
			_, err := s.RollbackUserMove(ctx, pd.UserMove{TenantID: "contoso", UserSipURI: sip})
			return err
		},
		"move-no-sip": func(ctx context.Context, s *pd.Service) error {
			_, _, err := s.GetUserMovePolicies(ctx, pd.UserMove{TenantID: tenant})
			return err
		},
		"validate-no-version": func(ctx context.Context, s *pd.Service) error {
			_, _, err := s.ValidateUserMove(ctx, pd.ValidateUserMoveRequest{UserMove: move, LocalDeployment: pd.LocalDeployment{PresenceFqdn: "p", RegistrarFqdn: "r"}})
			return err
		},
		"begin-no-major": func(ctx context.Context, s *pd.Service) error {
			_, _, err := s.BeginUserMove(ctx, pd.UserMoveVersionRequest{UserMove: move})
			return err
		},
		"resource-data-empty": func(ctx context.Context, s *pd.Service) error {
			_, err := s.SetMovedResourceData(ctx, pd.MovedResourceDataRequest{UserMove: move, MajorVersion: "6"})
			return err
		},
		"history-region": func(ctx context.Context, s *pd.Service) error {
			_, _, err := s.GetUserProvisioningHistory(ctx, tenant, probe, "LATAM")
			return err
		},
		"history-user-guid": func(ctx context.Context, s *pd.Service) error {
			_, _, err := s.GetUserProvisioningHistory(ctx, tenant, "user@example.invalid", pd.RegionNOAM)
			return err
		},
		"ngtprov-instance": func(ctx context.Context, s *pd.Service) error {
			_, err := s.InvokeNgtProvHandler(ctx, pd.NgtProvHandlerRequest{TenantID: tenant, Region: pd.DeploymentEastUS, HandlerName: "h", Target: "t", ServiceInstance: "other/x", Payload: payload})
			return err
		},
		"ngtprov-region": func(ctx context.Context, s *pd.Service) error {
			_, err := s.InvokeNgtProvHandler(ctx, pd.NgtProvHandlerRequest{TenantID: tenant, Region: "Mars", HandlerName: "h", Target: "t", Payload: payload})
			return err
		},
		"dosync-tenant-sync-class": func(ctx context.Context, s *pd.Service) error {
			r := valid
			r.SynchronizeTenantWithAllObjects = true
			_, err := s.SyncDirectoryObject(ctx, r)
			return err
		},
		"dosync-instance": func(ctx context.Context, s *pd.Service) error {
			r := valid
			r.ServiceInstance = "msteams"
			_, err := s.SyncDirectoryObject(ctx, r)
			return err
		},
		"dosync-object-ids": func(ctx context.Context, s *pd.Service) error {
			r := valid
			r.ObjectIDs = []string{probe, "x"}
			_, err := s.SyncDirectoryObject(ctx, r)
			return err
		},
		"dosync-tenant": func(ctx context.Context, s *pd.Service) error {
			r := valid
			r.TenantID = ""
			_, err := s.SyncDirectoryObject(ctx, r)
			return err
		},
	}
	for name, call := range invalid {
		run(t, contracttest.Case{Name: "invalid-" + name, Scope: "local", Want: contracttest.Expected{Error: pd.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return contracttest.Observation{Err: call(ctx, sdk.TeamsAdminAPI.ProvisioningDiagnostics)}
			}})
	}
}

// The OCE directory sync sends ObjectIds instead of ObjectId when given, with
// the zero GUID the module's default leaves in ObjectId.
func TestUnitProvisioningDiagnosticsOCEObjectIDsReplaceObjectID(t *testing.T) {
	t.Parallel()
	o := operation{method: "POST", kind: "oce", path: "/Teams.InternalSupport/ProvisioningOperations/DsRequest",
		body: `{"ServiceInstance":"` + instance + `","DeploymentName":19,"ObjectId":"00000000-0000-0000-0000-000000000000","ObjectIds":["` + probe + `"],"ObjectClass":0,"SynchronizeTenantWithAllObjects":true,"IsValidationRequest":false,"ScenariosToSuppress":"s"}`}
	run(t, contracttest.Case{Name: "oce-object-ids", Mappings: []contracttest.Mapping{{Command: "Invoke-CsDirectoryObjectSync", ParameterSet: "__AllParameterSets"}}, Steps: o.steps(t, 200, nil),
		Want: contracttest.Expected{Status: 200},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.ProvisioningDiagnostics.SyncDirectoryObject(ctx, pd.DirectoryObjectSyncRequest{TenantID: tenant, Deployment: pd.DeploymentChinaNorth3,
				ServiceInstance: instance, ObjectClass: pd.ObjectClassCompany, ObjectID: "ignored", ObjectIDs: []string{probe}, SynchronizeTenantWithAllObjects: true, ScenariosToSuppress: ptr("s")})
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitProvisioningDiagnosticsNilContextAndServiceAreRejected(t *testing.T) {
	t.Parallel()
	var s *pd.Service
	//nolint:staticcheck // A nil context is the guarded input.
	if _, _, err := s.GetUserSetting(nil, tenant, probe, "n"); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, err := s.RollbackUserMove(t.Context(), move); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
	if _, err := s.SyncDirectoryObject(t.Context(), pd.DirectoryObjectSyncRequest{}); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted for OCE")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, _, err := pd.NewService(nil).ListObjectChangelog(ctx, pd.ObjectChangelogQuery{}); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled context accepted")
	}
}
