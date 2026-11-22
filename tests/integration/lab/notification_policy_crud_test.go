package lab_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	notification "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_notification_and_feeds_policy"
)

type notificationPolicyCleanup struct{ service *notification.Service }

func (h notificationPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
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
func (h notificationPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
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
func (notificationPolicyCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }

// TestAccNotificationAndFeedsPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccNotificationAndFeedsPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("notification CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsNotificationAndFeedsPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private notification recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"notification-policy": notificationPolicyCleanup{service}}); err != nil {
			t.Error("notification policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("notification recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("notification-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("notification-policy", name)
	if err != nil {
		t.Fatal("notification create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	_, response, err := service.CreateByIdentity(ctx, name, notification.CreateRequest{Settings: notification.Settings{Description: policy.Set("SDK lab non-default initial description"), SuggestedFeedsEnabledType: policy.Set("EnabledUserOverride")}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("notification policy create failed; recovery journal retained")
	}
	labkit.WriteResponseFixture(t, "teams_notification_and_feeds_policy", "create", response, replacements)
	var value *notification.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && labkit.SamePolicyName(*read.Identity, name) && read.Description != nil && *read.Description == "SDK lab non-default initial description" && read.SuggestedFeedsEnabledType != nil && *read.SuggestedFeedsEnabledType == "EnabledUserOverride"
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_notification_and_feeds_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("notification non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("notification creation receipt not journaled")
	}
	compare := func(value *notification.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("notification tenant binding unavailable")
		}
		actual := notificationPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesNotificationAndFeedsPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell notification policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	_, response, err = service.UpdateByIdentity(ctx, identity, notification.UpdateRequest{Settings: notification.Settings{Description: policy.Set("SDK lab non-default updated description")}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("notification sparse update failed")
	}
	labkit.WriteResponseFixture(t, "teams_notification_and_feeds_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == "SDK lab non-default updated description" && read.SuggestedFeedsEnabledType != nil && *read.SuggestedFeedsEnabledType == "EnabledUserOverride"
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_notification_and_feeds_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("notification update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	_, response, err = service.PatchByIdentity(ctx, identity, notification.PatchRequest{Settings: notification.Settings{Description: policy.Null[string]()}})
	labkit.WriteResponseFixture(t, "teams_notification_and_feeds_policy", "clear_description", response, replacements)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("notification explicit-null description update failed")
	}
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, r, e := service.GetByIdentity(ctx, identity)
		matched := e == nil && read != nil && read.Description == nil && read.SuggestedFeedsEnabledType != nil && *read.SuggestedFeedsEnabledType == "EnabledUserOverride"
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_notification_and_feeds_policy", "clear_description_readback", r, replacements)
		}
		return matched, e
	}); err != nil {
		t.Fatal("notification explicit-null read-back did not converge")
	}
	compare(value)
	t.Log("notification explicit-null update/read-back: passed")
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("notification policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && labkit.SamePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned notification policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_notification_and_feeds_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if _, deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_notification_and_feeds_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (notificationPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("notification deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("notification deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_notification_and_feeds_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("notification policy create/read/update/list/delete and omitted-field preservation verified")
}
