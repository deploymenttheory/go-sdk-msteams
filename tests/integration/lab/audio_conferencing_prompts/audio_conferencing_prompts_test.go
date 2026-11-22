package audioconferencingpromptsacc

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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/audio_conferencing_prompts"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

type conferencingPromptsObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

var conferencingPackageArrays = []string{"InboundStartRecordingPrompt", "InboundEndRecordingPrompt", "OutboundStartRecordingPrompt", "OutboundEndRecordingPrompt", "MeetingStartRecordingPrompt", "MeetingEndRecordingPrompt"}

func nullableStrings(item map[string]json.RawMessage, names []string) bool {
	for _, name := range names {
		var text *string
		raw, ok := item[name]
		if !ok || json.Unmarshal(raw, &text) != nil {
			return false
		}
	}
	return true
}

func decodeConferencingPromptsOracle(raw []byte, tenant string) (conferencingPromptsObservation, error) {
	var v conferencingPromptsObservation
	if len(raw) > labkit.MaxOracleBytes || !labkit.DirectoryID.MatchString(tenant) {
		return v, labkit.ErrOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || !strings.EqualFold(v.Tenant, tenant) || !strings.EqualFold(v.Identity, "Global") || len(v.Settings) != 2 {
		return conferencingPromptsObservation{}, labkit.ErrOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return conferencingPromptsObservation{}, labkit.ErrOracle
	}
	for _, key := range []string{"Prompts", "Packages"} {
		body, ok := v.Settings[key]
		var items []map[string]json.RawMessage
		if !ok || json.Unmarshal(body, &items) != nil || items == nil {
			return conferencingPromptsObservation{}, labkit.ErrOracle
		}
		for _, item := range items {
			if item == nil {
				continue
			}
			if key == "Prompts" {
				if len(item) != 7 || !nullableStrings(item, []string{"Identity", "Id", "Name", "AudioPrompt", "TextPrompt", "Type", "Locale"}) {
					return conferencingPromptsObservation{}, labkit.ErrOracle
				}
			} else {
				if len(item) != 10 || !nullableStrings(item, []string{"Identity", "Id", "Name", "DefaultLocale"}) {
					return conferencingPromptsObservation{}, labkit.ErrOracle
				}
				for _, name := range conferencingPackageArrays {
					var values []*string
					body, ok := item[name]
					if !ok || json.Unmarshal(body, &values) != nil || values == nil {
						return conferencingPromptsObservation{}, labkit.ErrOracle
					}
				}
			}
		}
	}
	return v, nil
}

func conferencingPromptsProjection(v *resource.Configuration) []byte {
	prompts := make([]map[string]any, len(v.Prompts))
	for i, x := range v.Prompts {
		if x != nil {
			prompts[i] = map[string]any{"Identity": x.Identity, "Id": x.ID, "Name": x.Name, "AudioPrompt": x.AudioPrompt, "TextPrompt": x.TextPrompt, "Type": x.Type, "Locale": x.Locale}
		}
	}
	array := func(v []*string) []*string {
		if v == nil {
			return []*string{}
		}
		return v
	}
	packages := make([]map[string]any, len(v.Packages))
	for i, x := range v.Packages {
		if x != nil {
			packages[i] = map[string]any{"Identity": x.Identity, "Id": x.ID, "Name": x.Name, "DefaultLocale": x.DefaultLocale,
				"InboundStartRecordingPrompt":  array(x.InboundStartRecordingPrompt),
				"InboundEndRecordingPrompt":    array(x.InboundEndRecordingPrompt),
				"OutboundStartRecordingPrompt": array(x.OutboundStartRecordingPrompt),
				"OutboundEndRecordingPrompt":   array(x.OutboundEndRecordingPrompt),
				"MeetingStartRecordingPrompt":  array(x.MeetingStartRecordingPrompt),
				"MeetingEndRecordingPrompt":    array(x.MeetingEndRecordingPrompt),
			}
		}
	}
	raw, _ := json.Marshal(map[string]any{"Prompts": prompts, "Packages": packages})
	return raw
}

func conferencingPromptsOracle(t *testing.T, ctx context.Context, tenant string) conferencingPromptsObservation {
	t.Helper()
	input, e := json.Marshal(map[string]string{"tenant_id": tenant})
	if e != nil {
		t.Fatal("invalid prompts selector")
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-audio-conferencing-prompts.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if v, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+v)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"), "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned custom prompts read failed; no output retained")
	}
	v, e := decodeConferencingPromptsOracle(output.Data, tenant)
	if e != nil {
		t.Fatal("pinned prompts projection unavailable")
	}
	return v
}

func globalConferencingPrompts(values []*resource.Configuration) (*resource.Configuration, error) {
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

// TestAccAudioConferencingPromptsReadOnly performs only Global and filtered reads.
// It never edits prompts, packages, roles, consent or effective conferencing state.
func TestAccAudioConferencingPromptsReadOnly(t *testing.T) {
	cfg, e := lab.Load()
	if e != nil {
		t.Fatal("invalid lab enablement")
	}
	if e := cfg.Require(lab.ReadOnly); errors.Is(e, lab.ErrDisabled) {
		t.Skip("custom prompt reads require explicit read-only opt-in")
	} else if e != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, e := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "audio_conferencing_prompts", "get_global", response, nil)
	if e != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("custom prompts Global read failed (HTTP %d); no response payload retained", status)
	}
	listed, listResponse, e := sdk.TeamsAdminAPI.AudioConferencingCustomPrompts.ListFiltered(ctx, "Global")
	labkit.WriteResponseFixture(t, "audio_conferencing_prompts", "list", listResponse, nil)
	selected, selectionErr := globalConferencingPrompts(listed)
	if e != nil || selectionErr != nil || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || !bytes.Equal(conferencingPromptsProjection(value), conferencingPromptsProjection(selected)) {
		t.Fatal("custom prompts Global/list projections differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := conferencingPromptsOracle(t, ctx, tenant)
		raw, _ := json.Marshal(actual.Settings)
		var a, b any
		if json.Unmarshal(conferencingPromptsProjection(value), &a) != nil || json.Unmarshal(raw, &b) != nil || !reflect.DeepEqual(a, b) {
			t.Fatal("Go and pinned custom prompts projections differ")
		}
	}
	t.Log("selected custom prompt reads verified; writes, playback and cleanup require separate qualification")
}

func TestUnitConferencingPromptsOracleRejectsIncompleteOrMixedProjections(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := `{"Prompts":[{"Identity":null,"Id":null,"Name":"","AudioPrompt":null,"TextPrompt":null,"Type":"","Locale":""}],"Packages":[{"Identity":null,"Id":null,"Name":null,"DefaultLocale":"","InboundStartRecordingPrompt":["future",null],"InboundEndRecordingPrompt":[],"OutboundStartRecordingPrompt":[],"OutboundEndRecordingPrompt":[],"MeetingStartRecordingPrompt":[],"MeetingEndRecordingPrompt":[]}]}`
	valid := `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":` + settings + `}`
	for _, raw := range []string{valid + "{}", strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, `"identity":"Global"`, `"identity":"Other"`, 1), strings.Replace(valid, `"Type":""`, `"Type":false`, 1), strings.Replace(valid, `,"Locale":""`, "", 1), strings.Replace(valid, `"MeetingEndRecordingPrompt":[]`, `"MeetingEndRecordingPrompt":null`, 1), strings.Replace(valid, `["future",null]`, `[false]`, 1)} {
		if _, e := decodeConferencingPromptsOracle([]byte(raw), tenant); e == nil {
			t.Fatal("invalid prompts oracle accepted")
		}
	}
	if _, e := decodeConferencingPromptsOracle([]byte(valid), tenant); e != nil {
		t.Fatal("complete nullable projection rejected")
	}
	global := &resource.Configuration{Identity: new("Global")}
	tagged := &resource.Configuration{Identity: new("Tag:Global")}
	if got, e := globalConferencingPrompts([]*resource.Configuration{tagged, global}); e != nil || got != global {
		t.Fatal("tagged match hid Global")
	}
	if _, e := globalConferencingPrompts([]*resource.Configuration{global, {Identity: new("global")}}); e == nil {
		t.Fatal("ambiguous Global accepted")
	}
}
