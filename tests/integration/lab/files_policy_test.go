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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_files_policy"
)

var filesPolicyScalars = []string{"NativeFileEntryPoints", "SPChannelFilesTab", "DefaultFileUploadAppId", "FileSharingInChatswithExternalUsers"}

type filesPolicyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

func decodeFilesPolicyOracle(raw []byte, tenant string) (filesPolicyObservation, error) {
	var v filesPolicyObservation
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || !strings.EqualFold(v.Tenant, tenant) || !strings.EqualFold(v.Identity, "Global") || len(v.Settings) != len(filesPolicyScalars) {
		return filesPolicyObservation{}, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return filesPolicyObservation{}, errTeamOracle
	}
	for _, name := range filesPolicyScalars {
		if _, ok := v.Settings[name]; !ok {
			return filesPolicyObservation{}, errTeamOracle
		}
	}
	settings, err := json.Marshal(v.Settings)
	var model resource.Policy
	if err != nil || json.Unmarshal(settings, &model) != nil {
		return filesPolicyObservation{}, errTeamOracle
	}
	return v, nil
}

func filesPolicyOracle(t *testing.T, ctx context.Context, tenant string) filesPolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant})
	if err != nil {
		t.Fatal("invalid files policy oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-files-policy.ps1")
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
		t.Fatal("pinned files policy read failed; no output retained")
	}
	value, err := decodeFilesPolicyOracle(output.data, tenant)
	if err != nil {
		t.Fatal("pinned files policy projection unavailable")
	}
	return value
}

// TestAccFilesPolicyReadOnly performs configuration reads only.
// Creates, updates, deletes, grants and effective file behavior need separate acceptance.
func TestAccFilesPolicyReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("files policy reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.TeamsFilesPolicies.GetByIdentity(ctx, "Global")
	writeResponseFixture(t, "teams_files_policy", "get_global", response, nil)
	if err != nil {
		status := 0
		phase := "preparation"
		if response != nil {
			status = response.StatusCode()
			if response.Request != nil && strings.Contains(response.Request.URL, "/Skype.Policy/configurations/TeamsFilesPolicy/configuration/Global") {
				phase = "resource"
			}
		}
		t.Fatalf("files policy %s Global read failed (HTTP %d); no response payload retained", phase, status)
	}
	listed, listResponse, err := sdk.TeamsAdminAPI.TeamsFilesPolicies.ListFiltered(ctx, "Global")
	writeResponseFixture(t, "teams_files_policy", "list", listResponse, nil)
	selected, selectionErr := globalFilesPolicy(listed)
	if err != nil || selectionErr != nil || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || !bytes.Equal(filesPolicyProjection(value), filesPolicyProjection(selected)) {
		t.Fatal("files policy Global read/list projections differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := filesPolicyOracle(t, ctx, tenant)
		if !matchesFilesPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("Go and pinned files policy getter projections differ")
		}

	}
	t.Log("selected files policy reads verified; persistence and effective file behavior remain separately qualified")
}

func TestUnitFilesPolicyOracleRejectsIncompleteWrongTenantAndMixedOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := `{"NativeFileEntryPoints":"","SPChannelFilesTab":null,"DefaultFileUploadAppId":"not-a-guid","FileSharingInChatswithExternalUsers":"Future"}`
	valid := `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":` + settings + `}`
	for _, raw := range []string{`{"success":false}`, strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, "Global", "Other", 1), valid + "{}", strings.Replace(valid, `"NativeFileEntryPoints":""`, `"NativeFileEntryPoints":false`, 1), strings.Replace(valid, `"SPChannelFilesTab":null`, `"SPChannelFilesTab":[]`, 1), strings.Replace(valid, `,"FileSharingInChatswithExternalUsers":"Future"`, "", 1)} {
		if _, e := decodeFilesPolicyOracle([]byte(raw), tenant); e == nil {
			t.Fatal("invalid files policy oracle accepted")
		}
	}
	if _, e := decodeFilesPolicyOracle([]byte(valid), tenant); e != nil {
		t.Fatal("complete nullable files policy projection rejected")
	}
}

func filesPolicyProjection(v *resource.Policy) []byte {
	raw, _ := json.Marshal(map[string]*string{"NativeFileEntryPoints": v.NativeFileEntryPoints, "SPChannelFilesTab": v.SPChannelFilesTab, "DefaultFileUploadAppId": v.DefaultFileUploadAppID, "FileSharingInChatswithExternalUsers": v.FileSharingInChatsWithExternalUsers})
	return raw
}
func globalFilesPolicy(values []*resource.Policy) (*resource.Policy, error) {
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
func TestUnitFilesPolicyGlobalSelection(t *testing.T) {
	t.Parallel()
	global := &resource.Policy{Identity: new("Global")}
	tagged := &resource.Policy{Identity: new("Tag:Global")}
	if v, e := globalFilesPolicy([]*resource.Policy{tagged, global}); e != nil || v != global {
		t.Fatal("tagged match hid Global")
	}
	if _, e := globalFilesPolicy([]*resource.Policy{global, {Identity: new("global")}}); e == nil {
		t.Fatal("ambiguous Global accepted")
	}
}

func matchesFilesPolicyGetterProjection(value *resource.Policy, settings map[string]json.RawMessage) bool {
	if value == nil {
		return false
	}
	raw, err := json.Marshal(settings)
	if err != nil {
		return false
	}
	var left, right any
	return json.Unmarshal(filesPolicyProjection(value), &left) == nil && json.Unmarshal(raw, &right) == nil && reflect.DeepEqual(left, right)
}

func TestUnitFilesPolicyOracleUsesLastCaseInsensitiveGetterValue(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"nativefileentrypoints":"fixture"}`, `{"NativeFileEntryPoints":"earlier","nativefileentrypoints":"fixture"}`, `{"NativeFileEntryPoints":null,"nativefileentrypoints":"fixture"}`} {
		var value resource.Policy
		if json.Unmarshal([]byte(raw), &value) != nil {
			t.Fatal("invalid synthetic model")
		}
		settings := map[string]json.RawMessage{}
		for _, name := range filesPolicyScalars {
			settings[name] = json.RawMessage("null")
		}
		settings["NativeFileEntryPoints"] = json.RawMessage(`"fixture"`)
		if !matchesFilesPolicyGetterProjection(&value, settings) {
			t.Fatal("canonical getter comparison used raw key spelling/earlier duplicate")
		}
	}
}
