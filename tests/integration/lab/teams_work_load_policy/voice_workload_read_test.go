package teamsworkloadpolicyacc

import (
	"context"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func TestAccWorkLoadPolicyReadOnly(t *testing.T) {
	cfg := labkit.ConfigurationReadEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	s := sdk.TeamsAdminAPI.TeamsWorkLoadPolicies
	v, r, e := s.GetByIdentity(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_work_load_policy", "get_global", r, nil)
	if e != nil || v == nil || v.Identity != "Global" {
		labkit.LogAccessResponse(t, r, e)
		t.Fatal("Global read failed")
	}
	values, r, e := s.ListFiltered(ctx, "Global")
	labkit.WriteResponseFixture(t, "teams_work_load_policy", "list", r, nil)
	if e != nil || len(values) != 1 || values[0].RawFields == nil {
		t.Fatal("Global list selection failed")
	}
	if !labkit.SameConfigurationFields(v.RawFields, values[0].RawFields) {
		reason := labkit.ResponseComparisonFailure(v.RawFields, values[0].RawFields)
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
