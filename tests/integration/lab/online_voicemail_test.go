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
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
)

var voicemailScalars = []string{"Description", "EnableTranscription", "MaximumRecordingLength"}

type voicemailObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

var voicemailOwnedIdentity = regexp.MustCompile(`(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-voicemail-policy$`)

func decodeVoicemailOracle(raw []byte, tenant string) (voicemailObservation, error) {
	return decodeVoicemailOracleForIdentity(raw, tenant, "Global")
}
func decodeVoicemailOracleForIdentity(raw []byte, tenant, identity string) (voicemailObservation, error) {
	var v voicemailObservation
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) || !strings.EqualFold(identity, "Global") && !voicemailOwnedIdentity.MatchString(identity) {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil {
		return voicemailObservation{}, errTeamOracle
	}
	identityMatches := strings.EqualFold(v.Identity, "Global")
	if !strings.EqualFold(identity, "Global") {
		identityMatches = samePolicyName(v.Identity, identity)
	}
	if !v.Success || !strings.EqualFold(v.Tenant, tenant) || !identityMatches || len(v.Settings) != len(voicemailScalars) {
		return voicemailObservation{}, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return voicemailObservation{}, errTeamOracle
	}
	for _, name := range voicemailScalars {
		if _, ok := v.Settings[name]; !ok {
			return voicemailObservation{}, errTeamOracle
		}
	}
	return v, nil
}

func voicemailOracle(t *testing.T, ctx context.Context, tenant string) voicemailObservation {
	return voicemailOracleForIdentity(t, ctx, tenant, "Global")
}
func voicemailOracleForIdentity(t *testing.T, ctx context.Context, tenant, identity string) voicemailObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": identity})
	if err != nil {
		t.Fatal("invalid voicemail oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-online-voicemail.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"), "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output teamOracleBuffer
	defer func() { clear(output.data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.exceeded {
		t.Fatal("pinned voicemail read failed; no output retained")
	}
	value, err := decodeVoicemailOracleForIdentity(output.data, tenant, identity)
	if err != nil {
		t.Fatal("pinned voicemail projection unavailable")
	}
	return value
}

// TestAccOnlineVoicemailReadOnly never edits Global, creates assets or grants a
// policy. Tenant writes and effective assignments need separate acceptance.
func TestAccOnlineVoicemailReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("meeting voicemail reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.OnlineVoicemailPolicies.GetByIdentity(ctx, "Global")
	writeResponseFixture(t, "online_voicemail_policy", "get_global", response, nil)
	if err != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("voicemail Global read failed (HTTP %d); no response payload retained", status)
	}
	listed, listResponse, err := sdk.TeamsAdminAPI.OnlineVoicemailPolicies.ListFiltered(ctx, "Global")
	writeResponseFixture(t, "online_voicemail_policy", "list", listResponse, nil)
	if err != nil {
		logAccessResponse(t, listResponse, err)
		t.Fatal("voicemail Global list request failed")
	}
	if value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") {
		t.Fatal("voicemail Global read identity differs")
	}
	if len(listed) != 1 || listed[0] == nil {
		t.Fatal("voicemail Global list selection differs")
	}
	if !sameConfigurationFields(value.RawFields, listed[0].RawFields) {
		reason := responseComparisonFailure(value.RawFields, listed[0].RawFields)
		t.Log(reason)
		t.Fatal("voicemail Global read/list observations differ")
	}

	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := voicemailOracle(t, ctx, tenant)
		for _, key := range voicemailScalars {
			want := value.RawFields[key]
			if len(want) == 0 {
				want = json.RawMessage("null") // Nullable CLR scalar presentation.
			}
			var left, right any
			if json.Unmarshal(want, &left) != nil || json.Unmarshal(actual.Settings[key], &right) != nil || !reflect.DeepEqual(left, right) {
				t.Fatal("Go and pinned voicemail scalar projections differ")
			}
		}
	}
	t.Log("selected voicemail reads verified; persistence, feature effects and assignments remain separately qualified")
}

func TestUnitVoicemailOracleRejectsIncompleteWrongTenantAndMixedOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := `{"Description":null,"EnableTranscription":false,"MaximumRecordingLength":"00:01:30"}`
	valid := `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":` + settings + `}`
	for _, raw := range []string{`{"success":false}`, strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, "Global", "Other", 1), valid + "{}", `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":{}}`} {
		if _, err := decodeVoicemailOracle([]byte(raw), tenant); err == nil {
			t.Fatal("invalid voicemail oracle output accepted")
		}
	}
	if _, err := decodeVoicemailOracle([]byte(valid), tenant); err != nil {
		t.Fatal("complete nullable voicemail projection rejected")
	}
}

func TestUnitVoicemailOracleRestrictsOwnedIdentity(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	const name = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-voicemail-policy"
	raw, err := os.ReadFile("../../../msteams/teams_admin_api/online_voicemail_policy/testdata/get_global.json")
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]json.RawMessage
	if json.Unmarshal(raw, &settings) != nil {
		t.Fatal("invalid acceptance response")
	}
	projection := map[string]json.RawMessage{}
	for _, field := range voicemailScalars {
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
		{"wrong-kind", strings.Replace(name, "voicemail-policy", "calling-policy", 1), name, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]any{"success": true, "tenant_id": tenant, "identity": tc.returned, "settings": projection})
			if err != nil {
				t.Fatal(err)
			}
			_, err = decodeVoicemailOracleForIdentity(body, tenant, tc.selected)
			if (err == nil) != tc.valid {
				t.Fatal("oracle accepted a different identity or rejected its owned policy")
			}
		})
	}
}
