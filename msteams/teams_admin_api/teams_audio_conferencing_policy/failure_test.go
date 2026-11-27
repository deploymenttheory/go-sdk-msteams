package teamsaudioconferencingpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_audio_conferencing_policy"
)

func TestUnitAudioPolicyMutationErrorsNeverReplay(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{"create", "update", "delete", "user", "tenant", "group"} {
		for _, status := range []int{400, 401, 403, 429, 503} {
			method, command, path, body := "POST", newCommand, collection, json.RawMessage(`{"Identity":"Global"}`)
			switch operation {
			case "update":
				method, command, path = "PUT", setCommand, item
			case "delete":
				method, command, path, body = "DELETE", deleteCommand, item, nil
			case "user":
				method, command, path, body = "PATCH", grantCommand, "/Skype.Policy/users/fixture%2Fuser/policies/TeamsAudioConferencingPolicy", json.RawMessage(`{"policyName":""}`)
			case "tenant":
				method, command, path, body = "PATCH", grantCommand, "/Skype.Policy/tenants/policies/TeamsAudioConferencingPolicy", json.RawMessage(`{"policyName":"","forceSwitchPresent":false}`)
			case "group":
				method, command, path, body = "POST", grantCommand, "/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsAudioConferencingPolicy", json.RawMessage(`{"policyName":""}`)
			}
			run(t, testutil.Case{Name: fmt.Sprintf("%s-error-%d", operation, status), Steps: []testutil.Exchange{discovery(t), wire(method, path, command, body, status, `{"Message":"fixture-private"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				s := sdk.TeamsAdminAPI.TeamsAudioConferencingPolicies
				var o testutil.Observation
				switch operation {
				case "create":
					r, e := s.CreateByIdentity(ctx, "Global", resource.Settings{})
					o = testutil.Observation{Response: r, Err: e}
				case "update":
					r, e := s.UpdateByIdentity(ctx, "Global", resource.Settings{})
					o = testutil.Observation{Response: r, Err: e}
				case "delete":
					r, e := s.DeleteByIdentity(ctx, "Global")
					o = testutil.Observation{Response: r, Err: e}
				case "user":
					r, e := s.AssignToUser(ctx, "fixture/user", "")
					o = testutil.Observation{Response: r, Err: e}
				case "tenant":
					r, e := s.AssignTenantDefault(ctx, "", false)
					o = testutil.Observation{Response: r, Err: e}
				case "group":
					v, r, e := s.AssignToGroup(ctx, "fixture/group", "", nil)
					if v != nil && v.Accepted {
						return testutil.Observation{Err: errors.New("failed group POST reported accepted")}
					}
					o = testutil.Observation{Response: r, Err: e}
				}
				var api *client.APIError
				if !errors.As(o.Err, &api) || api.StatusCode != status || strings.Contains(o.Err.Error(), "fixture-private") {
					return testutil.Observation{Err: errors.New("mutation status/redaction changed")}
				}
				o.Err = nil
				return o
			}})
		}
	}
}

func TestUnitAudioPolicyInputsAndBuildersHaveNoHTTP(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "local-input-context-and-builders", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.TeamsAudioConferencingPolicies
		for _, id := range []string{"", " ", ".", "..", "bad\\identity", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid identity accepted")}
			}
			if _, e := s.CreateByIdentity(ctx, id, resource.Settings{}); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid create identity accepted")}
			}
		}
		if _, _, e := s.ListFiltered(ctx, "[broken"); !errors.Is(e, resource.ErrInput) {
			return testutil.Observation{Err: errors.New("invalid wildcard accepted")}
		}
		if _, e := s.UpdateGlobal(ctx, resource.Settings{MeetingInvitePhoneNumbers: new(resource.ListChange)}); !errors.Is(e, resource.ErrInput) {
			return testutil.Observation{Err: errors.New("zero list change acquired credentials")}
		}
		for _, input := range []resource.Assignment{{}, {User: &resource.UserAssignment{Identity: "fixture"}, Tenant: &resource.TenantAssignment{}}} {
			if _, _, e := s.Assign(ctx, input); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid assignment scope acquired credentials")}
			}
		}
		if _, e := s.AssignToUser(ctx, "", ""); !errors.Is(e, resource.ErrInput) {
			return testutil.Observation{Err: errors.New("empty user accepted")}
		}
		if _, _, e := s.AssignToGroup(ctx, "", "", nil); !errors.Is(e, resource.ErrInput) {
			return testutil.Observation{Err: errors.New("empty group accepted")}
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, e := s.UpdateGlobal(canceled, resource.Settings{}); !errors.Is(e, context.Canceled) {
			return testutil.Observation{Err: errors.New("cancellation lost")}
		}
		for _, missing := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := missing.List(ctx); !errors.Is(e, client.ErrClosed) {
				return testutil.Observation{Err: errors.New("missing backend accepted")}
			}
		}
		if _, e := json.Marshal(resource.ChangePhoneNumbers(nil, nil)); e != nil {
			return testutil.Observation{Err: e}
		}
		if _, e := resource.ReplacePhoneValues([]json.RawMessage{json.RawMessage(`null`)}); !errors.Is(e, resource.ErrInput) {
			return testutil.Observation{Err: errors.New("null list item accepted")}
		}
		return testutil.Observation{}
	}})
}
