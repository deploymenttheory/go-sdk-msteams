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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_enhanced_encryption_policy"
)

var enhancedEncryptionFields = []string{"Description", "CallingEndtoEndEncryptionEnabledType", "MeetingEndToEndEncryption"}

type enhancedEncryptionObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

func decodeEnhancedEncryptionOracle(raw []byte, tenant string) (enhancedEncryptionObservation, error) {
	var v enhancedEncryptionObservation
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || !strings.EqualFold(v.Tenant, tenant) || !strings.EqualFold(v.Identity, "Global") || len(v.Settings) != len(enhancedEncryptionFields) {
		return enhancedEncryptionObservation{}, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return enhancedEncryptionObservation{}, errTeamOracle
	}
	for _, key := range enhancedEncryptionFields {
		var text *string
		if raw, ok := v.Settings[key]; !ok || json.Unmarshal(raw, &text) != nil {
			return enhancedEncryptionObservation{}, errTeamOracle
		}
	}
	return v, nil
}

func enhancedEncryptionOracle(t *testing.T, ctx context.Context, tenant string) enhancedEncryptionObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant})
	if err != nil {
		t.Fatal("invalid enhanced encryption policy oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-enhanced-encryption-policy.ps1")
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
		t.Fatal("pinned enhanced encryption policy read failed; no output retained")
	}
	value, err := decodeEnhancedEncryptionOracle(output.data, tenant)
	if err != nil {
		t.Fatal("pinned enhanced encryption policy projection unavailable")
	}
	return value
}

// TestLabEnhancedEncryptionReadOnly never edits Global, creates resources or grants a
// policy. Tenant writes and effective assignments need separate acceptance.
func TestLabEnhancedEncryptionReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("enhanced encryption policy reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.TeamsEnhancedEncryptionPolicies.GetByIdentity(ctx, "Global")
	if err != nil {
		status := 0
		phase := "preparation"
		if response != nil {
			status = response.StatusCode()
			if response.Request != nil && strings.Contains(response.Request.URL, "/Skype.Policy/configurations/TeamsEnhancedEncryptionPolicy/configuration/Global") {
				phase = "resource"
			}
		}
		t.Fatalf("enhanced encryption policy %s Global read failed (HTTP %d); no response payload retained", phase, status)
	}
	listed, _, err := sdk.TeamsAdminAPI.TeamsEnhancedEncryptionPolicies.ListFiltered(ctx, "Global")
	selected, selectionErr := globalEnhancedEncryption(listed)
	if err != nil || selectionErr != nil || value == nil || !strings.EqualFold(value.Identity, "Global") {
		t.Fatal("enhanced encryption policy Global selection differs")
	}
	if !bytes.Equal(enhancedEncryptionProjection(value), enhancedEncryptionProjection(selected)) {
		t.Fatal("enhanced encryption Global read/list projections differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := enhancedEncryptionOracle(t, ctx, tenant)
		if !matchesEnhancedEncryptionProjection(value, actual.Settings) {
			t.Fatal("Go and pinned enhanced encryption getter projections differ")
		}
	}
	t.Log("selected enhanced encryption policy reads verified; persistence, feature effects and assignments remain separately qualified")
}

func enhancedEncryptionProjection(value *resource.Policy) []byte {
	raw, _ := json.Marshal(map[string]*string{"Description": value.Description, "CallingEndtoEndEncryptionEnabledType": value.CallingEndToEndEncryptionEnabledType, "MeetingEndToEndEncryption": value.MeetingEndToEndEncryption})
	return raw
}
func matchesEnhancedEncryptionProjection(value *resource.Policy, settings map[string]json.RawMessage) bool {
	if value == nil {
		return false
	}
	raw, err := json.Marshal(settings)
	if err != nil {
		return false
	}
	var left, right any
	return json.Unmarshal(enhancedEncryptionProjection(value), &left) == nil && json.Unmarshal(raw, &right) == nil && reflect.DeepEqual(left, right)
}
func globalEnhancedEncryption(values []resource.Policy) (*resource.Policy, error) {
	var found *resource.Policy
	for i := range values {
		if strings.EqualFold(values[i].Identity, "Global") {
			if found != nil {
				return nil, errTeamOracle
			}
			found = &values[i]
		}
	}
	if found == nil {
		return nil, errTeamOracle
	}
	return found, nil
}
func TestEnhancedEncryptionOracleRejectsIncompleteWrongTenantAndMixedOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := `{"Description":"","CallingEndtoEndEncryptionEnabledType":"FutureCalling","MeetingEndToEndEncryption":null}`
	valid := `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":` + settings + `}`
	for _, raw := range []string{`{"success":false}`, strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, `"identity":"Global"`, `"identity":"Other"`, 1), valid + "{}", strings.Replace(valid, `"CallingEndtoEndEncryptionEnabledType":"FutureCalling"`, `"CallingEndtoEndEncryptionEnabledType":false`, 1), strings.Replace(valid, `"MeetingEndToEndEncryption":null`, `"MeetingEndToEndEncryption":0`, 1), strings.Replace(valid, `,"MeetingEndToEndEncryption":null`, "", 1)} {
		if _, e := decodeEnhancedEncryptionOracle([]byte(raw), tenant); e == nil {
			t.Fatal("invalid enhanced encryption oracle accepted")
		}
	}
	if _, e := decodeEnhancedEncryptionOracle([]byte(valid), tenant); e != nil {
		t.Fatal("complete nullable enhanced encryption projection rejected")
	}
	if _, e := globalEnhancedEncryption([]resource.Policy{{Identity: "Tag:Global"}, {Identity: "Global"}}); e != nil {
		t.Fatal("tagged match hid Global")
	}
	if _, e := globalEnhancedEncryption([]resource.Policy{{Identity: "Global"}, {Identity: "global"}}); e == nil {
		t.Fatal("ambiguous Global accepted")
	}
}
