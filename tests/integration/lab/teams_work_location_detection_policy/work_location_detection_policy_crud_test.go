package teamsworklocationdetectionpolicyacc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_work_location_detection_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// sameExceptEnableWorkLocationDetection reports whether two snapshots differ only
// in EnableWorkLocationDetection and server-maintained metadata, proving the
// sparse update did not reset UserSettingsDefault or other returned fields.
func sameExceptEnableWorkLocationDetection(before, after map[string]json.RawMessage) bool {
	if len(before) != len(after) {
		return false
	}
	for name, value := range before {
		switch name {
		case "EnableWorkLocationDetection", "ConfigMetadata", "DataSource":
			if _, ok := after[name]; !ok {
				return false
			}
			continue
		}
		if !bytes.Equal(value, after[name]) {
			return false
		}
	}
	return true
}

// settingsAre reports the typed EnableWorkLocationDetection and UserSettingsDefault values.
func settingsAre(value *resource.Policy, enabled bool, defaults string) bool {
	return value != nil && value.EnableWorkLocationDetection != nil && *value.EnableWorkLocationDetection == enabled &&
		value.UserSettingsDefault != nil && *value.UserSettingsDefault == defaults
}

// TestAccWorkLocationDetectionPolicyCRUD manages one uniquely named unassigned
// policy. It never changes Global or grants a policy to a tenant, user or group.
func TestAccWorkLocationDetectionPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("work location detection CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsWorkLocationDetectionPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private work location detection recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"work-location-detection-policy": workLocationDetectionPolicyCleanup{service}}); err != nil {
			t.Error("work location detection policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("work location detection recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("work-location-detection-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("work-location-detection-policy", name)
	if err != nil {
		t.Fatal("work location detection create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	_, response, err := service.CreateByIdentity(ctx, name, resource.CreateRequest{Settings: resource.Settings{EnableWorkLocationDetection: policy.Set(true), UserSettingsDefault: policy.Set("Enabled")}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("work location detection policy create failed; recovery journal retained")
	}
	labkit.WriteResponseFixture(t, "teams_work_location_detection_policy", "create", response, replacements)
	var value *resource.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && labkit.SamePolicyName(*read.Identity, name) && settingsAre(read, true, "Enabled")
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_work_location_detection_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("work location detection non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("work location detection creation receipt not journaled")
	}
	compare := func(value *resource.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("work location detection tenant binding unavailable")
		}
		actual := workLocationDetectionPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesWorkLocationDetectionPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell work location detection policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	created := value.RawFields
	updated, response, err := service.UpdateByIdentity(ctx, identity, resource.UpdateRequest{Settings: resource.Settings{EnableWorkLocationDetection: policy.Set(false)}})
	if err != nil || updated == nil || !updated.HTTPAccepted || !settingsAre(updated.Before, true, "Enabled") {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("work location detection sparse update failed")
	}
	labkit.WriteResponseFixture(t, "teams_work_location_detection_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := settingsAre(read, false, "Enabled") && sameExceptEnableWorkLocationDetection(created, read.RawFields)
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_work_location_detection_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("work location detection update or unchanged-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("work location detection policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && labkit.SamePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned work location detection policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_work_location_detection_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if _, deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_work_location_detection_policy", "delete", deletedResponse, replacements)
	}
	if err := (workLocationDetectionPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("work location detection deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("work location detection deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_work_location_detection_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("work location detection policy create/read/update/list/delete verified")
}
