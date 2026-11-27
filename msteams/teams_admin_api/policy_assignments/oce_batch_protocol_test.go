package policyassignments_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

// Authored from the pinned OCE LROS shapes and ordered MultiTenant regions,
// not captured regional eligibility. ProcessGetBatchOperationInfo stops at
// its first200; its status/definition projection is selected at IL79/117.
// Retaining earlier failed Attempts on later success is an SDK enhancement.
func TestUnitOCEAuthoredDefaultLookupKeepsEarlierFailureAndWinningRegion(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"status", "definition"} {
		t.Run(kind, func(t *testing.T) {
			reply := authoredOCEStatus

			if kind == "definition" {
				reply = authoredOCEDefinition

			}
			// Unknown body metadata must not override the actual request region.
			reply = strings.TrimSuffix(reply, "}") + `,"Region":"NOAM"}`
			run(t, testutil.Case{Name: "default-success-" + kind, Steps: []testutil.Exchange{batchWire(kind, resource.RegionNOAM, 404, ""), batchWire(kind, resource.RegionEMEA, 200, reply)}, Want: testutil.Expected{Body: []byte(reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				if kind == "status" {
					value, response, err := sdk.TeamsAdminAPI.PolicyAssignments.GetOCEBatchStatus(ctx, resource.OCEContext{TenantID: oceTenant}, oceOperation, nil)
					if err != nil {
						return testutil.Observation{Response: response, Err: err}
					}
					if value == nil || value.Value == nil || value.Region != resource.RegionEMEA || value.Value.Region != resource.RegionEMEA || value.Value.OperationID == nil || *value.Value.OperationID != oceOperation || string(value.Value.RawFields["Region"]) != `"NOAM"` || len(value.Attempts) != 2 || value.Attempts[0].Region != resource.RegionNOAM || value.Attempts[0].Response == nil || value.Attempts[0].Response.StatusCode() != 404 || value.Attempts[0].Err == nil || value.Attempts[1].Region != resource.RegionEMEA || value.Attempts[1].Err != nil {
						t.Fatal("successful default lookup lost the winning region, job or earlier failed receipt")
					}
					return testutil.Observation{Response: response}
				}
				value, response, err := sdk.TeamsAdminAPI.PolicyAssignments.GetOCEBatchDefinition(ctx, resource.OCEContext{TenantID: oceTenant}, oceOperation, nil)
				if err != nil {
					return testutil.Observation{Response: response, Err: err}
				}
				if value == nil || value.Value == nil || value.Region != resource.RegionEMEA || value.Value.Region != resource.RegionEMEA || value.Value.OperationID == nil || *value.Value.OperationID != oceOperation || value.Value.AcceptTime == nil || *value.Value.AcceptTime != 9007199254740993 || string(value.Value.RawFields["Region"]) != `"NOAM"` || len(value.Attempts) != 2 || value.Attempts[0].Response == nil || value.Attempts[0].Response.StatusCode() != 404 || value.Attempts[0].Err == nil || value.Attempts[1].Region != resource.RegionEMEA || value.Attempts[1].Err != nil {
					t.Fatal("successful definition lookup lost exact data, request region or prior failure")
				}
				return testutil.Observation{Response: response}
			}})
		})
	}
}

func TestUnitOCEAuthoredRegionalListPreservesPartialSuccessAndFailure(t *testing.T) {
	t.Parallel()
	const other = "00000000-0000-4000-8000-000000000006"
	noam := "[" + authoredOCESimpleStatus + "]"
	apac := "[" + strings.ReplaceAll(authoredOCESimpleStatus, oceOperation, other) + "]"
	run(t, testutil.Case{Name: "regional-partial-success", Steps: []testutil.Exchange{batchWire("status-list", resource.RegionNOAM, 200, noam), batchWire("status-list", resource.RegionEMEA, 403, ""), batchWire("status-list", resource.RegionAPAC, 200, apac)}, Want: testutil.Expected{Body: []byte(apac), Error: auth.ErrForbidden}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		value, response, err := sdk.TeamsAdminAPI.PolicyAssignments.ListOCEBatchStatuses(ctx, resource.OCEContext{TenantID: oceTenant}, nil)
		if !errors.Is(err, auth.ErrForbidden) || value == nil || len(value.Values) != 2 || value.Values[0] == nil || value.Values[1] == nil || value.Values[0].Region != resource.RegionNOAM || value.Values[1].Region != resource.RegionAPAC || value.Values[0].OperationID == nil || *value.Values[0].OperationID != oceOperation || value.Values[1].OperationID == nil || *value.Values[1].OperationID != other || len(value.Attempts) != 3 || value.Attempts[0].Err != nil || value.Attempts[1].Region != resource.RegionEMEA || value.Attempts[1].Response == nil || value.Attempts[1].Response.StatusCode() != 403 || !errors.Is(value.Attempts[1].Err, auth.ErrForbidden) || value.Attempts[2].Err != nil || response == nil || response.StatusCode() != 200 {
			t.Fatal("regional partial failure erased successful jobs or their exact request-region receipts")
		}
		return testutil.Observation{Response: response, Err: err}
	}})
}

func TestUnitOCEBodyCloseCancellationPreservesEarlierDecodedRegionOnly(t *testing.T) {
	t.Parallel()
	const other = "00000000-0000-4000-8000-000000000006"
	first := "[" + authoredOCESimpleStatus + "]"
	second := "[" + strings.ReplaceAll(authoredOCESimpleStatus, oceOperation, other) + "]"
	var stop context.CancelFunc
	step := batchWire("status-list", resource.RegionEMEA, 200, second)
	step.Reply.OnBodyClose = func() { stop() }
	run(t, testutil.Case{Name: "regional-body-close-cancel", Steps: []testutil.Exchange{batchWire("status-list", resource.RegionNOAM, 200, first), step}, Want: testutil.Expected{Body: []byte(second), Error: context.Canceled}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		operation, cancel := context.WithCancel(ctx)
		defer cancel()
		stop = cancel
		value, response, err := sdk.TeamsAdminAPI.PolicyAssignments.ListOCEBatchStatuses(operation, resource.OCEContext{TenantID: oceTenant}, nil)
		// NOAM decoded before cancellation. EMEA's body was received/closed
		// but cancellation won before its model decoder; APAC must not run.
		if !errors.Is(err, context.Canceled) || value == nil || len(value.Values) != 1 || value.Values[0] == nil || value.Values[0].Region != resource.RegionNOAM || value.Values[0].OperationID == nil || *value.Values[0].OperationID != oceOperation || len(value.Attempts) != 2 || value.Attempts[0].Err != nil || value.Attempts[1].Region != resource.RegionEMEA || !errors.Is(value.Attempts[1].Err, context.Canceled) || value.Attempts[1].Response != response || response == nil || response.StatusCode() != 200 {
			t.Fatal("regional cancellation lost prior results or published an undecoded later job")
		}
		return testutil.Observation{Response: response, Err: err}
	}})
}
