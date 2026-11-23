package teamscallholdpolicyacc

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	callHold "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_call_hold_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccCallHoldPolicyCRUD manages one uniquely named unassigned policy.
// It never changes Global or grants a policy to a tenant, user or group.
func TestAccCallHoldPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("call hold CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	raw, err := labkit.PrivateFile(os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	if err != nil {
		t.Fatal("private call-hold audio fixture configuration unavailable")
	}
	defer clear(raw)
	var fixture struct {
		AudioFileID string `json:"call_hold_audio_file_id"`
	}
	if json.Unmarshal(raw, &fixture) != nil {
		t.Fatal("invalid private audio fixture configuration")
	}
	audioFileID := fixture.AudioFileID
	if !regexp.MustCompile(`(?i)^(?:[0-9a-f]{32}|[0-9a-f]{8}(?:-[0-9a-f]{4}){3}-[0-9a-f]{12})$`).MatchString(audioFileID) {
		t.Fatal("explicit lab call-hold audio fixture ID is required")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsCallHoldPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private call hold recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"call-hold-policy": callHoldPolicyCleanup{service}}); err != nil {
			t.Error("call hold policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("call hold recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("call-hold-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("call-hold-policy", name)
	if err != nil {
		t.Fatal("call hold create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	initial := "SDK lab non-default initial description"
	_, response, err := service.CreateByIdentity(ctx, name, callHold.CreateRequest{Settings: callHold.Settings{Description: policy.Set(initial), AudioFileID: policy.Set(audioFileID)}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("call hold policy create failed; recovery journal retained")
	}
	labkit.WriteResponseFixture(t, "teams_call_hold_policy", "create", response, replacements)
	var value *callHold.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && labkit.SamePolicyName(*read.Identity, name) && read.Description != nil && *read.Description == initial && read.AudioFileID != nil && *read.AudioFileID == audioFileID
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_call_hold_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("call hold non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("call hold creation receipt not journaled")
	}
	compare := func(value *callHold.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("call hold tenant binding unavailable")
		}
		actual := callHoldPolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesCallHoldGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell call hold policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	initialValue := value
	updated := "SDK lab non-default updated description"
	_, response, err = service.UpdateByIdentity(ctx, identity, callHold.UpdateRequest{Settings: callHold.Settings{Description: policy.Set(updated)}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("call hold sparse update failed")
	}
	labkit.WriteResponseFixture(t, "teams_call_hold_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == updated && reflect.DeepEqual(read.AudioFileID, initialValue.AudioFileID) && reflect.DeepEqual(read.StreamingSourceURL, initialValue.StreamingSourceURL) && reflect.DeepEqual(read.StreamingSourceAuthType, initialValue.StreamingSourceAuthType)
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_call_hold_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("call hold update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("call hold policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && labkit.SamePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned call hold policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_call_hold_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if _, deletedResponse, err := service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_call_hold_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (callHoldPolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("call hold deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("call hold deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_call_hold_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("call hold policy create/read/update/list/delete and omitted-field preservation verified")
}
