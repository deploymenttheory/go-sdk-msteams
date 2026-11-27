package onlinevoicemailvalidationconfiguration_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_validation_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

// These controlled 400 request probes establish wire bytes, not successful decoding.
func TestUnitValidationRejectedProviderRequests(t *testing.T) {
	t.Parallel()

	for _, f := range requestInputs(t) {
		var body json.RawMessage
		if f.Request.Body != "" {
			body = json.RawMessage(f.Request.Body)
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: f.Request.CmdletName, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{wire(f.Request.Method, strings.TrimPrefix(f.Request.Url, origin), f.Request.CmdletName, body, f.Status, f.Response)}}
			if f.Status != 400 {
				t.Fatal("rejected request probe unexpectedly contains a success contract")
			}
			c.Want = contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations
				if f.Operation == "list" {
					v, r, e := s.List(ctx)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				if f.Operation == "patch" {
					uri := policy.Set("HTTPS://example.invalid/validate")
					if f.NullURI {
						uri = policy.Null[string]()
					}
					controls := policy.Controls{}
					if f.Force {
						no := false
						controls.Force = &no
					}
					v, r, e := s.PatchByIdentity(ctx, scoped, resource.PatchRequest{Settings: resource.Settings{AudioFileValidationEnabled: policy.Set(false), AudioFileValidationURI: uri}, Controls: controls, BodyIdentity: policy.Set("read-model-identity")})
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				v, r, e := s.GetByIdentity(ctx, scoped)
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}
			run(t, c)
		}
	}
}
