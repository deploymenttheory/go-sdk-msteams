package customappsettingsacc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	appcatalog "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/app_catalog"
	teamsartifact "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/teams_artifact"
	customappsettings "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/custom_app_settings"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// probeApp is a reserved fixture-form ID that cannot name a real catalog app.
const probeApp = "00000000-0000-4000-8000-000000000001"

// emptyZip is a valid zip archive with no entries: never a usable app package.
var emptyZip = []byte{0x50, 0x4b, 0x05, 0x06, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0, 0}

type artifactObservation struct {
	Outcome        string `json:"outcome"`
	Count          int    `json:"count"`
	Errors         int    `json:"errors"`
	SitesForbidden bool   `json:"sites_forbidden"`
}

type appCatalogObservation struct {
	Success    bool                `json:"success"`
	Sideloaded bool                `json:"sideloaded"`
	App        string              `json:"app"`
	Artifacts  artifactObservation `json:"artifacts"`
}

func powerShellAppCatalog(t *testing.T, ctx context.Context) appCatalogObservation {
	t.Helper()
	command := labkit.OracleCommand(ctx, "read-app-catalog.ps1")
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
		t.Fatal("pinned PowerShell app catalog read failed; no output retained")
	}
	var observed appCatalogObservation
	decoder := json.NewDecoder(bytes.NewReader(output.Data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&observed) != nil || !observed.Success {
		t.Fatal("pinned PowerShell app catalog projection unavailable")
	}
	return observed
}

// TestAccAppCatalogReads reads the tenant custom app settings, lists the
// organization app catalog and scans SharePoint for meeting artifacts. The lab
// identity lacks AppCatalog.Read.All and Sites.Read.All, so the catalog list
// and the site search are denied; the denials are checked with PowerShell
// parity.
func TestAccAppCatalogReads(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, _, _ := labkit.SDKClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()

	settings, response, err := sdk.TeamsAdminAPI.CustomAppSettings.Get(ctx)
	if err != nil || response.StatusCode() != 200 || settings.IsSideloadedAppsInteractionEnabled == nil || settings.IsAppsEnabled == nil || settings.RawFields["appSettingsList"] == nil {
		t.Fatal("custom app settings read failed or lost fields")
	}
	t.Logf("custom app settings read: %d per-app entries", len(settings.AppSettingsList))

	apps, err := sdk.GraphAPI.AppCatalog.List(ctx, appcatalog.ListOptions{DistributionMethod: "organization"})
	labkit.LogAccessResponse(t, apps.Response, err)
	if len(apps.Items) != 0 || apps.Response == nil || apps.Response.StatusCode() != 403 || !errors.Is(err, auth.ErrForbidden) {
		t.Fatal("expected app catalog authorization denial changed; successful lists need separate qualification")
	}

	result, err := sdk.GraphAPI.TeamsArtifacts.List(ctx, teamsartifact.Options{SharePoint: true})
	if err != nil || len(result.Artifacts) != 0 || len(result.Failures) != 1 || result.Failures[0].Stage != "sites" || !errors.Is(result.Failures[0].Err, auth.ErrForbidden) {
		t.Fatal("expected SharePoint site search denial changed")
	}

	if labkit.OracleEnabled(t) {
		observed := powerShellAppCatalog(t, ctx)
		if observed.Sideloaded != settings.SideloadedAppsInteractionEnabled() || observed.App != "forbidden" {
			t.Fatal("PowerShell and Go app catalog outcomes differ")
		}
		switch observed.Artifacts.Outcome {
		case "scanned":
			if observed.Artifacts.Count != 0 || !observed.Artifacts.SitesForbidden || observed.Artifacts.Errors != len(result.Failures) {
				t.Fatal("PowerShell and Go artifact scan outcomes differ")
			}
		case "skipped":
			t.Log("Get-TeamsArtifacts parity needs client-secret auth; skipped")
		default:
			t.Fatal("unknown PowerShell artifact outcome")
		}
		t.Log("PowerShell 8.1.0 / Go app catalog parity: passed")
	}
	t.Log("confirmed: settings read, app catalog HTTP 403, SharePoint search HTTP 403")
}

// TestAccAppCatalogWrites writes the CURRENT sideloading flag back through the
// module's read-then-replace flow (no state change) and checks the settings
// afterwards. Catalog create/update/delete probes use an empty package and a
// reserved app ID; the lab identity lacks AppCatalog.ReadWrite.All, so each is
// denied before validation and nothing is created.
func TestAccAppCatalogWrites(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, _, _ := labkit.SDKClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	settings := sdk.TeamsAdminAPI.CustomAppSettings

	before, response, err := settings.Get(ctx)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatalf("custom app settings read failed: %v", err)
	}
	response, err = settings.SetSideloadedAppsInteraction(ctx, before.SideloadedAppsInteractionEnabled())
	if err != nil || response.StatusCode() != 204 {
		t.Fatal("same-value sideloading write was not accepted")
	}
	after, _, err := settings.Get(ctx)
	if err != nil {
		t.Fatal("custom app settings re-read failed")
	}
	// appSettingsList mirrors the service-managed app catalog and its length
	// varies between reads independently of this write, so only the switches
	// and branding fields are compared.
	if changed := changedScalars(before, after); len(changed) != 0 {
		t.Fatalf("same-value sideloading write changed the settings: %v", changed)
	}
	t.Log("custom app settings same-value write: HTTP 204, settings unchanged")

	catalog := sdk.GraphAPI.AppCatalog
	if _, response, err = catalog.Create(ctx, appcatalog.CreateRequest{DistributionMethod: "Organization", Package: emptyZip}); response == nil || response.StatusCode() != 403 || !errors.Is(err, auth.ErrForbidden) {
		t.Fatal("app create was not denied")
	}
	if response, err = catalog.UpdateByID(ctx, probeApp, emptyZip); response == nil || response.StatusCode() != 403 || !errors.Is(err, auth.ErrForbidden) {
		t.Fatal("app update was not denied")
	}
	if response, err = catalog.DeleteByID(ctx, probeApp); response == nil || response.StatusCode() != 403 || !errors.Is(err, auth.ErrForbidden) {
		t.Fatal("app delete was not denied")
	}
	t.Log("confirmed: app catalog create/update/delete HTTP 403")
}

// changedScalars names the switch and branding fields that differ; values are
// not reported.
func changedScalars(left, right *customappsettings.Settings) []string {
	var changed []string
	flags := map[string][2]*bool{
		"isAppsEnabled":                      {left.IsAppsEnabled, right.IsAppsEnabled},
		"isAppsPurchaseEnabled":              {left.IsAppsPurchaseEnabled, right.IsAppsPurchaseEnabled},
		"isExternalAppsEnabledByDefault":     {left.IsExternalAppsEnabledByDefault, right.IsExternalAppsEnabledByDefault},
		"isLicenseBasedPinnedAppsEnabled":    {left.IsLicenseBasedPinnedAppsEnabled, right.IsLicenseBasedPinnedAppsEnabled},
		"isSideloadedAppsInteractionEnabled": {left.IsSideloadedAppsInteractionEnabled, right.IsSideloadedAppsInteractionEnabled},
		"isTenantWideAutoInstallEnabled":     {left.IsTenantWideAutoInstallEnabled, right.IsTenantWideAutoInstallEnabled},
	}
	for name, pair := range flags {
		if (pair[0] == nil) != (pair[1] == nil) || pair[0] != nil && *pair[0] != *pair[1] {
			changed = append(changed, name)
		}
	}
	text := map[string][2]*string{
		"lobTextColor": {left.LobTextColor, right.LobTextColor}, "lobBackground": {left.LobBackground, right.LobBackground},
		"lobLogo": {left.LobLogo, right.LobLogo}, "lobLogomark": {left.LobLogomark, right.LobLogomark},
	}
	for name, pair := range text {
		if (pair[0] == nil) != (pair[1] == nil) || pair[0] != nil && *pair[0] != *pair[1] {
			changed = append(changed, name)
		}
	}
	slices.Sort(changed)
	return changed
}
