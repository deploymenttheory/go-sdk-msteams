package policies_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	firstparty "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_first_party_meeting_template_configuration"
)

// Reads retain permanent primitive acceptance bytes. The compound sequence,
// later denial/malformed reply and recovery are authored offline. The malformed
// Description object is rejected by the own pinned DTO probe. Replaying real
// readbacks does not establish persistence, atomicity or live authorization.
func TestUnitFirstPartyTemplatePreparedFailureRetainsReceiptsAndCallerRecovery(t *testing.T) {
	t.Parallel()
	const resource = "TeamsFirstPartyMeetingTemplateConfiguration"
	const set = "Set-CsTeamsFirstPartyMeetingTemplateConfiguration"
	const get = "Get-CsTeamsFirstPartyMeetingTemplateConfiguration"
	base := pol91ResponseFixture(t, "msteams/teams_admin_api/teams_first_party_meeting_template_configuration/testdata/get_global.json")
	updated := pol91ResponseFixture(t, "msteams/teams_admin_api/teams_first_party_meeting_template_configuration/testdata/update_readback.json")
	restored := pol91ResponseFixture(t, "msteams/teams_admin_api/teams_first_party_meeting_template_configuration/testdata/restored_readback.json")
	independent := pol91ResponseFixture(t, "msteams/teams_admin_api/teams_meeting_template_configuration/testdata/get_global.json")
	denied := pol91ResponseFixture(t, "msteams/teams_admin_api/privacy_configuration/testdata/get_global_403.json")
	var original, changed firstparty.Configuration
	if json.Unmarshal(base, &original) != nil || json.Unmarshal(updated, &changed) != nil || changed.Description == nil || original.Description != nil || len(original.TeamsMeetingTemplates) == 0 {
		t.Fatal("permanent first-party preparation/readback fixtures unavailable")
	}
	request := func(description *string) json.RawMessage {
		t.Helper()
		raw, err := json.Marshal(map[string]any{"Identity": "Global", "Description": description})
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	rejected := "offline rejected update"
	for _, tc := range []struct {
		name     string
		status   int
		reply    []byte
		want     error
		accepted bool
	}{
		{"denied", 403, denied, auth.ErrForbidden, false},
		{"accepted-malformed", 200, []byte(`{"Description":{}}`), client.ErrDecode, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := testutil.Case{Name: "firstparty-" + tc.name + "-recovery",

				Steps: []testutil.Exchange{
					pol91ConfigurationWire(resource, "GET", set, nil, 200, base),
					pol91ConfigurationWire(resource, "PATCH", set, request(changed.Description), 204, nil),
					pol91ConfigurationWire(resource, "GET", set, nil, 200, updated),
					pol91ConfigurationWire(resource, "PATCH", set, request(&rejected), tc.status, tc.reply),
					pol91ConfigurationWire("TeamsMeetingTemplateConfiguration", "GET", "Get-CsTeamsMeetingTemplateConfiguration", nil, 200, independent),
					pol91ConfigurationWire(resource, "PATCH", set, request(original.Description), 204, nil),
					pol91ConfigurationWire(resource, "GET", get, nil, 200, restored),
				}, Want: testutil.Expected{Status: 200, Body: restored}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					service := sdk.TeamsAdminAPI.TeamsFirstPartyMeetingTemplateConfigurations
					first, response, err := service.UpdateByIdentity(ctx, "Global", firstparty.UpdateRequest{ScalarSettings: firstparty.ScalarSettings{Description: policy.Set(*changed.Description)}})
					if err != nil || response == nil || response.StatusCode() != 204 || first == nil || !first.PatchInvoked || !first.HTTPAccepted || first.After != nil || first.Before == nil || first.Before.Identity == nil || *first.Before.Identity != "Global" || first.Before.Description != nil || len(first.Before.TeamsMeetingTemplates) != len(original.TeamsMeetingTemplates) || first.ReadResponse == nil || first.ReadResponse.StatusCode() != 200 || !bytes.Equal(first.ReadResponse.Body(), base) || !bytes.Equal(first.Before.RawJSON, bytes.TrimSpace(base)) {
						t.Fatal("accepted first update lost preparation or invented a typed write result")
					}
					failed, response, err := service.UpdateByIdentity(ctx, "Global", firstparty.UpdateRequest{ScalarSettings: firstparty.ScalarSettings{Description: policy.Set(rejected)}})
					if !errors.Is(err, tc.want) || response == nil || response.StatusCode() != tc.status || !bytes.Equal(response.Body(), tc.reply) || failed == nil || !failed.PatchInvoked || failed.HTTPAccepted != tc.accepted || failed.After != nil || failed.Before == nil || failed.Before.Description == nil || *failed.Before.Description != *changed.Description || failed.ReadResponse == nil || failed.ReadResponse.StatusCode() != 200 || !bytes.Equal(failed.ReadResponse.Body(), updated) {
						t.Fatal("later write failure lost typed preparation, raw receipt or actual acceptance state")
					}
					other, response, err := sdk.TeamsAdminAPI.TeamsMeetingTemplateConfigurations.GetByIdentity(ctx, "Global")
					if err != nil || response == nil || response.StatusCode() != 200 || other == nil || other.Identity == nil || *other.Identity != "Global" || !bytes.Equal(response.Body(), independent) || !bytes.Equal(other.RawJSON, bytes.TrimSpace(independent)) {
						t.Fatal("independent same-Global family read was unusable or crossed response state")
					}
					// Recovery is caller-selected from the original preparation. Update never
					// retries, rolls back or infers persistence after either failure outcome.
					restore := policy.Null[string]()
					if original.Description != nil {
						restore = policy.Set(*original.Description)
					}
					value, response, err := service.PatchByIdentity(ctx, "Global", firstparty.PatchRequest{Settings: firstparty.Settings{ScalarSettings: firstparty.ScalarSettings{Description: restore}}})
					if err != nil || response == nil || response.StatusCode() != 204 || value != nil {
						t.Fatal("explicit caller restoration failed or fabricated a typed204 value")
					}
					readback, response, err := service.GetByIdentity(ctx, "Global")
					if err == nil && (response == nil || readback == nil || readback.Identity == nil || *readback.Identity != "Global" || readback.Description != nil || len(readback.TeamsMeetingTemplates) != len(original.TeamsMeetingTemplates) || !first.HTTPAccepted || first.After != nil || !bytes.Equal(first.Before.RawJSON, bytes.TrimSpace(base)) || failed.HTTPAccepted != tc.accepted || !bytes.Equal(failed.Before.RawJSON, bytes.TrimSpace(updated)) || !bytes.Equal(other.RawJSON, bytes.TrimSpace(independent))) {
						t.Fatal("authored restoration/readback changed prior snapshots or confused acceptance with persistence")
					}
					return testutil.Observation{Response: response, Err: err}
				}}
			result, err := testutil.Run(t.Context(), c)
			if err != nil || !result.Passed() {
				t.Fatalf("first-party recovery composition differs: %v %+v", err, result.Differences())
			}
		})
	}
}
