package lab_test

import (
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	meeting "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_policy"
)

type meetingPolicyCleanup struct{ service *meeting.Service }

func (h meetingPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	values, _, err := h.service.List(ctx)
	if err != nil {
		return "", err
	}
	found := ""
	for _, value := range values {
		if value == nil || value.Identity == nil {
			return "", lab.ErrUnresolved
		}
		if samePolicyName(*value.Identity, entry.Fixture) {
			if found != "" {
				return "", lab.ErrUnresolved
			}
			found = *value.Identity
		}
	}
	if found == "" {
		_, _, err := h.service.GetByIdentity(ctx, entry.Fixture)
		if isHTTP(err, 404) {
			return "", nil
		}
		if err != nil {
			return "", err
		}
		return "", lab.ErrUnresolved
	}
	return found, nil
}
func (h meetingPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !strings.HasPrefix(entry.Fixture, "msteams-sdk-") || !samePolicyName(entry.ResourceID, entry.Fixture) {
		return lab.ErrUnresolved
	}
	_, _, err := h.service.DeleteByIdentity(ctx, entry.ResourceID, policy.Controls{})
	if err != nil && !isHTTP(err, 404) {
		return err
	}
	return observeUntil(ctx, func(ctx context.Context) (bool, error) {
		_, _, err := h.service.GetByIdentity(ctx, entry.ResourceID)
		if err == nil {
			return false, nil
		}
		if !isHTTP(err, 404) {
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
			if samePolicyName(*value.Identity, entry.Fixture) {
				return false, nil
			}
		}
		return true, nil
	})
}
func (meetingPolicyCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }

// TestAccMeetingPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccMeetingPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("meeting CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsMeetingPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private meeting recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"meeting-policy": meetingPolicyCleanup{service}}); err != nil {
			t.Error("meeting policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("meeting recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("meeting-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("meeting-policy", name)
	if err != nil {
		t.Fatal("meeting create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	initial := "SDK lab non-default initial description"
	_, response, err := service.CreateByIdentity(ctx, name, meeting.CreateRequest{Settings: meeting.Settings{Description: policy.Set(initial), AllowIPVideo: policy.Set(false)}})
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("meeting policy create failed; recovery journal retained")
	}
	writeResponseFixture(t, "teams_meeting_policy", "create", response, replacements)
	var value *meeting.Policy
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if isHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && samePolicyName(*read.Identity, name) && read.Description != nil && *read.Description == initial && read.AllowIPVideo != nil && !*read.AllowIPVideo
		if matched {
			value = read
			writeResponseFixture(t, "teams_meeting_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("meeting non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("meeting creation receipt not journaled")
	}
	compare := func(value *meeting.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("meeting tenant binding unavailable")
		}
		actual := meetingPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesMeetingPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell meeting policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	updated := "SDK lab non-default updated description"
	_, response, err = service.UpdateByIdentity(ctx, identity, meeting.UpdateRequest{Settings: meeting.Settings{Description: policy.Set(updated)}})
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("meeting sparse update failed")
	}
	writeResponseFixture(t, "teams_meeting_policy", "update", response, replacements)
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == updated && read.AllowIPVideo != nil && !*read.AllowIPVideo
		if matched {
			value = read
			writeResponseFixture(t, "teams_meeting_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("meeting update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("meeting policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && samePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned meeting policy missing or ambiguous in list")
	}
	writeResponseFixture(t, "teams_meeting_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if _, deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		logAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		writeResponseFixture(t, "teams_meeting_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (meetingPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("meeting deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !isHTTP(err, 404) {
		t.Fatal("meeting deleted policy did not return typed404")
	}
	writeResponseFixture(t, "teams_meeting_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("meeting policy create/read/update/list/delete and omitted-field preservation verified")
}
