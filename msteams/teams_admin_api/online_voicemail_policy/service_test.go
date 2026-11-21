package onlinevoicemailpolicy_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
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
			} else if f.Operation != "delete" && f.Response != "" {
				c.Want.JSON = json.RawMessage(f.Response)
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.OnlineVoicemailPolicies
				switch f.Operation {
				case "list":
					v, r, e := s.List(ctx)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "create":
					v, r, e := s.Create(ctx, resource.CreateRequest{Identity: scoped, Settings: settings(f.Duration)})
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "patch":
					v, r, e := s.PatchByIdentity(ctx, scoped, resource.PatchRequest{Settings: settings(f.Duration), BodyIdentity: policy.Set("read-model-identity")})
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "delete":
					controls := policy.Controls{}
					if f.Force {
						no := false
						controls.Force = &no
					}
					r, e := s.DeleteByIdentity(ctx, scoped, controls)
					return contracttest.Observation{Response: r, Err: e}
				default:
					v, r, e := s.GetByIdentity(ctx, scoped)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
			}
			results = append(results, run(t, c))
		}
	}
	evidence(t, results)
}
