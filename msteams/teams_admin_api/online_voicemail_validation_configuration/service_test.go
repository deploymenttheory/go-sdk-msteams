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

func TestValidationPinnedProviderRequests(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, f := range captured(t) {
		var body json.RawMessage
		if f.Capture.Body != "" {
			body = json.RawMessage(f.Capture.Body)
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: f.Capture.CmdletName, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{wire(f.Capture.Method, strings.TrimPrefix(f.Capture.Url, origin), f.Capture.CmdletName, body, f.Status, f.Response)}}
			if f.Capture.FailureType != nil {
				c.Want.Error = client.ErrDecode
				if f.Status == 403 {
					c.Want.Error = auth.ErrForbidden
				}
			} else if f.Response != "" {
				c.Want.JSON = json.RawMessage(f.Response)
			}
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
			results = append(results, run(t, c))
		}
	}
	evidence(t, results)
}
