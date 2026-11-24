package teamsmedialoggingpolicyacc

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	assignments "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_media_logging_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccMediaLoggingPolicyGroupGrant grants the built-in Enabled instance to
// the dedicated empty security group through the public GrantToGroup set, reads
// the assignment configuration back, then removes it and verifies absence. It
// does not infer member propagation from the group configuration write.
func TestAccMediaLoggingPolicyGroupGrant(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("group grant requires explicit mutation opt-in and a private fixture")
	} else if err != nil {
		t.Fatal("invalid lab mutation enablement")
	}
	sdk, binding, appID := labkit.SDKClient(t)
	fixture := loadGroupFixture(t, binding)
	handler := groupGrantCleanup{sdk, fixture.GroupID}
	handlers := map[string]lab.Handler{"media-logging-group-assignment": handler}
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private group recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, handlers); err != nil {
			t.Error("group grant cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("group recovery journal close failed")
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	before, response, err := sdk.TeamsAdminAPI.PolicyAssignments.ListGroup(ctx, fixture.GroupID)
	if err != nil || before == nil || len(before) != 0 {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("owned group must have no pre-existing assignments")
	}
	compareGroupOracle(t, ctx, fixture, before)
	replacements := map[string]string{fixture.GroupID: "00000000-0000-4000-8000-000000000020", appID: "00000000-0000-4000-8000-000000000001"}
	baseline, err := json.Marshal(before)
	if err != nil {
		t.Fatal("group baseline cannot be journaled")
	}
	entry, err := journal.PlanRestore("media-logging-group-assignment", fixture.GroupID, baseline)
	if err != nil {
		t.Fatal("group restoration intent unavailable")
	}
	const rank = int32(1)
	grant, response, err := sdk.TeamsAdminAPI.TeamsMediaLoggingPolicies.Assign(ctx, resource.Assignment{Group: &resource.GroupGrant{ID: fixture.GroupID, PolicyName: builtInPolicy, Rank: new(rank)}})
	labkit.WriteResponseFixture(t, "teams_media_logging_policy", "assign_group", response, replacements)
	if err != nil || grant == nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("owned group media logging grant failed")
	}
	if err := journal.Applied(entry, fixture.GroupID); err != nil {
		t.Fatal("group grant receipt unavailable")
	}
	var selected *assignments.GroupAssignment
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		value, _, err := sdk.TeamsAdminAPI.PolicyAssignments.GetGroup(ctx, fixture.GroupID, policyType)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := value != nil && value.GroupID != nil && strings.EqualFold(*value.GroupID, fixture.GroupID) && value.PolicyType != nil && *value.PolicyType == policyType && value.PolicyName != nil && labkit.SamePolicyName(*value.PolicyName, builtInPolicy) && value.Priority != nil && *value.Priority == rank
		if matched {
			selected = value
		}
		return matched, nil
	}); err != nil {
		t.Fatal("group grant read-back did not converge")
	}
	values, _, err := sdk.TeamsAdminAPI.PolicyAssignments.ListGroup(ctx, fixture.GroupID)
	if err != nil || len(values) != 1 || values[0] == nil || !labkit.SameResponseFields(selected.RawFields, values[0].RawFields) {
		t.Fatal("selected group assignment and list differ")
	}
	t.Log("group policy grant/read-back: passed")
	compareGroupOracle(t, ctx, fixture, values)
	result, response, err := sdk.TeamsAdminAPI.PolicyAssignments.DeleteGroup(ctx, fixture.GroupID, policyType)
	if err != nil || result == nil || !result.HTTPAccepted {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("owned group assignment removal failed")
	}
	if err := labkit.ObserveUntil(ctx, handler.absent); err != nil {
		t.Fatal("owned group assignment absence not verified")
	}
	values, _, err = sdk.TeamsAdminAPI.PolicyAssignments.ListGroup(ctx, fixture.GroupID)
	if err != nil || values == nil || len(values) != 0 {
		t.Fatal("removed group assignment remains in list")
	}
	t.Log("group policy delete/absence: passed")
	compareGroupOracle(t, ctx, fixture, values)
	if _, err := journal.Recover(ctx, handlers); err != nil {
		t.Fatal("group restoration was not independently verified")
	}
	t.Log("group assignment cleanup: passed")
}

func TestUnitGroupGrantRecoveryRejectsForeignTargetsAndSnapshots(t *testing.T) {
	t.Parallel()
	handler := groupGrantCleanup{groupID: "00000000-0000-4000-8000-000000000020"}
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
