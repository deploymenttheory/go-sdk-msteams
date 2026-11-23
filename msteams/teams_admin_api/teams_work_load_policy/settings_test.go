package teamsworkloadpolicy_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_work_load_policy"
)

func TestUnitAllBooleanBindsAndOmissionUseSparseModernPut(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		input      resource.Settings
	}{
		{"unbound", `{"Identity":"Global"}`, resource.Settings{}},
		{"all-false-empty", `{"Identity":"Global","Description":"","AllowCalling":false,"AllowCallingPinned":false,"AllowMeeting":false,"AllowMeetingPinned":false,"AllowMessaging":false,"AllowMessagingPinned":false}`, resource.Settings{Description: new(""), AllowCalling: new(false), AllowCallingPinned: new(false), AllowMeeting: new(false), AllowMeetingPinned: new(false), AllowMessaging: new(false), AllowMessagingPinned: new(false)}},
		{"all-true", `{"Identity":"Global","AllowCalling":true,"AllowCallingPinned":true,"AllowMeeting":true,"AllowMeetingPinned":true,"AllowMessaging":true,"AllowMessagingPinned":true}`, resource.Settings{AllowCalling: new(true), AllowCallingPinned: new(true), AllowMeeting: new(true), AllowMeetingPinned: new(true), AllowMessaging: new(true), AllowMessagingPinned: new(true)}},
	} {
		run(t, contracttest.Case{Name: tc.name, Mappings: []contracttest.Mapping{{Command: set, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{compiledWire("PUT", selected, set, json.RawMessage(tc.body), 503, "")}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 503}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			r, e := sdk.TeamsAdminAPI.TeamsWorkLoadPolicies.UpdateGlobal(ctx, tc.input)
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
}
