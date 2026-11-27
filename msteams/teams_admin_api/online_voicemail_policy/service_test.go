package onlinevoicemailpolicy_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func settings(duration string) resource.Settings {
	length := policy.Set(resource.TimeSpanFromTicks(900000000))
	if duration == "tick" {
		length = policy.Set(resource.TimeSpanFromTicks(1))
	}
	if duration == "null" {
		length = policy.Null[resource.TimeSpan]()
	}
	return resource.Settings{Description: policy.Set(""), ShareData: policy.Set("FutureValue"), PrimarySystemPromptLanguage: policy.Set("FuturePrimary"), SecondarySystemPromptLanguage: policy.Set("FutureSecondary"), PreambleAudioFile: policy.Set("fixture-audio-reference"), PostambleAudioFile: policy.Set(""), EnableTranscription: policy.Set(false), EnableTranscriptionProfanityMasking: policy.Set(false), EnableEditingCallAnswerRulesSetting: policy.Set(false), EnableTranscriptionTranslation: policy.Set(false), PreamblePostambleMandatory: policy.Set(false), EnableVoicemailTriage: policy.Set(false), MaximumRecordingLength: length}
}

func TestUnitVoicemailCompiledProviderRequests(t *testing.T) {
	t.Parallel()

	for _, f := range requestInputs(t) {
		var body json.RawMessage
		if f.Request.Body != "" {
			body = json.RawMessage(f.Request.Body)
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			c := testutil.Case{Name: f.Name + "-" + string(mode), Mode: mode, Steps: []testutil.Exchange{wire(f.Request.Method, strings.TrimPrefix(f.Request.Url, origin), f.Request.CmdletName, body, f.Status, f.Response)}}
			if f.Request.FailureType != nil {
				c.Want.Error = client.ErrDecode
				if f.Status == 403 {
					c.Want.Error = auth.ErrForbidden
				}
			} else if f.Operation != "delete" && f.Response != "" {
				c.Want.JSON = json.RawMessage(f.Response)
			}
			if len(c.Steps) != 0 {
				c.Want = testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				s := sdk.TeamsAdminAPI.OnlineVoicemailPolicies
				switch f.Operation {
				case "list":
					v, r, e := s.List(ctx)
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "create":
					v, r, e := s.Create(ctx, resource.CreateRequest{Identity: scoped, Settings: settings(f.Duration)})
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "patch":
					v, r, e := s.PatchByIdentity(ctx, scoped, resource.PatchRequest{Settings: settings(f.Duration), BodyIdentity: policy.Set("read-model-identity")})
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "delete":
					controls := policy.Controls{}
					if f.Force {
						no := false
						controls.Force = &no
					}
					r, e := s.DeleteByIdentity(ctx, scoped, controls)
					return testutil.Observation{Response: r, Err: e}
				default:
					v, r, e := s.GetByIdentity(ctx, scoped)
					return testutil.Observation{Value: v, Response: r, Err: e}
				}
			}
			run(t, c)
		}
	}
}
