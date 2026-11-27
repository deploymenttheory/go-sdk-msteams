package privacyconfiguration_test

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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/privacy_configuration"
)

// These controlled 400 request probes establish wire bytes, not successful decoding.
func TestUnitPrivacyRejectedProviderRequests(t *testing.T) {
	t.Parallel()

	for _, f := range requestInputs(t) {
		var body json.RawMessage
		if f.Request.Body != "" {
			body = json.RawMessage(f.Request.Body)
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			c := testutil.Case{Name: f.Name + "-" + string(mode), Mode: mode, Steps: []testutil.Exchange{wire(f.Request.Method, strings.TrimPrefix(f.Request.Url, origin), f.Request.CmdletName, body, f.Status, f.Response)}}
			if f.Status != 400 {
				t.Fatal("rejected request probe unexpectedly contains a success contract")
			}
			c.Want = testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				s := sdk.TeamsAdminAPI.PrivacyConfigurations
				if f.Operation == "list" {
					v, r, e := s.List(ctx)
					return testutil.Observation{Value: v, Response: r, Err: e}
				}
				if f.Operation == "patch" {
					location := policy.Set(false)
					if f.NullLocation {
						location = policy.Null[bool]()
					}
					controls := policy.Controls{}
					if f.Force {
						no := false
						controls.Force = &no
					}
					v, r, e := s.PatchByIdentity(ctx, scoped, resource.PatchRequest{Settings: resource.Settings{EnablePrivacyMode: policy.Set(false), AutoInitiateContacts: policy.Set(true), PublishLocationDataDefault: location, DisplayPublishedPhotoDefault: policy.Set(false)}, Controls: controls, BodyIdentity: policy.Set("read-model-identity")})
					return testutil.Observation{Value: v, Response: r, Err: e}
				}
				v, r, e := s.GetByIdentity(ctx, scoped)
				return testutil.Observation{Value: v, Response: r, Err: e}
			}
			run(t, c)
		}
	}
}
