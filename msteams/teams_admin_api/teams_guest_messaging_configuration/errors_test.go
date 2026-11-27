package teamsguestmessagingconfiguration_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_guest_messaging_configuration"
)

// The common PolicyRP authorization envelope is from a real privacy request.
// Replaying it here verifies error propagation; it is not a live denial claim
// for this configuration or a fabricated successful service response.
func TestUnitAuthorizationFailureRetainsRealEnvelopeWithoutReplay(t *testing.T) {
	t.Parallel()
	raw, e := os.ReadFile("../privacy_configuration/testdata/get_global_403.json")
	if e != nil {
		t.Fatal(e)
	}
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
	c := testutil.Case{Name: "denied", Steps: []testutil.Exchange{compiledWire("PUT", selected, set, json.RawMessage(`{"Identity":"Global","AllowMemes":false}`), 403, string(raw))}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 403, Codes: codes, ServiceErrorCodes: []int{envelope.ErrorCode}}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		r, e := sdk.TeamsAdminAPI.TeamsGuestMessagingConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{AllowMemes: new(false)})
		return testutil.Observation{Response: r, Err: e}
	}}
	run(t, c)
}
