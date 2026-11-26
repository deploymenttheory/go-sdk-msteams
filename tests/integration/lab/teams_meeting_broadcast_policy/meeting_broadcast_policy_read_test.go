package teamsmeetingbroadcastpolicyacc

import (
	"context"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccMeetingBroadcastPolicyReadOnly reads Global, the full list and a local
// Global filter, then compares the public getters with pinned PowerShell.
func TestAccMeetingBroadcastPolicyReadOnly(t *testing.T) {
	cfg := labkit.ConfigurationReadEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TeamsMeetingBroadcastPolicies
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	global, response, err := service.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_meeting_broadcast_policy", "get_global", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("meeting broadcast policy Global read failed")
	}
	if global == nil || !strings.EqualFold(global.Identity, "Global") {
		t.Fatal("meeting broadcast policy Global identity differs")
	}
	values, response, err := service.List(ctx)
	labkit.WriteResponseFixture(t, "teams_meeting_broadcast_policy", "list", response, nil)
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("meeting broadcast policy list request failed")
	}
	if len(values) == 0 || !strings.EqualFold(values[0].Identity, "Global") {
		t.Fatal("meeting broadcast policy list did not place Global first")
	}
	if !labkit.SameConfigurationFields(global.RawFields, values[0].RawFields) {
		t.Log(labkit.ResponseComparisonFailure(global.RawFields, values[0].RawFields))
		t.Fatal("meeting broadcast policy Global read/list differ")
	}
	filtered, _, err := service.ListFiltered(ctx, "Global")
	if err != nil || len(filtered) != 1 || !labkit.SameConfigurationFields(global.RawFields, filtered[0].RawFields) {
		t.Fatal("meeting broadcast policy Global filter selection differs")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("meeting broadcast tenant binding unavailable")
		}
		actual := meetingBroadcastPolicyOracleForIdentity(t, ctx, tenant, "Global")
		if !matchesMeetingBroadcastPolicyGetterProjection(global, actual.Settings) {
			t.Fatal("Go and PowerShell meeting broadcast policy getters differ")
		}
		t.Log("PowerShell 8.1.0 / Go Global read parity: passed")
	}
	t.Log("meeting broadcast policy Global read/list/filter verified")
}
