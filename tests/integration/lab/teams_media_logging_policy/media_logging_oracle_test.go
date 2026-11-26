package teamsmedialoggingpolicyacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"reflect"
	"strings"
	"testing"
	"time"

	assignments "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_media_logging_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const oracleScript = "read-media-logging-policy.ps1"

var mediaLoggingScalars = []string{"Description", "AllowMediaLogging"}

// runOracle invokes the dedicated pinned read script with only the private
// files its selected kind needs; output is bounded and never logged.
func runOracle(t *testing.T, ctx context.Context, input map[string]string, fixtureEnv string) []byte {
	t.Helper()
	raw, err := json.Marshal(input)
	if err != nil {
		t.Fatal("invalid media logging oracle selector")
	}
	defer clear(raw)
	cmd := labkit.OracleCommand(ctx, oracleScript)
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	keys := []string{"MSTEAMS_LAB_AUTH_METHOD", "MSTEAMS_LAB_MODULE_PATH", "MSTEAMS_LAB_IDENTITY_FILE"}
	if fixtureEnv != "" {
		keys = append(keys, fixtureEnv)
	}
	for _, key := range keys {
		cmd.Env = append(cmd.Env, key+"="+os.Getenv(key))
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1")
	var output labkit.OracleBuffer
	cmd.Stdin = bytes.NewReader(raw)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.Exceeded {
		clear(output.Data)
		t.Fatal("pinned media logging read failed; no output retained")
	}
	if stage := labkit.ReadOracleFailureStage(output.Data); stage != "" {
		t.Log("oracle failure stage: " + stage)
	}
	return output.Data
}

func strictDecode(raw []byte, value any) bool {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	var trailing any
	return d.Decode(value) == nil && d.Decode(&trailing) == io.EOF
}

type policyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

func decodePolicyOracle(raw []byte, tenant, identity string) (policyObservation, error) {
	var v policyObservation
	if len(raw) > labkit.MaxOracleBytes || !labkit.DirectoryID.MatchString(tenant) || (identity != "Global" && identity != "Tag:Enabled") || !strictDecode(raw, &v) {
		return policyObservation{}, labkit.ErrOracle
	}
	if !v.Success || !strings.EqualFold(v.Tenant, tenant) || v.Identity != identity || len(v.Settings) != len(mediaLoggingScalars) {
		return policyObservation{}, labkit.ErrOracle
	}
	for _, name := range mediaLoggingScalars {
		if _, ok := v.Settings[name]; !ok {
			return policyObservation{}, labkit.ErrOracle
		}
	}
	var model struct {
		Description       *string
		AllowMediaLogging *bool
	}
	settings, err := json.Marshal(v.Settings)
	if err != nil || json.Unmarshal(settings, &model) != nil {
		return policyObservation{}, labkit.ErrOracle
	}
	return v, nil
}

func comparePolicyOracle(t *testing.T, ctx context.Context, tenant string, v *resource.Policy) {
	t.Helper()
	if v == nil {
		t.Fatal("missing media logging policy value")
	}
	observed, err := decodePolicyOracle(runOracle(t, ctx, map[string]string{"tenant_id": tenant, "kind": "policy", "identity": v.Identity}, ""), tenant, v.Identity)
	if err != nil {
		t.Fatal("pinned media logging policy projection unavailable")
	}
	if !labkit.SameProjection(map[string]any{"Description": v.Description, "AllowMediaLogging": v.AllowMediaLogging}, observed.Settings) {
		t.Fatal("Go and PowerShell media logging getters differ")
	}
}

type userSource struct {
	PolicyName     *string `json:"policy_name"`
	PolicyType     *string `json:"policy_type"`
	AssignmentType *string `json:"assignment_type"`
	Reference      *string `json:"reference"`
}

type userProjection struct {
	PolicyType *string      `json:"policy_type"`
	PolicyName *string      `json:"policy_name"`
	Sources    []userSource `json:"sources"`
}

func projectUser(values []*assignments.EffectivePolicy) ([]userProjection, error) {
	out := make([]userProjection, 0, len(values))
	for _, value := range values {
		if value == nil || value.PolicyType == nil || !strings.EqualFold(*value.PolicyType, policyType) {
			return nil, labkit.ErrOracle
		}
		item := userProjection{PolicyType: value.PolicyType, PolicyName: value.PolicyName, Sources: make([]userSource, 0, len(value.PolicySource))}
		for _, source := range value.PolicySource {
			if source == nil {
				return nil, labkit.ErrOracle
			}
			item.Sources = append(item.Sources, userSource{source.PolicyName, source.PolicyType, source.AssignmentType, source.Reference})
		}
		out = append(out, item)
	}
	return out, nil
}

func compareUserOracle(t *testing.T, ctx context.Context, fixture userFixture, values []*assignments.EffectivePolicy) {
	t.Helper()
	if !labkit.OracleEnabled(t) {
		return
	}
	raw := runOracle(t, ctx, map[string]string{"tenant_id": fixture.TenantID, "kind": "user", "user_id": fixture.UserID}, "MSTEAMS_LAB_USER_FIXTURE_FILE")
	defer clear(raw)
	var report struct {
		Success  bool             `json:"success"`
		Tenant   string           `json:"tenant_id"`
		User     string           `json:"user_id"`
		Policies []userProjection `json:"policies"`
	}
	if !strictDecode(raw, &report) || !report.Success || !strings.EqualFold(report.Tenant, fixture.TenantID) || !strings.EqualFold(report.User, fixture.UserID) || report.Policies == nil {
		t.Fatal("pinned user-assignment projection or binding differs")
	}
	want, err := projectUser(values)
	if err != nil || !reflect.DeepEqual(want, report.Policies) {
		t.Fatal("SDK and PowerShell user-assignment projections differ")
	}
	t.Log("PowerShell 8.1.0 / Go user assignment parity: passed")
}

type groupProjection struct {
	GroupID    *string `json:"group_id"`
	PolicyType *string `json:"policy_type"`
	PolicyName *string `json:"policy_name"`
	Priority   *int32  `json:"priority"`
}

func compareGroupOracle(t *testing.T, ctx context.Context, fixture groupFixture, values []*assignments.GroupAssignment) {
	t.Helper()
	if !labkit.OracleEnabled(t) {
		return
	}
	raw := runOracle(t, ctx, map[string]string{"tenant_id": fixture.TenantID, "kind": "group", "group_id": fixture.GroupID}, "MSTEAMS_LAB_GROUP_FIXTURE_FILE")
	defer clear(raw)
	var report struct {
		Success     bool              `json:"success"`
		Tenant      string            `json:"tenant_id"`
		Group       string            `json:"group_id"`
		Assignments []groupProjection `json:"assignments"`
	}
	if !strictDecode(raw, &report) || !report.Success || !strings.EqualFold(report.Tenant, fixture.TenantID) || !strings.EqualFold(report.Group, fixture.GroupID) || report.Assignments == nil {
		t.Fatal("pinned group-assignment projection or binding differs")
	}
	want := make([]groupProjection, 0, len(values))
	for _, value := range values {
		if value == nil {
			t.Fatal("nil group assignment")
		}
		want = append(want, groupProjection{value.GroupID, value.PolicyType, value.PolicyName, value.Priority})
	}
	if !reflect.DeepEqual(want, report.Assignments) {
		t.Fatal("SDK and PowerShell group-assignment projections differ")
	}
	t.Log("PowerShell 8.1.0 / Go media logging group grant parity: passed")
}

func TestUnitMediaLoggingOracleRejectsForeignSelectorsAndIncompleteOutput(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000001"
	good := `{"success":true,"tenant_id":"` + tenant + `","identity":"Tag:Enabled","settings":{"Description":null,"AllowMediaLogging":true}}`
	if _, err := decodePolicyOracle([]byte(good), tenant, "Tag:Enabled"); err != nil {
		t.Fatal("valid projection rejected")
	}
	for _, tc := range []struct{ raw, identity string }{
		{good, "Global"},
		{good, "msteams-sdk-foreign"},
		{strings.Replace(good, `"success":true`, `"success":false`, 1), "Tag:Enabled"},
		{strings.Replace(good, `,"AllowMediaLogging":true`, ``, 1), "Tag:Enabled"},
		{strings.Replace(good, `"AllowMediaLogging":true`, `"AllowMediaLogging":"yes"`, 1), "Tag:Enabled"},
		{strings.Replace(good, `}}`, `},"extra":1}`, 1), "Tag:Enabled"},
		{good + `{}`, "Tag:Enabled"},
	} {
		if _, err := decodePolicyOracle([]byte(tc.raw), tenant, tc.identity); err == nil {
			t.Fatal("invalid oracle projection accepted")
		}
	}
}
