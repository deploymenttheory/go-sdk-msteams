package teamssipdevicesconfiguration_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_sip_devices_configuration"
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
	c := testutil.Case{Name: "denied", Steps: []testutil.Exchange{compiledWire("PATCH", selected, set, json.RawMessage(`{"Identity":"Global","BulkSignIn":null}`), 403, string(raw))}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 403, Codes: codes, ServiceErrorCodes: []int{envelope.ErrorCode}}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		_, r, e := sdk.TeamsAdminAPI.TeamsSipDevicesConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{BulkSignIn: policy.Null[string]()}})
		return testutil.Observation{Response: r, Err: e}
	}}
	run(t, c)
}
