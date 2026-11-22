package lab_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	recordingRollOut "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_recording_roll_out_policy"
)

type recordingRollOutPolicyCleanup struct{ service *recordingRollOut.Service }

func (h recordingRollOutPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
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
func (h recordingRollOutPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
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
func (recordingRollOutPolicyCleanup) Restore(context.Context, lab.Entry) error {
	return lab.ErrUnresolved
}

// TestAccRecordingRollOutPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccRecordingRollOutPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("recording rollout CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsRecordingRollOutPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private recording rollout recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"recording-rollout-policy": recordingRollOutPolicyCleanup{service}}); err != nil {
			t.Error("recording rollout policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("recording rollout recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("recording-rollout-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("recording-rollout-policy", name)
	if err != nil {
		t.Fatal("recording rollout create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	_, response, err := service.CreateByIdentity(ctx, name, recordingRollOut.CreateRequest{Settings: recordingRollOut.Settings{MeetingRecordingOwnership: policy.Set("MeetingInitiator")}})
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("recording rollout policy create failed; recovery journal retained")
	}
	writeResponseFixture(t, "teams_recording_roll_out_policy", "create", response, replacements)
	var value *recordingRollOut.Policy
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if isHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && samePolicyName(*read.Identity, name) && read.MeetingRecordingOwnership != nil && *read.MeetingRecordingOwnership == "MeetingInitiator"
		if matched {
			value = read
			writeResponseFixture(t, "teams_recording_roll_out_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("recording rollout non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("recording rollout creation receipt not journaled")
	}
	compare := func(value *recordingRollOut.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("recording rollout tenant binding unavailable")
		}
		actual := recordingRollOutPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesRecordingRollOutPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell recording rollout policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	_, response, err = service.UpdateByIdentity(ctx, identity, recordingRollOut.UpdateRequest{Settings: recordingRollOut.Settings{MeetingRecordingOwnership: policy.Set("MeetingOrganizer")}})
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("recording rollout sparse update failed")
	}
	if response == nil || response.Request == nil {
		t.Fatal("recording rollout update request receipt unavailable")
	}
	sent, ok := response.Request.Body.([]byte)
	var fields map[string]json.RawMessage
	if !ok || json.Unmarshal(sent, &fields) != nil || len(fields) != 2 || fields["Identity"] == nil || fields["MeetingRecordingOwnership"] == nil {
		t.Fatal("recording rollout sparse update included an unbound field")
	}
	writeResponseFixture(t, "teams_recording_roll_out_policy", "update", response, replacements)
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.MeetingRecordingOwnership != nil && *read.MeetingRecordingOwnership == "MeetingOrganizer"
		if matched {
			value = read
			writeResponseFixture(t, "teams_recording_roll_out_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("recording rollout update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("recording rollout policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && samePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned recording rollout policy missing or ambiguous in list")
	}
	writeResponseFixture(t, "teams_recording_roll_out_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if _, deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		logAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		writeResponseFixture(t, "teams_recording_roll_out_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (recordingRollOutPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("recording rollout deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !isHTTP(err, 404) {
		t.Fatal("recording rollout deleted policy did not return typed404")
	}
	writeResponseFixture(t, "teams_recording_roll_out_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("recording rollout policy create/read/update/list/delete and omitted-field preservation verified")
}
