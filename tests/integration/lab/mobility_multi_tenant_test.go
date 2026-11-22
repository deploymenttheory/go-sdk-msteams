package lab_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/go-resty/resty/v2"
)

type compiledScalarView struct {
	identity *string
	settings map[string]any
	fields   map[string]json.RawMessage
}

type compiledScalarRead func(context.Context) (*compiledScalarView, *resty.Response, error)

func mobilityRead(sdk *msteams.Client) compiledScalarRead {
	return func(ctx context.Context) (*compiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsMobilityPolicies.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &compiledScalarView{v.Identity, map[string]any{"Description": v.Description, "IPAudioMobileMode": v.IPAudioMobileMode, "IPVideoMobileMode": v.IPVideoMobileMode, "MobileDialerPreference": v.MobileDialerPreference, "LinksInTeams": v.LinksInTeams}, v.RawFields}, r, nil
	}
}
func multiTenantRead(sdk *msteams.Client) compiledScalarRead {
	return func(ctx context.Context) (*compiledScalarView, *resty.Response, error) {
		v, r, e := sdk.TeamsAdminAPI.TeamsMultiTenantOrganizationConfigurations.GetByIdentity(ctx, "Global")
		if e != nil || v == nil {
			return nil, r, e
		}
		return &compiledScalarView{v.Identity, map[string]any{"CopilotFromHomeTenant": v.CopilotFromHomeTenant}, v.RawFields}, r, nil
	}
}
func testCompiledScalarGlobal(t *testing.T, kind, family string, read func(*msteams.Client) compiledScalarRead, list func(context.Context, *msteams.Client) ([]*compiledScalarView, *resty.Response, error)) {
	t.Helper()
	cfg := configurationReadEnabled(t)
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	v, response, err := read(sdk)(ctx)
	writeResponseFixture(t, family, "get_global", response, nil)
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("compiled scalar Global read failed")
	}
	if v == nil || v.identity == nil || !strings.EqualFold(*v.identity, "Global") {
		t.Fatal("compiled scalar Global identity differs")
	}
	values, response, err := list(ctx, sdk)
	writeResponseFixture(t, family, "list", response, nil)
	if err != nil {
		logAccessResponse(t, response, err)
		t.Fatal("compiled scalar Global list request failed")
	}
	if len(values) != 1 || values[0] == nil {
		t.Fatal("compiled scalar Global list selection differs")
	}
	if !sameConfigurationFields(v.fields, values[0].fields) {
		reason := responseComparisonFailure(v.fields, values[0].fields)
		t.Log(reason)
		logMetadataMismatches(t, v.fields, values[0].fields)
		t.Fatal("compiled scalar Global read/list differ")
	}
	if !sameTemplateProjection(v.settings, values[0].settings) {
		t.Fatal("compiled scalar Global typed getters differ")
	}

	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !sameTemplateProjection(v.settings, tenantConfigurationOracle(t, ctx, kind, tenant)) {
			t.Fatal("Go and PowerShell compiled scalar getters differ")
		}
	}
	t.Log("compiled scalar Global read/list and pinned getter comparison verified")
}
func TestAccMobilityPolicyReadOnly(t *testing.T) {
	testCompiledScalarGlobal(t, "mobility", "teams_mobility_policy", mobilityRead, func(ctx context.Context, sdk *msteams.Client) ([]*compiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsMobilityPolicies.ListFiltered(ctx, "Global")
		var out []*compiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &compiledScalarView{v.Identity, map[string]any{"Description": v.Description, "IPAudioMobileMode": v.IPAudioMobileMode, "IPVideoMobileMode": v.IPVideoMobileMode, "MobileDialerPreference": v.MobileDialerPreference, "LinksInTeams": v.LinksInTeams}, v.RawFields})
			}
		}
		return out, r, e
	})
}
func TestAccMultiTenantOrganizationConfigurationReadOnly(t *testing.T) {
	testCompiledScalarGlobal(t, "multi-tenant", "teams_multi_tenant_organization_configuration", multiTenantRead, func(ctx context.Context, sdk *msteams.Client) ([]*compiledScalarView, *resty.Response, error) {
		values, r, e := sdk.TeamsAdminAPI.TeamsMultiTenantOrganizationConfigurations.ListFiltered(ctx, "Global")
		var out []*compiledScalarView
		for _, v := range values {
			if v != nil {
				out = append(out, &compiledScalarView{v.Identity, map[string]any{"CopilotFromHomeTenant": v.CopilotFromHomeTenant}, v.RawFields})
			}
		}
		return out, r, e
	})
}
