package conversationrole_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	roles "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/conversation_role"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

const source = "msteams/teams_admin_api/conversation_role/service_test.go"
const route = "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/ConversationRole"
const roleID = "00000000-0000-4000-8000-000000000001"
const identity = "Global/" + roleID

// Replies are authored vectors constrained by pinned DTOs, not captured tenant
// responses. Expected write bodies below are actual local serializer captures.
const reply = `{"Identity":"Global/00000000-0000-4000-8000-000000000001","Id":"00000000-0000-4000-8000-000000000001","Name":"Controlled role","Permissions":{"Identity":null,"Chat":{"AddParticipants":false},"Calling":{"SendAudio":true}},"IsAssignable":false,"CreatedAt":"2027-01-01T00:00:00","UpdatedAt":"2027-01-02T00:00:00Z","Future":[]}`

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
	c.Family = "conversation_role"
	if c.Scope == "" {
		c.Scope = "service"
	}
	result, e := contracttest.Run(t.Context(), c)
	if e != nil || !result.Passed() {
		t.Fatalf("conversation role vector differed: %v %+v", e, result.Differences())
	}
}
func ptr[T any](v T) *T { return &v }
func captured(t *testing.T, name string) string {
	t.Helper()
	raw, e := os.ReadFile("testdata/pinned-projections.json")
	if e != nil {
		t.Fatal(e)
	}
	var value struct {
		Cases []struct {
			Scenario string
			Accepted bool
			Body     string
		}
	}
	if json.Unmarshal(raw, &value) != nil {
		t.Fatal("invalid pinned request capture")
	}
	for _, v := range value.Cases {
		if v.Scenario == name && v.Accepted {
			return v.Body
		}
	}
	t.Fatal("pinned request case missing")
	return ""
}
func parameters(t *testing.T) roles.CreateParameters {
	t.Helper()
	created, e := roles.ParseDateTime("2027-01-01T00:00:00")
	if e != nil {
		t.Fatal(e)
	}
	updated, e := roles.ParseDateTime("2027-01-02T00:00:00Z")
	if e != nil {
		t.Fatal(e)
	}
	permission, e := roles.NewPermissions(roles.NewChatPermissions(roles.ChatPermissionsValues{AddParticipants: ptr(false), SendMessage: ptr(true)}), roles.NewCallingPermissions(roles.CallingPermissionsValues{AddParticipants: ptr(true), SendAudio: ptr(false)}))
	if e != nil {
		t.Fatal(e)
	}
	return roles.CreateParameters{Permissions: permission, CreatedAt: created, UpdatedAt: updated}
}

func TestUnitConversationRoleAllGetSetsAndRawSnapshot(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "get-selectors", Steps: []contracttest.Exchange{
		wire("GET", "/configuration/"+identity, "Get-CsConversationRole", "", 200, reply),
		wire("GET", "/configuration/"+identity, "Get-CsConversationRole", "", 200, reply),
		wire("GET", "", "Get-CsConversationRole", "", 200, `[`+reply+`]`),
		wire("GET", "", "Get-CsConversationRole", "", 200, `[`+reply+`]`),
		wire("GET", "", "Get-CsConversationRole", "", 200, `[`+reply+`]`),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		service := sdk.TeamsAdminAPI.ConversationRoles
		v, _, e := service.GetByIdentity(ctx, identity)
		if e != nil {
			return contracttest.Observation{Err: e}
		}
		if v == nil || v.ID == nil || *v.ID != roleID || v.Permissions == nil || v.Permissions.Chat == nil || v.Permissions.Chat.AddParticipants == nil || *v.Permissions.Chat.AddParticipants || v.Permissions.Calling == nil || v.Permissions.Calling.SendAudio == nil || !*v.Permissions.Calling.SendAudio || v.CreatedAt.Kind() != roles.DateTimeUnspecified || v.UpdatedAt.Kind() != roles.DateTimeUTC {
			t.Fatal("typed local values or date kinds were lost")
		}
		raw, e := json.Marshal(v)
		var actual, expected any
		if e != nil || json.Unmarshal(raw, &actual) != nil || json.Unmarshal([]byte(reply), &expected) != nil || !reflect.DeepEqual(actual, expected) {
			t.Fatal("full response metadata was lost")
		}
		if _, _, e = service.GetByID(ctx, roleID); e != nil {
			return contracttest.Observation{Err: e}
		}
		values, _, e := service.List(ctx)
		if e != nil {
			return contracttest.Observation{Err: e}
		}
		if len(values) != 1 {
			t.Fatal("collection lost role")
		}
		values, _, e = service.ListFiltered(ctx, "Global/*")
		if e != nil {
			return contracttest.Observation{Err: e}
		}
		if len(values) != 1 {
			t.Fatal("wildcard lost full identity")
		}
		values, _, e = service.ListFiltered(ctx, " ")
		if e == nil && len(values) != 0 {
			t.Fatal("whitespace was trimmed")
		}
		return contracttest.Observation{Err: e}
	}})
}
func TestUnitConversationCreateSetsUseInspectableFrozenObservedProjection(t *testing.T) {
	t.Parallel()
	for _, relative := range []bool{false, true} {
		name := "identity-create"
		if relative {
			name = "relative-id-create"
		}
		t.Run(name, func(t *testing.T) {
			input := parameters(t)
			var request *roles.CreateRequest
			var e error
			if relative {
				request, e = roles.ProjectCreateByID(roleID, input)
			} else {
				request, e = roles.ProjectCreateByIdentity(identity, input)
			}
			if e != nil {
				t.Fatal(e)
			}
			projection := request.Projection()
			if len(projection.OmittedPaths) != 2 || request.ResourceIdentity() != identity || relative != (len(projection.ChangedPaths) == 1) {
				t.Fatal("source loss or selector difference was hidden")
			}
			projection.Body[0] = '['
			projection.OmittedPaths[0] = "changed"
			input.Permissions.Identity = ptr("changed")
			run(t, contracttest.Case{Name: name, Steps: []contracttest.Exchange{wire("POST", "", "New-CsConversationRole", captured(t, name), 201, reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, response, e := sdk.TeamsAdminAPI.ConversationRoles.Create(ctx, *request)
				if e == nil && (v == nil || v.ID == nil || *v.ID != roleID || response == nil) {
					t.Fatal("creation receipt lost authored response")
				}
				return contracttest.Observation{Err: e}
			}})
			if strings.Contains(fmt.Sprintf("%v %#v %v", request, request, request.Projection()), roleID) {
				t.Fatal("request diagnostics exposed identity")
			}
		})
	}
}
func TestUnitConversationSetBothSelectorsPrepareThenApplyFrozenBody(t *testing.T) {
	t.Parallel()
	for _, relative := range []bool{false, true} {
		name := "identity-set"
		if relative {
			name = "relative-id-set"
		}
		t.Run(name, func(t *testing.T) {
			run(t, contracttest.Case{Name: name, Steps: []contracttest.Exchange{wire("GET", "/configuration/"+identity, "Set-CsConversationRole", "", 200, reply), wire("PATCH", "/configuration/"+identity, "Set-CsConversationRole", captured(t, name), 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				service := sdk.TeamsAdminAPI.ConversationRoles
				input := roles.UpdateParameters{Name: policy.Null[string]()}
				var request *roles.UpdateRequest
				var e error
				if relative {
					request, _, e = service.PrepareUpdateByID(ctx, roleID, input)
				} else {
					request, _, e = service.PrepareUpdateByIdentity(ctx, identity, input)
				}
				if e != nil {
					return contracttest.Observation{Err: e}
				}
				if request.Before() == nil || request.ReadResponse() == nil || request.ResourceIdentity() != identity {
					t.Fatal("preparation receipt missing")
				}
				before := request.Before()
				before.Identity = ptr("retarget")
				projection := request.Projection()
				projection.Body[0] = '['
				var result *roles.UpdateResult
				if relative {
					result, _, e = service.UpdateByID(ctx, roleID, *request)
				} else {
					result, _, e = service.UpdateByIdentity(ctx, identity, *request)
				}
				if e == nil && (result == nil || result.Before == nil || result.Before.Identity == nil || *result.Before.Identity != identity || !result.PatchInvoked || !result.HTTPAccepted || result.ReadResponse == nil) {
					t.Fatal("prepared body/stages were changed by exposed snapshots")
				}
				return contracttest.Observation{Err: e}
			}})
		})
	}
}
func TestUnitConversationSelectedPermissionSetReportsLossWithoutRejectingInput(t *testing.T) {
	t.Parallel()
	permission := parameters(t).Permissions
	run(t, contracttest.Case{Name: "permission-set-observed", Steps: []contracttest.Exchange{wire("GET", "/configuration/"+identity, "Set-CsConversationRole", "", 200, reply), wire("PATCH", "/configuration/"+identity, "Set-CsConversationRole", captured(t, "permissions-set"), 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		request, _, e := sdk.TeamsAdminAPI.ConversationRoles.PrepareUpdateByIdentity(ctx, identity, roles.UpdateParameters{Permissions: policy.Set(permission)})
		if e != nil {
			return contracttest.Observation{Err: e}
		}
		if len(request.Projection().OmittedPaths) != 2 {
			t.Fatal("upstream projection loss was hidden")
		}
		result, _, e := sdk.TeamsAdminAPI.ConversationRoles.UpdateByIdentity(ctx, identity, *request)
		if e == nil && (result == nil || !result.PatchInvoked) {
			t.Fatal("accepted local flags were rejected instead of explicitly projected")
		}
		return contracttest.Observation{Err: e}
	}})
}
func TestUnitConversationDeleteBothSetsUseOriginalBoundParameters(t *testing.T) {
	t.Parallel()
	flag := false
	for _, relative := range []bool{false, true} {
		name := "identity-delete"
		body := captured(t, "identity-delete")
		if relative {
			name = "relative-delete"
			body = captured(t, "relative-id-delete")
		}
		t.Run(name, func(t *testing.T) {
			run(t, contracttest.Case{Name: name, Steps: []contracttest.Exchange{wire("DELETE", "/configuration/"+identity, "Remove-CsConversationRole", body, 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				var e error
				if relative {
					_, _, e = sdk.TeamsAdminAPI.ConversationRoles.DeleteByID(ctx, roleID, policy.Controls{Force: &flag})
				} else {
					_, _, e = sdk.TeamsAdminAPI.ConversationRoles.DeleteByIdentity(ctx, identity, policy.Controls{Force: &flag})
				}
				return contracttest.Observation{Err: e}
			}})
		})
	}
}
func TestUnitConversationPreparationDenialAndAcceptedDecodeFailureRetainStages(t *testing.T) {
	t.Parallel()
	for _, denied := range []bool{true, false} {
		name := "denied-preparation"
		want := auth.ErrForbidden
		steps := []contracttest.Exchange{wire("GET", "/configuration/"+identity, "Set-CsConversationRole", "", 403, `{"code":"Forbidden"}`)}
		if !denied {
			name = "accepted-malformed"
			want = client.ErrDecode
			steps = []contracttest.Exchange{wire("GET", "/configuration/"+identity, "Set-CsConversationRole", "", 200, reply), wire("PATCH", "/configuration/"+identity, "Set-CsConversationRole", captured(t, "identity-set"), 200, `[]`)}
		}
		t.Run(name, func(t *testing.T) {
			run(t, contracttest.Case{Name: name, Steps: steps, Want: contracttest.Expected{Error: want}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				request, response, e := sdk.TeamsAdminAPI.ConversationRoles.PrepareUpdateByIdentity(ctx, identity, roles.UpdateParameters{Name: policy.Null[string]()})
				if denied {
					if !errors.Is(e, want) || request == nil || request.ReadResponse() == nil || response == nil {
						t.Fatal("denial preparation evidence lost")
					}
					return contracttest.Observation{Err: e}
				}
				if e != nil {
					return contracttest.Observation{Err: e}
				}
				result, response, e := sdk.TeamsAdminAPI.ConversationRoles.UpdateByIdentity(ctx, identity, *request)
				if !errors.Is(e, want) || result == nil || result.Before == nil || !result.PatchInvoked || !result.HTTPAccepted || result.After != nil || response == nil {
					t.Fatal("accepted malformed reply lost its stages")
				}
				return contracttest.Observation{Err: e}
			}})
		})
	}
}
func TestUnitConversationPreparedSelectorCannotRetargetOrChangeSet(t *testing.T) {
	t.Parallel()
	for _, differentSet := range []bool{false, true} {
		name := "retarget-prepared"
		if differentSet {
			name = "different-set"
		}
		t.Run(name, func(t *testing.T) {
			run(t, contracttest.Case{Name: name, Steps: []contracttest.Exchange{wire("GET", "/configuration/"+identity, "Set-CsConversationRole", "", 200, reply)}, Want: contracttest.Expected{Error: roles.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				request, _, e := sdk.TeamsAdminAPI.ConversationRoles.PrepareUpdateByIdentity(ctx, identity, roles.UpdateParameters{Name: policy.Null[string]()})
				if e != nil {
					return contracttest.Observation{Err: e}
				}
				var result *roles.UpdateResult
				if differentSet {
					result, _, e = sdk.TeamsAdminAPI.ConversationRoles.UpdateByID(ctx, roleID, *request)
				} else {
					result, _, e = sdk.TeamsAdminAPI.ConversationRoles.UpdateByIdentity(ctx, "Global/other", *request)
				}
				if !errors.Is(e, roles.ErrInput) || result == nil || result.Before == nil || result.PatchInvoked {
					t.Fatal("prepared request escaped its exact selector")
				}
				return contracttest.Observation{Err: e}
			}})
		})
	}
}
func TestUnitConversationUnboundSetIsNoOpAndEmptyFilterIsLocalError(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "unbound-set", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		request, _, e := sdk.TeamsAdminAPI.ConversationRoles.PrepareUpdateByID(ctx, roleID, roles.UpdateParameters{})
		if e != nil {
			return contracttest.Observation{Err: e}
		}
		result, _, e := sdk.TeamsAdminAPI.ConversationRoles.UpdateByID(ctx, roleID, *request)
		if e == nil && (result == nil || !result.NoOp || result.PatchInvoked) {
			t.Fatal("unbound set made an implicit request")
		}
		return contracttest.Observation{Err: e}
	}})
	run(t, contracttest.Case{Name: "empty-filter", Scope: "local", Want: contracttest.Expected{Error: roles.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, _, e := sdk.TeamsAdminAPI.ConversationRoles.ListFiltered(ctx, "")
		return contracttest.Observation{Err: e}
	}})
}

func TestUnitConversationRelativeSetKeepsMissingModelIDNull(t *testing.T) {
	t.Parallel()
	missingID := strings.Replace(reply, `"Id":"`+roleID+`",`, "", 1)
	run(t, contracttest.Case{Name: "relative-set-missing-model-id", Steps: []contracttest.Exchange{
		wire("GET", "/configuration/"+identity, "Set-CsConversationRole", "", 200, missingID),
		wire("PATCH", "/configuration/"+identity, "Set-CsConversationRole", captured(t, "relative-id-set-missing-model-id"), 204, ""),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		request, _, e := sdk.TeamsAdminAPI.ConversationRoles.PrepareUpdateByID(ctx, roleID, roles.UpdateParameters{Name: policy.Null[string]()})
		if e != nil {
			return contracttest.Observation{Err: e}
		}
		if request.Before() == nil || request.Before().ID != nil {
			t.Fatal("missing response Id was copied from the selector")
		}
		result, _, e := sdk.TeamsAdminAPI.ConversationRoles.UpdateByID(ctx, roleID, *request)
		if e == nil && (result == nil || !result.HTTPAccepted) {
			t.Fatal("source null model Id projection failed")
		}
		return contracttest.Observation{Err: e}
	}})
}
