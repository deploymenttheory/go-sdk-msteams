package teamsclientconfiguration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_client_configuration"
)

// The successful body is a retained read replayed as an authored PATCH reply.
// It proves native typed response handling, not a captured write, persistence
// or convergence of the submitted null setting.
func TestUnitClientConfigurationPatchRetainsTypedSuccessAndMalformedReceipt(t *testing.T) {
	t.Parallel()
	read := readResponseFixture(t, "get_global.json")
	for _, tc := range []struct {
		name string
		body []byte
		want error
	}{
		{name: "typed-success", body: read},
		{name: "malformed-success", body: []byte(`{"Identity":"Global","AllowEmailIntoChannel":"invalid"}`), want: client.ErrDecode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{
				Name: tc.name, Source: "msteams/teams_admin_api/teams_client_configuration/provider_success_test.go",
				Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}},
				Steps: []contracttest.Exchange{compiledWire("PATCH", item, setCommand,
					json.RawMessage(`{"Identity":"request-body-identity","AllowEmailIntoChannel":false,"RestrictedSenderList":null}`), 200, string(tc.body))},
				Want: contracttest.Expected{Status: 200, Body: tc.body, Error: tc.want},
				Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					value, response, err := sdk.TeamsAdminAPI.TeamsClientConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{
						Settings:     resource.Settings{AllowEmailIntoChannel: policy.Set(false), RestrictedSenderList: policy.Null[string]()},
						BodyIdentity: policy.Set("request-body-identity"),
					})
					if response == nil || response.StatusCode() != 200 || !bytes.Equal(response.Body(), tc.body) {
						t.Fatal("PATCH lost the received response")
					}
					if tc.want != nil {
						if value != nil || !errors.Is(err, tc.want) {
							t.Fatal("malformed successful response became a typed result or silent success")
						}
					} else {
						if err != nil || value == nil || value.Identity == nil || *value.Identity != "Global" {
							t.Fatal("PATCH did not return the response's typed identity")
						}
						if value.AllowEmailIntoChannel == nil || *value.AllowEmailIntoChannel || value.AllowOrganizationTab == nil || !*value.AllowOrganizationTab {
							t.Fatal("typed response lost false or non-default settings")
						}
						if value.RestrictedSenderList == nil || *value.RestrictedSenderList != "" {
							t.Fatal("response was replaced with the submitted null setting")
						}
						// encoding/json passes the object without outer whitespace to
						// UnmarshalJSON; the HTTP receipt above keeps every byte.
						if !bytes.Equal(value.RawJSON, bytes.TrimSpace(tc.body)) || value.RawFields["Key"] == nil {
							t.Fatal("typed response lost its complete received snapshot")
						}
					}
					return contracttest.Observation{Value: value, Response: response, Err: err}
				},
			})
		})
	}
}
