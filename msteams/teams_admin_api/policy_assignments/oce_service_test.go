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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

const oceSource = "contracts/endpoints/8.1.0/fixtures/policy-assignment-oce.json"
const oceTenant = "00000000-0000-4000-8000-000000000003"
const oceOperation = "00000000-0000-4000-8000-000000000004"
const oceGroup = "00000000-0000-4000-8000-000000000005"

func oceWire(method, url string, status int, reply string) contracttest.Exchange {
	h := http.Header{"X-MS-Tenant-Id": {oceTenant}}
	for _, name := range []string{"MPACmdlet", "X-MS-CmdletName", "X-MS-Target-Uri", "Content-Type"} {
		h[http.CanonicalHeaderKey(name)] = nil
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: url, Audience: auth.TeamsAdminAudience, Headers: h}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}

func oceMapping(name string) []contracttest.Mapping {
	command, set := "Get-CsGroupPolicyAssignments", "GetGroupPolicyAssignment"
	switch name {
	case "batch-status":
		command, set = "Get-CsBatchOperationStatus", "GetBatchOperationStatusForRegion"
	case "batch-list":
		command, set = "Get-CsBatchOperationStatus", "GetAllBatchOperationStatusForRegion"
	case "batch-definition":
		command, set = "Get-CsBatchOperationDefinition", "GetBatchOperationDefinitionForRegion"
	case "batch-reprocess":
		command, set = "Invoke-CsReprocessBatchOperation", "__AllParameterSets"
	case "group-list":
		set = "GetGroupPolicyAssignments"
	case "group-list-type":
		set = "GetPolicyAssignmentForGroups"
	case "group-list-all":
		set = "GetAllGroupPolicyAssignments"
	case "group-members", "group-members-all":
		command, set = "Get-GPAGroupMembers", "__AllParameterSets"
	case "user-membership":
		command, set = "Get-GPAUserMembership", "__AllParameterSets"
	case "group-refresh":
		command, set = "Invoke-CsRefreshGroupUsers", "__AllParameterSets"
	case "group-reprocess":
		command, set = "Invoke-CsReprocessGroupPolicyAssignment", "__AllParameterSets"
	}
	return []contracttest.Mapping{{Command: command, ParameterSet: set}}
}

func invokeOCE(name string, ctx context.Context, sdk *msteams.Client) contracttest.Observation {
	s := sdk.TeamsAdminAPI.PolicyAssignments
	input := resource.OCEContext{TenantID: oceTenant}
	var o contracttest.Observation
	region := resource.RegionNOAM
	switch name {
	case "batch-status":
		o.Value, o.Response, o.Err = s.GetOCEBatchStatus(ctx, input, oceOperation, &region)
	case "batch-list":
		o.Value, o.Response, o.Err = s.ListOCEBatchStatuses(ctx, input, &region)
	case "batch-definition":
		o.Value, o.Response, o.Err = s.GetOCEBatchDefinition(ctx, input, oceOperation, &region)
	case "batch-reprocess":
		o.Value, o.Response, o.Err = s.ReprocessOCEBatch(ctx, input, oceOperation, region)
	case "group-list":
		o.Value, o.Response, o.Err = s.ListOCEGroup(ctx, input, oceGroup)
	case "group-list-type":
		o.Value, o.Response, o.Err = s.ListOCEGroupsByPolicy(ctx, input, "Fixture")
	case "group-list-all":
		o.Value, o.Response, o.Err = s.ListOCEGroups(ctx, input)
	case "group-members", "group-members-all":
		o.Value, o.Response, o.Err = s.ListOCEGroupMembers(ctx, input, oceGroup, name == "group-members-all")
	case "user-membership":
		o.Value, o.Response, o.Err = s.ListOCEUserMembership(ctx, input, "00000000-0000-0000-0000-000000000000")
	case "group-refresh":
		o.Value, o.Response, o.Err = s.RefreshOCEGroupUsers(ctx, input, oceGroup, region, resource.OCERefreshOptions{PolicyType: new("Fixture")})
	case "group-reprocess":
		o.Value, o.Response, o.Err = s.ReprocessOCEGroupAssignment(ctx, input, oceGroup, "Fixture", region)
	default:
		o.Value, o.Response, o.Err = s.GetOCEGroup(ctx, input, oceGroup, "Fixture")
	}
	return o
}

func TestUnitPinnedOCEDirectRequests(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../" + oceSource)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name     string
			Status   int
			Response string
			Capture  struct {
				Method, URL                                                           string
				Body                                                                  *string
				Sends                                                                 int
				TenantMatches, CorrelationIsGuid, HasMpa, HasCmdletName, HasTargetUri bool
			}
		}
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Cases) != 18 {
		t.Fatal("invalid OCE capture")
	}
	var results []contracttest.Result
	for _, f := range fixture.Cases {
		if f.Capture.Sends != 1 || f.Capture.Body != nil || !f.Capture.TenantMatches || !f.Capture.CorrelationIsGuid || f.Capture.HasMpa || f.Capture.HasCmdletName || f.Capture.HasTargetUri {
			t.Fatal("unsafe or incomplete OCE sender observation")
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			c := contracttest.Case{Name: "oce-" + f.Name + "-" + string(mode), Mode: mode, Source: oceSource, Mappings: oceMapping(f.Name), Steps: []contracttest.Exchange{oceWire(f.Capture.Method, f.Capture.URL, f.Status, f.Response)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return invokeOCE(f.Name, ctx, sdk)
			}}
			if f.Status == 206 {
				c.Want.Error = resource.ErrResponse
			}
			if f.Status >= 400 {
				codes := []string(nil)
				if f.Status == 403 {
					codes = []string{"FixtureDenied"}
				}
				if f.Status == 503 {
					codes = []string{"FixtureUnavailable"}
				}
				c.Want.APIError = &client.APIError{Backend: client.TeamsAdmin, StatusCode: f.Status, Codes: codes}
			}
			results = append(results, run(t, c))
		}
	}
	evidence(t, results)
}

func TestUnitOCEOptionalQueryAndScopedEscaping(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	cases := []struct {
		name, method, path, reply string
		invoke                    func(context.Context, *resource.Service) contracttest.Observation
	}{
		{"empty-policy-query", "GET", "/Skype.Policy/groupPolicyAssignments/oce?policyType=", "[]", func(ctx context.Context, s *resource.Service) contracttest.Observation {
			v, r, e := s.ListOCEGroupsByPolicy(ctx, resource.OCEContext{TenantID: oceTenant}, "")
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}},
		{"escaped-group-policy", "GET", "/Skype.Policy/groupPolicyAssignments/oce/" + oceGroup + "/policyTypes/Fixture%2Fchild%3A%40", "{}", func(ctx context.Context, s *resource.Service) contracttest.Observation {
			v, r, e := s.GetOCEGroup(ctx, resource.OCEContext{TenantID: oceTenant}, strings.ToUpper(oceGroup), "Fixture/child:@")
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}},
		{"refresh-defaults", "POST", "/Skype.Policy/groupPolicyAssignments/" + oceGroup + "/refresh/users?tenantRegion=NOAM&userId=00000000-0000-0000-0000-000000000000&policyType=", "null", func(ctx context.Context, s *resource.Service) contracttest.Observation {
			v, r, e := s.RefreshOCEGroupUsers(ctx, resource.OCEContext{TenantID: oceTenant}, oceGroup, resource.RegionNOAM, resource.OCERefreshOptions{})
			if e == nil && (v == nil || !v.HTTPAccepted || v.Value != nil) {
				e = errors.New("nullable refresh acceptance changed")
			}
			return contracttest.Observation{Response: r, Err: e}
		}},
		{"refresh-explicit", "POST", "/Skype.Policy/groupPolicyAssignments/" + oceGroup + "/refresh/users?tenantRegion=NOAM&userId=" + oceOperation + "&policyType=Fixture%20%26%3F%40", "\"\"", func(ctx context.Context, s *resource.Service) contracttest.Observation {
			v, r, e := s.RefreshOCEGroupUsers(ctx, resource.OCEContext{TenantID: oceTenant}, oceGroup, resource.Region("noam"), resource.OCERefreshOptions{UserID: new(oceOperation), PolicyType: new("Fixture &?@")})
			if e == nil && (v == nil || v.Value == nil || *v.Value != "") {
				e = errors.New("explicit empty action value lost")
			}
			return contracttest.Observation{Response: r, Err: e}
		}},
	}
	for _, row := range cases {
		results = append(results, run(t, contracttest.Case{Name: "oce-" + row.name, Source: oceSource, Steps: []contracttest.Exchange{oceWire(row.method, origin+row.path, 200, row.reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			return row.invoke(ctx, sdk.TeamsAdminAPI.PolicyAssignments)
		}}))
	}
	evidence(t, results)
}
