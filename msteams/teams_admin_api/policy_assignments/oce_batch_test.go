package policyassignments_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
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

func TestUnitOCERegionalLookupRetainsFailedAttemptsAndOrder(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"status", "definition"} {
		command, set := "Get-CsBatchOperationStatus", "GetBatchOperationStatus"
		if kind == "definition" {
			command, set = "Get-CsBatchOperationDefinition", "GetBatchOperationDefinition"
		}
		c := contracttest.Case{Name: "regional-failures-" + kind, Source: oceSource, Mappings: []contracttest.Mapping{{Command: command, ParameterSet: set}}, Steps: []contracttest.Exchange{batchWire(kind, resource.RegionNOAM, 404, ""), batchWire(kind, resource.RegionEMEA, 404, ""), batchWire(kind, resource.RegionAPAC, 404, "")}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 404}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			if kind == "status" {
				v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.GetOCEBatchStatus(ctx, resource.OCEContext{TenantID: oceTenant}, oceOperation, nil)
				if v == nil || v.Value != nil || len(v.Attempts) != 3 || v.Attempts[0].Region != resource.RegionNOAM || v.Attempts[1].Region != resource.RegionEMEA || v.Attempts[2].Region != resource.RegionAPAC {
					e = errors.New("regional failure history lost")
				}
				return contracttest.Observation{Response: r, Err: e}
			}
			v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.GetOCEBatchDefinition(ctx, resource.OCEContext{TenantID: oceTenant}, oceOperation, nil)
			if v == nil || v.Value != nil || len(v.Attempts) != 3 || v.Attempts[0].Region != resource.RegionNOAM || v.Attempts[1].Region != resource.RegionEMEA || v.Attempts[2].Region != resource.RegionAPAC {
				e = errors.New("regional definition history lost")
			}
			return contracttest.Observation{Response: r, Err: e}
		}}
		run(t, c)
	}
}

func TestUnitOCERegionalListRetainsEveryFailedAttempt(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "regional-list-failures", Source: oceSource, Mappings: []contracttest.Mapping{{Command: "Get-CsBatchOperationStatus", ParameterSet: "GetAllBatchOperationStatus"}}, Steps: []contracttest.Exchange{batchWire("status-list", resource.RegionNOAM, 403, ""), batchWire("status-list", resource.RegionEMEA, 403, ""), batchWire("status-list", resource.RegionAPAC, 403, "")}, Want: contracttest.Expected{Error: auth.ErrForbidden}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.ListOCEBatchStatuses(ctx, resource.OCEContext{TenantID: oceTenant}, nil)
		if v == nil || len(v.Values) != 0 || len(v.Attempts) != 3 || v.Attempts[0].Err == nil || v.Attempts[1].Err == nil || v.Attempts[2].Err == nil || e == nil {
			e = errors.New("failed regional list invented success")
		}
		return contracttest.Observation{Response: r, Err: e}
	}})
}

func TestUnitOCERegionalFailuresAndMalformedSuccessRemainVisible(t *testing.T) {
	t.Parallel()

	c := contracttest.Case{Name: "oce-regional-all-failed", Source: oceSource, Steps: []contracttest.Exchange{batchWire("status", resource.RegionNOAM, 404, `{}`), batchWire("status", resource.RegionEMEA, 404, `{}`), batchWire("status", resource.RegionAPAC, 404, `{}`)}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 404}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.GetOCEBatchStatus(ctx, resource.OCEContext{TenantID: oceTenant}, oceOperation, nil)
		if v == nil || v.Value != nil || len(v.Attempts) != 3 || e == nil {
			e = errors.New("all-region failure collapsed into success")
		}
		return contracttest.Observation{Response: r, Err: e}
	}}
	run(t, c)
	for _, reply := range []struct{ name, body string }{{"invalid", "not-json"}, {"null", "null"}} {
		c = contracttest.Case{Name: "oce-regional-malformed-" + reply.name, Source: oceSource, Steps: []contracttest.Exchange{batchWire("status", resource.RegionNOAM, 200, reply.body)}, Want: contracttest.Expected{Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.GetOCEBatchStatus(ctx, resource.OCEContext{TenantID: oceTenant}, oceOperation, nil)
			if v == nil || v.Value != nil || len(v.Attempts) != 1 || v.Attempts[0].Err == nil {
				e = errors.New("decode error triggered regional fallback")
			}
			return contracttest.Observation{Response: r, Err: e}
		}}
		run(t, c)
	}
	for _, reply := range []struct{ name, body string }{{"null", "null"}, {"null-element", "[null]"}, {"empty-body", ""}} {
		c = contracttest.Case{Name: "oce-regional-list-malformed-" + reply.name, Source: oceSource, Steps: []contracttest.Exchange{batchWire("status-list", resource.RegionNOAM, 200, reply.body)}, Want: contracttest.Expected{Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.ListOCEBatchStatuses(ctx, resource.OCEContext{TenantID: oceTenant}, new(resource.RegionNOAM))
			if v == nil || len(v.Attempts) != 1 || v.Attempts[0].Err == nil {
				e = errors.New("invalid regional list accepted")
			}
			return contracttest.Observation{Response: r, Err: e}
		}}
		run(t, c)
	}
}

func TestUnitOCEActionRequestsNeverReplay(t *testing.T) {
	t.Parallel()

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
			run(t, c)
		}
	}
}
