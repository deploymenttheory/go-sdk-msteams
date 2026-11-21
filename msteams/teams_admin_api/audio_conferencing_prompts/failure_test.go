package audioconferencingprompts_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/audio_conferencing_prompts"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestUnitPromptsInputContextAndBackendRejectBeforeHTTP(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "local-input-context-backend", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts
		for _, id := range []string{"", " ", ".", "..", "scope/", "/empty", "scope/..", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid identity accepted")}
			}
		}
		for _, pattern := range []string{"", "[broken"} {
			if _, _, e := s.ListFiltered(ctx, pattern); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid filter accepted")}
			}
		}
		for _, input := range []resource.PatchRequest{{}, {Controls: policy.Controls{Force: new(true)}}, {Settings: resource.Settings{Prompts: policy.Null[[]*resource.Prompt]()}}, {Settings: resource.Settings{Packages: policy.Null[[]*resource.PromptPackage]()}}, {Settings: resource.Settings{Prompts: policy.Set([]*resource.Prompt{{ID: new("invalid-guid")}})}}} {
			if _, _, e := s.PatchByIdentity(ctx, "Global", input); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid patch acquired credentials")}
			}
		}
		for _, input := range []resource.UpdateRequest{{}, {Controls: policy.Controls{Force: new(false)}}, {Prompts: new(resource.PromptEdit)}, {Packages: new(resource.PackageEdit)}, {Prompts: resource.ReplacePrompts(&resource.Prompt{ID: new("invalid-guid")})}} {
			if _, _, e := s.UpdateByIdentity(ctx, "Global", input); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid update performed a read")}
			}
		}
		graph, e := client.NewScope(client.Graph, sdk.TeamsAdminAPI.Scope().Session())
		if e != nil {
			return contracttest.Observation{Err: e}
		}
		if _, _, e := resource.NewService(graph).List(ctx); !errors.Is(e, client.ErrDestination) {
			return contracttest.Observation{Err: errors.New("wrong backend accepted")}
		}
		for _, missing := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := missing.List(ctx); !errors.Is(e, client.ErrClosed) {
				return contracttest.Observation{Err: errors.New("missing service accepted")}
			}
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.List(canceled); !errors.Is(e, context.Canceled) {
			return contracttest.Observation{Err: errors.New("cancellation lost")}
		}
		return contracttest.Observation{}
	}})
}

func TestUnitPromptsPatchPreservesStatusAndNeverReplays(t *testing.T) {
	t.Parallel()

	for _, status := range []int{400, 401, 403, 429, 503} {
		run(t, contracttest.Case{Name: fmt.Sprintf("patch-error-%d", status), Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{wire("PATCH", item, setCommand, json.RawMessage(`{"Identity":"Global","Prompts":[]}`), status, `{"Message":"fixture-private"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, r, e := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{Prompts: policy.Set[[]*resource.Prompt](nil)}})
			var api *client.APIError
			if !errors.As(e, &api) || api.StatusCode != status || strings.Contains(fmt.Sprint(e), "fixture-private") {
				return contracttest.Observation{Err: errors.New("status/redaction changed")}
			}
			return contracttest.Observation{Response: r}
		}})
	}
}

func TestUnitPromptsBuildersNeverAcquireCredentialsOrHTTP(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "local-prompt-package-builders", Source: sourceModels, Invoke: func(_ context.Context, _ *msteams.Client) contracttest.Observation {
		prompt, e := resource.NewPrompt(firstID, "", "")
		if e != nil || prompt.Type == nil || *prompt.Type != "" || prompt.Locale == nil || *prompt.Locale != "" {
			return contracttest.Observation{Err: errors.New("prompt builder lost empty values")}
		}
		pkg, e := resource.NewPromptPackage(firstID, "future-locale")
		if e != nil || pkg.DefaultLocale == nil || *pkg.DefaultLocale != "future-locale" {
			return contracttest.Observation{Err: errors.New("package builder lost future locale")}
		}
		if _, e := resource.NewPrompt("invalid-guid", "", ""); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("invalid constructor id accepted")}
		}
		prompts, e := resource.ReplacePrompts(prompt, prompt, nil).Apply(nil)
		if e != nil || len(prompts) != 3 || prompts[2] != nil {
			return contracttest.Observation{Err: errors.New("local prompt edit lost duplicates/null")}
		}
		packages, e := resource.ReplacePackages(pkg, pkg).Apply(nil)
		if e != nil || len(packages) != 2 {
			return contracttest.Observation{Err: errors.New("local package edit changed")}
		}
		return contracttest.Observation{}
	}})
}
