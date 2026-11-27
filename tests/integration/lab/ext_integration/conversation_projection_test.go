package extintegration

import (
	"context"
	"errors"
	"fmt"
	"io"
	"net/http"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	roles "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/conversation_role"
	settings "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/conversation_roles_setting"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

const conversationID = "00000000-0000-4000-8000-000000000001"
const conversationIdentity = "Global/" + conversationID
const conversationBefore = `{"Identity":"Global/00000000-0000-4000-8000-000000000001","Id":"00000000-0000-4000-8000-000000000001","Name":"fixture-secret-name","Permissions":{}}`
const conversationSettingsBefore = `{"Identity":"Global","ConversationRoles":[]}`

func conversationClient(t *testing.T, tenant, identity string, transport http.RoundTripper) *msteams.Client {
	t.Helper()
	sdk, e := msteams.NewClient(&auth.Config{Method: auth.SuppliedTokens, TenantID: tenant, IdentityID: identity, TokenMode: auth.Application, AccessTokens: map[string]auth.Token{auth.TeamsAdminAudience: {AccessToken: "synthetic-offline-token", ExpiresAt: time.Now().Add(time.Hour)}}}, msteams.WithHTTPClient(&http.Client{Transport: transport}))
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if e := sdk.Close(); e != nil {
			t.Error(e)
		}
	})
	return sdk
}
func conversationResponse(request *http.Request, body string) *http.Response {
	return &http.Response{StatusCode: 200, Header: http.Header{"Content-Type": {"application/json"}}, Body: io.NopCloser(strings.NewReader(body)), Request: request}
}

// This is offline only. Neither new conditional resource is registered for live
// mutation, and authored replies do not prove effective permission persistence.
func TestUnitConversationPreparedWritesBindClientIdentityAndFrozenSnapshot(t *testing.T) {
	t.Parallel()
	for _, family := range []string{"role", "settings"} {
		t.Run(family, func(t *testing.T) {
			reads, writes, otherCalls := 0, 0, 0
			transport := cancelTransport(func(request *http.Request) (*http.Response, error) {
				if request.Method == http.MethodGet {
					reads++
					body := conversationBefore
					if family == "settings" {
						body = conversationSettingsBefore
					}
					return conversationResponse(request, body), nil
				}
				writes++
				return nil, errors.New("unexpected offline mutation")
			})
			first := conversationClient(t, "fixture-tenant-one", "fixture-principal-one", transport)
			other := conversationClient(t, "fixture-tenant-two", "fixture-principal-two", cancelTransport(func(*http.Request) (*http.Response, error) {
				otherCalls++
				return nil, errors.New("unexpected cross-client request")
			}))
			if family == "role" {
				request, _, e := first.TeamsAdminAPI.ConversationRoles.PrepareUpdateByIdentity(t.Context(), conversationIdentity, roles.UpdateParameters{Name: policy.Null[string]()})
				if e != nil {
					t.Fatal(e)
				}
				result, _, e := other.TeamsAdminAPI.ConversationRoles.UpdateByIdentity(t.Context(), conversationIdentity, *request)
				if !errors.Is(e, roles.ErrInput) || result == nil || result.Before == nil || result.PatchInvoked {
					t.Fatal("role snapshot transferred between principals")
				}
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				result, _, e = first.TeamsAdminAPI.ConversationRoles.UpdateByIdentity(ctx, conversationIdentity, *request)
				if !errors.Is(e, context.Canceled) || result == nil || result.Before == nil || result.PatchInvoked || result.ReadResponse == nil {
					t.Fatal("role cancellation lost preparation")
				}
				if strings.Contains(fmt.Sprintf("%v %#v %v %v", request, request, request.Projection(), request.Before()), "fixture-secret-name") {
					t.Fatal("role diagnostic exposed prepared data")
				}
			} else {
				request, _, e := first.TeamsAdminAPI.ConversationRoleSettings.PrepareUpdateByIdentity(t.Context(), "Global", settings.UpdateParameters{ConversationRoles: policy.Null[*settings.ListModifier]()})
				if e != nil {
					t.Fatal(e)
				}
				result, _, e := other.TeamsAdminAPI.ConversationRoleSettings.UpdateByIdentity(t.Context(), "Global", *request)
				if !errors.Is(e, settings.ErrInput) || result == nil || result.Before == nil || result.PatchInvoked {
					t.Fatal("settings snapshot transferred between principals")
				}
				ctx, cancel := context.WithCancel(t.Context())
				cancel()
				result, _, e = first.TeamsAdminAPI.ConversationRoleSettings.UpdateByIdentity(ctx, "Global", *request)
				if !errors.Is(e, context.Canceled) || result == nil || result.Before == nil || result.PatchInvoked || result.ReadResponse == nil {
					t.Fatal("settings cancellation lost preparation")
				}
			}
			if reads != 1 || writes != 0 || otherCalls != 0 {
				t.Fatal("preparation/rejection invoked an unexpected request")
			}
		})
	}
}
func TestUnitConversationApplyLostResponseAndCancellationRetainUncertainWrite(t *testing.T) {
	t.Parallel()
	for _, family := range []string{"role", "settings"} {
		for _, failure := range []string{"lost-response", "cancel"} {
			t.Run(family+"-"+failure, func(t *testing.T) {
				ctx, cancel := context.WithCancel(t.Context())
				defer cancel()
				reads, writes := 0, 0
				sdk := conversationClient(t, "fixture-tenant", "fixture-principal", cancelTransport(func(request *http.Request) (*http.Response, error) {
					if request.Method == http.MethodGet {
						reads++
						body := conversationBefore
						if family == "settings" {
							body = conversationSettingsBefore
						}
						return conversationResponse(request, body), nil
					}
					if request.Method == http.MethodPatch {
						writes++
						if failure == "cancel" {
							cancel()
							return nil, context.Canceled
						}
						return nil, io.ErrUnexpectedEOF
					}
					t.Fatal("unexpected method")
					return nil, io.ErrUnexpectedEOF
				}))
				var e error
				before, prepared, invoked, accepted := false, false, false, false
				if family == "role" {
					request, _, prepareErr := sdk.TeamsAdminAPI.ConversationRoles.PrepareUpdateByIdentity(ctx, conversationIdentity, roles.UpdateParameters{Name: policy.Null[string]()})
					if prepareErr != nil {
						t.Fatal(prepareErr)
					}
					result, _, operationErr := sdk.TeamsAdminAPI.ConversationRoles.UpdateByIdentity(ctx, conversationIdentity, *request)
					e = operationErr
					if result != nil {
						before = result.Before != nil
						prepared = result.ReadResponse != nil
						invoked = result.PatchInvoked
						accepted = result.HTTPAccepted
					}
				} else {
					request, _, prepareErr := sdk.TeamsAdminAPI.ConversationRoleSettings.PrepareUpdateByIdentity(ctx, "Global", settings.UpdateParameters{ConversationRoles: policy.Null[*settings.ListModifier]()})
					if prepareErr != nil {
						t.Fatal(prepareErr)
					}
					result, _, operationErr := sdk.TeamsAdminAPI.ConversationRoleSettings.UpdateByIdentity(ctx, "Global", *request)
					e = operationErr
					if result != nil {
						before = result.Before != nil
						prepared = result.ReadResponse != nil
						invoked = result.PatchInvoked
						accepted = result.HTTPAccepted
					}
				}
				if e == nil || !before || !prepared || !invoked || accepted || reads != 1 || writes != 1 {
					t.Fatal("uncertain write lost stages or was replayed")
				}
				if failure == "cancel" && !errors.Is(e, context.Canceled) {
					t.Fatal("context cancellation was hidden")
				}
			})
		}
	}
}

func TestUnitConversationRoleReceiptRecoversIndependentlyAfterSettingsDenial(t *testing.T) {
	t.Parallel()
	created, e := roles.ParseDateTime("2027-01-01T00:00:00")
	if e != nil {
		t.Fatal(e)
	}
	updated, e := roles.ParseDateTime("2027-01-02T00:00:00Z")
	if e != nil {
		t.Fatal(e)
	}
	parameters := roles.CreateParameters{Permissions: &roles.Permissions{}, CreatedAt: created, UpdatedAt: updated}
	request, e := roles.ProjectCreateByIdentity(conversationIdentity, parameters)
	if e != nil {
		t.Fatal(e)
	}
	body := string(request.Projection().Body)
	result, e := contracttest.Run(t.Context(), contracttest.Case{Name: "role-receipt-independent-recovery", Family: "ext_integration", Scope: "service", Source: "tests/integration/lab/ext_integration/conversation_projection_test.go", Steps: []contracttest.Exchange{
		exchange(true, "POST", "/Skype.Policy/configurations/ConversationRole", "New-CsConversationRole", body, 201, conversationBefore),
		exchange(true, "GET", "/Skype.Policy/configurations/ConversationRolesSetting/configuration/Global", "Set-CsConversationRolesSetting", "", 403, `{"code":"Forbidden"}`),
		exchange(true, "DELETE", "/Skype.Policy/configurations/ConversationRole/configuration/Global/"+conversationID, "Remove-CsConversationRole", `{"Id":"`+conversationID+`"}`, 204, ""),
	}, Want: contracttest.Expected{Error: auth.ErrForbidden}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		owned, _, e := sdk.TeamsAdminAPI.ConversationRoles.Create(ctx, *request)
		if e != nil {
			return contracttest.Observation{Err: e}
		}
		if owned == nil || owned.ID == nil {
			t.Fatal("owned role receipt missing")
		}
		prepared, _, denied := sdk.TeamsAdminAPI.ConversationRoleSettings.PrepareUpdateByIdentity(ctx, "Global", settings.UpdateParameters{ConversationRoles: policy.Null[*settings.ListModifier]()})
		if !errors.Is(denied, auth.ErrForbidden) || prepared == nil || prepared.ReadResponse() == nil {
			t.Fatal("settings denial did not retain its receipt")
		}
		if _, _, e = sdk.TeamsAdminAPI.ConversationRoles.DeleteByID(ctx, *owned.ID, policy.Controls{}); e != nil {
			t.Fatal("independent owned receipt recovery failed")
		}
		return contracttest.Observation{Err: denied}
	}})
	if e != nil || !result.Passed() {
		t.Fatalf("independent recovery vector differed: %v %+v", e, result.Differences())
	}
}
