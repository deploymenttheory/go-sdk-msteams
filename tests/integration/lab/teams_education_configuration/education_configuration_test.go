package teamseducationconfigurationacc

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

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_education_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

var educationConfigurationScalars = []string{"ParentGuardianPreferredContactMethod", "EduGenerativeAIEnhancements", "UpdateParentInformation"}

type educationConfigurationObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

func decodeEducationConfigurationOracle(raw []byte, tenant string) (educationConfigurationObservation, error) {
	var v educationConfigurationObservation
	if len(raw) > labkit.MaxOracleBytes || !labkit.DirectoryID.MatchString(tenant) {
		return v, labkit.ErrOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || !strings.EqualFold(v.Tenant, tenant) || !strings.EqualFold(v.Identity, "Global") || len(v.Settings) != len(educationConfigurationScalars) {
		return educationConfigurationObservation{}, labkit.ErrOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return educationConfigurationObservation{}, labkit.ErrOracle
	}
	for _, name := range educationConfigurationScalars {
		if _, ok := v.Settings[name]; !ok {
			return educationConfigurationObservation{}, labkit.ErrOracle
		}
	}
	settings, err := json.Marshal(v.Settings)
	var model resource.Configuration
	if err != nil || json.Unmarshal(settings, &model) != nil {
		return educationConfigurationObservation{}, labkit.ErrOracle
	}
	return v, nil
}

func educationConfigurationOracle(t *testing.T, ctx context.Context, tenant string) educationConfigurationObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant})
	if err != nil {
		t.Fatal("invalid education configuration oracle selector")
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-education-configuration.ps1")
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
		t.Fatal("pinned education configuration read failed; no output retained")
	}
	value, err := decodeEducationConfigurationOracle(output.Data, tenant)
	if err != nil {
		t.Fatal("pinned education configuration projection unavailable")
	}
	return value
}

// TestAccEducationConfigurationReadOnly performs configuration reads only.
// Tenant updates and effective education behavior need separate acceptance.
func TestAccEducationConfigurationReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("education configuration reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.TeamsEducationConfigurations.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_education_configuration", "get_global", response, nil)
	if err != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("education configuration Global read failed (HTTP %d); no response payload retained", status)
	}
	listed, listResponse, err := sdk.TeamsAdminAPI.TeamsEducationConfigurations.ListFiltered(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_education_configuration", "list", listResponse, nil)
	selected, selectionErr := globalEducationConfiguration(listed)
	if err != nil || selectionErr != nil || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || !bytes.Equal(educationConfigurationProjection(value), educationConfigurationProjection(selected)) {
		t.Fatal("education configuration Global read/list projections differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := educationConfigurationOracle(t, ctx, tenant)
		if !matchesEducationConfigurationGetterProjection(value, actual.Settings) {
			t.Fatal("Go and pinned education configuration getter projections differ")
		}

	}
	t.Log("selected education configuration reads verified; persistence and effective education behavior remain separately qualified")
}

func TestUnitEducationConfigurationOracleRejectsIncompleteWrongTenantAndMixedOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := `{"ParentGuardianPreferredContactMethod":null,"EduGenerativeAIEnhancements":"","UpdateParentInformation":"fixture"}`
	valid := `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":` + settings + `}`
	for _, raw := range []string{`{"success":false}`, strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, "Global", "Other", 1), valid + "{}", strings.Replace(valid, `"EduGenerativeAIEnhancements":""`, `"EduGenerativeAIEnhancements":false`, 1), strings.Replace(valid, `,"UpdateParentInformation":"fixture"`, "", 1), strings.Replace(valid, `"UpdateParentInformation":"fixture"`, `"UpdateParentInformation":false`, 1)} {
		if _, e := decodeEducationConfigurationOracle([]byte(raw), tenant); e == nil {
			t.Fatal("invalid education configuration oracle accepted")
		}
	}
	if _, e := decodeEducationConfigurationOracle([]byte(valid), tenant); e != nil {
		t.Fatal("complete nullable education configuration projection rejected")
	}
}

func educationConfigurationProjection(v *resource.Configuration) []byte {
	raw, _ := json.Marshal(map[string]*string{"ParentGuardianPreferredContactMethod": v.ParentGuardianPreferredContactMethod, "EduGenerativeAIEnhancements": v.EduGenerativeAIEnhancements, "UpdateParentInformation": v.UpdateParentInformation})
	return raw
}

func globalEducationConfiguration(values []*resource.Configuration) (*resource.Configuration, error) {
	var found *resource.Configuration
	for _, v := range values {
		if v != nil && v.Identity != nil && strings.EqualFold(*v.Identity, "Global") {
			if found != nil {
				return nil, labkit.ErrOracle
			}
			found = v
		}
	}
	if found == nil {
		return nil, labkit.ErrOracle
	}
	return found, nil
}

func TestUnitEducationConfigurationGlobalSelection(t *testing.T) {
	t.Parallel()
	global := &resource.Configuration{Identity: new("Global")}
	tagged := &resource.Configuration{Identity: new("Tag:Global")}
	if v, e := globalEducationConfiguration([]*resource.Configuration{tagged, global}); e != nil || v != global {
		t.Fatal("tagged match hid Global")
	}
	if _, e := globalEducationConfiguration([]*resource.Configuration{global, {Identity: new("global")}}); e == nil {
		t.Fatal("ambiguous Global accepted")
	}
}

func matchesEducationConfigurationGetterProjection(value *resource.Configuration, settings map[string]json.RawMessage) bool {
	if value == nil {
		return false
	}
	raw, err := json.Marshal(settings)
	if err != nil {
		return false
	}
	var left, right any
	return json.Unmarshal(educationConfigurationProjection(value), &left) == nil && json.Unmarshal(raw, &right) == nil && reflect.DeepEqual(left, right)
}

func TestUnitEducationConfigurationOracleUsesLastCaseInsensitiveGetterValue(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"edugenerativeaienhancements":"fixture"}`, `{"EduGenerativeAIEnhancements":"earlier","edugenerativeaienhancements":"fixture"}`, `{"EduGenerativeAIEnhancements":null,"edugenerativeaienhancements":"fixture"}`} {
		var value resource.Configuration
		if json.Unmarshal([]byte(raw), &value) != nil {
			t.Fatal("invalid synthetic model")
		}
		settings := map[string]json.RawMessage{}
		for _, name := range educationConfigurationScalars {
			settings[name] = json.RawMessage("null")
		}
		settings["EduGenerativeAIEnhancements"] = json.RawMessage(`"fixture"`)
		if !matchesEducationConfigurationGetterProjection(&value, settings) {
			t.Fatal("canonical getter comparison used raw key spelling/earlier duplicate")
		}
	}
}
