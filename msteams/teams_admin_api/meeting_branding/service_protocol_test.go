package meetingbranding_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/meeting_branding"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

// Primary-authored empty acceptance, not a captured nested-asset mutation.
// The settled pinned builder/property contracts supply the request vector;
// Set shares the PolicyRp generic policy parser's empty-success branch.
// Assets are local URI DTOs: no upload, fetch, rendering or persistence claim.
func TestUnitBrandingNestedBuildersReachSuccessfulPatch(t *testing.T) {
	t.Parallel()
	seen := 0
	for _, vector := range requestInputs(t) {
		if vector.Name != "patch-nested" {
			continue
		}
		seen++
		t.Run(vector.Name, func(t *testing.T) {
			const status = 204
			mappings := []contracttest.Mapping{
				{Command: vector.Request.CmdletName, ParameterSet: "Identity"},
				{Command: "New-CsTeamsMeetingBackgroundImage", ParameterSet: "Identity"},
				{Command: "New-CsTeamsMeetingBrandingTheme", ParameterSet: "Identity"},
				{Command: "New-CsTeamsNdiAssuranceSlate", ParameterSet: "Identity"},
			}
			result, err := contracttest.Run(t.Context(), contracttest.Case{Name: "accepted-" + vector.Name, Family: "meeting_branding", Scope: "service", Source: "msteams/teams_admin_api/meeting_branding/service_protocol_test.go", Mappings: mappings, Steps: []contracttest.Exchange{wire(vector.Request.Method, strings.TrimPrefix(vector.Request.Url, origin), vector.Request.CmdletName, json.RawMessage(vector.Request.Body), status, "")}, Want: contracttest.Expected{Status: status, Body: []byte{}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				input := settings(t, true)
				s := sdk.TeamsAdminAPI.MeetingBrandingPolicies
				value, response, err := s.PatchByIdentity(ctx, scoped, resource.PatchRequest{Settings: input, BodyIdentity: policy.Set("read-model-identity")})
				if value != nil {
					t.Fatal("empty nested patch receipt fabricated a readback")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
			if err != nil || !result.Passed() {
				t.Fatalf("nested builder acceptance differs: %v %+v", err, result.Differences())
			}
		})
	}
	if seen != 1 {
		t.Fatal("nested patch contract vector missing")
	}
}
