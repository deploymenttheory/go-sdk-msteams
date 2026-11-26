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

// TestAccMediaLoggingPolicyUserGrant targets only the dedicated SDK user. It
// journals the empty direct-assignment baseline before granting the built-in
// Enabled instance, then clears it with the Grant clear form and verifies the
// original state independently. Media logging grants do not touch licensing.
func TestAccMediaLoggingPolicyUserGrant(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("user grant requires explicit mutation opt-in and a private fixture")
	} else if err != nil {
		t.Fatal("invalid lab mutation enablement")
	}
	sdk, binding, appID := labkit.SDKClient(t)
	fixture := loadUserFixture(t, binding)
	handlers := map[string]lab.Handler{"media-logging-user-assignment": userGrantCleanup{sdk, fixture.UserID}}
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private assignment recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, handlers); err != nil {
			t.Error("user grant cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("assignment recovery journal close failed")
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	read := func(ctx context.Context) ([]*assignments.EffectivePolicy, error) {
		values, response, err := sdk.TeamsAdminAPI.PolicyAssignments.GetUser(ctx, fixture.UserID, policyType)
		if err != nil {
			labkit.LogAccessResponse(t, response, err)
		}
		return values, err
	}
	before, err := read(ctx)
	if err != nil || before == nil || len(before) != 0 {
		t.Fatal("owned user must have no pre-existing media logging assignment")
	}
	compareUserOracle(t, ctx, fixture, before)
	replacements := map[string]string{fixture.UserID: "00000000-0000-4000-8000-000000000010", fixture.PrincipalName: "fixture-user@example.invalid", appID: "00000000-0000-4000-8000-000000000001"}
	baseline, err := json.Marshal(before)
	if err != nil {
		t.Fatal("assignment baseline cannot be journaled")
	}
	restore, err := journal.PlanRestore("media-logging-user-assignment", fixture.UserID, baseline)
	if err != nil {
		t.Fatal("user restoration intent not journaled")
	}
	_, response, err := sdk.TeamsAdminAPI.TeamsMediaLoggingPolicies.Assign(ctx, resource.Assignment{User: &resource.UserAssignment{Identity: fixture.UserID, PolicyName: builtInPolicy}})
	labkit.WriteResponseFixture(t, "teams_media_logging_policy", "assign_user", response, replacements)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("owned user media logging grant failed")
	}
	if err := journal.Applied(restore, fixture.UserID); err != nil {
		t.Fatal("user grant receipt not journaled")
	}
	var assigned []*assignments.EffectivePolicy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		values, err := read(ctx)
		if err != nil {
			return false, err
		}
		matched := len(values) == 1 && values[0] != nil && values[0].PolicyType != nil && strings.EqualFold(*values[0].PolicyType, policyType) && values[0].PolicyName != nil && labkit.SamePolicyName(*values[0].PolicyName, builtInPolicy)
		if matched {
			assigned = values
		}
		return matched, nil
	}); err != nil {
		t.Fatal("owned user grant did not converge")
	}
	t.Log("user policy assignment/read-back: passed")
	compareUserOracle(t, ctx, fixture, assigned)
	_, response, err = sdk.TeamsAdminAPI.TeamsMediaLoggingPolicies.Assign(ctx, resource.Assignment{User: &resource.UserAssignment{Identity: fixture.UserID}})
	labkit.WriteResponseFixture(t, "teams_media_logging_policy", "clear_user", response, replacements)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("owned user media logging clear failed")
	}
	var restored []*assignments.EffectivePolicy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		values, err := read(ctx)
		if err != nil {
			return false, err
		}
		restored = values
		return values != nil && len(values) == 0, nil
	}); err != nil {
		t.Fatal("owned user original assignment state was not restored")
	}
	t.Log("user policy clear/restoration: passed")
	compareUserOracle(t, ctx, fixture, restored)
	if _, err := journal.Recover(ctx, handlers); err != nil {
		t.Fatal("user restoration was not independently verified")
	}
	t.Log("user assignment cleanup: passed")
}

func TestUnitUserGrantRecoveryRejectsForeignTargetsAndSnapshots(t *testing.T) {
	t.Parallel()
	const user = "00000000-0000-4000-8000-000000000010"
	handler := userGrantCleanup{userID: user}
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
