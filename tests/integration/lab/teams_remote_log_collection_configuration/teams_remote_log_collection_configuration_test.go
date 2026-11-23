package teamsremotelogcollectionconfigurationacc

import (
	"context"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const fixtureService = "teams_remote_log_collection_configuration"

func tenantReplacements(t *testing.T, binding string) map[string]string {
	t.Helper()
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok || !labkit.DirectoryID.MatchString(tenant) {
		t.Fatal("tenant binding required")
	}
	return map[string]string{tenant: "00000000-0000-4000-8000-0000000000aa", strings.ToLower(tenant): "00000000-0000-4000-8000-0000000000aa"}
}

// TestAccRemoteLogCollectionConfigurationReads lists the configuration and
// reads Global. It changes nothing.
func TestAccRemoteLogCollectionConfigurationReads(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	replacements := tenantReplacements(t, binding)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	service := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionConfigurations

	listed, response, err := service.List(ctx)
	if err != nil || len(listed) == 0 || listed[0].Identity == nil || *listed[0].Identity != "Global" {
		t.Fatal("remote log collection configuration list failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "list", response, replacements)
	global, response, err := service.GetByIdentity(ctx, "Global")
	if err != nil || global == nil || global.RawFields["Devices"] == nil {
		t.Fatal("global remote log collection configuration read failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "get_global", response, replacements)
	if filtered, _, err := service.ListFiltered(ctx, "Glob*"); err != nil || len(filtered) != 1 {
		t.Fatal("remote log collection configuration filter failed")
	}
	t.Log("remote log collection configuration reads: passed")
}
