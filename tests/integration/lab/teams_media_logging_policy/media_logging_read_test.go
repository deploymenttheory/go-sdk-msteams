package teamsmedialoggingpolicyacc

import (
	"context"
	"strings"
	"testing"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_media_logging_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccMediaLoggingPolicyReadOnly reads both built-in instances by identity,
// the full collection and a Tag filter, and compares every typed getter with
// the pinned PowerShell 8.1.0 function. It performs no tenant writes.
func TestAccMediaLoggingPolicyReadOnly(t *testing.T) {
	cfg := labkit.ConfigurationReadEnabled(t)
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok {
		t.Fatal("canonical tenant binding required")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	s := sdk.TeamsAdminAPI.TeamsMediaLoggingPolicies
	selected := map[string]*resource.Policy{}
	for _, tc := range []struct {
		identity, fixture string
		allow             bool
	}{{"Global", "get_global", false}, {"Tag:Enabled", "get_tag_enabled", true}} {
		v, response, err := s.GetByIdentity(ctx, tc.identity)
		labkit.WriteResponseFixture(t, "teams_media_logging_policy", tc.fixture, response, nil)
		if err != nil || v == nil {
			labkit.LogAccessResponse(t, response, err)
			t.Fatal("media logging selected read failed")
		}
		if v.Identity != tc.identity || v.AllowMediaLogging == nil || *v.AllowMediaLogging != tc.allow {
			t.Fatal("built-in media logging instance differs")
		}
		selected[tc.identity] = v
	}
	values, response, err := s.List(ctx)
	labkit.WriteResponseFixture(t, "teams_media_logging_policy", "list", response, nil)
	if err != nil || len(values) < 2 || values[0].Identity != "Global" {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("media logging list failed or lost Global ordering")
	}
	for identity, v := range selected {
		found := false
		for _, listed := range values {
			if listed.Identity == identity {
				found = true
				if !labkit.SameConfigurationFields(v.RawFields, listed.RawFields) {
					t.Log(labkit.ResponseComparisonFailure(v.RawFields, listed.RawFields))
					labkit.LogMetadataMismatches(t, v.RawFields, listed.RawFields)
					t.Fatal("selected media logging read and list differ")
				}
			}
		}
		if !found {
			t.Fatal("selected media logging instance missing from list")
		}
	}
	filtered, _, err := s.ListFiltered(ctx, "Enabled")
	if err != nil || len(filtered) != 1 || filtered[0].Identity != "Tag:Enabled" {
		t.Fatal("media logging Tag filter differs")
	}
	missing, response, err := s.GetByIdentity(ctx, "Tag:msteams-sdk-absent-media-logging")
	labkit.WriteResponseFixture(t, "teams_media_logging_policy", "get_missing", response, nil)
	if missing != nil || !labkit.IsHTTP(err, 404) {
		t.Fatal("absent media logging instance did not return HTTP404")
	}
	t.Log("media logging read/list/filter: passed")
	if oracle {
		for _, identity := range []string{"Global", "Tag:Enabled"} {
			comparePolicyOracle(t, ctx, tenant, selected[identity])
		}
		t.Log("PowerShell 8.1.0 / Go media logging policy read parity: passed")
	}
}
