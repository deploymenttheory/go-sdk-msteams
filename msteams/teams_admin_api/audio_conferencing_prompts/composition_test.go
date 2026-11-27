package audioconferencingprompts_test

import (
	"context"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/audio_conferencing_prompts"
)

func TestUnitPromptsUpdatePreparationFailures(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name string
		read int
		body string
	}{
		{"read-denied", 403, ""},
		{"read-empty", 204, ""},
		{"read-malformed", 200, "not-json"},
	}
	cases = append(cases, struct {
		name string
		read int
		body string
	}{"actual-read-denial", 400, string(readResponseFixture(t, "get_global_400.json"))})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			steps := []testutil.Exchange{wire("GET", item, setCommand, nil, tc.read, tc.body)}
			run(t, testutil.Case{Name: tc.name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Prompts: resource.ClearPrompts()})
				if e == nil || v == nil || v.ReadResponse == nil || v.PatchInvoked || v.HTTPAccepted {
					t.Error("failed update lost preparation state or invented acceptance")
				}
				if tc.read == 403 {
					var apiError *client.APIError
					expected := tc.read
					if !errors.As(e, &apiError) || apiError.StatusCode != expected {
						t.Error("failure lost its HTTP status")
					}
				}
				return testutil.Observation{Response: r}
			}})
		})
	}
}

func TestUnitPromptsEditsStopOnPreparationDenial(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "preparation-denial", Steps: []testutil.Exchange{wire("GET", item, setCommand, nil, 403, "")}, Want: testutil.Expected{Error: auth.ErrForbidden}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Packages: resource.ChangePackages(nil, []*resource.PromptPackage{{}})})
		if v == nil || v.ReadResponse == nil || v.Before != nil || v.PatchInvoked || v.HTTPAccepted {
			t.Error("failed preparation invented a snapshot, patch or acceptance")
		}
		return testutil.Observation{Response: r, Err: e}
	}})
}
