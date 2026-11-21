package teamseducationconfiguration_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_education_configuration"
)

func TestUnitEducationConfigurationPinnedProviderRequests(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, f := range captures(t) {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			x := f.Capture
			var body json.RawMessage
			if x.Body != "" {
				body = json.RawMessage(x.Body)
			}
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: x.CmdletName, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{compiledWire(x.Method, strings.TrimPrefix(x.Url, origin), x.CmdletName, body, f.Status, f.Response)}}
			if x.Sends == 0 {
				c.Steps = nil
				c.Mappings = []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}
				c.Want.Error = resource.ErrInput
			} else if x.FailureType != nil {
				c.Want.Error = client.ErrDecode
				if f.Status == 403 {
					c.Want.Error = auth.ErrForbidden
				}
			} else if f.Response != "" {
				c.Want.JSON = json.RawMessage(f.Response)
			} else if f.Operation == "read" {
				c.Want.JSON = json.RawMessage("null")
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TeamsEducationConfigurations
				controls := policy.Controls{}
				if f.ForceMode >= 0 {
					controls.Force = new(f.ForceMode == 1)
				}
				switch f.Operation {
				case "read":
					v, r, e := s.GetByIdentity(ctx, scoped)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "list":
					v, r, e := s.List(ctx)
					return contracttest.Observation{Value: v, Response: r, Err: e}

				}
				var settings resource.Settings
				if e := json.Unmarshal([]byte(f.Input), &settings); e != nil {
					return contracttest.Observation{Err: e}
				}

				v, r, e := s.PatchByIdentity(ctx, scoped, resource.PatchRequest{Settings: settings, Controls: controls, BodyIdentity: policy.Set("read-model-identity")})
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}
			results = append(results, run(t, c))
		}
	}
	evidence(t, results)
}
