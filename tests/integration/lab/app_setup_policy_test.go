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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_setup_policy"
)

type appSetupObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

var appSetupCollections = []string{"AppPresetList", "PinnedAppBarApps", "PinnedMessageBarApps", "AppPresetMeetingList", "AdditionalCustomizationApps", "PinnedCallingBarApps"}

func decodeAppSetupOracle(raw []byte, tenant string) (appSetupObservation, error) {
	var v appSetupObservation
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || !strings.EqualFold(v.Tenant, tenant) || !strings.EqualFold(v.Identity, "Global") || len(v.Settings) != 9 {
		return appSetupObservation{}, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return appSetupObservation{}, errTeamOracle
	}
	var description *string
	body, ok := v.Settings["Description"]
	if !ok || json.Unmarshal(body, &description) != nil {
		return appSetupObservation{}, errTeamOracle
	}
	for _, key := range []string{"AllowSideLoading", "AllowUserPinning"} {
		var value *bool
		body, ok := v.Settings[key]
		if !ok || json.Unmarshal(body, &value) != nil {
			return appSetupObservation{}, errTeamOracle
		}
	}
	for _, key := range appSetupCollections {
		var items []map[string]json.RawMessage
		body, ok := v.Settings[key]
		if !ok || json.Unmarshal(body, &items) != nil || items == nil {
			return appSetupObservation{}, errTeamOracle
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
				return appSetupObservation{}, errTeamOracle
			}
			for _, name := range []string{"Identity", "Id"} {
				var value *string
				raw, ok := item[name]
				if !ok || json.Unmarshal(raw, &value) != nil {
					return appSetupObservation{}, errTeamOracle
				}
			}
			if key == "PinnedAppBarApps" || key == "PinnedMessageBarApps" {
				var order int64
				raw, ok := item["Order"]
				if !ok || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) || json.Unmarshal(raw, &order) != nil {
					return appSetupObservation{}, errTeamOracle
				}
			}
			if key == "AdditionalCustomizationApps" {
				var value *string
				raw, ok := item["AdditionalCustomizationId"]
				if !ok || json.Unmarshal(raw, &value) != nil {
					return appSetupObservation{}, errTeamOracle
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
	t.Helper()
	input, e := json.Marshal(map[string]string{"tenant_id": tenant})
	if e != nil {
		t.Fatal("invalid setup oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-app-setup-policy.ps1")
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
		t.Fatal("pinned setup read failed; no output retained")
	}
	value, e := decodeAppSetupOracle(output.data, tenant)
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
				return nil, errTeamOracle
			}
			found = v
		}
	}
	if found == nil {
		return nil, errTeamOracle
	}
	return found, nil
}

// TestLabAppSetupPolicyReadOnly reads only Global and a filtered list. It never
// changes policies, app layouts, assignments, app installation or directory roles.
func TestLabAppSetupPolicyReadOnly(t *testing.T) {
	cfg, e := lab.Load()
	if e != nil {
		t.Fatal("invalid lab enablement")
	}
	if e := cfg.Require(lab.ReadOnly); errors.Is(e, lab.ErrDisabled) {
		t.Skip("setup policy reads require explicit read-only opt-in")
	} else if e != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, e := sdk.TeamsAdminAPI.TeamsAppSetupPolicies.GetByIdentity(ctx, "Global")
	if e != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("app setup Global read failed (HTTP %d); no response payload retained", status)
	}
	listed, _, e := sdk.TeamsAdminAPI.TeamsAppSetupPolicies.ListFiltered(ctx, "Global")
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
func TestAppSetupOracleRejectsWrongTenantMixedOutputAndOrderOverflow(t *testing.T) {
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
