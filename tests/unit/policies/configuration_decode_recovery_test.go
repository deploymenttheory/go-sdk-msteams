package policies_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	validation "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_validation_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	clientconfig "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_client_configuration"
)

// The invalid HTTP URI is an adversarial model reply against the pinned HTTPS
// setter, not captured server data. Earlier accepted configuration state must
// survive a later200 decode failure without a PATCH or guessed compensation.
func TestUnitConfigurationAcceptedMalformedPreparationRetainsEarlierSuccessAndCallerRecovery(t *testing.T) {
	t.Parallel()
	clientRead := pol91ResponseFixture(t, "msteams/teams_admin_api/teams_client_configuration/testdata/get_global.json")
	malformed := []byte(`{"Identity":"Global","AudioFileValidationEnabled":false,"AudioFileValidationUri":"http://example.invalid/rejected"}`)
	privacyRead := []byte(`{"Identity":"Global","EnablePrivacyMode":false,"AutoInitiateContacts":null,"Future":false}`)
	steps := []testutil.Exchange{
		pol91ConfigurationWire("TeamsClientConfiguration", "GET", "Set-CsTeamsClientConfiguration", nil, 200, clientRead),
		pol91ConfigurationWire("TeamsClientConfiguration", "PATCH", "Set-CsTeamsClientConfiguration", json.RawMessage(`{"Identity":"Global","AllowEmailIntoChannel":true}`), 204, nil),
		pol91ConfigurationWire("OnlineVoicemailValidationConfiguration", "GET", "Set-CsOnlineVoicemailValidationConfiguration", nil, 200, malformed),
		pol91ConfigurationWire("PrivacyConfiguration", "GET", "Get-CsPrivacyConfiguration", nil, 200, privacyRead),
		pol91ConfigurationWire("TeamsClientConfiguration", "PATCH", "Set-CsTeamsClientConfiguration", json.RawMessage(`{"Identity":"Global","AllowEmailIntoChannel":false}`), 204, nil),
		pol91ConfigurationWire("TeamsClientConfiguration", "GET", "Get-CsTeamsClientConfiguration", nil, 200, clientRead),
	}
	c := testutil.Case{Name: "config-malformed-preparation-recovery",

		Steps: steps, Want: testutil.Expected{Status: 200, Body: clientRead}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			clients := sdk.TeamsAdminAPI.TeamsClientConfigurations
			accepted, response, err := clients.UpdateByIdentity(ctx, "Global", clientconfig.UpdateRequest{Settings: clientconfig.Settings{AllowEmailIntoChannel: policy.Set(true)}})
			if err != nil || response == nil || response.StatusCode() != 204 || accepted == nil || !accepted.HTTPAccepted || !accepted.PatchInvoked || accepted.After != nil || accepted.Before == nil || accepted.Before.AllowEmailIntoChannel == nil || *accepted.Before.AllowEmailIntoChannel {
				t.Fatal("earlier successful update lost its separate preparation/acceptance")
			}
			failed, response, err := sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations.UpdateByIdentity(ctx, "Global", validation.UpdateRequest{Settings: validation.Settings{AudioFileValidationEnabled: policy.Set(false)}})
			if !errors.Is(err, client.ErrDecode) || response == nil || response.StatusCode() != 200 || string(response.Body()) != string(malformed) || failed == nil || failed.Before != nil || failed.After != nil || failed.PatchInvoked || failed.HTTPAccepted || failed.ReadResponse == nil || failed.ReadResponse.StatusCode() != 200 || string(failed.ReadResponse.Body()) != string(malformed) {
				t.Fatal("accepted malformed preparation became a typed snapshot, PATCH or silent success")
			}
			independent, response, err := sdk.TeamsAdminAPI.PrivacyConfigurations.GetByIdentity(ctx, "Global")
			if err != nil || response == nil || independent == nil || independent.EnablePrivacyMode == nil || *independent.EnablePrivacyMode || string(independent.RawFields["AutoInitiateContacts"]) != "null" || string(independent.RawFields["Future"]) != "false" {
				t.Fatal("another family became unusable after a received-body decode error")
			}
			if _, response, err = clients.PatchByIdentity(ctx, "Global", clientconfig.PatchRequest{Settings: clientconfig.Settings{AllowEmailIntoChannel: policy.Set(*accepted.Before.AllowEmailIntoChannel)}}); err != nil || response == nil || response.StatusCode() != 204 {
				t.Fatal("explicit caller recovery after decode failure failed")
			}
			restored, response, err := clients.GetByIdentity(ctx, "Global")
			if err == nil && (restored == nil || restored.AllowEmailIntoChannel == nil || *restored.AllowEmailIntoChannel || !accepted.HTTPAccepted || *accepted.Before.AllowEmailIntoChannel || failed.Before != nil || failed.PatchInvoked) {
				t.Fatal("recovery erased the prior acceptance or fabricated a decoded failing-family state")
			}
			return testutil.Observation{Response: response, Err: err}
		}}
	result, err := testutil.Run(t.Context(), c)
	if err != nil || !result.Passed() {
		t.Fatalf("malformed configuration recovery differs: %v %+v", err, result.Differences())
	}
}
