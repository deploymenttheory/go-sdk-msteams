package regionacc

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccMultiGeoRegionReadOnly requires a separately provisioned delegated
// scope-specific token file. It cannot reuse the dedicated application identity.
// No token is obtained from an ambient browser or application-mode fallback.
func TestAccMultiGeoRegionReadOnly(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	path := os.Getenv("MSTEAMS_LAB_REGION_FILE")
	if path == "" {
		t.Skip("region read requires an explicitly provisioned private delegated scope-token fixture")
	}
	_, binding, _ := labkit.SDKClient(t)
	raw, err := labkit.PrivateFile(path)
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
	if !ok || !strings.EqualFold(fixture.TenantID, tenant) || !labkit.DirectoryID.MatchString(fixture.IdentityID) || fixture.Cloud != "commercial" && fixture.Cloud != "gcc" || fixture.EntityID == "" || fixture.EntityType != "User" && fixture.EntityType != "Group" {
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
