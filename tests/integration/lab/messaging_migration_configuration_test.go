package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	messaging "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_messaging_configuration"
	migration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_migration_configuration"
)

func messagingConfigurationProjection(v *messaging.Configuration) map[string]any {
	return map[string]any{"EnableVideoMessageCaptions": v.EnableVideoMessageCaptions, "EnableInOrganizationChatControl": v.EnableInOrganizationChatControl, "CustomEmojis": v.CustomEmojis, "Storyline": v.Storyline, "Communities": v.Communities, "MessagingNotes": v.MessagingNotes, "FileTypeCheck": v.FileTypeCheck, "UrlReputationCheck": v.UrlReputationCheck, "ContentBasedPhishingCheck": v.ContentBasedPhishingCheck, "ReportIncorrectSecurityDetections": v.ReportIncorrectSecurityDetections, "FileTypeCustomizationSetting": v.FileTypeCustomizationSetting, "FileTypeBlockList": v.FileTypeBlockList}
}
func migrationConfigurationProjection(v *migration.Configuration) map[string]any {
	return map[string]any{"EnableLegacyClientInterop": v.EnableLegacyClientInterop}
}
func configurationReadEnabled(t *testing.T) lab.Config {
	t.Helper()
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid configuration lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("configuration reads require explicit read-only enablement")
	} else if err != nil {
		t.Fatal("invalid configuration lab enablement")
	}
	return cfg
}
func tenantConfigurationOracle(t *testing.T, ctx context.Context, kind, tenant string) map[string]json.RawMessage {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": "Global", "kind": kind})
	if err != nil {
		t.Fatal(err)
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-messaging-migration-configuration.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	for _, key := range []string{"MSTEAMS_LAB_AUTH_METHOD", "MSTEAMS_LAB_MODULE_PATH", "MSTEAMS_LAB_IDENTITY_FILE"} {
		cmd.Env = append(cmd.Env, key+"="+os.Getenv(key))
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1")
	var output teamOracleBuffer
	defer func() { clear(output.data) }()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = bytes.NewReader(input), &output, io.Discard
	if cmd.Run() != nil || output.exceeded {
		t.Fatal("pinned tenant configuration read process failed")
	}
	var result struct {
		Success  bool                       `json:"success"`
		Tenant   string                     `json:"tenant_id"`
		Identity string                     `json:"identity"`
		Settings map[string]json.RawMessage `json:"settings"`
	}
	decoder := json.NewDecoder(bytes.NewReader(output.data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&result) != nil || decoder.Decode(new(any)) != io.EOF || !result.Success || !strings.EqualFold(result.Tenant, tenant) || !strings.EqualFold(result.Identity, "Global") {
		t.Fatal("pinned tenant configuration projection unavailable")
	}
	names := []string{"EnableVideoMessageCaptions", "EnableInOrganizationChatControl", "CustomEmojis", "Storyline", "Communities", "MessagingNotes", "FileTypeCheck", "UrlReputationCheck", "ContentBasedPhishingCheck", "ReportIncorrectSecurityDetections", "FileTypeCustomizationSetting", "FileTypeBlockList"}
	switch kind {
	case "migration":
		names = []string{"EnableLegacyClientInterop"}
	case "mobility":
		names = []string{"Description", "IPAudioMobileMode", "IPVideoMobileMode", "MobileDialerPreference", "LinksInTeams"}
	case "multi-tenant":
		names = []string{"CopilotFromHomeTenant"}
	}
	if len(result.Settings) != len(names) {
		t.Fatal("incomplete configuration getter projection")
	}
	for _, name := range names {
		if _, ok := result.Settings[name]; !ok {
			t.Fatal("missing configuration getter")
		}
	}
	return result.Settings
}
func TestAccMessagingConfigurationReadOnly(t *testing.T) {
	cfg := configurationReadEnabled(t)
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	service := sdk.TeamsAdminAPI.TeamsMessagingConfigurations
	value, response, err := service.GetByIdentity(ctx, "Global")
	writeResponseFixture(t, "teams_messaging_configuration", "get_global", response, nil)
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("messaging configuration Global read failed")
	}
	if value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") {
		t.Fatal("messaging configuration identity differs")
	}
	values, response, err := service.ListFiltered(ctx, "Global")
	writeResponseFixture(t, "teams_messaging_configuration", "list", response, nil)
	var selected *messaging.Configuration
	for _, item := range values {
		if item != nil && item.Identity != nil && strings.EqualFold(*item.Identity, "Global") {
			if selected != nil {
				t.Fatal("ambiguous messaging configuration Global")
			}
			selected = item
		}
	}
	if err != nil || selected == nil || !sameTemplateProjection(messagingConfigurationProjection(value), messagingConfigurationProjection(selected)) {
		t.Fatal("messaging configuration read/list getters differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("configuration tenant binding unavailable")
		}
		actual := tenantConfigurationOracle(t, ctx, "messaging", tenant)
		if !sameTemplateProjection(messagingConfigurationProjection(value), actual) {
			t.Fatal("Go and PowerShell messaging configuration getters differ")
		}
	}
	t.Log("messaging configuration read/list and pinned getter comparison verified")
}
func TestAccMigrationConfigurationReadOnly(t *testing.T) {
	cfg := configurationReadEnabled(t)
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	service := sdk.TeamsAdminAPI.TeamsMigrationConfigurations
	value, response, err := service.GetByIdentity(ctx, "Global")
	writeResponseFixture(t, "teams_migration_configuration", "get_global", response, nil)
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("migration configuration Global read failed")
	}
	if value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") {
		t.Fatal("migration configuration identity differs")
	}
	values, response, err := service.ListFiltered(ctx, "Global")
	writeResponseFixture(t, "teams_migration_configuration", "list", response, nil)
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
	if err != nil || selected == nil || !sameTemplateProjection(migrationConfigurationProjection(value), migrationConfigurationProjection(selected)) {
		t.Fatal("migration configuration read/list getters differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("configuration tenant binding unavailable")
		}
		actual := tenantConfigurationOracle(t, ctx, "migration", tenant)
		if !sameTemplateProjection(migrationConfigurationProjection(value), actual) {
			t.Fatal("Go and PowerShell migration configuration getters differ")
		}
	}
	t.Log("migration configuration read/list and pinned getter comparison verified")
}
