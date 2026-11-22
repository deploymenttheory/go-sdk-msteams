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

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_personal_attendant_policy"
)

var personalPolicyScalars = []string{"PersonalAttendant", "CallScreening", "CalendarBookings", "InboundInternalCalls", "InboundFederatedCalls", "InboundPSTNCalls", "AutomaticTranscription", "AutomaticRecording"}

type personalPolicyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

var personalPolicyOwnedIdentity = regexp.MustCompile(`(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-personal-policy$`)

func decodePersonalAttendantPolicyOracleForIdentity(raw []byte, tenant, identity string) (personalPolicyObservation, error) {
	var v personalPolicyObservation
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) || !strings.EqualFold(identity, "Global") && !personalPolicyOwnedIdentity.MatchString(identity) {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil {
		return personalPolicyObservation{}, errTeamOracle
	}
	identityMatches := strings.EqualFold(v.Identity, "Global")
	if !strings.EqualFold(identity, "Global") {
		identityMatches = samePolicyName(v.Identity, identity)
	}
	if !v.Success || !strings.EqualFold(v.Tenant, tenant) || !identityMatches || len(v.Settings) != len(personalPolicyScalars) {
		return personalPolicyObservation{}, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return personalPolicyObservation{}, errTeamOracle
	}
	for _, name := range personalPolicyScalars {
		if _, ok := v.Settings[name]; !ok {
			return personalPolicyObservation{}, errTeamOracle
		}
	}
	settings, err := json.Marshal(v.Settings)
	var model resource.Policy
	if err != nil || json.Unmarshal(settings, &model) != nil {
		return personalPolicyObservation{}, errTeamOracle
	}
	return v, nil
}

func personalPolicyOracleForIdentity(t *testing.T, ctx context.Context, tenant, identity string) personalPolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": identity})
	if err != nil {
		t.Fatal("invalid personal attendant policy oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-personal-policy.ps1")
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
		t.Fatal("pinned personal attendant policy read failed; no output retained")
	}
	if stage := readOracleFailureStage(output.data); stage != "" {
		t.Log("oracle failure stage: " + stage)
	}
	value, err := decodePersonalAttendantPolicyOracleForIdentity(output.data, tenant, identity)
	if err != nil {
		t.Fatal("pinned personal attendant policy projection unavailable")
	}
	return value
}

func matchesPersonalAttendantPolicyGetterProjection(v *resource.Policy, settings map[string]json.RawMessage) bool {
	if v == nil {
		return false
	}
	return sameTemplateProjection(map[string]any{"PersonalAttendant": v.PersonalAttendant, "CallScreening": v.CallScreening, "CalendarBookings": v.CalendarBookings, "InboundInternalCalls": v.InboundInternalCalls, "InboundFederatedCalls": v.InboundFederatedCalls, "InboundPSTNCalls": v.InboundPSTNCalls, "AutomaticTranscription": v.AutomaticTranscription, "AutomaticRecording": v.AutomaticRecording}, settings)
}
