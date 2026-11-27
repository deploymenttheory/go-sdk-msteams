package teamsmeetingbroadcastconfiguration_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_broadcast_configuration"
)

func envelopeCodes(t *testing.T, raw []byte) ([]string, int) {
	t.Helper()
	var envelope struct {
		Code      string `json:"code"`
		ErrorCode int    `json:"errorCode"`
	}
	if json.Unmarshal(raw, &envelope) != nil {
		t.Fatal("invalid actual error fixture")
	}
	var codes []string
	if envelope.Code != "" {
		codes = []string{envelope.Code}
	}
	return codes, envelope.ErrorCode
}

// The common PolicyRP authorization envelope is from a real privacy request.
// Replaying it here verifies error propagation; it is not a live denial claim
// for this configuration or a fabricated successful service response.
func TestUnitAuthorizationFailureRetainsRealEnvelopeWithoutReplay(t *testing.T) {
	t.Parallel()
	raw, e := os.ReadFile("../privacy_configuration/testdata/get_global_403.json")
	if e != nil {
		t.Fatal(e)
	}
	codes, code := envelopeCodes(t, raw)
	url := "https://support.example.com/broadcast"
	c := testutil.Case{Name: "denied", Steps: []testutil.Exchange{compiledWire("PUT", selected, set, json.RawMessage(`{"Identity":"Global","SupportURL":"https://support.example.com/broadcast"}`), 403, string(raw))}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 403, Codes: codes, ServiceErrorCodes: []int{code}}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		r, e := sdk.TeamsAdminAPI.TeamsMeetingBroadcastConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{SupportURL: &url})
		return testutil.Observation{Response: r, Err: e}
	}}
	run(t, c)
}

// The live service rejected enabling the SDN provider without SdnProviderName
// (40013 SkypecastSdnMissingParameter). The SDK sends the sparse body as bound,
// performs no local cross-field validation and surfaces the service error.
func TestUnitServiceValidationRejectsProviderWithoutNameWithoutReplay(t *testing.T) {
	t.Parallel()
	const name = "update_allow_without_provider_400.json"
	raw := readResponseFixture(t, name)
	codes, code := envelopeCodes(t, raw)
	c := testutil.Case{Name: "sdn-missing-provider", Steps: []testutil.Exchange{compiledWire("PUT", selected, set, json.RawMessage(`{"Identity":"Global","AllowSdnProviderForBroadcastMeeting":true}`), 400, string(raw))}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400, Codes: codes, ServiceErrorCodes: []int{code}}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		r, e := sdk.TeamsAdminAPI.TeamsMeetingBroadcastConfigurations.UpdateGlobal(ctx, resource.UpdateRequest{AllowSdnProviderForBroadcastMeeting: new(true)})
		return testutil.Observation{Response: r, Err: e}
	}}
	run(t, c)
}
