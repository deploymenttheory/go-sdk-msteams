package acsfederationconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/acs_federation_configuration"
)

func TestUnitProviderFilteringIsLocalAndGlobalOrderIsExact(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	reply := `[{"Identity":"global","Future":null},{"Identity":"tag:Fixture","AllowedAcsResources":[]},{"Identity":"Global","AllowedAcsResources":["a"],"Future":{"value":false}},{"Identity":"Other"}]`
	for index, pattern := range []string{"", "Fixture*"} {
		results = append(results, run(t, contracttest.Case{Name: fmt.Sprint("filter-", index), Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsAcsFederationConfiguration", ParameterSet: "Filter"}}, Steps: []contracttest.Exchange{wire("GET", "/Skype.Policy/configurations/TeamsAcsFederationConfiguration", "Get-CsTeamsAcsFederationConfiguration", nil, 200, reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			values, r, e := sdk.TeamsAdminAPI.ACSFederationConfigurations.ListFiltered(ctx, pattern)
			if e == nil {
				want := []string{"Global", "global", "tag:Fixture", "Other"}
				if pattern != "" {
					want = []string{"tag:Fixture"}
				}
				var ids []string
				for _, v := range values {
					if v == nil || v.Identity == nil {
						e = errors.New("identity lost")
						break
					}
					ids = append(ids, *v.Identity)
					encoded, err := json.Marshal(v)
					var fields map[string]json.RawMessage
					if err != nil || json.Unmarshal(encoded, &fields) != nil || !reflect.DeepEqual(fields, v.RawFields) {
						e = errors.New("raw snapshot changed")
					}
				}
				if !reflect.DeepEqual(ids, want) {
					e = errors.New("local filtering or exact Global-first order changed")
				}
			}
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}
func TestUnitWrongBackendAndNilServiceRejectLocally(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "wrong-backend", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		graph, e := client.NewScope(client.Graph, sdk.TeamsAdminAPI.Scope().Session())
		if e != nil {
			return contracttest.Observation{Err: e}
		}
		if _, _, e := resource.NewService(graph).GetByIdentity(ctx, "Global"); !errors.Is(e, client.ErrDestination) {
			return contracttest.Observation{Err: errors.New("wrong backend accepted")}
		}
		var missing *resource.Service
		if _, _, e := missing.List(ctx); !errors.Is(e, client.ErrClosed) {
			return contracttest.Observation{Err: errors.New("nil service accepted")}
		}
		return contracttest.Observation{}
	}})
}
