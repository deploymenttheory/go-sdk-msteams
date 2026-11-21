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
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_feedback_policy"
)

var feedbackPolicyScalars = []string{"UserInitiatedMode", "ReceiveSurveysMode", "AllowScreenshotCollection", "AllowEmailCollection", "AllowLogCollection", "EnableFeatureSuggestions"}

type feedbackPolicyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

func decodeFeedbackPolicyOracle(raw []byte, tenant string) (feedbackPolicyObservation, error) {
	var v feedbackPolicyObservation
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || !strings.EqualFold(v.Tenant, tenant) || !strings.EqualFold(v.Identity, "Global") || len(v.Settings) != len(feedbackPolicyScalars) {
		return feedbackPolicyObservation{}, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return feedbackPolicyObservation{}, errTeamOracle
	}
	for _, name := range feedbackPolicyScalars {
		if _, ok := v.Settings[name]; !ok {
			return feedbackPolicyObservation{}, errTeamOracle
		}
	}
	settings, err := json.Marshal(v.Settings)
	var model resource.Policy
	if err != nil || json.Unmarshal(settings, &model) != nil {
		return feedbackPolicyObservation{}, errTeamOracle
	}
	return v, nil
}

func feedbackPolicyOracle(t *testing.T, ctx context.Context, tenant string) feedbackPolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant})
	if err != nil {
		t.Fatal("invalid feedback policy oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-feedback-policy.ps1")
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
		t.Fatal("pinned feedback policy read failed; no output retained")
	}
	value, err := decodeFeedbackPolicyOracle(output.data, tenant)
	if err != nil {
		t.Fatal("pinned feedback policy projection unavailable")
	}
	return value
}

// TestAccFeedbackPolicyReadOnly performs configuration reads only.
// Creates, updates, deletes, grants and effective feedback behavior need separate acceptance.
func TestAccFeedbackPolicyReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("feedback policy reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.TeamsFeedbackPolicies.GetByIdentity(ctx, "Global")
	writeResponseFixture(t, "teams_feedback_policy", "get_global", response, nil)
	if err != nil {
		status := 0
		phase := "preparation"
		if response != nil {
			status = response.StatusCode()
			if response.Request != nil && strings.Contains(response.Request.URL, "/Skype.Policy/configurations/TeamsFeedbackPolicy/configuration/Global") {
				phase = "resource"
			}
		}
		t.Fatalf("feedback policy %s Global read failed (HTTP %d); no response payload retained", phase, status)
	}
	listed, listResponse, err := sdk.TeamsAdminAPI.TeamsFeedbackPolicies.ListFiltered(ctx, "Global")
	writeResponseFixture(t, "teams_feedback_policy", "list", listResponse, nil)
	selected, selectionErr := globalFeedbackPolicy(listed)
	if err != nil || selectionErr != nil || value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, "Global") || !bytes.Equal(feedbackPolicyProjection(value), feedbackPolicyProjection(selected)) {
		t.Fatal("feedback policy Global read/list projections differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := feedbackPolicyOracle(t, ctx, tenant)
		if !matchesFeedbackPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("Go and pinned feedback policy getter projections differ")
		}

	}
	t.Log("selected feedback policy reads verified; persistence and effective feedback behavior remain separately qualified")
}

func TestUnitFeedbackPolicyOracleRejectsIncompleteWrongTenantAndMixedOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := map[string]any{}
	for _, name := range feedbackPolicyScalars {
		settings[name] = nil
	}
	settings["UserInitiatedMode"] = ""
	settings["AllowEmailCollection"] = false
	settings["AllowLogCollection"] = true
	raw, err := json.Marshal(feedbackPolicyObservation{Success: true, Tenant: tenant, Identity: "Global", Settings: func() map[string]json.RawMessage {
		encoded, _ := json.Marshal(settings)
		var value map[string]json.RawMessage
		_ = json.Unmarshal(encoded, &value)
		return value
	}()})
	if err != nil {
		t.Fatal(err)
	}
	valid := string(raw)
	for _, input := range []string{`{"success":false}`, strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, "Global", "Other", 1), valid + "{}", strings.Replace(valid, `"UserInitiatedMode":""`, `"UserInitiatedMode":false`, 1), strings.Replace(valid, `"AllowLogCollection":true`, `"AllowLogCollection":"true"`, 1), strings.Replace(valid, `,"AllowLogCollection":true`, "", 1)} {
		if _, e := decodeFeedbackPolicyOracle([]byte(input), tenant); e == nil {
			t.Fatal("invalid feedback policy oracle accepted")
		}
	}
	if _, e := decodeFeedbackPolicyOracle(raw, tenant); e != nil {
		t.Fatal("complete nullable feedback policy projection rejected")
	}
}

func feedbackPolicyProjection(v *resource.Policy) []byte {
	raw, _ := json.Marshal(map[string]any{"UserInitiatedMode": v.UserInitiatedMode, "ReceiveSurveysMode": v.ReceiveSurveysMode, "AllowScreenshotCollection": v.AllowScreenshotCollection, "AllowEmailCollection": v.AllowEmailCollection, "AllowLogCollection": v.AllowLogCollection, "EnableFeatureSuggestions": v.EnableFeatureSuggestions})
	return raw
}
func globalFeedbackPolicy(values []*resource.Policy) (*resource.Policy, error) {
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
func TestUnitFeedbackPolicyGlobalSelection(t *testing.T) {
	t.Parallel()
	global := &resource.Policy{Identity: new("Global")}
	tagged := &resource.Policy{Identity: new("Tag:Global")}
	if v, e := globalFeedbackPolicy([]*resource.Policy{tagged, global}); e != nil || v != global {
		t.Fatal("tagged match hid Global")
	}
	if _, e := globalFeedbackPolicy([]*resource.Policy{global, {Identity: new("global")}}); e == nil {
		t.Fatal("ambiguous Global accepted")
	}
}

func matchesFeedbackPolicyGetterProjection(value *resource.Policy, settings map[string]json.RawMessage) bool {
	if value == nil {
		return false
	}
	raw, err := json.Marshal(settings)
	if err != nil {
		return false
	}
	var left, right any
	return json.Unmarshal(feedbackPolicyProjection(value), &left) == nil && json.Unmarshal(raw, &right) == nil && reflect.DeepEqual(left, right)
}

func TestUnitFeedbackPolicyOracleUsesLastCaseInsensitiveGetterValue(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"userinitiatedmode":"fixture"}`, `{"UserInitiatedMode":"earlier","userinitiatedmode":"fixture"}`, `{"UserInitiatedMode":null,"userinitiatedmode":"fixture"}`} {
		var value resource.Policy
		if json.Unmarshal([]byte(raw), &value) != nil {
			t.Fatal("invalid synthetic model")
		}
		settings := map[string]json.RawMessage{}
		for _, name := range feedbackPolicyScalars {
			settings[name] = json.RawMessage("null")
		}
		settings["UserInitiatedMode"] = json.RawMessage(`"fixture"`)
		if !matchesFeedbackPolicyGetterProjection(&value, settings) {
			t.Fatal("canonical getter comparison used raw key spelling/earlier duplicate")
		}
	}
}
