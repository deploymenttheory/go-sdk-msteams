package audioconferencingprompts_test

import (
	"context"
	"encoding/json"

	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/audio_conferencing_prompts"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestUnitPromptsRejectedProviderRequests(t *testing.T) {
	t.Parallel()

	for _, f := range requestInputs(t) {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			x := f.Request
			var body json.RawMessage
			if x.Body != "" {
				body = json.RawMessage(x.Body)
			}
			c := testutil.Case{Name: f.Name + "-" + string(mode), Mode: mode, Steps: []testutil.Exchange{wire(x.Method, strings.TrimPrefix(x.Url, origin), x.CmdletName, body, f.Status, f.Response)}}
			// These retained request vectors all use a controlled400. They do not
			// assert typed success or malformed-success decoding.
			c.Want = testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				s := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts
				if f.Operation == "read" {
					v, r, e := s.GetByIdentity(ctx, scoped)
					return testutil.Observation{Value: v, Response: r, Err: e}
				}
				if f.Operation == "list" {
					v, r, e := s.List(ctx)
					return testutil.Observation{Value: v, Response: r, Err: e}
				}
				var prompts []*resource.Prompt
				var packages []*resource.PromptPackage
				if f.Fields != "empty-lists" {
					prompts = []*resource.Prompt{{Identity: new("prompt-meta"), ID: new(firstID), Name: new(""), TextPrompt: new("text"), Type: new("FutureType"), Locale: new("xx-future")}}
					packages = []*resource.PromptPackage{{Identity: new("package-meta"), ID: new(secondID), DefaultLocale: new(""), InboundStartRecordingPrompt: []*string{new(firstID), new(firstID), nil}, InboundEndRecordingPrompt: []*string{new("")}, OutboundStartRecordingPrompt: []*string{new("future-reference")}, OutboundEndRecordingPrompt: []*string{new(firstID)}, MeetingStartRecordingPrompt: []*string{new(secondID)}}}
				}
				input := resource.PatchRequest{BodyIdentity: policy.Set("read-model-identity")}
				if f.Fields != "packages" {
					input.Prompts = policy.Set(prompts)
				}
				if f.Fields != "prompts" {
					input.Packages = policy.Set(packages)
				}
				if f.ForceMode >= 0 {
					input.Force = new(f.ForceMode == 1)
				}
				v, r, e := s.PatchByIdentity(ctx, scoped, input)
				return testutil.Observation{Value: v, Response: r, Err: e}
			}
			run(t, c)
		}
	}
}

func TestUnitPromptsRejectedFilterRequests(t *testing.T) {
	t.Parallel()
	// No successful list was observed for this service. Exercise only
	// local filter validation/request construction against a controlled HTTP error.
	for _, pattern := range []string{"*", "Global", "Fixture*"} {
		result, err := testutil.Run(t.Context(), testutil.Case{
			Name:  "request-only-filter",
			Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/TeamsAudioConferencingCustomPromptsConfiguration", 400, nil),
			Want:  testutil.Expected{Status: 400, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				value, response, err := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts.ListFiltered(ctx, pattern)
				return testutil.Observation{Value: value, Response: response, Err: err}
			},
		})
		if err != nil || !result.Passed() {
			t.Fatalf("request-only filter differs: %v %+v", err, result.Differences())
		}
	}

}
