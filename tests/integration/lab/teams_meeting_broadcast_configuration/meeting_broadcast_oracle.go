package teamsmeetingbroadcastconfigurationacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	broadcast "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_broadcast_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// meetingBroadcastProjection is the default public getter projection.
func meetingBroadcastProjection(v *broadcast.Configuration) map[string]any {
	return map[string]any{"SupportURL": v.SupportURL, "AllowSdnProviderForBroadcastMeeting": v.AllowSdnProviderForBroadcastMeeting}
}

// meetingBroadcastSDNProjection is the -ExposeSDNConfigurationJsonBlob getter
// projection. The provider token is reduced to presence and never compared.
func meetingBroadcastSDNProjection(v *broadcast.Configuration) map[string]any {
	return map[string]any{
		"SupportURL":                          v.SupportURL,
		"AllowSdnProviderForBroadcastMeeting": v.AllowSdnProviderForBroadcastMeeting,
		"SdnName":                             v.SdnName,
		"SdnLicenseId":                        v.SdnLicenseID,
		"SdnAzureSubscriptionId":              v.SdnAzureSubscriptionID,
		"SdnApiTemplateUrl":                   v.SdnAPITemplateURL,
		"SdnRuntimeconfiguration":             v.SdnRuntimeConfiguration,
		"SdnAttendeeFallbackCount":            v.SdnAttendeeFallbackCount,
		"SdnApiTokenPresent":                  v.SdnAPIToken != nil && *v.SdnAPIToken != "",
	}
}

type meetingBroadcastOracleResult struct {
	Settings map[string]json.RawMessage
	SDN      map[string]json.RawMessage
}

// meetingBroadcastOracle runs the pinned 8.1.0 Global getter, with and without
// -ExposeSDNConfigurationJsonBlob, in a dedicated read-only PowerShell process.
func meetingBroadcastOracle(t *testing.T, ctx context.Context, tenant string) meetingBroadcastOracleResult {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": "Global", "kind": "meeting-broadcast"})
	if err != nil {
		t.Fatal(err)
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-meeting-broadcast-configuration.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	for _, key := range []string{"MSTEAMS_LAB_AUTH_METHOD", "MSTEAMS_LAB_MODULE_PATH", "MSTEAMS_LAB_IDENTITY_FILE"} {
		cmd.Env = append(cmd.Env, key+"="+os.Getenv(key))
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1")
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = bytes.NewReader(input), &output, io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned meeting broadcast configuration read process failed")
	}
	var result struct {
		Success  bool                       `json:"success"`
		Tenant   string                     `json:"tenant_id"`
		Identity string                     `json:"identity"`
		Settings map[string]json.RawMessage `json:"settings"`
		SDN      map[string]json.RawMessage `json:"sdn"`
	}
	decoder := json.NewDecoder(bytes.NewReader(output.Data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&result) != nil || decoder.Decode(new(any)) != io.EOF || !result.Success || !strings.EqualFold(result.Tenant, tenant) || !strings.EqualFold(result.Identity, "Global") {
		t.Fatal("pinned meeting broadcast configuration projection unavailable")
	}
	for _, name := range []string{"SupportURL", "AllowSdnProviderForBroadcastMeeting"} {
		if _, ok := result.Settings[name]; !ok {
			t.Fatal("incomplete meeting broadcast configuration getter projection")
		}
	}
	if len(result.Settings) != 2 || len(result.SDN) != 9 {
		t.Fatal("incomplete meeting broadcast configuration getter projection")
	}
	if _, ok := result.SDN["SdnApiToken"]; ok {
		t.Fatal("meeting broadcast oracle emitted a provider token")
	}
	return meetingBroadcastOracleResult{Settings: result.Settings, SDN: result.SDN}
}
