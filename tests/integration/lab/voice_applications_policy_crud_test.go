package lab_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	voice "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_voice_applications_policy"
)

type voiceApplicationsPolicyCleanup struct{ service *voice.Service }

func (h voiceApplicationsPolicyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
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
func (h voiceApplicationsPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
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
func (voiceApplicationsPolicyCleanup) Restore(context.Context, lab.Entry) error {
	return lab.ErrUnresolved
}

// TestAccVoiceApplicationsPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccVoiceApplicationsPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("voice applications CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsVoiceApplicationsPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private voice applications recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"voice-applications-policy": voiceApplicationsPolicyCleanup{service}}); err != nil {
			t.Error("voice applications policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("voice applications recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("voice-applications-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("voice-applications-policy", name)
	if err != nil {
		t.Fatal("voice applications create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	_, response, err := service.CreateByIdentity(ctx, name, voice.CreateRequest{Settings: voice.Settings{Description: policy.Set("SDK lab non-default initial description"), AllowAutoAttendantHolidaysChange: policy.Set(true), AllowCallQueueOptOutChange: policy.Set(true)}})
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("voice applications policy create failed; recovery journal retained")
	}
	writeResponseFixture(t, "teams_voice_applications_policy", "create", response, replacements)
	var value *voice.Policy
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if isHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && samePolicyName(*read.Identity, name) && read.Description != nil && *read.Description == "SDK lab non-default initial description" && read.AllowAutoAttendantHolidaysChange != nil && *read.AllowAutoAttendantHolidaysChange && read.AllowCallQueueOptOutChange != nil && *read.AllowCallQueueOptOutChange
		if matched {
			value = read
			writeResponseFixture(t, "teams_voice_applications_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("voice applications non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("voice applications creation receipt not journaled")
	}
	compare := func(value *voice.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("voice applications tenant binding unavailable")
		}
		actual := voiceApplicationsPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesVoiceApplicationsPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell voice applications policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	voiceApplicationsExpected := voiceApplicationsProjection(value)
	voiceApplicationsExpected["Description"] = new("SDK lab non-default updated description")
	_, response, err = service.UpdateByIdentity(ctx, identity, voice.UpdateRequest{Settings: voice.Settings{Description: policy.Set("SDK lab non-default updated description")}})
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("voice applications sparse update failed")
	}
	if response == nil || response.Request == nil {
		t.Fatal("voice applications update request receipt unavailable")
	}
	sent, ok := response.Request.Body.([]byte)
	var fields map[string]json.RawMessage
	if !ok || json.Unmarshal(sent, &fields) != nil || len(fields) != 2 || fields["Identity"] == nil || fields["Description"] == nil {
		t.Fatal("voice applications sparse update included an unbound field")
	}
	writeResponseFixture(t, "teams_voice_applications_policy", "update", response, replacements)
	if err := observeUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == "SDK lab non-default updated description" && sameTemplateProjection(voiceApplicationsExpected, voiceApplicationsProjection(read))
		if matched {
			value = read
			writeResponseFixture(t, "teams_voice_applications_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("voice applications update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("voice applications policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && samePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned voice applications policy missing or ambiguous in list")
	}
	writeResponseFixture(t, "teams_voice_applications_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if _, deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		logAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		writeResponseFixture(t, "teams_voice_applications_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (voiceApplicationsPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("voice applications deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !isHTTP(err, 404) {
		t.Fatal("voice applications deleted policy did not return typed404")
	}
	writeResponseFixture(t, "teams_voice_applications_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("voice applications policy create/read/update/list/delete and omitted-field preservation verified")
}
