package policyassignments_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

const ngtProvOrigin = "https://api.interfaces.records.teams.microsoft.com/Teams.InternalSupport/NgtProvOperations/"

func TestUnitOCENgtProvFailoverStatusUsesAcceptanceResponse(t *testing.T) {
	t.Parallel()
	reply, err := os.ReadFile("../migration/testdata/ngtprov_failover_status_403.json")
	if err != nil {
		t.Fatal(err)
	}
	run(t, testutil.Case{Name: "oce-ngtprov-status-forbidden",

		Steps: []testutil.Exchange{oceWire("GET", ngtProvOrigin+"FailoverStatus?region=NOAM", 403, string(reply))},
		Want:  testutil.Expected{Status: 403, Body: reply, Error: auth.ErrForbidden},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.GetOCENgtProvFailoverStatus(ctx, resource.OCEContext{TenantID: oceTenant}, resource.RegionNOAM)
			return testutil.Observation{Value: v, Response: r, Err: e}
		}})
}

func TestUnitOCENgtProvRequestVectors(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "oce-ngtprov-status",
		Steps: []testutil.Exchange{oceWire("GET", ngtProvOrigin+"FailoverStatus?region=EMEA", 200, `{"responseMessage":"Side A 100%","Extra":1}`)},
		Want:  testutil.Expected{Status: 200},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.GetOCENgtProvFailoverStatus(ctx, resource.OCEContext{TenantID: oceTenant}, resource.Region("emea"))
			// The module's Newtonsoft decode is case-insensitive.
			if e == nil && (v == nil || v.ResponseMessage == nil || *v.ResponseMessage != "Side A 100%" || len(v.RawFields) != 2) {
				e = errors.New("failover status was not decoded")
			}
			return testutil.Observation{Response: r, Err: e}
		}})
	// Synthetic only: the failover action is never sent to the lab.
	run(t, testutil.Case{Name: "oce-ngtprov-move",
		Steps: []testutil.Exchange{oceWire("POST", ngtProvOrigin+"Failover/70/30?region=APAC", 200, "")},
		Want:  testutil.Expected{Status: 200},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.MoveOCENgtProvInstance(ctx, resource.OCEContext{TenantID: oceTenant}, resource.RegionAPAC, 70, 30)
			if e == nil && (v == nil || v.ResponseMessage != nil) {
				e = errors.New("empty failover reply was not an empty result")
			}
			return testutil.Observation{Response: r, Err: e}
		}})
	run(t, testutil.Case{Name: "oce-ngtprov-move-malformed",
		Steps: []testutil.Exchange{oceWire("POST", ngtProvOrigin+"Failover/0/100?region=NOAM", 200, "[]")},
		Want:  testutil.Expected{Status: 200, Error: client.ErrDecode},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			_, r, e := sdk.TeamsAdminAPI.PolicyAssignments.MoveOCENgtProvInstance(ctx, resource.OCEContext{TenantID: oceTenant}, resource.RegionNOAM, 0, 100)
			return testutil.Observation{Response: r, Err: e}
		}})
}

func TestUnitOCENgtProvInvalidRegionFailsBeforeIO(t *testing.T) {
	t.Parallel()
	for name, region := range map[string]resource.Region{"int": resource.RegionINT, "gcch": resource.RegionGCCH, "blank": ""} {
		run(t, testutil.Case{Name: "oce-ngtprov-region-" + name, Want: testutil.Expected{Error: resource.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				_, _, e1 := sdk.TeamsAdminAPI.PolicyAssignments.GetOCENgtProvFailoverStatus(ctx, resource.OCEContext{TenantID: oceTenant}, region)
				_, _, e2 := sdk.TeamsAdminAPI.PolicyAssignments.MoveOCENgtProvInstance(ctx, resource.OCEContext{TenantID: oceTenant}, region, 50, 50)
				if !errors.Is(e2, resource.ErrInput) {
					return testutil.Observation{Err: e2}
				}
				return testutil.Observation{Err: e1}
			}})
	}
}
