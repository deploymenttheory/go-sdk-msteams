package teamsremotelogcollectiondeviceacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	device "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_remote_log_collection_device"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const (
	deviceFixtures = "teams_remote_log_collection_device"
	// Reserved fixture-form IDs that cannot name a real device or user,
	// so every probe below targets something that does not exist.
	probeDevice = "00000000-0000-4000-8000-000000000001"
	probeUser   = "00000000-0000-4000-8000-000000000002"
	// The service's not-found error code for a missing device.
	missingCode = "40407"
)

func tenantReplacements(t *testing.T, binding string) map[string]string {
	t.Helper()
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok || !labkit.DirectoryID.MatchString(tenant) {
		t.Fatal("tenant binding required")
	}
	return map[string]string{tenant: "00000000-0000-4000-8000-0000000000aa", strings.ToLower(tenant): "00000000-0000-4000-8000-0000000000aa"}
}

type remoteLogObservation struct {
	Success                  bool   `json:"success"`
	ConfigurationCount       int    `json:"configuration_count"`
	ConfigurationDeviceCount int    `json:"configuration_device_count"`
	DeviceCount              int    `json:"device_count"`
	MissingErrorCode         string `json:"missing_error_code"`
}

func powerShellRemoteLogCollection(t *testing.T, ctx context.Context) remoteLogObservation {
	t.Helper()
	command := labkit.OracleCommand(ctx, "read-remote-log-collection.ps1")
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
	command.Env = append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_REMOTE_LOG_DEVICE_PROBE="+probeDevice)
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	command.Stdout, command.Stderr = &output, io.Discard
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned PowerShell remote log collection read failed; no output retained")
	}
	var observed remoteLogObservation
	decoder := json.NewDecoder(bytes.NewReader(output.Data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&observed) != nil || !observed.Success {
		t.Fatal("pinned PowerShell remote log collection projection unavailable")
	}
	return observed
}

// TestAccRemoteLogCollectionDeviceReads lists devices and reads one
// nonexistent device, then compares them and the Global configuration with
// pinned PowerShell. It changes nothing.
func TestAccRemoteLogCollectionDeviceReads(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	replacements := tenantReplacements(t, binding)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()

	devices := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionDevices
	all, response, err := devices.List(ctx)
	if err != nil {
		t.Fatal("remote log collection device list failed")
	}
	labkit.WriteResponseFixture(t, deviceFixtures, "list", response, replacements)
	if _, response, err = devices.GetByID(ctx, probeDevice); !labkit.IsHTTP(err, 404) {
		t.Fatal("missing device was not a 404")
	}
	labkit.WriteResponseFixture(t, deviceFixtures, "get_missing", response, replacements)
	t.Log("remote log collection device reads: passed")

	if labkit.OracleEnabled(t) {
		listed, _, err := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionConfigurations.List(ctx)
		if err != nil || len(listed) == 0 {
			t.Fatal("remote log collection configuration list failed")
		}
		global := listed[0]
		observed := powerShellRemoteLogCollection(t, ctx)
		if observed.ConfigurationCount != len(listed) || observed.ConfigurationDeviceCount != len(global.Devices) ||
			observed.DeviceCount != len(all) || observed.MissingErrorCode != missingCode {
			t.Fatal("PowerShell and Go remote log collection observations differ")
		}
		t.Log("PowerShell 8.1.0 / Go remote log collection parity: passed")
	}
}

// TestAccRemoteLogCollectionDeviceRejectedWrites sends each device write
// where the service rejects it or its target does not exist, so nothing is
// created or changed: a device with an invalid DeviceId is refused, and
// Set/Remove target a reserved, nonexistent device.
func TestAccRemoteLogCollectionDeviceRejectedWrites(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	replacements := tenantReplacements(t, binding)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()

	devices := sdk.TeamsAdminAPI.TeamsRemoteLogCollectionDevices
	expiry := time.Now().UTC().Add(72 * time.Hour).Format(time.RFC3339)
	_, response, err := devices.CreateByID(ctx, probeDevice, device.CreateRequest{DeviceID: "msteams-sdk-not-a-guid", UserID: probeUser, ExpireAfter: expiry})
	if !labkit.IsHTTP(err, 400) {
		t.Fatal("device with an invalid DeviceId was not rejected")
	}
	labkit.WriteResponseFixture(t, deviceFixtures, "create_invalid_device", response, replacements)
	update := device.UpdateRequest{Settings: device.Settings{UserID: policy.Set(probeUser)}}
	if result, response, err := devices.UpdateByID(ctx, probeDevice, update); !labkit.IsHTTP(err, 404) || result == nil || result.PatchInvoked {
		t.Fatal("update of a missing device was not stopped by its read")
	} else {
		labkit.WriteResponseFixture(t, deviceFixtures, "update_missing", response, replacements)
	}
	if _, response, err = devices.PatchByID(ctx, probeDevice, device.PatchRequest{Settings: update.Settings, BodyID: policy.Set(probeDevice)}); !labkit.IsHTTP(err, 404) {
		t.Fatal("patch of a missing device was not a 404")
	}
	labkit.WriteResponseFixture(t, deviceFixtures, "patch_missing", response, replacements)
	if _, response, err = devices.DeleteByID(ctx, probeDevice, policy.Controls{}); !labkit.IsHTTP(err, 404) {
		t.Fatal("delete of a missing device was not a 404")
	}
	labkit.WriteResponseFixture(t, deviceFixtures, "delete_missing", response, replacements)
	t.Log("remote log collection device rejected writes: passed")
}
