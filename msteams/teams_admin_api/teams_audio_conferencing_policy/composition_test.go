package teamsaudioconferencingpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"strconv"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_audio_conferencing_policy"
)

func TestUnitAudioPolicyCreateDenialNeverReadsBack(t *testing.T) {
	t.Parallel()
	for _, status := range []int{403, 429, 503} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			run(t, testutil.Case{Name: "create-denial", Steps: []testutil.Exchange{discovery(t), wire("POST", collection, newCommand, json.RawMessage(`{"Identity":"Global","AllowTollFreeDialin":false,"MeetingInvitePhoneNumbers":[]}`), status, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsAudioConferencingPolicies.Create(ctx, resource.CreateRequest{Identity: "Global", Settings: resource.Settings{AllowTollFreeDialin: new(false), MeetingInvitePhoneNumbers: resource.ReplacePhoneNumbers()}})
				var apiError *client.APIError
				var readBack *resource.ReadBackError
				if !errors.As(e, &apiError) || apiError.StatusCode != status || v != nil || errors.As(e, &readBack) {
					t.Error("failed POST fabricated creation or attempted a success read-back")
				}
				return testutil.Observation{Response: r}
			}})
		})
	}
}

func TestUnitAudioPolicyGroupGrantDenialsNeverReportAcceptance(t *testing.T) {
	t.Parallel()
	for _, status := range []int{403, 429, 503} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			run(t, testutil.Case{Name: "group-denial", Steps: []testutil.Exchange{discovery(t), wire("POST", "/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsAudioConferencingPolicy", grantCommand, json.RawMessage(`{"policyName":"Fixture"}`), status, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsAudioConferencingPolicies.AssignToGroup(ctx, "fixture/group", "Fixture", nil)
				var apiError *client.APIError
				if !errors.As(e, &apiError) || apiError.StatusCode != status || v != nil && v.Accepted {
					t.Error("denied group grant lost its HTTP failure or invented acceptance")
				}
				return testutil.Observation{Response: r}
			}})
		})
	}
}

func TestUnitAudioPolicyGrantRankAndSparseGlobalCompositionRequestDenials(t *testing.T) {
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
		run(t, testutil.Case{Want: testutil.Expected{Error: auth.ErrForbidden}, Name: name, Steps: []testutil.Exchange{discovery(t), wire("POST", "/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsAudioConferencingPolicy", grantCommand, raw, 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsAudioConferencingPolicies.Assign(ctx, resource.Assignment{Group: &resource.GroupGrant{ID: "fixture/group", PolicyName: "Fixture", Rank: new(rank)}})
			return testutil.Observation{Response: r, Err: e}
		}})
	}
	for _, bound := range []bool{false, true} {
		name := "identity-only"
		body := `{"Identity":"Global"}`
		input := resource.UpdateRequest{}
		if bound {
			name = "false-and-empty-operators"
			body = `{"Identity":"Global","AllowTollFreeDialin":false,"MeetingInvitePhoneNumbers":{}}`
			input.AllowTollFreeDialin = new(false)
			input.MeetingInvitePhoneNumbers = resource.ChangePhoneNumbers(nil, nil)
		}
		run(t, testutil.Case{Want: testutil.Expected{Error: auth.ErrForbidden}, Name: name, Steps: []testutil.Exchange{discovery(t), wire("PUT", item, setCommand, json.RawMessage(body), 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			r, e := sdk.TeamsAdminAPI.TeamsAudioConferencingPolicies.UpdateGlobal(ctx, input)
			return testutil.Observation{Response: r, Err: e}
		}})
	}
}
