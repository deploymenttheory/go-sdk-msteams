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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_setup_policy"
)

type appSetupObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

var appSetupCollections = []string{"AppPresetList", "PinnedAppBarApps", "PinnedMessageBarApps", "AppPresetMeetingList", "AdditionalCustomizationApps", "PinnedCallingBarApps"}

var appSetupOwnedIdentity = regexp.MustCompile(`(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-app-setup-policy$`)

func decodeAppSetupOracle(raw []byte, tenant string) (appSetupObservation, error) {
	return decodeAppSetupOracleForIdentity(raw, tenant, "Global")
}
func decodeAppSetupOracleForIdentity(raw []byte, tenant, identity string) (appSetupObservation, error) {
	var v appSetupObservation
	if len(raw) > labkit.MaxOracleBytes || !labkit.DirectoryID.MatchString(tenant) || !strings.EqualFold(identity, "Global") && !appSetupOwnedIdentity.MatchString(identity) {
		return v, labkit.ErrOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil {
		return appSetupObservation{}, labkit.ErrOracle
	}
	identityMatches := strings.EqualFold(v.Identity, "Global")
	if !strings.EqualFold(identity, "Global") {
		identityMatches = labkit.SamePolicyName(v.Identity, identity)
	}
	if !v.Success || !strings.EqualFold(v.Tenant, tenant) || !identityMatches || len(v.Settings) != 9 {
		return appSetupObservation{}, labkit.ErrOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return appSetupObservation{}, labkit.ErrOracle
	}
	var description *string
	body, ok := v.Settings["Description"]
	if !ok || json.Unmarshal(body, &description) != nil {
		return appSetupObservation{}, labkit.ErrOracle
	}
	for _, key := range []string{"AllowSideLoading", "AllowUserPinning"} {
		var value *bool
		body, ok := v.Settings[key]
		if !ok || json.Unmarshal(body, &value) != nil {
			return appSetupObservation{}, labkit.ErrOracle
		}
	}
	for _, key := range appSetupCollections {
		var items []map[string]json.RawMessage
		body, ok := v.Settings[key]
		if !ok || json.Unmarshal(body, &items) != nil || items == nil {
			return appSetupObservation{}, labkit.ErrOracle
		}
		for _, item := range items {
			if item == nil {
				continue
			}
			count := 2
			if key == "PinnedAppBarApps" || key == "PinnedMessageBarApps" || key == "AdditionalCustomizationApps" {
				count = 3
			}
			if len(item) != count {
				return appSetupObservation{}, labkit.ErrOracle
			}
			for _, name := range []string{"Identity", "Id"} {
				var value *string
				raw, ok := item[name]
				if !ok || json.Unmarshal(raw, &value) != nil {
					return appSetupObservation{}, labkit.ErrOracle
				}
			}
			if key == "PinnedAppBarApps" || key == "PinnedMessageBarApps" {
				var order int64
				raw, ok := item["Order"]
				if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &order) != nil {
					return appSetupObservation{}, labkit.ErrOracle
				}
			}
			if key == "AdditionalCustomizationApps" {
				var value *string
				raw, ok := item["AdditionalCustomizationId"]
				if !ok || json.Unmarshal(raw, &value) != nil {
					return appSetupObservation{}, labkit.ErrOracle
				}
			}
		}
	}
	return v, nil
}
func appSetupProjection(v *resource.Policy) []byte {
	apps := func(items []*resource.App) []map[string]any {
		out := make([]map[string]any, len(items))
		for i, x := range items {
			if x != nil {
				out[i] = map[string]any{"Identity": x.Identity, "Id": x.ID}
			}
		}
		return out
	}
	pins := func(items []*resource.PinnedApp) []map[string]any {
		out := make([]map[string]any, len(items))
		for i, x := range items {
			if x != nil {
				out[i] = map[string]any{"Identity": x.Identity, "Id": x.ID, "Order": x.Order}
			}
		}
		return out
	}
	customs := func(items []*resource.CustomizationApp) []map[string]any {
		out := make([]map[string]any, len(items))
		for i, x := range items {
			if x != nil {
				out[i] = map[string]any{"Identity": x.Identity, "Id": x.ID, "AdditionalCustomizationId": x.AdditionalCustomizationID}
			}
		}
		return out
	}
	raw, _ := json.Marshal(map[string]any{"Description": v.Description, "AllowSideLoading": v.AllowSideLoading, "AllowUserPinning": v.AllowUserPinning, "AppPresetList": apps(v.AppPresetList), "PinnedAppBarApps": pins(v.PinnedAppBarApps), "PinnedMessageBarApps": pins(v.PinnedMessageBarApps), "AppPresetMeetingList": apps(v.AppPresetMeetingList), "AdditionalCustomizationApps": customs(v.AdditionalCustomizationApps), "PinnedCallingBarApps": apps(v.PinnedCallingBarApps)})
	return raw
}
func appSetupOracle(t *testing.T, ctx context.Context, tenant string) appSetupObservation {
	return appSetupOracleForIdentity(t, ctx, tenant, "Global")
}
func appSetupOracleForIdentity(t *testing.T, ctx context.Context, tenant, identity string) appSetupObservation {
	t.Helper()
	input, e := json.Marshal(map[string]string{"tenant_id": tenant, "identity": identity})
	if e != nil {
		t.Fatal("invalid setup oracle selector")
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-app-setup-policy.ps1")
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
		t.Fatal("pinned setup read failed; no output retained")
	}
	value, e := decodeAppSetupOracleForIdentity(output.Data, tenant, identity)
	if e != nil {
		t.Fatal("pinned setup projection unavailable")
	}
	return value
}
func globalAppSetupPolicy(values []*resource.Policy) (*resource.Policy, error) {
	var found *resource.Policy
	for _, v := range values {
		if v != nil && v.Identity != nil && strings.EqualFold(*v.Identity, "Global") {
			if found != nil {
				return nil, labkit.ErrOracle
			}
			found = v
		}
	}
	if found == nil {
		return nil, labkit.ErrOracle
	}
	return found, nil
}

// TestAccAppSetupPolicyReadOnly reads only Global and a filtered list. It never
// changes policies, app layouts, assignments, app installation or directory roles.
func TestAccAppSetupPolicyReadOnly(t *testing.T) {
	cfg, e := lab.Load()
	if e != nil {
		t.Fatal("invalid lab enablement")
	}
	if e := cfg.Require(lab.ReadOnly); errors.Is(e, lab.ErrDisabled) {
		t.Skip("setup policy reads require explicit read-only opt-in")
	} else if e != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, e := sdk.TeamsAdminAPI.TeamsAppSetupPolicies.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_app_setup_policy", "get_global", response, nil)
	if e != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("app setup Global read failed (HTTP %d); no response payload retained", status)
	}
	listed, listResponse, e := sdk.TeamsAdminAPI.TeamsAppSetupPolicies.ListFiltered(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_app_setup_policy", "list", listResponse, nil)
	selected, selectionErr := globalAppSetupPolicy(listed)
	if e != nil || selectionErr != nil || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || !bytes.Equal(appSetupProjection(value), appSetupProjection(selected)) {
		t.Fatal("setup Global read/list projections differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := appSetupOracle(t, ctx, tenant)
		raw, _ := json.Marshal(actual.Settings)
		left, right := json.NewDecoder(bytes.NewReader(appSetupProjection(value))), json.NewDecoder(bytes.NewReader(raw))
		left.UseNumber()
		right.UseNumber()
		var a, b any
		if left.Decode(&a) != nil || right.Decode(&b) != nil || !reflect.DeepEqual(a, b) {
			t.Fatal("Go and pinned setup projections differ")
		}
	}
	t.Log("selected app setup reads verified; writes, assignments, cleanup and effective layout remain separately qualified")
}
func TestUnitAppSetupOracleRejectsWrongTenantMixedOutputAndOrderOverflow(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := `{"Description":null,"AllowSideLoading":false,"AllowUserPinning":null,"AppPresetList":[],"PinnedAppBarApps":[{"Identity":null,"Id":null,"Order":9223372036854775807}],"PinnedMessageBarApps":[],"AppPresetMeetingList":[],"AdditionalCustomizationApps":[],"PinnedCallingBarApps":[]}`
	valid := `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":` + settings + `}`
	for _, raw := range []string{`{"success":false}`, valid + "{}", strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, `"identity":"Global"`, `"identity":"Other"`, 1), strings.Replace(valid, `"Description":null,`, "", 1), strings.Replace(valid, `"Id":null`, `"Id":false`, 1), strings.Replace(valid, "9223372036854775807", "9223372036854775808", 1), strings.Replace(valid, "9223372036854775807", "null", 1)} {
		if _, e := decodeAppSetupOracle([]byte(raw), tenant); e == nil {
			t.Fatal("invalid setup oracle output accepted")
		}
	}
	if _, e := decodeAppSetupOracle([]byte(valid), tenant); e != nil {
		t.Fatal("complete Int64 setup projection rejected")
	}
	global := &resource.Policy{Identity: new("Global")}
	tagged := &resource.Policy{Identity: new("Tag:Global")}
	if selected, e := globalAppSetupPolicy([]*resource.Policy{tagged, global}); e != nil || selected != global {
		t.Fatal("tagged filter match hid Global")
	}
	if _, e := globalAppSetupPolicy([]*resource.Policy{global, {Identity: new("global")}}); e == nil {
		t.Fatal("ambiguous Global accepted")
	}
}

func TestUnitAppSetupOracleRestrictsOwnedIdentity(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	const name = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-app-setup-policy"
	raw, err := os.ReadFile(labkit.Root + "/msteams/teams_admin_api/teams_app_setup_policy/testdata/get_global.json")
	if err != nil {
		t.Fatal(err)
	}
	var value resource.Policy
	if json.Unmarshal(raw, &value) != nil {
		t.Fatal("invalid actual policy fixture")
	}
	var settings map[string]json.RawMessage
	if json.Unmarshal(appSetupProjection(&value), &settings) != nil {
		t.Fatal("invalid canonical projection")
	}
	for _, tc := range []struct {
		selected, returned string
		valid              bool
	}{
		{"Global", "global", true}, {name, "Tag:" + name, true}, {"Tag:" + name, name, true},
		{"other-policy", "other-policy", false}, {name, strings.Replace(name, "aaaa", "bbbb", 1), false},
		{"Global", "Tag:Global", false}, {strings.Replace(name, "app-setup-policy", "calling-policy", 1), name, false},
	} {
		body, err := json.Marshal(map[string]any{"success": true, "tenant_id": tenant, "identity": tc.returned, "settings": settings})
		if err != nil {
			t.Fatal(err)
		}
		_, err = decodeAppSetupOracleForIdentity(body, tenant, tc.selected)
		if (err == nil) != tc.valid {
			t.Fatal("oracle accepted a foreign policy or rejected its owned selection")
		}
	}
}
