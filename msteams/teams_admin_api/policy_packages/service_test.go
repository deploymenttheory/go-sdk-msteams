package policypackages_test

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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_packages"
)

const source = "msteams/teams_admin_api/policy_packages/service_test.go"
const origin = "https://api.interfaces.records.teams.microsoft.com"
const forest = "https://admin.example.invalid/"

type requestCase struct {
	Name, Method, URL string
	Body              json.RawMessage
	Headers           http.Header
	Status            int
	Reply             string
}

func requestCases() []requestCase {
	return []requestCase{
		{Name: "list", Method: "GET", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.VerticalPackaging/package", Body: nil, Headers: http.Header{}, Status: 403},
		{Name: "get", Method: "GET", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.VerticalPackaging/package/Fixture%2FPackage%3A%40", Body: nil, Headers: http.Header{}, Status: 403},
		{Name: "user-get", Method: "GET", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.VerticalPackaging/package/assigned", Body: nil, Headers: http.Header{"user": {"fixture-user@example.invalid"}}, Status: 403},
		{Name: "user-recommendations", Method: "GET", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.VerticalPackaging/package", Body: nil, Headers: http.Header{"user": {"fixture-user@example.invalid"}}, Status: 403},
		{Name: "create", Method: "POST", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.VerticalPackaging/package/custompackage", Body: json.RawMessage("{\"Description\":\"\",\"PackageName\":\"Fixture/Package:@\",\"PolicyList\":[{\"PolicyName\":\"FixtureMeeting\",\"PolicyType\":\"TeamsMeetingPolicy\"}]}"), Headers: http.Header{}, Status: 403},
		{Name: "create-empty-dto", Method: "POST", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.VerticalPackaging/package/custompackage", Body: json.RawMessage("{}"), Headers: http.Header{}, Status: 403},
		{Name: "update", Method: "PUT", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.VerticalPackaging/package/custompackage", Body: json.RawMessage("{\"Description\":\"\",\"PackageName\":\"Fixture/Package:@\",\"PolicyList\":[]}"), Headers: http.Header{}, Status: 403},
		{Name: "delete", Method: "DELETE", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.VerticalPackaging/package/custompackage/Fixture%2FPackage%3A%40", Body: nil, Headers: http.Header{}, Status: 403},
		{Name: "group-assign", Method: "POST", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.VerticalPackaging/package/group", Body: json.RawMessage("{\"GroupId\":\"fixture/group\",\"PackageName\":\"FixturePackage\",\"PolicyRankings\":[{\"PolicyType\":\"TeamsMeetingPolicy\",\"Rank\":0}]}"), Headers: http.Header{}, Status: 403},
		{Name: "group-empty-package", Method: "POST", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.VerticalPackaging/package/group", Body: json.RawMessage("{\"GroupId\":\"fixture/group\",\"PackageName\":\"\",\"PolicyRankings\":[]}"), Headers: http.Header{}, Status: 403},
		{Name: "group-null-package-default-rank", Method: "POST", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.VerticalPackaging/package/group", Body: json.RawMessage("{\"GroupId\":\"fixture/group\",\"PolicyRankings\":[{\"PolicyType\":\"Fixture\",\"Rank\":0}]}"), Headers: http.Header{}, Status: 403},
		{Name: "user-grant", Method: "POST", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.VerticalPackaging/package/batch", Body: json.RawMessage("{\"PackageType\":\"FixturePackage\",\"UserIds\":[\"fixture/a\",\"fixture/a\",\"fixture/b\"]}"), Headers: http.Header{}, Status: 403},
		{Name: "batch-grant", Method: "POST", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.VerticalPackaging/package/operation/batchpackage", Body: json.RawMessage("{\"PackageType\":\"FixturePackage\",\"UserIds\":[\"fixture/a\",\"fixture/a\",\"fixture/b\"]}"), Headers: http.Header{}, Status: 403},
	}
}

func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family = "policy_packages"
	if c.Source == "" {
		c.Source = source
	}
	if c.Mode == "" {
		c.Mode = auth.Delegated
	}
	if len(c.Steps) == 0 {
		c.Scope = "local"
	} else {
		c.Scope = "service"
		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	}
	r, err := contracttest.Run(t.Context(), c)
	if err != nil || !r.Passed() {
		t.Fatalf("package contract %s differs: %v %+v", c.Name, err, r.Differences())
	}
	return r
}

func discovery(t *testing.T) contracttest.Exchange {
	t.Helper()
	body, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Body: body}}
}

func wire(f requestCase, command string) contracttest.Exchange {
	h := f.Headers.Clone()
	if h == nil {
		h = http.Header{}
	}
	h.Set("Accept", "application/json")
	h.Set("X-MS-CmdletName", command)
	h.Set("X-MS-Target-Uri", forest)
	h[http.CanonicalHeaderKey("MPACmdlet")] = nil
	body := f.Body
	if string(body) == "null" {
		body = nil
	}
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: f.Method, URL: f.URL, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: contracttest.Reply{Status: f.Status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(f.Reply)}}
}
func customInput(update bool) resource.CustomPackageRequest {
	request := resource.CustomPackageRequest{Name: new("Fixture/Package:@"), Description: new(""), Policies: []resource.PolicyPair{{PolicyType: new("TeamsMeetingPolicy"), PolicyName: new("FixtureMeeting")}}}
	if update {
		request.Policies = []resource.PolicyPair{}
	}
	return request
}
func groupInput(name string) resource.GroupAssignmentRequest {
	input := resource.GroupAssignmentRequest{GroupID: "fixture/group", PackageName: new("FixturePackage"), Rankings: []resource.PolicyRanking{{PolicyType: new("TeamsMeetingPolicy"), Rank: 0}}}
	if name == "group-empty-package" {
		input.PackageName = new("")
		input.Rankings = []resource.PolicyRanking{}
	}
	if name == "group-null-package-default-rank" {
		input.PackageName = nil
		input.Rankings = []resource.PolicyRanking{{PolicyType: new("Fixture")}}
	}
	return input
}
func userInput() resource.UserAssignmentRequest {
	return resource.UserAssignmentRequest{PackageName: new("FixturePackage"), UserIDs: []string{"fixture/a", "fixture/a", "fixture/b"}}
}
func mapping(name string) []contracttest.Mapping {
	command, set := "Get-CsPolicyPackage", "Get1"
	switch {
	case name == "list" || name == "list-null":
		set = "Get"
	case name == "user-get":
		command, set = "Get-CsUserPolicyPackage", "Get"
	case name == "user-recommendations":
		command, set = "Get-CsUserPolicyPackageRecommendation", "Get"
	case strings.HasPrefix(name, "create"):
		command, set = "New-CsCustomPolicyPackage", "RequiredPolicyList"
	case name == "update":
		command, set = "Update-CsCustomPolicyPackage", "RequiredPolicyList"
	case name == "delete":
		command, set = "Remove-CsCustomPolicyPackage", "Remove"
	case strings.HasPrefix(name, "group"):
		command, set = "Grant-CsGroupPolicyPackageAssignment", "RequiredPolicyList"
	case strings.HasPrefix(name, "batch"):
		command, set = "New-CsBatchPolicyPackageAssignmentOperation", "NewExpanded"
	case strings.HasPrefix(name, "user"):
		command, set = "Grant-CsUserPolicyPackage", "GrantExpanded"
	}
	return []contracttest.Mapping{{Command: command, ParameterSet: set}}
}
func invoke(name string, ctx context.Context, sdk *msteams.Client) contracttest.Observation {
	s := sdk.TeamsAdminAPI.PolicyPackages
	var o contracttest.Observation
	switch {
	case name == "list" || name == "list-null":
		o.Value, o.Response, o.Err = s.List(ctx)
	case name == "user-get":
		o.Value, o.Response, o.Err = s.GetForUser(ctx, "fixture-user@example.invalid")
	case name == "user-recommendations":
		o.Value, o.Response, o.Err = s.RecommendationsForUser(ctx, "fixture-user@example.invalid")
	case name == "create-empty-dto":
		o.Value, o.Response, o.Err = s.Create(ctx, resource.CustomPackageRequest{})
	case strings.HasPrefix(name, "create"):
		o.Value, o.Response, o.Err = s.Create(ctx, customInput(false))
	case name == "update":
		o.Value, o.Response, o.Err = s.Update(ctx, customInput(true))
	case name == "delete":
		o.Value, o.Response, o.Err = s.DeleteByName(ctx, "Fixture/Package:@")
	case strings.HasPrefix(name, "group"):
		o.Value, o.Response, o.Err = s.AssignGroup(ctx, groupInput(name))
	case strings.HasPrefix(name, "batch"):
		o.Value, o.Response, o.Err = s.SubmitBatch(ctx, userInput())
	case strings.HasPrefix(name, "user"):
		o.Value, o.Response, o.Err = s.AssignUsers(ctx, userInput())
	default:
		o.Value, o.Response, o.Err = s.GetByName(ctx, "Fixture/Package:@")
	}
	if v, ok := o.Value.(*resource.AssignmentResult); ok && v != nil {
		if name == "user-partial" && (len(v.FailedUserIDs) != 1 || v.FailedUserErrorMessages == nil || len(v.FailedUserErrorMessages.LegacyUser) != 1 || string(v.FailedUserErrorMessages.RawFields["Future"]) != `["fixture-future"]` || !v.HTTPAccepted) {
			o.Err = errors.New("partial user/future error details lost")
		}
		if name == "group-domain-error" && (len(v.FailedAssignPolicies) != 1 || v.HTTPAccepted) {
			o.Err = errors.New("typed group failure data lost")
		}
		if name == "user-domain-error" && (len(v.FailedUserIDs) != 1 || v.HTTPAccepted) {
			o.Err = errors.New("typed user failure data lost")
		}
		if name == "batch-error" && (len(v.FailedCreatePolicies) != 1 || v.HTTPAccepted) {
			o.Err = errors.New("typed batch failure data lost")
		}
		if name == "batch-missing-id" && (v.OperationID != nil || !v.HTTPAccepted) {
			o.Err = errors.New("missing batch handle invented")
		}
	}
	return o
}
func TestUnitPackageRequestVectors(t *testing.T) {
	t.Parallel()
	for _, tc := range requestCases() {
		t.Run(tc.Name, func(t *testing.T) {
			mappings := mapping(tc.Name)
			c := contracttest.Case{Name: tc.Name, Mappings: mappings, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return invoke(tc.Name, ctx, sdk)
			}}
			if tc.Name == "create-empty-dto" {
				c.Want.Error = resource.ErrInput
			} else {
				c.Steps = []contracttest.Exchange{discovery(t), wire(tc, mappings[0].Command)}
				c.Want.Error = auth.ErrForbidden
			}
			run(t, c)
		})
	}
}

func TestUnitPackageInputObjectVariantsUsePackageName(t *testing.T) {
	t.Parallel()

	for _, remove := range []bool{false, true} {
		f := requestCase{Name: "get-input-object", Method: "GET", URL: origin + "/Teams.VerticalPackaging/package/Fixture%2FPackage%3A%40", Status: 403}
		command, set := "Get-CsPolicyPackage", "GetViaIdentity"
		if remove {
			f.Name, f.Method, f.URL = "delete-input-object", "DELETE", origin+"/Teams.VerticalPackaging/package/custompackage/Fixture%2FPackage%3A%40"
			command, set = "Remove-CsCustomPolicyPackage", "RemoveViaIdentity"
		}
		c := contracttest.Case{Want: contracttest.Expected{Error: auth.ErrForbidden}, Name: f.Name, Mappings: []contracttest.Mapping{{Command: command, ParameterSet: set}}, Steps: []contracttest.Exchange{discovery(t), wire(f, command)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			input := resource.Identity{PackageName: "Fixture/Package:@"}
			var o contracttest.Observation
			if remove {
				o.Value, o.Response, o.Err = sdk.TeamsAdminAPI.PolicyPackages.DeleteByIdentity(ctx, input)
			} else {
				o.Value, o.Response, o.Err = sdk.TeamsAdminAPI.PolicyPackages.GetByIdentity(ctx, input)
			}
			return o
		}}
		run(t, c)
	}
}
