package teamsmeetingconfiguration_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_configuration"
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
	c := contracttest.Case{Name: "denied", Source: "msteams/teams_admin_api/privacy_configuration/testdata/get_global_403.json", Steps: []contracttest.Exchange{compiledWire("PATCH", selected, set, json.RawMessage(`{"Identity":"Global","CustomFooterText":null}`), 403, string(raw))}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 403, Codes: codes, ServiceErrorCodes: []int{envelope.ErrorCode}}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, r, e := sdk.TeamsAdminAPI.TeamsMeetingConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{ScalarSettings: resource.ScalarSettings{CustomFooterText: policy.Null[string]()}}})
		return contracttest.Observation{Response: r, Err: e}
	}}
	run(t, c)
}

// Deliberately malformed replies (not service fixtures) fail decoding, and a
// failed preparation read never reaches the PATCH.
func TestUnitMalformedResponsesFailDecoding(t *testing.T) {
	t.Parallel()
	for _, reply := range []string{`{`, `[]`, `{"PublishedEntraAuthenticationContexts":{}}`, `{"ClientVideoPort":"x"}`} {
		run(t, contracttest.Case{Name: "malformed-get", Source: "msteams/teams_admin_api/teams_meeting_configuration/errors_test.go", Steps: []contracttest.Exchange{compiledWire("GET", selected, get, nil, 200, reply)}, Want: contracttest.Expected{Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsMeetingConfigurations.GetByIdentity(ctx, "Global")
			return contracttest.Observation{Response: r, Err: e}
		}})
		run(t, contracttest.Case{Name: "malformed-preparation", Source: "msteams/teams_admin_api/teams_meeting_configuration/errors_test.go", Steps: []contracttest.Exchange{compiledWire("GET", selected, set, nil, 200, reply)}, Want: contracttest.Expected{Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsMeetingConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{EnableQoS: policy.Set(true)}})
			if v == nil || v.PatchInvoked || v.HTTPAccepted {
				t.Fatal("failed preparation reached the mutation")
			}
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
	run(t, contracttest.Case{Name: "malformed-list", Source: "msteams/teams_admin_api/teams_meeting_configuration/errors_test.go", Steps: []contracttest.Exchange{compiledWire("GET", collection, get, nil, 200, `{}`)}, Want: contracttest.Expected{Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, r, e := sdk.TeamsAdminAPI.TeamsMeetingConfigurations.List(ctx)
		return contracttest.Observation{Response: r, Err: e}
	}})
}
