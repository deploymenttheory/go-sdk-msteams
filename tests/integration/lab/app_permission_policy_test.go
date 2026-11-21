package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_permission_policy"
)

type appPermissionObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

var appPermissionScalars = []string{"Description", "DefaultCatalogAppsType", "GlobalCatalogAppsType", "PrivateCatalogAppsType"}
var appPermissionCatalogs = []string{"DefaultCatalogApps", "GlobalCatalogApps", "PrivateCatalogApps"}

func decodeAppPermissionOracle(raw []byte, tenant string) (appPermissionObservation, error) {
	var v appPermissionObservation
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || !strings.EqualFold(v.Tenant, tenant) || !strings.EqualFold(v.Identity, "Global") || len(v.Settings) != 7 {
		return appPermissionObservation{}, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return appPermissionObservation{}, errTeamOracle
	}
	for _, key := range appPermissionScalars {
		var value *string
		body, ok := v.Settings[key]
		if !ok || json.Unmarshal(body, &value) != nil {
			return appPermissionObservation{}, errTeamOracle
		}
	}
	for _, key := range appPermissionCatalogs {
		var items []map[string]*string
		body, ok := v.Settings[key]
		if !ok || json.Unmarshal(body, &items) != nil || items == nil {
			return appPermissionObservation{}, errTeamOracle
		}
		for _, item := range items {
			if item == nil {
				continue
			}
			if len(item) != 2 {
				return appPermissionObservation{}, errTeamOracle
			}
			if _, ok := item["Identity"]; !ok {
				return appPermissionObservation{}, errTeamOracle
			}
			if _, ok := item["Id"]; !ok {
				return appPermissionObservation{}, errTeamOracle
			}
		}
	}
	return v, nil
}
func appPermissionProjection(v *resource.Policy) []byte {
	project := func(items []*resource.CatalogApp) []map[string]*string {
		out := make([]map[string]*string, len(items))
		for i, x := range items {
			if x != nil {
				out[i] = map[string]*string{"Identity": x.Identity, "Id": x.ID}
			}
		}
		return out
	}
	raw, _ := json.Marshal(map[string]any{"Description": v.Description, "DefaultCatalogAppsType": v.DefaultCatalogAppsType, "GlobalCatalogAppsType": v.GlobalCatalogAppsType, "PrivateCatalogAppsType": v.PrivateCatalogAppsType, "DefaultCatalogApps": project(v.DefaultCatalogApps), "GlobalCatalogApps": project(v.GlobalCatalogApps), "PrivateCatalogApps": project(v.PrivateCatalogApps)})
	return raw
}
func appPermissionOracle(t *testing.T, ctx context.Context, tenant string) appPermissionObservation {
	t.Helper()
	input, e := json.Marshal(map[string]string{"tenant_id": tenant})
	if e != nil {
		t.Fatal("invalid app permission oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-app-permission-policy.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output teamOracleBuffer
	defer func() { clear(output.data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.exceeded {
		t.Fatal("pinned app permission read failed; no output retained")
	}
	v, e := decodeAppPermissionOracle(output.data, tenant)
	if e != nil {
		t.Fatal("pinned app permission projection unavailable")
	}
	return v
}

func globalAppPermissionPolicy(values []*resource.Policy) (*resource.Policy, error) {
	var found *resource.Policy
	for _, value := range values {
		if value != nil && value.Identity != nil && strings.EqualFold(*value.Identity, "Global") {
			if found != nil {
				return nil, errTeamOracle
			}
			found = value
		}
	}
	if found == nil {
		return nil, errTeamOracle
	}
	return found, nil
}

// TestAccAppPermissionPolicyReadOnly never creates or edits policies, grants an
// app policy to a principal, installs apps or changes directory permissions.
func TestAccAppPermissionPolicyReadOnly(t *testing.T) {
	cfg, e := lab.Load()
	if e != nil {
		t.Fatal("invalid lab enablement")
	}
	if e := cfg.Require(lab.ReadOnly); errors.Is(e, lab.ErrDisabled) {
		t.Skip("app permission reads require explicit read-only opt-in")
	} else if e != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, e := sdk.TeamsAdminAPI.TeamsAppPermissionPolicies.GetByIdentity(ctx, "Global")
	if e != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("app permission Global read failed (HTTP %d); no response payload retained", status)
	}
	listed, _, e := sdk.TeamsAdminAPI.TeamsAppPermissionPolicies.ListFiltered(ctx, "Global")
	selected, selectionErr := globalAppPermissionPolicy(listed)
	if e != nil || selectionErr != nil || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || !bytes.Equal(appPermissionProjection(value), appPermissionProjection(selected)) {
		t.Fatal("app permission Global read/list observations differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := appPermissionOracle(t, ctx, tenant)
		var left, right any
		raw, _ := json.Marshal(actual.Settings)
		if json.Unmarshal(appPermissionProjection(value), &left) != nil || json.Unmarshal(raw, &right) != nil || !reflect.DeepEqual(left, right) {
			t.Fatal("Go and pinned app permission projections differ")
		}
	}
	t.Log("selected app permission reads verified; writes, assignments, cleanup and effective behavior remain separately qualified")
}
func TestUnitAppPermissionGlobalSelectionAllowsTaggedMatches(t *testing.T) {
	t.Parallel()
	global := &resource.Policy{Identity: new("Global")}
	tagged := &resource.Policy{Identity: new("Tag:Global")}
	if selected, e := globalAppPermissionPolicy([]*resource.Policy{tagged, global}); e != nil || selected != global {
		t.Fatal("tagged filter match hid Global")
	}
	for _, values := range [][]*resource.Policy{{tagged}, {global, {Identity: new("global")}}} {
		if _, e := globalAppPermissionPolicy(values); e == nil {
			t.Fatal("missing or ambiguous Global accepted")
		}
	}
}
func TestUnitAppPermissionOracleRejectsWrongTenantIncompleteAndMixedOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := `{"Description":null,"DefaultCatalogAppsType":"FutureMode","GlobalCatalogAppsType":"","PrivateCatalogAppsType":null,"DefaultCatalogApps":[{"Identity":null,"Id":null}],"GlobalCatalogApps":[],"PrivateCatalogApps":[]}`
	valid := `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":` + settings + `}`
	for _, raw := range []string{`{"success":false}`, valid + "{}", strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, `"identity":"Global"`, `"identity":"Other"`, 1), strings.Replace(valid, `"Description":null,`, "", 1), strings.Replace(valid, `"Id":null`, `"Id":false`, 1), strings.Replace(valid, `"GlobalCatalogApps":[]`, `"GlobalCatalogApps":null`, 1)} {
		if _, e := decodeAppPermissionOracle([]byte(raw), tenant); e == nil {
			t.Fatal("invalid app permission oracle output accepted")
		}
	}
	if _, e := decodeAppPermissionOracle([]byte(valid), tenant); e != nil {
		t.Fatal("complete nullable app permission projection rejected")
	}
}
