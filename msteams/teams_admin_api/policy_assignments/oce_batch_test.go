package policyassignments_test

import (
	"context"
	"errors"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

func batchWire(kind string, region resource.Region, status int, reply string) testutil.Exchange {
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

		c := testutil.Case{Name: "regional-failures-" + kind, Steps: []testutil.Exchange{batchWire(kind, resource.RegionNOAM, 404, ""), batchWire(kind, resource.RegionEMEA, 404, ""), batchWire(kind, resource.RegionAPAC, 404, "")}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 404}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			if kind == "status" {
				v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.GetOCEBatchStatus(ctx, resource.OCEContext{TenantID: oceTenant}, oceOperation, nil)
				if v == nil || v.Value != nil || len(v.Attempts) != 3 || v.Attempts[0].Region != resource.RegionNOAM || v.Attempts[1].Region != resource.RegionEMEA || v.Attempts[2].Region != resource.RegionAPAC {
					e = errors.New("regional failure history lost")
				}
				return testutil.Observation{Response: r, Err: e}
			}
			v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.GetOCEBatchDefinition(ctx, resource.OCEContext{TenantID: oceTenant}, oceOperation, nil)
			if v == nil || v.Value != nil || len(v.Attempts) != 3 || v.Attempts[0].Region != resource.RegionNOAM || v.Attempts[1].Region != resource.RegionEMEA || v.Attempts[2].Region != resource.RegionAPAC {
				e = errors.New("regional definition history lost")
			}
			return testutil.Observation{Response: r, Err: e}
		}}
		run(t, c)
	}
}

func TestUnitOCERegionalListRetainsEveryFailedAttempt(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "regional-list-failures", Steps: []testutil.Exchange{batchWire("status-list", resource.RegionNOAM, 403, ""), batchWire("status-list", resource.RegionEMEA, 403, ""), batchWire("status-list", resource.RegionAPAC, 403, "")}, Want: testutil.Expected{Error: auth.ErrForbidden}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.ListOCEBatchStatuses(ctx, resource.OCEContext{TenantID: oceTenant}, nil)
		if v == nil || len(v.Values) != 0 || len(v.Attempts) != 3 || v.Attempts[0].Err == nil || v.Attempts[1].Err == nil || v.Attempts[2].Err == nil || e == nil {
			e = errors.New("failed regional list invented success")
		}
		return testutil.Observation{Response: r, Err: e}
	}})
}

func TestUnitOCERegionalFailuresAndMalformedSuccessRemainVisible(t *testing.T) {
	t.Parallel()

	c := testutil.Case{Name: "oce-regional-all-failed", Steps: []testutil.Exchange{batchWire("status", resource.RegionNOAM, 404, `{}`), batchWire("status", resource.RegionEMEA, 404, `{}`), batchWire("status", resource.RegionAPAC, 404, `{}`)}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 404}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.GetOCEBatchStatus(ctx, resource.OCEContext{TenantID: oceTenant}, oceOperation, nil)
		if v == nil || v.Value != nil || len(v.Attempts) != 3 || e == nil {
			e = errors.New("all-region failure collapsed into success")
		}
		return testutil.Observation{Response: r, Err: e}
	}}
	run(t, c)
	for _, reply := range []struct{ name, body string }{{"invalid", "not-json"}, {"null", "null"}} {
		c = testutil.Case{Name: "oce-regional-malformed-" + reply.name, Steps: []testutil.Exchange{batchWire("status", resource.RegionNOAM, 200, reply.body)}, Want: testutil.Expected{Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.GetOCEBatchStatus(ctx, resource.OCEContext{TenantID: oceTenant}, oceOperation, nil)
			if v == nil || v.Value != nil || len(v.Attempts) != 1 || v.Attempts[0].Err == nil {
				e = errors.New("decode error triggered regional fallback")
			}
			return testutil.Observation{Response: r, Err: e}
		}}
		run(t, c)
	}
	for _, reply := range []struct{ name, body string }{{"null", "null"}, {"null-element", "[null]"}, {"empty-body", ""}} {
		c = testutil.Case{Name: "oce-regional-list-malformed-" + reply.name, Steps: []testutil.Exchange{batchWire("status-list", resource.RegionNOAM, 200, reply.body)}, Want: testutil.Expected{Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.ListOCEBatchStatuses(ctx, resource.OCEContext{TenantID: oceTenant}, new(resource.RegionNOAM))
			if v == nil || len(v.Attempts) != 1 || v.Attempts[0].Err == nil {
				e = errors.New("invalid regional list accepted")
			}
			return testutil.Observation{Response: r, Err: e}
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
			c := testutil.Case{Name: fmt.Sprintf("oce-%s-denial-%d", operation, status), Steps: []testutil.Exchange{oceWire("POST", origin+path, status, `{"message":"fixture-private"}`)}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				return invokeOCE(name, ctx, sdk)
			}}
			run(t, c)
		}
	}
}
