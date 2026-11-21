package applicationaccesspolicy_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"reflect"
	"strconv"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	policy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/application_access_policy"
)

const policyOrigin = "https://api.interfaces.records.teams.microsoft.com"
const policyForest = "https://admin.example.invalid/"

func policyDiscovery(t *testing.T) contracttest.Exchange {
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: policyOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience,
		Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}},
		Body: []byte(discoveryResponseFixture(t))}}
}

func policyWire(t *testing.T, name, command string) contracttest.Exchange {
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
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: item.Method, URL: address, Audience: auth.TeamsAdminAudience, Headers: headers, JSON: body},
		Reply: contracttest.Reply{Status: item.Status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(item.Response)}}
}

func policyContract(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family = "application_access_policy"
	if c.Source == "" {
		c.Source = "msteams/teams_admin_api/application_access_policy/testdata/list.json"
	}
	if c.Scope == "" {
		c.Scope = "service"
	}
	result, err := contracttest.Run(t.Context(), c)
	if err != nil || !result.Passed() {
		t.Fatalf("policy service contract differs: %+v", result.Differences())
	}
	return result
}

func TestUnitPolicyRequestsMapEveryPinnedSet(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}

	for _, tc := range []struct {
		name, command, set string
		fixtures           []string
		invoke             func(context.Context, *policy.Service) contracttest.Observation
	}{
		{name: "get", command: "Get-CsApplicationAccessPolicy", set: "Identity", fixtures: []string{"policy-read-escaped"}, invoke: func(ctx context.Context, s *policy.Service) contracttest.Observation {
			v, r, e := s.GetByIdentity(ctx, "tag:Fixture / Policy")
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}},
		{name: "list", command: "Get-CsApplicationAccessPolicy", set: "Identity", fixtures: []string{"policy-list"}, invoke: func(ctx context.Context, s *policy.Service) contracttest.Observation {
			v, r, e := s.List(ctx)
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}},
		{name: "filter", command: "Get-CsApplicationAccessPolicy", set: "Filter", fixtures: []string{"policy-list"}, invoke: func(ctx context.Context, s *policy.Service) contracttest.Observation {
			v, r, e := s.ListFiltered(ctx, "Fixture*")
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}},
		{name: "create-request", command: "New-CsApplicationAccessPolicy", set: "Identity", fixtures: []string{"policy-create", "policy-read-escaped"}, invoke: func(ctx context.Context, s *policy.Service) contracttest.Observation {
			empty := ""
			v, r, e := s.Create(ctx, policy.CreateRequest{Identity: "tag:Fixture / Policy", AppIDs: policy.ReplaceAppIDs([]string{"fixture-app"}), Description: &empty})
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}},
		{name: "update", command: "Set-CsApplicationAccessPolicy", set: "Identity", fixtures: []string{"policy-update-put"}, invoke: func(ctx context.Context, s *policy.Service) contracttest.Observation {
			empty := ""
			r, e := s.UpdateByIdentity(ctx, "tag:Fixture / Policy", policy.UpdateRequest{AppIDs: policy.ReplaceAppIDs([]string{"fixture-app"}), Description: &empty})
			return contracttest.Observation{Response: r, Err: e}
		}},
		{name: "delete", command: "Remove-CsApplicationAccessPolicy", set: "Identity", fixtures: []string{"policy-delete"}, invoke: func(ctx context.Context, s *policy.Service) contracttest.Observation {
			r, e := s.DeleteByIdentity(ctx, "tag:Fixture / Policy")
			return contracttest.Observation{Response: r, Err: e}
		}},
		{name: "assign-user", command: "Grant-CsApplicationAccessPolicy", set: "Identity", fixtures: []string{"user-assignment"}, invoke: func(ctx context.Context, s *policy.Service) contracttest.Observation {
			r, e := s.AssignToUser(ctx, "fixture/user", "Fixture")
			return contracttest.Observation{Response: r, Err: e}
		}},
		{name: "clear-user", command: "Grant-CsApplicationAccessPolicy", set: "Identity", fixtures: []string{"user-clear-empty-policy"}, invoke: func(ctx context.Context, s *policy.Service) contracttest.Observation {
			r, e := s.AssignToUser(ctx, "fixture/user", "")
			return contracttest.Observation{Response: r, Err: e}
		}},
		{name: "assign-tenant", command: "Grant-CsApplicationAccessPolicy", set: "GrantToTenant", fixtures: []string{"tenant-assignment"}, invoke: func(ctx context.Context, s *policy.Service) contracttest.Observation {
			r, e := s.AssignTenantDefault(ctx, "Fixture", false)
			return contracttest.Observation{Response: r, Err: e}
		}},
		{name: "assign-group-accepted", command: "Grant-CsApplicationAccessPolicy", set: "GrantToGroup", fixtures: []string{"group-assignment-accepted"}, invoke: func(ctx context.Context, s *policy.Service) contracttest.Observation {
			rank := 2
			v, r, e := s.AssignToGroup(ctx, "fixture/group", "Fixture", &rank)
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}},
		{name: "assign-group-no-content", command: "Grant-CsApplicationAccessPolicy", set: "GrantToGroup", fixtures: []string{"group-assignment-no-content"}, invoke: func(ctx context.Context, s *policy.Service) contracttest.Observation {
			rank := 2
			v, r, e := s.AssignToGroup(ctx, "fixture/group", "Fixture", &rank)
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}},
	} {
		seen[tc.command+"/"+tc.set] = true
		t.Run(tc.name, func(t *testing.T) {
			step := policyWire(t, tc.fixtures[0], tc.command)
			// This table checks every public parameter-set request. Successful CRUD
			// decoding is tested separately using actual acceptance responses.
			step.Reply.Status, step.Reply.Body = 400, nil
			steps := []contracttest.Exchange{policyDiscovery(t), step}
			want := contracttest.Expected{Status: 400, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			policyContract(t, contracttest.Case{Name: tc.name, Steps: steps, Mappings: []contracttest.Mapping{{Command: tc.command, ParameterSet: tc.set}},
				Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					return tc.invoke(ctx, sdk.TeamsAdminAPI.ApplicationAccessPolicies)
				}, Want: want})
		})
	}
	var ledger struct {
		Records []struct {
			Name  string
			Owner string                  `json:"implementation_owner"`
			Sets  []struct{ Name string } `json:"parameter_sets"`
		}
	}
	data, err := os.ReadFile("../../../contracts/coverage/8.1.0.json")
	if err != nil || json.Unmarshal(data, &ledger) != nil {
		t.Fatal("pinned policy ownership unavailable")
	}
	want := map[string]bool{}
	for _, record := range ledger.Records {
		if record.Owner == "POL-20" {
			for _, set := range record.Sets {
				want[record.Name+"/"+set.Name] = true
			}
		}
	}
	if len(want) != 8 || !reflect.DeepEqual(seen, want) {
		t.Fatal("policy service cases do not map every assigned pinned set exactly")
	}
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
		invoke                      func(context.Context, *policy.Service) contracttest.Observation
	}{
		{name: "create", fixture: "policy-create", command: "New-CsApplicationAccessPolicy", set: "Identity", invoke: func(ctx context.Context, s *policy.Service) contracttest.Observation {
			v, r, e := s.Create(ctx, policy.CreateRequest{Identity: "tag:Fixture / Policy"})
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}},
		{name: "update", fixture: "policy-update-put", command: "Set-CsApplicationAccessPolicy", set: "Identity", invoke: func(ctx context.Context, s *policy.Service) contracttest.Observation {
			r, e := s.UpdateByIdentity(ctx, "tag:Fixture / Policy", policy.UpdateRequest{})
			return contracttest.Observation{Response: r, Err: e}
		}},
		{name: "delete", fixture: "policy-delete", command: "Remove-CsApplicationAccessPolicy", set: "Identity", invoke: func(ctx context.Context, s *policy.Service) contracttest.Observation {
			r, e := s.DeleteByIdentity(ctx, "tag:Fixture / Policy")
			return contracttest.Observation{Response: r, Err: e}
		}},
		{name: "user", fixture: "user-assignment", command: "Grant-CsApplicationAccessPolicy", set: "Identity", invoke: func(ctx context.Context, s *policy.Service) contracttest.Observation {
			r, e := s.AssignToUser(ctx, "fixture/user", "Fixture")
			return contracttest.Observation{Response: r, Err: e}
		}},
		{name: "tenant", fixture: "tenant-assignment", command: "Grant-CsApplicationAccessPolicy", set: "GrantToTenant", invoke: func(ctx context.Context, s *policy.Service) contracttest.Observation {
			r, e := s.AssignTenantDefault(ctx, "Fixture", false)
			return contracttest.Observation{Response: r, Err: e}
		}},
		{name: "group", fixture: "group-assignment-accepted", command: "Grant-CsApplicationAccessPolicy", set: "GrantToGroup", invoke: func(ctx context.Context, s *policy.Service) contracttest.Observation {
			rank := 2
			v, r, e := s.AssignToGroup(ctx, "fixture/group", "Fixture", &rank)
			return contracttest.Observation{Value: v, Response: r, Err: e}
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
				policyContract(t, contracttest.Case{Name: name, Steps: []contracttest.Exchange{policyDiscovery(t), step},
					Mappings: []contracttest.Mapping{{Command: tc.command, ParameterSet: tc.set}},
					Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
						return tc.invoke(ctx, sdk.TeamsAdminAPI.ApplicationAccessPolicies)
					},
					Want: contracttest.Expected{Status: status, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}})
			})
		}
	}
}
