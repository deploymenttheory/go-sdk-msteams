package teamsfilespolicy_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_files_policy"
)

func TestUnitFilesPolicyRequestShapes(t *testing.T) {
	t.Parallel()

	for _, f := range requestInputs() {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			var body json.RawMessage
			if len(f.Body) != 0 {
				body = f.Body
			}
			set := "Identity"
			switch f.Operation {
			case "user":
				set = "GrantToUser"
			case "tenant":
				set = "GrantToTenant"
			case "group":
				set = "GrantToGroup"
			}
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: f.Command, ParameterSet: set}}, Steps: []contracttest.Exchange{compiledWire(f.Method, f.Path, f.Command, body, 400, "")}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TeamsFilesPolicies
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
				case "delete":
					v, r, e := s.DeleteByIdentity(ctx, scoped, controls)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "user":
					v, r, e := s.AssignUser(ctx, "user/fixture", f.PolicyName)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "tenant":
					v, r, e := s.AssignTenantDefault(ctx, f.PolicyName, f.ForceMode == 1)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "group":
					v, r, e := s.AssignGroup(ctx, "group/fixture", f.PolicyName, f.Rank)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				var settings resource.Settings
				if e := json.Unmarshal([]byte(f.Input), &settings); e != nil {
					return contracttest.Observation{Err: e}
				}
				if f.Operation == "create" {
					v, r, e := s.CreateByIdentity(ctx, scoped, resource.CreateRequest{Settings: settings, Controls: controls})
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				v, r, e := s.PatchByIdentity(ctx, scoped, resource.PatchRequest{Settings: settings, Controls: controls, BodyIdentity: policy.Set("read-model-identity")})
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}
			run(t, c)
		}
	}
}
