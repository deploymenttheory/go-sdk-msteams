package teamsaudioconferencingpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_audio_conferencing_policy"
)

func TestAudioPolicyCreateRetainsSubmissionOnReadBackFailure(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, c := range []struct {
		name       string
		post, read int
		reply      string
	}{
		{"success", 201, 200, `{"Identity":"Global","AllowTollFreeDialin":false,"MeetingInvitePhoneNumbers":[],"Future":false}`},
		{"post-denied", 403, 0, `{"Message":"fixture-private"}`},
		{"read-denied", 201, 403, `{"Message":"fixture-private"}`},
		{"read-malformed", 201, 200, "not-json"},
		{"read-empty", 201, 204, ""},
	} {
		postReply := ""
		if c.post != 201 {
			postReply = c.reply
		}
		steps := []contracttest.Exchange{discovery(), wire("POST", collection, newCommand, json.RawMessage(`{"Identity":"Global","AllowTollFreeDialin":false,"MeetingInvitePhoneNumbers":[]}`), c.post, postReply)}
		if c.read != 0 {
			steps = append(steps, wire("GET", item, getCommand, nil, c.read, c.reply))
		}
		results = append(results, run(t, contracttest.Case{Name: "create-" + c.name, Mappings: []contracttest.Mapping{{Command: newCommand, ParameterSet: "Identity"}}, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsAudioConferencingPolicies.Create(ctx, resource.CreateRequest{Identity: "Global", Settings: resource.Settings{AllowTollFreeDialin: new(false), MeetingInvitePhoneNumbers: resource.ReplacePhoneNumbers()}})
			if c.name == "success" {
				if e != nil || v == nil || v.AllowTollFreeDialin == nil || *v.AllowTollFreeDialin || v.RawFields["Future"] == nil {
					return contracttest.Observation{Err: errors.New("create result changed")}
				}
				return contracttest.Observation{Response: r}
			}
			var partial *resource.ReadBackError
			if c.read != 0 {
				if v == nil || v.Identity != "Global" || !errors.As(e, &partial) || partial.CreationResponse == nil || partial.CreationResponse.StatusCode() != 201 {
					return contracttest.Observation{Err: errors.New("successful submission lost after failed read")}
				}
			} else if e == nil || v != nil || errors.As(e, &partial) {
				return contracttest.Observation{Err: errors.New("failed POST reported as created")}
			}
			return contracttest.Observation{Response: r}
		}}))
	}
	evidence(t, results)
}

func TestAudioPolicyGroupAcceptanceSurvivesOnlyActualPostResponses(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, c := range []struct {
		name   string
		status int
		reply  string
		bad    bool
	}{
		{"accepted-malformed", 202, "not-json", true},
		{"accepted-empty", 202, "", true},
		{"accepted-null", 202, "null", false},
		{"accepted-future", 202, `{"OperationId":"fixture-operation","Future":false}`, false},
		{"completed-empty", 204, "", false},
	} {
		results = append(results, run(t, contracttest.Case{Name: c.name, Mappings: []contracttest.Mapping{{Command: grantCommand, ParameterSet: "GrantToGroup"}}, Steps: []contracttest.Exchange{discovery(), wire("POST", "/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsAudioConferencingPolicy", grantCommand, json.RawMessage(`{"policyName":"Fixture"}`), c.status, c.reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsAudioConferencingPolicies.AssignToGroup(ctx, "fixture/group", "Fixture", nil)
			if v == nil || v.Accepted != (c.status == 202) || c.bad && !errors.Is(e, client.ErrDecode) || !c.bad && e != nil {
				return contracttest.Observation{Err: errors.New("group acceptance/decoding changed")}
			}
			if c.name == "accepted-future" && (v.OperationID != "fixture-operation" || len(v.RawJSON) == 0) {
				return contracttest.Observation{Err: errors.New("group future payload lost")}
			}
			return contracttest.Observation{Response: r}
		}}))
	}
	evidence(t, results)
}

func TestAudioPolicyGrantRankAndSparseGlobalComposition(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
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
		results = append(results, run(t, contracttest.Case{Name: name, Source: wrapperSource, Mappings: []contracttest.Mapping{{Command: grantCommand, ParameterSet: "GrantToGroup"}}, Steps: []contracttest.Exchange{discovery(), wire("POST", "/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsAudioConferencingPolicy", grantCommand, raw, 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsAudioConferencingPolicies.Assign(ctx, resource.Assignment{Group: &resource.GroupGrant{ID: "fixture/group", PolicyName: "Fixture", Rank: new(rank)}})
			return contracttest.Observation{Response: r, Err: e}
		}}))
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
		results = append(results, run(t, contracttest.Case{Name: name, Source: wrapperSource, Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{discovery(), wire("PUT", item, setCommand, json.RawMessage(body), 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			r, e := sdk.TeamsAdminAPI.TeamsAudioConferencingPolicies.UpdateGlobal(ctx, input)
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}
