package audioconferencingprompts_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/audio_conferencing_prompts"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestPromptsPinnedProviderRequests(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, f := range captures(t) {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			x := f.Capture
			var body json.RawMessage
			if x.Body != "" {
				body = json.RawMessage(x.Body)
			}
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: x.CmdletName, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{wire(x.Method, strings.TrimPrefix(x.Url, origin), x.CmdletName, body, f.Status, f.Response)}}
			if x.FailureType != nil {
				c.Want.Error = client.ErrDecode
				if f.Status == 403 {
					c.Want.Error = auth.ErrForbidden
				}
			} else if f.Response != "" {
				c.Want.JSON = json.RawMessage(f.Response)
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts
				if f.Operation == "read" {
					v, r, e := s.GetByIdentity(ctx, scoped)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				if f.Operation == "list" {
					v, r, e := s.List(ctx)
					return contracttest.Observation{Value: v, Response: r, Err: e}
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
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}
			results = append(results, run(t, c))
		}
	}
	evidence(t, results)
}

func TestPromptsFilterKeepsGlobalTagAndWhitespace(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	const reply = `[{"Identity":"global"},{"Identity":"Tag:Fixture"},{"Identity":"Global"},{"Identity":" "}]`
	for _, pattern := range []string{"", "Fixture*", " "} {
		name, set, want := "list", "Identity", "Global,global,Tag:Fixture, "
		if pattern == "Fixture*" {
			name, set, want = "filter-tag", "Filter", "Tag:Fixture"
		}
		if pattern == " " {
			name, set, want = "filter-whitespace", "Filter", " "
		}
		results = append(results, run(t, contracttest.Case{Name: name, Mappings: []contracttest.Mapping{{Command: getCommand, ParameterSet: set}}, Steps: []contracttest.Exchange{wire("GET", collection, getCommand, nil, 200, reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts
			var o contracttest.Observation
			if pattern == "" {
				v, r, e := s.List(ctx)
				o = contracttest.Observation{Value: v, Response: r, Err: e}
			} else {
				v, r, e := s.ListFiltered(ctx, pattern)
				o = contracttest.Observation{Value: v, Response: r, Err: e}
			}
			values, _ := o.Value.([]*resource.Configuration)
			var ids []string
			for _, v := range values {
				if v == nil || v.Identity == nil {
					ids = append(ids, "missing")
				} else {
					ids = append(ids, *v.Identity)
				}
			}
			if o.Err == nil && strings.Join(ids, ",") != want {
				o.Err = errors.New("prompts filter/order differs")
			}
			return o
		}}))
	}
	evidence(t, results)
}
