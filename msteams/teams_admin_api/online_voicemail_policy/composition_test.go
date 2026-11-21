package onlinevoicemailpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestUnitVoicemailComposedUpdateRetainsPreparationAndBoundValues(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, tc := range []struct {
		name        string
		read, patch int
		reply       string
		accepted    bool
	}{
		{"success", 200, 204, "", true}, {"read-denied", 403, 0, "", false}, {"patch-throttled", 200, 429, `{"Message":"fixture-private"}`, false}, {"accepted-malformed", 200, 200, "not-json", true},
	} {
		steps := []contracttest.Exchange{wire("GET", item, "Set-CsOnlineVoicemailPolicy", nil, tc.read, `{"Identity":"read-model-identity","Description":"preserved","EnableTranscription":true,"MaximumRecordingLength":"00:01:30","Future":false}`)}
		if tc.patch != 0 {
			steps = append(steps, wire("PATCH", item, "Set-CsOnlineVoicemailPolicy", json.RawMessage(`{"Identity":"read-model-identity","EnableTranscription":false,"MaximumRecordingLength":null}`), tc.patch, tc.reply))
		}
		results = append(results, run(t, contracttest.Case{Name: "update-" + tc.name, Mappings: []contracttest.Mapping{{Command: "Set-CsOnlineVoicemailPolicy", ParameterSet: "Identity"}}, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.OnlineVoicemailPolicies.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{EnableTranscription: policy.Set(false), MaximumRecordingLength: policy.Null[resource.TimeSpan]()}})
			if v == nil || v.ReadResponse == nil || v.PatchInvoked != (tc.patch != 0) || v.HTTPAccepted != tc.accepted {
				return contracttest.Observation{Err: errors.New("voicemail update preparation lost")}
			}
			if tc.name == "success" {
				return contracttest.Observation{Response: r, Err: e}
			}
			if e == nil || tc.accepted && !errors.Is(e, client.ErrDecode) {
				return contracttest.Observation{Err: errors.New("partial update error lost")}
			}
			return contracttest.Observation{Response: r}
		}}))
	}
	evidence(t, results)
}

func TestUnitVoicemailCreateComparisonDoesNotCloneDefaultsOrBlockSubmission(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, status := range []int{200, 403} {
		results = append(results, run(t, contracttest.Case{Name: fmt.Sprint("comparison-", status), Mappings: []contracttest.Mapping{{Command: "New-CsOnlineVoicemailPolicy", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{wire("GET", item, "New-CsOnlineVoicemailPolicy", nil, status, `{"Identity":"Global","MaximumRecordingLength":"00:05:00","EnableTranscription":true}`), wire("POST", collection, "New-CsOnlineVoicemailPolicy", json.RawMessage(`{"Identity":"Fixture","Description":""}`), 201, `{"Identity":"Fixture"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.OnlineVoicemailPolicies.CreateWithGlobalComparison(ctx, resource.CreateRequest{Identity: "Fixture", Settings: resource.Settings{Description: policy.Set("")}})
			if e == nil && (v == nil || !v.CreateInvoked || !v.HTTPAccepted || v.Created == nil || (v.ComparisonError != nil) != (status == 403)) {
				e = errors.New("comparison/submission observations lost")
			}
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}

func TestUnitVoicemailListFilteringAndExactGlobalOrder(t *testing.T) {
	t.Parallel()
	const reply = `[{"Identity":"global"},{"Identity":"Tag:Fixture"},{"Identity":"Global"}]`
	var results []contracttest.Result
	for _, filter := range []bool{false, true} {
		set, name := "Identity", "global-order"
		if filter {
			set, name = "Filter", "filter-local"
		}
		results = append(results, run(t, contracttest.Case{Name: name, Mappings: []contracttest.Mapping{{Command: "Get-CsOnlineVoicemailPolicy", ParameterSet: set}}, Steps: []contracttest.Exchange{wire("GET", collection, "Get-CsOnlineVoicemailPolicy", nil, 200, reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			var values []*resource.Policy
			var e error
			if filter {
				v, r, err := sdk.TeamsAdminAPI.OnlineVoicemailPolicies.ListFiltered(ctx, "Fixture*")
				if err == nil && (len(v) != 1 || v[0].Identity == nil || *v[0].Identity != "Tag:Fixture") {
					err = errors.New("local filtering changed")
				}
				return contracttest.Observation{Response: r, Err: err}
			}
			values, r, e := sdk.TeamsAdminAPI.OnlineVoicemailPolicies.List(ctx)
			if e == nil && (len(values) != 3 || values[0].Identity == nil || *values[0].Identity != "Global" || *values[1].Identity != "global") {
				e = errors.New("exact Global order changed")
			}
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}

func TestUnitVoicemailBindingAndOperationReadComposeWithoutRediscovery(t *testing.T) {
	t.Parallel()
	result := run(t, contracttest.Case{Name: "binding-group-zero-then-operation", Source: generatedSource, Mappings: []contracttest.Mapping{{Command: "Grant-CsOnlineVoicemailPolicy", ParameterSet: "GrantToGroup"}}, Steps: []contracttest.Exchange{discovery(), generatedWire("POST", origin+"/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/OnlineVoicemailPolicy", "Grant-CsOnlineVoicemailPolicy", json.RawMessage(`{"policyName":""}`), 202, `{"OperationId":"fixture/operation"}`), generatedWire("GET", origin+"/Skype.Policy/assignments/operations/fixture%2Foperation", "Get-CsBatchPolicyAssignmentOperation", nil, 200, `{"operationId":"fixture/operation","overallStatus":"Future","completedCount":0,"Future":false}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.OnlineVoicemailPolicies
		zero := int32(0)
		v, _, e := s.AssignWithPowerShellBinding(ctx, resource.Assignment{PolicyName: policy.Null[string](), Group: &resource.GroupAssignment{ID: "fixture/group", Priority: &zero}})
		if e != nil || v == nil || !v.Accepted || v.OperationID == nil {
			return contracttest.Observation{Err: errors.New("group binding submission failed")}
		}
		status, r, e := s.GetAssignmentOperation(ctx, *v.OperationID)
		if e == nil && (status == nil || status.OverallStatus == nil || *status.OverallStatus != "Future" || status.CompletedCount == nil || *status.CompletedCount != 0 || string(status.RawFields["Future"]) != "false") {
			e = errors.New("operation presence/future state lost")
		}
		return contracttest.Observation{Response: r, Err: e}
	}})
	evidence(t, []contracttest.Result{result})
}
