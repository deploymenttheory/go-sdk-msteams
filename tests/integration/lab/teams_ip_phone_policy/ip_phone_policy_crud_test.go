package teamsipphonepolicyacc

import (
	"bytes"
	"context"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	ipphone "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_ip_phone_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccIPPhonePolicyCRUD manages one uniquely named unassigned policy.
// It creates non-default sign-in mode, hot-desking and idle-timeout values,
// sparsely updates the description, home screen and Better Together settings,
// and requires the omitted settings to persist. It never changes Global or grants a policy to a tenant, user or group.
func TestAccIPPhonePolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("IP phone CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsIPPhonePolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private IP phone recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"ip-phone-policy": ipPhonePolicyCleanup{service}}); err != nil {
			t.Error("IP phone policy cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("IP phone recovery journal close failed")
		}
	})
	name, err := journal.FixtureName("ip-phone-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("ip-phone-policy", name)
	if err != nil {
		t.Fatal("IP phone create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	initial := "SDK lab non-default initial description"
	const signIn = "CommonAreaPhoneSignIn"
	const timeout = int64(30)
	const homeScreen = "Disabled"
	const betterTogether = "Disabled"
	response, err := service.CreateByIdentity(ctx, name, ipphone.Settings{Description: new(initial), AllowHotDesking: new(false), HotDeskingIdleTimeoutInMinutes: new(timeout), SignInMode: new(signIn)})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("IP phone policy create failed; recovery journal retained")
	}
	labkit.WriteResponseFixture(t, "teams_ip_phone_policy", "create", response, replacements)
	var value *ipphone.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != "" && labkit.SamePolicyName(read.Identity, name) && read.Description != nil && *read.Description == initial && createdValues(read, signIn, timeout)
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_ip_phone_policy", "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("IP phone non-default create read-back did not converge")
	}
	identity := value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("IP phone creation receipt not journaled")
	}
	compare := func(value *ipphone.Policy) {
		if !oracle {
			return
		}
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("IP phone tenant binding unavailable")
		}
		actual := ipPhonePolicyOracleForIdentity(t, ctx, tenant, identity)
		if !matchesIPPhonePolicyGetterProjection(value, actual.Settings) {
			t.Fatal("SDK and PowerShell IP phone policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}
	t.Log("policy create/read-back: passed")
	compare(value)
	updated := "SDK lab non-default updated description"
	response, err = service.UpdateByIdentity(ctx, identity, ipphone.UpdateRequest{Description: new(updated), AllowHomeScreen: new(homeScreen), AllowBetterTogether: new(betterTogether)})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("IP phone sparse update failed")
	}
	labkit.WriteResponseFixture(t, "teams_ip_phone_policy", "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == updated && read.AllowHomeScreen != nil && *read.AllowHomeScreen == homeScreen && read.AllowBetterTogether != nil && *read.AllowBetterTogether == betterTogether && createdValues(read, signIn, timeout) && sameUntouched(value, read)
		if matched {
			value = read
			labkit.WriteResponseFixture(t, "teams_ip_phone_policy", "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("IP phone update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)
	values, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("IP phone policy list failed")
	}
	count := 0
	for _, item := range values {
		if item.Identity != "" && labkit.SamePolicyName(item.Identity, name) {
			count++
		}
	}
	if count != 1 {
		t.Fatal("owned IP phone policy missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, "teams_ip_phone_policy", "owned_list", listResponse, replacements)
	t.Log("policy list: passed")
	if deletedResponse, err := service.DeleteByIdentity(ctx, identity); err != nil {
		labkit.LogAccessResponse(t, deletedResponse, err)
		t.Fatal("owned policy delete failed")
	} else {
		labkit.WriteResponseFixture(t, "teams_ip_phone_policy", "delete", deletedResponse, map[string]string{name: "fixture-policy"})
	}
	if err := (ipPhonePolicyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("IP phone deletion and independent absence were not verified")
	}
	_, response, err = service.GetByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("IP phone deleted policy did not return typed404")
	}
	labkit.WriteResponseFixture(t, "teams_ip_phone_policy", "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
	t.Log("IP phone policy create/read/update/list/delete and omitted-field preservation verified")
}

// sameUntouched requires every returned field except the three updated settings
// (and server-maintained metadata) to keep its created value.
func sameUntouched(before, after *ipphone.Policy) bool {
	if before == nil || after == nil || len(before.RawFields) != len(after.RawFields) {
		return false
	}
	for name, value := range before.RawFields {
		switch name {
		case "Description", "AllowHomeScreen", "AllowBetterTogether", "ConfigMetadata", "DataSource", "Key":
			continue
		}
		other, ok := after.RawFields[name]
		if !ok || !bytes.Equal(bytes.TrimSpace(value), bytes.TrimSpace(other)) {
			return false
		}
	}
	return true
}

// createdValues checks the non-default creation values the sparse update must keep.
func createdValues(v *ipphone.Policy, signIn string, timeout int64) bool {
	return v.AllowHotDesking != nil && !*v.AllowHotDesking && v.HotDeskingIdleTimeoutInMinutes != nil && *v.HotDeskingIdleTimeoutInMinutes == timeout && v.SignInMode != nil && *v.SignInMode == signIn
}
