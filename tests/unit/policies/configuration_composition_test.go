package policies_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	validation "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_validation_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	privacy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/privacy_configuration"
	clientconfig "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_client_configuration"
)

func pol91ResponseFixture(t *testing.T, file string) []byte {
	t.Helper()
	root, err := os.OpenRoot("../../..")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := root.Close(); err != nil {
			t.Error("fixture root close failed")
		}
	})
	raw, err := root.ReadFile(file)
	if err != nil || !json.Valid(raw) {
		t.Fatal("permanent configuration response fixture is unavailable or invalid")
	}
	return raw
}

func pol91ConfigurationWire(resource, method, command string, body json.RawMessage, status int, reply []byte) testutil.Exchange {
	h := http.Header{}
	h.Set("MPACmdlet", "true")
	h.Set("X-MS-CmdletName", command)
	h[http.CanonicalHeaderKey("X-MS-Target-Uri")] = nil
	if body != nil {
		h.Set("Content-Type", "application/json; charset=utf-8")
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/" + resource + "/configuration/Global", Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: reply}}
}

// Client/validation reads reuse permanent primitive response bytes. Privacy DTOs,
// write receipts, denial, readbacks and the compound sequence are authored offline
// from pinned Core properties and PolicyRp request/status rules. The three Set
// guards use bound-key presence, so false/null are settings. This is not captured
// tenant persistence, a transaction, effective privacy or service authorization.
func TestUnitConfigurationGlobalIdentitiesKeepPartialFailureAndExplicitRecoverySeparate(t *testing.T) {
	t.Parallel()
	clientRead := pol91ResponseFixture(t, "msteams/teams_admin_api/teams_client_configuration/testdata/get_global.json")
	validationRead := pol91ResponseFixture(t, "msteams/teams_admin_api/online_voicemail_validation_configuration/testdata/get_global.json")
	privacyRead := []byte(`{"Identity":"Global","EnablePrivacyMode":false,"PublishLocationDataDefault":true,"AutoInitiateContacts":null,"Future":false}`)
	steps := []testutil.Exchange{
		pol91ConfigurationWire("TeamsClientConfiguration", "GET", "Set-CsTeamsClientConfiguration", nil, 200, clientRead),
		pol91ConfigurationWire("TeamsClientConfiguration", "PATCH", "Set-CsTeamsClientConfiguration", json.RawMessage(`{"Identity":"Global","AllowEmailIntoChannel":true}`), 204, nil),
		pol91ConfigurationWire("PrivacyConfiguration", "GET", "Set-CsPrivacyConfiguration", nil, 200, privacyRead),
		pol91ConfigurationWire("PrivacyConfiguration", "PATCH", "Set-CsPrivacyConfiguration", json.RawMessage(`{"Identity":"Global","EnablePrivacyMode":null,"PublishLocationDataDefault":false}`), 204, nil),
		pol91ConfigurationWire("OnlineVoicemailValidationConfiguration", "GET", "Set-CsOnlineVoicemailValidationConfiguration", nil, 200, validationRead),
		pol91ConfigurationWire("OnlineVoicemailValidationConfiguration", "PATCH", "Set-CsOnlineVoicemailValidationConfiguration", json.RawMessage(`{"Identity":"Global","AudioFileValidationEnabled":false,"AudioFileValidationUri":null}`), 403, []byte(`{"Message":"fixture denied"}`)),
		pol91ConfigurationWire("OnlineVoicemailValidationConfiguration", "GET", "Get-CsOnlineVoicemailValidationConfiguration", nil, 200, validationRead),
		pol91ConfigurationWire("PrivacyConfiguration", "PATCH", "Set-CsPrivacyConfiguration", json.RawMessage(`{"Identity":"Global","EnablePrivacyMode":false,"PublishLocationDataDefault":true}`), 204, nil),
		pol91ConfigurationWire("PrivacyConfiguration", "GET", "Get-CsPrivacyConfiguration", nil, 200, privacyRead),
		pol91ConfigurationWire("TeamsClientConfiguration", "PATCH", "Set-CsTeamsClientConfiguration", json.RawMessage(`{"Identity":"Global","AllowEmailIntoChannel":false}`), 204, nil),
		pol91ConfigurationWire("TeamsClientConfiguration", "GET", "Get-CsTeamsClientConfiguration", nil, 200, clientRead),
	}
	c := testutil.Case{Name: "global-config-partial-recovery",

		Steps: steps, Want: testutil.Expected{Status: 200, Body: clientRead}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			clients, privacies, validations := sdk.TeamsAdminAPI.TeamsClientConfigurations, sdk.TeamsAdminAPI.PrivacyConfigurations, sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations
			first, response, err := clients.UpdateByIdentity(ctx, "Global", clientconfig.UpdateRequest{Settings: clientconfig.Settings{AllowEmailIntoChannel: policy.Set(true)}})
			if err != nil || response == nil || response.StatusCode() != 204 || first == nil || !first.PatchInvoked || !first.HTTPAccepted || first.After != nil || first.ReadResponse == nil || first.ReadResponse.StatusCode() != 200 || first.Before == nil || first.Before.Identity == nil || *first.Before.Identity != "Global" || first.Before.AllowEmailIntoChannel == nil || *first.Before.AllowEmailIntoChannel {
				t.Fatal("client update lost typed preparation or confused its receipt with persistence")
			}
			second, response, err := privacies.UpdateByIdentity(ctx, "Global", privacy.UpdateRequest{Settings: privacy.Settings{EnablePrivacyMode: policy.Null[bool](), PublishLocationDataDefault: policy.Set(false)}})
			if err != nil || response == nil || response.StatusCode() != 204 || second == nil || !second.PatchInvoked || !second.HTTPAccepted || second.After != nil || second.Before == nil || second.Before.Identity == nil || *second.Before.Identity != "Global" || second.Before.EnablePrivacyMode == nil || *second.Before.EnablePrivacyMode || second.Before.PublishLocationDataDefault == nil || !*second.Before.PublishLocationDataDefault || string(second.Before.RawFields["Future"]) != "false" {
				t.Fatal("same Global identity reused the wrong family state or dropped null/false selection")
			}
			failed, response, err := validations.UpdateByIdentity(ctx, "Global", validation.UpdateRequest{Settings: validation.Settings{AudioFileValidationEnabled: policy.Set(false), AudioFileValidationURI: policy.Null[string]()}})
			if !errors.Is(err, auth.ErrForbidden) || response == nil || response.StatusCode() != 403 || failed == nil || !failed.PatchInvoked || failed.HTTPAccepted || failed.After != nil || failed.Before == nil || failed.Before.AudioFileValidationEnabled == nil || !*failed.Before.AudioFileValidationEnabled || failed.Before.AudioFileValidationURI == nil || failed.ReadResponse == nil || failed.ReadResponse.StatusCode() != 200 {
				t.Fatal("later denial lost the prepared validation state or earlier receipts")
			}
			independent, response, err := validations.GetByIdentity(ctx, "Global")
			if err != nil || response == nil || independent == nil || independent.AudioFileValidationEnabled == nil || !*independent.AudioFileValidationEnabled {
				t.Fatal("explicit independent read was unusable after another operation failed")
			}
			if _, response, err = privacies.PatchByIdentity(ctx, "Global", privacy.PatchRequest{Settings: privacy.Settings{EnablePrivacyMode: policy.Set(*second.Before.EnablePrivacyMode), PublishLocationDataDefault: policy.Set(*second.Before.PublishLocationDataDefault)}}); err != nil || response == nil || response.StatusCode() != 204 {
				t.Fatal("caller-selected privacy restoration failed")
			}
			restoredPrivacy, response, err := privacies.GetByIdentity(ctx, "Global")
			if err != nil || response == nil || restoredPrivacy == nil || restoredPrivacy.EnablePrivacyMode == nil || *restoredPrivacy.EnablePrivacyMode || restoredPrivacy.PublishLocationDataDefault == nil || !*restoredPrivacy.PublishLocationDataDefault || string(restoredPrivacy.RawFields["AutoInitiateContacts"]) != "null" {
				t.Fatal("authored privacy restoration read lost its separate model state")
			}
			if _, response, err = clients.PatchByIdentity(ctx, "Global", clientconfig.PatchRequest{Settings: clientconfig.Settings{AllowEmailIntoChannel: policy.Set(*first.Before.AllowEmailIntoChannel)}}); err != nil || response == nil || response.StatusCode() != 204 {
				t.Fatal("caller-selected client restoration failed")
			}
			restoredClient, response, err := clients.GetByIdentity(ctx, "Global")
			if err == nil && (restoredClient == nil || restoredClient.AllowEmailIntoChannel == nil || *restoredClient.AllowEmailIntoChannel || first.Before.AllowEmailIntoChannel == nil || *first.Before.AllowEmailIntoChannel || second.Before.EnablePrivacyMode == nil || *second.Before.EnablePrivacyMode || string(second.Before.RawFields["Future"]) != "false" || !first.HTTPAccepted || !second.HTTPAccepted || failed.HTTPAccepted) {
				t.Fatal("explicit restoration mutated prior snapshots or erased independent acceptance/error states")
			}
			return testutil.Observation{Response: response, Err: err}
		}}
	result, err := testutil.Run(t.Context(), c)
	if err != nil || !result.Passed() {
		t.Fatalf("configuration composition differs: %v %+v", err, result.Differences())
	}
}
