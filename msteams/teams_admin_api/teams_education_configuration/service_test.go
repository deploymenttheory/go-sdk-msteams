package teamseducationconfiguration_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_education_configuration"
)

// These controlled 400 request probes establish wire bytes, not successful decoding.
func TestUnitEducationConfigurationRejectedProviderRequests(t *testing.T) {
	t.Parallel()

	for _, f := range requestInputs(t) {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			x := f.Request
			var body json.RawMessage
			if x.Body != "" {
				body = json.RawMessage(x.Body)
			}
			c := testutil.Case{Name: f.Name + "-" + string(mode), Mode: mode, Steps: []testutil.Exchange{compiledWire(x.Method, strings.TrimPrefix(x.Url, origin), x.CmdletName, body, f.Status, f.Response)}}
			if x.Method == "" {
				c.Steps = nil

				c.Want.Error = resource.ErrInput
			} else {
				if f.Status != 400 {
					t.Fatal("rejected request probe unexpectedly contains a success contract")
				}
				c.Want = testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				s := sdk.TeamsAdminAPI.TeamsEducationConfigurations
				controls := policy.Controls{}
				if f.ForceMode >= 0 {
					controls.Force = new(f.ForceMode == 1)
				}
				switch f.Operation {
				case "read":
					v, r, e := s.GetByIdentity(ctx, scoped)
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "list":
					v, r, e := s.List(ctx)
					return testutil.Observation{Value: v, Response: r, Err: e}

				}
				var settings resource.Settings
				if e := json.Unmarshal([]byte(f.Input), &settings); e != nil {
					return testutil.Observation{Err: e}
				}

				v, r, e := s.PatchByIdentity(ctx, scoped, resource.PatchRequest{Settings: settings, Controls: controls, BodyIdentity: policy.Set("read-model-identity")})
				return testutil.Observation{Value: v, Response: r, Err: e}
			}
			run(t, c)
		}
	}
}
