package lab_test

import (
	"context"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	messaging "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_messaging_configuration"
	migration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_migration_configuration"
)

func messagingConfigurationProjection(v *messaging.Configuration) map[string]any {
	return map[string]any{"EnableVideoMessageCaptions": v.EnableVideoMessageCaptions, "EnableInOrganizationChatControl": v.EnableInOrganizationChatControl, "CustomEmojis": v.CustomEmojis, "Storyline": v.Storyline, "Communities": v.Communities, "MessagingNotes": v.MessagingNotes, "FileTypeCheck": v.FileTypeCheck, "UrlReputationCheck": v.UrlReputationCheck, "ContentBasedPhishingCheck": v.ContentBasedPhishingCheck, "ReportIncorrectSecurityDetections": v.ReportIncorrectSecurityDetections, "FileTypeCustomizationSetting": v.FileTypeCustomizationSetting, "FileTypeBlockList": v.FileTypeBlockList}
}
func migrationConfigurationProjection(v *migration.Configuration) map[string]any {
	return map[string]any{"EnableLegacyClientInterop": v.EnableLegacyClientInterop}
}

func TestAccMessagingConfigurationReadOnly(t *testing.T) {
	cfg := labkit.ConfigurationReadEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	service := sdk.TeamsAdminAPI.TeamsMessagingConfigurations
	value, response, err := service.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_messaging_configuration", "get_global", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("messaging configuration Global read failed")
	}
	if value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") {
		t.Fatal("messaging configuration identity differs")
	}
	values, response, err := service.ListFiltered(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_messaging_configuration", "list", response, nil)
	var selected *messaging.Configuration
	for _, item := range values {
		if item != nil && item.Identity != nil && strings.EqualFold(*item.Identity, "Global") {
			if selected != nil {
				t.Fatal("ambiguous messaging configuration Global")
			}
			selected = item
		}
	}
	if err != nil || selected == nil || !labkit.SameProjection(messagingConfigurationProjection(value), messagingConfigurationProjection(selected)) {
		t.Fatal("messaging configuration read/list getters differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("configuration tenant binding unavailable")
		}
		actual := labkit.TenantConfigurationOracle(t, ctx, "messaging", tenant)
		if !labkit.SameProjection(messagingConfigurationProjection(value), actual) {
			t.Fatal("Go and PowerShell messaging configuration getters differ")
		}
	}
	t.Log("messaging configuration read/list and pinned getter comparison verified")
}
func TestAccMigrationConfigurationReadOnly(t *testing.T) {
	cfg := labkit.ConfigurationReadEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	service := sdk.TeamsAdminAPI.TeamsMigrationConfigurations
	value, response, err := service.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_migration_configuration", "get_global", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("migration configuration Global read failed")
	}
	if value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") {
		t.Fatal("migration configuration identity differs")
	}
	values, response, err := service.ListFiltered(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_migration_configuration", "list", response, nil)
	var selected *migration.Configuration
	for i := range values {
		item := &values[i]
		if item.Identity != nil && strings.EqualFold(*item.Identity, "Global") {
			if selected != nil {
				t.Fatal("ambiguous migration configuration Global")
			}
			selected = item
		}
	}
	if err != nil || selected == nil || !labkit.SameProjection(migrationConfigurationProjection(value), migrationConfigurationProjection(selected)) {
		t.Fatal("migration configuration read/list getters differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("configuration tenant binding unavailable")
		}
		actual := labkit.TenantConfigurationOracle(t, ctx, "migration", tenant)
		if !labkit.SameProjection(migrationConfigurationProjection(value), actual) {
			t.Fatal("Go and PowerShell migration configuration getters differ")
		}
	}
	t.Log("migration configuration read/list and pinned getter comparison verified")
}
