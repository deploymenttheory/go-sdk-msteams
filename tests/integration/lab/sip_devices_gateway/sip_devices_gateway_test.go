package sipdevicesgatewayacc

import (
	"context"
	"strings"
	"testing"

	sdg "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/sip_devices_gateway"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const (
	fixtureService = "sip_devices_gateway"
	// A reserved fixture-form ID that cannot name a real batch or device.
	probeID = "00000000-0000-4000-8000-000000000001"
	// The fixture tenant ID; the real tenant never reaches the gateway body.
	probeTenant = "00000000-0000-4000-8000-0000000000aa"
	// An example.invalid account, a documentation MAC address and a region
	// that does not exist, so no request can act on a real device.
	probeUser     = "msteams-sdk-probe@example.invalid"
	probeHardware = "00-00-5E-00-53-01"
	probeRegion   = "msteams-sdk-invalid"
)

func tenantReplacements(t *testing.T, binding string) map[string]string {
	t.Helper()
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok || !labkit.DirectoryID.MatchString(tenant) {
		t.Fatal("tenant binding required")
	}
	return map[string]string{tenant: probeTenant, strings.ToLower(tenant): probeTenant}
}

// TestAccSIPDevicesGatewayDenied reads the bulk sign-in summary and the
// status of a nonexistent batch. The lab's app-only identity is denied every
// gateway operation, so this proves routing and the denial contract only.
func TestAccSIPDevicesGatewayDenied(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	replacements := tenantReplacements(t, binding)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	gateway := sdk.TeamsAdminAPI.SIPDevicesGateway

	_, response, err := gateway.ListBulkSignInRequests(ctx)
	if !labkit.IsHTTP(err, 403) {
		t.Fatal("bulk sign-in summary was not denied")
	}
	labkit.WriteResponseFixture(t, fixtureService, "summary", response, replacements)
	if _, response, err = gateway.GetBulkSignInRequestStatus(ctx, probeID); !labkit.IsHTTP(err, 403) {
		t.Fatal("bulk sign-in status was not denied")
	}
	labkit.WriteResponseFixture(t, fixtureService, "status", response, replacements)
	t.Log("SIP devices gateway reads denied: passed")
}

// TestAccSIPDevicesGatewayRejectedWrites sends bulk sign-in, tagging and
// transfer with inert values. The gateway denies the app-only identity, and
// the values could not name a real account, device or region anyway.
func TestAccSIPDevicesGatewayRejectedWrites(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	replacements := tenantReplacements(t, binding)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	gateway := sdk.TeamsAdminAPI.SIPDevicesGateway
	user, hardware, tenant := probeUser, probeHardware, probeTenant

	_, response, err := gateway.CreateBulkSignInRequest(ctx, probeRegion, []sdg.BulkSignInItem{{UserName: &user, HardwareID: &hardware}})
	if !labkit.IsHTTP(err, 403) {
		t.Fatal("bulk sign-in request was not denied")
	}
	labkit.WriteResponseFixture(t, fixtureService, "bulk_sign_in", response, replacements)
	if _, response, err = gateway.CreateDeviceTaggingRequest(ctx, probeRegion, sdg.DeviceTaggingRequest{HardwareID: &hardware, TenantID: &tenant}); !labkit.IsHTTP(err, 403) {
		t.Fatal("device tagging request was not denied")
	}
	labkit.WriteResponseFixture(t, fixtureService, "tagging", response, replacements)
	if _, response, err = gateway.CreateDeviceTransferRequest(ctx, sdg.DeviceTransferRequest{SdhDeviceID: probeID, SourceRegion: probeRegion, TargetRegion: probeRegion}); !labkit.IsHTTP(err, 403) {
		t.Fatal("device transfer request was not denied")
	}
	labkit.WriteResponseFixture(t, fixtureService, "transfer", response, replacements)
	t.Log("SIP devices gateway writes denied: passed")
}
