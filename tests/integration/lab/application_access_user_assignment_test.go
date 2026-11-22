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
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	application "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/application_access_policy"
	assignments "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

type assignmentUserFixture struct {
	TenantID      string `json:"tenant_id"`
	UserID        string `json:"user_id"`
	PrincipalName string `json:"user_principal_name"`
	DisplayName   string `json:"display_name"`
}

func loadAssignmentUserFixture(t *testing.T, binding string) assignmentUserFixture {
	t.Helper()
	raw, err := privateLabFile(os.Getenv("MSTEAMS_LAB_USER_FIXTURE_FILE"))
	if err != nil {
		t.Fatal("explicit private assignment-user fixture is required")
	}
	defer clear(raw)
	var fixture assignmentUserFixture
	tenant, _, ok := strings.Cut(binding, "|")
	if json.Unmarshal(raw, &fixture) != nil || !ok || !strings.EqualFold(fixture.TenantID, tenant) || !labDirectoryID.MatchString(fixture.UserID) || !strings.HasPrefix(fixture.PrincipalName, "go-sdk-msteams-acceptance@") || fixture.DisplayName != "go-sdk-msteams acceptance user" {
		t.Fatal("assignment fixture ownership or tenant binding differs")
	}
	return fixture
}

type userAssignmentCleanup struct {
	sdk    *msteams.Client
	userID string
}

func (userAssignmentCleanup) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}
func (userAssignmentCleanup) Delete(context.Context, lab.Entry) error { return lab.ErrUnresolved }
func (h userAssignmentCleanup) Restore(ctx context.Context, entry lab.Entry) error {
	if entry.ResourceID != h.userID || !labDirectoryID.MatchString(h.userID) {
		return lab.ErrUnresolved
	}
	var before []*assignments.EffectivePolicy
	if json.Unmarshal(entry.Snapshot, &before) != nil || before == nil || len(before) != 0 {
		return lab.ErrUnresolved
	}
	if _, err := h.sdk.TeamsAdminAPI.ApplicationAccessPolicies.AssignToUser(ctx, h.userID, ""); err != nil {
		return err
	}
	return observeUntil(ctx, func(ctx context.Context) (bool, error) {
		values, _, err := h.sdk.TeamsAdminAPI.PolicyAssignments.GetUser(ctx, h.userID, "ApplicationAccessPolicy")
		return err == nil && len(values) == 0, err
	})
}

type userAssignmentSource struct {
	PolicyName     *string `json:"policy_name"`
	PolicyType     *string `json:"policy_type"`
	AssignmentType *string `json:"assignment_type"`
	Reference      *string `json:"reference"`
}
type userAssignmentProjection struct {
	PolicyType *string                `json:"policy_type"`
	PolicyName *string                `json:"policy_name"`
	Sources    []userAssignmentSource `json:"sources"`
}

func projectUserAssignments(values []*assignments.EffectivePolicy) ([]userAssignmentProjection, error) {
	out := make([]userAssignmentProjection, 0, len(values))
	for _, value := range values {
		if value == nil || value.PolicyType == nil || !strings.EqualFold(*value.PolicyType, "ApplicationAccessPolicy") {
			return nil, errTeamOracle
		}
		item := userAssignmentProjection{PolicyType: value.PolicyType, PolicyName: value.PolicyName, Sources: make([]userAssignmentSource, 0, len(value.PolicySource))}
		for _, source := range value.PolicySource {
			if source == nil {
				return nil, errTeamOracle
			}
			item.Sources = append(item.Sources, userAssignmentSource{source.PolicyName, source.PolicyType, source.AssignmentType, source.Reference})
		}
		out = append(out, item)
	}
	return out, nil
}
func compareUserAssignmentOracle(t *testing.T, ctx context.Context, fixture assignmentUserFixture, values []*assignments.EffectivePolicy) {
	t.Helper()
	if !teamOracleEnabled(t) {
		return
	}
	input, err := json.Marshal(map[string]string{"tenant_id": fixture.TenantID, "user_id": fixture.UserID})
	if err != nil {
		t.Fatal("invalid assignment oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-user-application-assignment.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	for _, key := range []string{"MSTEAMS_LAB_AUTH_METHOD", "MSTEAMS_LAB_MODULE_PATH", "MSTEAMS_LAB_IDENTITY_FILE", "MSTEAMS_LAB_USER_FIXTURE_FILE"} {
		cmd.Env = append(cmd.Env, key+"="+os.Getenv(key))
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1")
	var output teamOracleBuffer
	defer func() { clear(output.data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout = &output
	cmd.Stderr = io.Discard
	if cmd.Run() != nil || output.exceeded {
		t.Fatal("pinned user-assignment read failed")
	}
	var report struct {
		Success  bool                       `json:"success"`
		Tenant   string                     `json:"tenant_id"`
		User     string                     `json:"user_id"`
		Policies []userAssignmentProjection `json:"policies"`
	}
	d := json.NewDecoder(bytes.NewReader(output.data))
	d.DisallowUnknownFields()
	var trailing any
	if d.Decode(&report) != nil || d.Decode(&trailing) != io.EOF || !report.Success || !strings.EqualFold(report.Tenant, fixture.TenantID) || !strings.EqualFold(report.User, fixture.UserID) || report.Policies == nil {
		t.Fatal("pinned user-assignment projection or binding differs")
	}
	want, err := projectUserAssignments(values)
	if err != nil || !reflect.DeepEqual(want, report.Policies) {
		t.Fatal("SDK and PowerShell user-assignment projections differ")
	}
	t.Log("PowerShell 8.1.0 / Go user assignment parity: passed")
}

// TestAccApplicationAccessPolicyUserAssignment targets only the dedicated SDK
// user. It journals restoration before assignment and deletes its owned policy
// after the user's original empty assignment state is independently restored.
func TestAccApplicationAccessPolicyUserAssignment(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("user assignment requires explicit mutation opt-in and a private fixture")
	} else if err != nil {
		t.Fatal("invalid lab mutation enablement")
	}
	sdk, binding, appID := labSDKClient(t)
	fixture := loadAssignmentUserFixture(t, binding)
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private assignment recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		_, err := journal.Recover(ctx, map[string]lab.Handler{"application-access-policy": policyCleanup{sdk.TeamsAdminAPI.ApplicationAccessPolicies}, "app-access-user-assignment": userAssignmentCleanup{sdk, fixture.UserID}})
		if err != nil {
			t.Error("user-assignment cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("assignment recovery journal close failed")
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	before, response, err := sdk.TeamsAdminAPI.PolicyAssignments.GetUser(ctx, fixture.UserID, "ApplicationAccessPolicy")
	if err != nil || len(before) != 0 {
		logAccessResponse(t, response, err)
		t.Fatal("owned user must have no pre-existing application-access assignment")
	}
	replacements := map[string]string{fixture.UserID: "00000000-0000-4000-8000-000000000010", fixture.PrincipalName: "fixture-user@example.invalid", appID: "00000000-0000-4000-8000-000000000001"}
	writeResponseFixture(t, "policy_assignments", "user_application_access_before", response, replacements)
	compareUserAssignmentOracle(t, ctx, fixture, before)
	name, err := journal.FixtureName("app-access")
	if err != nil {
		t.Fatal("assignment policy name unavailable")
	}
	replacements[name] = "fixture-policy"
	entry, err := journal.PlanCreate("application-access-policy", name)
	if err != nil {
		t.Fatal("assignment policy intent not journaled")
	}
	description := "SDK lab non-default initial description"
	created, response, err := sdk.TeamsAdminAPI.ApplicationAccessPolicies.Create(ctx, application.CreateRequest{Identity: name, AppIDs: application.ReplaceAppIDs([]string{appID}), Description: &description})
	if err != nil || created == nil || !samePolicyName(created.Identity, name) {
		logAccessResponse(t, response, err)
		t.Fatal("owned assignment policy creation failed")
	}
	if err := journal.Applied(entry, name); err != nil {
		t.Fatal("assignment policy receipt not journaled")
	}
	baseline, err := json.Marshal(before)
	if err != nil {
		t.Fatal("assignment baseline cannot be journaled")
	}
	restore, err := journal.PlanRestore("app-access-user-assignment", fixture.UserID, baseline)
	if err != nil {
		t.Fatal("user restoration intent not journaled")
	}
	response, err = sdk.TeamsAdminAPI.ApplicationAccessPolicies.AssignToUser(ctx, fixture.UserID, name)
	writeResponseFixture(t, "application_access_policy", "assign_user", response, replacements)
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("owned user policy assignment failed")
	}
	if err := journal.Applied(restore, fixture.UserID); err != nil {
		t.Fatal("user assignment receipt not journaled")
	}
	var assigned []*assignments.EffectivePolicy
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		values, response, err := sdk.TeamsAdminAPI.PolicyAssignments.GetUser(ctx, fixture.UserID, "ApplicationAccessPolicy")
		if err != nil {
			return false, err
		}
		matched := len(values) == 1 && values[0] != nil && values[0].PolicyType != nil && strings.EqualFold(*values[0].PolicyType, "ApplicationAccessPolicy") && values[0].PolicyName != nil && samePolicyName(*values[0].PolicyName, name)
		if matched {
			assigned = values
			writeResponseFixture(t, "policy_assignments", "user_application_access_assigned", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("owned user assignment did not converge")
	}
	t.Log("user policy assignment/read-back: passed")
	compareUserAssignmentOracle(t, ctx, fixture, assigned)
	response, err = sdk.TeamsAdminAPI.ApplicationAccessPolicies.AssignToUser(ctx, fixture.UserID, "")
	writeResponseFixture(t, "application_access_policy", "clear_user", response, replacements)
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("owned user policy clear failed")
	}
	var restored []*assignments.EffectivePolicy
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		values, response, err := sdk.TeamsAdminAPI.PolicyAssignments.GetUser(ctx, fixture.UserID, "ApplicationAccessPolicy")
		if err != nil {
			return false, err
		}
		if len(values) == 0 {
			restored = values
			writeResponseFixture(t, "policy_assignments", "user_application_access_restored", response, replacements)
			return true, nil
		}
		return false, nil
	}); err != nil {
		t.Fatal("owned user original assignment state was not restored")
	}
	t.Log("user policy clear/restoration: passed")
	compareUserAssignmentOracle(t, ctx, fixture, restored)
	if _, err := journal.Recover(ctx, map[string]lab.Handler{"application-access-policy": policyCleanup{sdk.TeamsAdminAPI.ApplicationAccessPolicies}, "app-access-user-assignment": userAssignmentCleanup{sdk, fixture.UserID}}); err != nil {
		t.Fatal("assignment restoration or policy deletion was not independently verified")
	}
	t.Log("user assignment cleanup: passed")
}

func TestUnitUserAssignmentRecoveryRejectsForeignTargetsAndSnapshots(t *testing.T) {
	t.Parallel()
	const user = "00000000-0000-4000-8000-000000000010"
	handler := userAssignmentCleanup{userID: user}
	for _, entry := range []lab.Entry{
		{ResourceID: "00000000-0000-4000-8000-000000000011", Snapshot: json.RawMessage(`[]`)},
		{ResourceID: user, Snapshot: json.RawMessage(`null`)},
		{ResourceID: user, Snapshot: json.RawMessage(`{}`)},
		{ResourceID: user, Snapshot: json.RawMessage(`[{}]`)},
		{ResourceID: user, Snapshot: json.RawMessage(`[] {}`)},
	} {
		if !errors.Is(handler.Restore(t.Context(), entry), lab.ErrUnresolved) {
			t.Fatal("recovery accepted an unowned target or unsupported baseline")
		}
	}
}
