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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_channels_policy"
)

var channelsPolicyScalars = []string{"Description", "AllowOrgWideTeamCreation", "EnablePrivateTeamDiscovery", "AllowPrivateChannelCreation", "AllowSharedChannelCreation", "AllowChannelSharingToExternalUser", "AllowUserToParticipateInExternalSharedChannel", "ThreadedChannelCreation", "AllowCreateChannel", "AllowCreateClassicChannel", "AllowCreatePrivateChannel", "AllowCreateSharedChannel", "CreateSharedChannelsByDefault", "AllowUsersFromOutsideTeam", "AllowGuestsFromOutsideTeam", "AllowUsersFromOutsideTeamInPrivateChannel", "AllowGuestsFromOutsideTeamInPrivateChannel", "AllowSharingWithTeamInOrg", "AllowSharingPrivateChannelWithTeamInOrg", "DefaultChannelTypeOnCreation"}

type channelsPolicyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

func decodeChannelsPolicyOracle(raw []byte, tenant string) (channelsPolicyObservation, error) {
	var v channelsPolicyObservation
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || !strings.EqualFold(v.Tenant, tenant) || !strings.EqualFold(v.Identity, "Global") || len(v.Settings) != len(channelsPolicyScalars) {
		return channelsPolicyObservation{}, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return channelsPolicyObservation{}, errTeamOracle
	}
	for _, name := range channelsPolicyScalars {
		if _, ok := v.Settings[name]; !ok {
			return channelsPolicyObservation{}, errTeamOracle
		}
	}
	settings, err := json.Marshal(v.Settings)
	var model resource.Policy
	if err != nil || json.Unmarshal(settings, &model) != nil {
		return channelsPolicyObservation{}, errTeamOracle
	}
	return v, nil
}

func channelsPolicyOracle(t *testing.T, ctx context.Context, tenant string) channelsPolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant})
	if err != nil {
		t.Fatal("invalid channels policy oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-channels-policy.ps1")
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
		t.Fatal("pinned channels policy read failed; no output retained")
	}
	value, err := decodeChannelsPolicyOracle(output.data, tenant)
	if err != nil {
		t.Fatal("pinned channels policy projection unavailable")
	}
	return value
}

// TestAccChannelsPolicyReadOnly never edits Global, creates resources or grants a
// policy. Tenant writes and effective assignments need separate acceptance.
func TestAccChannelsPolicyReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("channels policy reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.TeamsChannelsPolicies.GetByIdentity(ctx, "Global")
	if err != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("channels policy Global read failed (HTTP %d); no response payload retained", status)
	}
	listed, _, err := sdk.TeamsAdminAPI.TeamsChannelsPolicies.ListFiltered(ctx, "Global")
	selected, selectionErr := globalChannelsPolicy(listed)
	if err != nil || selectionErr != nil || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || !bytes.Equal(channelsProjection(value), channelsProjection(selected)) {
		t.Fatal("channels Global read/list projections differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := channelsPolicyOracle(t, ctx, tenant)
		if !matchesChannelsGetterProjection(value, actual.Settings) {
			t.Fatal("Go and pinned channels getter projections differ")
		}

	}
	t.Log("selected channels policy reads verified; persistence, feature effects and assignments remain separately qualified")
}

func TestUnitChannelsPolicyOracleRejectsIncompleteWrongTenantAndMixedOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := `{"Description":"","AllowOrgWideTeamCreation":null,"EnablePrivateTeamDiscovery":null,"AllowPrivateChannelCreation":false,"AllowSharedChannelCreation":null,"AllowChannelSharingToExternalUser":null,"AllowUserToParticipateInExternalSharedChannel":null,"ThreadedChannelCreation":null,"AllowCreateChannel":null,"AllowCreateClassicChannel":null,"AllowCreatePrivateChannel":"FutureMode","AllowCreateSharedChannel":null,"CreateSharedChannelsByDefault":null,"AllowUsersFromOutsideTeam":null,"AllowGuestsFromOutsideTeam":null,"AllowUsersFromOutsideTeamInPrivateChannel":null,"AllowGuestsFromOutsideTeamInPrivateChannel":null,"AllowSharingWithTeamInOrg":null,"AllowSharingPrivateChannelWithTeamInOrg":null,"DefaultChannelTypeOnCreation":null}`
	valid := `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":` + settings + `}`
	for _, raw := range []string{`{"success":false}`, strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, "Global", "Other", 1), valid + "{}", strings.Replace(valid, `"AllowPrivateChannelCreation":false`, `"AllowPrivateChannelCreation":"false"`, 1), strings.Replace(valid, `"AllowCreatePrivateChannel":"FutureMode"`, `"AllowCreatePrivateChannel":false`, 1), strings.Replace(valid, `"Description":""`, `"Description":false`, 1), `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":{}}`} {
		if _, err := decodeChannelsPolicyOracle([]byte(raw), tenant); err == nil {
			t.Fatal("invalid channels policy oracle output accepted")
		}
	}
	if _, err := decodeChannelsPolicyOracle([]byte(valid), tenant); err != nil {
		t.Fatal("complete nullable channels policy projection rejected")
	}
}

func channelsProjection(v *resource.Policy) []byte {
	raw, _ := json.Marshal(map[string]any{"Description": v.Description, "AllowOrgWideTeamCreation": v.AllowOrgWideTeamCreation, "EnablePrivateTeamDiscovery": v.EnablePrivateTeamDiscovery, "AllowPrivateChannelCreation": v.AllowPrivateChannelCreation, "AllowSharedChannelCreation": v.AllowSharedChannelCreation, "AllowChannelSharingToExternalUser": v.AllowChannelSharingToExternalUser, "AllowUserToParticipateInExternalSharedChannel": v.AllowUserToParticipateInExternalSharedChannel, "ThreadedChannelCreation": v.ThreadedChannelCreation, "AllowCreateChannel": v.AllowCreateChannel, "AllowCreateClassicChannel": v.AllowCreateClassicChannel, "AllowCreatePrivateChannel": v.AllowCreatePrivateChannel, "AllowCreateSharedChannel": v.AllowCreateSharedChannel, "CreateSharedChannelsByDefault": v.CreateSharedChannelsByDefault, "AllowUsersFromOutsideTeam": v.AllowUsersFromOutsideTeam, "AllowGuestsFromOutsideTeam": v.AllowGuestsFromOutsideTeam, "AllowUsersFromOutsideTeamInPrivateChannel": v.AllowUsersFromOutsideTeamInPrivateChannel, "AllowGuestsFromOutsideTeamInPrivateChannel": v.AllowGuestsFromOutsideTeamInPrivateChannel, "AllowSharingWithTeamInOrg": v.AllowSharingWithTeamInOrg, "AllowSharingPrivateChannelWithTeamInOrg": v.AllowSharingPrivateChannelWithTeamInOrg, "DefaultChannelTypeOnCreation": v.DefaultChannelTypeOnCreation})
	return raw
}
func globalChannelsPolicy(values []*resource.Policy) (*resource.Policy, error) {
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
func TestUnitChannelsPolicyGlobalSelection(t *testing.T) {
	t.Parallel()
	global := &resource.Policy{Identity: new("Global")}
	tagged := &resource.Policy{Identity: new("Tag:Global")}
	if v, e := globalChannelsPolicy([]*resource.Policy{tagged, global}); e != nil || v != global {
		t.Fatal("tagged match hid Global")
	}
	if _, e := globalChannelsPolicy([]*resource.Policy{global, {Identity: new("global")}}); e == nil {
		t.Fatal("ambiguous Global accepted")
	}
}

func matchesChannelsGetterProjection(value *resource.Policy, settings map[string]json.RawMessage) bool {
	if value == nil {
		return false
	}
	raw, err := json.Marshal(settings)
	if err != nil {
		return false
	}
	var left, right any
	return json.Unmarshal(channelsProjection(value), &left) == nil && json.Unmarshal(raw, &right) == nil && reflect.DeepEqual(left, right)
}

func TestUnitChannelsOracleUsesLastCaseInsensitiveGetterValue(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"description":"fixture"}`, `{"Description":"earlier","description":"fixture"}`, `{"Description":null,"description":"fixture"}`} {
		var value resource.Policy
		if json.Unmarshal([]byte(raw), &value) != nil {
			t.Fatal("invalid synthetic model")
		}
		settings := map[string]json.RawMessage{}
		for _, name := range channelsPolicyScalars {
			settings[name] = json.RawMessage("null")
		}
		settings["Description"] = json.RawMessage(`"fixture"`)
		if !matchesChannelsGetterProjection(&value, settings) {
			t.Fatal("canonical getter comparison used raw key spelling/earlier duplicate")
		}
	}
}
