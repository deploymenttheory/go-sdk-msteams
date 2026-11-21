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

var voicemailScalars = []string{"Description", "EnableTranscription", "MaximumRecordingLength"}

type voicemailObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

func decodeVoicemailOracle(raw []byte, tenant string) (voicemailObservation, error) {
	var v voicemailObservation
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || !strings.EqualFold(v.Tenant, tenant) || !strings.EqualFold(v.Identity, "Global") || len(v.Settings) != len(voicemailScalars) {
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
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant})
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
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output teamOracleBuffer
	defer func() { clear(output.data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.exceeded {
		t.Fatal("pinned voicemail read failed; no output retained")
	}
	value, err := decodeVoicemailOracle(output.data, tenant)
	if err != nil {
		t.Fatal("pinned voicemail projection unavailable")
	}
	return value
}

// TestLabOnlineVoicemailReadOnly never edits Global, creates assets or grants a
// policy. Tenant writes and effective assignments need separate acceptance.
func TestLabOnlineVoicemailReadOnly(t *testing.T) {
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
	if err != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("voicemail Global read failed (HTTP %d); no response payload retained", status)
	}
	listed, _, err := sdk.TeamsAdminAPI.OnlineVoicemailPolicies.ListFiltered(ctx, "Global")
	if err != nil || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || len(listed) != 1 || !reflect.DeepEqual(value.RawFields, listed[0].RawFields) {
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

func TestVoicemailOracleRejectsIncompleteWrongTenantAndMixedOutput(t *testing.T) {
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
