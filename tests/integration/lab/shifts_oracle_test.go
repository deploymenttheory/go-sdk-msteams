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

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_shifts_policy"
)

var shiftsPolicyScalars = []string{"ShiftNoticeFrequency", "ShiftNoticeMessageType", "ShiftNoticeMessageCustom", "AccessType", "AccessGracePeriodMinutes", "EnableScheduleOwnerPermissions"}

type shiftsPolicyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

var shiftsPolicyOwnedIdentity = regexp.MustCompile(`(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-shifts-policy$`)

func decodeShiftsPolicyOracleForIdentity(raw []byte, tenant, identity string) (shiftsPolicyObservation, error) {
	var v shiftsPolicyObservation
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) || !strings.EqualFold(identity, "Global") && !shiftsPolicyOwnedIdentity.MatchString(identity) {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil {
		return shiftsPolicyObservation{}, errTeamOracle
	}
	identityMatches := strings.EqualFold(v.Identity, "Global")
	if !strings.EqualFold(identity, "Global") {
		identityMatches = samePolicyName(v.Identity, identity)
	}
	if !v.Success || !strings.EqualFold(v.Tenant, tenant) || !identityMatches || len(v.Settings) != len(shiftsPolicyScalars) {
		return shiftsPolicyObservation{}, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return shiftsPolicyObservation{}, errTeamOracle
	}
	for _, name := range shiftsPolicyScalars {
		if _, ok := v.Settings[name]; !ok {
			return shiftsPolicyObservation{}, errTeamOracle
		}
	}
	settings, err := json.Marshal(v.Settings)
	var model resource.Policy
	if err != nil || json.Unmarshal(settings, &model) != nil {
		return shiftsPolicyObservation{}, errTeamOracle
	}
	return v, nil
}

func shiftsPolicyOracleForIdentity(t *testing.T, ctx context.Context, tenant, identity string) shiftsPolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": identity})
	if err != nil {
		t.Fatal("invalid Shifts policy oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-shifts-policy.ps1")
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
		t.Fatal("pinned Shifts policy read failed; no output retained")
	}
	if stage := readOracleFailureStage(output.data); stage != "" {
		t.Log("oracle failure stage: " + stage)
	}
	value, err := decodeShiftsPolicyOracleForIdentity(output.data, tenant, identity)
	if err != nil {
		t.Fatal("pinned Shifts policy projection unavailable")
	}
	return value
}

func matchesShiftsPolicyGetterProjection(v *resource.Policy, settings map[string]json.RawMessage) bool {
	if v == nil {
		return false
	}
	return sameTemplateProjection(map[string]any{"ShiftNoticeFrequency": v.ShiftNoticeFrequency, "ShiftNoticeMessageType": v.ShiftNoticeMessageType, "ShiftNoticeMessageCustom": v.ShiftNoticeMessageCustom, "AccessType": v.AccessType, "AccessGracePeriodMinutes": v.AccessGracePeriodMinutes, "EnableScheduleOwnerPermissions": v.EnableScheduleOwnerPermissions}, settings)
}
