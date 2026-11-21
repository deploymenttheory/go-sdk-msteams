package policyassignments_test

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
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

func TestUnitAssignmentInputAndContextBeforeCredentials(t *testing.T) {
	t.Parallel()
	r := run(t, contracttest.Case{Name: "input-context-before-tokens", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.PolicyAssignments
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nid"} {
			if _, _, e := s.GetGroup(ctx, id, "Fixture"); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid group identity accepted")}
			}
			if _, _, e := s.CreateGroup(ctx, "fixture-group", id, resource.GroupSettings{}); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid policy type accepted")}
			}
			if _, _, e := s.GetBatch(ctx, id); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid operation identity accepted")}
			}
			if _, _, e := s.GetUserByIdentity(ctx, resource.UserIdentity{Identity: id}, ""); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid InputObject identity accepted")}
			}
		}
		if _, _, e := s.CreateBatch(ctx, resource.BatchCreateRequest{}); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("missing required identity collection accepted")}
		}
		invalid := resource.BatchCreateRequest{Identities: []string{}, AdditionalParameters: map[string]json.RawMessage{"fixture-private": []byte("not-json")}}
		if _, _, e := s.CreateBatch(ctx, invalid); !errors.Is(e, resource.ErrInput) || strings.Contains(fmt.Sprint(e), "fixture-private") {
			return contracttest.Observation{Err: errors.New("malformed additional parameters acquired credentials or leaked input")}
		}
		graph, e := client.NewScope(client.Graph, sdk.TeamsAdminAPI.Scope().Session())
		if e != nil {
			return contracttest.Observation{Err: e}
		}
		if _, _, e := resource.NewService(graph).ListBatches(ctx, ""); !errors.Is(e, client.ErrDestination) {
			return contracttest.Observation{Err: errors.New("wrong backend accepted")}
		}
		for _, missing := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := missing.ListBatches(ctx, ""); !errors.Is(e, client.ErrClosed) {
				return contracttest.Observation{Err: errors.New("missing service accepted")}
			}
		}
		//nolint:contextcheck,staticcheck // Deliberately verify the public nil-context guard before credential acquisition.
		if _, _, e := s.ListBatches(nil, ""); !errors.Is(e, client.ErrNilContext) {
			return contracttest.Observation{Err: errors.New("nil context accepted")}
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.CreateGroup(canceled, "fixture-group", "Fixture", resource.GroupSettings{}); !errors.Is(e, context.Canceled) {
			return contracttest.Observation{Err: errors.New("cancellation lost")}
		}
		return contracttest.Observation{}
	}})
	evidence(t, []contracttest.Result{r})
}

func TestUnitAssignmentWritesKeepDenialAndNeverReplay(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, operation := range []string{"create", "update", "delete", "batch"} {
		for _, status := range []int{401, 403, 429, 500, 503} {
			method, command, path, body := "PUT", "New-CsGroupPolicyAssignment", "/Skype.Policy/groupPolicyAssignments/fixture-group/policyTypes/Fixture", json.RawMessage(`{}`)
			switch operation {
			case "update":
				method, command = "PATCH", "Set-CsGroupPolicyAssignment"
			case "delete":
				method, command, body = "DELETE", "Remove-CsGroupPolicyAssignment", nil
			case "batch":
				method, command, path, body = "POST", "New-CsBatchPolicyAssignmentOperation", "/Skype.Policy/assignments/operations", json.RawMessage(`{"identities":[]}`)
			}
			c := contracttest.Case{Name: fmt.Sprintf("%s-denial-%d", operation, status), Steps: []contracttest.Exchange{discovery(), wire(method, origin+path, command, body, status, `{"code":"FixtureDenied","message":"fixture-private"}`)}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status, Codes: []string{"FixtureDenied"}}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.PolicyAssignments
				var o contracttest.Observation
				switch operation {
				case "create":
					o.Value, o.Response, o.Err = s.CreateGroup(ctx, "fixture-group", "Fixture", resource.GroupSettings{})
				case "update":
					o.Value, o.Response, o.Err = s.UpdateGroup(ctx, "fixture-group", "Fixture", resource.GroupSettings{})
				case "delete":
					o.Value, o.Response, o.Err = s.DeleteGroup(ctx, "fixture-group", "Fixture")
				case "batch":
					o.Value, o.Response, o.Err = s.CreateBatch(ctx, resource.BatchCreateRequest{Identities: []string{}})
				}
				if o.Err == nil || strings.Contains(fmt.Sprintf("%v %#v", o.Err, o.Err), "fixture-private") {
					o.Err = errors.New("denial lost or leaked response message")
				}
				return o
			}}
			results = append(results, run(t, c))
		}
	}
	evidence(t, results)
}

func TestUnitAssignmentSubmissionRetainsAcceptanceAndHandlePresence(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, row := range []struct {
		name, reply string
		status      int
		wantID      *string
		decode      bool
	}{
		{"future-id", `{"OperationId":"opaque/future:handle","future":false}`, 202, new("opaque/future:handle"), false},
		{"missing-id", `{}`, 202, nil, false},
		{"null-id", `{"OperationId":null}`, 202, nil, false},
		{"invalid-id", `{"OperationId":3}`, 202, nil, true},
		{"malformed", `not-json`, 202, nil, true},
	} {
		c := contracttest.Case{Name: "submission-" + row.name, Mappings: mapping("batch-create"), Steps: []contracttest.Exchange{discovery(), wire("POST", origin+"/Skype.Policy/assignments/operations", "New-CsBatchPolicyAssignmentOperation", json.RawMessage(`{"identities":[]}`), row.status, row.reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.CreateBatch(ctx, resource.BatchCreateRequest{Identities: []string{}})
			if v == nil || !v.HTTPAccepted || r == nil || (row.wantID == nil && v.OperationID != nil) || (row.wantID != nil && (v.OperationID == nil || *v.OperationID != *row.wantID)) {
				e = errors.New("submission acceptance or optional handle changed")
			}
			return contracttest.Observation{Response: r, Err: e}
		}}
		if row.decode {
			c.Want.Error = client.ErrDecode
		}
		results = append(results, run(t, c))
	}
	evidence(t, results)
}

func TestUnitAssignmentInputObjectSharesOpaqueUserRoute(t *testing.T) {
	t.Parallel()
	c := contracttest.Case{Name: "user-input-object", Mappings: []contracttest.Mapping{{Command: "Get-CsUserPolicyAssignment", ParameterSet: "GetViaIdentity"}}, Steps: []contracttest.Exchange{discovery(), wire("GET", origin+"/Skype.User/users/fixture%2Fuser%3A%40/policies", "Get-CsUserPolicyAssignment", nil, 200, `[{"policyType":"Fixture","policyName":"","policySource":[{"assignmentType":"Direct","reference":"fixture-direct"},{"assignmentType":"Group","reference":"fixture-group"}]}]`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.GetUserByIdentity(ctx, resource.UserIdentity{Identity: "fixture/user:@"}, "")
		if e == nil && (len(v) != 1 || len(v[0].PolicySource) != 2 || v[0].PolicySource[1].Reference == nil || *v[0].PolicySource[1].Reference != "fixture-group") {
			e = errors.New("effective policy source provenance lost")
		}
		return contracttest.Observation{Response: r, Err: e}
	}}
	evidence(t, []contracttest.Result{run(t, c)})
}

func TestUnitAssignmentSettingsOmitNullAndRetainZero(t *testing.T) {
	t.Parallel()
	settings := resource.GroupSettings{PolicyName: policy.Null[string](), Priority: policy.Set(int32(0))}
	raw, e := json.Marshal(settings)
	if e != nil || string(raw) != `{"priority":0}` {
		t.Fatal("generated DTO null/zero semantics changed")
	}
}
