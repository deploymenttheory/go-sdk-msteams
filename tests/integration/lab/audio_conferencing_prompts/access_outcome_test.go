package audioconferencingpromptsacc

import (
	"context"
	"errors"
	"slices"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func TestAccAudioConferencingPromptsCapabilityUnavailable(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	if !labkit.OracleEnabled(t) {
		t.Fatal("PowerShell comparison is required for tenant access outcomes")
	}
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	value, response, err := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts.GetByIdentity(ctx, "Global")
	labkit.LogAccessResponse(t, response, err)
	var api *client.APIError
	if value != nil || response == nil || response.StatusCode() != 400 || !errors.Is(err, auth.ErrTenantCapabilityUnavailable) || !errors.As(err, &api) || !slices.Equal(api.ServiceErrorCodes, []int{40003}) {
		t.Fatal("expected typed tenant capability denial changed; successful prompt operations need separate qualification")
	}
	labkit.WriteResponseFixture(t, "audio_conferencing_prompts", "get_global", response, nil)
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok {
		t.Fatal("tenant binding required")
	}
	labkit.AccessOutcomeOracle(t, ctx, tenant, "prompts", "Get-CsTeamsAudioConferencingCustomPromptsConfiguration", "feature-unavailable")
	t.Log("confirmed capability denial: HTTP 400, PolicyRP 40003; PowerShell feature denial matched")
}
