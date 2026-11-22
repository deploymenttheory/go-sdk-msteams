package teamsworkloadpolicy_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_work_load_policy"
)

func TestUnitActualModernReadAndCollection(t *testing.T) {
	t.Parallel()
	for _, list := range []bool{false, true} {
		file, path := "get_global.json", collection+"/configuration/Global"
		if list {
			file, path = "list.json", collection
		}
		raw := readResponseFixture(t, file)
		run(t, contracttest.Case{Name: strings.TrimSuffix(file, ".json"), Steps: []contracttest.Exchange{compiledWire("GET", path, "Get-CsTeamsWorkLoadPolicy", nil, 200, string(raw))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsWorkLoadPolicies
			if !list {
				v, r, e := s.GetByIdentity(ctx, "Global")
				if e == nil && (v.Identity != "Global" || v.AllowCalling == nil || v.AllowMeeting == nil || v.AllowMessaging == nil) {
					t.Fatal("actual getter projection lost")
				}
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}
			values, r, e := s.List(ctx)
			var expected []map[string]json.RawMessage
			if json.Unmarshal(raw, &expected) != nil || len(values) != len(expected) || len(values) == 0 || values[0].Identity != "Global" {
				t.Fatal("actual collection or Global ordering lost")
			}
			return contracttest.Observation{Value: values, Response: r, Err: e}
		}})
	}
}
func TestUnitModernFiltersUseActualCollection(t *testing.T) {
	t.Parallel()
	var values []resource.Policy
	if json.Unmarshal(readResponseFixture(t, "list.json"), &values) != nil {
		t.Fatal("invalid actual list")
	}
	for _, tc := range []struct {
		pattern string
		count   int
	}{{"Global", 1}, {"Glo?al", 1}, {"missing", 0}, {" ", len(values)}, {"", len(values)}} {
		run(t, contracttest.Case{Name: "filter", Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsWorkLoadPolicy", ParameterSet: "Filter"}}, Steps: []contracttest.Exchange{compiledWire("GET", collection, "Get-CsTeamsWorkLoadPolicy", nil, 200, string(readResponseFixture(t, "list.json")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			values, r, e := sdk.TeamsAdminAPI.TeamsWorkLoadPolicies.ListFiltered(ctx, tc.pattern)
			if e == nil && len(values) != tc.count {
				t.Fatal("modern filter differs")
			}
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
}
