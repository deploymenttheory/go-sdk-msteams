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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_events_policy"
)

var eventsPolicyScalars = []string{"AllowWebinars", "EventAccessType", "AllowTownhalls", "TownhallEventAttendeeAccess", "ExternalPresenterJoinVerification", "AllowEmailEditing", "AllowedQuestionTypesInRegistrationForm", "AllowEventIntegrations", "AllowedWebinarTypesForRecordingPublish", "AllowedTownhallTypesForRecordingPublish", "RecordingForTownhall", "RecordingForWebinar", "TranscriptionForTownhall", "TranscriptionForWebinar", "TownhallChatExperience", "BroadcastPremiumApps", "UseMicrosoftECDN", "ImmersiveEvents", "TownhallMaxResolution", "HighBitrateForTownhall", "BackroomChat", "Registration", "AllowEngagementReport", "InfoShownInReportMode", "Description"}

type eventsPolicyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

func decodeEventsPolicyOracle(raw []byte, tenant string) (eventsPolicyObservation, error) {
	var v eventsPolicyObservation
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || !strings.EqualFold(v.Tenant, tenant) || !strings.EqualFold(v.Identity, "Global") || len(v.Settings) != len(eventsPolicyScalars) {
		return eventsPolicyObservation{}, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return eventsPolicyObservation{}, errTeamOracle
	}
	for _, name := range eventsPolicyScalars {
		if _, ok := v.Settings[name]; !ok {
			return eventsPolicyObservation{}, errTeamOracle
		}
	}
	settings, err := json.Marshal(v.Settings)
	var model resource.Policy
	if err != nil || json.Unmarshal(settings, &model) != nil {
		return eventsPolicyObservation{}, errTeamOracle
	}
	return v, nil
}

func eventsPolicyOracle(t *testing.T, ctx context.Context, tenant string) eventsPolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant})
	if err != nil {
		t.Fatal("invalid events policy oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-events-policy.ps1")
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
		t.Fatal("pinned events policy read failed; no output retained")
	}
	value, err := decodeEventsPolicyOracle(output.data, tenant)
	if err != nil {
		t.Fatal("pinned events policy projection unavailable")
	}
	return value
}

// TestLabEventsPolicyReadOnly performs configuration reads only.
// Creates, updates, deletes, grants and effective event behavior need separate acceptance.
func TestLabEventsPolicyReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("events policy reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.TeamsEventsPolicies.GetByIdentity(ctx, "Global")
	if err != nil {
		status := 0
		phase := "preparation"
		if response != nil {
			status = response.StatusCode()
			if response.Request != nil && strings.Contains(response.Request.URL, "/Skype.Policy/configurations/TeamsEventsPolicy/configuration/Global") {
				phase = "resource"
			}
		}
		t.Fatalf("events policy %s Global read failed (HTTP %d); no response payload retained", phase, status)
	}
	listed, _, err := sdk.TeamsAdminAPI.TeamsEventsPolicies.ListFiltered(ctx, "Global")
	selected, selectionErr := globalEventsPolicy(listed)
	if err != nil || selectionErr != nil || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || !bytes.Equal(eventsPolicyProjection(value), eventsPolicyProjection(selected)) {
		t.Fatal("events policy Global read/list projections differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := eventsPolicyOracle(t, ctx, tenant)
		if !matchesEventsPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("Go and pinned events policy getter projections differ")
		}

	}
	t.Log("selected events policy reads verified; persistence and effective event behavior remain separately qualified")
}

func TestEventsPolicyOracleRejectsIncompleteWrongTenantAndMixedOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := map[string]any{}
	for _, name := range eventsPolicyScalars {
		settings[name] = nil
	}
	settings["Description"] = ""
	settings["AllowEventIntegrations"] = false
	settings["UseMicrosoftECDN"] = true
	raw, err := json.Marshal(eventsPolicyObservation{Success: true, Tenant: tenant, Identity: "Global", Settings: func() map[string]json.RawMessage {
		encoded, _ := json.Marshal(settings)
		var value map[string]json.RawMessage
		_ = json.Unmarshal(encoded, &value)
		return value
	}()})
	if err != nil {
		t.Fatal(err)
	}
	valid := string(raw)
	for _, input := range []string{`{"success":false}`, strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, "Global", "Other", 1), valid + "{}", strings.Replace(valid, `"Description":""`, `"Description":false`, 1), strings.Replace(valid, `"UseMicrosoftECDN":true`, `"UseMicrosoftECDN":"true"`, 1), strings.Replace(valid, `,"UseMicrosoftECDN":true`, "", 1)} {
		if _, e := decodeEventsPolicyOracle([]byte(input), tenant); e == nil {
			t.Fatal("invalid events policy oracle accepted")
		}
	}
	if _, e := decodeEventsPolicyOracle(raw, tenant); e != nil {
		t.Fatal("complete nullable events policy projection rejected")
	}
}

func eventsPolicyProjection(v *resource.Policy) []byte {
	raw, _ := json.Marshal(map[string]any{"AllowWebinars": v.AllowWebinars, "EventAccessType": v.EventAccessType, "AllowTownhalls": v.AllowTownhalls, "TownhallEventAttendeeAccess": v.TownhallEventAttendeeAccess, "ExternalPresenterJoinVerification": v.ExternalPresenterJoinVerification, "AllowEmailEditing": v.AllowEmailEditing, "AllowedQuestionTypesInRegistrationForm": v.AllowedQuestionTypesInRegistrationForm, "AllowEventIntegrations": v.AllowEventIntegrations, "AllowedWebinarTypesForRecordingPublish": v.AllowedWebinarTypesForRecordingPublish, "AllowedTownhallTypesForRecordingPublish": v.AllowedTownhallTypesForRecordingPublish, "RecordingForTownhall": v.RecordingForTownhall, "RecordingForWebinar": v.RecordingForWebinar, "TranscriptionForTownhall": v.TranscriptionForTownhall, "TranscriptionForWebinar": v.TranscriptionForWebinar, "TownhallChatExperience": v.TownhallChatExperience, "BroadcastPremiumApps": v.BroadcastPremiumApps, "UseMicrosoftECDN": v.UseMicrosoftECDN, "ImmersiveEvents": v.ImmersiveEvents, "TownhallMaxResolution": v.TownhallMaxResolution, "HighBitrateForTownhall": v.HighBitrateForTownhall, "BackroomChat": v.BackroomChat, "Registration": v.Registration, "AllowEngagementReport": v.AllowEngagementReport, "InfoShownInReportMode": v.InfoShownInReportMode, "Description": v.Description})
	return raw
}
func globalEventsPolicy(values []*resource.Policy) (*resource.Policy, error) {
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
func TestEventsPolicyGlobalSelection(t *testing.T) {
	t.Parallel()
	global := &resource.Policy{Identity: new("Global")}
	tagged := &resource.Policy{Identity: new("Tag:Global")}
	if v, e := globalEventsPolicy([]*resource.Policy{tagged, global}); e != nil || v != global {
		t.Fatal("tagged match hid Global")
	}
	if _, e := globalEventsPolicy([]*resource.Policy{global, {Identity: new("global")}}); e == nil {
		t.Fatal("ambiguous Global accepted")
	}
}

func matchesEventsPolicyGetterProjection(value *resource.Policy, settings map[string]json.RawMessage) bool {
	if value == nil {
		return false
	}
	raw, err := json.Marshal(settings)
	if err != nil {
		return false
	}
	var left, right any
	return json.Unmarshal(eventsPolicyProjection(value), &left) == nil && json.Unmarshal(raw, &right) == nil && reflect.DeepEqual(left, right)
}

func TestEventsPolicyOracleUsesLastCaseInsensitiveGetterValue(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"description":"fixture"}`, `{"Description":"earlier","description":"fixture"}`, `{"Description":null,"description":"fixture"}`} {
		var value resource.Policy
		if json.Unmarshal([]byte(raw), &value) != nil {
			t.Fatal("invalid synthetic model")
		}
		settings := map[string]json.RawMessage{}
		for _, name := range eventsPolicyScalars {
			settings[name] = json.RawMessage("null")
		}
		settings["Description"] = json.RawMessage(`"fixture"`)
		if !matchesEventsPolicyGetterProjection(&value, settings) {
			t.Fatal("canonical getter comparison used raw key spelling/earlier duplicate")
		}
	}
}
