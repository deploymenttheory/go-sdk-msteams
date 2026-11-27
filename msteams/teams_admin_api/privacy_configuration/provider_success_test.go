package privacyconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/privacy_configuration"
)

// These responses are authored DTO/protocol cases, not captured privacy writes.
// The retained privacy fixture is an authorization denial. False/null are bound
// settings; omitted fields and the route identity remain independent of them.
func TestUnitPrivacyPatchTypedSuccessAndMalformedResponse(t *testing.T) {
	t.Parallel()
	const reply = `{"Identity":"prepared-identity","EnablePrivacyMode":false,"AutoInitiateContacts":null,"FutureSetting":{"enabled":true}}`
	for _, tc := range []struct {
		name, response string
		wantError      error
	}{
		{name: "typed-success", response: reply},
		{name: "malformed-success", response: `{"Identity":"prepared-identity","EnablePrivacyMode":"invalid"}`, wantError: client.ErrDecode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{
				Name: tc.name, Source: "msteams/teams_admin_api/privacy_configuration/provider_success_test.go",
				Mappings: []contracttest.Mapping{{Command: "Set-CsPrivacyConfiguration", ParameterSet: "Identity"}},
				Steps:    []contracttest.Exchange{wire("PATCH", "/Skype.Policy/configurations/PrivacyConfiguration/configuration/Global", "Set-CsPrivacyConfiguration", json.RawMessage(`{"Identity":"prepared-identity","EnablePrivacyMode":false,"AutoInitiateContacts":null}`), 200, tc.response)},
				Want:     contracttest.Expected{Status: 200, Body: []byte(tc.response), Error: tc.wantError},
				Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					value, response, err := sdk.TeamsAdminAPI.PrivacyConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{
						Settings:     resource.Settings{EnablePrivacyMode: policy.Set(false), AutoInitiateContacts: policy.Null[bool]()},
						BodyIdentity: policy.Set("prepared-identity"),
					})
					if response == nil || response.StatusCode() != 200 || string(response.Body()) != tc.response {
						t.Fatal("patch lost the received HTTP response")
					}
					if tc.wantError != nil {
						if value != nil || !errors.Is(err, tc.wantError) {
							t.Fatal("malformed success did not remain a decode failure")
						}
					} else {
						if err != nil || value == nil || value.Identity == nil || *value.Identity != "prepared-identity" {
							t.Fatal("successful patch did not return its typed identity")
						}
						if value.EnablePrivacyMode == nil || *value.EnablePrivacyMode || value.AutoInitiateContacts != nil {
							t.Fatal("typed patch response lost explicit false or null")
						}
						if value.PublishLocationDataDefault != nil || value.DisplayPublishedPhotoDefault != nil {
							t.Fatal("typed patch response invented omitted settings")
						}
						if string(value.RawFields["AutoInitiateContacts"]) != "null" || string(value.RawFields["FutureSetting"]) != `{"enabled":true}` {
							t.Fatal("typed patch response lost null or future fields")
						}
					}
					return contracttest.Observation{Value: value, Response: response, Err: err}
				},
			})
		})
	}
}
