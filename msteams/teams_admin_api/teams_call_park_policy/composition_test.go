package teamscallparkpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_call_park_policy"
)

func TestUnitCallParkCreateDenialNeverReadsBack(t *testing.T) {
	t.Parallel()
	for _, status := range []int{403, 429, 503} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			run(t, contracttest.Case{Name: "create-denial", Source: "msteams/teams_admin_api/teams_call_park_policy/composition_test.go", Steps: []contracttest.Exchange{discovery(t), wire("POST", collection, newCommand, json.RawMessage(`{"Identity":"Global","AllowCallPark":false,"Description":"","ParkTimeoutSeconds":0,"PickupRangeStart":9007199254740993,"PickupRangeEnd":9223372036854775807}`), status, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsCallParkPolicies.Create(ctx, resource.CreateRequest{Identity: "Global", Settings: fixtureSettings()})
				var apiError *client.APIError
				var readBack *resource.ReadBackError
				if !errors.As(e, &apiError) || apiError.StatusCode != status || v != nil || errors.As(e, &readBack) {
					t.Error("failed POST fabricated creation or attempted a success read-back")
				}
				return contracttest.Observation{Response: r}
			}})
		})
	}
}

func TestUnitCallParkGroupGrantDenialsNeverReportAcceptance(t *testing.T) {
	t.Parallel()
	for _, status := range []int{403, 429, 503} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			run(t, contracttest.Case{Name: "group-denial", Source: "msteams/teams_admin_api/teams_call_park_policy/composition_test.go", Steps: []contracttest.Exchange{discovery(t), wire("POST", "/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsCallParkPolicy", grantCommand, json.RawMessage(`{"policyName":"Fixture"}`), status, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsCallParkPolicies.AssignToGroup(ctx, "fixture/group", "Fixture", nil)
				var apiError *client.APIError
				if !errors.As(e, &apiError) || apiError.StatusCode != status || v != nil && v.Accepted {
					t.Error("denied group grant lost its HTTP failure or invented acceptance")
				}
				return contracttest.Observation{Response: r}
			}})
		})
	}
}

func TestUnitCallParkGrantRankAndSparseGlobalCompositionRequestDenials(t *testing.T) {
	t.Parallel()

	for _, rank := range []int32{0, 2, -1} {
		body := map[string]any{"policyName": "Fixture"}
		name := "rank-zero-omitted"
		if rank != 0 {
			body["priority"] = rank
			name = "rank-positive"
			if rank < 0 {
				name = "rank-negative"
			}
		}
		raw, _ := json.Marshal(body)
		run(t, contracttest.Case{Want: contracttest.Expected{Error: auth.ErrForbidden}, Name: name, Source: wrapperSource, Mappings: []contracttest.Mapping{{Command: grantCommand, ParameterSet: "GrantToGroup"}}, Steps: []contracttest.Exchange{discovery(t), wire("POST", "/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsCallParkPolicy", grantCommand, raw, 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsCallParkPolicies.Assign(ctx, resource.Assignment{Group: &resource.GroupGrant{ID: "fixture/group", PolicyName: "Fixture", Rank: new(rank)}})
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
	for _, bound := range []bool{false, true} {
		name := "identity-only"
		body := `{"Identity":"Global"}`
		input := resource.UpdateRequest{}
		if bound {
			name = "false-zero-empty"
			body = `{"Identity":"Global","AllowCallPark":false,"Description":"","ParkTimeoutSeconds":0,"PickupRangeStart":0,"PickupRangeEnd":0}`
			input.AllowCallPark = new(false)
			input.Description = new("")
			input.ParkTimeoutSeconds = new(int64(0))
			input.PickupRangeStart = new(int64(0))
			input.PickupRangeEnd = new(int64(0))
		}
		run(t, contracttest.Case{Want: contracttest.Expected{Error: auth.ErrForbidden}, Name: name, Source: wrapperSource, Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{discovery(t), wire("PUT", item, setCommand, json.RawMessage(body), 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			r, e := sdk.TeamsAdminAPI.TeamsCallParkPolicies.UpdateGlobal(ctx, input)
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
}
