package policyassignments_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

func TestUnitAssignmentInputAndContextBeforeCredentials(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "input-context-before-tokens", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.PolicyAssignments
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nid"} {
			if _, _, e := s.GetGroup(ctx, id, "Fixture"); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid group identity accepted")}
			}
			if _, _, e := s.CreateGroup(ctx, "fixture-group", id, resource.GroupSettings{PolicyName: policy.Set("Fixture")}); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid policy type accepted")}
			}
			if _, _, e := s.GetBatch(ctx, id); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid operation identity accepted")}
			}
			if _, _, e := s.GetUserByIdentity(ctx, resource.UserIdentity{Identity: id}, ""); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid InputObject identity accepted")}
			}
		}
		if _, _, e := s.CreateBatch(ctx, resource.BatchCreateRequest{}); !errors.Is(e, resource.ErrInput) {
			return testutil.Observation{Err: errors.New("missing required identity collection accepted")}
		}
		invalid := resource.BatchCreateRequest{Identities: []string{}, AdditionalParameters: map[string]json.RawMessage{"fixture-private": []byte("not-json")}}
		if _, _, e := s.CreateBatch(ctx, invalid); !errors.Is(e, resource.ErrInput) || strings.Contains(fmt.Sprint(e), "fixture-private") {
			return testutil.Observation{Err: errors.New("malformed additional parameters acquired credentials or leaked input")}
		}
		graph, e := client.NewScope(client.Graph, sdk.TeamsAdminAPI.Scope().Session())
		if e != nil {
			return testutil.Observation{Err: e}
		}
		if _, _, e := resource.NewService(graph).ListBatches(ctx, ""); !errors.Is(e, client.ErrDestination) {
			return testutil.Observation{Err: errors.New("wrong backend accepted")}
		}
		for _, missing := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := missing.ListBatches(ctx, ""); !errors.Is(e, client.ErrClosed) {
				return testutil.Observation{Err: errors.New("missing service accepted")}
			}
		}
		//nolint:contextcheck,staticcheck // Deliberately verify the public nil-context guard before credential acquisition.
		if _, _, e := s.ListBatches(nil, ""); !errors.Is(e, client.ErrNilContext) {
			return testutil.Observation{Err: errors.New("nil context accepted")}
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.CreateGroup(canceled, "fixture-group", "Fixture", resource.GroupSettings{}); !errors.Is(e, context.Canceled) {
			return testutil.Observation{Err: errors.New("cancellation lost")}
		}
		return testutil.Observation{}
	}})
}

func TestUnitAssignmentWritesKeepDenialAndNeverReplay(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{"create", "update", "delete", "batch"} {
		for _, status := range []int{401, 403, 429, 500, 503} {
			method, command, path, body := "PUT", "New-CsGroupPolicyAssignment", "/Skype.Policy/groupPolicyAssignments/fixture-group/policyTypes/Fixture", json.RawMessage(`{}`)
			switch operation {
			case "create":
				body = json.RawMessage(`{"policyName":"Fixture"}`)
			case "update":
				method, command = "PATCH", "Set-CsGroupPolicyAssignment"
			case "delete":
				method, command, body = "DELETE", "Remove-CsGroupPolicyAssignment", nil
			case "batch":
				method, command, path, body = "POST", "New-CsBatchPolicyAssignmentOperation", "/Skype.Policy/assignments/operations", json.RawMessage(`{"identities":[]}`)
			}
			c := testutil.Case{Name: fmt.Sprintf("%s-denial-%d", operation, status), Steps: []testutil.Exchange{discovery(t), wire(method, origin+path, command, body, status, `{"code":"FixtureDenied","message":"fixture-private"}`)}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status, Codes: []string{"FixtureDenied"}}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				s := sdk.TeamsAdminAPI.PolicyAssignments
				var o testutil.Observation
				switch operation {
				case "create":
					o.Value, o.Response, o.Err = s.CreateGroup(ctx, "fixture-group", "Fixture", resource.GroupSettings{PolicyName: policy.Set("Fixture")})
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
			run(t, c)
		}
	}
}

func TestUnitAssignmentLocalOperationHandlePresence(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		want       *string
		invalid    bool
	}{
		{"future-id", `{"OperationId":"opaque/future:handle","future":false}`, new("opaque/future:handle"), false},
		{"missing-id", `{}`, nil, false}, {"null-id", `{"OperationId":null}`, nil, false}, {"invalid-id", `{"OperationId":3}`, nil, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var v resource.OperationResult
			err := json.Unmarshal([]byte(tc.body), &v)
			if tc.invalid {
				if err == nil {
					t.Fatal("invalid local handle accepted")
				}
				return
			}
			if err != nil || (tc.want == nil && v.OperationID != nil) || (tc.want != nil && (v.OperationID == nil || *v.OperationID != *tc.want)) {
				t.Fatal("local handle presence differs")
			}
		})
	}
}

func TestUnitAssignmentInputObjectSharesOpaqueUserRoute(t *testing.T) {
	t.Parallel()
	c := testutil.Case{Want: testutil.Expected{Error: auth.ErrForbidden}, Name: "user-input-object", Steps: []testutil.Exchange{discovery(t), wire("GET", origin+"/Skype.User/users/fixture%2Fuser%3A%40/policies", "Get-CsUserPolicyAssignment", nil, 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.PolicyAssignments.GetUserByIdentity(ctx, resource.UserIdentity{Identity: "fixture/user:@"}, "")
		// This case proves the opaque InputObject route and denial only.
		// Actual source provenance is asserted by TestUnitUserReadsUseAcceptanceResponses.
		return testutil.Observation{Value: v, Response: r, Err: e}
	}}
	run(t, c)
}

func TestUnitAssignmentSettingsOmitNullAndRetainZero(t *testing.T) {
	t.Parallel()
	settings := resource.GroupSettings{PolicyName: policy.Null[string](), Priority: policy.Set(int32(0))}
	raw, e := json.Marshal(settings)
	if e != nil || string(raw) != `{"priority":0}` {
		t.Fatal("generated DTO null/zero semantics changed")
	}
}
