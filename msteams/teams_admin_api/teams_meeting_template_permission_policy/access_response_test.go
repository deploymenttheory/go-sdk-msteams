package teamsmeetingtemplatepermissionpolicy_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_template_permission_policy"
)

func TestUnitPopulatedCreateUsesActualAuthorizationResponse(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "create_populated_denied_400.json")
	for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
		run(t, testutil.Case{Name: "populated-create-" + string(mode), Mode: mode, Steps: []testutil.Exchange{compiledWire("POST", collection, "New-CsTeamsMeetingTemplatePermissionPolicy", json.RawMessage(`{"Identity":"fixture-policy","Description":"SDK lab non-default initial description","HiddenMeetingTemplates":[{"Identity":null,"Id":"caller-template-id"}]}`), 400, string(raw))}, Want: testutil.Expected{Status: 400, Body: raw, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400, Codes: []string{"ClientError"}, ServiceErrorCodes: []int{40013}}}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			_, response, err := sdk.TeamsAdminAPI.TeamsMeetingTemplatePermissionPolicies.CreateByIdentity(ctx, "fixture-policy", resource.CreateRequest{ScalarSettings: resource.ScalarSettings{Description: policy.Set("SDK lab non-default initial description")}, HiddenMeetingTemplates: resource.ReplaceHiddenTemplates(resource.NewHiddenTemplate("caller-template-id"))})
			return testutil.Observation{Response: response, Err: err}
		}})
	}
}
