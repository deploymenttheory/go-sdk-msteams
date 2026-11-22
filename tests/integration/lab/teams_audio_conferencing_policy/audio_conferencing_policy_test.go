package teamsaudioconferencingpolicyacc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_audio_conferencing_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

var audioConferencingPolicyFields = []string{"AllowTollFreeDialin", "MeetingInvitePhoneNumbers"}

type audioConferencingPolicyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

var audioConferencingPolicyOwnedIdentity = regexp.MustCompile(`(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-audio-policy$`)

func decodeAudioConferencingPolicyOracle(raw []byte, tenant string) (audioConferencingPolicyObservation, error) {
	return decodeAudioConferencingPolicyOracleForIdentity(raw, tenant, "Global")
}

func decodeAudioConferencingPolicyOracleForIdentity(raw []byte, tenant, identity string) (audioConferencingPolicyObservation, error) {
	var v audioConferencingPolicyObservation
	if len(raw) > labkit.MaxOracleBytes || !labkit.DirectoryID.MatchString(tenant) || !strings.EqualFold(identity, "Global") && !audioConferencingPolicyOwnedIdentity.MatchString(identity) {
		return v, labkit.ErrOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil {
		return audioConferencingPolicyObservation{}, labkit.ErrOracle
	}
	identityMatches := strings.EqualFold(v.Identity, "Global")
	if !strings.EqualFold(identity, "Global") {
		identityMatches = labkit.SamePolicyName(v.Identity, identity)
	}
	if !v.Success || !strings.EqualFold(v.Tenant, tenant) || !identityMatches || len(v.Settings) != len(audioConferencingPolicyFields) {
		return audioConferencingPolicyObservation{}, labkit.ErrOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return audioConferencingPolicyObservation{}, labkit.ErrOracle
	}
	var boolean *bool
	if raw, ok := v.Settings["AllowTollFreeDialin"]; !ok || json.Unmarshal(raw, &boolean) != nil {
		return audioConferencingPolicyObservation{}, labkit.ErrOracle
	}
	var phones []json.RawMessage
	if raw, ok := v.Settings["MeetingInvitePhoneNumbers"]; !ok || json.Unmarshal(raw, &phones) != nil {
		return audioConferencingPolicyObservation{}, labkit.ErrOracle
	}
	return v, nil
}

func audioConferencingPolicyOracle(t *testing.T, ctx context.Context, tenant string) audioConferencingPolicyObservation {
	return audioConferencingPolicyOracleForIdentity(t, ctx, tenant, "Global")
}

func audioConferencingPolicyOracleForIdentity(t *testing.T, ctx context.Context, tenant, identity string) audioConferencingPolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": identity})
	if err != nil {
		t.Fatal("invalid audio conferencing policy oracle selector")
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-audio-conferencing-policy.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"), "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned audio conferencing policy read failed; no output retained")
	}
	value, err := decodeAudioConferencingPolicyOracleForIdentity(output.Data, tenant, identity)
	if err != nil {
		t.Fatal("pinned audio conferencing policy projection unavailable")
	}
	return value
}

// TestAccAudioConferencingPolicyReadOnly never edits Global, creates resources or grants a
// policy. Tenant writes and effective assignments need separate acceptance.
func TestAccAudioConferencingPolicyReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("audio conferencing policy reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.TeamsAudioConferencingPolicies.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_audio_conferencing_policy", "get_global", response, nil)
	if err != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("audio conferencing policy Global read failed (HTTP %d); no response payload retained", status)
	}
	listed, listResponse, err := sdk.TeamsAdminAPI.TeamsAudioConferencingPolicies.ListFiltered(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_audio_conferencing_policy", "list", listResponse, nil)
	selected, selectionErr := globalAudioConferencingPolicy(listed)
	if err != nil || selectionErr != nil || value == nil || !strings.EqualFold(value.Identity, "Global") {
		t.Fatal("audio policy Global selection differs")
	}
	for _, key := range audioConferencingPolicyFields {
		var a, b any
		left, right := json.NewDecoder(bytes.NewReader(value.RawFields[key])), json.NewDecoder(bytes.NewReader(selected.RawFields[key]))
		left.UseNumber()
		right.UseNumber()
		if left.Decode(&a) != nil || right.Decode(&b) != nil || !reflect.DeepEqual(a, b) {
			t.Fatal("audio policy Global read/list projections differ")
		}
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := audioConferencingPolicyOracle(t, ctx, tenant)
		for _, key := range audioConferencingPolicyFields {
			want := value.RawFields[key]
			var left, right any
			if decodeAudioPolicyValue(want, &left) != nil || decodeAudioPolicyValue(actual.Settings[key], &right) != nil || !reflect.DeepEqual(left, right) {
				t.Fatal("Go and pinned audio conferencing policy known projections differ")
			}
		}
	}
	t.Log("selected audio conferencing policy reads verified; persistence, feature effects and assignments remain separately qualified")
}

func decodeAudioPolicyValue(raw []byte, out *any) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	return d.Decode(out)
}

func globalAudioConferencingPolicy(values []resource.Policy) (*resource.Policy, error) {
	var found *resource.Policy
	for i := range values {
		if strings.EqualFold(values[i].Identity, "Global") {
			if found != nil {
				return nil, labkit.ErrOracle
			}
			found = &values[i]
		}
	}
	if found == nil {
		return nil, labkit.ErrOracle
	}
	return found, nil
}

func TestUnitAudioConferencingPolicyOracleRejectsIncompleteWrongTenantAndMixedOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := `{"AllowTollFreeDialin":false,"MeetingInvitePhoneNumbers":["future-phone",null,{"Future":9007199254740993}]}`
	valid := `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":` + settings + `}`
	for _, raw := range []string{`{"success":false}`, strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, `"identity":"Global"`, `"identity":"Other"`, 1), valid + "{}", strings.Replace(valid, `"AllowTollFreeDialin":false`, `"AllowTollFreeDialin":0`, 1), strings.Replace(valid, `"MeetingInvitePhoneNumbers":["future-phone",null,{"Future":9007199254740993}]`, `"MeetingInvitePhoneNumbers":false`, 1), `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":{}}`} {
		if _, e := decodeAudioConferencingPolicyOracle([]byte(raw), tenant); e == nil {
			t.Fatal("invalid audio policy oracle accepted")
		}
	}
	if _, e := decodeAudioConferencingPolicyOracle([]byte(valid), tenant); e != nil {
		t.Fatal("complete nullable/future projection rejected")
	}
	if _, e := globalAudioConferencingPolicy([]resource.Policy{{Identity: "Tag:Global"}, {Identity: "Global"}}); e != nil {
		t.Fatal("tagged match hid Global")
	}
	if _, e := globalAudioConferencingPolicy([]resource.Policy{{Identity: "Global"}, {Identity: "global"}}); e == nil {
		t.Fatal("ambiguous Global accepted")
	}
}

func TestUnitAudioConferencingPolicyOracleRestrictsOwnedIdentity(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	const name = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-audio-policy"
	raw, err := os.ReadFile(labkit.Root + "/msteams/teams_admin_api/teams_audio_conferencing_policy/testdata/get_global.json")
	if err != nil {
		t.Fatal(err)
	}
	var settings map[string]json.RawMessage
	if json.Unmarshal(raw, &settings) != nil {
		t.Fatal("invalid acceptance response")
	}
	projection := map[string]json.RawMessage{}
	for _, field := range audioConferencingPolicyFields {
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
		{"wrong-kind", strings.Replace(name, "audio-policy", "calling-policy", 1), name, false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]any{"success": true, "tenant_id": tenant, "identity": tc.returned, "settings": projection})
			if err != nil {
				t.Fatal(err)
			}
			_, err = decodeAudioConferencingPolicyOracleForIdentity(body, tenant, tc.selected)
			if (err == nil) != tc.valid {
				t.Fatal("oracle accepted a different identity or rejected its owned policy")
			}
		})
	}
}
