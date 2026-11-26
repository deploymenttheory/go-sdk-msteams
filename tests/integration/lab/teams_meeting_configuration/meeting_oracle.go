package teamsmeetingconfigurationacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	meeting "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// meetingSettings are the 21 public getters of the pinned compiled type.
var meetingSettings = []string{
	"PublishedEntraAuthenticationContexts", "LogoURL", "LegalURL", "HelpURL", "CustomFooterText",
	"DisableAnonymousJoin", "DisableAppInteractionForAnonymousUsers", "EnableQoS",
	"EnableAttributedTranscripts", "EnableGraphTranscriptAccess", "EnableWorkIQTranscriptAccess",
	"ClientAudioPort", "ClientAudioPortRange", "ClientVideoPort", "ClientVideoPortRange",
	"ClientAppSharingPort", "ClientAppSharingPortRange", "ClientMediaPortRangeEnabled",
	"LimitPresenterRolePermissions", "FeedbackSurveyForAnonymousUsers", "ReportMeeting",
}

// contextProjection keeps the seven public element getters, matching the
// oracle's per-element projection.
func contextProjection(items []*meeting.PublishedEntraAuthenticationContext) []map[string]any {
	values := make([]map[string]any, 0, len(items))
	for _, item := range items {
		if item == nil {
			values = append(values, nil)
			continue
		}
		values = append(values, map[string]any{
			"Identity": item.Identity, "Id": item.ID, "PublishedName": item.PublishedName,
			"PublishedDescription": item.PublishedDescription, "EntraName": item.EntraName,
			"EntraDescription": item.EntraDescription, "Published": item.Published,
		})
	}
	return values
}

func meetingProjection(v *meeting.Configuration) map[string]any {
	return map[string]any{
		"PublishedEntraAuthenticationContexts":   contextProjection(v.PublishedEntraAuthenticationContexts),
		"LogoURL":                                v.LogoURL,
		"LegalURL":                               v.LegalURL,
		"HelpURL":                                v.HelpURL,
		"CustomFooterText":                       v.CustomFooterText,
		"DisableAnonymousJoin":                   v.DisableAnonymousJoin,
		"DisableAppInteractionForAnonymousUsers": v.DisableAppInteractionForAnonymousUsers,
		"EnableQoS":                              v.EnableQoS,
		"EnableAttributedTranscripts":            v.EnableAttributedTranscripts,
		"EnableGraphTranscriptAccess":            v.EnableGraphTranscriptAccess,
		"EnableWorkIQTranscriptAccess":           v.EnableWorkIQTranscriptAccess,
		"ClientAudioPort":                        v.ClientAudioPort,
		"ClientAudioPortRange":                   v.ClientAudioPortRange,
		"ClientVideoPort":                        v.ClientVideoPort,
		"ClientVideoPortRange":                   v.ClientVideoPortRange,
		"ClientAppSharingPort":                   v.ClientAppSharingPort,
		"ClientAppSharingPortRange":              v.ClientAppSharingPortRange,
		"ClientMediaPortRangeEnabled":            v.ClientMediaPortRangeEnabled,
		"LimitPresenterRolePermissions":          v.LimitPresenterRolePermissions,
		"FeedbackSurveyForAnonymousUsers":        v.FeedbackSurveyForAnonymousUsers,
		"ReportMeeting":                          v.ReportMeeting,
	}
}

// meetingOracle runs the pinned 8.1.0 Global getter in a dedicated read-only
// PowerShell process and returns its 21 public settings.
func meetingOracle(t *testing.T, ctx context.Context, tenant string) map[string]json.RawMessage {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": "Global", "kind": "meeting"})
	if err != nil {
		t.Fatal(err)
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-meeting-configuration.ps1")
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
		t.Fatal("pinned meeting configuration read process failed")
	}
	settings, err := decodeMeetingOracle(output.Data, tenant)
	if err != nil {
		t.Fatal("pinned meeting configuration projection unavailable")
	}
	return settings
}

type oracleError string

func (e oracleError) Error() string { return string(e) }

// decodeMeetingOracle accepts exactly one successful, tenant-bound Global
// report carrying all 21 getters.
func decodeMeetingOracle(raw []byte, tenant string) (map[string]json.RawMessage, error) {
	var result struct {
		Success  bool                       `json:"success"`
		Tenant   string                     `json:"tenant_id"`
		Identity string                     `json:"identity"`
		Settings map[string]json.RawMessage `json:"settings"`
	}
	decoder := json.NewDecoder(bytes.NewReader(raw))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&result) != nil || decoder.Decode(new(any)) != io.EOF || !result.Success || !strings.EqualFold(result.Tenant, tenant) || !strings.EqualFold(result.Identity, "Global") {
		return nil, oracleError("unverified oracle report")
	}
	if len(result.Settings) != len(meetingSettings) {
		return nil, oracleError("incomplete oracle report")
	}
	for _, name := range meetingSettings {
		if _, ok := result.Settings[name]; !ok {
			return nil, oracleError("incomplete oracle report")
		}
	}
	return result.Settings, nil
}
