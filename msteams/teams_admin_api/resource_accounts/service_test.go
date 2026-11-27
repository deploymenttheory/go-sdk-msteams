package resourceaccounts_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	ra "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/resource_accounts"
	"github.com/go-resty/resty/v2"
)

const (
	configOrigin = "https://api.interfaces.records.teams.microsoft.com"
	targetForest = "https://admin.example.invalid/"

	probe         = "00000000-0000-4000-8000-000000000001"
	configuration = "00000000-0000-4000-8000-000000000002"
	jsonType      = "application/json; charset=utf-8"
	legacy        = "/Teams.PlatformService/v2/Legacy/ApplicationInstances"
	instances     = "/Teams.PlatformService/v2/ApplicationInstances"
	associations  = "/Teams.VoiceApps/applicationinstanceassociations"
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

func wire(method, path, command string, body json.RawMessage, status int, reply []byte) testutil.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {targetForest}}
	if body != nil {
		h.Set("Content-Type", "application/json")
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
		t.Fatalf("resource account contract %s differed: %+v", c.Name, result.Differences())
	}
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

func TestUnitResourceAccountReadsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	list, search, status := readFile(t, "testdata/list.json"), readFile(t, "testdata/search.json"), readFile(t, "testdata/association_status.json")
	missing, noAssociation := readFile(t, "testdata/get_missing_404.json"), readFile(t, "testdata/association_missing_404.json")
	run(t, testutil.Case{Name: "actual-reads",

		Steps: []testutil.Exchange{discovery(t),
			wire("GET", legacy+"?resultSize=100", "Get-CsOnlineApplicationInstance", nil, 200, list),
			wire("GET", "/Teams.VoiceApps/applicationinstances?MaxResults=100&SearchQuery=msteams-sdk", "Find-CsOnlineApplicationInstance", nil, 200, search),
			wire("GET", associations+"/"+probe+"/status", "Get-CsOnlineApplicationInstanceAssociationStatus", nil, 200, status)},
		Want: testutil.Expected{Status: 200, Body: status},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.ResourceAccounts
			values, _, err := s.List(ctx, &ra.ListOptions{ResultSize: ptr(int32(100))})
			if err != nil || values == nil || len(values) != 0 {
				t.Fatal("empty list was not a non-nil empty slice")
			}
			found, _, err := s.Search(ctx, ra.SearchRequest{SearchQuery: "msteams-sdk", MaxResults: ptr(int32(100))})
			if err != nil || found == nil || len(found) != 0 {
				t.Fatal("empty search was not a non-nil empty slice")
			}
			record, response, err := s.GetAssociationStatusByID(ctx, probe)
			if err != nil || !record.Failed() || *record.ErrorCode != "ApplicationInstanceAssociationStatusProvider_AppEndpointNotFound" ||
				*record.Status != 0 || string(record.AuxiliaryData) != "null" || len(record.RawFields) != 6 {
				t.Fatal("actual status record was changed")
			}
			return testutil.Observation{Response: response, Err: err}
		}})
	run(t, testutil.Case{Name: "actual-get-missing",
		Steps: []testutil.Exchange{discovery(t), wire("GET", legacy+"?identity="+probe, "Get-CsOnlineApplicationInstance", nil, 404, missing)},
		Want:  testutil.Expected{Status: 404, Body: missing, APIError: apiError(404, missing)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.ResourceAccounts.GetByID(ctx, probe)
			return testutil.Observation{Value: value, Response: response, Err: err}
		}})
	run(t, testutil.Case{Name: "actual-association-missing",
		Steps: []testutil.Exchange{discovery(t), wire("GET", associations+"/"+probe, "Get-CsOnlineApplicationInstanceAssociation", nil, 404, noAssociation)},
		Want:  testutil.Expected{Status: 404, Body: noAssociation, APIError: apiError(404, noAssociation)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.ResourceAccounts.GetAssociationByID(ctx, probe)
			return testutil.Observation{Value: value, Response: response, Err: err}
		}})
}

type invocation func(context.Context, *ra.Service) (any, *resty.Response, error)

func TestUnitResourceAccountRejectedWritesUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	app, name := ra.CallQueueApplicationID, "msteams-sdk-probe"
	for _, c := range []struct {
		name, method, path, command, fixture string
		body                                 string
		invoke                               invocation
	}{
		{"create-unverified-domain", "POST", instances, "New-CsOnlineApplicationInstance", "create_unverified_domain_400.json",
			`{"applicationId":"` + app + `","displayName":"` + name + `","userPrincipalName":"msteams-sdk-probe@example.invalid"}`,
			func(ctx context.Context, s *ra.Service) (any, *resty.Response, error) {
				return s.Create(ctx, ra.CreateRequest{UserPrincipalName: "msteams-sdk-probe@example.invalid", ApplicationID: &app, DisplayName: &name})
			}},
		{"update-without-application", "PATCH", instances + "/" + probe, "Set-CsOnlineApplicationInstance", "update_without_application_400.json",
			`{"displayName":"` + name + `"}`,
			func(ctx context.Context, s *ra.Service) (any, *resty.Response, error) {
				return s.UpdateByID(ctx, probe, ra.UpdateRequest{DisplayName: &name})
			}},
		{"update-app-only", "PATCH", instances + "/" + probe, "Set-CsOnlineApplicationInstance", "update_app_only_400.json",
			`{"applicationId":"` + app + `"}`,
			func(ctx context.Context, s *ra.Service) (any, *resty.Response, error) {
				return s.UpdateByID(ctx, probe, ra.UpdateRequest{ApplicationID: &app})
			}},
		{"telephone-number-deprecated", "PATCH", instances + "/" + probe + "/telephoneNumber", "Set-CsOnlineVoiceApplicationInstance", "telephone_number_deprecated_400.json",
			`{"telephoneNumber":"+15555550100"}`,
			func(ctx context.Context, s *ra.Service) (any, *resty.Response, error) {
				return s.UpdateTelephoneNumberByID(ctx, probe, ptr("+15555550100"))
			}},
		{"sync-without-ids", "POST", instances + "/" + probe + "/sync", "Sync-CsOnlineApplicationInstance", "sync_without_ids_400.json", "",
			func(ctx context.Context, s *ra.Service) (any, *resty.Response, error) {
				response, err := s.SyncByID(ctx, probe, ra.SyncRequest{})
				return nil, response, err
			}},
	} {
		reply := readFile(t, "testdata/"+c.fixture)
		var body json.RawMessage
		if c.body != "" {
			body = json.RawMessage(c.body)
		}
		run(t, testutil.Case{Name: "actual-" + c.name,
			Steps: []testutil.Exchange{discovery(t), wire(c.method, c.path, c.command, body, 400, reply)},
			Want:  testutil.Expected{Status: 400, Body: reply, APIError: apiError(400, reply)},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				value, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.ResourceAccounts)
				if v, ok := value.(*ra.ApplicationInstance); ok && v != nil {
					t.Fatal("rejected write returned a value")
				}
				return testutil.Observation{Response: response, Err: err}
			}})
	}
}

func TestUnitResourceAccountAssociationWritesUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	associated, unassociated := readFile(t, "testdata/associate_missing.json"), readFile(t, "testdata/unassociate_missing.json")
	body := json.RawMessage(`{"CallPriority":3,"ConfigurationId":"` + configuration + `","ConfigurationType":"CallQueue","EndpointsIds":["` + probe + `"]}`)
	run(t, testutil.Case{Name: "actual-association-writes",

		Steps: []testutil.Exchange{discovery(t),
			wire("POST", instances+"/"+probe+"/sync?applicationId="+ra.CallQueueApplicationID, "Sync-CsOnlineApplicationInstance", nil, 204, nil),
			wire("POST", associations, "New-CsOnlineApplicationInstanceAssociation", body, 200, associated),
			wire("DELETE", associations+"/"+probe, "Remove-CsOnlineApplicationInstanceAssociation", nil, 200, unassociated)},
		Want: testutil.Expected{Status: 200, Body: unassociated},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.ResourceAccounts
			if response, err := s.SyncByID(ctx, probe, ra.SyncRequest{ApplicationID: ptr(ra.CallQueueApplicationID)}); err != nil || response.StatusCode() != 204 {
				t.Fatal("sync was not accepted")
			}
			results, _, err := s.CreateAssociations(ctx, ra.AssociationRequest{EndpointIDs: []string{probe}, ConfigurationID: configuration,
				ConfigurationType: ra.ConfigurationTypeCallQueue, CallPriority: ptr(int16(3))})
			if err != nil || len(results) != 1 || results[0].Association != nil || !results[0].Status.Failed() ||
				*results[0].Status.ErrorCode != "CreateApplicationInstanceAssociation_AppEndpointNotFound" {
				t.Fatal("per-account association failure was not preserved")
			}
			results, response, err := s.DeleteAssociationByID(ctx, probe)
			if err != nil || len(results) != 1 || *results[0].Status.ErrorCode != "RemoveApplicationInstanceAssociation_AssociationNotFound" {
				t.Fatal("per-account removal failure was not preserved")
			}
			return testutil.Observation{Response: response, Err: err}
		}})
}

func TestUnitResourceAccountSuccessShapesPreserveFields(t *testing.T) {
	t.Parallel()
	account := `{"objectId":"` + probe + `","tenantId":null,"userPrincipalName":"ra@example.invalid","applicationId":"` + ra.CallQueueApplicationID +
		`","acsResourceId":null,"displayName":"Front desk","phoneNumber":null,"future":{"x":1}}`
	list := []byte(`[` + account + `]`)
	single := []byte(account)
	created := []byte(account)
	found := []byte(`{"ApplicationInstances":[{"Id":"` + probe + `","DisplayName":"Front desk","SipUri":"sip:ra@example.invalid","ApplicationId":null,"IsOnPrem":false,"Extra":true}],"Diagnostics":[]}`)
	association := []byte(`{"ApplicationInstanceAssociation":{"Id":"` + probe + `","ApplicationConfigurationType":"CallQueue","ConfigurationId":"` + configuration + `","CallPriority":2}}`)
	createBody := json.RawMessage(`{"userPrincipalName":"ra@example.invalid"}`)
	run(t, testutil.Case{Name: "success-shapes",

		Steps: []testutil.Exchange{discovery(t),
			wire("GET", legacy+"?skip=0", "Get-CsOnlineApplicationInstance", nil, 200, list),
			wire("GET", legacy+"?identity=ra%40example.invalid", "Get-CsOnlineApplicationInstance", nil, 200, list),
			wire("GET", legacy+"?identity="+probe, "Get-CsOnlineApplicationInstance", nil, 200, single),
			wire("POST", instances, "New-CsOnlineApplicationInstance", createBody, 201, created),
			wire("PATCH", instances+"/ra@example.invalid", "Set-CsOnlineApplicationInstance", json.RawMessage(`{"acsResourceId":"`+configuration+`","applicationId":"`+probe+`"}`), 204, nil),
			wire("GET", "/Teams.VoiceApps/applicationinstances?AssociatedOnly=False&ExactMatchOnly=True&SearchQuery=front+desk", "Find-CsOnlineApplicationInstance", nil, 200, found),
			wire("GET", associations+"/ra@example%5BDOT%5Dinvalid", "Get-CsOnlineApplicationInstanceAssociation", nil, 200, association)},
		Want: testutil.Expected{Status: 200, Body: association},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.ResourceAccounts
			values, _, err := s.List(ctx, &ra.ListOptions{Skip: ptr(int32(0))})
			if err != nil || len(values) != 1 || values[0].TenantID != nil || *values[0].DisplayName != "Front desk" || values[0].RawFields["future"] == nil || len(values[0].RawFields) != 8 {
				t.Fatal("list lost null or unknown fields")
			}
			for _, identity := range []string{"ra@example.invalid", probe} {
				value, _, err := s.GetByID(ctx, identity)
				if err != nil || *value.ObjectID != probe {
					t.Fatal("array and object detail shapes were not both accepted")
				}
			}
			value, response, err := s.Create(ctx, ra.CreateRequest{UserPrincipalName: "ra@example.invalid"})
			if err != nil || response.StatusCode() != 201 || *value.ApplicationID != ra.CallQueueApplicationID {
				t.Fatal("create receipt was not decoded")
			}
			// Braced uppercase GUIDs are normalized as a bound System.Guid is.
			upper := "{00000000-0000-4000-8000-000000000001}"
			value, _, err = s.UpdateByID(ctx, "ra@example.invalid", ra.UpdateRequest{ApplicationID: &upper, AcsResourceID: ptr("00000000000040008000000000000002")})
			if err != nil || value != nil {
				t.Fatal("empty update reply was not a nil value")
			}
			results, _, err := s.Search(ctx, ra.SearchRequest{SearchQuery: "front desk", ExactMatchOnly: ptr(true), AssociatedOnly: ptr(false)})
			if err != nil || len(results) != 1 || results[0].ApplicationID != nil || results[0].IsOnPrem == nil || *results[0].IsOnPrem || results[0].RawFields["Extra"] == nil {
				t.Fatal("search result lost fields")
			}
			got, response, err := s.GetAssociationByID(ctx, "ra@example.invalid")
			if err != nil || *got.ApplicationConfigurationType != "CallQueue" || *got.CallPriority != 2 {
				t.Fatal("association was not decoded")
			}
			return testutil.Observation{Response: response, Err: err}
		}})
}

func TestUnitResourceAccountMalformedResponsesFailDecode(t *testing.T) {
	t.Parallel()
	for name, c := range map[string]struct {
		path, command, body string
		invoke              invocation
	}{
		"list-object":    {legacy, "Get-CsOnlineApplicationInstance", `{}`, func(ctx context.Context, s *ra.Service) (any, *resty.Response, error) { return s.List(ctx, nil) }},
		"list-null-item": {legacy, "Get-CsOnlineApplicationInstance", `[null]`, func(ctx context.Context, s *ra.Service) (any, *resty.Response, error) { return s.List(ctx, nil) }},
		"get-two":        {legacy + "?identity=x", "Get-CsOnlineApplicationInstance", `[{},{}]`, func(ctx context.Context, s *ra.Service) (any, *resty.Response, error) { return s.GetByID(ctx, "x") }},
		"get-bad-type":   {legacy + "?identity=x", "Get-CsOnlineApplicationInstance", `{"objectId":7}`, func(ctx context.Context, s *ra.Service) (any, *resty.Response, error) { return s.GetByID(ctx, "x") }},
		"search-missing": {"/Teams.VoiceApps/applicationinstances?SearchQuery=x", "Find-CsOnlineApplicationInstance", `{}`, func(ctx context.Context, s *ra.Service) (any, *resty.Response, error) {
			return s.Search(ctx, ra.SearchRequest{SearchQuery: "x"})
		}},
		"association-null": {associations + "/x", "Get-CsOnlineApplicationInstanceAssociation", `{"ApplicationInstanceAssociation":null}`, func(ctx context.Context, s *ra.Service) (any, *resty.Response, error) {
			return s.GetAssociationByID(ctx, "x")
		}},
		"status-missing": {associations + "/x/status", "Get-CsOnlineApplicationInstanceAssociationStatus", `{"Diagnostics":[]}`, func(ctx context.Context, s *ra.Service) (any, *resty.Response, error) {
			return s.GetAssociationStatusByID(ctx, "x")
		}},
	} {
		run(t, testutil.Case{Name: name, Steps: []testutil.Exchange{discovery(t), wire("GET", c.path, c.command, nil, 200, []byte(c.body))},
			Want: testutil.Expected{Status: 200, Error: client.ErrDecode},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				_, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.ResourceAccounts)
				return testutil.Observation{Response: response, Err: err}
			}})
	}
	results := []byte(`{"AssociationOperationResults":null}`)
	run(t, testutil.Case{Name: "results-null", Steps: []testutil.Exchange{discovery(t), wire("DELETE", associations+"/x", "Remove-CsOnlineApplicationInstanceAssociation", nil, 200, results)},
		Want: testutil.Expected{Status: 200, Error: client.ErrDecode},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			_, response, err := sdk.TeamsAdminAPI.ResourceAccounts.DeleteAssociationByID(ctx, "x")
			return testutil.Observation{Response: response, Err: err}
		}})
	run(t, testutil.Case{Name: "undeclared-status", Steps: []testutil.Exchange{discovery(t), wire("POST", instances+"/"+probe+"/sync?acsResourceId="+probe, "Sync-CsOnlineApplicationInstance", nil, 202, nil)},
		Want: testutil.Expected{Status: 202, Error: ra.ErrResponse},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			response, err := sdk.TeamsAdminAPI.ResourceAccounts.SyncByID(ctx, probe, ra.SyncRequest{AcsResourceID: ptr(probe)})
			return testutil.Observation{Response: response, Err: err}
		}})
}

func TestUnitResourceAccountInvalidInputFailsBeforeIO(t *testing.T) {
	t.Parallel()
	bad := "not-a-guid"
	for name, call := range map[string]func(context.Context, *ra.Service) error{
		"get-blank":    func(ctx context.Context, s *ra.Service) error { _, _, e := s.GetByID(ctx, " "); return e },
		"get-control":  func(ctx context.Context, s *ra.Service) error { _, _, e := s.GetByID(ctx, "a\nb"); return e },
		"search-blank": func(ctx context.Context, s *ra.Service) error { _, _, e := s.Search(ctx, ra.SearchRequest{}); return e },
		"search-negative-max": func(ctx context.Context, s *ra.Service) error {
			_, _, e := s.Search(ctx, ra.SearchRequest{SearchQuery: "x", MaxResults: ptr(int32(-1))})
			return e
		},
		"create-no-upn": func(ctx context.Context, s *ra.Service) error { _, _, e := s.Create(ctx, ra.CreateRequest{}); return e },
		"create-bad-app": func(ctx context.Context, s *ra.Service) error {
			_, _, e := s.Create(ctx, ra.CreateRequest{UserPrincipalName: "a@b", ApplicationID: &bad})
			return e
		},
		"update-traversal": func(ctx context.Context, s *ra.Service) error {
			_, _, e := s.UpdateByID(ctx, "..", ra.UpdateRequest{})
			return e
		},
		"update-bad-acs": func(ctx context.Context, s *ra.Service) error {
			_, _, e := s.UpdateByID(ctx, "x", ra.UpdateRequest{AcsResourceID: &bad})
			return e
		},
		"phone-slash": func(ctx context.Context, s *ra.Service) error {
			_, _, e := s.UpdateTelephoneNumberByID(ctx, "a/b", nil)
			return e
		},
		"sync-upn": func(ctx context.Context, s *ra.Service) error {
			_, e := s.SyncByID(ctx, "a@b", ra.SyncRequest{})
			return e
		},
		"sync-half-braced": func(ctx context.Context, s *ra.Service) error {
			_, e := s.SyncByID(ctx, "{"+probe, ra.SyncRequest{})
			return e
		},
		"association-blank": func(ctx context.Context, s *ra.Service) error { _, _, e := s.GetAssociationByID(ctx, ""); return e },
		"status-slash": func(ctx context.Context, s *ra.Service) error {
			_, _, e := s.GetAssociationStatusByID(ctx, "a/b")
			return e
		},
		"associate-none": func(ctx context.Context, s *ra.Service) error {
			_, _, e := s.CreateAssociations(ctx, ra.AssociationRequest{ConfigurationID: "c", ConfigurationType: "CallQueue"})
			return e
		},
		"associate-blank-id": func(ctx context.Context, s *ra.Service) error {
			_, _, e := s.CreateAssociations(ctx, ra.AssociationRequest{EndpointIDs: []string{""}, ConfigurationID: "c", ConfigurationType: "CallQueue"})
			return e
		},
		"associate-no-type": func(ctx context.Context, s *ra.Service) error {
			_, _, e := s.CreateAssociations(ctx, ra.AssociationRequest{EndpointIDs: []string{"e"}, ConfigurationID: "c"})
			return e
		},
		"associate-priority": func(ctx context.Context, s *ra.Service) error {
			_, _, e := s.CreateAssociations(ctx, ra.AssociationRequest{EndpointIDs: []string{"e"}, ConfigurationID: "c", ConfigurationType: "CallQueue", CallPriority: ptr(int16(6))})
			return e
		},
		"unassociate-dot": func(ctx context.Context, s *ra.Service) error { _, _, e := s.DeleteAssociationByID(ctx, "."); return e },
	} {
		run(t, testutil.Case{Name: name, Want: testutil.Expected{Error: ra.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				return testutil.Observation{Err: call(ctx, sdk.TeamsAdminAPI.ResourceAccounts)}
			}})
	}
}

func TestUnitResourceAccountNilContextAndServiceAreRejected(t *testing.T) {
	t.Parallel()
	var s *ra.Service
	//nolint:staticcheck // A nil context is the guarded input.
	if _, _, err := s.List(nil, nil); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, _, err := s.DeleteAssociationByID(t.Context(), "x"); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := ra.NewService(nil).SyncByID(ctx, probe, ra.SyncRequest{}); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled context accepted")
	}
}
