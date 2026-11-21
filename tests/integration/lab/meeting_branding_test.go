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
)

var brandingScalars = []string{"DefaultTheme", "EnableMeetingOptionsThemeOverride", "EnableMeetingOptionsBrandedReactionsOverride", "EnableNdiAssuranceSlate", "EnableMeetingBackgroundImages", "RequireBackgroundEffect"}

type brandingObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

func decodeBrandingOracle(raw []byte, tenant string) (brandingObservation, error) {
	var v brandingObservation
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || !strings.EqualFold(v.Tenant, tenant) || !strings.EqualFold(v.Identity, "Global") || len(v.Settings) != len(brandingScalars) {
		return brandingObservation{}, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return brandingObservation{}, errTeamOracle
	}
	for _, name := range brandingScalars {
		if _, ok := v.Settings[name]; !ok {
			return brandingObservation{}, errTeamOracle
		}
	}
	return v, nil
}

func brandingOracle(t *testing.T, ctx context.Context, tenant string) brandingObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant})
	if err != nil {
		t.Fatal("invalid branding oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-meeting-branding.ps1")
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
		t.Fatal("pinned branding read failed; no output retained")
	}
	value, err := decodeBrandingOracle(output.data, tenant)
	if err != nil {
		t.Fatal("pinned branding projection unavailable")
	}
	return value
}

// TestAccMeetingBrandingReadOnly never edits Global, creates assets or grants a
// policy. Tenant writes and effective assignments need separate acceptance.
func TestAccMeetingBrandingReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("meeting branding reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.MeetingBrandingPolicies.GetByIdentity(ctx, "Global")
	writeResponseFixture(t, "meeting_branding", "get_global", response, nil)
	if err != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("branding Global read failed (HTTP %d); no response payload retained", status)
	}
	listed, listResponse, err := sdk.TeamsAdminAPI.MeetingBrandingPolicies.ListFiltered(ctx, "Global")
	writeResponseFixture(t, "meeting_branding", "list", listResponse, nil)
	if err != nil || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || len(listed) != 1 || !reflect.DeepEqual(value.RawFields, listed[0].RawFields) {
		t.Fatal("branding Global read/list observations differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := brandingOracle(t, ctx, tenant)
		for _, key := range brandingScalars {
			want := value.RawFields[key]
			if len(want) == 0 {
				want = json.RawMessage("null") // Nullable CLR scalar presentation.
			}
			var left, right any
			if json.Unmarshal(want, &left) != nil || json.Unmarshal(actual.Settings[key], &right) != nil || !reflect.DeepEqual(left, right) {
				t.Fatal("Go and pinned branding scalar projections differ")
			}
		}
	}
	t.Log("selected branding reads verified; persistence, Premium effects and assignments remain separately qualified")
}

func TestUnitBrandingOracleRejectsIncompleteWrongTenantAndMixedOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := `{"DefaultTheme":null,"EnableMeetingOptionsThemeOverride":null,"EnableMeetingOptionsBrandedReactionsOverride":false,"EnableNdiAssuranceSlate":false,"EnableMeetingBackgroundImages":false,"RequireBackgroundEffect":null}`
	valid := `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":` + settings + `}`
	for _, raw := range []string{`{"success":false}`, strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, "Global", "Other", 1), valid + "{}", `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":{}}`} {
		if _, err := decodeBrandingOracle([]byte(raw), tenant); err == nil {
			t.Fatal("invalid branding oracle output accepted")
		}
	}
	if _, err := decodeBrandingOracle([]byte(valid), tenant); err != nil {
		t.Fatal("complete nullable branding projection rejected")
	}
}
