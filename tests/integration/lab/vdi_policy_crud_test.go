package lab_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	vdi "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_vdi_policy"
)

type vdiPolicyCleanup struct{ service *vdi.Service }

func (h vdiPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	values, _, err := h.service.List(ctx)
	if err != nil {
		return "", err
	}
	found := ""
	for _, value := range values {
		if value == nil || value.Identity == nil {
			return "", lab.ErrUnresolved
		}
		if labkit.SamePolicyName(*value.Identity, entry.Fixture) {
			if found != "" {
				return "", lab.ErrUnresolved
			}
			found = *value.Identity
		}
	}
	if found == "" {
		_, _, err := h.service.GetByIdentity(ctx, entry.Fixture)
		if labkit.IsHTTP(err, 404) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		return "", lab.ErrUnresolved
	}
	return found, nil
}
func (h vdiPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !strings.HasPrefix(entry.Fixture, "msteams-sdk-") || !labkit.SamePolicyName(entry.ResourceID, entry.Fixture) {
		return lab.ErrUnresolved
	}
	_, _, err := h.service.DeleteByIdentity(ctx, entry.ResourceID, policy.Controls{})
	if err != nil && !labkit.IsHTTP(err, 404) {
		return err
	}
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		_, _, err := h.service.GetByIdentity(ctx, entry.ResourceID)
		if err == nil {
			return false, nil
		}
		if !labkit.IsHTTP(err, 404) {
			return false, err
		}
		values, _, err := h.service.List(ctx)
		if err != nil {
			return false, err
		}
		for _, value := range values {
			if value == nil || value.Identity == nil {
				return false, lab.ErrUnresolved
			}
			if labkit.SamePolicyName(*value.Identity, entry.Fixture) {
				return false, nil
			}
		}
		return true, nil
	})
}
func (vdiPolicyCleanup) Restore(context.Context, lab.Entry) error {
	return lab.ErrUnresolved
}

// TestAccVDIPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccVDIPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("vdi CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsVDIPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private vdi recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"vdi-policy": vdiPolicyCleanup{service}}); err != nil {
			t.Error("vdi policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("vdi recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("vdi-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("vdi-policy", name)
	if err != nil {
		t.Fatal("vdi create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	_, response, err := service.CreateByIdentity(ctx, name, vdi.CreateRequest{Settings: vdi.Settings{VDI2Optimization: policy.Set("Disabled"), DisableCallsAndMeetings: policy.Set(true), DisableAudioVideoInCallsAndMeetings: policy.Set(true)}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("vdi policy create failed; recovery journal retained")
	}
	labkit.WriteResponseFixture(t, "teams_vdi_policy", "create", response, replacements)
	var value *vdi.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && labkit.SamePolicyName(*read.Identity, name) && read.VDI2Optimization != nil && *read.VDI2Optimization == "Disabled" && read.DisableCallsAndMeetings != nil && *read.DisableCallsAndMeetings && read.DisableAudioVideoInCallsAndMeetings != nil && *read.DisableAudioVideoInCallsAndMeetings
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_vdi_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("vdi non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("vdi creation receipt not journaled")
	}
	compare := func(value *vdi.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("vdi tenant binding unavailable")
		}
		actual := vdiPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesVDIPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell vdi policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	_, response, err = service.UpdateByIdentity(ctx, identity, vdi.UpdateRequest{Settings: vdi.Settings{VDI2Optimization: policy.Set("Enabled")}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("vdi sparse update failed")
	}
	if response == nil || response.Request == nil {
		t.Fatal("vdi update request receipt unavailable")
	}
	sent, ok := response.Request.Body.([]byte)
	var fields map[string]json.RawMessage
	if !ok || json.Unmarshal(sent, &fields) != nil || len(fields) != 2 || fields["Identity"] == nil || fields["VDI2Optimization"] == nil {
		t.Fatal("vdi sparse update included an unbound field")
	}
	labkit.WriteResponseFixture(t, "teams_vdi_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.VDI2Optimization != nil && *read.VDI2Optimization == "Enabled" && read.DisableCallsAndMeetings != nil && *read.DisableCallsAndMeetings && read.DisableAudioVideoInCallsAndMeetings != nil && *read.DisableAudioVideoInCallsAndMeetings
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_vdi_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("vdi update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("vdi policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && labkit.SamePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned vdi policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_vdi_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if _, deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_vdi_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (vdiPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("vdi deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("vdi deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_vdi_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("vdi policy create/read/update/list/delete and omitted-field preservation verified")
}
