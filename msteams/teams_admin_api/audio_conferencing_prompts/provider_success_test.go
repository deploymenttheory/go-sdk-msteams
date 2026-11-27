package audioconferencingprompts_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/audio_conferencing_prompts"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

const getCommand = "Get-CsTeamsAudioConferencingCustomPromptsConfiguration"

// The fixture is an actual Core getter projection authored locally by the pinned
// probe. It is not a captured successful tenant response or uploaded asset.
func assertAuthoredPrompts(t *testing.T, v *resource.Configuration, raw []byte) {
	t.Helper()
	if v == nil || v.Identity == nil || *v.Identity != "Global" || len(v.Prompts) != 1 || len(v.Packages) != 1 || !bytes.Equal(v.RawJSON, bytes.TrimSpace(raw)) {
		t.Fatal("typed configuration or owned raw projection lost")
	}
	p, q := v.Prompts[0], v.Packages[0]
	if p == nil || p.Identity != nil || p.ID == nil || *p.ID != "00000000-0000-0000-0000-000000000001" || p.Name == nil || *p.Name != "example" || p.AudioPrompt != nil || p.TextPrompt == nil || *p.TextPrompt != "authored text" || p.Type == nil || *p.Type != "FutureType" || p.Locale == nil || *p.Locale != "xx-future" {
		t.Fatal("prompt getter projection lost")
	}
	if q == nil || q.Identity != nil || q.ID == nil || *q.ID != "00000000-0000-0000-0000-000000000002" || q.Name == nil || *q.Name != "example package" || q.DefaultLocale == nil || *q.DefaultLocale != "xx-future" {
		t.Fatal("package getter projection lost")
	}
	for _, refs := range [][]*string{q.InboundStartRecordingPrompt, q.InboundEndRecordingPrompt, q.OutboundStartRecordingPrompt, q.OutboundEndRecordingPrompt, q.MeetingStartRecordingPrompt, q.MeetingEndRecordingPrompt} {
		if !reflect.DeepEqual(refs, []*string{new("00000000-0000-0000-0000-000000000001"), nil}) {
			t.Fatal("nullable package references lost")
		}
	}
}

func TestUnitPromptsAuthoredTypedReadListAndFilters(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "primary_authored_get.json")
	for _, op := range []string{"read", "list", "Glo?al", " "} {
		path, body := collection, "["+string(raw)+"]"
		if op == "read" {
			path, body = item, string(raw)
		}
		run(t, testutil.Case{Name: map[string]string{"read": "read", "list": "list", "Glo?al": "wildcard-filter", " ": "whitespace-filter"}[op], Steps: []testutil.Exchange{wire("GET", path, getCommand, nil, 200, body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts
			if op == "read" {
				v, r, e := s.GetByIdentity(ctx, "Global")
				if e == nil {
					assertAuthoredPrompts(t, v, raw)
				}
				return testutil.Observation{Response: r, Err: e}
			}
			var values []*resource.Configuration
			var r *resty.Response
			var e error
			if op == "list" {
				values, r, e = s.List(ctx)
			} else {
				values, r, e = s.ListFiltered(ctx, op)
			}
			if e == nil {
				if op == " " {
					if len(values) != 0 {
						t.Fatal("whitespace became unfiltered")
					}
				} else {
					if len(values) != 1 {
						t.Fatal("typed list count changed")
					}
					assertAuthoredPrompts(t, values[0], raw)
				}
			}
			return testutil.Observation{Response: r, Err: e}
		}})
	}
}

func TestUnitPromptsAcceptedUpdateRetainsLaterDenialAndCallerRestoration(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "primary_authored_get.json")
	run(t, testutil.Case{Name: "accepted-later-denial-recovery", Steps: []testutil.Exchange{
		wire("GET", item, setCommand, nil, 200, string(raw)), wire("PATCH", item, setCommand, json.RawMessage(`{"Identity":"Global","Prompts":[]}`), 204, ""),
		wire("GET", item, setCommand, nil, 200, string(raw)), wire("PATCH", item, setCommand, json.RawMessage(`{"Identity":"Global","Packages":[]}`), 403, `{"Message":"controlled denial"}`),
		wire("PATCH", item, setCommand, json.RawMessage(raw), 204, ""), wire("GET", item, getCommand, nil, 200, string(raw)),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts
		first, r, e := s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Prompts: resource.ClearPrompts()})
		if e != nil || first == nil || first.ReadResponse == nil || string(first.ReadResponse.Body()) != string(raw) || !first.PatchInvoked || !first.HTTPAccepted || first.After != nil || r == nil || r.StatusCode() != 204 {
			t.Fatal("accepted update lost preparation or invented persistence")
		}
		assertAuthoredPrompts(t, first.Before, raw)
		later, r, e := s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Packages: resource.ClearPackages()})
		var api *client.APIError
		if !errors.As(e, &api) || api.StatusCode != 403 || later == nil || later.Before == nil || later.ReadResponse == nil || !later.PatchInvoked || later.HTTPAccepted || later.After != nil || r == nil || r.StatusCode() != 403 || !first.HTTPAccepted {
			t.Fatal("later patch failure lost partial or earlier accepted receipt")
		}
		assertAuthoredPrompts(t, later.Before, raw)
		assertAuthoredPrompts(t, first.Before, raw)
		if v, r, e := s.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{Prompts: policy.Set(first.Before.Prompts), Packages: policy.Set(first.Before.Packages)}}); e != nil || v != nil || r == nil || r.StatusCode() != 204 {
			t.Fatal("explicit caller restoration failed")
		}
		v, r, e := s.GetByIdentity(ctx, "Global")
		if e == nil {
			assertAuthoredPrompts(t, v, raw)
		}
		return testutil.Observation{Response: r, Err: e}
	}})
}

func TestUnitPromptsNullRemovalFromEmptyListRetainsReadWithoutPatch(t *testing.T) {
	t.Parallel()
	const empty = `{"Identity":"Global","Prompts":[],"Packages":[]}`
	for _, packages := range []bool{false, true} {
		run(t, testutil.Case{Name: "null-remove-empty", Steps: []testutil.Exchange{wire("GET", item, setCommand, nil, 200, empty)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			input := resource.UpdateRequest{Prompts: resource.ChangePrompts(nil, []*resource.Prompt{nil})}
			if packages {
				input = resource.UpdateRequest{Packages: resource.ChangePackages(nil, []*resource.PromptPackage{nil})}
			}
			v, r, e := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts.UpdateByIdentity(ctx, "Global", input)
			if !errors.Is(e, resource.ErrInput) || r == nil || r.StatusCode() != 200 || v == nil || v.Before == nil || v.ReadResponse != r || v.PatchInvoked || v.HTTPAccepted || v.After != nil {
				t.Fatal("null Remove lost source error/preparation or sent PATCH")
			}
			return testutil.Observation{Response: r}
		}})
	}
	p, e := resource.ChangePrompts([]*resource.Prompt{nil}, nil).Apply(nil)
	if e != nil || len(p) != 1 || p[0] != nil {
		t.Fatal("nullable prompt Add rejected")
	}
	q, e := resource.ChangePackages([]*resource.PromptPackage{nil}, nil).Apply(nil)
	if e != nil || len(q) != 1 || q[0] != nil {
		t.Fatal("nullable package Add rejected")
	}
	q, e = resource.ChangePackages(nil, []*resource.PromptPackage{{}}).Apply(nil)
	if e != nil || len(q) != 0 {
		t.Fatal("nonnull package Remove from empty rejected")
	}
}

func TestUnitPromptsPatchTypedAndMalformedAcceptedReplies(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "primary_authored_get.json")
	for _, malformed := range []bool{false, true} {
		body := string(raw)
		if malformed {
			body = `{"Prompts":[{"Id":"invalid-guid"}]}`
		}
		run(t, testutil.Case{Name: "authored-accepted-patch", Steps: []testutil.Exchange{wire("PATCH", item, setCommand, json.RawMessage(`{"Identity":"Global","Prompts":[]}`), 200, body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, r, e := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{Prompts: policy.Set[[]*resource.Prompt](nil)}})
			if r == nil || r.StatusCode() != 200 || string(r.Body()) != body {
				t.Fatal("accepted PATCH lost raw receipt")
			}
			if malformed {
				if !errors.Is(e, client.ErrDecode) || v != nil {
					t.Fatal("malformed accepted PATCH lost decode error or invented value")
				}
				return testutil.Observation{Response: r}
			}
			if e == nil {
				assertAuthoredPrompts(t, v, raw)
			}
			return testutil.Observation{Response: r, Err: e}
		}})
	}
}
