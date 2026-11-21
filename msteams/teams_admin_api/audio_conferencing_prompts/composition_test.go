package audioconferencingprompts_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/audio_conferencing_prompts"
)

func TestUnitPromptsUpdateKeepsPreparationAndHTTPAcceptance(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, c := range []struct {
		name            string
		read, patch     int
		readBody, reply string
		accepted        bool
	}{
		{"success", 200, 204, `{"Identity":"read-model-identity","Prompts":[{}],"Packages":[{}],"Future":false}`, "", true},
		{"read-denied", 403, 0, `{"Message":"fixture-private"}`, "", false},
		{"read-empty", 204, 0, "", "", false},
		{"read-malformed", 200, 0, "not-json", "", false},
		{"patch-throttled", 200, 429, `{"Identity":"read-model-identity"}`, `{"Message":"fixture-private"}`, false},
		{"patch-unavailable", 200, 503, `{"Identity":"read-model-identity"}`, "not-json", false},
		{"accepted-malformed", 200, 200, `{"Identity":"read-model-identity"}`, "not-json", true},
	} {
		steps := []contracttest.Exchange{wire("GET", item, setCommand, nil, c.read, c.readBody)}
		if c.patch != 0 {
			steps = append(steps, wire("PATCH", item, setCommand, json.RawMessage(`{"Identity":"read-model-identity","Prompts":[]}`), c.patch, c.reply))
		}
		results = append(results, run(t, contracttest.Case{Name: "update-" + c.name, Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Prompts: resource.ClearPrompts()})
			if v == nil || v.ReadResponse == nil || v.PatchInvoked != (c.patch != 0) || v.HTTPAccepted != c.accepted {
				return contracttest.Observation{Err: errors.New("preparation/acceptance lost")}
			}
			if c.name == "success" {
				if v.Before == nil || len(v.Before.Prompts) != 1 || len(v.Before.Packages) != 1 || v.Before.RawFields["Future"] == nil || v.After != nil {
					return contracttest.Observation{Err: errors.New("read snapshot changed")}
				}
				return contracttest.Observation{Response: r, Err: e}
			}
			if e == nil || c.accepted && !errors.Is(e, client.ErrDecode) {
				return contracttest.Observation{Err: errors.New("partial failure lost")}
			}
			return contracttest.Observation{Response: r}
		}}))
	}
	evidence(t, results)
}

func TestUnitPromptsEditsFailBeforePatchOrPreserveSourceKeys(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, name := range []string{"package-remove-nonempty", "package-remove-empty", "prompt-remove-first"} {
		read := `{"Identity":"read-model-identity","Packages":[{}],"Future":false}`
		body := `{"Identity":"read-model-identity","Packages":[]}`
		if name == "package-remove-empty" {
			read = `{"Identity":"read-model-identity","Packages":[],"Future":false}`
		}
		if name == "prompt-remove-first" {
			read = `{"Identity":"read-model-identity","Prompts":[{"Id":"11111111-1111-1111-1111-111111111111"}],"prompts":[{"Id":"11111111-1111-1111-1111-111111111111","Future":false}],"Future":false}`
			body = `{"Identity":"read-model-identity","Prompts":[{"Identity":null,"Id":"11111111-1111-1111-1111-111111111111","Name":null,"AudioPrompt":null,"TextPrompt":null,"Type":null,"Locale":null},{"Identity":null,"Id":"22222222-2222-2222-2222-222222222222","Name":null,"AudioPrompt":null,"TextPrompt":null,"Type":null,"Locale":null}]}`
		}
		steps := []contracttest.Exchange{wire("GET", item, setCommand, nil, 200, read)}
		patch := name != "package-remove-nonempty"
		if patch {
			steps = append(steps, wire("PATCH", item, setCommand, json.RawMessage(body), 204, ""))
		}
		results = append(results, run(t, contracttest.Case{Name: name, Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			input := resource.UpdateRequest{Packages: resource.ChangePackages(nil, []*resource.PromptPackage{{}})}
			if name == "prompt-remove-first" {
				input = resource.UpdateRequest{Prompts: resource.ChangePrompts([]*resource.Prompt{{ID: new(secondID)}}, []*resource.Prompt{{Identity: new("ignored"), ID: new(firstID)}})}
			}
			v, r, e := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts.UpdateByIdentity(ctx, "Global", input)
			if v == nil || v.Before == nil || v.Before.RawFields["Future"] == nil || v.PatchInvoked != patch || v.HTTPAccepted != patch {
				return contracttest.Observation{Err: errors.New("source edit preparation changed")}
			}
			if !patch {
				if !errors.Is(e, resource.ErrInput) || len(v.Before.Packages) != 1 {
					return contracttest.Observation{Err: errors.New("nonempty package removal was dispatched or lost snapshot")}
				}
				return contracttest.Observation{Response: r}
			}
			if name == "prompt-remove-first" && len(v.Before.Prompts) != 2 {
				return contracttest.Observation{Err: errors.New("edit mutated preparation snapshot")}
			}
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}
