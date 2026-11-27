package onlinevoicemailvalidationconfiguration_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_validation_configuration"
)

// The effective public Filter inherits ValidateNotNullOrEmpty. The helper's
// null/empty branch describes omission after binding, not an explicit empty
// argument. Exact metadata/hash and helper IL are in the POL91 source audit.
func TestUnitValidationFilterBindingDistinguishesOmissionEmptyAndWhitespace(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "bound-empty-filter", Source: "msteams/teams_admin_api/online_voicemail_validation_configuration/filter_binding_test.go", Mappings: []contracttest.Mapping{{Command: "Get-CsOnlineVoicemailValidationConfiguration", ParameterSet: "Filter"}}, Want: contracttest.Expected{Error: resource.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		value, response, err := sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations.ListFiltered(ctx, "")
		return contracttest.Observation{Value: value, Response: response, Err: err}
	}})
	raw := readResponseFixture(t, "list.json")
	var expected []*resource.Configuration
	if err := json.Unmarshal(raw, &expected); err != nil || len(expected) == 0 {
		t.Fatal("permanent list fixture must retain typed configurations")
	}
	for _, whitespace := range []bool{false, true} {
		name, set := "omitted-filter", "Identity"
		if whitespace {
			name, set = "bound-whitespace-filter", "Filter"
		}
		run(t, contracttest.Case{Name: name, Source: "msteams/teams_admin_api/online_voicemail_validation_configuration/testdata/list.json", Mappings: []contracttest.Mapping{{Command: "Get-CsOnlineVoicemailValidationConfiguration", ParameterSet: set}}, Steps: []contracttest.Exchange{wire("GET", collection, "Get-CsOnlineVoicemailValidationConfiguration", nil, 200, string(raw))}, Want: contracttest.Expected{Body: raw}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			var observed contracttest.Observation
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
