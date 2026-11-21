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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/recording_messages"
)

type recordingOracleObservation struct {
	Identity string          `json:"identity"`
	Settings json.RawMessage `json:"settings"`
}

func decodeRecordingOracle(raw []byte, tenant string, identities []string) (map[string]json.RawMessage, error) {
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) || len(identities) > 5 {
		return nil, errTeamOracle
	}
	var v struct {
		Success      bool                         `json:"success"`
		Tenant       string                       `json:"tenant_id"`
		Observations []recordingOracleObservation `json:"observations"`
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || !strings.EqualFold(v.Tenant, tenant) || len(v.Observations) != len(identities) {
		return nil, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return nil, errTeamOracle
	}
	allowed := map[string]bool{}
	for _, identity := range identities {
		if identity == "" || allowed[identity] {
			return nil, errTeamOracle
		}
		allowed[identity] = true
	}
	result := map[string]json.RawMessage{}
	for _, o := range v.Observations {
		if !allowed[o.Identity] || result[o.Identity] != nil {
			return nil, errTeamOracle
		}
		var fields map[string]json.RawMessage
		if json.Unmarshal(o.Settings, &fields) != nil || len(fields) != 3 {
			return nil, errTeamOracle
		}
		for _, key := range []string{"Id", "Description", "RecordingAndTranscriptionLocalizationCustomMessage"} {
			if _, ok := fields[key]; !ok {
				return nil, errTeamOracle
			}
		}
		// Validate ID/text using the SDK rule, then validate the canonical child
		// presentation separately: an absent CLR Language projects as null here.
		projection := map[string]json.RawMessage{"Id": fields["Id"], "Description": fields["Description"]}
		encoded, e := json.Marshal(projection)
		var model resource.Message
		if e != nil || json.Unmarshal(encoded, &model) != nil {
			return nil, errTeamOracle
		}
		var children []map[string]*string
		if json.Unmarshal(fields["RecordingAndTranscriptionLocalizationCustomMessage"], &children) != nil || children == nil {
			return nil, errTeamOracle
		}
		for _, child := range children {
			if child == nil {
				continue
			}
			if len(child) != 8 {
				return nil, errTeamOracle
			}
			for _, name := range []string{"Identity", "Language", "InitiatorImplicit", "ParticipantImplicit", "InitiatorExplicit", "ParticipantExplicitRequested", "ParticipantExplicitProvided", "AgreementDialogue"} {
				if _, ok := child[name]; !ok {
					return nil, errTeamOracle
				}
			}
		}
		result[o.Identity] = o.Settings
	}
	return result, nil
}

func recordingProjection(v *resource.Message) []byte {
	items := make([]map[string]*string, len(v.Localizations))
	for i, x := range v.Localizations {
		if x != nil {
			items[i] = map[string]*string{"Identity": x.Identity, "Language": x.Language, "InitiatorImplicit": x.InitiatorImplicit, "ParticipantImplicit": x.ParticipantImplicit, "InitiatorExplicit": x.InitiatorExplicit, "ParticipantExplicitRequested": x.ParticipantExplicitRequested, "ParticipantExplicitProvided": x.ParticipantExplicitProvided, "AgreementDialogue": x.AgreementDialogue}
		}
	}
	raw, _ := json.Marshal(map[string]any{"Id": v.ID, "Description": v.Description, "RecordingAndTranscriptionLocalizationCustomMessage": items})
	return raw
}

func sameRecordingCollections(left, right []*resource.Message) bool {
	if len(left) != len(right) {
		return false
	}
	byIdentity := map[string][]byte{}
	for _, v := range right {
		if v == nil || v.Identity == nil || *v.Identity == "" || byIdentity[*v.Identity] != nil {
			return false
		}
		byIdentity[*v.Identity] = recordingProjection(v)
	}
	for _, v := range left {
		if v == nil || v.Identity == nil || !bytes.Equal(byIdentity[*v.Identity], recordingProjection(v)) {
			return false
		}
		delete(byIdentity, *v.Identity)
	}
	return len(byIdentity) == 0
}
func recordingOracle(t *testing.T, ctx context.Context, tenant string, identities []string) map[string]json.RawMessage {
	t.Helper()
	input, e := json.Marshal(map[string]any{"tenant_id": tenant, "identities": identities})
	if e != nil {
		t.Fatal("invalid recording oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-recording-messages.ps1")
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
		t.Fatal("pinned recording reads failed; no output retained")
	}
	v, e := decodeRecordingOracle(output.data, tenant, identities)
	if e != nil {
		t.Fatal("pinned recording projection unavailable")
	}
	return v
}

// TestLabRecordingMessagesReadOnly lists and reads at most five selected records.
// It never creates messages, edits policy, changes assignments or grants roles.
func TestLabRecordingMessagesReadOnly(t *testing.T) {
	cfg, e := lab.Load()
	if e != nil {
		t.Fatal("invalid lab enablement")
	}
	if e := cfg.Require(lab.ReadOnly); errors.Is(e, lab.ErrDisabled) {
		t.Skip("recording reads require explicit read-only opt-in")
	} else if e != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	values, response, e := sdk.TeamsAdminAPI.RecordingMessages.List(ctx)
	if e != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("recording list failed (HTTP %d); no response payload retained", status)
	}
	filtered, _, e := sdk.TeamsAdminAPI.RecordingMessages.ListFiltered(ctx, "*")
	if e != nil || !sameRecordingCollections(values, filtered) {
		t.Fatal("recording filtered list differs")
	}
	identities := []string{}
	projections := map[string][]byte{}
	for _, v := range values {
		if len(identities) == 5 {
			break
		}
		if v == nil || v.Identity == nil || *v.Identity == "" {
			t.Fatal("recording identity unavailable")
		}
		read, _, e := sdk.TeamsAdminAPI.RecordingMessages.GetByIdentity(ctx, *v.Identity)
		if e != nil || read == nil || read.Identity == nil || *read.Identity != *v.Identity {
			t.Fatal("selected recording read failed; no output retained")
		}
		before, after := recordingProjection(v), recordingProjection(read)
		if !bytes.Equal(before, after) {
			t.Fatal("recording list/read projections differ")
		}
		if v.ID != nil && labDirectoryID.MatchString(*v.ID) && *v.Identity == "Global/"+*v.ID {
			relative, _, e := sdk.TeamsAdminAPI.RecordingMessages.GetByID(ctx, *v.ID)
			if e != nil || relative == nil || relative.Identity == nil || *relative.Identity != *v.Identity || !bytes.Equal(after, recordingProjection(relative)) {
				t.Fatal("recording relative-ID read differs")
			}
		}
		identities = append(identities, *v.Identity)
		projections[*v.Identity] = after
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := recordingOracle(t, ctx, tenant, identities)
		for identity, want := range projections {
			var left, right any
			if json.Unmarshal(want, &left) != nil || json.Unmarshal(actual[identity], &right) != nil || !reflect.DeepEqual(left, right) {
				t.Fatal("Go and pinned recording projections differ")
			}
		}
	}
	t.Log("bounded recording reads verified; writes, cleanup and effective behavior remain separately qualified")
}

func TestRecordingCollectionComparisonIgnoresResponseOrder(t *testing.T) {
	t.Parallel()
	a := &resource.Message{Identity: new("Global/a")}
	b := &resource.Message{Identity: new("Global/b")}
	if !sameRecordingCollections([]*resource.Message{a, b}, []*resource.Message{b, a}) {
		t.Fatal("independent response order changed the comparison")
	}
	for _, other := range [][]*resource.Message{{a, a}, {a, nil}, {a, {Identity: new("Global/b"), Description: new("changed")}}} {
		if sameRecordingCollections([]*resource.Message{a, b}, other) {
			t.Fatal("duplicate, missing or changed record accepted")
		}
	}
}

func TestRecordingOracleRejectsWrongTenantDuplicateAndMixedOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	const identity = "Global/fixture"
	settings := `{"Id":null,"Description":null,"RecordingAndTranscriptionLocalizationCustomMessage":[]}`
	observation := `{"identity":"` + identity + `","settings":` + settings + `}`
	valid := `{"success":true,"tenant_id":"` + tenant + `","observations":[` + observation + `]}`
	for _, raw := range []string{`{"success":false}`, valid + "{}", strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, identity, "Other", 1), strings.Replace(valid, `"Description":null,`, "", 1), strings.Replace(valid, observation, observation+","+observation, 1), strings.Replace(valid, `"Id":null`, `"Id":"invalid"`, 1)} {
		if _, e := decodeRecordingOracle([]byte(raw), tenant, []string{identity}); e == nil {
			t.Fatal("invalid recording oracle output accepted")
		}
	}
	if _, e := decodeRecordingOracle([]byte(valid), tenant, []string{identity}); e != nil {
		t.Fatal("complete recording projection rejected")
	}
	child := `{"Identity":null,"Language":null,"InitiatorImplicit":null,"ParticipantImplicit":null,"InitiatorExplicit":null,"ParticipantExplicitRequested":null,"ParticipantExplicitProvided":null,"AgreementDialogue":null}`
	projected := strings.Replace(valid, `"RecordingAndTranscriptionLocalizationCustomMessage":[]`, `"RecordingAndTranscriptionLocalizationCustomMessage":[`+child+`]`, 1)
	if _, e := decodeRecordingOracle([]byte(projected), tenant, []string{identity}); e != nil {
		t.Fatal("canonical absent-Language presentation rejected")
	}
	for _, bad := range []string{strings.Replace(projected, `"Language":null,`, "", 1), strings.Replace(projected, `"Language":null`, `"Language":false`, 1)} {
		if _, e := decodeRecordingOracle([]byte(bad), tenant, []string{identity}); e == nil {
			t.Fatal("invalid child projection accepted")
		}
	}
}
