package applicationaccesspolicy_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"

	"strconv"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	policy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/application_access_policy"
)

const policyOrigin = "https://api.interfaces.records.teams.microsoft.com"
const policyForest = "https://admin.example.invalid/"

func policyDiscovery(t *testing.T) testutil.Exchange {
	return testutil.Exchange{Request: testutil.WireRequest{Method: "GET", URL: policyOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience,
		Headers: http.Header{"Accept": {"application/json"}}}, Reply: testutil.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}},
		Body: []byte(discoveryResponseFixture(t))}}
}

func policyWire(t *testing.T, name, command string) testutil.Exchange {
	t.Helper()
	item, ok := requestCases(t)[name]
	if !ok {
		t.Fatal("request input missing")
	}
	u, err := url.Parse(item.URL)
	if err != nil {
		t.Fatal(err)
	}
	// Go and .NET use different valid colon spellings. Normalize each segment
	// exactly once; encoded slashes must remain within their original segment.
	parts := strings.Split(u.EscapedPath(), "/")
	for i, part := range parts {
		decoded, err := url.PathUnescape(part)
		if err != nil {
			t.Fatal(err)
		}
		parts[i] = url.PathEscape(decoded)
	}
	address := u.Scheme + "://" + u.Host + strings.Join(parts, "/")
	if u.RawQuery != "" {
		address += "?" + u.RawQuery
	}
	headers := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {policyForest}}
	var body json.RawMessage
	if len(item.Body) != 0 && string(item.Body) != "null" {
		body = item.Body
		headers.Set("Content-Type", "application/json")
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: item.Method, URL: address, Audience: auth.TeamsAdminAudience, Headers: headers, JSON: body},
		Reply: testutil.Reply{Status: item.Status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(item.Response)}}
}

func policyContract(t *testing.T, c testutil.Case) testutil.Result {
	t.Helper()

	result, err := testutil.Run(t.Context(), c)
	if err != nil || !result.Passed() {
		t.Fatalf("policy service contract differs: %+v", result.Differences())
	}
	return result
}

func TestUnitPolicyListsUseAcceptanceGlobalFirstResponse(t *testing.T) {
	t.Parallel()
	for _, filtered := range []bool{false, true} {
		t.Run(strconv.FormatBool(filtered), func(t *testing.T) {
			sdk, calls := scriptedClient(t, []wireCase{requestCases(t)["policy-list"]})
			var values []policy.Policy
			var err error
			if filtered {
				values, _, err = sdk.TeamsAdminAPI.ApplicationAccessPolicies.ListFiltered(t.Context(), "*")
			} else {
				values, _, err = sdk.TeamsAdminAPI.ApplicationAccessPolicies.List(t.Context())
			}
			if err != nil || len(values) != 2 || values[0].Identity != "Global" || !responsePolicyIdentity(values[1].Identity) || *calls != 2 {
				t.Fatal("acceptance list ordering or request count changed")
			}
		})
	}
}

func TestUnitPolicyMutationsDoNotReplayUnderDefaultTransportPolicy(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, fixture, command, set string
		invoke                      func(context.Context, *policy.Service) testutil.Observation
	}{
		{name: "create", fixture: "policy-create", command: "New-CsApplicationAccessPolicy", set: "Identity", invoke: func(ctx context.Context, s *policy.Service) testutil.Observation {
			v, r, e := s.Create(ctx, policy.CreateRequest{Identity: "tag:Fixture / Policy"})
			return testutil.Observation{Value: v, Response: r, Err: e}
		}},
		{name: "update", fixture: "policy-update-put", command: "Set-CsApplicationAccessPolicy", set: "Identity", invoke: func(ctx context.Context, s *policy.Service) testutil.Observation {
			r, e := s.UpdateByIdentity(ctx, "tag:Fixture / Policy", policy.UpdateRequest{})
			return testutil.Observation{Response: r, Err: e}
		}},
		{name: "delete", fixture: "policy-delete", command: "Remove-CsApplicationAccessPolicy", set: "Identity", invoke: func(ctx context.Context, s *policy.Service) testutil.Observation {
			r, e := s.DeleteByIdentity(ctx, "tag:Fixture / Policy")
			return testutil.Observation{Response: r, Err: e}
		}},
		{name: "user", fixture: "user-assignment", command: "Grant-CsApplicationAccessPolicy", set: "Identity", invoke: func(ctx context.Context, s *policy.Service) testutil.Observation {
			r, e := s.AssignToUser(ctx, "fixture/user", "Fixture")
			return testutil.Observation{Response: r, Err: e}
		}},
		{name: "tenant", fixture: "tenant-assignment", command: "Grant-CsApplicationAccessPolicy", set: "GrantToTenant", invoke: func(ctx context.Context, s *policy.Service) testutil.Observation {
			r, e := s.AssignTenantDefault(ctx, "Fixture", false)
			return testutil.Observation{Response: r, Err: e}
		}},
		{name: "group", fixture: "group-assignment-accepted", command: "Grant-CsApplicationAccessPolicy", set: "GrantToGroup", invoke: func(ctx context.Context, s *policy.Service) testutil.Observation {
			rank := 2
			v, r, e := s.AssignToGroup(ctx, "fixture/group", "Fixture", &rank)
			return testutil.Observation{Value: v, Response: r, Err: e}
		}},
	} {
		for _, status := range []int{429, 503} {
			name := tc.name + "-" + strconv.Itoa(status)
			t.Run(name, func(t *testing.T) {
				step := policyWire(t, tc.fixture, tc.command)
				if tc.name == "create" || tc.name == "update" {
					step.Request.JSON = json.RawMessage(`{"Identity":"tag:Fixture / Policy"}`)
				}
				step.Reply.Status, step.Reply.Body = status, nil
				policyContract(t, testutil.Case{Name: name, Steps: []testutil.Exchange{policyDiscovery(t), step},

					Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
						return tc.invoke(ctx, sdk.TeamsAdminAPI.ApplicationAccessPolicies)
					},
					Want: testutil.Expected{Status: status, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}})
			})
		}
	}
}
