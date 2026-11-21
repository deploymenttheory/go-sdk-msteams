package teamscortanapolicy_test

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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_cortana_policy"
)

func TestUnitCortanaCreateDenialNeverReadsBack(t *testing.T) {
	t.Parallel()
	for _, status := range []int{403, 429, 503} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			run(t, contracttest.Case{Name: "create-denial", Source: "msteams/teams_admin_api/teams_cortana_policy/composition_test.go", Steps: []contracttest.Exchange{discovery(t), wire("POST", collection, newCommand, json.RawMessage(`{"Identity":"Global","AllowCortanaVoiceInvocation":false,"Description":"","CortanaVoiceInvocationMode":"FutureMode","AllowCortanaAmbientListening":true,"AllowCortanaInContextSuggestions":false}`), status, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsCortanaPolicies.Create(ctx, resource.CreateRequest{Identity: "Global", Settings: fixtureSettings()})
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

func TestUnitCortanaGroupGrantDenialsNeverReportAcceptance(t *testing.T) {
	t.Parallel()
	for _, status := range []int{403, 429, 503} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			run(t, contracttest.Case{Name: "group-denial", Source: "msteams/teams_admin_api/teams_cortana_policy/composition_test.go", Steps: []contracttest.Exchange{discovery(t), wire("POST", "/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsCortanaPolicy", grantCommand, json.RawMessage(`{"policyName":"Fixture"}`), status, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsCortanaPolicies.AssignToGroup(ctx, "fixture/group", "Fixture", nil)
				var apiError *client.APIError
				if !errors.As(e, &apiError) || apiError.StatusCode != status || v != nil && v.Accepted {
					t.Error("denied group grant lost its HTTP failure or invented acceptance")
				}
				return contracttest.Observation{Response: r}
			}})
		})
	}
}

func TestUnitCortanaGrantRankAndSparseGlobalCompositionRequestDenials(t *testing.T) {
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
		run(t, contracttest.Case{Want: contracttest.Expected{Error: auth.ErrForbidden}, Name: name, Source: wrapperSource, Mappings: []contracttest.Mapping{{Command: grantCommand, ParameterSet: "GrantToGroup"}}, Steps: []contracttest.Exchange{discovery(t), wire("POST", "/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsCortanaPolicy", grantCommand, raw, 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsCortanaPolicies.Assign(ctx, resource.Assignment{Group: &resource.GroupGrant{ID: "fixture/group", PolicyName: "Fixture", Rank: new(rank)}})
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
	for _, bound := range []bool{false, true} {
		name := "identity-only"
		body := `{"Identity":"Global"}`
		input := resource.UpdateRequest{}
		if bound {
			name = "false-empty"
			body = `{"Identity":"Global","AllowCortanaVoiceInvocation":false,"Description":"","CortanaVoiceInvocationMode":"","AllowCortanaAmbientListening":false,"AllowCortanaInContextSuggestions":false}`
			input.AllowCortanaVoiceInvocation = new(false)
			input.Description = new("")
			input.CortanaVoiceInvocationMode = new("")
			input.AllowCortanaAmbientListening = new(false)
			input.AllowCortanaInContextSuggestions = new(false)
		}
		run(t, contracttest.Case{Want: contracttest.Expected{Error: auth.ErrForbidden}, Name: name, Source: wrapperSource, Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{discovery(t), wire("PUT", item, setCommand, json.RawMessage(body), 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			r, e := sdk.TeamsAdminAPI.TeamsCortanaPolicies.UpdateGlobal(ctx, input)
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
}
