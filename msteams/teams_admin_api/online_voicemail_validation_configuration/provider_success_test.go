package onlinevoicemailvalidationconfiguration_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_validation_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

// Core's nullable Boolean/string DTO and URI setter permit false/null. The
// controlled pinned request probes already establish these sparse request bytes;
// these accepted replies are primary-authored, not captured tenant writes.
func TestUnitValidationSparsePatchAndPreparedUpdateReachAcceptedBranches(t *testing.T) {
	t.Parallel()
	const typed = `{"Identity":"prepared-identity","AudioFileValidationEnabled":false,"AudioFileValidationUri":null,"Future":false}`
	run(t, contracttest.Case{Name: "typed-sparse-patch", Source: "msteams/teams_admin_api/online_voicemail_validation_configuration/provider_success_test.go", Mappings: []contracttest.Mapping{{Command: "Set-CsOnlineVoicemailValidationConfiguration", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{wire("PATCH", item, "Set-CsOnlineVoicemailValidationConfiguration", json.RawMessage(`{"Identity":"prepared-identity","AudioFileValidationEnabled":false,"AudioFileValidationUri":null}`), 200, typed)}, Want: contracttest.Expected{Body: []byte(typed), JSON: json.RawMessage(typed)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		value, response, err := sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{AudioFileValidationEnabled: policy.Set(false), AudioFileValidationURI: policy.Null[string]()}, BodyIdentity: policy.Set("prepared-identity")})
		if err == nil && (value == nil || value.Identity == nil || *value.Identity != "prepared-identity" || value.AudioFileValidationEnabled == nil || *value.AudioFileValidationEnabled || value.AudioFileValidationURI != nil || string(value.RawFields["AudioFileValidationUri"]) != "null" || string(value.RawFields["Future"]) != "false") {
			t.Fatal("typed PATCH lost body identity, false/null or raw future data")
		}
		return contracttest.Observation{Value: value, Response: response, Err: err}
	}})

	before := readResponseFixture(t, "get_global.json")
	run(t, contracttest.Case{Name: "prepared-update-acceptance", Source: "msteams/teams_admin_api/online_voicemail_validation_configuration/provider_success_test.go", Mappings: []contracttest.Mapping{{Command: "Set-CsOnlineVoicemailValidationConfiguration", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{
		wire("GET", item, "Set-CsOnlineVoicemailValidationConfiguration", nil, 200, string(before)),
		wire("PATCH", item, "Set-CsOnlineVoicemailValidationConfiguration", json.RawMessage(`{"Identity":"Global","AudioFileValidationEnabled":false,"AudioFileValidationUri":null}`), 204, ""),
	}, Want: contracttest.Expected{Body: []byte{}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		value, response, err := sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{AudioFileValidationEnabled: policy.Set(false), AudioFileValidationURI: policy.Null[string]()}})
		if err == nil && (response == nil || response.StatusCode() != 204 || value == nil || !value.PatchInvoked || !value.HTTPAccepted || value.After != nil || value.Before == nil || value.Before.Identity == nil || *value.Before.Identity != "Global" || value.Before.AudioFileValidationEnabled == nil || !*value.Before.AudioFileValidationEnabled || value.Before.AudioFileValidationURI == nil || value.ReadResponse == nil || value.ReadResponse.StatusCode() != 200 || string(value.ReadResponse.Body()) != string(before)) {
			t.Fatal("accepted Update lost its preparation or fabricated persistence/readback")
		}
		return contracttest.Observation{Response: response, Err: err}
	}})
}
