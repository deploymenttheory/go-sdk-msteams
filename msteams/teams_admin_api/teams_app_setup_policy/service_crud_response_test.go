package teamsappsetuppolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_setup_policy"
)

func TestUnitOwnedCRUDUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	const initial = "SDK lab non-default initial description"
	const updated = "SDK lab non-default updated description"
	const base = "/Skype.Policy/configurations/TeamsAppSetupPolicy"
	const tagged = base + "/configuration/Tag%3Afixture-policy"
	create := readResponseFixture(t, "create_201.body")
	before := readResponseFixture(t, "create_readback.json")
	patch := readResponseFixture(t, "update_204.body")
	after := readResponseFixture(t, "update_readback.json")
	list := readResponseFixture(t, "owned_list.json")
	deleted := readResponseFixture(t, "delete_204.body")
	missing := readResponseFixture(t, "get_deleted_404.json")
	var beforeRaw, afterRaw map[string]json.RawMessage
	if json.Unmarshal(before, &beforeRaw) != nil || json.Unmarshal(after, &afterRaw) != nil {
		t.Fatal("invalid actual policy snapshots")
	}
	steps := []testutil.Exchange{
		wire("POST", base, "New-CsTeamsAppSetupPolicy", json.RawMessage(`{"Identity":"fixture-policy","Description":"`+initial+`","AllowUserPinning":false}`), 201, string(create)),
		wire("GET", base+"/configuration/fixture-policy", "Get-CsTeamsAppSetupPolicy", nil, 200, string(before)),
		wire("GET", tagged, "Set-CsTeamsAppSetupPolicy", nil, 200, string(before)),
		wire("PATCH", tagged, "Set-CsTeamsAppSetupPolicy", json.RawMessage(`{"Identity":"Tag:fixture-policy","Description":"`+updated+`"}`), 204, string(patch)),
		wire("GET", tagged, "Get-CsTeamsAppSetupPolicy", nil, 200, string(after)),
		wire("GET", base, "Get-CsTeamsAppSetupPolicy", nil, 200, string(list)),
		wire("DELETE", tagged, "Remove-CsTeamsAppSetupPolicy", json.RawMessage(`{"Identity":"Tag:fixture-policy"}`), 204, string(deleted)),
		wire("GET", tagged, "Get-CsTeamsAppSetupPolicy", nil, 404, string(missing)),
	}
	for i := range steps {
		steps[i].Reply.Headers.Set("Content-Type", "application/json")
	}
	var envelope struct {
		Code      string `json:"code"`
		ErrorCode int    `json:"errorCode"`
	}
	if json.Unmarshal(missing, &envelope) != nil {
		t.Fatal("invalid error fixture")
	}
	run(t, testutil.Case{Name: "owned-crud", Steps: steps,
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.TeamsAppSetupPolicies
			_, response, err := s.CreateByIdentity(ctx, "fixture-policy", resource.CreateRequest{ScalarSettings: resource.ScalarSettings{Description: policy.Set(initial), AllowUserPinning: policy.Set(false)}})
			if err != nil || response == nil || response.StatusCode() != 201 {
				t.Fatal("create request or acceptance changed")
			}
			value, response, err := s.GetByIdentity(ctx, "fixture-policy")
			if err != nil || response == nil || response.StatusCode() != 200 || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Tag:fixture-policy") || value.Description == nil || *value.Description != initial || value.AllowUserPinning == nil || *value.AllowUserPinning {
				t.Fatal("actual create response was decoded incorrectly")
			}
			if !reflect.DeepEqual(value.RawFields, beforeRaw) {
				t.Fatal("complete actual create snapshot changed")
			}
			changed, response, err := s.UpdateByIdentity(ctx, "Tag:fixture-policy", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{Description: policy.Set(updated)}})
			if err != nil || response == nil || response.StatusCode() != 204 || changed == nil || !changed.HTTPAccepted || changed.Before == nil || changed.Before.Description == nil || *changed.Before.Description != initial || !changed.PatchInvoked {
				t.Fatal("sparse update lost preparation or HTTP acceptance")
			}
			value, response, err = s.GetByIdentity(ctx, "Tag:fixture-policy")
			if err != nil || response == nil || response.StatusCode() != 200 || value == nil || value.Description == nil || *value.Description != updated || value.AllowUserPinning == nil || *value.AllowUserPinning || value.RawFields["ConfigMetadata"] == nil {
				t.Fatal("actual update response lost the changed or omitted setting")
			}
			if !reflect.DeepEqual(value.RawFields, afterRaw) || !reflect.DeepEqual(value.AppPresetList, changed.Before.AppPresetList) || !reflect.DeepEqual(value.PinnedAppBarApps, changed.Before.PinnedAppBarApps) || !reflect.DeepEqual(value.PinnedMessageBarApps, changed.Before.PinnedMessageBarApps) || !reflect.DeepEqual(value.AppPresetMeetingList, changed.Before.AppPresetMeetingList) || !reflect.DeepEqual(value.AdditionalCustomizationApps, changed.Before.AdditionalCustomizationApps) || !reflect.DeepEqual(value.PinnedCallingBarApps, changed.Before.PinnedCallingBarApps) {
				t.Fatal("complete actual snapshot or omitted app collection changed")
			}
			values, response, err := s.List(ctx)
			if err != nil || response == nil || response.StatusCode() != 200 {
				t.Fatal("actual owned list response was not decoded")
			}
			found := 0
			for _, item := range values {
				if item != nil && item.Identity != nil && *item.Identity == "Tag:fixture-policy" {
					found++
				}
			}
			if found != 1 {
				t.Fatal("owned policy missing or duplicated")
			}
			_, deleted, err := s.DeleteByIdentity(ctx, "Tag:fixture-policy", policy.Controls{})
			if err != nil || deleted.StatusCode() != 204 {
				t.Fatal("actual delete response changed")
			}
			_, response, err = s.GetByIdentity(ctx, "Tag:fixture-policy")
			var api *client.APIError
			if !errors.As(err, &api) || api.StatusCode != 404 {
				t.Fatal("actual absence response lost its typed error")
			}
			return testutil.Observation{Response: response, Err: err}
		}, Want: testutil.Expected{Status: 404, Body: missing, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 404, Codes: []string{envelope.Code}, ServiceErrorCodes: []int{envelope.ErrorCode}}}})
}
