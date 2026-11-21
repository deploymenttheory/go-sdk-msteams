package lab_test

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

type privateMigrationObservation struct {
	TenantID *string `json:"tenant_id"`
	Status   *string `json:"status"`
	Details  *string `json:"details"`
}

// TestLabBatchDeploymentStatusReadOnly reads one explicitly selected existing
// orchestration. It never submits a deployment or sends notification email.
func TestLabBatchDeploymentStatusReadOnly(t *testing.T) {
	cfg := targetingReadOnly(t)
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	raw, err := privateLabFile(os.Getenv("MSTEAMS_LAB_BATCH_STATUS_FILE"))
	if err != nil {
		t.Fatal("explicit private batch-status fixture required")
	}
	defer clear(raw)
	var fixture struct {
		TenantID        string `json:"tenant_id"`
		OrchestrationID string `json:"orchestration_id"`
	}
	tenant, _, ok := strings.Cut(binding, "|")
	if json.Unmarshal(raw, &fixture) != nil || !ok || !strings.EqualFold(fixture.TenantID, tenant) || strings.TrimSpace(fixture.OrchestrationID) == "" {
		t.Fatal("invalid batch-status fixture binding")
	}
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, _, err := sdk.TeamsAdminAPI.BatchDeployments.GetByID(ctx, fixture.OrchestrationID)
	if err != nil || value == nil {
		t.Fatal("selected batch status read failed")
	}
	if oracle {
		actual := runTeamOracle(t, ctx, teamOracleRequest{Kind: "batch-status", ID: tenant, OrchestrationID: fixture.OrchestrationID}).BatchStatus
		if actual == nil || *actual != value.Text {
			t.Fatal("PowerShell and Go batch status text differs")
		}
	}
	t.Log("selected raw batch status read observed; text does not establish deployment success")
}

func TestLabPrivateChannelMigrationReadOnly(t *testing.T) {
	cfg := targetingReadOnly(t)
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, _, err := sdk.TeamsAdminAPI.PrivateChannelMigration.Get(ctx)
	if err != nil || value == nil {
		t.Fatal("private-channel migration status read failed")
	}
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok || value.TenantID == nil || !strings.EqualFold(*value.TenantID, tenant) {
		t.Fatal("migration status tenant does not match selected identity")
	}
	if oracle {
		actual := runTeamOracle(t, ctx, teamOracleRequest{Kind: "private-migration", ID: tenant}).PrivateMigration
		want := privateMigrationObservation{TenantID: value.TenantID, Status: value.MigrationStatus, Details: value.Details}
		if actual == nil || !reflect.DeepEqual(want, *actual) {
			t.Fatal("PowerShell and Go selected private migration fields differ")
		}
	}
	t.Log("tenant private-channel migration observation read; completion/permissions remain separately qualified")
}

// TestLabMultiGeoRegionReadOnly requires a separately provisioned delegated
// scope-specific token file. It cannot reuse the dedicated application identity.
// No token is obtained from an ambient browser or application-mode fallback.
func TestLabMultiGeoRegionReadOnly(t *testing.T) {
	cfg := targetingReadOnly(t)
	path := os.Getenv("MSTEAMS_LAB_REGION_FILE")
	if path == "" {
		t.Skip("region read requires an explicitly provisioned private delegated scope-token fixture")
	}
	_, binding, _ := labSDKClient(t)
	raw, err := privateLabFile(path)
	if err != nil {
		t.Fatal("private delegated region fixture unavailable")
	}
	defer clear(raw)
	var fixture struct {
		TenantID   string                `json:"tenant_id"`
		IdentityID string                `json:"identity_id"`
		Cloud      string                `json:"cloud"`
		Tokens     map[string]auth.Token `json:"tokens"`
		EntityID   string                `json:"entity_id"`
		EntityType string                `json:"entity_type"`
	}
	if json.Unmarshal(raw, &fixture) != nil {
		t.Fatal("invalid delegated region fixture")
	}
	defer clear(fixture.Tokens)
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok || !strings.EqualFold(fixture.TenantID, tenant) || !labDirectoryID.MatchString(fixture.IdentityID) || fixture.Cloud != "commercial" && fixture.Cloud != "gcc" || fixture.EntityID == "" || fixture.EntityType != "User" && fixture.EntityType != "Group" {
		t.Fatal("invalid delegated region binding/selector")
	}
	const audience = "https://migrationsvc.teams.microsoft.com"
	request := auth.TokenRequest{Audience: audience, Scope: audience + "/Region.Read.All"}
	if len(fixture.Tokens) != 1 || fixture.Tokens[request.TokenKey()].AccessToken == "" {
		t.Fatal("exact scope-keyed region token is required")
	}
	sdk, err := msteams.NewClient(&auth.Config{Method: auth.SuppliedTokens, TenantID: fixture.TenantID, Cloud: fixture.Cloud, IdentityID: fixture.IdentityID, TokenMode: auth.Delegated, AccessTokens: fixture.Tokens})
	if err != nil {
		t.Fatal("delegated region SDK configuration failed")
	}
	t.Cleanup(func() { _ = sdk.Close() })
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	if fixture.EntityType == "User" {
		value, _, err := sdk.TeamsAdminAPI.Regions.GetForUser(ctx, fixture.EntityID)
		if err != nil || value == nil || value.Region == nil {
			t.Fatal("selected user region read failed")
		}
	} else {
		value, _, err := sdk.TeamsAdminAPI.Regions.GetForGroup(ctx, fixture.EntityID)
		if err != nil || value == nil || value.Region == nil {
			t.Fatal("selected group region read failed")
		}
	}
	t.Log("selected delegated Multi-Geo region observed; no relocation or directory mutation performed")
}
