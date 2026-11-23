package policyassignments_test

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

const ngtProvOrigin = "https://api.interfaces.records.teams.microsoft.com/Teams.InternalSupport/NgtProvOperations/"

func ngtProvMapping(command string) []contracttest.Mapping {
	return []contracttest.Mapping{{Command: command, ParameterSet: "__AllParameterSets"}}
}

func TestUnitOCENgtProvFailoverStatusUsesAcceptanceResponse(t *testing.T) {
	t.Parallel()
	reply, err := os.ReadFile("../migration/testdata/ngtprov_failover_status_403.json")
	if err != nil {
		t.Fatal(err)
	}
	run(t, contracttest.Case{Name: "oce-ngtprov-status-forbidden", Source: "msteams/teams_admin_api/migration/testdata/ngtprov_failover_status_403.json",
		Mappings: ngtProvMapping("Get-NgtProvInstanceFailOverStatus"),
		Steps:    []contracttest.Exchange{oceWire("GET", ngtProvOrigin+"FailoverStatus?region=NOAM", 403, string(reply))},
		Want:     contracttest.Expected{Status: 403, Body: reply, Error: auth.ErrForbidden},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.GetOCENgtProvFailoverStatus(ctx, resource.OCEContext{TenantID: oceTenant}, resource.RegionNOAM)
			return contracttest.Observation{Value: v, Response: r, Err: e}
		}})
}

func TestUnitOCENgtProvRequestVectors(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "oce-ngtprov-status", Source: oceSource, Mappings: ngtProvMapping("Get-NgtProvInstanceFailOverStatus"),
		Steps: []contracttest.Exchange{oceWire("GET", ngtProvOrigin+"FailoverStatus?region=EMEA", 200, `{"responseMessage":"Side A 100%","Extra":1}`)},
		Want:  contracttest.Expected{Status: 200},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.GetOCENgtProvFailoverStatus(ctx, resource.OCEContext{TenantID: oceTenant}, resource.Region("emea"))
			// The module's Newtonsoft decode is case-insensitive.
			if e == nil && (v == nil || v.ResponseMessage == nil || *v.ResponseMessage != "Side A 100%" || len(v.RawFields) != 2) {
				e = errors.New("failover status was not decoded")
			}
			return contracttest.Observation{Response: r, Err: e}
		}})
	// Synthetic only: the failover action is never sent to the lab.
	run(t, contracttest.Case{Name: "oce-ngtprov-move", Source: oceSource, Mappings: ngtProvMapping("Move-NgtProvInstance"),
		Steps: []contracttest.Exchange{oceWire("POST", ngtProvOrigin+"Failover/70/30?region=APAC", 200, "")},
		Want:  contracttest.Expected{Status: 200},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.MoveOCENgtProvInstance(ctx, resource.OCEContext{TenantID: oceTenant}, resource.RegionAPAC, 70, 30)
			if e == nil && (v == nil || v.ResponseMessage != nil) {
				e = errors.New("empty failover reply was not an empty result")
			}
			return contracttest.Observation{Response: r, Err: e}
		}})
	run(t, contracttest.Case{Name: "oce-ngtprov-move-malformed", Source: oceSource, Mappings: ngtProvMapping("Move-NgtProvInstance"),
		Steps: []contracttest.Exchange{oceWire("POST", ngtProvOrigin+"Failover/0/100?region=NOAM", 200, "[]")},
		Want:  contracttest.Expected{Status: 200, Error: client.ErrDecode},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, r, e := sdk.TeamsAdminAPI.PolicyAssignments.MoveOCENgtProvInstance(ctx, resource.OCEContext{TenantID: oceTenant}, resource.RegionNOAM, 0, 100)
			return contracttest.Observation{Response: r, Err: e}
		}})
}

func TestUnitOCENgtProvInvalidRegionFailsBeforeIO(t *testing.T) {
	t.Parallel()
	for name, region := range map[string]resource.Region{"int": resource.RegionINT, "gcch": resource.RegionGCCH, "blank": ""} {
		run(t, contracttest.Case{Name: "oce-ngtprov-region-" + name, Scope: "local", Source: oceSource, Want: contracttest.Expected{Error: resource.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				_, _, e1 := sdk.TeamsAdminAPI.PolicyAssignments.GetOCENgtProvFailoverStatus(ctx, resource.OCEContext{TenantID: oceTenant}, region)
				_, _, e2 := sdk.TeamsAdminAPI.PolicyAssignments.MoveOCENgtProvInstance(ctx, resource.OCEContext{TenantID: oceTenant}, region, 50, 50)
				if !errors.Is(e2, resource.ErrInput) {
					return contracttest.Observation{Err: e2}
				}
				return contracttest.Observation{Err: e1}
			}})
	}
}
