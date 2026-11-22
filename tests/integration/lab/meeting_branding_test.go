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
)

var brandingScalars = []string{"DefaultTheme", "EnableMeetingOptionsThemeOverride", "EnableMeetingOptionsBrandedReactionsOverride", "EnableNdiAssuranceSlate", "EnableMeetingBackgroundImages", "RequireBackgroundEffect"}

type brandingObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

var brandingOwnedIdentity = regexp.MustCompile(`(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-branding-policy$`)

func decodeBrandingOracle(raw []byte, tenant string) (brandingObservation, error) {
	return decodeBrandingOracleForIdentity(raw, tenant, "Global")
}
func decodeBrandingOracleForIdentity(raw []byte, tenant, identity string) (brandingObservation, error) {
	var v brandingObservation
	if len(raw) > labkit.MaxOracleBytes || !labkit.DirectoryID.MatchString(tenant) || !strings.EqualFold(identity, "Global") && !brandingOwnedIdentity.MatchString(identity) {
		return v, labkit.ErrOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil {
		return brandingObservation{}, labkit.ErrOracle
	}
	identityMatches := strings.EqualFold(v.Identity, "Global")
	if !strings.EqualFold(identity, "Global") {
		identityMatches = labkit.SamePolicyName(v.Identity, identity)
	}
	if !v.Success || !strings.EqualFold(v.Tenant, tenant) || !identityMatches || len(v.Settings) != len(brandingScalars) {
		return brandingObservation{}, labkit.ErrOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return brandingObservation{}, labkit.ErrOracle
	}
	for _, name := range brandingScalars {
		if _, ok := v.Settings[name]; !ok {
			return brandingObservation{}, labkit.ErrOracle
		}
	}
	return v, nil
}

func brandingOracle(t *testing.T, ctx context.Context, tenant string) brandingObservation {
	return brandingOracleForIdentity(t, ctx, tenant, "Global")
}
func brandingOracleForIdentity(t *testing.T, ctx context.Context, tenant, identity string) brandingObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": identity})
	if err != nil {
		t.Fatal("invalid branding oracle selector")
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-meeting-branding.ps1")
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
		t.Fatal("pinned branding read failed; no output retained")
	}
	value, err := decodeBrandingOracleForIdentity(output.Data, tenant, identity)
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
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.MeetingBrandingPolicies.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "meeting_branding", "get_global", response, nil)
	if err != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("branding Global read failed (HTTP %d); no response payload retained", status)
	}
	listed, listResponse, err := sdk.TeamsAdminAPI.MeetingBrandingPolicies.ListFiltered(ctx, "Global")
	labkit.WriteResponseFixture(t, "meeting_branding", "list", listResponse, nil)
	if err != nil || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || len(listed) != 1 || !labkit.SameConfigurationFields(value.RawFields, listed[0].RawFields) {
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

func TestUnitBrandingOracleRestrictsOwnedIdentity(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	const name = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-branding-policy"
	raw, err := os.ReadFile(labkit.Root + "/msteams/teams_admin_api/meeting_branding/testdata/get_global.json")
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]json.RawMessage
	if json.Unmarshal(raw, &settings) != nil {
		t.Fatal("invalid acceptance response")
	}
	projection := map[string]json.RawMessage{}
	for _, field := range brandingScalars {
		projection[field] = settings[field]
	}
	for _, tc := range []struct {
		name, selected, returned string
		valid                    bool
	}{
		{"global", "Global", "global", true},
		{"owned-bare", name, "Tag:" + name, true},
		{"owned-tagged", "Tag:" + name, name, true},
		{"foreign-selection", "other-policy", "other-policy", false},
		{"wrong-owned-policy", name, strings.Replace(name, "aaaa", "bbbb", 1), false},
		{"tagged-global", "Global", "Tag:Global", false},
		{"wrong-kind", strings.Replace(name, "branding-policy", "calling-policy", 1), name, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]any{"success": true, "tenant_id": tenant, "identity": tc.returned, "settings": projection})
			if err != nil {
				t.Fatal(err)
			}
			_, err = decodeBrandingOracleForIdentity(body, tenant, tc.selected)
			if (err == nil) != tc.valid {
				t.Fatal("oracle accepted a different identity or rejected its owned policy")
			}
		})
	}
}
