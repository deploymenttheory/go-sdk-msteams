package teamsaipolicy_test

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_ai_policy"
)

// TestUnitDescriptionLengthIsNotLocallyValidated proves the SDK sends empty and
// over-1040-character descriptions unchanged. The pinned TeamsAIPolicy
// Description setter evaluates 1/1040 length constants but discards the result
// and never throws, so length validation stays with the tenant.
func TestUnitDescriptionLengthIsNotLocallyValidated(t *testing.T) {
	t.Parallel()
	for _, description := range []string{"", strings.Repeat("d", 1041)} {
		body, err := json.Marshal(map[string]string{"Identity": "fixture-policy", "Description": description})
		if err != nil {
			t.Fatal(err)
		}
		run(t, contracttest.Case{Name: fmt.Sprintf("description-length-%d", len(description)), Mappings: []contracttest.Mapping{{Command: "Set-CsTeamsAIPolicy", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{compiledWire("PATCH", collection+"/configuration/fixture-policy", "Set-CsTeamsAIPolicy", body, 400, "")}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsAIPolicies.PatchByIdentity(ctx, "fixture-policy", resource.PatchRequest{Settings: resource.Settings{Description: policy.Set(description)}})
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
}
