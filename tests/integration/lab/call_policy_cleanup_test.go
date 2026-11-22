package lab_test

import (
	"context"
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

func TestUnitCallPolicyCleanupRequiresIndependentAbsence(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		kind, resource, service string
		modern                  bool
		handler                 func(*msteams.Client) lab.Handler
	}{
		{"events-policy", "TeamsEventsPolicy", "teams_events_policy", false, func(sdk *msteams.Client) lab.Handler {
			return eventsPolicyCleanup{sdk.TeamsAdminAPI.TeamsEventsPolicies}
		}},
		{"enhanced-encryption-policy", "TeamsEnhancedEncryptionPolicy", "teams_enhanced_encryption_policy", true, func(sdk *msteams.Client) lab.Handler {
			return enhancedEncryptionPolicyCleanup{sdk.TeamsAdminAPI.TeamsEnhancedEncryptionPolicies}
		}},
		{"channels-policy", "TeamsChannelsPolicy", "teams_channels_policy", false, func(sdk *msteams.Client) lab.Handler {
			return channelsPolicyCleanup{sdk.TeamsAdminAPI.TeamsChannelsPolicies}
		}},
		{"cortana-policy", "TeamsCortanaPolicy", "teams_cortana_policy", true, func(sdk *msteams.Client) lab.Handler {
			return cortanaPolicyCleanup{sdk.TeamsAdminAPI.TeamsCortanaPolicies}
		}},
		{"call-hold-policy", "TeamsCallHoldPolicy", "teams_call_hold_policy", false, func(sdk *msteams.Client) lab.Handler {
			return callHoldPolicyCleanup{sdk.TeamsAdminAPI.TeamsCallHoldPolicies}
		}},
		{"call-park-policy", "TeamsCallParkPolicy", "teams_call_park_policy", true, func(sdk *msteams.Client) lab.Handler {
			return callParkPolicyCleanup{sdk.TeamsAdminAPI.TeamsCallParkPolicies}
		}},
	} {
		t.Run(tc.kind, func(t *testing.T) {
			name := "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-" + tc.kind
			base := "/Skype.Policy/configurations/" + tc.resource
			list, err := os.ReadFile("../../../msteams/teams_admin_api/" + tc.service + "/testdata/list.json")
			if err != nil {
				t.Fatal(err)
			}
			missing, err := os.ReadFile("../../../msteams/teams_admin_api/" + tc.service + "/testdata/get_deleted_404.json")
			if err != nil {
				t.Fatal(err)
			}
			wire := func(method, path, command string, status int, body []byte) contracttest.Exchange {
				headers := http.Header{"X-MS-CmdletName": {command}}
				var input json.RawMessage
				if tc.modern {
					headers.Set("Accept", "application/json")
					headers.Set("X-MS-Target-Uri", "https://admin.example.invalid/")
					headers[http.CanonicalHeaderKey("MPACmdlet")] = nil
				} else {
					headers.Set("MPACmdlet", "true")
					headers[http.CanonicalHeaderKey("X-MS-Target-Uri")] = nil
					if method == "DELETE" {
						headers.Set("Content-Type", "application/json; charset=utf-8")
						input = json.RawMessage(`{"Identity":"` + name + `"}`)
					}
				}
				return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: "https://api.interfaces.records.teams.microsoft.com" + path, Audience: auth.TeamsAdminAudience, Headers: headers, JSON: input}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body}}
			}
			for _, scenario := range []string{"resolve-absence", "resolve-denied", "delete-denied"} {
				t.Run(scenario, func(t *testing.T) {
					steps := []contracttest.Exchange{}
					if tc.modern {
						discovery, err := os.ReadFile("../../../msteams/teams_admin_api/internal/configapi/testdata/service_discovery.json")
						if err != nil {
							t.Fatal(err)
						}
						steps = append(steps, contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: discovery}})
					}
					entry := lab.Entry{Fixture: name}
					if scenario == "delete-denied" {
						entry.ResourceID = name
						steps = append(steps, wire("DELETE", base+"/configuration/"+name, "Remove-Cs"+tc.resource, 404, nil))
					} else {
						steps = append(steps, wire("GET", base, "Get-Cs"+tc.resource, 200, list))
					}
					status, body := 404, missing
					var want error
					if scenario != "resolve-absence" {
						// Inject an authorization failure after the first observation. A list
						// miss or a DELETE404 must never hide this independent read failure.
						status, body, want = 403, nil, auth.ErrForbidden
					}
					steps = append(steps, wire("GET", base+"/configuration/"+name, "Get-Cs"+tc.resource, status, body))
					result, err := contracttest.Run(t.Context(), contracttest.Case{Name: scenario, Family: tc.kind, Scope: "service", Source: "tests/integration/lab/call_policy_cleanup_test.go", Steps: steps, Want: contracttest.Expected{Error: want}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
						handler := tc.handler(sdk)
						if scenario == "delete-denied" {
							return contracttest.Observation{Err: handler.Delete(ctx, entry)}
						}
						id, err := handler.Resolve(ctx, entry)
						if id != "" {
							t.Fatal("no matching owned policy may resolve to a resource ID")
						}
						return contracttest.Observation{Err: err}
					}})
					if err != nil || !result.Passed() {
						t.Fatalf("cleanup accepted an unverified observation: %v %+v", err, result.Differences())
					}
				})
			}
		})
	}
}
