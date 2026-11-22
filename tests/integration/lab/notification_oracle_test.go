package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"os/exec"
	"regexp"
	"strings"
	"testing"
	"time"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_notification_and_feeds_policy"
)

var notificationPolicyScalars = []string{"Description", "SuggestedFeedsEnabledType", "TrendingFeedsEnabledType"}

type notificationPolicyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

var notificationPolicyOwnedIdentity = regexp.MustCompile(`(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-notification-policy$`)

func decodeNotificationAndFeedsPolicyOracleForIdentity(raw []byte, tenant, identity string) (notificationPolicyObservation, error) {
	var v notificationPolicyObservation
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) || !strings.EqualFold(identity, "Global") && !notificationPolicyOwnedIdentity.MatchString(identity) {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil {
		return notificationPolicyObservation{}, errTeamOracle
	}
	identityMatches := strings.EqualFold(v.Identity, "Global")
	if !strings.EqualFold(identity, "Global") {
		identityMatches = samePolicyName(v.Identity, identity)
	}
	if !v.Success || !strings.EqualFold(v.Tenant, tenant) || !identityMatches || len(v.Settings) != len(notificationPolicyScalars) {
		return notificationPolicyObservation{}, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return notificationPolicyObservation{}, errTeamOracle
	}
	for _, name := range notificationPolicyScalars {
		if _, ok := v.Settings[name]; !ok {
			return notificationPolicyObservation{}, errTeamOracle
		}
	}
	settings, err := json.Marshal(v.Settings)
	var model resource.Policy
	if err != nil || json.Unmarshal(settings, &model) != nil {
		return notificationPolicyObservation{}, errTeamOracle
	}
	return v, nil
}

func notificationPolicyOracleForIdentity(t *testing.T, ctx context.Context, tenant, identity string) notificationPolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": identity})
	if err != nil {
		t.Fatal("invalid notification policy oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-notification-policy.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"), "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output teamOracleBuffer
	defer func() { clear(output.data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.exceeded {
		t.Fatal("pinned notification policy read failed; no output retained")
	}
	if stage := readOracleFailureStage(output.data); stage != "" {
		t.Log("oracle failure stage: " + stage)
	}
	value, err := decodeNotificationAndFeedsPolicyOracleForIdentity(output.data, tenant, identity)
	if err != nil {
		t.Fatal("pinned notification policy projection unavailable")
	}
	return value
}

func matchesNotificationAndFeedsPolicyGetterProjection(v *resource.Policy, settings map[string]json.RawMessage) bool {
	if v == nil {
		return false
	}
	return sameTemplateProjection(map[string]any{"Description": v.Description, "SuggestedFeedsEnabledType": v.SuggestedFeedsEnabledType, "TrendingFeedsEnabledType": v.TrendingFeedsEnabledType}, settings)
}
