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

var privacyConfigurationScalars = []string{"EnablePrivacyMode", "AutoInitiateContacts", "PublishLocationDataDefault", "DisplayPublishedPhotoDefault"}

type privacyConfigurationObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

func decodePrivacyConfigurationOracle(raw []byte, tenant string) (privacyConfigurationObservation, error) {
	var v privacyConfigurationObservation
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || !strings.EqualFold(v.Tenant, tenant) || !strings.EqualFold(v.Identity, "Global") || len(v.Settings) != len(privacyConfigurationScalars) {
		return privacyConfigurationObservation{}, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return privacyConfigurationObservation{}, errTeamOracle
	}
	for _, name := range privacyConfigurationScalars {
		var value *bool
		if raw, ok := v.Settings[name]; !ok || json.Unmarshal(raw, &value) != nil {
			return privacyConfigurationObservation{}, errTeamOracle
		}
	}
	return v, nil
}

func privacyConfigurationOracle(t *testing.T, ctx context.Context, tenant string) privacyConfigurationObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant})
	if err != nil {
		t.Fatal("invalid privacy configuration oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-privacy-configuration.ps1")
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
		t.Fatal("pinned privacy configuration read failed; no output retained")
	}
	value, err := decodePrivacyConfigurationOracle(output.data, tenant)
	if err != nil {
		t.Fatal("pinned privacy configuration projection unavailable")
	}
	return value
}

// TestAccPrivacyConfigurationReadOnly never edits Global, creates resources or grants a
// policy. Tenant writes and effective assignments need separate acceptance.
func TestAccPrivacyConfigurationReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("privacy configuration reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.PrivacyConfigurations.GetByIdentity(ctx, "Global")
	writeResponseFixture(t, "privacy_configuration", "get_global", response, nil)
	if err != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("privacy configuration Global read failed (HTTP %d); no response payload retained", status)
	}
	listed, listResponse, err := sdk.TeamsAdminAPI.PrivacyConfigurations.ListFiltered(ctx, "Global")
	writeResponseFixture(t, "privacy_configuration", "list", listResponse, nil)
	if err != nil || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || len(listed) != 1 || !reflect.DeepEqual(value.RawFields, listed[0].RawFields) {
		t.Fatal("privacy configuration Global read/list observations differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := privacyConfigurationOracle(t, ctx, tenant)
		for _, key := range privacyConfigurationScalars {
			want := value.RawFields[key]
			if len(want) == 0 {
				want = json.RawMessage("null") // Nullable CLR scalar presentation.
			}
			var left, right any
			if json.Unmarshal(want, &left) != nil || json.Unmarshal(actual.Settings[key], &right) != nil || !reflect.DeepEqual(left, right) {
				t.Fatal("Go and pinned privacy configuration scalar projections differ")
			}
		}
	}
	t.Log("selected privacy configuration reads verified; persistence, feature effects and assignments remain separately qualified")
}

func TestUnitPrivacyConfigurationOracleRejectsIncompleteWrongTenantAndMixedOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := `{"EnablePrivacyMode":false,"AutoInitiateContacts":true,"PublishLocationDataDefault":null,"DisplayPublishedPhotoDefault":false}`
	valid := `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":` + settings + `}`
	for _, raw := range []string{`{"success":false}`, strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, "Global", "Other", 1), valid + "{}", strings.Replace(valid, `"EnablePrivacyMode":false`, `"EnablePrivacyMode":0`, 1), `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":{}}`} {
		if _, err := decodePrivacyConfigurationOracle([]byte(raw), tenant); err == nil {
			t.Fatal("invalid privacy configuration oracle output accepted")
		}
	}
	if _, err := decodePrivacyConfigurationOracle([]byte(valid), tenant); err != nil {
		t.Fatal("complete nullable privacy configuration projection rejected")
	}
}
