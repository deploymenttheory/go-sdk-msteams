package tenantidentityacc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	ti "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_identity"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const (
	fixtureService = "tenant_identity"
	// The owned acceptance group's display name is fixed by the lab contract.
	ownedGroupName = "go-sdk-msteams acceptance policy group"
	// A reserved fixture-form ID that cannot name a real directory object.
	probeUser = "00000000-0000-4000-8000-000000000001"
)

func tenantReplacements(t *testing.T, binding string) (string, map[string]string) {
	t.Helper()
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok || !labkit.DirectoryID.MatchString(tenant) {
		t.Fatal("tenant binding required")
	}
	return tenant, map[string]string{tenant: "00000000-0000-4000-8000-0000000000aa", strings.ToLower(tenant): "00000000-0000-4000-8000-0000000000aa"}
}

func ownedGroupID(t *testing.T, tenant string) string {
	t.Helper()
	raw, err := labkit.PrivateFile(os.Getenv("MSTEAMS_LAB_GROUP_FIXTURE_FILE"))
	if err != nil {
		t.Fatal("group fixture unavailable")
	}
	var fixture struct {
		TenantID    string `json:"tenant_id"`
		GroupID     string `json:"group_id"`
		DisplayName string `json:"display_name"`
	}
	if json.Unmarshal(raw, &fixture) != nil || !strings.EqualFold(fixture.TenantID, tenant) || !labkit.DirectoryID.MatchString(fixture.GroupID) || fixture.DisplayName != ownedGroupName {
		t.Fatal("group fixture ownership or tenant binding differs")
	}
	return strings.ToLower(fixture.GroupID)
}

type tenantIdentityObservation struct {
	Success             bool     `json:"success"`
	TenantID            string   `json:"tenant_id"`
	VerifiedDomainCount int      `json:"verified_domain_count"`
	SIPDomainCount      int      `json:"sip_domain_count"`
	AssignedPlanCount   int      `json:"assigned_plan_count"`
	GroupIDs            []string `json:"group_ids"`
	DirectoryDeprecated bool     `json:"directory_tenant_deprecated"`
	LicensingForbidden  bool     `json:"licensing_forbidden"`
}

func powerShellTenantIdentity(t *testing.T, ctx context.Context) tenantIdentityObservation {
	t.Helper()
	command := labkit.OracleCommand(ctx, "read-tenant-identity.ps1")
	command.WaitDelay = time.Second
	command.Env = []string{}
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			command.Env = append(command.Env, key+"="+value)
		}
	}
	for _, key := range []string{"MSTEAMS_LAB_IDENTITY_FILE", "MSTEAMS_LAB_MODULE_PATH", "MSTEAMS_LAB_AUTH_METHOD"} {
		command.Env = append(command.Env, key+"="+os.Getenv(key))
	}
	command.Env = append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_GROUP_SEARCH="+ownedGroupName)
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	command.Stdout, command.Stderr = &output, io.Discard
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned PowerShell tenant identity read failed; no output retained")
	}
	var observed tenantIdentityObservation
	decoder := json.NewDecoder(bytes.NewReader(output.Data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&observed) != nil || !observed.Success {
		t.Fatal("pinned PowerShell tenant identity projection unavailable")
	}
	return observed
}

func forbidden(err error) bool {
	var apiErr *client.APIError
	return labkit.IsHTTP(err, 403) && errors.As(err, &apiErr) && slices.Contains(apiErr.Codes, "Forbidden")
}

// TestAccTenantIdentityReads reads the extended tenant record and searches for
// the owned acceptance group, then records the app-only outcomes of the
// support-only directory reads and the restricted configuration reads. It
// changes nothing.
func TestAccTenantIdentityReads(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.TenantIdentity
	tenant, replacements := tenantReplacements(t, binding)
	groupID := ownedGroupID(t, tenant)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()

	value, response, err := service.GetTenant(ctx)
	if err != nil || value.ObjectID == nil || !strings.EqualFold(*value.ObjectID, tenant) || len(value.VerifiedDomains) == 0 {
		t.Fatal("tenant read failed or is bound to another tenant")
	}
	labkit.WriteResponseFixture(t, fixtureService, "tenant", response, replacements)
	t.Log("tenant read: passed")

	exact, limit := true, uint32(10)
	groups, response, err := service.SearchGroups(ctx, ti.GroupSearchRequest{SearchQuery: ownedGroupName, ExactMatchOnly: &exact, MaxResults: &limit})
	if err != nil || len(groups) != 1 || groups[0].ID == nil || !strings.EqualFold(*groups[0].ID, groupID) || groups[0].DisplayName == nil || *groups[0].DisplayName != ownedGroupName {
		t.Fatal("owned group search did not return exactly the owned group")
	}
	labkit.WriteResponseFixture(t, fixtureService, "search_groups", response, replacements)
	none, response, err := service.SearchGroups(ctx, ti.GroupSearchRequest{SearchQuery: "msteams-sdk-no-such-group-zz"})
	if err != nil || len(none) != 0 {
		t.Fatal("unmatched group search was not empty")
	}
	labkit.WriteResponseFixture(t, fixtureService, "search_groups_empty", response, replacements)
	if _, response, err = service.SearchGroups(ctx, ti.GroupSearchRequest{SearchQuery: "ab"}); !labkit.IsHTTP(err, 400) {
		t.Fatal("short group search query was not rejected")
	}
	labkit.WriteResponseFixture(t, fixtureService, "search_groups_short", response, replacements)
	t.Log("group search: passed")

	if err := service.GetOnlineDirectoryTenant(ctx); !errors.Is(err, ti.ErrOnlineDirectoryTenantDeprecated) {
		t.Fatal("online directory tenant did not report the module deprecation")
	}

	// Support-only directory reads are denied to the lab's app-only identity.
	if _, response, err = service.GetAadTenant(ctx); !forbidden(err) {
		t.Fatal("app-only AAD tenant read was not forbidden")
	}
	labkit.WriteResponseFixture(t, fixtureService, "aad_tenant_forbidden", response, replacements)
	if _, response, err = service.GetAadUser(ctx, probeUser); !forbidden(err) {
		t.Fatal("app-only AAD user read was not forbidden")
	}
	labkit.WriteResponseFixture(t, fixtureService, "aad_user_forbidden", response, replacements)
	if _, response, err = service.GetCloudTenant(ctx); !forbidden(err) {
		t.Fatal("app-only cloud tenant read was not forbidden")
	}
	labkit.WriteResponseFixture(t, fixtureService, "cloud_tenant_forbidden", response, replacements)
	if _, response, err = service.GetCloudUser(ctx, probeUser); !forbidden(err) {
		t.Fatal("app-only cloud user read was not forbidden")
	}
	labkit.WriteResponseFixture(t, fixtureService, "cloud_user_forbidden", response, replacements)
	t.Log("support-only directory reads: forbidden as expected")

	// Hosting provider and tenant licensing configurations are restricted.
	if _, response, err = service.ListHostingProviders(ctx); !forbidden(err) {
		t.Fatal("hosting provider list was not forbidden")
	}
	labkit.WriteResponseFixture(t, fixtureService, "hosting_providers_forbidden", response, replacements)
	if _, response, err = service.ListTenantLicensingConfigurations(ctx); !forbidden(err) {
		t.Fatal("tenant licensing configuration list was not forbidden")
	}
	labkit.WriteResponseFixture(t, fixtureService, "tenant_licensing_forbidden", response, replacements)
	if _, response, err = service.GetTenantLicensingConfigurationByIdentity(ctx, "Global"); !forbidden(err) {
		t.Fatal("tenant licensing configuration read was not forbidden")
	}
	labkit.WriteResponseFixture(t, fixtureService, "tenant_licensing_global_forbidden", response, replacements)
	t.Log("restricted configuration reads: forbidden as expected")

	if labkit.OracleEnabled(t) {
		observed := powerShellTenantIdentity(t, ctx)
		if !strings.EqualFold(observed.TenantID, tenant) || observed.VerifiedDomainCount != len(value.VerifiedDomains) || observed.SIPDomainCount != len(value.SIPDomains) ||
			observed.AssignedPlanCount != len(value.AssignedPlans) || len(observed.GroupIDs) != 1 || !strings.EqualFold(observed.GroupIDs[0], groupID) ||
			!observed.DirectoryDeprecated || !observed.LicensingForbidden {
			t.Fatal("PowerShell and Go tenant identity observations differ")
		}
		t.Log("PowerShell 8.1.0 / Go tenant identity parity: passed")
	}
}
