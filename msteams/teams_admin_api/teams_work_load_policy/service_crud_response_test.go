package teamsworkloadpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_work_load_policy"
)

func TestUnitOwnedCRUDUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	const initial = "SDK lab non-default initial description"
	const updated = "SDK lab non-default updated description"
	const base = "/Skype.Policy/configurations/TeamsWorkLoadPolicy"
	const tagged = base + "/configuration/Tag:fixture-policy"
	create := readResponseFixture(t, "create_201.body")
	before := readResponseFixture(t, "create_readback.json")
	patch := readResponseFixture(t, "update_204.body")
	after := readResponseFixture(t, "update_readback.json")
	list := readResponseFixture(t, "owned_list.json")
	deleted := readResponseFixture(t, "delete_204.body")
	missing := readResponseFixture(t, "get_deleted_404.json")
	steps := []contracttest.Exchange{
		compiledWire("POST", base, "New-CsTeamsWorkLoadPolicy", json.RawMessage(`{"Identity":"fixture-policy","Description":"`+initial+`","AllowCalling":false,"AllowCallingPinned":false,"AllowMeeting":false,"AllowMeetingPinned":false,"AllowMessaging":false,"AllowMessagingPinned":false}`), 201, string(create)),
		compiledWire("GET", base+"/configuration/fixture-policy", "Get-CsTeamsWorkLoadPolicy", nil, 200, string(before)),
		compiledWire("PUT", tagged, "Set-CsTeamsWorkLoadPolicy", json.RawMessage(`{"Identity":"Tag:fixture-policy","Description":"`+updated+`"}`), 204, string(patch)),
		compiledWire("GET", tagged, "Get-CsTeamsWorkLoadPolicy", nil, 200, string(after)),
		compiledWire("GET", base, "Get-CsTeamsWorkLoadPolicy", nil, 200, string(list)),
		compiledWire("DELETE", tagged, "Remove-CsTeamsWorkLoadPolicy", nil, 204, string(deleted)),
		compiledWire("GET", tagged, "Get-CsTeamsWorkLoadPolicy", nil, 404, string(missing)),
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
	run(t, contracttest.Case{Name: "owned-crud", Source: "msteams/teams_admin_api/teams_work_load_policy/testdata/create_readback.json", Steps: steps,
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsWorkLoadPolicies
			response, err := s.CreateByIdentity(ctx, "fixture-policy", resource.Settings{Description: new(initial), AllowCalling: new(false), AllowCallingPinned: new(false), AllowMeeting: new(false), AllowMeetingPinned: new(false), AllowMessaging: new(false), AllowMessagingPinned: new(false)})
			if err != nil || response == nil || response.StatusCode() != 201 {
				t.Fatal("create request or acceptance changed")
			}
			value, response, err := s.GetByIdentity(ctx, "fixture-policy")
			if err != nil || response == nil || response.StatusCode() != 200 || value == nil || value.Identity != "Tag:fixture-policy" || value.Description == nil || *value.Description != initial || !allDisabled(value) {
				t.Fatal("actual create response was decoded incorrectly")
			}
			response, err = s.UpdateByIdentity(ctx, "Tag:fixture-policy", resource.UpdateRequest{Description: new(updated)})
			if err != nil || response == nil || response.StatusCode() != 204 {
				t.Fatal("sparse update lost its HTTP acceptance")
			}
			value, response, err = s.GetByIdentity(ctx, "Tag:fixture-policy")
			if err != nil || response == nil || response.StatusCode() != 200 || value == nil || value.Description == nil || *value.Description != updated || !allDisabled(value) || value.RawFields["ConfigMetadata"] == nil {
				t.Fatal("actual update response lost the changed or omitted setting")
			}
			values, response, err := s.List(ctx)
			if err != nil || response == nil || response.StatusCode() != 200 {
				t.Fatal("actual owned list response was not decoded")
			}
			found := 0
			for _, item := range values {
				if item.Identity == "Tag:fixture-policy" {
					found++
				}
			}
			if found != 1 {
				t.Fatal("owned policy missing or duplicated")
			}
			deleted, err := s.DeleteByIdentity(ctx, "Tag:fixture-policy")
			if err != nil || deleted.StatusCode() != 204 {
				t.Fatal("actual delete response changed")
			}
			_, response, err = s.GetByIdentity(ctx, "Tag:fixture-policy")
			var api *client.APIError
			if !errors.As(err, &api) || api.StatusCode != 404 {
				t.Fatal("actual absence response lost its typed error")
			}
			return contracttest.Observation{Response: response, Err: err}
		}, Want: contracttest.Expected{Status: 404, Body: missing, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 404, Codes: []string{envelope.Code}, ServiceErrorCodes: []int{envelope.ErrorCode}}}})
}

func allDisabled(v *resource.Policy) bool {
	for _, flag := range []*bool{v.AllowCalling, v.AllowCallingPinned, v.AllowMeeting, v.AllowMeetingPinned, v.AllowMessaging, v.AllowMessagingPinned} {
		if flag == nil || *flag {
			return false
		}
	}
	return true
}
