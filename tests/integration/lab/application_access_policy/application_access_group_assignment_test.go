package applicationaccesspolicyacc

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
	application "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/application_access_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	assignments "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

type groupAssignmentProjection struct {
	GroupID    *string `json:"group_id"`
	PolicyType *string `json:"policy_type"`
	PolicyName *string `json:"policy_name"`
	Priority   *int32  `json:"priority"`
}

func compareGroupAssignmentOracle(t *testing.T, ctx context.Context, fixture assignmentGroupFixture, values []*assignments.GroupAssignment) {
	t.Helper()
	if !labkit.OracleEnabled(t) {
		return
	}
	input, err := json.Marshal(map[string]string{"tenant_id": fixture.TenantID, "group_id": fixture.GroupID})
	if err != nil {
		t.Fatal("invalid group oracle selector")
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-group-application-assignment.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath", "MSTEAMS_LAB_AUTH_METHOD", "MSTEAMS_LAB_MODULE_PATH", "MSTEAMS_LAB_IDENTITY_FILE", "MSTEAMS_LAB_GROUP_FIXTURE_FILE"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1")
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout = &output
	cmd.Stderr = io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned group-assignment read failed")
	}
	var report struct {
		Success     bool                        `json:"success"`
		Tenant      string                      `json:"tenant_id"`
		Group       string                      `json:"group_id"`
		Assignments []groupAssignmentProjection `json:"assignments"`
	}
	d := json.NewDecoder(bytes.NewReader(output.Data))
	d.DisallowUnknownFields()
	var trailing any
	if d.Decode(&report) != nil || d.Decode(&trailing) != io.EOF || !report.Success || !strings.EqualFold(report.Tenant, fixture.TenantID) || !strings.EqualFold(report.Group, fixture.GroupID) || report.Assignments == nil {
		t.Fatal("pinned group-assignment projection or binding differs")
	}
	want := make([]groupAssignmentProjection, 0, len(values))
	for _, value := range values {
		if value == nil {
			t.Fatal("nil group assignment")
		}
		want = append(want, groupAssignmentProjection{value.GroupID, value.PolicyType, value.PolicyName, value.Priority})
	}
	if !reflect.DeepEqual(want, report.Assignments) {
		t.Fatal("SDK and PowerShell group-assignment projections differ")
	}
	t.Log("PowerShell 8.1.0 / Go group assignment parity: passed")
}

// TestAccApplicationAccessPolicyGroupAssignment exercises the assignment
// configuration on the dedicated empty security group. It does not infer member
// propagation or an effective user policy from a successful configuration write.
func TestAccApplicationAccessPolicyGroupAssignment(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("group assignment requires explicit mutation opt-in and a private fixture")
	} else if err != nil {
		t.Fatal("invalid lab mutation enablement")
	}
	sdk, binding, appID := labkit.SDKClient(t)
	fixture := loadAssignmentGroupFixture(t, binding)
	handler := groupAssignmentCleanup{sdk, fixture.GroupID}
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private group recovery journal unavailable")
	}
	handlers := map[string]lab.Handler{"group-app-access-policy": groupOwnedPolicyCleanup{policyCleanup{sdk.TeamsAdminAPI.ApplicationAccessPolicies}, handler}, "app-access-group-assignment": handler}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, handlers); err != nil {
			t.Error("group assignment cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("group recovery journal close failed")
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	before, response, err := sdk.TeamsAdminAPI.PolicyAssignments.ListGroup(ctx, fixture.GroupID)
	if err != nil || len(before) != 0 {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("owned group must have no pre-existing assignments")
	}
	replacements := map[string]string{fixture.GroupID: "00000000-0000-4000-8000-000000000020", appID: "00000000-0000-4000-8000-000000000001"}
	labkit.WriteResponseFixture(t, "policy_assignments", "group_application_access_before", response, replacements)
	compareGroupAssignmentOracle(t, ctx, fixture, before)
	names := make([]string, 0, 2)
	for _, label := range []string{"group-access-a", "group-access-b"} {
		name, err := journal.FixtureName(label)
		if err != nil {
			t.Fatal("group policy fixture name unavailable")
		}
		entry, err := journal.PlanCreate("group-app-access-policy", name)
		if err != nil {
			t.Fatal("group policy creation intent unavailable")
		}
		description := "SDK lab non-default initial description"
		created, response, err := sdk.TeamsAdminAPI.ApplicationAccessPolicies.Create(ctx, application.CreateRequest{Identity: name, AppIDs: application.ReplaceAppIDs([]string{appID}), Description: &description})
		if err != nil || created == nil || !labkit.SamePolicyName(created.Identity, name) {
			labkit.LogAccessResponse(t, response, err)
			t.Fatal("owned group policy creation failed")
		}
		if err := journal.Applied(entry, name); err != nil {
			t.Fatal("group policy receipt not journaled")
		}
		names = append(names, name)
	}
	replacements[names[0]] = "fixture-policy-a"
	replacements[names[1]] = "fixture-policy-b"
	baseline, err := json.Marshal(before)
	if err != nil {
		t.Fatal("group baseline cannot be journaled")
	}
	entry, err := journal.PlanRestore("app-access-group-assignment", fixture.GroupID, baseline)
	if err != nil {
		t.Fatal("group restoration intent unavailable")
	}
	settings := func(name string) assignments.GroupSettings {
		return assignments.GroupSettings{PolicyName: policy.Set(name), Priority: policy.Set(int32(1))}
	}
	result, response, err := sdk.TeamsAdminAPI.PolicyAssignments.CreateGroup(ctx, fixture.GroupID, "ApplicationAccessPolicy", settings(names[0]))
	labkit.WriteResponseFixture(t, "policy_assignments", "group_application_access_create", response, replacements)
	if err != nil || result == nil || !result.HTTPAccepted {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("owned group assignment create failed")
	}
	if err := journal.Applied(entry, fixture.GroupID); err != nil {
		t.Fatal("group assignment receipt unavailable")
	}
	read := func(name, label string) {
		t.Helper()
		var selected *assignments.GroupAssignment
		if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
			value, response, err := sdk.TeamsAdminAPI.PolicyAssignments.GetGroup(ctx, fixture.GroupID, "ApplicationAccessPolicy")
			if labkit.IsHTTP(err, 404) {
				return false, nil
			}
			if err != nil {
				return false, err
			}
			matched := value != nil && value.GroupID != nil && strings.EqualFold(*value.GroupID, fixture.GroupID) && value.PolicyType != nil && *value.PolicyType == "ApplicationAccessPolicy" && value.PolicyName != nil && labkit.SamePolicyName(*value.PolicyName, name) && value.Priority != nil && *value.Priority == 1
			if matched {
				selected = value
				labkit.WriteResponseFixture(t, "policy_assignments", label+"_get", response, replacements)
			}
			return matched, nil
		}); err != nil {
			t.Fatal("group assignment read-back did not converge")
		}
		values, response, err := sdk.TeamsAdminAPI.PolicyAssignments.ListGroup(ctx, fixture.GroupID)
		if err != nil || len(values) != 1 || values[0] == nil || !labkit.SameResponseFields(selected.RawFields, values[0].RawFields) {
			t.Fatal("selected group assignment and list differ")
		}
		labkit.WriteResponseFixture(t, "policy_assignments", label+"_list", response, replacements)
		compareGroupAssignmentOracle(t, ctx, fixture, values)
	}
	read(names[0], "group_application_access_created")
	t.Log("group policy create/read-back: passed")
	result, response, err = sdk.TeamsAdminAPI.PolicyAssignments.UpdateGroup(ctx, fixture.GroupID, "ApplicationAccessPolicy", settings(names[1]))
	labkit.WriteResponseFixture(t, "policy_assignments", "group_application_access_update", response, replacements)
	if err != nil || result == nil || !result.HTTPAccepted {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("owned group assignment update failed")
	}
	read(names[1], "group_application_access_updated")
	t.Log("group policy update/read-back: passed")
	remove := func(label string) {
		t.Helper()
		result, response, err := sdk.TeamsAdminAPI.PolicyAssignments.DeleteGroup(ctx, fixture.GroupID, "ApplicationAccessPolicy")
		labkit.WriteResponseFixture(t, "policy_assignments", label+"_delete", response, replacements)
		if err != nil || result == nil || !result.HTTPAccepted {
			labkit.LogAccessResponse(t, response, err)
			t.Fatal("owned group assignment delete failed")
		}
		if err := labkit.ObserveUntil(ctx, handler.absent); err != nil {
			t.Fatal("owned group assignment absence not verified")
		}
		_, response, err = sdk.TeamsAdminAPI.PolicyAssignments.GetGroup(ctx, fixture.GroupID, "ApplicationAccessPolicy")
		if !labkit.IsHTTP(err, 404) {
			t.Fatal("deleted group assignment did not return HTTP404")
		}
		labkit.WriteResponseFixture(t, "policy_assignments", label+"_get_deleted", response, replacements)
		values, response, err := sdk.TeamsAdminAPI.PolicyAssignments.ListGroup(ctx, fixture.GroupID)
		if err != nil || len(values) != 0 {
			t.Fatal("deleted group assignment remains in list")
		}
		labkit.WriteResponseFixture(t, "policy_assignments", label+"_list_deleted", response, replacements)
		compareGroupAssignmentOracle(t, ctx, fixture, values)
	}
	remove("group_application_access")
	t.Log("group policy delete/absence: passed")
	rank := 1
	grant, response, err := sdk.TeamsAdminAPI.ApplicationAccessPolicies.AssignToGroup(ctx, fixture.GroupID, names[0], &rank)
	labkit.WriteResponseFixture(t, "application_access_policy", "assign_group", response, replacements)
	if err != nil || grant == nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("owned group grant failed")
	}
	read(names[0], "group_application_access_granted")
	t.Log("group policy grant/read-back: passed")
	remove("group_application_access_restored")
	if _, err := journal.Recover(ctx, handlers); err != nil {
		t.Fatal("group restoration or policy cleanup incomplete")
	}
	t.Log("group assignment cleanup: passed")
}

func TestUnitGroupAssignmentRecoveryRejectsForeignTargetsAndSnapshots(t *testing.T) {
	handler := groupAssignmentCleanup{groupID: "00000000-0000-4000-8000-000000000020"}
	for _, entry := range []lab.Entry{
		{ResourceID: "00000000-0000-4000-8000-000000000021", Snapshot: json.RawMessage(`[]`)},
		{ResourceID: handler.groupID, Snapshot: json.RawMessage(`null`)},
		{ResourceID: handler.groupID, Snapshot: json.RawMessage(`[{}]`)},
		{ResourceID: handler.groupID, Snapshot: json.RawMessage(`{`)},
	} {
		if !errors.Is(handler.Restore(t.Context(), entry), lab.ErrUnresolved) {
			t.Fatal("group recovery accepted a foreign target or unsupported baseline")
		}
	}
}
