package appinstallation_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"net/url"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	ai "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/app_installation"
)

const source = "msteams/graph_api/app_installation/service_test.go"
const catalogID = "00000000-0000-4000-8000-000000000001"
const installedID = "installation-id-distinct-from-catalog-id"
const item = `{"id":"` + installedID + `","teamsAppDefinition":{"id":"definition-id","teamsAppId":"` + catalogID + `","version":"1.0"},"future":false,"nullable":null,"empty":[]}`

func exchange(method, target, body string, status int, reply string) contracttest.Exchange {
	headers := http.Header{"Accept": {"application/json"}}
	if body != "" {
		headers.Set("Content-Type", "application/json")
	}
	response := http.Header{}
	if reply != "" {
		response.Set("Content-Type", "application/json")
	}
	var input json.RawMessage
	if body != "" {
		input = json.RawMessage(body)
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: target, Audience: auth.GraphAudience, Headers: headers, JSON: input}, Reply: contracttest.Reply{Status: status, Headers: response, Body: []byte(reply)}}
}
func run(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Source = source
	c.Family = "app_installation"
	if c.Scope == "" {
		c.Scope = "service"
	}
	result, err := contracttest.Run(t.Context(), c)
	if err != nil || !result.Passed() {
		t.Fatalf("installation contract differed: %v %+v", err, result.Differences())
	}
}
func query(app string) string {
	q := url.Values{"$expand": {"teamsAppDefinition"}}
	if app != "" {
		q.Set("$filter", "teamsAppDefinition/teamsAppId eq '"+app+"'")
	}
	return "?" + q.Encode()
}

func TestUnitTeamAndUserInstallationLifecycle(t *testing.T) {
	t.Parallel()
	for _, scope := range []struct {
		name, path string
		target     ai.Target
	}{
		{"team", "/v1.0/teams/owned-team/installedApps", ai.Target{TeamID: "owned-team"}},
		{"user", "/v1.0/users/owned-user/teamwork/installedApps", ai.Target{UserID: "owned-user"}},
	} {
		t.Run(scope.name, func(t *testing.T) {
			base := "https://graph.microsoft.com" + scope.path
			consent := &ai.PermissionSet{ResourceSpecificPermissions: []ai.ResourcePermission{{PermissionValue: "ChannelMessage.Read.Group", PermissionType: "application"}, {PermissionValue: "ChannelMeeting.ReadBasic.Group", PermissionType: "delegated"}}}
			run(t, contracttest.Case{Name: scope.name + "-lifecycle", Steps: []contracttest.Exchange{
				exchange("POST", base, `{"teamsApp@odata.bind":"https://graph.microsoft.com/v1.0/appCatalogs/teamsApps/`+catalogID+`","consentedPermissionSet":{"resourceSpecificPermissions":[{"permissionValue":"ChannelMessage.Read.Group","permissionType":"application"},{"permissionValue":"ChannelMeeting.ReadBasic.Group","permissionType":"delegated"}]}}`, 201, ""),
				exchange("GET", base+query(""), "", 200, `{"value":[`+item+`]}`),
				exchange("GET", base+"/"+installedID+query(""), "", 200, item),
				exchange("POST", base+"/"+installedID+"/upgrade", `{}`, 204, ""),
				exchange("DELETE", base+"/"+installedID, "", 204, ""),
			}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				service := sdk.GraphAPI.AppInstallations
				if _, err := service.Create(ctx, scope.target, ai.CreateRequest{CatalogID: catalogID, Consent: consent}); err != nil {
					return contracttest.Observation{Err: err}
				}
				listed, err := service.List(ctx, scope.target, ai.ListOptions{})
				if err != nil {
					return contracttest.Observation{Err: err}
				}
				if len(listed.Items) != 1 || listed.Items[0].ID != installedID {
					t.Fatal("installation identity replaced by catalog identity")
				}
				value, _, err := service.GetByID(ctx, scope.target, installedID)
				if err != nil {
					return contracttest.Observation{Err: err}
				}
				if string(value.RawFields["future"]) != "false" || string(value.RawFields["nullable"]) != "null" || string(value.RawFields["empty"]) != "[]" || value.AppDefinition == nil || value.AppDefinition.TeamsAppID == nil || *value.AppDefinition.TeamsAppID != catalogID {
					t.Fatal("expanded installation presence lost")
				}
				if _, err = service.UpgradeByID(ctx, scope.target, value.ID, nil); err != nil {
					return contracttest.Observation{Err: err}
				}
				_, err = service.DeleteByID(ctx, scope.target, value.ID)
				return contracttest.Observation{Err: err}
			}})
		})
	}
}

func TestUnitCatalogResolutionRequiresCompleteUniqueInstallation(t *testing.T) {
	t.Parallel()
	for _, scope := range []struct {
		name, path string
		target     ai.Target
	}{
		{"team", "/v1.0/teams/owned-team/installedApps", ai.Target{TeamID: "owned-team"}},
		{"user", "/v1.0/users/owned-user/teamwork/installedApps", ai.Target{UserID: "owned-user"}},
	} {
		base := "https://graph.microsoft.com" + scope.path
		for _, scenario := range []string{"missing", "duplicate", "mismatch", "partial-denied", "unique-delete", "unique-without-definition", "unique-upgrade"} {
			t.Run(scope.name+"/"+scenario, func(t *testing.T) {
				body := `{"value":[]}`
				var want error
				switch scenario {
				case "missing":
					want = ai.ErrNotFound
				case "duplicate":
					body = `{"value":[` + item + `,` + item + `]}`
					want = ai.ErrAmbiguous
				case "unique-without-definition":
					body = `{"value":[{"id":"` + installedID + `"}]}`
				case "mismatch":
					body = `{"value":[` + strings.Replace(item, catalogID, "other-catalog-id", 1) + `]}`
					want = ai.ErrIdentityMismatch
				case "partial-denied":
					body = `{"value":[` + item + `],"@odata.nextLink":"` + base + `?skiptoken=opaque"}`
					want = auth.ErrForbidden
				default:
					body = `{"value":[` + item + `]}`
				}
				steps := []contracttest.Exchange{exchange("GET", base+query(catalogID), "", 200, body)}
				if scenario == "partial-denied" {
					steps = append(steps, exchange("GET", base+"?skiptoken=opaque", "", 403, `{"error":{"code":"Forbidden"}}`))
				}
				if scenario == "unique-delete" || scenario == "unique-without-definition" {
					steps = append(steps, exchange("DELETE", base+"/"+installedID, "", 204, ""))
				}
				if scenario == "unique-upgrade" {
					steps = append(steps, exchange("POST", base+"/"+installedID+"/upgrade", `{"consentedPermissionSet":{"resourceSpecificPermissions":[]}}`, 204, ""))
				}
				run(t, contracttest.Case{Name: scenario, Steps: steps, Want: contracttest.Expected{Error: want}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					service := sdk.GraphAPI.AppInstallations
					if scenario == "unique-upgrade" {
						empty := &ai.PermissionSet{ResourceSpecificPermissions: []ai.ResourcePermission{}}
						_, err := service.UpgradeByCatalogID(ctx, scope.target, catalogID, empty)
						return contracttest.Observation{Err: err}
					}
					result, err := service.DeleteByCatalogID(ctx, scope.target, catalogID)
					if scenario == "mismatch" && (result == nil || !result.LookupComplete || result.LookupResponse == nil || result.LookupResponse.StatusCode() != 200 || result.InstallationID != "" || result.ActionInvoked) {
						t.Fatal("mismatched expanded catalog identity triggered or obscured a write")
					}
					return contracttest.Observation{Err: err}
				}})
			})
		}
	}
}

func TestUnitInstallationInvalidTargetAndIDsSendNothing(t *testing.T) {
	t.Parallel()
	for _, target := range []ai.Target{{}, {TeamID: "team", UserID: "user"}, {TeamID: "../escape"}} {
		run(t, contracttest.Case{Name: "invalid-target", Scope: "local", Want: contracttest.Expected{Error: ai.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, err := sdk.GraphAPI.AppInstallations.Create(ctx, target, ai.CreateRequest{CatalogID: catalogID})
			return contracttest.Observation{Err: err}
		}})
	}
	run(t, contracttest.Case{Name: "invalid-id", Scope: "local", Want: contracttest.Expected{Error: ai.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, err := sdk.GraphAPI.AppInstallations.UpgradeByID(ctx, ai.Target{UserID: "owned-user"}, "../escape", nil)
		return contracttest.Observation{Err: err}
	}})
}

func TestUnitInstallationMutationDenialDoesNotReplay(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "install-denied", Steps: []contracttest.Exchange{exchange("POST", "https://graph.microsoft.com/v1.0/teams/owned-team/installedApps", `{"teamsApp@odata.bind":"https://graph.microsoft.com/v1.0/appCatalogs/teamsApps/`+catalogID+`"}`, 403, `{"error":{"code":"Forbidden"}}`)}, Want: contracttest.Expected{Error: auth.ErrForbidden}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, err := sdk.GraphAPI.AppInstallations.Create(ctx, ai.Target{TeamID: "owned-team"}, ai.CreateRequest{CatalogID: catalogID})
		return contracttest.Observation{Err: err}
	}})
}

func TestUnitInstallationRejectsMissingIdentity(t *testing.T) {
	var value ai.Installation
	if !errors.Is(json.Unmarshal([]byte(`{"teamsAppDefinition":{}}`), &value), client.ErrDecode) {
		t.Fatal("missing installation ID accepted")
	}
}

func TestUnitInstallationBindingUsesSelectedGraphCloud(t *testing.T) {
	t.Parallel()
	step := exchange("POST", "https://graph.microsoft.us/v1.0/users/owned-user/teamwork/installedApps", `{"teamsApp@odata.bind":"https://graph.microsoft.us/v1.0/appCatalogs/teamsApps/`+catalogID+`"}`, 201, "")
	step.Request.Audience = "https://graph.microsoft.us"
	run(t, contracttest.Case{Name: "selected-cloud-binding", Cloud: "gcch", Steps: []contracttest.Exchange{step}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, err := sdk.GraphAPI.AppInstallations.Create(ctx, ai.Target{UserID: "owned-user"}, ai.CreateRequest{CatalogID: catalogID})
		return contracttest.Observation{Err: err}
	}})
}

func TestUnitNativeConsentValidatesPermissionTypeBeforeIO(t *testing.T) {
	t.Parallel()
	for _, consent := range []*ai.PermissionSet{{}, {ResourceSpecificPermissions: []ai.ResourcePermission{{PermissionValue: "Channel.Read.Group", PermissionType: "invented"}}}} {
		run(t, contracttest.Case{Name: "invalid-native-consent", Scope: "local", Want: contracttest.Expected{Error: ai.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, err := sdk.GraphAPI.AppInstallations.Create(ctx, ai.Target{TeamID: "owned-team"}, ai.CreateRequest{CatalogID: catalogID, Consent: consent})
			return contracttest.Observation{Err: err}
		}})
	}
}

func TestUnitNativeTeamCreateAcceptsDocumentedEmpty200(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "team-empty-200", Steps: []contracttest.Exchange{exchange("POST", "https://graph.microsoft.com/v1.0/teams/owned-team/installedApps", `{"teamsApp@odata.bind":"https://graph.microsoft.com/v1.0/appCatalogs/teamsApps/`+catalogID+`"}`, 200, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, err := sdk.GraphAPI.AppInstallations.Create(ctx, ai.Target{TeamID: "owned-team"}, ai.CreateRequest{CatalogID: catalogID})
		return contracttest.Observation{Err: err}
	}})
}

func TestUnitExpandedDefinitionPreservesTypedIdentitiesAndNestedPresence(t *testing.T) {
	t.Parallel()
	var value ai.Installation
	raw := `{"id":"installed","teamsAppDefinition":{"id":"version-definition","teamsAppId":"catalog","azureADAppId":"aad-application","displayName":null,"version":"","authorization":{"requiredPermissionSet":{"resourceSpecificPermissions":[{"permissionValue":"Channel.Read.Group","permissionType":"delegated"}]}},"future":[]}}`
	if json.Unmarshal([]byte(raw), &value) != nil || value.AppDefinition == nil {
		t.Fatal("typed expanded definition failed to decode")
	}
	d := value.AppDefinition
	if d.ID == nil || *d.ID != "version-definition" || d.TeamsAppID == nil || *d.TeamsAppID != "catalog" || d.AzureADAppID == nil || *d.AzureADAppID != "aad-application" || d.DisplayName != nil || d.Version == nil || *d.Version != "" || string(d.RawFields["displayName"]) != "null" || string(d.RawFields["future"]) != "[]" || d.Authorization == nil || d.Authorization.RequiredPermissionSet == nil || len(d.Authorization.RequiredPermissionSet.ResourceSpecificPermissions) != 1 || d.Authorization.RequiredPermissionSet.ResourceSpecificPermissions[0].PermissionType != ai.PermissionDelegated {
		t.Fatal("expanded definition collapsed identities, known fields or raw presence")
	}
}
