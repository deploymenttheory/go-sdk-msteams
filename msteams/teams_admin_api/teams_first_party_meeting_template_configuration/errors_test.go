package teamsfirstpartymeetingtemplateconfiguration_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_first_party_meeting_template_configuration"
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
	want := contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 403, Codes: codes, ServiceErrorCodes: []int{envelope.ErrorCode}}}
	run(t, contracttest.Case{Name: "denied-write", Source: "msteams/teams_admin_api/privacy_configuration/testdata/get_global_403.json", Steps: []contracttest.Exchange{compiledWire("PATCH", selected, set, json.RawMessage(`{"Identity":"Global","Description":"x"}`), 403, string(raw))}, Want: want, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, r, e := sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{ScalarSettings: resource.ScalarSettings{Description: policy.Set("x")}}})
		return contracttest.Observation{Response: r, Err: e}
	}})
	run(t, contracttest.Case{Name: "denied-preparation", Source: "msteams/teams_admin_api/privacy_configuration/testdata/get_global_403.json", Steps: []contracttest.Exchange{compiledWire("GET", selected, set, nil, 403, string(raw))}, Want: want, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{Description: policy.Set("x")}})
		if v == nil || v.PatchInvoked || v.HTTPAccepted || v.ReadResponse == nil {
			t.Fatal("denied preparation lost its receipt or attempted a mutation")
		}
		return contracttest.Observation{Response: r, Err: e}
	}})
}
