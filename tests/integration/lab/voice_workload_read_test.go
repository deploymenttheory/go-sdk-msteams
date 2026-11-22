package lab_test

import (
	"context"
	"strings"
	"testing"
)

func TestAccVoiceApplicationsPolicyReadOnly(t *testing.T) {
	cfg := configurationReadEnabled(t)
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	s := sdk.TeamsAdminAPI.TeamsVoiceApplicationsPolicies
	v, r, e := s.GetByIdentity(ctx, "Global")
	writeResponseFixture(t, "teams_voice_applications_policy", "get_global", r, nil)
	if e != nil || v == nil || v.Identity == nil || *v.Identity != "Global" {
		logAccessResponse(t, r, e)
		t.Fatal("Global read failed")
	}
	values, r, e := s.ListFiltered(ctx, "Global")
	writeResponseFixture(t, "teams_voice_applications_policy", "list", r, nil)
	if e != nil || len(values) != 1 || values[0] == nil {
		t.Fatal("Global list selection failed")
	}
	if !sameConfigurationFields(v.RawFields, values[0].RawFields) {
		reason := responseComparisonFailure(v.RawFields, values[0].RawFields)
		t.Log(reason)
		t.Fatal("Global read/list mismatch")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !matchesVoiceApplicationsPolicyGetterProjection(v, voiceApplicationsPolicyOracleForIdentity(t, ctx, tenant, "Global").Settings) {
			t.Fatal("pinned Global getter comparison failed")
		}
	}
	t.Log("Global read/list and PowerShell 8.1.0 getters: passed")
}
func TestAccWorkLoadPolicyReadOnly(t *testing.T) {
	cfg := configurationReadEnabled(t)
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	s := sdk.TeamsAdminAPI.TeamsWorkLoadPolicies
	v, r, e := s.GetByIdentity(ctx, "Global")
	writeResponseFixture(t, "teams_work_load_policy", "get_global", r, nil)
	if e != nil || v == nil || v.Identity != "Global" {
		logAccessResponse(t, r, e)
		t.Fatal("Global read failed")
	}
	values, r, e := s.ListFiltered(ctx, "Global")
	writeResponseFixture(t, "teams_work_load_policy", "list", r, nil)
	if e != nil || len(values) != 1 || values[0].RawFields == nil {
		t.Fatal("Global list selection failed")
	}
	if !sameConfigurationFields(v.RawFields, values[0].RawFields) {
		reason := responseComparisonFailure(v.RawFields, values[0].RawFields)
		t.Log(reason)
		t.Fatal("Global read/list mismatch")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !matchesWorkLoadPolicyGetterProjection(v, workLoadPolicyOracleForIdentity(t, ctx, tenant, "Global").Settings) {
			t.Fatal("pinned Global getter comparison failed")
		}
	}
	t.Log("Global read/list and PowerShell 8.1.0 getters: passed")
}
