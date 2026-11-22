package teamsvirtualappointmentspolicyacc

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	virtualAppointments "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_virtual_appointments_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccVirtualAppointmentsPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccVirtualAppointmentsPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("virtual-appointments CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsVirtualAppointmentsPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private virtual-appointments recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"virtual-appointments-policy": virtualAppointmentsPolicyCleanup{service}}); err != nil {
			t.Error("virtual-appointments policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("virtual-appointments recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("virtual-appointments-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("virtual-appointments-policy", name)
	if err != nil {
		t.Fatal("virtual-appointments create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	_, response, err := service.CreateByIdentity(ctx, name, virtualAppointments.CreateRequest{Settings: virtualAppointments.Settings{EnableSmsNotifications: policy.Set(false)}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("virtual-appointments policy create failed; recovery journal retained")
	}
	labkit.WriteResponseFixture(t, "teams_virtual_appointments_policy", "create", response, replacements)
	var value *virtualAppointments.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && labkit.SamePolicyName(*read.Identity, name) && read.EnableSmsNotifications != nil && *read.EnableSmsNotifications == false
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_virtual_appointments_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("virtual-appointments non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("virtual-appointments creation receipt not journaled")
	}
	compare := func(value *virtualAppointments.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("virtual-appointments tenant binding unavailable")
		}
		actual := virtualAppointmentsPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesVirtualAppointmentsPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell virtual-appointments policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	_, response, err = service.UpdateByIdentity(ctx, identity, virtualAppointments.UpdateRequest{Settings: virtualAppointments.Settings{EnableSmsNotifications: policy.Set(true)}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("virtual-appointments sparse update failed")
	}
	if response == nil || response.Request == nil {
		t.Fatal("virtual-appointments update request receipt unavailable")
	}
	sent, ok := response.Request.Body.([]byte)
	var fields map[string]json.RawMessage
	if !ok || json.Unmarshal(sent, &fields) != nil || len(fields) != 2 || fields["Identity"] == nil || fields["EnableSmsNotifications"] == nil {
		t.Fatal("virtual-appointments sparse update included an unbound field")
	}
	labkit.WriteResponseFixture(t, "teams_virtual_appointments_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.EnableSmsNotifications != nil && *read.EnableSmsNotifications == true
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_virtual_appointments_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("virtual-appointments update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("virtual-appointments policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && labkit.SamePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned virtual-appointments policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_virtual_appointments_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if _, deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_virtual_appointments_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (virtualAppointmentsPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("virtual-appointments deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("virtual-appointments deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_virtual_appointments_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("virtual-appointments policy create/read/update/list/delete and omitted-field preservation verified")
}
