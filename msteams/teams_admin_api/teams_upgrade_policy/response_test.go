package teamsupgradepolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_upgrade_policy"
)

const collection = "/Skype.Policy/configurations/TeamsUpgradePolicy"

func TestUnitActualModernReadAndCollection(t *testing.T) {
	t.Parallel()
	for _, list := range []bool{false, true} {
		file, path := "get_global.json", collection+"/configuration/Global"
		if list {
			file, path = "list.json", collection
		}
		raw := readResponseFixture(t, file)
		run(t, contracttest.Case{Name: strings.TrimSuffix(file, ".json"), Steps: []contracttest.Exchange{compiledWire("GET", path, "Get-CsTeamsUpgradePolicy", nil, 200, string(raw))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsUpgradePolicies
			if !list {
				v, r, e := s.GetByIdentity(ctx, "Global")
				if e == nil && (v.Identity != "Global" || v.Mode == nil || v.NotifySfbUsers == nil || v.Action == nil) {
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
		run(t, contracttest.Case{Name: "filter", Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsUpgradePolicy", ParameterSet: "Filter"}}, Steps: []contracttest.Exchange{compiledWire("GET", collection, "Get-CsTeamsUpgradePolicy", nil, 200, string(readResponseFixture(t, "list.json")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			values, r, e := sdk.TeamsAdminAPI.TeamsUpgradePolicies.ListFiltered(ctx, tc.pattern)
			if e == nil && len(values) != tc.count {
				t.Fatal("modern filter differs")
			}
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
}
func TestUnitGrantScopeMigrationAndCancellationRejectBeforeHTTP(t *testing.T) {
	t.Parallel()
	runExplicit(t, contracttest.Case{Name: "local-validation", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.TeamsUpgradePolicies
		for _, input := range []resource.Assignment{{}, {User: &resource.UserAssignment{Identity: "fixture"}, Tenant: &resource.TenantAssignment{}}} {
			if _, _, e := s.Assign(ctx, input); !errors.Is(e, resource.ErrInput) {
				t.Fatal("ambiguous scope reached HTTP")
			}
		}
		for _, flag := range []bool{false, true} {
			if _, _, e := s.Assign(ctx, resource.Assignment{User: &resource.UserAssignment{Identity: "fixture"}, MigrateMeetingsToTeams: &flag}); !errors.Is(e, resource.ErrMigrationParameterBinding) {
				t.Fatal("pinned binding failure did not reject before HTTP")
			}
		}
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid identity reached HTTP")
			}
		}
		if _, _, e := s.ListFiltered(ctx, "[broken"); !errors.Is(e, resource.ErrInput) {
			t.Fatal("malformed filter reached HTTP")
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.Assign(canceled, resource.Assignment{Tenant: &resource.TenantAssignment{}}); !errors.Is(e, context.Canceled) {
			t.Fatal("cancellation lost")
		}
		for _, value := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := value.List(ctx); !errors.Is(e, client.ErrClosed) {
				t.Fatal("nil backend reached HTTP")
			}
		}
		return contracttest.Observation{}
	}})
}
