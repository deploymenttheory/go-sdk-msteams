package policytransport

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func TestUnitFlightDisabledUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	root, err := os.OpenRoot("../..")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := root.Close(); err != nil {
			t.Error(err)
		}
	})
	for _, test := range []struct {
		name, path  string
		status      int
		unavailable bool
	}{
		{"migration", "assignment_migration/testdata/list_400.json", 400, true},
		{"prompts", "audio_conferencing_prompts/testdata/get_global_400.json", 400, true},
		{"privacy", "privacy_configuration/testdata/get_global_403.json", 403, false},
	} {
		t.Run(test.name, func(t *testing.T) {
			body, err := root.ReadFile(test.path)
			if err != nil {
				t.Fatal(err)
			}
			api := client.NewAPIError(client.TeamsAdmin, test.status, nil, body)
			got := classifyError(api, test.status, body)
			var retained *client.APIError
			if errors.Is(got, auth.ErrTenantCapabilityUnavailable) != test.unavailable || !errors.As(got, &retained) || retained != api {
				t.Fatal("capability classification lost the original API error or inferred a generic restriction")
			}
			for _, format := range []string{"%v", "%+v", "%#v"} {
				if strings.Contains(fmt.Sprintf(format, got), "flighting") {
					t.Fatal("error formatting retained server text")
				}
			}
		})
	}
}

func TestUnitFlightClassificationRejectsAmbiguousErrors(t *testing.T) {
	t.Parallel()
	body, err := os.ReadFile("../../audio_conferencing_prompts/testdata/get_global_400.json")
	if err != nil {
		t.Fatal(err)
	}
	for _, test := range []struct {
		name    string
		status  int
		backend client.Backend
		field   string
		value   any
	}{
		{"wrong-status", 403, client.TeamsAdmin, "", nil},
		{"wrong-backend", 400, client.Graph, "", nil},
		{"different-code", 400, client.TeamsAdmin, "errorCode", 40004},
		{"wrong-code-type", 400, client.TeamsAdmin, "errorCode", "40003"},
		{"different-category", 400, client.TeamsAdmin, "code", "Forbidden"},
		{"generic-message", 400, client.TeamsAdmin, "message", "feature unavailable"},
	} {
		t.Run(test.name, func(t *testing.T) {
			var object map[string]any
			if json.Unmarshal(body, &object) != nil {
				t.Fatal("invalid acceptance fixture")
			}
			if test.field != "" {
				object[test.field] = test.value
			}
			mutated, err := json.Marshal(object)
			if err != nil {
				t.Fatal(err)
			}
			api := client.NewAPIError(test.backend, test.status, nil, mutated)
			if result := classifyError(api, test.status, mutated); !errors.Is(result, api) || errors.Is(result, auth.ErrTenantCapabilityUnavailable) {
				t.Fatal("ambiguous error inferred a tenant restriction")
			}
		})
	}
}
