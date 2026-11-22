package lab_test

import (
	"context"
	"strings"
	"testing"
)

func TestAccVDIPolicyReadOnly(t *testing.T) {
	cfg := configurationReadEnabled(t)
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	s := sdk.TeamsAdminAPI.TeamsVDIPolicies
	v, r, e := s.GetByIdentity(ctx, "Global")
	writeResponseFixture(t, "teams_vdi_policy", "get_global", r, nil)
	if e != nil || v == nil || v.Identity == nil || *v.Identity != "Global" {
		logAccessResponse(t, r, e)
		t.Fatal("Global read failed")
	}
	values, r, e := s.ListFiltered(ctx, "Global")
	writeResponseFixture(t, "teams_vdi_policy", "list", r, nil)
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
		if !ok || !matchesVDIPolicyGetterProjection(v, vdiPolicyOracleForIdentity(t, ctx, tenant, "Global").Settings) {
			t.Fatal("pinned Global getter comparison failed")
		}
	}
	t.Log("Global read/list and PowerShell 8.1.0 getters: passed")
}
func TestAccVirtualAppointmentsPolicyReadOnly(t *testing.T) {
	cfg := configurationReadEnabled(t)
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	s := sdk.TeamsAdminAPI.TeamsVirtualAppointmentsPolicies
	v, r, e := s.GetByIdentity(ctx, "Global")
	writeResponseFixture(t, "teams_virtual_appointments_policy", "get_global", r, nil)
	if e != nil || v == nil || v.Identity == nil || *v.Identity != "Global" {
		logAccessResponse(t, r, e)
		t.Fatal("Global read failed")
	}
	values, r, e := s.ListFiltered(ctx, "Global")
	writeResponseFixture(t, "teams_virtual_appointments_policy", "list", r, nil)
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
		if !ok || !matchesVirtualAppointmentsPolicyGetterProjection(v, virtualAppointmentsPolicyOracleForIdentity(t, ctx, tenant, "Global").Settings) {
			t.Fatal("pinned Global getter comparison failed")
		}
	}
	t.Log("Global read/list and PowerShell 8.1.0 getters: passed")
}
