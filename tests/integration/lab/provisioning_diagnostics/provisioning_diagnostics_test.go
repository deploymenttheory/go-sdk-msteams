package provisioningdiagnosticsacc

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	pd "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/provisioning_diagnostics"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
	"github.com/go-resty/resty/v2"
)

const (
	fixtureService = "provisioning_diagnostics"
	// Reserved fixture-form ID that cannot name a real directory object.
	probeID = "00000000-0000-4000-8000-000000000026"
	// example.invalid can never be a tenant SIP domain.
	probeSip = "sip:msteams-sdk-probe@example.invalid"
)

type binding struct {
	tenant string
	user   string
}

// labBinding returns the authenticated tenant and the private lab user, whose
// object ID is only ever sent as a read target.
func labBinding(t *testing.T, tenantBinding string) binding {
	t.Helper()
	tenant, _, ok := strings.Cut(tenantBinding, "|")
	if !ok || !labkit.DirectoryID.MatchString(tenant) {
		t.Fatal("tenant binding required")
	}
	raw, err := labkit.PrivateFile(os.Getenv("MSTEAMS_LAB_USER_FIXTURE_FILE"))
	if err != nil {
		t.Fatal("explicit private lab user fixture is required")
	}
	defer clear(raw)
	var fixture struct {
		TenantID string `json:"tenant_id"`
		UserID   string `json:"user_id"`
	}
	if json.Unmarshal(raw, &fixture) != nil || !strings.EqualFold(fixture.TenantID, tenant) || !labkit.DirectoryID.MatchString(fixture.UserID) {
		t.Fatal("lab user fixture tenant binding differs")
	}
	return binding{tenant: tenant, user: fixture.UserID}
}

// denied requires the expected authorization status, keeps the redacted
// response as a fixture and reports the outcome without identifiers.
func denied(t *testing.T, name string, status int, response *resty.Response, err error) {
	t.Helper()
	labkit.LogAccessResponse(t, response, err)
	want := auth.ErrForbidden
	if status == 401 {
		want = auth.ErrUnauthenticated
	}
	if response == nil || response.StatusCode() != status || !labkit.IsHTTP(err, status) || !errors.Is(err, want) {
		t.Fatalf("%s: expected HTTP %d denial changed; successful calls need separate qualification", name, status)
	}
	labkit.WriteResponseFixture(t, fixtureService, name, response, nil)
	t.Logf("%s: confirmed HTTP %d denial", name, status)
}

// TestAccProvisioningDiagnosticsReadsDenied sends every read-only diagnostic.
// App-only Teams Administrator credentials are denied each one; nothing is
// changed. The lab user is only read.
func TestAccProvisioningDiagnosticsReadsDenied(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, tenantBinding, _ := labkit.SDKClient(t)
	lab := labBinding(t, tenantBinding)
	service := sdk.TeamsAdminAPI.ProvisioningDiagnostics
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	size := int32(1)

	entities, response, err := service.GetDirectoryDiagnosticData(ctx, pd.DirectoryDiagnosticQuery{Table: "Users", Region: "NOAM", PartitionKey: lab.tenant, ResultSize: &size})
	denied(t, "directory_diagnostic_data", 403, response, err)
	changelog, response, err := service.ListObjectChangelog(ctx, pd.ObjectChangelogQuery{Identity: lab.user})
	denied(t, "object_changelog", 403, response, err)
	schema, response, err := service.ListVersionedSchemaData(ctx, pd.VersionedSchemaQuery{SchemaName: "UserAuthoredProps", Identity: lab.user})
	denied(t, "versioned_schema_data", 403, response, err)
	setting, response, err := service.GetUserSetting(ctx, lab.tenant, lab.user, "msteams-sdk-probe")
	denied(t, "user_setting", 403, response, err)
	history, response, err := service.GetUserProvisioningHistory(ctx, lab.tenant, lab.user, pd.RegionNOAM)
	denied(t, "user_provisioning_history", 403, response, err)
	user := pd.UserMove{TenantID: lab.tenant, UserSipURI: probeSip}
	policies, response, err := service.GetUserMovePolicies(ctx, user)
	denied(t, "user_move_policies", 403, response, err)
	deployment, response, err := service.ValidateUserMove(ctx, pd.ValidateUserMoveRequest{UserMove: user, CmdletVersion: "msteams-sdk-probe",
		LocalDeployment: pd.LocalDeployment{PresenceFqdn: "presence.example.invalid", RegistrarFqdn: "registrar.example.invalid", MajorVersion: 6}})
	denied(t, "validate_user_move", 403, response, err)
	if entities != nil || changelog != nil || schema != nil || setting != "" || history != nil || policies != nil || deployment != nil {
		t.Fatal("a denied read returned a value")
	}
}

// TestAccProvisioningDiagnosticsActionsDenied sends every action against a
// reserved object ID, an example.invalid SIP address, an invalid operation or
// handler name, or as a validation-only request, so none could act even if
// authorized. App-only Teams Administrator credentials are denied each one.
func TestAccProvisioningDiagnosticsActionsDenied(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, tenantBinding, _ := labkit.SDKClient(t)
	lab := labBinding(t, tenantBinding)
	service := sdk.TeamsAdminAPI.ProvisioningDiagnostics
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	validate, user, class, id := true, "User", "User", probeID
	empty := json.RawMessage(`{}`)

	_, response, err := service.SetUserSetting(ctx, lab.tenant, probeID, "msteams-sdk-probe", empty)
	denied(t, "set_user_setting", 403, response, err)
	_, response, err = service.SetTenantUserBackfill(ctx, "msteams-sdk-invalid", empty)
	denied(t, "tenant_user_backfill", 403, response, err)
	response, err = service.SyncDirectObjects(ctx, pd.DirectObjectSyncRequest{ObjectClass: &user, ObjectID: &id, IsValidationRequest: &validate})
	denied(t, "direct_object_sync", 403, response, err)
	response, err = service.SyncMsodsObjects(ctx, pd.MsodsSyncRequest{TenantID: probeID, ObjectClass: class, ObjectIDs: []string{probeID}, IsValidationRequest: &validate})
	denied(t, "msods_sync", 403, response, err)
	response, err = service.RehomeUser(ctx, probeID, empty)
	denied(t, "rehome_user", 403, response, err)
	handler, payload := "msteams-sdk-invalid", "{}"
	_, response, err = service.InvokeCustomHandler(ctx, pd.CustomHandlerPayload{HandlerFullyQualifiedName: &handler, Payload: &payload})
	denied(t, "custom_handler", 401, response, err)
	_, response, err = service.InvokeCustomHandlerCallback(ctx, pd.CustomHandlerCallbackRequest{ID: probeID, Operation: pd.StatusQueryGetURI})
	denied(t, "custom_handler_callback", 401, response, err)

	move := pd.UserMove{TenantID: lab.tenant, UserSipURI: probeSip}
	_, response, err = service.BeginUserMove(ctx, pd.UserMoveVersionRequest{UserMove: move, MajorVersion: "6"})
	denied(t, "begin_user_move", 403, response, err)
	response, err = service.CompleteUserMove(ctx, pd.UserMoveVersionRequest{UserMove: move, MajorVersion: "6"})
	denied(t, "complete_user_move", 403, response, err)
	response, err = service.RollbackUserMove(ctx, move)
	denied(t, "rollback_user_move", 403, response, err)
	response, err = service.RehomeUserMove(ctx, pd.RehomeUserMoveRequest{UserMove: move})
	denied(t, "rehome_user_move", 403, response, err)
	response, err = service.SetMovedResourceData(ctx, pd.MovedResourceDataRequest{UserMove: move, ResourceData: "msteams-sdk-probe", MajorVersion: "6"})
	denied(t, "moved_resource_data", 403, response, err)

	response, err = service.InvokeNgtProvHandler(ctx, pd.NgtProvHandlerRequest{TenantID: lab.tenant, Region: pd.DeploymentEastUS, HandlerName: "msteams-sdk-invalid", Target: probeID, Payload: empty})
	denied(t, "ngtprov_handler", 403, response, err)
	response, err = service.SyncDirectoryObject(ctx, pd.DirectoryObjectSyncRequest{TenantID: lab.tenant, Deployment: pd.DeploymentEastUS,
		ServiceInstance: "MicrosoftCommunicationsOnline/msteams-sdk-probe", ObjectClass: pd.ObjectClassUser, ObjectID: probeID, IsValidationRequest: true})
	denied(t, "directory_object_sync", 403, response, err)
}
