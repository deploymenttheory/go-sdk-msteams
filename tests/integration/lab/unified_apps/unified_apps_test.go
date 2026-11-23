// Package unifiedappsacc exercises the unified app catalog against the lab.
//
// The lab identity is application-only. Microsoft documents the unified app
// cmdlets as unsupported under application authentication, and the AdminAppCatalog
// service returns 401 for app-only tokens, so the SDK refuses application mode
// locally. Real reads need a separately provisioned delegated token file.
package unifiedappsacc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	ua "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/unified_apps"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const (
	audience = "08ff1ce2-4973-4b08-86a3-ebed13badc7f"
	scope    = audience + "/AdminAppCatalog.ReadWrite.All"
	probeApp = "00000000-0000-4000-8000-000000000001"
)

type oracleObservation struct {
	Success  bool   `json:"success"`
	List     string `json:"list"`
	Pending  string `json:"pending"`
	Settings string `json:"settings"`
}

func powerShellUnifiedApps(t *testing.T, ctx context.Context) oracleObservation {
	t.Helper()
	command := labkit.OracleCommand(ctx, "read-unified-apps.ps1")
	command.WaitDelay = time.Second
	command.Env = []string{}
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			command.Env = append(command.Env, key+"="+value)
		}
	}
	for _, key := range []string{"MSTEAMS_LAB_IDENTITY_FILE", "MSTEAMS_LAB_MODULE_PATH", "MSTEAMS_LAB_AUTH_METHOD"} {
		command.Env = append(command.Env, key+"="+os.Getenv(key))
	}
	command.Env = append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1")
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	command.Stdout, command.Stderr = &output, io.Discard
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned PowerShell unified app read failed; no output retained")
	}
	var observed oracleObservation
	decoder := json.NewDecoder(bytes.NewReader(output.Data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&observed) != nil || !observed.Success {
		t.Fatal("pinned PowerShell unified app projection unavailable")
	}
	return observed
}

// refused reports whether PowerShell refused an application-identity read,
// either locally (no Substrate token) or with the service's 401.
func refused(outcome string) bool { return outcome == "local" || outcome == "401" }

// TestAccUnifiedAppsApplicationIdentityRefused proves every unified app
// operation refuses the lab's application identity before any token request, and
// that pinned PowerShell refuses the same reads.
func TestAccUnifiedAppsApplicationIdentityRefused(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, _, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.UnifiedApps
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()

	calls := map[string]func() error{
		"list":     func() error { _, _, err := service.List(ctx); return err },
		"get":      func() error { _, _, err := service.Get(ctx, probeApp); return err },
		"pending":  func() error { _, _, err := service.ListCustomPending(ctx); return err },
		"settings": func() error { _, _, err := service.GetTenantSettings(ctx); return err },
		"update":   func() error { _, err := service.Update(ctx, probeApp, ua.AppUpdate{IsBlocked: new(bool)}); return err },
		"review":   func() error { _, err := service.ReviewCustomPending(ctx, probeApp, ua.ReviewRejected); return err },
		"settings-update": func() error {
			_, err := service.UpdateTenantSettings(ctx, ua.TenantSettingsUpdate{SettingName: ua.SettingDefaultApp, SettingValue: ua.SettingValueAll})
			return err
		},
	}
	for name, call := range calls {
		if err := call(); !errors.Is(err, client.ErrAuthenticationMode) {
			t.Fatalf("unified app %s did not refuse application mode locally", name)
		}
	}
	t.Log("unified app application-mode refusal: passed")

	if labkit.OracleEnabled(t) {
		observed := powerShellUnifiedApps(t, ctx)
		if !refused(observed.List) || !refused(observed.Pending) || !refused(observed.Settings) {
			t.Fatal("pinned PowerShell allowed an application-identity unified app read")
		}
		t.Log("PowerShell 8.1.0 / Go unified app application refusal parity: passed")
	}
}

// TestAccUnifiedAppsDelegatedReadOnly requires MSTEAMS_LAB_APP_CATALOG_FILE, a
// private delegated token fixture shaped like the region fixture: tenant_id,
// identity_id, cloud ("commercial") and tokens holding exactly the scope-keyed
// AdminAppCatalog token. Without it the test skips; no token is ever obtained
// interactively.
func TestAccUnifiedAppsDelegatedReadOnly(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	path := os.Getenv("MSTEAMS_LAB_APP_CATALOG_FILE")
	if path == "" {
		t.Skip("unified app reads require an explicitly provisioned private delegated scope-token fixture")
	}
	_, binding, _ := labkit.SDKClient(t)
	raw, err := labkit.PrivateFile(path)
	if err != nil {
		t.Fatal("private delegated app catalog fixture unavailable")
	}
	defer clear(raw)
	var fixture struct {
		TenantID   string                `json:"tenant_id"`
		IdentityID string                `json:"identity_id"`
		Cloud      string                `json:"cloud"`
		Tokens     map[string]auth.Token `json:"tokens"`
	}
	if json.Unmarshal(raw, &fixture) != nil {
		t.Fatal("invalid delegated app catalog fixture")
	}
	defer clear(fixture.Tokens)
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok || !strings.EqualFold(fixture.TenantID, tenant) || !labkit.DirectoryID.MatchString(fixture.IdentityID) || fixture.Cloud != "commercial" {
		t.Fatal("invalid delegated app catalog binding")
	}
	request := auth.TokenRequest{Audience: audience, Scope: scope}
	if len(fixture.Tokens) != 1 || fixture.Tokens[request.TokenKey()].AccessToken == "" {
		t.Fatal("exact scope-keyed app catalog token is required")
	}
	sdk, err := msteams.NewClient(&auth.Config{Method: auth.SuppliedTokens, TenantID: fixture.TenantID, Cloud: fixture.Cloud, IdentityID: fixture.IdentityID, TokenMode: auth.Delegated, AccessTokens: fixture.Tokens})
	if err != nil {
		t.Fatal("delegated app catalog SDK configuration failed")
	}
	t.Cleanup(func() { _ = sdk.Close() })
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	service := sdk.TeamsAdminAPI.UnifiedApps
	apps, _, err := service.List(ctx)
	if err != nil || apps == nil {
		t.Fatal("unified app list failed")
	}
	if len(apps) > 0 && apps[0].ID != nil {
		if app, _, err := service.Get(ctx, *apps[0].ID); err != nil || app == nil {
			t.Fatal("unified app get failed")
		}
	}
	if pending, _, err := service.ListCustomPending(ctx); err != nil || pending == nil {
		t.Fatal("custom pending app list failed")
	}
	if settings, _, err := service.GetTenantSettings(ctx); err != nil || settings == nil {
		t.Fatal("unified app tenant settings read failed")
	}
	t.Log("delegated unified app reads: passed; no catalog change performed")
}
