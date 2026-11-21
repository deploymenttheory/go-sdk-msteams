package client_test

import (
	"errors"
	"fmt"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func TestUnitPolicyRPErrorCodesUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	root, err := os.OpenRoot("../teams_admin_api")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := root.Close(); err != nil {
			t.Error(err)
		}
	})
	for _, test := range []struct {
		name, path, code    string
		status, serviceCode int
	}{
		{"migration", "assignment_migration/testdata/list_400.json", "ClientError", 400, 40003},
		{"prompts", "audio_conferencing_prompts/testdata/get_global_400.json", "ClientError", 400, 40003},
		{"privacy", "privacy_configuration/testdata/get_global_403.json", "Forbidden", 403, 40301},
	} {
		t.Run(test.name, func(t *testing.T) {
			body, err := root.ReadFile(test.path)
			if err != nil {
				t.Fatal(err)
			}
			api := client.NewAPIError(client.TeamsAdmin, test.status, http.Header{}, body)
			if !reflect.DeepEqual(api.Codes, []string{test.code}) || !reflect.DeepEqual(api.ServiceErrorCodes, []int{test.serviceCode}) {
				t.Fatalf("service error codes lost: %v", api.ServiceErrorCodes)
			}
			if errors.Is(api, auth.ErrTenantCapabilityUnavailable) {
				t.Fatal("generic transport inferred resource capability")
			}
			for _, format := range []string{"%v", "%+v", "%#v"} {
				if strings.Contains(fmt.Sprintf(format, api), fmt.Sprint(test.serviceCode)) {
					t.Fatal("error formatting leaked service diagnostics")
				}
			}
		})
	}
}

func TestUnitMalformedServiceErrorCodePreservesOtherDiagnostics(t *testing.T) {
	t.Parallel()
	for _, value := range []string{`null`, `"40003"`, `{}`, `[]`, `40003.5`, `999999999999999999999999999999`} {
		t.Run(value, func(t *testing.T) {
			api := client.NewAPIError(client.TeamsAdmin, 400, nil, []byte(`{"code":"ClientError","errorCode":`+value+`}`))
			if !reflect.DeepEqual(api.Codes, []string{"ClientError"}) || len(api.ServiceErrorCodes) != 0 {
				t.Fatal("malformed numeric code discarded unrelated diagnostics or was coerced")
			}
		})
	}
}
