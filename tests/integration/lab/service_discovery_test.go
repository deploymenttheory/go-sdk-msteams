package lab_test

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func TestAccTeamsServiceDiscovery(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, _, _ := labkit.SDKClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	var result struct {
		Endpoints struct {
			Admin string `json:"AdminServiceEndpoint"`
		} `json:"Endpoints"`
	}
	response, err := sdk.TeamsAdminAPI.Scope().Do(ctx, client.Request{
		Method: "GET", Path: []string{"Teams.Tenant", "serviceDiscovery"}, Replay: client.ReplaySafe,
	}, &result)
	if err != nil || response == nil || response.StatusCode() != 200 || result.Endpoints.Admin == "" {
		t.Fatal("Teams service discovery did not return an administration endpoint")
	}
	labkit.WriteResponseFixture(t, "internal/configapi", "service_discovery", response, nil)
}
