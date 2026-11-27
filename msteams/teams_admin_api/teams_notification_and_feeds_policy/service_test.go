package teamsnotificationandfeedspolicy_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_notification_and_feeds_policy"
)

const collection = "/Skype.Policy/configurations/TeamsNotificationAndFeedsPolicy"
const item = collection + "/configuration/Global"
const selected = collection + "/configuration/Tag%3Afixture-policy"

func TestUnitExplicitNullUsesItsAcceptanceResponse(t *testing.T) {
	t.Parallel()
	read := readResponseFixture(t, "clear_description_readback.json")
	run(t, testutil.Case{Name: "clear-description", Steps: []testutil.Exchange{
		compiledWire("PATCH", selected, "Set-CsTeamsNotificationAndFeedsPolicy", json.RawMessage(`{"Identity":"Tag:fixture-policy","Description":null}`), 204, string(readResponseFixture(t, "clear_description_204.body"))),
		compiledWire("GET", selected, "Get-CsTeamsNotificationAndFeedsPolicy", nil, 200, string(read)),
	}, Want: testutil.Expected{JSON: read}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.TeamsNotificationAndFeedsPolicies
		if _, _, e := s.PatchByIdentity(ctx, "Tag:fixture-policy", resource.PatchRequest{Settings: resource.Settings{Description: policy.Null[string]()}}); e != nil {
			t.Fatal("explicit null was not sent")
		}
		v, r, e := s.GetByIdentity(ctx, "Tag:fixture-policy")
		if e == nil && (v == nil || v.Description != nil || v.SuggestedFeedsEnabledType == nil || *v.SuggestedFeedsEnabledType != "EnabledUserOverride") {
			t.Fatal("actual cleared response lost null or the omitted setting")
		}
		return testutil.Observation{Value: v, Response: r, Err: e}
	}})
}

func TestUnitFiltersUseTheAcceptanceCollection(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, pattern string
		count         int
	}{{"global", "[gG]lobal", 1}, {"tagged", "fixture-polic?", 1}, {"missing", "missing", 0}, {"literal-whitespace", " ", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: tc.name, Steps: []testutil.Exchange{compiledWire("GET", collection, "Get-CsTeamsNotificationAndFeedsPolicy", nil, 200, string(readResponseFixture(t, "owned_list.json")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsNotificationAndFeedsPolicies.ListFiltered(ctx, tc.pattern)
				if e == nil && len(v) != tc.count {
					t.Fatal("Identity/Tag filter selection differs")
				}
				return testutil.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitMutationsKeepErrorsAndNeverReplay(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"create", "patch", "delete"} {
		for _, status := range []int{401, 403, 429, 503} {
			method, command, path, body := "POST", "New-CsTeamsNotificationAndFeedsPolicy", collection, json.RawMessage(`{"Identity":"Global"}`)
			switch operation {
			case "patch":
				method, command, path, body = "PATCH", "Set-CsTeamsNotificationAndFeedsPolicy", item, json.RawMessage(`{"Identity":"Global","Description":null}`)
			case "delete":
				method, command, path = "DELETE", "Remove-CsTeamsNotificationAndFeedsPolicy", item
			}
			run(t, testutil.Case{Name: operation, Steps: []testutil.Exchange{compiledWire(method, path, command, body, status, "")}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				s := sdk.TeamsAdminAPI.TeamsNotificationAndFeedsPolicies
				switch operation {
				case "create":
					_, r, e := s.CreateByIdentity(ctx, "Global", resource.CreateRequest{})
					return testutil.Observation{Response: r, Err: e}
				case "patch":
					_, r, e := s.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{Description: policy.Null[string]()}})
					return testutil.Observation{Response: r, Err: e}
				default:
					_, r, e := s.DeleteByIdentity(ctx, "Global", policy.Controls{})
					return testutil.Observation{Response: r, Err: e}
				}
			}})
		}
	}
}
