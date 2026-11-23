package resourceaccountsacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	ra "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/resource_accounts"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const (
	fixtureService = "resource_accounts"
	// Reserved fixture-form IDs that cannot name a real directory object, so
	// every probe below targets an account or configuration that does not exist.
	probeEndpoint      = "00000000-0000-4000-8000-000000000001"
	probeConfiguration = "00000000-0000-4000-8000-000000000002"
	probeStatusCode    = "ApplicationInstanceAssociationStatusProvider_AppEndpointNotFound"
)

func tenantReplacements(t *testing.T, binding string) map[string]string {
	t.Helper()
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok || !labkit.DirectoryID.MatchString(tenant) {
		t.Fatal("tenant binding required")
	}
	return map[string]string{tenant: "00000000-0000-4000-8000-0000000000aa", strings.ToLower(tenant): "00000000-0000-4000-8000-0000000000aa"}
}

type resourceAccountObservation struct {
	Success     bool   `json:"success"`
	ListCount   int    `json:"list_count"`
	SearchCount int    `json:"search_count"`
	StatusCode  string `json:"status_code"`
}

func powerShellResourceAccounts(t *testing.T, ctx context.Context) resourceAccountObservation {
	t.Helper()
	command := labkit.OracleCommand(ctx, "read-resource-accounts.ps1")
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
	command.Env = append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_RESOURCE_ACCOUNT_PROBE="+probeEndpoint)
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	command.Stdout, command.Stderr = &output, io.Discard
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned PowerShell resource account read failed; no output retained")
	}
	var observed resourceAccountObservation
	decoder := json.NewDecoder(bytes.NewReader(output.Data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&observed) != nil || !observed.Success {
		t.Fatal("pinned PowerShell resource account projection unavailable")
	}
	return observed
}

// TestAccResourceAccountReads lists and searches resource accounts and reads
// the association and status of a nonexistent endpoint. It changes nothing.
func TestAccResourceAccountReads(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.ResourceAccounts
	replacements := tenantReplacements(t, binding)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()

	size := int32(100)
	listed, response, err := service.List(ctx, &ra.ListOptions{ResultSize: &size})
	if err != nil || listed == nil {
		t.Fatal("resource account list failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "list", response, replacements)
	if _, response, err = service.GetByID(ctx, probeEndpoint); !labkit.IsHTTP(err, 404) {
		t.Fatal("missing resource account was not a 404")
	}
	labkit.WriteResponseFixture(t, fixtureService, "get_missing", response, replacements)
	t.Log("resource account list/get: passed")

	query, limit := "msteams-sdk", int32(100)
	found, response, err := service.Search(ctx, ra.SearchRequest{SearchQuery: query, MaxResults: &limit})
	if err != nil || found == nil {
		t.Fatal("resource account search failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "search", response, replacements)
	t.Log("resource account search: passed")

	if _, response, err = service.GetAssociationByID(ctx, probeEndpoint); !labkit.IsHTTP(err, 404) {
		t.Fatal("missing association was not a 404")
	}
	labkit.WriteResponseFixture(t, fixtureService, "association_missing", response, replacements)
	status, response, err := service.GetAssociationStatusByID(ctx, probeEndpoint)
	if err != nil || !status.Failed() || *status.ErrorCode != probeStatusCode {
		t.Fatal("association status for a missing endpoint changed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "association_status", response, replacements)
	t.Log("resource account association reads: passed")

	if labkit.OracleEnabled(t) {
		observed := powerShellResourceAccounts(t, ctx)
		if observed.ListCount != len(listed) || observed.SearchCount != len(found) || !strings.HasSuffix(probeStatusCode, "_"+observed.StatusCode) {
			t.Fatal("PowerShell and Go resource account observations differ")
		}
		t.Log("PowerShell 8.1.0 / Go resource account parity: passed")
	}
}

// TestAccResourceAccountRejectedWrites sends each write operation where the
// service rejects it or where its target does not exist, proving the wire
// contract without creating or changing a directory object. App-only
// credentials cannot create or update resource accounts.
func TestAccResourceAccountRejectedWrites(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.ResourceAccounts
	replacements := tenantReplacements(t, binding)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	app, name := ra.CallQueueApplicationID, "msteams-sdk-probe"

	// example.invalid is never a tenant domain, so nothing can be created.
	_, response, err := service.Create(ctx, ra.CreateRequest{UserPrincipalName: "msteams-sdk-probe@example.invalid", ApplicationID: &app, DisplayName: &name})
	if !labkit.IsHTTP(err, 400) {
		t.Fatal("unverified-domain create was not rejected")
	}
	labkit.WriteResponseFixture(t, fixtureService, "create_unverified_domain", response, replacements)

	if _, response, err = service.UpdateByID(ctx, probeEndpoint, ra.UpdateRequest{DisplayName: &name}); !labkit.IsHTTP(err, 400) {
		t.Fatal("update without application ID was not rejected")
	}
	labkit.WriteResponseFixture(t, fixtureService, "update_without_application", response, replacements)
	if _, response, err = service.UpdateByID(ctx, probeEndpoint, ra.UpdateRequest{ApplicationID: &app}); !labkit.IsHTTP(err, 400) {
		t.Fatal("app-only application ID update was not rejected")
	}
	labkit.WriteResponseFixture(t, fixtureService, "update_app_only", response, replacements)
	number := "+15555550100"
	if _, response, err = service.UpdateTelephoneNumberByID(ctx, probeEndpoint, &number); !labkit.IsHTTP(err, 400) {
		t.Fatal("deprecated telephone number update was not rejected")
	}
	labkit.WriteResponseFixture(t, fixtureService, "telephone_number_deprecated", response, replacements)
	t.Log("resource account create/update rejection: passed")

	if response, err = service.SyncByID(ctx, probeEndpoint, ra.SyncRequest{}); !labkit.IsHTTP(err, 400) {
		t.Fatal("sync without IDs was not rejected")
	}
	labkit.WriteResponseFixture(t, fixtureService, "sync_without_ids", response, replacements)
	if response, err = service.SyncByID(ctx, probeEndpoint, ra.SyncRequest{ApplicationID: &app}); err != nil || response.StatusCode() != 204 {
		t.Fatal("sync request was not accepted")
	}
	t.Log("resource account sync: passed")

	priority := int16(3)
	results, response, err := service.CreateAssociations(ctx, ra.AssociationRequest{EndpointIDs: []string{probeEndpoint}, ConfigurationID: probeConfiguration,
		ConfigurationType: ra.ConfigurationTypeCallQueue, CallPriority: &priority})
	if err != nil || len(results) != 1 || results[0].Association != nil || !results[0].Status.Failed() ||
		*results[0].Status.ErrorCode != "CreateApplicationInstanceAssociation_AppEndpointNotFound" {
		t.Fatal("association create for a missing endpoint changed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "associate_missing", response, replacements)
	results, response, err = service.DeleteAssociationByID(ctx, probeEndpoint)
	if err != nil || len(results) != 1 || !results[0].Status.Failed() || *results[0].Status.ErrorCode != "RemoveApplicationInstanceAssociation_AssociationNotFound" {
		t.Fatal("association remove for a missing endpoint changed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "unassociate_missing", response, replacements)
	t.Log("resource account association writes: passed")
}
