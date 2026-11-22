package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
)

var voicemailValidationScalars = []string{"AudioFileValidationEnabled", "AudioFileValidationUri"}

type voicemailValidationObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

func decodeVoicemailValidationOracle(raw []byte, tenant string) (voicemailValidationObservation, error) {
	var v voicemailValidationObservation
	if len(raw) > labkit.MaxOracleBytes || !labkit.DirectoryID.MatchString(tenant) {
		return v, labkit.ErrOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || !strings.EqualFold(v.Tenant, tenant) || !strings.EqualFold(v.Identity, "Global") || len(v.Settings) != len(voicemailValidationScalars) {
		return voicemailValidationObservation{}, labkit.ErrOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return voicemailValidationObservation{}, labkit.ErrOracle
	}
	for _, name := range voicemailValidationScalars {
		if _, ok := v.Settings[name]; !ok {
			return voicemailValidationObservation{}, labkit.ErrOracle
		}
	}
	return v, nil
}

func voicemailValidationOracle(t *testing.T, ctx context.Context, tenant string) voicemailValidationObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant})
	if err != nil {
		t.Fatal("invalid voicemail oracle selector")
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-voicemail-validation.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"), "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned voicemail read failed; no output retained")
	}
	value, err := decodeVoicemailValidationOracle(output.Data, tenant)
	if err != nil {
		t.Fatal("pinned voicemail projection unavailable")
	}
	return value
}

// TestAccVoicemailValidationReadOnly never edits Global, creates assets or grants a
// policy. Tenant writes and effective assignments need separate acceptance.
func TestAccVoicemailValidationReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("meeting voicemail reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "online_voicemail_validation_configuration", "get_global", response, nil)
	if err != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("voicemail Global read failed (HTTP %d); no response payload retained", status)
	}
	listed, listResponse, err := sdk.TeamsAdminAPI.OnlineVoicemailValidationConfigurations.ListFiltered(ctx, "Global")
	labkit.WriteResponseFixture(t, "online_voicemail_validation_configuration", "list", listResponse, nil)
	if err != nil || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || len(listed) != 1 || !labkit.SameConfigurationFields(value.RawFields, listed[0].RawFields) {
		t.Fatal("voicemail Global read/list observations differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := voicemailValidationOracle(t, ctx, tenant)
		for _, key := range voicemailValidationScalars {
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

func TestUnitVoicemailValidationOracleRejectsIncompleteWrongTenantAndMixedOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := `{"AudioFileValidationEnabled":false,"AudioFileValidationUri":null}`
	valid := `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":` + settings + `}`
	for _, raw := range []string{`{"success":false}`, strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, "Global", "Other", 1), valid + "{}", `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":{}}`} {
		if _, err := decodeVoicemailValidationOracle([]byte(raw), tenant); err == nil {
			t.Fatal("invalid voicemail oracle output accepted")
		}
	}
	if _, err := decodeVoicemailValidationOracle([]byte(valid), tenant); err != nil {
		t.Fatal("complete nullable voicemail projection rejected")
	}
}
