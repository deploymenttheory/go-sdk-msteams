package policies_test

import (
	"encoding/json"
	"os"
	"reflect"
	"slices"
	"strings"
	"testing"

	acs "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/acs_federation_configuration"
	external "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/external_access_policy"
	teams "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_external_access_configuration"
	tenant "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_federation_configuration"
)

func TestUnitFederationRequestFieldsMatchPinnedParameters(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/powershell/8.1.0/runtime.json")
	if err != nil {
		t.Fatal(err)
	}
	var runtime struct {
		Commands []struct {
			Name string `json:"name"`
			Sets []struct {
				Parameters []struct {
					Name string `json:"name"`
				} `json:"parameters"`
			} `json:"parameter_sets"`
		} `json:"commands"`
	}
	if json.Unmarshal(raw, &runtime) != nil {
		t.Fatal("invalid pinned signatures")
	}
	raw, err = os.ReadFile("../../../contracts/coverage/8.1.0.json")
	if err != nil {
		t.Fatal(err)
	}
	var ledger struct {
		Common map[string]string `json:"common_parameters"`
	}
	if json.Unmarshal(raw, &ledger) != nil {
		t.Fatal("invalid common parameter map")
	}
	requests := map[string]any{"Set-CsExternalAccessPolicy": external.Settings{}, "Set-CsTeamsAcsFederationConfiguration": acs.Settings{}, "Set-CsTeamsExternalAccessConfiguration": teams.Settings{}, "Set-CsTenantFederationConfiguration": tenant.UpdateRequest{}}
	seen := 0
	for _, command := range runtime.Commands {
		request, ok := requests[command.Name]
		if !ok {
			continue
		}
		seen++
		want := map[string]bool{}
		for _, set := range command.Sets {
			for _, p := range set.Parameters {
				// Identity/Force have separate request controls. Internal processing
				// mode chooses the PowerShell flight, not a native resource setting.
				if _, common := ledger.Common[p.Name]; common || p.Name == "Identity" || p.Name == "Force" || p.Name == "MsftInternalProcessingMode" {
					continue
				}
				want[p.Name] = true
			}
		}
		var fields func(reflect.Type) []string
		fields = func(typ reflect.Type) []string {
			var names []string
			for i := range typ.NumField() {
				field := typ.Field(i)
				if field.Anonymous {
					names = append(names, fields(field.Type)...)
					continue
				}
				name := strings.Split(field.Tag.Get("json"), ",")[0]
				if name == "-" {
					continue
				}
				if name == "" {
					name = field.Name
				}
				names = append(names, name)
			}
			return names
		}
		got := fields(reflect.TypeOf(request))
		slices.Sort(got)
		var expected []string
		for name := range want {
			expected = append(expected, name)
		}
		slices.Sort(expected)
		if !reflect.DeepEqual(got, expected) {
			t.Fatalf("%s request fields differ from pinned signature: got %v want %v", command.Name, got, expected)
		}
	}
	if seen != 4 {
		t.Fatal("federation signature inventory incomplete")
	}
}
