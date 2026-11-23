package teamsshiftspolicyacc

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	shifts "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_shifts_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccShiftsPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccShiftsPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("Shifts CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsShiftsPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private Shifts recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"shifts-policy": shiftsPolicyCleanup{service}}); err != nil {
			t.Error("Shifts policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("Shifts recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("shifts-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("shifts-policy", name)
	if err != nil {
		t.Fatal("Shifts create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	_, response, err := service.CreateByIdentity(ctx, name, shifts.CreateRequest{Settings: shifts.Settings{ShiftNoticeFrequency: policy.Set("Always"), AccessGracePeriodMinutes: policy.Set(int64(5)), EnableScheduleOwnerPermissions: policy.Set(true)}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("Shifts policy create failed; recovery journal retained")
	}
	labkit.WriteResponseFixture(t, "teams_shifts_policy", "create", response, replacements)
	var value *shifts.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && labkit.SamePolicyName(*read.Identity, name) && read.ShiftNoticeFrequency != nil && *read.ShiftNoticeFrequency == "Always" && read.AccessGracePeriodMinutes != nil && *read.AccessGracePeriodMinutes == 5 && read.EnableScheduleOwnerPermissions != nil && *read.EnableScheduleOwnerPermissions
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_shifts_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("Shifts non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("Shifts creation receipt not journaled")
	}
	compare := func(value *shifts.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("Shifts tenant binding unavailable")
		}
		actual := shiftsPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesShiftsPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell Shifts policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	_, response, err = service.UpdateByIdentity(ctx, identity, shifts.UpdateRequest{Settings: shifts.Settings{AccessGracePeriodMinutes: policy.Set(int64(10))}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("Shifts sparse update failed")
	}
	labkit.WriteResponseFixture(t, "teams_shifts_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.AccessGracePeriodMinutes != nil && *read.AccessGracePeriodMinutes == 10 && read.ShiftNoticeFrequency != nil && *read.ShiftNoticeFrequency == "Always" && read.EnableScheduleOwnerPermissions != nil && *read.EnableScheduleOwnerPermissions
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_shifts_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("Shifts update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("Shifts policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && labkit.SamePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned Shifts policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_shifts_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if _, deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_shifts_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (shiftsPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("Shifts deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("Shifts deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_shifts_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("Shifts policy create/read/update/list/delete and omitted-field preservation verified")
}
