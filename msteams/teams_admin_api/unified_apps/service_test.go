package unifiedapps_test

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
	ua "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/unified_apps"
	"github.com/go-resty/resty/v2"
)

const (
	origin   = "https://substrate.office.com"
	audience = "08ff1ce2-4973-4b08-86a3-ebed13badc7f"
	scope    = audience + "/AdminAppCatalog.ReadWrite.All"
	admin    = "/AdminAppCatalog/ps/v2/admin/unifiedApp"
	source   = "msteams/teams_admin_api/unified_apps/service_test.go"
	app      = "00000000-0000-4000-8000-000000000001"
	pending  = "00000000-0000-4000-8000-000000000006"
	user     = "00000000-0000-4000-8000-000000000004"
	group    = "00000000-0000-4000-8000-000000000005"
	jsonType = "application/json; charset=utf-8"
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

func wire(method, path, command string, body json.RawMessage, status int, reply []byte) contracttest.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": nil}
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	contentType := jsonType
	if reply == nil {
		contentType = ""
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: audience, TokenScope: scope, Headers: h, JSON: body},
		Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {contentType}}, Body: reply}}
}

func run(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family = "unified_apps"
	if c.Source == "" {
		c.Source = source
	}
	if c.Scope == "" {
		c.Scope = "service"
	}
	if c.Mode == "" {
		c.Mode = auth.Delegated
	}
	result, err := contracttest.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Fatalf("unified apps contract %s differed: %+v", c.Name, result.Differences())
	}
}

func mapping(commands ...string) []contracttest.Mapping {
	out := make([]contracttest.Mapping, 0, len(commands))
	for _, command := range commands {
		out = append(out, contracttest.Mapping{Command: command, ParameterSet: "__AllParameterSets"})
	}
	return out
}

func ptr[T any](v T) *T { return &v }

func TestUnitUnifiedAppReadsDecodeCatalogShapes(t *testing.T) {
	t.Parallel()
	apps, one, staged, settings := readFile(t, "testdata/apps.json"), readFile(t, "testdata/app.json"), readFile(t, "testdata/pending.json"), readFile(t, "testdata/settings.json")
	run(t, contracttest.Case{Name: "reads",
		Mappings: mapping("Get-AllM365TeamsApps", "Get-M365TeamsApp", "Get-M365UnifiedCustomPendingApps", "Get-M365UnifiedTenantSettings"),
		Steps: []contracttest.Exchange{
			wire("GET", admin+"?%24expand=DeployedTo%2CAvailableTo%2Cmetadata", "Get-AllM365TeamsApps", nil, 200, apps),
			wire("GET", admin+"/"+app+"?%24expand=metadata", "Get-M365TeamsApp", nil, 200, one),
			wire("GET", admin+"/stagedApps", "Get-M365UnifiedCustomPendingApps", nil, 200, staged),
			wire("GET", admin+"/settings", "Get-M365UnifiedTenantSettings", nil, 200, settings),
			wire("GET", admin+"/settings?SettingNames=DefaultApp%2CEnableCopilotExtensibility", "Get-M365UnifiedTenantSettings", nil, 200, settings)},
		Want: contracttest.Expected{Status: 200, Body: settings},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.UnifiedApps
			list, _, err := s.List(ctx)
			if err != nil || len(list) != 2 || *list[0].AvailableTo.AssignmentType != ua.TypeEveryone || list[0].InstalledFor.Version != nil ||
				list[0].RawFields["metadata"] == nil || !*list[1].IsBlocked || list[1].AvailableTo != nil || string(list[1].RawFields["futureField"]) != "7" {
				t.Fatalf("list decode changed: %v", err)
			}
			value, _, err := s.Get(ctx, app)
			if err != nil || *value.AppType != "ThirdParty" || value.AvailableTo.Users[0].ID[0] != user || len(value.AvailableTo.Groups) != 0 ||
				*value.InstalledFor.InstalledSource != "Admin" || value.InstalledFor.Groups[0].ID[0] != group || value.InstalledFor.Groups[0].LastUpdatedDate != nil {
				t.Fatalf("get decode changed: %v", err)
			}
			waiting, _, err := s.ListCustomPending(ctx)
			if err != nil || len(waiting) != 1 || *waiting[0].Iteration != "3" || *waiting[0].Metadata.CopilotEnabled || waiting[0].Metadata.CompatibleHubs[0] != "Teams" {
				t.Fatalf("pending decode changed: %v", err)
			}
			all, _, err := s.GetTenantSettings(ctx)
			if err != nil || len(all) != 2 || *all[1].SettingValue != ua.SettingValueSome || all[1].Users[0] != user || all[0].Users != nil {
				t.Fatalf("settings decode changed: %v", err)
			}
			named, response, err := s.GetTenantSettings(ctx, ua.SettingDefaultApp, ua.SettingEnableCopilotExtensibility)
			return contracttest.Observation{Value: named, Response: response, Err: err}
		}})
}

type write func(context.Context, *ua.Service) (*resty.Response, error)

func TestUnitUnifiedAppWritesSendModuleBodies(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name, method, path, command, body string
		invoke                            write
	}{
		{"block", "PATCH", admin, "Update-M365TeamsApp", `{"id":"` + app + `","isBlocked":true}`,
			func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
				return s.Update(ctx, app, ua.AppUpdate{IsBlocked: ptr(true)})
			}},
		{"available-users-and-groups", "PATCH", admin, "Update-M365TeamsApp",
			`{"id":"` + app + `","availableTo":{"type":"UsersAndGroups","users":["` + user + `"],"groups":["` + group + `"],"operation":"remove"}}`,
			func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
				return s.Update(ctx, app, ua.AppUpdate{AssignmentType: ua.TypeUsersAndGroups, Operation: "Remove", Users: []string{user, user}, Groups: []string{group}})
			}},
		{"everyone-and-install", "PATCH", admin, "Update-M365TeamsApp",
			`{"id":"` + app + `","isBlocked":false,"availableTo":{"type":"everyone","operation":"add"},"deployTo":{"type":"UsersAndGroups","users":["` + user + `"],"operation":"add","version":"1.0.0"}}`,
			func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
				return s.Update(ctx, app, ua.AppUpdate{IsBlocked: ptr(false), AssignmentType: "everyone", InstallType: ua.TypeUsersAndGroups, InstallOperation: ua.OperationAdd, InstallUsers: []string{user}, InstallVersion: "1.0.0"})
			}},
		{"install-noone", "PATCH", admin, "Update-M365TeamsApp", `{"id":"` + app + `","deployTo":{"type":"Noone","operation":"add"}}`,
			func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
				return s.Update(ctx, app, ua.AppUpdate{InstallType: ua.TypeNoone})
			}},
		{"settings-default", "PATCH", admin + "/settings", "Update-M365UnifiedTenantSettings", `{"settingName":"DefaultApp","settingValue":"None"}`,
			func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
				return s.UpdateTenantSettings(ctx, ua.TenantSettingsUpdate{SettingName: ua.SettingDefaultApp, SettingValue: ua.SettingValueNone})
			}},
		{"settings-copilot", "PATCH", admin + "/settings", "Update-M365UnifiedTenantSettings",
			`{"settingName":"EnableCopilotExtensibility","settingValue":"Some","groups":["` + group + `"],"operation":"add"}`,
			func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
				return s.UpdateTenantSettings(ctx, ua.TenantSettingsUpdate{SettingName: ua.SettingEnableCopilotExtensibility, SettingValue: ua.SettingValueSome, Groups: []string{group}, Operation: ua.OperationAdd})
			}},
	} {
		run(t, contracttest.Case{Name: c.name, Mappings: mapping(c.command),
			Steps: []contracttest.Exchange{wire(c.method, c.path, c.command, json.RawMessage(c.body), 204, nil)},
			Want:  contracttest.Expected{Status: 204},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				response, err := c.invoke(ctx, sdk.TeamsAdminAPI.UnifiedApps)
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
}

func TestUnitUnifiedAppReviewReadsEtagThenPuts(t *testing.T) {
	t.Parallel()
	const command = "Update-M365UnifiedCustomPendingApp"
	staged, details := readFile(t, "testdata/pending.json"), readFile(t, "testdata/pending_details.json")
	detailsPath := "/AdminAppCatalog/v2/admin/unifiedApps/stagedApps/" + pending + "/details"
	run(t, contracttest.Case{Name: "review-published", Mappings: mapping(command),
		Steps: []contracttest.Exchange{
			wire("GET", admin+"/stagedApps", command, nil, 200, staged),
			wire("GET", detailsPath, command, nil, 200, details),
			wire("PUT", admin+"/stagedApps/review", command, json.RawMessage(`{"Id":"`+pending+`","ReviewStatus":"Published","StagedAppDefinitionEtag":"3"}`), 200, []byte(`{}`))},
		Want: contracttest.Expected{Status: 200},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.UnifiedApps.ReviewCustomPending(ctx, pending, "PUBLISHED")
			return contracttest.Observation{Response: response, Err: err}
		}})
	run(t, contracttest.Case{Name: "review-unknown-app", Mappings: mapping(command),
		Steps: []contracttest.Exchange{wire("GET", admin+"/stagedApps", command, nil, 200, staged)},
		Want:  contracttest.Expected{Status: 200, Error: ua.ErrPendingAppNotFound},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.UnifiedApps.ReviewCustomPending(ctx, app, ua.ReviewRejected)
			return contracttest.Observation{Response: response, Err: err}
		}})
	for _, body := range []string{`{}`, `{"stagedAppEntitlement":{"iteration":""}}`, `{"stagedAppEntitlement":null}`} {
		run(t, contracttest.Case{Name: "review-missing-iteration", Mappings: mapping(command),
			Steps: []contracttest.Exchange{wire("GET", admin+"/stagedApps", command, nil, 200, staged), wire("GET", detailsPath, command, nil, 200, []byte(body))},
			Want:  contracttest.Expected{Status: 200, Error: ua.ErrMissingIteration},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				response, err := sdk.TeamsAdminAPI.UnifiedApps.ReviewCustomPending(ctx, pending, ua.ReviewRejected)
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
}

func TestUnitUnifiedAppFailuresStayTyped(t *testing.T) {
	t.Parallel()
	forbidden := readFile(t, "testdata/forbidden_403.json")
	run(t, contracttest.Case{Name: "forbidden", Mappings: mapping("Get-AllM365TeamsApps"),
		Steps: []contracttest.Exchange{wire("GET", admin+"?%24expand=DeployedTo%2CAvailableTo%2Cmetadata", "Get-AllM365TeamsApps", nil, 403, forbidden)},
		Want:  contracttest.Expected{Status: 403, Body: forbidden, Error: auth.ErrForbidden},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.UnifiedApps.List(ctx)
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}})
	for _, body := range []string{`null`, `{}`, `[{"id":7}]`, `[{"availableTo":{"users":{}}}]`} {
		run(t, contracttest.Case{Name: "malformed-list", Mappings: mapping("Get-AllM365TeamsApps"),
			Steps: []contracttest.Exchange{wire("GET", admin+"?%24expand=DeployedTo%2CAvailableTo%2Cmetadata", "Get-AllM365TeamsApps", nil, 200, []byte(body))},
			Want:  contracttest.Expected{Status: 200, Error: client.ErrDecode},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.UnifiedApps.List(ctx)
				return contracttest.Observation{Value: value, Response: response, Err: err}
			}})
	}
	for _, body := range []string{`null`, `[]`, `{"deployedTo":{"users":[{"id":"not-an-array"}]}}`} {
		run(t, contracttest.Case{Name: "malformed-get", Mappings: mapping("Get-M365TeamsApp"),
			Steps: []contracttest.Exchange{wire("GET", admin+"/"+app+"?%24expand=metadata", "Get-M365TeamsApp", nil, 200, []byte(body))},
			Want:  contracttest.Expected{Status: 200, Error: client.ErrDecode},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.UnifiedApps.Get(ctx, app)
				return contracttest.Observation{Value: value, Response: response, Err: err}
			}})
	}
	run(t, contracttest.Case{Name: "undeclared-write-status", Mappings: mapping("Update-M365TeamsApp"),
		Steps: []contracttest.Exchange{wire("PATCH", admin, "Update-M365TeamsApp", json.RawMessage(`{"id":"`+app+`","isBlocked":true}`), 201, []byte(`{}`))},
		Want:  contracttest.Expected{Status: 201, Error: ua.ErrResponse},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.UnifiedApps.Update(ctx, app, ua.AppUpdate{IsBlocked: ptr(true)})
			return contracttest.Observation{Response: response, Err: err}
		}})
}

func TestUnitUnifiedAppInvalidInputFailsLocally(t *testing.T) {
	t.Parallel()
	for name, invoke := range map[string]write{
		"update-nothing": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			return s.Update(ctx, app, ua.AppUpdate{})
		},
		"update-blank-id": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			return s.Update(ctx, " ", ua.AppUpdate{IsBlocked: ptr(true)})
		},
		"update-dot-id": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			return s.Update(ctx, "..", ua.AppUpdate{IsBlocked: ptr(true)})
		},
		"update-unknown-type": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			return s.Update(ctx, app, ua.AppUpdate{AssignmentType: "Some"})
		},
		"update-missing-operation": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			return s.Update(ctx, app, ua.AppUpdate{AssignmentType: ua.TypeUsersAndGroups, Users: []string{user}})
		},
		"update-missing-members": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			return s.Update(ctx, app, ua.AppUpdate{AssignmentType: ua.TypeUsersAndGroups, Operation: ua.OperationAdd})
		},
		"update-everyone-members": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			return s.Update(ctx, app, ua.AppUpdate{AssignmentType: ua.TypeEveryone, Users: []string{user}})
		},
		"update-orphan-members": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			return s.Update(ctx, app, ua.AppUpdate{IsBlocked: ptr(true), Users: []string{user}})
		},
		"update-orphan-version": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			return s.Update(ctx, app, ua.AppUpdate{IsBlocked: ptr(true), InstallVersion: "1.0"})
		},
		"update-install-noone-user": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			return s.Update(ctx, app, ua.AppUpdate{InstallType: ua.TypeNoone, InstallGroups: []string{group}})
		},
		"update-blank-member": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			return s.Update(ctx, app, ua.AppUpdate{AssignmentType: ua.TypeUsersAndGroups, Operation: ua.OperationAdd, Users: []string{""}})
		},
		"review-status": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			return s.ReviewCustomPending(ctx, pending, "Pending")
		},
		"review-id": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			return s.ReviewCustomPending(ctx, "bad\nid", ua.ReviewPublished)
		},
		"settings-name": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			return s.UpdateTenantSettings(ctx, ua.TenantSettingsUpdate{SettingName: "defaultapp", SettingValue: ua.SettingValueAll})
		},
		"settings-value": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			return s.UpdateTenantSettings(ctx, ua.TenantSettingsUpdate{SettingName: ua.SettingGlobalApp})
		},
		"settings-some": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			return s.UpdateTenantSettings(ctx, ua.TenantSettingsUpdate{SettingName: ua.SettingPrivateApp, SettingValue: ua.SettingValueSome})
		},
		"settings-members": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			return s.UpdateTenantSettings(ctx, ua.TenantSettingsUpdate{SettingName: ua.SettingDefaultApp, SettingValue: ua.SettingValueAll, Users: []string{user}})
		},
		"settings-operation": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			return s.UpdateTenantSettings(ctx, ua.TenantSettingsUpdate{SettingName: ua.SettingEnableCopilotExtensibility, SettingValue: ua.SettingValueSome, Operation: "Add"})
		},
		"read-setting-name": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			_, r, err := s.GetTenantSettings(ctx, "Unknown")
			return r, err
		},
		"read-app-control-id": func(ctx context.Context, s *ua.Service) (*resty.Response, error) {
			_, r, err := s.Get(ctx, "id\t")
			return r, err
		},
	} {
		run(t, contracttest.Case{Name: name, Scope: "local",
			Want: contracttest.Expected{Error: ua.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				response, err := invoke(ctx, sdk.TeamsAdminAPI.UnifiedApps)
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
}

func TestUnitUnifiedAppProfileRejectsBeforeTokens(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "application-mode", Scope: "local", Mode: auth.Application,
		Want: contracttest.Expected{Error: client.ErrAuthenticationMode},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.UnifiedApps.List(ctx)
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}})
	for _, cloud := range []string{"gcc", "gcch", "dod", "china"} {
		run(t, contracttest.Case{Name: cloud, Cloud: cloud, Scope: "local",
			Want: contracttest.Expected{Error: client.ErrDestination},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.UnifiedApps.GetTenantSettings(ctx)
				return contracttest.Observation{Value: value, Response: response, Err: err}
			}})
	}
}

func TestUnitUnifiedAppLifecycleGuards(t *testing.T) {
	t.Parallel()
	var s *ua.Service
	//nolint:staticcheck // the nil context is the case under test.
	if _, _, err := s.List(nil); !errors.Is(err, client.ErrNilContext) {
		t.Fatalf("nil context: %v", err)
	}
	if _, err := s.Update(t.Context(), app, ua.AppUpdate{IsBlocked: ptr(true)}); !errors.Is(err, client.ErrClosed) {
		t.Fatalf("nil service: %v", err)
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, _, err := ua.NewService(nil).ListCustomPending(ctx); !errors.Is(err, context.Canceled) {
		t.Fatalf("cancelled context: %v", err)
	}
	if _, err := ua.NewService(nil).ReviewCustomPending(t.Context(), pending, ua.ReviewPublished); !errors.Is(err, client.ErrClosed) {
		t.Fatalf("unbound service: %v", err)
	}
}
