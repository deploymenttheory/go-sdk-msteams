package policyassignments_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

const source = "contracts/endpoints/8.1.0/fixtures/policy-assignment-requests.json"
const origin = "https://api.interfaces.records.teams.microsoft.com"
const forest = "https://fixture-forest.example.invalid/"

func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family = "policy_assignments"
	if c.Source == "" {
		c.Source = source
	}
	if len(c.Steps) == 0 {
		c.Scope = "local"
	} else {
		c.Scope = "service"
		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	}
	r, err := contracttest.Run(t.Context(), c)
	if err != nil || !r.Passed() {
		t.Fatalf("assignment contract %s differs: %v %+v", c.Name, err, r.Differences())
	}
	return r
}
func evidence(t *testing.T, results []contracttest.Result) {
	t.Helper()
	revision := os.Getenv("MSTEAMS_CONTRACT_REVISION")
	if revision == "" {
		return
	}
	root, err := os.OpenRoot(t.ArtifactDir())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	f, err := root.Create("policy-assignment-service.json")
	if err != nil {
		t.Fatal(err)
	}
	if err = contracttest.WriteEvidence(f, revision, results); err != nil {
		_ = f.Close()
		t.Fatal(err)
	}
	if err = f.Close(); err != nil {
		t.Fatal(err)
	}
}
func discovery() contracttest.Exchange {
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Body: []byte(`{"endpoints":{"adminServiceEndpoint":"fixture-forest.example.invalid:443/path?q=f#x"}}`)}}
}
func wire(method, url, command string, body json.RawMessage, status int, reply string) contracttest.Exchange {
	h := http.Header{"Accept": {"application/json"}}
	h.Set("X-MS-CmdletName", command)
	h.Set("X-MS-Target-Uri", forest)
	h[http.CanonicalHeaderKey("MPACmdlet")] = nil
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: url, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}

func batchInput(name string) resource.BatchCreateRequest {
	input := resource.BatchCreateRequest{Identities: []string{}}
	if name == "batch-create" {
		input.Identities = []string{"fixture/a", "fixture/a", "fixture/b"}
		input.PolicyType, input.PolicyName = policy.Set("Fixture"), policy.Set("")
		input.OperationName = "Fixture /?&:@"
		input.AdditionalParameters = map[string]json.RawMessage{"flag": []byte("false"), "n": []byte("0"), "empty": []byte(`""`), "null": []byte("null")}
	}
	if name == "batch-create-nested-additional" {
		input.AdditionalParameters = map[string]json.RawMessage{"nested": []byte(`{"flag":false,"null":null}`), "array": []byte(`[0,false,"",null,{"null":null}]`)}
	}
	return input
}

func invoke(name string, ctx context.Context, sdk *msteams.Client) contracttest.Observation {
	s := sdk.TeamsAdminAPI.PolicyAssignments
	var o contracttest.Observation
	groupID, policyType := "fixture/group:@", "Fixture/Policy:@"
	if name == "group-get-null" || name == "group-get-malformed" || name == "group-forbidden" || name == "group-create-malformed" {
		groupID, policyType = "fixture/group", "Fixture"
	}
	switch {
	case name == "batch-list" || name == "batch-list-null":
		o.Value, o.Response, o.Err = s.ListBatches(ctx, "")
	case name == "batch-list-filter":
		o.Value, o.Response, o.Err = s.ListBatches(ctx, "Future /?&:@")
	case name == "batch-get":
		o.Value, o.Response, o.Err = s.GetBatch(ctx, "fixture/operation:@")
	case name == "group-list":
		o.Value, o.Response, o.Err = s.ListGroup(ctx, groupID)
	case name == "group-list-type":
		o.Value, o.Response, o.Err = s.ListGroupsByPolicy(ctx, "Fixture /?&:@")
	case name == "group-list-all":
		o.Value, o.Response, o.Err = s.ListGroupsByPolicy(ctx, "")
	case name == "user-effective":
		o.Value, o.Response, o.Err = s.GetUser(ctx, "fixture/user:@", "Fixture /?&:@")
	case name == "user-effective-all":
		o.Value, o.Response, o.Err = s.GetUser(ctx, "fixture/user:@", "")
	case strings.HasPrefix(name, "group-create"):
		settings := resource.GroupSettings{}
		if name == "group-create" {
			settings.PolicyName, settings.Priority = policy.Set("Fixture"), policy.Set(int32(0))
		}
		if name == "group-create-empty-name-negative-rank" || name == "group-create-malformed" {
			settings.PolicyName, settings.Priority = policy.Set(""), policy.Set(int32(-1))
		}
		o.Value, o.Response, o.Err = s.CreateGroup(ctx, groupID, policyType, settings)
	case strings.HasPrefix(name, "group-update"):
		settings := resource.GroupSettings{}
		if name == "group-update" {
			settings.PolicyName, settings.Priority = policy.Set("UpdatedFixture"), policy.Set(int32(0))
		}
		o.Value, o.Response, o.Err = s.UpdateGroup(ctx, groupID, policyType, settings)
	case strings.HasPrefix(name, "group-delete"):
		o.Value, o.Response, o.Err = s.DeleteGroup(ctx, groupID, policyType)
	case strings.HasPrefix(name, "batch-create"):
		o.Value, o.Response, o.Err = s.CreateBatch(ctx, batchInput(name))
	default:
		o.Value, o.Response, o.Err = s.GetGroup(ctx, groupID, policyType)
	}
	if result, ok := o.Value.(*resource.OperationResult); ok && result != nil && !result.HTTPAccepted {
		o.Err = errors.New("successful submission lost acceptance")
	}
	return o
}

func mapping(name string) []contracttest.Mapping {
	command, set := "Get-CsGroupPolicyAssignment", "Get"
	switch {
	case strings.HasPrefix(name, "group-update"):
		return nil // Additional operation, no exported command claim.
	case name == "group-list":
		set = "Get1"
	case name == "group-list-type" || name == "group-list-all":
		set = "Get2"
	case strings.HasPrefix(name, "user-effective"):
		command, set = "Get-CsUserPolicyAssignment", "Get"
	case strings.HasPrefix(name, "batch-list"):
		command, set = "Get-CsBatchPolicyAssignmentOperation", "Get"
	case name == "batch-get":
		command, set = "Get-CsBatchPolicyAssignmentOperation", "Get1"
	case strings.HasPrefix(name, "batch-create"):
		command, set = "New-CsBatchPolicyAssignmentOperation", "NewExpanded"
	case strings.HasPrefix(name, "group-create"):
		command, set = "New-CsGroupPolicyAssignment", "NewExpanded"
	case strings.HasPrefix(name, "group-delete"):
		command, set = "Remove-CsGroupPolicyAssignment", "Remove"
	}
	return []contracttest.Mapping{{Command: command, ParameterSet: set}}
}

func TestUnitPinnedGeneratedAssignmentRequests(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../" + source)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name    string          `json:"name"`
			Method  string          `json:"method"`
			URL     string          `json:"generated_url"`
			Body    json.RawMessage `json:"body"`
			Reply   string          `json:"response_body"`
			Status  int             `json:"response_status"`
			Failure *string         `json:"decoder_failure"`
			Sends   int             `json:"sends"`
		}
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Cases) != 25 {
		t.Fatal("invalid pinned request observations")
	}
	var results []contracttest.Result
	for _, f := range fixture.Cases {
		if f.Sends != 1 {
			t.Fatal("unsafe pinned request")
		}
		command := "Set-CsGroupPolicyAssignment"
		mappings := mapping(f.Name)
		if len(mappings) != 0 {
			command = mappings[0].Command
		}
		body := f.Body
		if string(body) == "null" {
			body = nil
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: mappings, Steps: []contracttest.Exchange{discovery(), wire(f.Method, f.URL, command, body, f.Status, f.Reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return invoke(f.Name, ctx, sdk)
			}}
			if f.Failure != nil {
				c.Want.Error = client.ErrDecode
			}
			if f.Status == 403 {
				c.Want.APIError = &client.APIError{StatusCode: 403, Backend: client.TeamsAdmin, Codes: []string{"FixtureDenied"}}
			}
			results = append(results, run(t, c))
		}
	}
	evidence(t, results)
}
