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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_client_configuration"
)

var clientConfigurationScalars = []string{"AllowEmailIntoChannel", "RestrictedSenderList", "AllowDropBox", "AllowBox", "AllowGoogleDrive", "AllowShareFile", "AllowEgnyte", "AllowOrganizationTab", "AllowSkypeBusinessInterop", "ContentPin", "AllowResourceAccountSendMessage", "ResourceAccountContentAccess", "AllowGuestUser", "AllowScopedPeopleSearchandAccess", "AllowRoleBasedChatPermissions", "ExtendedWorkInfoInPeopleSearch", "UseUnifiedDomain"}

type clientConfigurationObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

func decodeClientConfigurationOracle(raw []byte, tenant string) (clientConfigurationObservation, error) {
	var v clientConfigurationObservation
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || !strings.EqualFold(v.Tenant, tenant) || !strings.EqualFold(v.Identity, "Global") || len(v.Settings) != len(clientConfigurationScalars) {
		return clientConfigurationObservation{}, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return clientConfigurationObservation{}, errTeamOracle
	}
	for _, name := range clientConfigurationScalars {
		if _, ok := v.Settings[name]; !ok {
			return clientConfigurationObservation{}, errTeamOracle
		}
	}
	settings, err := json.Marshal(v.Settings)
	var model resource.Configuration
	if err != nil || json.Unmarshal(settings, &model) != nil {
		return clientConfigurationObservation{}, errTeamOracle
	}
	return v, nil
}

func clientConfigurationOracle(t *testing.T, ctx context.Context, tenant string) clientConfigurationObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant})
	if err != nil {
		t.Fatal("invalid client configuration oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-client-configuration.ps1")
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
		t.Fatal("pinned client configuration read failed; no output retained")
	}
	value, err := decodeClientConfigurationOracle(output.data, tenant)
	if err != nil {
		t.Fatal("pinned client configuration projection unavailable")
	}
	return value
}

// TestAccClientConfigurationReadOnly performs configuration reads only.
// Tenant updates and effective client behavior need separate acceptance.
func TestAccClientConfigurationReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("client configuration reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.TeamsClientConfigurations.GetByIdentity(ctx, "Global")
	if err != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("client configuration Global read failed (HTTP %d); no response payload retained", status)
	}
	listed, _, err := sdk.TeamsAdminAPI.TeamsClientConfigurations.ListFiltered(ctx, "Global")
	selected, selectionErr := globalClientConfiguration(listed)
	if err != nil || selectionErr != nil || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || !bytes.Equal(clientConfigurationProjection(value), clientConfigurationProjection(selected)) {
		t.Fatal("client configuration Global read/list projections differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := clientConfigurationOracle(t, ctx, tenant)
		if !matchesClientConfigurationGetterProjection(value, actual.Settings) {
			t.Fatal("Go and pinned client configuration getter projections differ")
		}

	}
	t.Log("selected client configuration reads verified; persistence and effective client behavior remain separately qualified")
}

func TestUnitClientConfigurationOracleRejectsIncompleteWrongTenantAndMixedOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := `{"AllowEmailIntoChannel":false,"RestrictedSenderList":"","AllowDropBox":null,"AllowBox":null,"AllowGoogleDrive":null,"AllowShareFile":null,"AllowEgnyte":null,"AllowOrganizationTab":null,"AllowSkypeBusinessInterop":null,"ContentPin":"0000","AllowResourceAccountSendMessage":null,"ResourceAccountContentAccess":null,"AllowGuestUser":null,"AllowScopedPeopleSearchandAccess":null,"AllowRoleBasedChatPermissions":null,"ExtendedWorkInfoInPeopleSearch":null,"UseUnifiedDomain":null}`
	valid := `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":` + settings + `}`
	for _, raw := range []string{`{"success":false}`, strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, "Global", "Other", 1), valid + "{}", strings.Replace(valid, `"AllowEmailIntoChannel":false`, `"AllowEmailIntoChannel":"false"`, 1), strings.Replace(valid, `"RestrictedSenderList":""`, `"RestrictedSenderList":false`, 1), strings.Replace(valid, `"ContentPin":"0000"`, `"ContentPin":false`, 1), `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":{}}`} {
		if _, err := decodeClientConfigurationOracle([]byte(raw), tenant); err == nil {
			t.Fatal("invalid client configuration oracle output accepted")
		}
	}
	if _, err := decodeClientConfigurationOracle([]byte(valid), tenant); err != nil {
		t.Fatal("complete nullable client configuration projection rejected")
	}
}

func clientConfigurationProjection(v *resource.Configuration) []byte {
	raw, _ := json.Marshal(map[string]any{"AllowEmailIntoChannel": v.AllowEmailIntoChannel, "RestrictedSenderList": v.RestrictedSenderList, "AllowDropBox": v.AllowDropBox, "AllowBox": v.AllowBox, "AllowGoogleDrive": v.AllowGoogleDrive, "AllowShareFile": v.AllowShareFile, "AllowEgnyte": v.AllowEgnyte, "AllowOrganizationTab": v.AllowOrganizationTab, "AllowSkypeBusinessInterop": v.AllowSkypeBusinessInterop, "ContentPin": v.ContentPin, "AllowResourceAccountSendMessage": v.AllowResourceAccountSendMessage, "ResourceAccountContentAccess": v.ResourceAccountContentAccess, "AllowGuestUser": v.AllowGuestUser, "AllowScopedPeopleSearchandAccess": v.AllowScopedPeopleSearchAndAccess, "AllowRoleBasedChatPermissions": v.AllowRoleBasedChatPermissions, "ExtendedWorkInfoInPeopleSearch": v.ExtendedWorkInfoInPeopleSearch, "UseUnifiedDomain": v.UseUnifiedDomain})
	return raw
}
func globalClientConfiguration(values []*resource.Configuration) (*resource.Configuration, error) {
	var found *resource.Configuration
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
func TestUnitClientConfigurationGlobalSelection(t *testing.T) {
	t.Parallel()
	global := &resource.Configuration{Identity: new("Global")}
	tagged := &resource.Configuration{Identity: new("Tag:Global")}
	if v, e := globalClientConfiguration([]*resource.Configuration{tagged, global}); e != nil || v != global {
		t.Fatal("tagged match hid Global")
	}
	if _, e := globalClientConfiguration([]*resource.Configuration{global, {Identity: new("global")}}); e == nil {
		t.Fatal("ambiguous Global accepted")
	}
}

func matchesClientConfigurationGetterProjection(value *resource.Configuration, settings map[string]json.RawMessage) bool {
	if value == nil {
		return false
	}
	raw, err := json.Marshal(settings)
	if err != nil {
		return false
	}
	var left, right any
	return json.Unmarshal(clientConfigurationProjection(value), &left) == nil && json.Unmarshal(raw, &right) == nil && reflect.DeepEqual(left, right)
}

func TestUnitClientConfigurationOracleUsesLastCaseInsensitiveGetterValue(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"restrictedsenderlist":"fixture"}`, `{"RestrictedSenderList":"earlier","restrictedsenderlist":"fixture"}`, `{"RestrictedSenderList":null,"restrictedsenderlist":"fixture"}`} {
		var value resource.Configuration
		if json.Unmarshal([]byte(raw), &value) != nil {
			t.Fatal("invalid synthetic model")
		}
		settings := map[string]json.RawMessage{}
		for _, name := range clientConfigurationScalars {
			settings[name] = json.RawMessage("null")
		}
		settings["RestrictedSenderList"] = json.RawMessage(`"fixture"`)
		if !matchesClientConfigurationGetterProjection(&value, settings) {
			t.Fatal("canonical getter comparison used raw key spelling/earlier duplicate")
		}
	}
}
