package externalaccesspolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	external "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/external_access_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

const itemPath = "/Skype.Policy/configurations/ExternalAccessPolicy/configuration/Global"

func TestUpdateFailurePreservesPreparationAndNeverReplays(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name                    string
		readStatus, patchStatus int
		patchBody               string
		accepted                bool
	}{
		{"read-denied", 403, 0, "", false}, {"patch-throttled", 200, 429, `{"Message":"fixture-private"}`, false}, {"patch-unavailable", 200, 503, "fixture-private", false}, {"accepted-malformed", 200, 200, "not-json", true},
	} {
		steps := []contracttest.Exchange{wire("GET", itemPath, "Set-CsExternalAccessPolicy", nil, tc.readStatus, `{"Identity":"read-model-identity","AllowedExternalDomains":["a"]}`)}
		if tc.patchStatus != 0 {
			steps = append(steps, wire("PATCH", itemPath, "Set-CsExternalAccessPolicy", json.RawMessage(`{"Identity":"read-model-identity","Description":""}`), tc.patchStatus, tc.patchBody))
		}
		run(t, contracttest.Case{Name: tc.name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.ExternalAccessPolicies.UpdateByIdentity(ctx, "Global", external.UpdateRequest{Fields: external.Fields{Description: policy.Set("")}})
			if e == nil || v == nil || v.ReadResponse == nil || v.PatchInvoked != (tc.patchStatus != 0) || v.HTTPAccepted != tc.accepted {
				return contracttest.Observation{Err: errors.New("partial update preparation lost")}
			}
			if tc.patchStatus != 0 && v.Before == nil {
				return contracttest.Observation{Err: errors.New("successful preparation lost")}
			}
			if tc.accepted && !errors.Is(e, client.ErrDecode) {
				return contracttest.Observation{Err: errors.New("malformed accepted body not classified")}
			}
			if strings.Contains(fmt.Sprint(e), "fixture-private") {
				return contracttest.Observation{Err: errors.New("error exposed payload")}
			}
			return contracttest.Observation{Response: r}
		}})
	}
}

func TestCreateComparisonIsExplicitAndDoesNotCloneGlobal(t *testing.T) {
	t.Parallel()
	for _, status := range []int{200, 403} {
		run(t, contracttest.Case{Name: fmt.Sprint("comparison-", status), Steps: []contracttest.Exchange{
			wire("GET", itemPath, "New-CsExternalAccessPolicy", nil, status, `{"Identity":"Global","Description":"server default","EnableFederationAccess":true}`),
			wire("POST", "/Skype.Policy/configurations/ExternalAccessPolicy", "New-CsExternalAccessPolicy", json.RawMessage(`{"Identity":"Fixture"}`), 201, `{"Identity":"Fixture","Future":false}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.ExternalAccessPolicies.CreateWithGlobalComparison(ctx, external.CreateRequest{Identity: "Fixture"})
			if e == nil && (v == nil || !v.CreateInvoked || !v.HTTPAccepted || v.Created == nil || (v.ComparisonError != nil) != (status == 403)) {
				e = errors.New("comparison/submission observations lost")
			}
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
	run(t, contracttest.Case{Name: "global-no-comparison", Steps: []contracttest.Exchange{wire("POST", "/Skype.Policy/configurations/ExternalAccessPolicy", "New-CsExternalAccessPolicy", json.RawMessage(`{"Identity":"GLOBAL"}`), 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.ExternalAccessPolicies.CreateWithGlobalComparison(ctx, external.CreateRequest{Identity: "GLOBAL"})
		if e == nil && (v.Comparison != nil || v.ComparisonResponse != nil || v.ComparisonError != nil || !v.HTTPAccepted || v.Created != nil) {
			e = errors.New("Global comparison or result invented")
		}
		return contracttest.Observation{Response: r, Err: e}
	}})
}

func TestAssignmentsAndMutationsRejectInvalidInputBeforeTokens(t *testing.T) {
	t.Parallel()
	result := run(t, contracttest.Case{Name: "invalid-scope-and-input", Mappings: []contracttest.Mapping{{Command: "Grant-CsExternalAccessPolicy", ParameterSet: "Identity"}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.ExternalAccessPolicies
		for _, input := range []policy.Assignment{{}, {User: &policy.UserAssignment{}, Tenant: &policy.TenantAssignment{}}, {User: &policy.UserAssignment{Identity: ""}}, {Group: &policy.GroupAssignment{ID: ""}}} {
			if _, _, e := s.Assign(ctx, input); !errors.Is(e, policy.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid scope accepted")}
			}
		}
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nidentity", "scope/..", "/empty-scope", "scope/"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, policy.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid read identity accepted")}
			}
		}
		for _, input := range []external.UpdateRequest{{}, {Controls: policy.Controls{Force: new(bool)}}, {AllowedExternalDomains: &policy.ListEdit{}}} {
			if _, _, e := s.UpdateByIdentity(ctx, "Global", input); !errors.Is(e, policy.ErrInput) {
				return contracttest.Observation{Err: errors.New("empty/invalid update accepted")}
			}
		}
		if _, _, e := s.PatchByIdentity(ctx, "Global", external.PatchRequest{}); !errors.Is(e, policy.ErrInput) {
			return contracttest.Observation{Err: errors.New("empty patch accepted")}
		}
		if _, _, e := s.Create(ctx, external.CreateRequest{}); !errors.Is(e, policy.ErrInput) {
			return contracttest.Observation{Err: errors.New("empty create identity accepted")}
		}
		if _, e := s.DeleteByIdentity(ctx, "", policy.Controls{}); !errors.Is(e, policy.ErrInput) {
			return contracttest.Observation{Err: errors.New("empty delete identity accepted")}
		}
		if _, _, e := s.ListFiltered(ctx, "[broken"); !errors.Is(e, policy.ErrInput) {
			return contracttest.Observation{Err: errors.New("invalid filter accepted")}
		}
		c, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.List(c); !errors.Is(e, context.Canceled) {
			return contracttest.Observation{Err: errors.New("cancellation lost")}
		}
		return contracttest.Observation{}
	}})
	evidence(t, []contracttest.Result{result})
}
