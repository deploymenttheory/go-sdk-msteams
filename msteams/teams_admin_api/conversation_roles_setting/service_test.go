package conversationrolessetting_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	roles "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/conversation_role"
	settings "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/conversation_roles_setting"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

const source = "msteams/teams_admin_api/conversation_roles_setting/service_test.go"
const route = "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ConversationRolesSetting"
const firstID = "00000000-0000-4000-8000-000000000001"
const nextID = "00000000-0000-4000-8000-000000000002"
const firstRole = `{"Identity":"Global/00000000-0000-4000-8000-000000000001","Id":"00000000-0000-4000-8000-000000000001","Name":"Controlled role","Permissions":{"Chat":{"AddParticipants":false},"Calling":{"SendAudio":true}},"IsAssignable":null,"CreatedAt":"2027-01-01T00:00:00","UpdatedAt":"2027-01-02T00:00:00Z"}`
const before = `{"Identity":"Global","ConversationRoles":[` + firstRole + `],"Future":[]}`
const empty = `{"Identity":"Global","ConversationRoles":[]}`

// Response bodies are authored source-constrained DTO vectors, never successful
// tenant evidence. Captured expected requests are offline pinned serializer output.
func wire(method, path, command, body string, status int, response string) contracttest.Exchange {
	headers := http.Header{"MPACmdlet": {"true"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": nil}
	var raw json.RawMessage
	if body != "" {
		raw = json.RawMessage(body)
		headers.Set("Content-Type", "application/json; charset=utf-8")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: route + path, Audience: auth.TeamsAdminAudience, Headers: headers, JSON: raw}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(response)}}
}
func run(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Source = source
	c.Family = "conversation_roles_setting"
	if c.Scope == "" {
		c.Scope = "service"
	}
	result, e := contracttest.Run(t.Context(), c)
	if e != nil || !result.Passed() {
		t.Fatalf("conversation settings vector differed: %v %+v", e, result.Differences())
	}
}
func ptr[T any](v T) *T { return &v }
func captured(t *testing.T, current, scenario string) string {
	t.Helper()
	raw, e := os.ReadFile("testdata/pinned-projections.json")
	if e != nil {
		t.Fatal(e)
	}
	var data struct {
		Cases []struct {
			Current, Scenario, Body string
			Accepted                bool
		}
	}
	if json.Unmarshal(raw, &data) != nil {
		t.Fatal("invalid source projection capture")
	}
	for _, v := range data.Cases {
		if v.Current == current && v.Scenario == scenario && v.Accepted {
			return v.Body
		}
	}
	t.Fatal("source projection case missing")
	return ""
}
func newRole(t *testing.T) *roles.Role {
	t.Helper()
	created, e := roles.ParseDateTime("2027-01-01T00:00:00")
	if e != nil {
		t.Fatal(e)
	}
	updated, e := roles.ParseDateTime("2027-01-02T00:00:00Z")
	if e != nil {
		t.Fatal(e)
	}
	permissions, e := roles.NewPermissions(roles.NewChatPermissions(roles.ChatPermissionsValues{AddParticipants: ptr(false), SendMessage: ptr(true)}), roles.NewCallingPermissions(roles.CallingPermissionsValues{AddParticipants: ptr(true), SendAudio: ptr(false)}))
	if e != nil {
		t.Fatal(e)
	}
	return &roles.Role{Identity: ptr("Global/" + nextID), ID: ptr(nextID), Name: ptr("Controlled role"), Permissions: permissions, CreatedAt: created, UpdatedAt: updated}
}
func TestUnitConversationSettingsGetSetsPreservePresenceAndLocalFilter(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "get-selectors", Steps: []contracttest.Exchange{
		wire("GET", "/configuration/Global", "Get-CsConversationRolesSetting", "", 200, before),
		wire("GET", "", "Get-CsConversationRolesSetting", "", 200, `[`+before+`]`),
		wire("GET", "", "Get-CsConversationRolesSetting", "", 200, `[`+before+`]`),
		wire("GET", "", "Get-CsConversationRolesSetting", "", 200, `[`+before+`]`),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		service := sdk.TeamsAdminAPI.ConversationRoleSettings
		value, _, e := service.GetByIdentity(ctx, "Global")
		if e != nil {
			return contracttest.Observation{Err: e}
		}
		if value == nil || len(value.ConversationRoles) != 1 || value.ConversationRoles[0].ID == nil || *value.ConversationRoles[0].ID != firstID || string(value.RawFields["Future"]) != "[]" {
			t.Fatal("typed list/unknown presence lost")
		}
		raw, e := json.Marshal(value)
		var actual, expected any
		if e != nil || json.Unmarshal(raw, &actual) != nil || json.Unmarshal([]byte(before), &expected) != nil || !reflect.DeepEqual(actual, expected) {
			t.Fatal("full settings response snapshot changed")
		}
		values, _, e := service.List(ctx)
		if e != nil {
			return contracttest.Observation{Err: e}
		}
		if len(values) != 1 {
			t.Fatal("list missing")
		}
		values, _, e = service.ListFiltered(ctx, "Glo*")
		if e != nil {
			return contracttest.Observation{Err: e}
		}
		if len(values) != 1 {
			t.Fatal("wildcard missing")
		}
		values, _, e = service.ListFiltered(ctx, " ")
		if e == nil && len(values) != 0 {
			t.Fatal("literal whitespace changed")
		}
		return contracttest.Observation{Err: e}
	}})
}
func TestUnitConversationSettingsFullTransitionPipelineMatchesPinnedBodies(t *testing.T) {
	t.Parallel()
	for _, current := range []string{"empty", "nonempty"} {
		for _, scenario := range []string{"null", "empty", "add", "replace", "remove", "replace-and-remove"} {
			t.Run(current+"-"+scenario, func(t *testing.T) {
				value := newRole(t)
				var modifier *settings.ListModifier
				switch scenario {
				case "empty":
					modifier = &settings.ListModifier{}
				case "add":
					modifier = &settings.ListModifier{Add: []*roles.Role{value}}
				case "replace":
					modifier = &settings.ListModifier{Replace: []*roles.Role{value}}
				case "remove":
					modifier = &settings.ListModifier{Remove: []*roles.Role{{ID: ptr(firstID)}}}
				case "replace-and-remove":
					modifier = &settings.ListModifier{Replace: []*roles.Role{value}, Remove: []*roles.Role{{ID: ptr(firstID)}}}
				}
				response := empty
				if current == "nonempty" {
					response = before
				}
				steps := []contracttest.Exchange{wire("GET", "/configuration/Global", "Set-CsConversationRolesSetting", "", 200, response)}
				want := contracttest.Expected{}
				rejected := current == "nonempty" && scenario == "remove"
				if rejected {
					want.Error = settings.ErrInput
				} else {
					steps = append(steps, wire("PATCH", "/configuration/Global", "Set-CsConversationRolesSetting", captured(t, current, scenario), 204, ""))
				}
				run(t, contracttest.Case{Name: current + "-" + scenario, Steps: steps, Want: want, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					input := settings.UpdateParameters{ConversationRoles: policy.Set(modifier)}
					if scenario == "null" {
						input.ConversationRoles = policy.Null[*settings.ListModifier]()
					}
					request, _, e := sdk.TeamsAdminAPI.ConversationRoleSettings.PrepareUpdateByIdentity(ctx, "Global", input)
					if rejected {
						if !errors.Is(e, want.Error) || request == nil || request.Before() == nil || request.ReadResponse() == nil {
							t.Fatal("comparison failure lost preparation")
						}
						return contracttest.Observation{Err: e}
					}
					if e != nil {
						return contracttest.Observation{Err: e}
					}
					if request.ResourceIdentity() != "Global" || request.Before() == nil {
						t.Fatal("preparation binding missing")
					}
					projection := request.Projection()
					if (scenario == "add" || scenario == "replace" || scenario == "replace-and-remove") && len(projection.OmittedPaths) < 2 {
						t.Fatal("retained/new role permission projection loss hidden")
					}
					projection.Body[0] = '['
					if modifier != nil && len(modifier.Add) > 0 {
						modifier.Add[0].Name = ptr("changed")
					}
					snapshot := request.Before()
					snapshot.Identity = ptr("retarget")
					result, _, e := sdk.TeamsAdminAPI.ConversationRoleSettings.UpdateByIdentity(ctx, "Global", *request)
					if e == nil && (result == nil || result.Before == nil || !result.PatchInvoked || !result.HTTPAccepted || result.Before.Identity == nil || *result.Before.Identity != "Global") {
						t.Fatal("frozen list request or stages changed")
					}
					return contracttest.Observation{Err: e}
				}})
			})
		}
	}
}
func TestUnitConversationSettingsNullOmittedAndEmptyResponseCollections(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{}`, `{"ConversationRoles":null}`, `{"ConversationRoles":[]}`} {
		var value settings.Configuration
		if json.Unmarshal([]byte(raw), &value) != nil || value.ConversationRoles == nil || len(value.ConversationRoles) != 0 {
			t.Fatal("CLR empty collection projection changed")
		}
		encoded, e := json.Marshal(value)
		var actual, expected any
		if e != nil || json.Unmarshal(encoded, &actual) != nil || json.Unmarshal([]byte(raw), &expected) != nil || !reflect.DeepEqual(actual, expected) {
			t.Fatal("null/omitted/empty raw presence lost")
		}
	}
	var value settings.Configuration
	if json.Unmarshal([]byte(`{"ConversationRoles":[null]}`), &value) != nil || len(value.ConversationRoles) != 1 || value.ConversationRoles[0] != nil {
		t.Fatal("source-accepted null element rejected")
	}
	if _, _, e := settings.ProjectListValue(nil, nil); !errors.Is(e, settings.ErrInput) {
		t.Fatal("null current helper collection accepted")
	}
}
func TestUnitConversationSettingsDenialAcceptedMalformedAndRetargetRetainStages(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"denied", "malformed", "retarget"} {
		t.Run(scenario, func(t *testing.T) {
			steps := []contracttest.Exchange{wire("GET", "/configuration/Global", "Set-CsConversationRolesSetting", "", 200, before)}
			want := settings.ErrInput
			if scenario == "denied" {
				steps = []contracttest.Exchange{wire("GET", "/configuration/Global", "Set-CsConversationRolesSetting", "", 403, `{"code":"Forbidden"}`)}
				want = auth.ErrForbidden
			}
			if scenario == "malformed" {
				steps = append(steps, wire("PATCH", "/configuration/Global", "Set-CsConversationRolesSetting", captured(t, "nonempty", "null"), 200, `[]`))
				want = client.ErrDecode
			}
			run(t, contracttest.Case{Name: scenario, Steps: steps, Want: contracttest.Expected{Error: want}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				request, response, e := sdk.TeamsAdminAPI.ConversationRoleSettings.PrepareUpdateByIdentity(ctx, "Global", settings.UpdateParameters{ConversationRoles: policy.Null[*settings.ListModifier]()})
				if scenario == "denied" {
					if !errors.Is(e, want) || request == nil || request.ReadResponse() == nil || response == nil {
						t.Fatal("denied preparation receipt missing")
					}
					return contracttest.Observation{Err: e}
				}
				if e != nil {
					return contracttest.Observation{Err: e}
				}
				target := "Global"
				if scenario == "retarget" {
					target = "Other"
				}
				result, response, e := sdk.TeamsAdminAPI.ConversationRoleSettings.UpdateByIdentity(ctx, target, *request)
				if !errors.Is(e, want) || result == nil || result.Before == nil || result.ReadResponse == nil {
					t.Fatal("failed apply lost its preparation")
				}
				if scenario == "retarget" && result.PatchInvoked {
					t.Fatal("prepared request retargeted")
				}
				if scenario == "malformed" && (!result.PatchInvoked || !result.HTTPAccepted || result.After != nil || response == nil) {
					t.Fatal("accepted malformed result lost stages")
				}
				return contracttest.Observation{Err: e}
			}})
		})
	}
}
func TestUnitConversationSettingsUnboundRejectsBeforeCredentialsAndEmptyFilterIsLocal(t *testing.T) {
	t.Parallel()
	flag := false
	for _, controls := range []policy.Controls{{}, {Force: &flag}} {
		run(t, contracttest.Case{Name: "unbound-set", Scope: "local", Want: contracttest.Expected{Error: settings.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			request, response, e := sdk.TeamsAdminAPI.ConversationRoleSettings.PrepareUpdateByIdentity(ctx, "Global", settings.UpdateParameters{Controls: controls})
			if !errors.Is(e, settings.ErrInput) || request != nil || response != nil {
				t.Fatal("unbound settings did not reject before authentication or preparation")
			}
			return contracttest.Observation{Err: e}
		}})
	}
	run(t, contracttest.Case{Name: "empty-filter", Scope: "local", Want: contracttest.Expected{Error: settings.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, _, e := sdk.TeamsAdminAPI.ConversationRoleSettings.ListFiltered(ctx, "")
		return contracttest.Observation{Err: e}
	}})
}

func TestUnitConversationSettingsReplacementMatchesActualCompiledBinding(t *testing.T) {
	t.Parallel()
	raw, e := os.ReadFile("testdata/pinned-binding-replacement.json")
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct{ Body string }
	if json.Unmarshal(raw, &fixture) != nil || fixture.Body == "" {
		t.Fatal("invalid actual-binding capture")
	}
	value := newRole(t)
	value.ID = ptr(firstID)
	value.Identity = ptr("Global/" + firstID)
	run(t, contracttest.Case{Name: "actual-bound-list-replacement", Steps: []contracttest.Exchange{
		wire("GET", "/configuration/Global", "Set-CsConversationRolesSetting", "", 200, empty),
		wire("PATCH", "/configuration/Global", "Set-CsConversationRolesSetting", fixture.Body, 204, ""),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		service := sdk.TeamsAdminAPI.ConversationRoleSettings
		request, _, e := service.PrepareUpdateByIdentity(ctx, "Global", settings.UpdateParameters{ConversationRoles: policy.Set(&settings.ListModifier{Replace: []*roles.Role{value}})})
		if e != nil {
			return contracttest.Observation{Err: e}
		}
		if len(request.Projection().OmittedPaths) != 2 {
			t.Fatal("actual nested permission omission was hidden")
		}
		result, _, e := service.UpdateByIdentity(ctx, "Global", *request)
		if e == nil && (result == nil || !result.HTTPAccepted) {
			t.Fatal("binding-replay action receipt missing")
		}
		return contracttest.Observation{Err: e}
	}})
}
