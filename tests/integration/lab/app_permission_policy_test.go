package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

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

var appPermissionOwnedIdentity = regexp.MustCompile(`(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-app-permission-policy$`)

func decodeAppPermissionOracle(raw []byte, tenant string) (appPermissionObservation, error) {
	return decodeAppPermissionOracleForIdentity(raw, tenant, "Global")
}
func decodeAppPermissionOracleForIdentity(raw []byte, tenant, identity string) (appPermissionObservation, error) {
	var v appPermissionObservation
	if len(raw) > labkit.MaxOracleBytes || !labkit.DirectoryID.MatchString(tenant) || !strings.EqualFold(identity, "Global") && !appPermissionOwnedIdentity.MatchString(identity) {
		return v, labkit.ErrOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil {
		return appPermissionObservation{}, labkit.ErrOracle
	}
	identityMatches := strings.EqualFold(v.Identity, "Global")
	if !strings.EqualFold(identity, "Global") {
		identityMatches = labkit.SamePolicyName(v.Identity, identity)
	}
	if !v.Success || !strings.EqualFold(v.Tenant, tenant) || !identityMatches || len(v.Settings) != 7 {
		return appPermissionObservation{}, labkit.ErrOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return appPermissionObservation{}, labkit.ErrOracle
	}
	for _, key := range appPermissionScalars {
		var value *string
		body, ok := v.Settings[key]
		if !ok || json.Unmarshal(body, &value) != nil {
			return appPermissionObservation{}, labkit.ErrOracle
		}
	}
	for _, key := range appPermissionCatalogs {
		var items []map[string]*string
		body, ok := v.Settings[key]
		if !ok || json.Unmarshal(body, &items) != nil || items == nil {
			return appPermissionObservation{}, labkit.ErrOracle
		}
		for _, item := range items {
			if item == nil {
				continue
			}
			if len(item) != 2 {
				return appPermissionObservation{}, labkit.ErrOracle
			}
			if _, ok := item["Identity"]; !ok {
				return appPermissionObservation{}, labkit.ErrOracle
			}
			if _, ok := item["Id"]; !ok {
				return appPermissionObservation{}, labkit.ErrOracle
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
	return appPermissionOracleForIdentity(t, ctx, tenant, "Global")
}
func appPermissionOracleForIdentity(t *testing.T, ctx context.Context, tenant, identity string) appPermissionObservation {
	t.Helper()
	input, e := json.Marshal(map[string]string{"tenant_id": tenant, "identity": identity})
	if e != nil {
		t.Fatal("invalid app permission oracle selector")
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-app-permission-policy.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"), "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned app permission read failed; no output retained")
	}
	v, e := decodeAppPermissionOracleForIdentity(output.Data, tenant, identity)
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
				return nil, labkit.ErrOracle
			}
			found = value
		}
	}
	if found == nil {
		return nil, labkit.ErrOracle
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
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, e := sdk.TeamsAdminAPI.TeamsAppPermissionPolicies.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_app_permission_policy", "get_global", response, nil)
	if e != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("app permission Global read failed (HTTP %d); no response payload retained", status)
	}
	listed, listResponse, e := sdk.TeamsAdminAPI.TeamsAppPermissionPolicies.ListFiltered(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_app_permission_policy", "list", listResponse, nil)
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

func TestUnitAppPermissionOracleRestrictsOwnedIdentity(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	const name = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-app-permission-policy"
	raw, err := os.ReadFile(labkit.Root + "/msteams/teams_admin_api/teams_app_permission_policy/testdata/get_global.json")
	if err != nil {
		t.Fatal(err)
	}
	var value resource.Policy
	if json.Unmarshal(raw, &value) != nil {
		t.Fatal("invalid actual policy fixture")
	}
	var settings map[string]json.RawMessage
	if json.Unmarshal(appPermissionProjection(&value), &settings) != nil {
		t.Fatal("invalid canonical projection")
	}
	for _, tc := range []struct {
		selected, returned string
		valid              bool
	}{
		{"Global", "global", true}, {name, "Tag:" + name, true}, {"Tag:" + name, name, true},
		{"other-policy", "other-policy", false}, {name, strings.Replace(name, "aaaa", "bbbb", 1), false},
		{"Global", "Tag:Global", false}, {strings.Replace(name, "app-permission-policy", "calling-policy", 1), name, false},
	} {
		body, err := json.Marshal(map[string]any{"success": true, "tenant_id": tenant, "identity": tc.returned, "settings": settings})
		if err != nil {
			t.Fatal(err)
		}
		_, err = decodeAppPermissionOracleForIdentity(body, tenant, tc.selected)
		if (err == nil) != tc.valid {
			t.Fatal("oracle accepted a foreign policy or rejected its owned selection")
		}
	}
}
