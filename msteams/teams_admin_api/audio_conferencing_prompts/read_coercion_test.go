package audioconferencingprompts_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/audio_conferencing_prompts"
)

func typedFields(value any) any {
	v := reflect.ValueOf(value)
	if v.IsNil() {
		return nil
	}
	v = v.Elem()
	out := map[string]any{}
	for i := 0; i < v.NumField(); i++ {
		name := v.Type().Field(i).Tag.Get("json")
		if name != "-" {
			out[name] = v.Field(i).Interface()
		}
	}
	return out
}

func typedPrompts(v *resource.Configuration) map[string]any {
	prompts, packages := make([]any, len(v.Prompts)), make([]any, len(v.Packages))
	for i, p := range v.Prompts {
		prompts[i] = typedFields(p)
	}
	for i, p := range v.Packages {
		packages[i] = typedFields(p)
	}
	return map[string]any{"Identity": v.Identity, "Prompts": prompts, "Packages": packages}
}

// Every expected typed projection comes from actual pinned JsonConvert<AudioDTO>,
// the same deserializer called by the provider at IL452. Raw snapshots deliberately
// retain the original tokens instead of rewriting them to the coerced typed view.
func TestUnitPromptsReadMatchesPinnedScalarAndCollectionSemantics(t *testing.T) {
	t.Parallel()
	var fixture struct {
		Observations []struct {
			Name, Input string
			Accepted    bool
			Expected    *string
		}
	}
	if json.Unmarshal(readResponseFixture(t, "primary_read_vectors.json"), &fixture) != nil || len(fixture.Observations) == 0 {
		t.Fatal("invalid source read vectors")
	}
	for _, tc := range fixture.Observations {
		run(t, contracttest.Case{Name: tc.Name, Source: "msteams/teams_admin_api/audio_conferencing_prompts/testdata/primary_read_vectors.json", Mappings: []contracttest.Mapping{{Command: getCommand, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{wire("GET", item, getCommand, nil, 200, tc.Input)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts.GetByIdentity(ctx, "Global")
			if r == nil || r.StatusCode() != 200 || string(r.Body()) != tc.Input {
				t.Fatal("read lost original successful HTTP receipt")
			}
			if !tc.Accepted {
				if !errors.Is(e, client.ErrDecode) || v != nil {
					t.Fatal("source-rejected shape or setter value accepted")
				}
				return contracttest.Observation{Response: r}
			}
			if e != nil || v == nil || tc.Expected == nil {
				t.Fatal("source-accepted read rejected")
			}
			actual, err := json.Marshal(typedPrompts(v))
			var a, b any
			if err != nil || json.Unmarshal(actual, &a) != nil || json.Unmarshal([]byte(*tc.Expected), &b) != nil || !reflect.DeepEqual(a, b) {
				t.Fatalf("typed getter projection differs from pinned source for %s", tc.Name)
			}
			if !bytes.Equal(v.RawJSON, bytes.TrimSpace([]byte(tc.Input))) {
				t.Fatal("typed coercion rewrote raw JSON snapshot")
			}
			return contracttest.Observation{Response: r}
		}})
	}
	p, e := resource.NewPromptPackage("00000000-0000-0000-0000-000000000002", "xx-future")
	if e != nil {
		t.Fatal(e)
	}
	for _, refs := range [][]*string{p.InboundStartRecordingPrompt, p.InboundEndRecordingPrompt, p.OutboundStartRecordingPrompt, p.OutboundEndRecordingPrompt, p.MeetingStartRecordingPrompt, p.MeetingEndRecordingPrompt} {
		if refs == nil || len(refs) != 0 {
			t.Fatal("mapped package constructor lost empty list defaults")
		}
	}
}
