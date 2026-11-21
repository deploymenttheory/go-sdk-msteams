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
	"github.com/go-resty/resty/v2"
)

const policyOrigin = "https://api.interfaces.records.teams.microsoft.com"
const policyForest = "https://forest.fixture.invalid/"

func policyDiscovery() contracttest.Exchange {
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: policyOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience,
		Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}},
		Body: []byte(`{"Endpoints":{"AdminServiceEndpoint":"forest.fixture.invalid:444/ignored?x=y#ignored"}}`)}}
}

func policyWire(t *testing.T, name, command string) contracttest.Exchange {
	t.Helper()
	item, ok := pinnedCases(t)[name]
	if !ok {
		t.Fatal("pinned operation fixture missing")
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
		c.Source = "contracts/endpoints/8.1.0/fixtures/generated-request-dispatch.json"
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

func policyEvidence(t *testing.T, name string, results []contracttest.Result) {
	t.Helper()
	revision := os.Getenv("MSTEAMS_CONTRACT_REVISION")
	if revision == "" {
		return
	}
	root, err := os.OpenRoot(t.ArtifactDir())
	if err != nil {
		t.Fatal("policy evidence directory unavailable")
	}
	defer func() { _ = root.Close() }()
	file, err := root.Create(name)
	if err != nil {
		t.Fatal("policy evidence file unavailable")
	}
	writeErr := contracttest.WriteEvidence(file, revision, results)
	closeErr := file.Close()
	if writeErr != nil || closeErr != nil {
		t.Fatal("policy evidence write failed")
	}
}

func TestPolicyServiceContractsMapEveryPinnedSet(t *testing.T) {
	t.Parallel()
	seen := map[string]bool{}
	var results []contracttest.Result
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
		{name: "create-readback", command: "New-CsApplicationAccessPolicy", set: "Identity", fixtures: []string{"policy-create", "policy-read-escaped"}, invoke: func(ctx context.Context, s *policy.Service) contracttest.Observation {
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
			steps := []contracttest.Exchange{policyDiscovery()}
			for i, name := range tc.fixtures {
				command := tc.command
				if i > 0 && tc.name == "create-readback" {
					command = "Get-CsApplicationAccessPolicy"
				}
				steps = append(steps, policyWire(t, name, command))
			}
			want := contracttest.Expected{Status: steps[len(steps)-1].Reply.Status}
			switch tc.name {
			case "get", "create-readback":
				want.JSON = json.RawMessage(`{"Identity":"tag:Fixture / Policy","AppIds":[]}`)
			case "list", "filter":
				want.JSON = json.RawMessage(`[]`)
			case "assign-group-accepted":
				want.JSON = json.RawMessage(`{"OperationId":"fixture-operation"}`)
			case "assign-group-no-content":
				want.JSON = json.RawMessage(`{"OperationId":""}`)
			}
			results = append(results, policyContract(t, contracttest.Case{Name: tc.name, Steps: steps, Mappings: []contracttest.Mapping{{Command: tc.command, ParameterSet: tc.set}},
				Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					return tc.invoke(ctx, sdk.TeamsAdminAPI.ApplicationAccessPolicies)
				}, Want: want}))
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
	policyEvidence(t, "application-access-policy-service.json", results)
}

func TestPolicyListsFollowPinnedGlobalFirstComposition(t *testing.T) {
	t.Parallel()
	data, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/policy-wrapper-compositions.json")
	if err != nil {
		t.Fatal(err)
	}
	var captured struct {
		Cases []struct {
			Name string
			IDs  []string `json:"output_identities"`
		}
	}
	if json.Unmarshal(data, &captured) != nil {
		t.Fatal("isolated wrapper observations unavailable")
	}
	var results []contracttest.Result
	for _, name := range []string{"list-global-first", "filter-global-first", "filter-tag-prefix"} {
		t.Run(name, func(t *testing.T) {
			step := policyWire(t, "policy-list", "Get-CsApplicationAccessPolicy")
			step.Reply.Body = []byte(`[{"Identity":"Other"},{"Identity":"gLoBaL"},{"Identity":"Global"},{"Identity":"Tag:Global"},{"Identity":"Fixture"}]`)
			var want []string
			index := -1
			for i, c := range captured.Cases {
				if c.Name == name {
					want, index = c.IDs, i
					break
				}
			}
			if index < 0 {
				t.Fatal("pinned list composition missing")
			}
			encoded, err := json.Marshal(want)
			if err != nil {
				t.Fatal(err)
			}
			set := "Filter"
			if name == "list-global-first" {
				set = "Identity"
			}
			results = append(results, policyContract(t, contracttest.Case{Name: name, Mappings: []contracttest.Mapping{{Command: "Get-CsApplicationAccessPolicy", ParameterSet: set}}, Source: "contracts/endpoints/8.1.0/fixtures/policy-wrapper-compositions.json#/cases/" + strconv.Itoa(index),
				Steps: []contracttest.Exchange{policyDiscovery(), step}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					s := sdk.TeamsAdminAPI.ApplicationAccessPolicies
					var values []policy.Policy
					var response *resty.Response
					var err error
					if name == "list-global-first" {
						values, response, err = s.List(ctx)
					} else {
						pattern := "*"
						if name == "filter-tag-prefix" {
							pattern = "Global"
						}
						values, response, err = s.ListFiltered(ctx, pattern)
					}
					ids := make([]string, 0, len(values))
					for _, v := range values {
						ids = append(ids, v.Identity)
					}
					return contracttest.Observation{Value: ids, Response: response, Err: err}
				}, Want: contracttest.Expected{Status: 200, JSON: encoded}}))
		})
	}
	policyEvidence(t, "application-access-policy-list-service.json", results)
}

func TestPolicyMutationsDoNotReplayUnderDefaultTransportPolicy(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
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
				results = append(results, policyContract(t, contracttest.Case{Name: name, Steps: []contracttest.Exchange{policyDiscovery(), step},
					Mappings: []contracttest.Mapping{{Command: tc.command, ParameterSet: tc.set}},
					Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
						return tc.invoke(ctx, sdk.TeamsAdminAPI.ApplicationAccessPolicies)
					},
					Want: contracttest.Expected{Status: status, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}}))
			})
		}
	}
	policyEvidence(t, "application-access-policy-failure-service.json", results)
}
