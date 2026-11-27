package customappsettings_test

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
	cas "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/custom_app_settings"
)

const (
	configOrigin = "https://api.interfaces.records.teams.microsoft.com"
	targetForest = "https://admin.example.invalid/"
	source       = "msteams/teams_admin_api/custom_app_settings/service_test.go"
	fixture      = "msteams/teams_admin_api/custom_app_settings/testdata/get.json"
	route        = "/Teams.MiddletierService/tenantWideAppsSettingsGlobal"
	getCommand   = "Get-CsTeamsSettingsCustomApp"
	setCommand   = "Set-CsTeamsSettingsCustomApp"
)

func readFile(t *testing.T, name string) []byte {
	t.Helper()
	root, err := os.OpenRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	body, err := root.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return body
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

func wire(method, command string, body json.RawMessage, status int, reply []byte) contracttest.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {targetForest}}
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: configOrigin + route, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body},
		Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json; charset=utf-8"}}, Body: reply}}
}

func run(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family = "custom_app_settings"
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
		t.Fatalf("custom app settings contract %s differed: %+v", c.Name, result.Differences())
	}
}

func mapping(command string) []contracttest.Mapping {
	return []contracttest.Mapping{{Command: command, ParameterSet: "__AllParameterSets"}}
}

func TestUnitCustomAppSettingsGetDecodesTheLabShape(t *testing.T) {
	t.Parallel()
	body := readFile(t, "testdata/get.json")
	run(t, contracttest.Case{Name: "get-lab-shape", Source: fixture, Mappings: mapping(getCommand),
		Steps: []contracttest.Exchange{discovery(t), wire("GET", getCommand, nil, 200, body)},
		Want:  contracttest.Expected{Status: 200, Body: body},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.CustomAppSettings.Get(ctx)
			if err != nil || !value.SideloadedAppsInteractionEnabled() || value.IsTenantWideAutoInstallEnabled == nil || *value.IsTenantWideAutoInstallEnabled ||
				value.LobLogo == nil || *value.LobLogo != "" || len(value.AppSettingsList) != 2 || *value.AppSettingsList[1].IsEnabled ||
				*value.AppSettingsList[0].Context != "GlobalApp" || value.AppSettingsList[0].RawFields["isAppLevelAutoInstallEnabled"] == nil || value.RawFields["appSettingsList"] == nil {
				t.Fatal("custom app settings fields were lost")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitCustomAppSettingsProjectionReportsNullAsFalse(t *testing.T) {
	t.Parallel()
	for name, body := range map[string]string{"null": `{"isSideloadedAppsInteractionEnabled":null}`, "absent": `{"isAppsEnabled":true}`, "false": `{"isSideloadedAppsInteractionEnabled":false}`, "wrong-case": `{"IsSideloadedAppsInteractionEnabled":true}`, "mixed-case": `{"isSideloadedAppsInteractionEnabled":false,"IsSideloadedAppsInteractionEnabled":true}`} {
		run(t, contracttest.Case{Name: "projection-" + name, Mappings: mapping(getCommand),
			Steps: []contracttest.Exchange{discovery(t), wire("GET", getCommand, nil, 200, []byte(body))},
			Want:  contracttest.Expected{Status: 200, Body: []byte(body)},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.CustomAppSettings.Get(ctx)
				if err != nil || value.SideloadedAppsInteractionEnabled() {
					t.Fatal("missing sideloading flag was not projected as false")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
	var missing *cas.Settings
	if missing.SideloadedAppsInteractionEnabled() {
		t.Fatal("nil settings projected true")
	}
}

// The module reads the settings, then PUTs the six switches and
// appAccessRequestConfig with the new sideloading flag and an empty
// appSettingsList. Lob* strings are not sent.
func TestUnitCustomAppSettingsSetUsesTheModuleReadReplaceFlow(t *testing.T) {
	t.Parallel()
	current := readFile(t, "testdata/get.json")
	put := json.RawMessage(`{"isAppsEnabled":true,"isAppsPurchaseEnabled":true,"isExternalAppsEnabledByDefault":true,"isLicenseBasedPinnedAppsEnabled":true,"isSideloadedAppsInteractionEnabled":false,"isTenantWideAutoInstallEnabled":false,"appSettingsList":[]}`)
	run(t, contracttest.Case{Name: "set-sideloading", Source: fixture, Mappings: mapping(setCommand),
		Steps: []contracttest.Exchange{discovery(t), wire("GET", getCommand, nil, 200, current), wire("PUT", setCommand, put, 204, nil)},
		Want:  contracttest.Expected{Status: 204},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.CustomAppSettings.SetSideloadedAppsInteraction(ctx, false)
			return contracttest.Observation{Response: response, Err: err}
		}})

	withConfig := []byte(`{"isSideloadedAppsInteractionEnabled":false,"appAccessRequestConfig":{"adminInstructionMessage":"Ask IT","approvalPortalUrl":"https://portal.example.invalid/","future":1}}`)
	put = json.RawMessage(`{"isSideloadedAppsInteractionEnabled":true,"appSettingsList":[],"appAccessRequestConfig":{"adminInstructionMessage":"Ask IT","approvalPortalUrl":"https://portal.example.invalid/","future":1}}`)
	run(t, contracttest.Case{Name: "set-keeps-access-request-config", Mappings: mapping(setCommand),
		Steps: []contracttest.Exchange{discovery(t), wire("GET", getCommand, nil, 200, withConfig), wire("PUT", setCommand, put, 204, nil)},
		Want:  contracttest.Expected{Status: 204},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.CustomAppSettings.SetSideloadedAppsInteraction(ctx, true)
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitCustomAppSettingsUpdateSendsTheGivenDocument(t *testing.T) {
	t.Parallel()
	yes, no, color, scope := true, false, "#FFFFFF", "GlobalApp"
	put := json.RawMessage(`{"isAppsEnabled":true,"isTenantWideAutoInstallEnabled":false,"LobTextColor":"#FFFFFF","appSettingsList":[{"id":"00000000-0000-4000-8000-000000000101","isEnabled":false,"context":"GlobalApp"}]}`)
	run(t, contracttest.Case{Name: "update-document", Mappings: mapping(setCommand),
		Steps: []contracttest.Exchange{discovery(t), wire("PUT", setCommand, put, 204, nil)},
		Want:  contracttest.Expected{Status: 204},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.CustomAppSettings.Update(ctx, cas.UpdateRequest{IsAppsEnabled: &yes, IsTenantWideAutoInstallEnabled: &no, LobTextColor: &color,
				AppSettingsList: []cas.AppSettingUpdate{{ID: "00000000-0000-4000-8000-000000000101", IsEnabled: &no, Context: &scope}}})
			return contracttest.Observation{Response: response, Err: err}
		}})
	run(t, contracttest.Case{Name: "update-nil-list-is-empty", Mappings: mapping(setCommand),
		Steps: []contracttest.Exchange{discovery(t), wire("PUT", setCommand, json.RawMessage(`{"appSettingsList":[]}`), 204, nil)},
		Want:  contracttest.Expected{Status: 204},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.CustomAppSettings.Update(ctx, cas.UpdateRequest{})
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitCustomAppSettingsFailuresStayTyped(t *testing.T) {
	t.Parallel()
	for name, c := range map[string]struct {
		body   string
		status int
		want   error
	}{
		"get-array":       {`[]`, 200, client.ErrDecode},
		"get-null":        {`null`, 200, client.ErrDecode},
		"get-bad-flag":    {`{"isAppsEnabled":"yes"}`, 200, client.ErrDecode},
		"get-null-entry":  {`{"appSettingsList":[null]}`, 200, client.ErrDecode},
		"get-undeclared":  {`{}`, 202, cas.ErrResponse},
		"get-bad-request": {`{"code":"BadRequest"}`, 400, nil},
	} {
		want := contracttest.Expected{Status: c.status, Error: c.want}
		if c.want == nil {
			want = contracttest.Expected{Status: c.status, Body: []byte(c.body), APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: c.status, Codes: []string{"BadRequest"}}}
		}
		run(t, contracttest.Case{Name: name, Mappings: mapping(getCommand),
			Steps: []contracttest.Exchange{discovery(t), wire("GET", getCommand, nil, c.status, []byte(c.body))},
			Want:  want,
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.CustomAppSettings.Get(ctx)
				if value != nil {
					t.Fatal("failed response returned a value")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
	denied := []byte(`{"code":"Forbidden"}`)
	run(t, contracttest.Case{Name: "set-stops-when-read-fails", Mappings: mapping(setCommand),
		Steps: []contracttest.Exchange{discovery(t), wire("GET", getCommand, nil, 403, denied)},
		Want:  contracttest.Expected{Status: 403, Body: denied, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 403, Codes: []string{"Forbidden"}}},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.CustomAppSettings.SetSideloadedAppsInteraction(ctx, true)
			if !errors.Is(err, auth.ErrForbidden) {
				t.Fatal("403 was not classified as forbidden")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
	run(t, contracttest.Case{Name: "update-undeclared-status", Mappings: mapping(setCommand),
		Steps: []contracttest.Exchange{discovery(t), wire("PUT", setCommand, json.RawMessage(`{"appSettingsList":[]}`), 202, nil)},
		Want:  contracttest.Expected{Status: 202, Error: cas.ErrResponse},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.CustomAppSettings.Update(ctx, cas.UpdateRequest{})
			return contracttest.Observation{Response: response, Err: err}
		}})
	run(t, contracttest.Case{Name: "update-undeclared-200", Mappings: mapping(setCommand),
		Steps: []contracttest.Exchange{discovery(t), wire("PUT", setCommand, json.RawMessage(`{"appSettingsList":[]}`), 200, nil)},
		Want:  contracttest.Expected{Status: 200, Error: cas.ErrResponse},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.CustomAppSettings.Update(ctx, cas.UpdateRequest{})
			return contracttest.Observation{Response: response, Err: err}
		}})
	run(t, contracttest.Case{Name: "update-entry-without-id", Scope: "local", Mappings: mapping(setCommand), Steps: []contracttest.Exchange{},
		Want: contracttest.Expected{Error: cas.ErrInput},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.CustomAppSettings.Update(ctx, cas.UpdateRequest{AppSettingsList: []cas.AppSettingUpdate{{}}})
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitCustomAppSettingsNilContextAndServiceAreRejected(t *testing.T) {
	t.Parallel()
	var s *cas.Service
	//nolint:staticcheck // A nil context is the guarded input.
	if _, _, err := s.Get(nil); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, err := s.Update(t.Context(), cas.UpdateRequest{}); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, err := cas.NewService(nil).SetSideloadedAppsInteraction(ctx, true); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled context accepted")
	}
}
