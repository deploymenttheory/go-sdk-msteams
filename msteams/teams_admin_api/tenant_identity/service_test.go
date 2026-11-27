package tenantidentity_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	ti "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_identity"
	"github.com/go-resty/resty/v2"
)

const (
	configOrigin = "https://api.interfaces.records.teams.microsoft.com"
	targetForest = "https://admin.example.invalid/"
	source       = "msteams/teams_admin_api/tenant_identity/service_test.go"
	probe        = "00000000-0000-4000-8000-000000000001"
	tenantID     = "00000000-0000-4000-8000-0000000000aa"
	jsonType     = "application/json; charset=utf-8"
	groupQuery   = "go-sdk-msteams acceptance policy group"
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
	// Historic fixtures capture response bytes. This private header follows the
	// independently audited Find wrapper/proxy path, not capture-time metadata.
	if command == "Find-CsGroup" {
		command = "Find-CsGroup_Find"
	}
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {targetForest}}
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	contentType := jsonType
	if reply == nil {
		contentType = ""
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: configOrigin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body},
		Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {contentType}}, Body: reply}}
}

func run(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family = "tenant_identity"
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
		t.Fatalf("tenant identity contract %s differed: %+v", c.Name, result.Differences())
	}
}

func mapping(command string, sets ...string) []contracttest.Mapping {
	if len(sets) == 0 {
		sets = []string{"__AllParameterSets"}
	}
	out := make([]contracttest.Mapping, 0, len(sets))
	for _, set := range sets {
		out = append(out, contracttest.Mapping{Command: command, ParameterSet: set})
	}
	return out
}

// apiError is the normalized error for a fixture: ConfigAPI reports the code
// at the top level or inside an "error" object.
func apiError(status int, body []byte) *client.APIError {
	var envelope struct {
		Code      string `json:"code"`
		ErrorCode *int   `json:"errorCode"`
		Error     struct {
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
	if envelope.ErrorCode != nil {
		result.ServiceErrorCodes = []int{*envelope.ErrorCode}
	}
	return result
}

func ptr[T any](v T) *T { return &v }

const groupPath = "/Teams.VoiceApps/groups?ExactMatchOnly=True&MaxResults=10&SearchQuery=go-sdk-msteams+acceptance+policy+group"

func TestUnitTenantIdentityReadsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	tenant, groups := readFile(t, "testdata/tenant.json"), readFile(t, "testdata/search_groups.json")
	empty := readFile(t, "testdata/search_groups_empty.json")
	run(t, contracttest.Case{Name: "actual-reads", Source: "msteams/teams_admin_api/tenant_identity/testdata/tenant.json",
		Mappings: append(mapping("Get-CsTenant", "Identity"), mapping("Find-CsGroup")...),
		Steps: []contracttest.Exchange{discovery(t),
			wire("GET", "/Teams.Tenant/tenants?defaultpropertyset=Extended", "Get-CsTenantObou_Get", nil, 200, tenant),
			wire("GET", groupPath, "Find-CsGroup", nil, 200, groups),
			wire("GET", "/Teams.VoiceApps/groups?SearchQuery=msteams-sdk-no-such-group-zz", "Find-CsGroup", nil, 200, empty)},
		Want: contracttest.Expected{Status: 200, Body: empty},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TenantIdentity
			value, _, err := s.GetTenant(ctx)
			if err != nil || *value.ObjectID != tenantID || len(value.AssignedPlans) != 5 || len(value.ProvisionedPlans) != 37 ||
				len(value.VerifiedDomains) != 3 || *value.VerifiedDomains[0].Status != "Enabled" || len(value.SIPDomains) != 2 ||
				*value.TeamsUpgradeEffectiveMode != "TeamsOnly" || *value.TeamsUpgradeNotificationsEnabled ||
				*value.AssignedPlans[0].Capability != "Teams" || *value.AssignedPlans[0].IsInGracePeriod ||
				len(value.LastProvisionTimeStamps) != 3 || len(value.LastPublishTimeStamps) != 5 || len(value.RawFields) != 25 {
				t.Fatal("actual tenant record was changed")
			}
			found, _, err := s.SearchGroups(ctx, ti.GroupSearchRequest{SearchQuery: groupQuery, ExactMatchOnly: ptr(true), MaxResults: ptr(uint32(10))})
			if err != nil || len(found) != 1 || *found[0].ID != "00000000-0000-4000-8000-31a3e60e6e84" || *found[0].DisplayName != "Fixture displayname" {
				t.Fatal("actual group search was changed")
			}
			none, response, err := s.SearchGroups(ctx, ti.GroupSearchRequest{SearchQuery: "msteams-sdk-no-such-group-zz"})
			if err != nil || none == nil || len(none) != 0 {
				t.Fatal("empty search was not a non-nil empty slice")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
}

type invocation func(context.Context, *ti.Service) (any, *resty.Response, error)

func TestUnitTenantIdentityDenialsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name, path, command, fixture string
		status                       int
		sets                         []string
		invoke                       invocation
	}{
		{"search-short-query", "/Teams.VoiceApps/groups?SearchQuery=ab", "Find-CsGroup", "search_groups_short_400.json", 400, nil,
			func(ctx context.Context, s *ti.Service) (any, *resty.Response, error) {
				return s.SearchGroups(ctx, ti.GroupSearchRequest{SearchQuery: "ab"})
			}},
		{"aad-tenant", "/Teams.Tenant/AadTenant", "Get-CsAadTenant", "aad_tenant_forbidden_403.json", 403, nil,
			func(ctx context.Context, s *ti.Service) (any, *resty.Response, error) { return s.GetAadTenant(ctx) }},
		{"aad-user", "/Teams.User/AadUser/" + probe, "Get-CsAadUser", "aad_user_forbidden_403.json", 403, nil,
			func(ctx context.Context, s *ti.Service) (any, *resty.Response, error) {
				return s.GetAadUser(ctx, probe)
			}},
		{"cloud-tenant", "/Teams.InternalSupport/GetAADTenantData", "Get-CsCloudTenant", "cloud_tenant_forbidden_403.json", 403, nil,
			func(ctx context.Context, s *ti.Service) (any, *resty.Response, error) { return s.GetCloudTenant(ctx) }},
		{"cloud-user", "/Teams.InternalSupport/GetAADUserData?userId=" + probe, "Get-CsCloudUser", "cloud_user_forbidden_403.json", 403, nil,
			func(ctx context.Context, s *ti.Service) (any, *resty.Response, error) {
				return s.GetCloudUser(ctx, probe)
			}},
		{"hosting-providers", "/Skype.Policy/configurations/HostingProvider", "Get-CsHostingProvider", "hosting_providers_forbidden_403.json", 403, nil,
			func(ctx context.Context, s *ti.Service) (any, *resty.Response, error) {
				return s.ListHostingProviders(ctx)
			}},
		{"tenant-licensing", "/Skype.Policy/configurations/TenantLicensingConfiguration", "Get-CsTenantLicensingConfiguration", "tenant_licensing_forbidden_403.json", 403, []string{"Filter"},
			func(ctx context.Context, s *ti.Service) (any, *resty.Response, error) {
				return s.ListTenantLicensingConfigurations(ctx)
			}},
		{"tenant-licensing-global", "/Skype.Policy/configurations/TenantLicensingConfiguration/configuration/Global", "Get-CsTenantLicensingConfiguration", "tenant_licensing_global_forbidden_403.json", 403, []string{"Identity"},
			func(ctx context.Context, s *ti.Service) (any, *resty.Response, error) {
				return s.GetTenantLicensingConfigurationByIdentity(ctx, "Global")
			}},
	} {
		reply := readFile(t, "testdata/"+c.fixture)
		run(t, contracttest.Case{Name: "actual-" + c.name, Source: "msteams/teams_admin_api/tenant_identity/testdata/" + c.fixture, Mappings: mapping(c.command, c.sets...),
			Steps: []contracttest.Exchange{discovery(t), wire("GET", c.path, c.command, nil, c.status, reply)},
			Want:  contracttest.Expected{Status: c.status, Body: reply, APIError: apiError(c.status, reply)},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				_, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.TenantIdentity)
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
}

func TestUnitTenantIdentitySuccessShapesPreserveFields(t *testing.T) {
	t.Parallel()
	directory := []byte(`{"objectId":"` + probe + `","displayName":"Contoso","city":null,"future":{"x":1}}`)
	configurations := []byte(`[{"Identity":"Tag:Custom","Enabled":true},{"Identity":"Global","Enabled":false},{"Identity":"Tag:Other","Enabled":true}]`)
	single := []byte(`{"Identity":"Global","Enabled":false}`)
	groups := []byte(`{"Groups":null,"Diagnostics":[{"x":1}]}`)
	run(t, contracttest.Case{Name: "success-shapes",
		Mappings: append(append(append(append(mapping("Get-CsAadTenant"), mapping("Get-CsAadUser")...), mapping("Get-CsCloudTenant")...), mapping("Get-CsCloudUser")...),
			append(append(mapping("Get-CsHostingProvider"), mapping("Get-CsTenantLicensingConfiguration", "Filter", "Identity")...), mapping("Find-CsGroup")...)...),
		Steps: []contracttest.Exchange{discovery(t),
			wire("GET", "/Teams.Tenant/AadTenant", "Get-CsAadTenant", nil, 200, directory),
			wire("GET", "/Teams.User/AadUser/user@example.invalid", "Get-CsAadUser", nil, 200, directory),
			wire("GET", "/Teams.InternalSupport/GetAADTenantData", "Get-CsCloudTenant", nil, 200, directory),
			wire("GET", "/Teams.InternalSupport/GetAADUserData?userId=user%40example.invalid", "Get-CsCloudUser", nil, 200, directory),
			wire("GET", "/Skype.Policy/configurations/HostingProvider", "Get-CsHostingProvider", nil, 200, configurations),
			wire("GET", "/Skype.Policy/configurations/HostingProvider/configuration/Tag:Custom", "Get-CsHostingProvider", nil, 200, single),
			wire("GET", "/Skype.Policy/configurations/TenantLicensingConfiguration", "Get-CsTenantLicensingConfiguration", nil, 200, configurations),
			wire("GET", "/Teams.VoiceApps/groups?MailEnabledOnly=False&SearchQuery=abc", "Find-CsGroup", nil, 200, groups)},
		Want: contracttest.Expected{Status: 200, Body: groups},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TenantIdentity
			check := func(value *ti.DirectoryObject, err error) {
				if err != nil || *value.ObjectID != probe || *value.DisplayName != "Contoso" || len(value.RawFields) != 4 || string(value.RawFields["future"]) != `{"x":1}` {
					t.Fatal("directory object fields were not preserved")
				}
			}
			value, _, err := s.GetAadTenant(ctx)
			check(value, err)
			value, _, err = s.GetAadUser(ctx, "user@example.invalid")
			check(value, err)
			value, _, err = s.GetCloudTenant(ctx)
			check(value, err)
			value, _, err = s.GetCloudUser(ctx, "user@example.invalid")
			check(value, err)
			listed, _, err := s.ListHostingProviders(ctx)
			if err != nil || len(listed) != 3 || *listed[0].Identity != "Global" || *listed[1].Identity != "Tag:Custom" || *listed[2].Identity != "Tag:Other" ||
				string(listed[1].RawFields["Enabled"]) != "true" {
				t.Fatal("configuration list was not ordered Global first")
			}
			one, _, err := s.GetHostingProviderByIdentity(ctx, "Tag:Custom")
			if err != nil || *one.Identity != "Global" {
				t.Fatal("configuration read was not decoded")
			}
			filtered, _, err := s.ListTenantLicensingConfigurationsFiltered(ctx, "Cust*")
			if err != nil || len(filtered) != 1 || *filtered[0].Identity != "Tag:Custom" {
				t.Fatal("local Tag: wildcard filter changed")
			}
			found, response, err := s.SearchGroups(ctx, ti.GroupSearchRequest{SearchQuery: "abc", MailEnabledOnly: ptr(false)})
			if err != nil || found == nil || len(found) != 0 {
				t.Fatal("null Groups was not an empty result")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
}

// TestUnitTenantIdentityGracePeriodRemovalWire proves the module body and
// status contract only; the acceptance suite never calls this licensing write.
func TestUnitTenantIdentityGracePeriodRemovalWire(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "grace-period-removal",
		Mappings: mapping("Remove-CsUserLicenseGracePeriod", "Remove", "RemoveExpanded", "RemoveViaIdentity", "RemoveViaIdentityExpanded"),
		Steps: []contracttest.Exchange{discovery(t),
			wire("PATCH", "/Teams.User/users/"+probe+"/AssignedPlans", "Remove-CsUserLicenseGracePeriod", json.RawMessage(`{}`), 204, nil),
			wire("PATCH", "/Teams.User/users/user@example.invalid/AssignedPlans", "Remove-CsUserLicenseGracePeriod", json.RawMessage(`{"action":"Remove","capabilities":["MCOEV","MCOMEETADV"]}`), 204, nil),
			wire("PATCH", "/Teams.User/users/"+probe+"/AssignedPlans", "Remove-CsUserLicenseGracePeriod", json.RawMessage(`{"capabilities":[]}`), 204, nil)},
		Want: contracttest.Expected{Status: 204},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TenantIdentity
			if _, err := s.RemoveUserLicenseGracePeriod(ctx, probe, ti.GracePeriodRemoval{}); err != nil {
				t.Fatal("empty removal was not accepted")
			}
			if _, err := s.RemoveUserLicenseGracePeriod(ctx, "user@example.invalid", ti.GracePeriodRemoval{Action: ptr("Remove"), Capabilities: []string{"MCOEV", "MCOMEETADV"}}); err != nil {
				t.Fatal("expanded removal was not accepted")
			}
			response, err := s.RemoveUserLicenseGracePeriod(ctx, probe, ti.GracePeriodRemoval{Capabilities: []string{}})
			return contracttest.Observation{Response: response, Err: err}
		}})
}

type countingClient struct{ calls int }

func (c *countingClient) Do(context.Context, client.Request, any) (*resty.Response, error) {
	c.calls++
	return nil, errors.New("unexpected I/O")
}

func TestUnitTenantIdentityRejectsInputBeforeIO(t *testing.T) {
	t.Parallel()
	base := &countingClient{}
	s := ti.NewService(base)
	ctx := t.Context()
	for name, call := range map[string]func() error{
		"empty-search": func() error { _, _, err := s.SearchGroups(ctx, ti.GroupSearchRequest{}); return err },
		"overflow-search": func() error {
			_, _, err := s.SearchGroups(ctx, ti.GroupSearchRequest{SearchQuery: "abc", MaxResults: ptr(uint32(2147483648))})
			return err
		},
		"blank-aad-user":    func() error { _, _, err := s.GetAadUser(ctx, ""); return err },
		"dot-aad-user":      func() error { _, _, err := s.GetAadUser(ctx, ".."); return err },
		"blank-cloud-user":  func() error { _, _, err := s.GetCloudUser(ctx, "\t"); return err },
		"blank-hosting":     func() error { _, _, err := s.GetHostingProviderByIdentity(ctx, ""); return err },
		"backslash-licence": func() error { _, _, err := s.GetTenantLicensingConfigurationByIdentity(ctx, `a\b`); return err },
		"bad-filter": func() error {
			_, _, err := s.ListTenantLicensingConfigurationsFiltered(ctx, "[")
			return err
		},
		"blank-grace-user": func() error { _, err := s.RemoveUserLicenseGracePeriod(ctx, "", ti.GracePeriodRemoval{}); return err },
		"control-action": func() error {
			_, err := s.RemoveUserLicenseGracePeriod(ctx, probe, ti.GracePeriodRemoval{Action: ptr("a\x00")})
			return err
		},
	} {
		if err := call(); !errors.Is(err, ti.ErrInput) {
			t.Fatalf("%s: got %v, want ErrInput", name, err)
		}
	}
	if err := s.GetOnlineDirectoryTenant(ctx); !errors.Is(err, ti.ErrOnlineDirectoryTenantDeprecated) {
		t.Fatal("online directory tenant was not the module deprecation")
	}
	if base.calls != 0 {
		t.Fatal("invalid input or deprecated command performed I/O")
	}
	var nilService *ti.Service
	if _, _, err := nilService.GetTenant(ctx); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service was not closed")
	}
}
