package teamsipphonepolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_ip_phone_policy"
)

func TestUnitOwnedCRUDUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	const initial = "SDK lab non-default initial description"
	const updated = "SDK lab non-default updated description"
	const signIn = "CommonAreaPhoneSignIn"
	const timeout = int64(30)
	const homeScreen = "Disabled"
	const betterTogether = "Disabled"
	const base = "/Skype.Policy/configurations/TeamsIPPhonePolicy"
	const tagged = base + "/configuration/Tag:fixture-policy"
	create := readResponseFixture(t, "create_201.body")
	before := readResponseFixture(t, "create_readback.json")
	patch := readResponseFixture(t, "update_204.body")
	after := readResponseFixture(t, "update_readback.json")
	list := readResponseFixture(t, "owned_list.json")
	deleted := readResponseFixture(t, "delete_204.body")
	missing := readResponseFixture(t, "get_deleted_404.json")
	steps := []contracttest.Exchange{
		compiledWire("POST", base, "New-CsTeamsIPPhonePolicy", json.RawMessage(`{"AllowHotDesking":false,"Description":"`+initial+`","HotDeskingIdleTimeoutInMinutes":30,"Identity":"fixture-policy","SignInMode":"`+signIn+`"}`), 201, string(create)),
		compiledWire("GET", base+"/configuration/fixture-policy", "Get-CsTeamsIPPhonePolicy", nil, 200, string(before)),
		compiledWire("PUT", tagged, "Set-CsTeamsIPPhonePolicy", json.RawMessage(`{"AllowBetterTogether":"`+betterTogether+`","AllowHomeScreen":"`+homeScreen+`","Description":"`+updated+`","Identity":"Tag:fixture-policy"}`), 204, string(patch)),
		compiledWire("GET", tagged, "Get-CsTeamsIPPhonePolicy", nil, 200, string(after)),
		compiledWire("GET", base, "Get-CsTeamsIPPhonePolicy", nil, 200, string(list)),
		compiledWire("DELETE", tagged, "Remove-CsTeamsIPPhonePolicy", nil, 204, string(deleted)),
		compiledWire("GET", tagged, "Get-CsTeamsIPPhonePolicy", nil, 404, string(missing)),
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
	run(t, contracttest.Case{Name: "owned-crud", Source: "msteams/teams_admin_api/teams_ip_phone_policy/testdata/create_readback.json", Steps: steps,
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsIPPhonePolicies
			response, err := s.CreateByIdentity(ctx, "fixture-policy", resource.Settings{Description: new(initial), AllowHotDesking: new(false), HotDeskingIdleTimeoutInMinutes: new(timeout), SignInMode: new(signIn)})
			if err != nil || response == nil || response.StatusCode() != 201 {
				t.Fatal("create request or acceptance changed")
			}
			value, response, err := s.GetByIdentity(ctx, "fixture-policy")
			if err != nil || response == nil || response.StatusCode() != 200 || value == nil || value.Identity != "Tag:fixture-policy" || value.Description == nil || *value.Description != initial || !createdSettings(value, signIn, timeout) {
				t.Fatal("actual create response was decoded incorrectly")
			}
			response, err = s.UpdateByIdentity(ctx, "Tag:fixture-policy", resource.UpdateRequest{Description: new(updated), AllowHomeScreen: new(homeScreen), AllowBetterTogether: new(betterTogether)})
			if err != nil || response == nil || response.StatusCode() != 204 {
				t.Fatal("sparse update lost its HTTP acceptance")
			}
			value, response, err = s.GetByIdentity(ctx, "Tag:fixture-policy")
			if err != nil || response == nil || response.StatusCode() != 200 || value == nil || value.Description == nil || *value.Description != updated || value.AllowHomeScreen == nil || *value.AllowHomeScreen != homeScreen || value.AllowBetterTogether == nil || *value.AllowBetterTogether != betterTogether || !createdSettings(value, signIn, timeout) || value.RawFields["ConfigMetadata"] == nil {
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

// createdSettings checks the non-default values written at creation, which the
// sparse update must preserve.
func createdSettings(v *resource.Policy, signIn string, timeout int64) bool {
	return v.AllowHotDesking != nil && !*v.AllowHotDesking && v.HotDeskingIdleTimeoutInMinutes != nil && *v.HotDeskingIdleTimeoutInMinutes == timeout && v.SignInMode != nil && *v.SignInMode == signIn
}
