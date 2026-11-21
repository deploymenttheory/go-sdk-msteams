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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_education_assignments_app_policy"
)

var educationAssignmentsScalars = []string{"ParentDigestEnabledType", "MakeCodeEnabledType", "TurnItInEnabledType", "TurnItInApiUrl", "TurnItInApiKey"}

type educationAssignmentsObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

func decodeEducationAssignmentsOracle(raw []byte, tenant string) (educationAssignmentsObservation, error) {
	var v educationAssignmentsObservation
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || !strings.EqualFold(v.Tenant, tenant) || !strings.EqualFold(v.Identity, "Global") || len(v.Settings) != len(educationAssignmentsScalars) {
		return educationAssignmentsObservation{}, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return educationAssignmentsObservation{}, errTeamOracle
	}
	for _, name := range educationAssignmentsScalars {
		if _, ok := v.Settings[name]; !ok {
			return educationAssignmentsObservation{}, errTeamOracle
		}
	}
	settings, err := json.Marshal(v.Settings)
	var model resource.Policy
	if err != nil || json.Unmarshal(settings, &model) != nil {
		return educationAssignmentsObservation{}, errTeamOracle
	}
	return v, nil
}

func educationAssignmentsOracle(t *testing.T, ctx context.Context, tenant string) educationAssignmentsObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant})
	if err != nil {
		t.Fatal("invalid education assignments policy oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-education-assignments.ps1")
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
		t.Fatal("pinned education assignments policy read failed; no output retained")
	}
	value, err := decodeEducationAssignmentsOracle(output.data, tenant)
	if err != nil {
		t.Fatal("pinned education assignments policy projection unavailable")
	}
	return value
}

// TestLabEducationAssignmentsReadOnly performs configuration reads only.
// Tenant updates and effective assignment app behavior need separate acceptance.
func TestLabEducationAssignmentsReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("education assignments policy reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.TeamsEducationAssignmentsAppPolicies.GetByIdentity(ctx, "Global")
	if err != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("education assignments policy Global read failed (HTTP %d); no response payload retained", status)
	}
	listed, _, err := sdk.TeamsAdminAPI.TeamsEducationAssignmentsAppPolicies.ListFiltered(ctx, "Global")
	selected, selectionErr := globalEducationAssignments(listed)
	if err != nil || selectionErr != nil || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || !bytes.Equal(educationAssignmentsProjection(value), educationAssignmentsProjection(selected)) {
		t.Fatal("education assignments policy Global read/list projections differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := educationAssignmentsOracle(t, ctx, tenant)
		if !matchesEducationAssignmentsGetterProjection(value, actual.Settings) {
			t.Fatal("Go and pinned education assignments policy getter projections differ")
		}

	}
	t.Log("selected education assignments policy reads verified; persistence and effective assignment app behavior remain separately qualified")
}

func TestEducationAssignmentsOracleRejectsIncompleteWrongTenantAndMixedOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := `{"ParentDigestEnabledType":null,"MakeCodeEnabledType":"","TurnItInEnabledType":null,"TurnItInApiUrl":"https://fixture.invalid","TurnItInApiKey":"fixture"}`
	valid := `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":` + settings + `}`
	for _, raw := range []string{`{"success":false}`, strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, "Global", "Other", 1), valid + "{}", strings.Replace(valid, `"MakeCodeEnabledType":""`, `"MakeCodeEnabledType":false`, 1), strings.Replace(valid, `https://fixture.invalid`, `http://fixture.invalid`, 1), strings.Replace(valid, `,"TurnItInApiKey":"fixture"`, "", 1), strings.Replace(valid, `"TurnItInApiKey":"fixture"`, `"TurnItInApiKey":false`, 1)} {
		if _, e := decodeEducationAssignmentsOracle([]byte(raw), tenant); e == nil {
			t.Fatal("invalid education assignments oracle accepted")
		}
	}
	if _, e := decodeEducationAssignmentsOracle([]byte(valid), tenant); e != nil {
		t.Fatal("complete nullable education assignments projection rejected")
	}
}

func educationAssignmentsProjection(v *resource.Policy) []byte {
	raw, _ := json.Marshal(map[string]*string{"ParentDigestEnabledType": v.ParentDigestEnabledType, "MakeCodeEnabledType": v.MakeCodeEnabledType, "TurnItInEnabledType": v.TurnItInEnabledType, "TurnItInApiUrl": v.TurnItInAPIURL, "TurnItInApiKey": v.TurnItInAPIKey})
	return raw
}
func globalEducationAssignments(values []*resource.Policy) (*resource.Policy, error) {
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
func TestEducationAssignmentsGlobalSelection(t *testing.T) {
	t.Parallel()
	global := &resource.Policy{Identity: new("Global")}
	tagged := &resource.Policy{Identity: new("Tag:Global")}
	if v, e := globalEducationAssignments([]*resource.Policy{tagged, global}); e != nil || v != global {
		t.Fatal("tagged match hid Global")
	}
	if _, e := globalEducationAssignments([]*resource.Policy{global, {Identity: new("global")}}); e == nil {
		t.Fatal("ambiguous Global accepted")
	}
}

func matchesEducationAssignmentsGetterProjection(value *resource.Policy, settings map[string]json.RawMessage) bool {
	if value == nil {
		return false
	}
	raw, err := json.Marshal(settings)
	if err != nil {
		return false
	}
	var left, right any
	return json.Unmarshal(educationAssignmentsProjection(value), &left) == nil && json.Unmarshal(raw, &right) == nil && reflect.DeepEqual(left, right)
}

func TestEducationAssignmentsOracleUsesLastCaseInsensitiveGetterValue(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"makecodeenabledtype":"fixture"}`, `{"MakeCodeEnabledType":"earlier","makecodeenabledtype":"fixture"}`, `{"MakeCodeEnabledType":null,"makecodeenabledtype":"fixture"}`} {
		var value resource.Policy
		if json.Unmarshal([]byte(raw), &value) != nil {
			t.Fatal("invalid synthetic model")
		}
		settings := map[string]json.RawMessage{}
		for _, name := range educationAssignmentsScalars {
			settings[name] = json.RawMessage("null")
		}
		settings["MakeCodeEnabledType"] = json.RawMessage(`"fixture"`)
		if !matchesEducationAssignmentsGetterProjection(&value, settings) {
			t.Fatal("canonical getter comparison used raw key spelling/earlier duplicate")
		}
	}
}
