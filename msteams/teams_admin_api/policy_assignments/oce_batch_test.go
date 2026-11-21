package policyassignments_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

func batchWire(kind string, region resource.Region, status int, reply string) contracttest.Exchange {
	path := "/Skype.Policy/assignments/operations/" + kind
	if kind != "status-list" {
		path += "/" + oceOperation
	} else {
		path = "/Skype.Policy/assignments/operations/status"
	}
	return oceWire("GET", origin+path+"?region="+string(region), status, reply)
}

func TestOCERegionalLookupPreservesAttemptsAndSourceOrder(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, kind := range []string{"status", "definition"} {
		command, set := "Get-CsBatchOperationStatus", "GetBatchOperationStatus"
		if kind == "definition" {
			command, set = "Get-CsBatchOperationDefinition", "GetBatchOperationDefinition"
		}
		c := contracttest.Case{Name: "oce-regional-" + kind, Source: oceSource, Mappings: []contracttest.Mapping{{Command: command, ParameterSet: set}}, Steps: []contracttest.Exchange{batchWire(kind, resource.RegionNOAM, 404, `{}`), batchWire(kind, resource.RegionEMEA, 200, `{"OperationId":"`+oceOperation+`","Region":"spoofed","OverallStatus":"Future","Commands":[],"AcceptTime":0}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.PolicyAssignments
			input := resource.OCEContext{TenantID: oceTenant}
			if kind == "status" {
				v, r, e := s.GetOCEBatchStatus(ctx, input, oceOperation, nil)
				if e == nil && (v == nil || v.Value == nil || v.Region != resource.RegionEMEA || v.Value.Region != resource.RegionEMEA || len(v.Attempts) != 2 || v.Attempts[0].Err == nil || v.Attempts[0].Response.StatusCode() != 404 || v.Attempts[1].Err != nil) {
					e = errors.New("regional status attempt history or source region lost")
				}
				return contracttest.Observation{Response: r, Err: e}
			}
			v, r, e := s.GetOCEBatchDefinition(ctx, input, oceOperation, nil)
			if e == nil && (v == nil || v.Value == nil || v.Region != resource.RegionEMEA || v.Value.Region != resource.RegionEMEA || len(v.Attempts) != 2 || v.Attempts[0].Err == nil || v.Attempts[1].Err != nil || v.Value.AcceptTime == nil || *v.Value.AcceptTime != 0) {
				e = errors.New("regional definition history or zero accept time lost")
			}
			return contracttest.Observation{Response: r, Err: e}
		}}
		results = append(results, run(t, c))
	}
	evidence(t, results)
}

func TestOCERegionalListRetainsPartialResultsAndFailures(t *testing.T) {
	t.Parallel()
	c := contracttest.Case{Name: "oce-regional-partial-list", Source: oceSource, Mappings: []contracttest.Mapping{{Command: "Get-CsBatchOperationStatus", ParameterSet: "GetAllBatchOperationStatus"}}, Steps: []contracttest.Exchange{batchWire("status-list", resource.RegionNOAM, 200, `[{"OperationId":"`+oceOperation+`","OverallStatus":"Future","Region":"spoofed"}]`), batchWire("status-list", resource.RegionEMEA, 503, `{"code":"FixtureUnavailable","message":"fixture-private"}`), batchWire("status-list", resource.RegionAPAC, 200, `[{"OperationId":"`+oceGroup+`","CompletedCount":0}]`)}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 503, Codes: []string{"FixtureUnavailable"}}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.ListOCEBatchStatuses(ctx, resource.OCEContext{TenantID: oceTenant}, nil)
		if v == nil || len(v.Values) != 2 || len(v.Attempts) != 3 || v.Values[0].Region != resource.RegionNOAM || v.Values[1].Region != resource.RegionAPAC || v.Attempts[0].Err != nil || v.Attempts[1].Err == nil || v.Attempts[2].Err != nil || e == nil {
			e = errors.New("regional partial list or errors lost")
		}
		return contracttest.Observation{Response: r, Err: e}
	}}
	evidence(t, []contracttest.Result{run(t, c)})
}

func TestOCERegionalFailuresAndMalformedSuccessRemainVisible(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	c := contracttest.Case{Name: "oce-regional-all-failed", Source: oceSource, Steps: []contracttest.Exchange{batchWire("status", resource.RegionNOAM, 404, `{}`), batchWire("status", resource.RegionEMEA, 404, `{}`), batchWire("status", resource.RegionAPAC, 404, `{}`)}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 404}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.GetOCEBatchStatus(ctx, resource.OCEContext{TenantID: oceTenant}, oceOperation, nil)
		if v == nil || v.Value != nil || len(v.Attempts) != 3 || e == nil {
			e = errors.New("all-region failure collapsed into success")
		}
		return contracttest.Observation{Response: r, Err: e}
	}}
	results = append(results, run(t, c))
	for _, reply := range []struct{ name, body string }{{"invalid", "not-json"}, {"null", "null"}} {
		c = contracttest.Case{Name: "oce-regional-malformed-" + reply.name, Source: oceSource, Steps: []contracttest.Exchange{batchWire("status", resource.RegionNOAM, 200, reply.body)}, Want: contracttest.Expected{Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.GetOCEBatchStatus(ctx, resource.OCEContext{TenantID: oceTenant}, oceOperation, nil)
			if v == nil || v.Value != nil || len(v.Attempts) != 1 || v.Attempts[0].Err == nil {
				e = errors.New("decode error triggered regional fallback")
			}
			return contracttest.Observation{Response: r, Err: e}
		}}
		results = append(results, run(t, c))
	}
	for _, reply := range []struct{ name, body string }{{"null", "null"}, {"null-element", "[null]"}, {"empty-body", ""}} {
		c = contracttest.Case{Name: "oce-regional-list-malformed-" + reply.name, Source: oceSource, Steps: []contracttest.Exchange{batchWire("status-list", resource.RegionNOAM, 200, reply.body)}, Want: contracttest.Expected{Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.ListOCEBatchStatuses(ctx, resource.OCEContext{TenantID: oceTenant}, new(resource.RegionNOAM))
			if v == nil || len(v.Attempts) != 1 || v.Attempts[0].Err == nil {
				e = errors.New("invalid regional list accepted")
			}
			return contracttest.Observation{Response: r, Err: e}
		}}
		results = append(results, run(t, c))
	}
	evidence(t, results)
}

func TestOCEActionsKeepAcceptanceAndNeverReplay(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, operation := range []string{"batch", "group", "refresh"} {
		for _, status := range []int{401, 403, 429, 500, 503} {
			path := "/Skype.Policy/assignments/operations/reprocess/" + oceOperation + "?region=NOAM"
			if operation == "group" {
				path = "/Skype.Policy/groupPolicyAssignments/" + oceGroup + "/policyTypes/Fixture/reprocess?tenantRegion=NOAM"
			}
			if operation == "refresh" {
				path = "/Skype.Policy/groupPolicyAssignments/" + oceGroup + "/refresh/users?tenantRegion=NOAM&userId=00000000-0000-0000-0000-000000000000&policyType=Fixture"
			}
			name := "batch-reprocess"
			if operation == "group" {
				name = "group-reprocess"
			}
			if operation == "refresh" {
				name = "group-refresh"
			}
			c := contracttest.Case{Name: fmt.Sprintf("oce-%s-denial-%d", operation, status), Source: oceSource, Mappings: oceMapping(name), Steps: []contracttest.Exchange{oceWire("POST", origin+path, status, `{"message":"fixture-private"}`)}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return invokeOCE(name, ctx, sdk)
			}}
			results = append(results, run(t, c))
		}
	}
	for _, reply := range []struct {
		name, body string
		valid      bool
	}{{"quoted", `"00000000-0000-4000-8000-000000000006"`, true}, {"bare", "00000000-0000-4000-8000-000000000006", true}, {"wrong-envelope", `{"OperationId":"fixture"}`, false}, {"malformed", `not-json`, false}, {"null", `null`, false}} {
		c := contracttest.Case{Name: "oce-reprocess-reply-" + reply.name, Source: oceSource, Mappings: oceMapping("batch-reprocess"), Steps: []contracttest.Exchange{oceWire("POST", origin+"/Skype.Policy/assignments/operations/reprocess/"+oceOperation+"?region=NOAM", 202, reply.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.ReprocessOCEBatch(ctx, resource.OCEContext{TenantID: oceTenant}, oceOperation, resource.RegionNOAM)
			if v == nil || !v.HTTPAccepted || (reply.valid && (v.OperationID == nil || *v.OperationID != "00000000-0000-4000-8000-000000000006")) || (!reply.valid && v.OperationID != nil) {
				e = errors.New("reprocess acceptance or validated handle lost")
			}
			if strings.Contains(fmt.Sprintf("%v %#v", v, v), "00000000-") {
				e = errors.New("reprocess formatter exposed a handle")
			}
			return contracttest.Observation{Response: r, Err: e}
		}}
		if !reply.valid {
			c.Want.Error = client.ErrDecode
		}
		results = append(results, run(t, c))
	}
	evidence(t, results)
}
