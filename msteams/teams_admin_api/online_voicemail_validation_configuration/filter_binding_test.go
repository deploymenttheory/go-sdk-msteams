package onlinevoicemailvalidationconfiguration_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_validation_configuration"
)

// The effective public Filter inherits ValidateNotNullOrEmpty. The helper's
// null/empty branch describes omission after binding, not an explicit empty
// argument. Exact metadata/hash and helper IL are in the POL91 source audit.
func TestUnitValidationFilterBindingDistinguishesOmissionEmptyAndWhitespace(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "bound-empty-filter", Want: testutil.Expected{Error: resource.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		value, response, err := sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations.ListFiltered(ctx, "")
		return testutil.Observation{Value: value, Response: response, Err: err}
	}})
	raw := readResponseFixture(t, "list.json")
	var expected []*resource.Configuration
	if err := json.Unmarshal(raw, &expected); err != nil || len(expected) == 0 {
		t.Fatal("permanent list fixture must retain typed configurations")
	}
	for _, whitespace := range []bool{false, true} {
		name := "omitted-filter"
		if whitespace {
			name = "bound-whitespace-filter"
		}
		run(t, testutil.Case{Name: name, Steps: []testutil.Exchange{wire("GET", collection, "Get-CsOnlineVoicemailValidationConfiguration", nil, 200, string(raw))}, Want: testutil.Expected{Body: raw}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			var observed testutil.Observation
			var values []*resource.Configuration
			if whitespace {
				values, observed.Response, observed.Err = sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations.ListFiltered(ctx, " ")
			} else {
				values, observed.Response, observed.Err = sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations.List(ctx)
			}
			if observed.Err == nil && ((!whitespace && len(values) != len(expected)) || (whitespace && len(values) != 0)) {
				t.Fatal("omitted collection and literal whitespace filter were conflated")
			}
			observed.Value = values
			return observed
		}})
	}
}
