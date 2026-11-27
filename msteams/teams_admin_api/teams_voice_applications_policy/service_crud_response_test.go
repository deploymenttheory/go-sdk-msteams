package teamsvoiceapplicationspolicy_test

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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_voice_applications_policy"
)

func TestUnitOwnedCRUDUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	const base = "/Skype.Policy/configurations/TeamsVoiceApplicationsPolicy"
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
		compiledWire("POST", base, "New-CsTeamsVoiceApplicationsPolicy", json.RawMessage(`{"Identity":"fixture-policy","Description":"SDK lab non-default initial description","AllowAutoAttendantHolidaysChange":true,"AllowCallQueueOptOutChange":true}`), 201, string(create)),
		compiledWire("GET", base+"/configuration/fixture-policy", "Get-CsTeamsVoiceApplicationsPolicy", nil, 200, string(before)),
		compiledWire("GET", tagged, "Set-CsTeamsVoiceApplicationsPolicy", nil, 200, string(before)),
		compiledWire("PATCH", tagged, "Set-CsTeamsVoiceApplicationsPolicy", json.RawMessage(`{"Identity":"Tag:fixture-policy","Description":"SDK lab non-default updated description"}`), 204, string(patch)),
		compiledWire("GET", tagged, "Get-CsTeamsVoiceApplicationsPolicy", nil, 200, string(after)),
		compiledWire("GET", base, "Get-CsTeamsVoiceApplicationsPolicy", nil, 200, string(list)),
		compiledWire("DELETE", tagged, "Remove-CsTeamsVoiceApplicationsPolicy", json.RawMessage(`{"Identity":"Tag:fixture-policy"}`), 204, string(deleted)),
		compiledWire("GET", tagged, "Get-CsTeamsVoiceApplicationsPolicy", nil, 404, string(missing)),
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
			s := sdk.TeamsAdminAPI.TeamsVoiceApplicationsPolicies
			_, response, err := s.CreateByIdentity(ctx, "fixture-policy", resource.CreateRequest{Settings: resource.Settings{Description: policy.Set("SDK lab non-default initial description"), AllowAutoAttendantHolidaysChange: policy.Set(true), AllowCallQueueOptOutChange: policy.Set(true)}})
			if err != nil || response == nil || response.StatusCode() != 201 {
				t.Fatal("create request or acceptance changed")
			}
			value, response, err := s.GetByIdentity(ctx, "fixture-policy")
			if err != nil || response == nil || response.StatusCode() != 200 || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Tag:fixture-policy") || value.Description == nil || *value.Description != "SDK lab non-default initial description" {
				t.Fatal("actual create response was decoded incorrectly")
			}
			if !reflect.DeepEqual(value.RawFields, beforeRaw) {
				t.Fatal("complete actual create snapshot changed")
			}
			changed, response, err := s.UpdateByIdentity(ctx, "Tag:fixture-policy", resource.UpdateRequest{Settings: resource.Settings{Description: policy.Set("SDK lab non-default updated description")}})
			if err != nil || response == nil || response.StatusCode() != 204 || changed == nil || !changed.HTTPAccepted || changed.Before == nil || changed.Before.Description == nil || *changed.Before.Description != "SDK lab non-default initial description" || !changed.PatchInvoked {
				t.Fatal("sparse update lost preparation or HTTP acceptance")
			}
			value, response, err = s.GetByIdentity(ctx, "Tag:fixture-policy")
			if err != nil || response == nil || response.StatusCode() != 200 || value == nil || value.Description == nil || *value.Description != "SDK lab non-default updated description" || value.RawFields["ConfigMetadata"] == nil {
				t.Fatal("actual update response lost the changed or omitted setting")
			}
			if !reflect.DeepEqual(value.RawFields, afterRaw) {
				t.Fatal("complete actual update snapshot changed")
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
