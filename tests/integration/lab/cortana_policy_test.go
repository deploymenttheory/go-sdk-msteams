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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_cortana_policy"
)

var cortanaPolicyFields = []string{"Description", "CortanaVoiceInvocationMode", "AllowCortanaVoiceInvocation", "AllowCortanaAmbientListening", "AllowCortanaInContextSuggestions"}

type cortanaPolicyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

var cortanaOwnedIdentity = regexp.MustCompile(`(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-cortana-policy$`)

func decodeCortanaPolicyOracle(raw []byte, tenant string) (cortanaPolicyObservation, error) {
	return decodeCortanaPolicyOracleForIdentity(raw, tenant, "Global")
}
func decodeCortanaPolicyOracleForIdentity(raw []byte, tenant, identity string) (cortanaPolicyObservation, error) {
	var v cortanaPolicyObservation
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) || !strings.EqualFold(identity, "Global") && !cortanaOwnedIdentity.MatchString(identity) {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil {
		return cortanaPolicyObservation{}, errTeamOracle
	}
	identityMatches := strings.EqualFold(v.Identity, "Global")
	if !strings.EqualFold(identity, "Global") {
		identityMatches = samePolicyName(v.Identity, identity)
	}
	if !v.Success || !strings.EqualFold(v.Tenant, tenant) || !identityMatches || len(v.Settings) != len(cortanaPolicyFields) {
		return cortanaPolicyObservation{}, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return cortanaPolicyObservation{}, errTeamOracle
	}
	for _, key := range []string{"Description", "CortanaVoiceInvocationMode"} {
		var text *string
		if raw, ok := v.Settings[key]; !ok || json.Unmarshal(raw, &text) != nil {
			return cortanaPolicyObservation{}, errTeamOracle
		}
	}
	for _, key := range []string{"AllowCortanaVoiceInvocation", "AllowCortanaAmbientListening", "AllowCortanaInContextSuggestions"} {
		var boolean *bool
		if raw, ok := v.Settings[key]; !ok || json.Unmarshal(raw, &boolean) != nil {
			return cortanaPolicyObservation{}, errTeamOracle
		}
	}
	return v, nil
}

func cortanaPolicyOracle(t *testing.T, ctx context.Context, tenant string) cortanaPolicyObservation {
	t.Helper()
	return cortanaPolicyOracleForIdentity(t, ctx, tenant, "Global")
}
func cortanaPolicyOracleForIdentity(t *testing.T, ctx context.Context, tenant, identity string) cortanaPolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": identity})
	if err != nil {
		t.Fatal("invalid Cortana policy oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-cortana-policy.ps1")
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
		t.Fatal("pinned Cortana policy read failed; no output retained")
	}
	value, err := decodeCortanaPolicyOracleForIdentity(output.data, tenant, identity)
	if err != nil {
		t.Fatal("pinned Cortana policy projection unavailable")
	}
	return value
}

// TestAccCortanaPolicyReadOnly never edits Global, creates resources or grants a
// policy. Tenant writes and effective assignments need separate acceptance.
func TestAccCortanaPolicyReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("Cortana policy reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.TeamsCortanaPolicies.GetByIdentity(ctx, "Global")
	writeResponseFixture(t, "teams_cortana_policy", "get_global", response, nil)
	if err != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("Cortana policy Global read failed (HTTP %d); no response payload retained", status)
	}
	listed, listResponse, err := sdk.TeamsAdminAPI.TeamsCortanaPolicies.ListFiltered(ctx, "Global")
	writeResponseFixture(t, "teams_cortana_policy", "list", listResponse, nil)
	selected, selectionErr := globalCortanaPolicy(listed)
	if err != nil || selectionErr != nil || value == nil || !strings.EqualFold(value.Identity, "Global") {
		t.Fatal("Cortana policy Global selection differs")
	}
	if !bytes.Equal(cortanaProjection(value), cortanaProjection(selected)) {
		t.Fatal("Cortana Global read/list projections differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := cortanaPolicyOracle(t, ctx, tenant)
		if !matchesCortanaProjection(value, actual.Settings) {
			t.Fatal("Go and pinned Cortana getter projections differ")
		}
	}
	t.Log("selected Cortana policy reads verified; persistence, feature effects and assignments remain separately qualified")
}

func cortanaProjection(value *resource.Policy) []byte {
	raw, _ := json.Marshal(map[string]any{"Description": value.Description, "CortanaVoiceInvocationMode": value.CortanaVoiceInvocationMode, "AllowCortanaVoiceInvocation": value.AllowCortanaVoiceInvocation, "AllowCortanaAmbientListening": value.AllowCortanaAmbientListening, "AllowCortanaInContextSuggestions": value.AllowCortanaInContextSuggestions})
	return raw
}
func matchesCortanaProjection(value *resource.Policy, settings map[string]json.RawMessage) bool {
	if value == nil {
		return false
	}
	raw, err := json.Marshal(settings)
	if err != nil {
		return false
	}
	var left, right any
	return json.Unmarshal(cortanaProjection(value), &left) == nil && json.Unmarshal(raw, &right) == nil && reflect.DeepEqual(left, right)
}
func globalCortanaPolicy(values []resource.Policy) (*resource.Policy, error) {
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
func TestUnitCortanaPolicyOracleRejectsIncompleteWrongTenantAndMixedOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := `{"Description":"","CortanaVoiceInvocationMode":"FutureMode","AllowCortanaVoiceInvocation":false,"AllowCortanaAmbientListening":true,"AllowCortanaInContextSuggestions":null}`
	valid := `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":` + settings + `}`
	for _, raw := range []string{`{"success":false}`, strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, `"identity":"Global"`, `"identity":"Other"`, 1), valid + "{}", strings.Replace(valid, `"AllowCortanaVoiceInvocation":false`, `"AllowCortanaVoiceInvocation":0`, 1), strings.Replace(valid, `"CortanaVoiceInvocationMode":"FutureMode"`, `"CortanaVoiceInvocationMode":false`, 1), strings.Replace(valid, `"AllowCortanaAmbientListening":true`, `"AllowCortanaAmbientListening":"true"`, 1), `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":{}}`} {
		if _, e := decodeCortanaPolicyOracle([]byte(raw), tenant); e == nil {
			t.Fatal("invalid Cortana oracle accepted")
		}
	}
	if _, e := decodeCortanaPolicyOracle([]byte(valid), tenant); e != nil {
		t.Fatal("complete nullable Cortana projection rejected")
	}
	if _, e := globalCortanaPolicy([]resource.Policy{{Identity: "Tag:Global"}, {Identity: "Global"}}); e != nil {
		t.Fatal("tagged match hid Global")
	}
	if _, e := globalCortanaPolicy([]resource.Policy{{Identity: "Global"}, {Identity: "global"}}); e == nil {
		t.Fatal("ambiguous Global accepted")
	}
}

func TestUnitCortanaOracleRestrictsOwnedIdentity(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	const name = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-cortana-policy"
	raw, err := os.ReadFile("../../../msteams/teams_admin_api/teams_cortana_policy/testdata/get_global.json")
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]json.RawMessage
	if json.Unmarshal(raw, &settings) != nil {
		t.Fatal("invalid acceptance response")
	}
	projection := map[string]json.RawMessage{}
	for _, field := range cortanaPolicyFields {
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
		{"wrong-kind", strings.Replace(name, "cortana-policy", "calling-policy", 1), name, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]any{"success": true, "tenant_id": tenant, "identity": tc.returned, "settings": projection})
			if err != nil {
				t.Fatal(err)
			}
			_, err = decodeCortanaPolicyOracleForIdentity(body, tenant, tc.selected)
			if (err == nil) != tc.valid {
				t.Fatal("oracle accepted a different identity or rejected its owned policy")
			}
		})
	}
}
