package teamsmeetingbroadcastpolicyacc

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	broadcast "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_broadcast_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccMeetingBroadcastPolicyCRUD manages one uniquely named unassigned policy.
// It creates non-default scheduling/recording values, sparsely updates the
// description and attendee visibility, and requires the omitted settings to
// persist. It never changes Global or grants a policy to a tenant, user or group.
func TestAccMeetingBroadcastPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("meeting broadcast CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsMeetingBroadcastPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private meeting broadcast recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"meeting-broadcast-policy": meetingBroadcastPolicyCleanup{service}}); err != nil {
			t.Error("meeting broadcast policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("meeting broadcast recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("meeting-broadcast-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("meeting-broadcast-policy", name)
	if err != nil {
		t.Fatal("meeting broadcast create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	initial := "SDK lab non-default initial description"
	const recording = "AlwaysDisabled"
	const visibility = "InvitedUsersInCompany"
	response, err := service.CreateByIdentity(ctx, name, broadcast.Settings{Description: new(initial), AllowBroadcastScheduling: new(false), BroadcastRecordingMode: new(recording)})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("meeting broadcast policy create failed; recovery journal retained")
	}
	labkit.WriteResponseFixture(t, "teams_meeting_broadcast_policy", "create", response, replacements)
	var value *broadcast.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != "" && labkit.SamePolicyName(read.Identity, name) && read.Description != nil && *read.Description == initial && read.AllowBroadcastScheduling != nil && !*read.AllowBroadcastScheduling && read.BroadcastRecordingMode != nil && *read.BroadcastRecordingMode == recording
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_meeting_broadcast_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("meeting broadcast non-default create read-back did not converge")
	}
	identity := value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("meeting broadcast creation receipt not journaled")
	}
	compare := func(value *broadcast.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("meeting broadcast tenant binding unavailable")
		}
		actual := meetingBroadcastPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesMeetingBroadcastPolicyGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell meeting broadcast policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	updated := "SDK lab non-default updated description"
	response, err = service.UpdateByIdentity(ctx, identity, broadcast.UpdateRequest{Description: new(updated), BroadcastAttendeeVisibilityMode: new(visibility)})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("meeting broadcast sparse update failed")
	}
	labkit.WriteResponseFixture(t, "teams_meeting_broadcast_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == updated && read.BroadcastAttendeeVisibilityMode != nil && *read.BroadcastAttendeeVisibilityMode == visibility && read.AllowBroadcastScheduling != nil && !*read.AllowBroadcastScheduling && read.BroadcastRecordingMode != nil && *read.BroadcastRecordingMode == recording && sameUntouched(value, read)
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_meeting_broadcast_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("meeting broadcast update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("meeting broadcast policy list failed")
	}
	count := 0
	for _, item := range values {
		if item.Identity != "" && labkit.SamePolicyName(item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned meeting broadcast policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_meeting_broadcast_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if deletedResponse, err := service.DeleteByIdentity(ctx, identity); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_meeting_broadcast_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (meetingBroadcastPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("meeting broadcast deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("meeting broadcast deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_meeting_broadcast_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("meeting broadcast policy create/read/update/list/delete and omitted-field preservation verified")
}

// sameUntouched requires every returned field except the two updated settings
// (and server-maintained metadata) to keep its created value.
func sameUntouched(before, after *broadcast.Policy) bool {
	if before == nil || after == nil || len(before.RawFields) != len(after.RawFields) {
		return false
	}
	for name, value := range before.RawFields {
		switch name {
		case "Description", "BroadcastAttendeeVisibilityMode", "ConfigMetadata", "DataSource", "Key":
			continue
		}
		other, ok := after.RawFields[name]
		if !ok || !bytes.Equal(bytes.TrimSpace(value), bytes.TrimSpace(other)) {
			return false
		}
	}
	return true
}
