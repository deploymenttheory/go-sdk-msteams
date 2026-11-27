package teamsfilespolicy_test

import (
	"context"
	"fmt"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_files_policy"
	"github.com/go-resty/resty/v2"
)

func TestUnitFilesPolicyFiltersUseRealList(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "list.json")
	for i, c := range []struct {
		Pattern string
		Want    []string
	}{
		{"", []string{"Global"}}, {"*", []string{"Global"}}, {"global", []string{"Global"}},
		{"Fixture*", []string{}}, {" ", []string{}}, {"Global/*", []string{}},
	} {
		run(t, testutil.Case{Name: fmt.Sprintf("filter-%d", i), Steps: []testutil.Exchange{compiledWire("GET", collection, getCommand, nil, 200, string(raw))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			service := sdk.TeamsAdminAPI.TeamsFilesPolicies
			operation := service.List
			if c.Pattern != "" {
				operation = func(ctx context.Context) ([]*resource.Policy, *resty.Response, error) {
					return service.ListFiltered(ctx, c.Pattern)
				}
			}
			values, response, err := operation(ctx)
			identities := make([]string, 0, len(values))
			for _, value := range values {
				if value == nil || value.Identity == nil {
					t.Fatal("filtered resource lost identity")
				}
				identities = append(identities, *value.Identity)
			}
			if !reflect.DeepEqual(identities, c.Want) {
				t.Fatal("local filtering changed actual list identities")
			}
			return testutil.Observation{Response: response, Err: err}
		}})
	}
}

func TestUnitFilesPolicyParameterlessGrantHasNoDispatch(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "grant-no-scope", Want: testutil.Expected{Error: policy.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		value, response, err := sdk.TeamsAdminAPI.TeamsFilesPolicies.Assign(ctx, policy.Assignment{})
		return testutil.Observation{Value: value, Response: response, Err: err}
	}})
}
