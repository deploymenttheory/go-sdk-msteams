package audioconferencingprompts_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/audio_conferencing_prompts"
)

func TestUnitPromptsUpdatePreparationFailures(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		read, patch int
		body        string
	}{
		{"read-denied", 403, 0, ""},
		{"read-empty", 204, 0, ""},
		{"read-malformed", 200, 0, "not-json"},
	}
	cases = append(cases, struct {
		name        string
		read, patch int
		body        string
	}{"actual-read-denial", 400, 0, string(readResponseFixture(t, "get_global_400.json"))})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			steps := []contracttest.Exchange{wire("GET", item, setCommand, nil, tc.read, tc.body)}
			if tc.patch != 0 {
				steps = append(steps, wire("PATCH", item, setCommand, json.RawMessage(`{"Identity":"Global","Prompts":[]}`), tc.patch, ""))
			}
			run(t, contracttest.Case{Name: tc.name, Source: "msteams/teams_admin_api/audio_conferencing_prompts/composition_test.go", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Prompts: resource.ClearPrompts()})
				if e == nil || v == nil || v.ReadResponse == nil || v.PatchInvoked != (tc.patch != 0) || v.HTTPAccepted {
					t.Error("failed update lost preparation state or invented acceptance")
				}
				if tc.read == 403 || tc.patch != 0 {
					var apiError *client.APIError
					expected := tc.read
					if tc.patch != 0 {
						expected = tc.patch
					}
					if !errors.As(e, &apiError) || apiError.StatusCode != expected {
						t.Error("failure lost its HTTP status")
					}
				}
				return contracttest.Observation{Response: r}
			}})
		})
	}
}

func TestUnitPromptsEditsStopOnPreparationDenial(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "preparation-denial", Source: "msteams/teams_admin_api/audio_conferencing_prompts/composition_test.go", Steps: []contracttest.Exchange{wire("GET", item, setCommand, nil, 403, "")}, Want: contracttest.Expected{Error: auth.ErrForbidden}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Packages: resource.ChangePackages(nil, []*resource.PromptPackage{{}})})
		if v == nil || v.ReadResponse == nil || v.Before != nil || v.PatchInvoked || v.HTTPAccepted {
			t.Error("failed preparation invented a snapshot, patch or acceptance")
		}
		return contracttest.Observation{Response: r, Err: e}
	}})
}
