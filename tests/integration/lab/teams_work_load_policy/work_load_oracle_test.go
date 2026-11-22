package teamsworkloadpolicyacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_work_load_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

var workLoadPolicyScalars = []string{"AllowCalling", "AllowCallingPinned", "AllowMeeting", "AllowMeetingPinned", "AllowMessaging", "AllowMessagingPinned", "Description"}

type workLoadPolicyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

var workLoadPolicyOwnedIdentity = regexp.MustCompile(`(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-work-load-policy$`)

func decodeWorkLoadPolicyOracleForIdentity(raw []byte, tenant, identity string) (workLoadPolicyObservation, error) {
	var v workLoadPolicyObservation
	if len(raw) > labkit.MaxOracleBytes || !labkit.DirectoryID.MatchString(tenant) || !strings.EqualFold(identity, "Global") && !workLoadPolicyOwnedIdentity.MatchString(identity) {
		return v, labkit.ErrOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil {
		return workLoadPolicyObservation{}, labkit.ErrOracle
	}
	identityMatches := strings.EqualFold(v.Identity, "Global")
	if !strings.EqualFold(identity, "Global") {
		identityMatches = labkit.SamePolicyName(v.Identity, identity)
	}
	if !v.Success || !strings.EqualFold(v.Tenant, tenant) || !identityMatches || len(v.Settings) != len(workLoadPolicyScalars) {
		return workLoadPolicyObservation{}, labkit.ErrOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return workLoadPolicyObservation{}, labkit.ErrOracle
	}
	for _, name := range workLoadPolicyScalars {
		if _, ok := v.Settings[name]; !ok {
			return workLoadPolicyObservation{}, labkit.ErrOracle
		}
	}
	settings, err := json.Marshal(v.Settings)
	var model resource.Policy
	if err != nil || json.Unmarshal(settings, &model) != nil {
		return workLoadPolicyObservation{}, labkit.ErrOracle
	}
	return v, nil
}

func workLoadPolicyOracleForIdentity(t *testing.T, ctx context.Context, tenant, identity string) workLoadPolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": identity})
	if err != nil {
		t.Fatal("invalid work-load policy oracle selector")
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-work-load-policy.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"), "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned work-load policy read failed; no output retained")
	}
	if stage := labkit.ReadOracleFailureStage(output.Data); stage != "" {
		t.Log("oracle failure stage: " + stage)
	}
	value, err := decodeWorkLoadPolicyOracleForIdentity(output.Data, tenant, identity)
	if err != nil {
		t.Fatal("pinned work-load policy projection unavailable")
	}
	return value
}

func matchesWorkLoadPolicyGetterProjection(v *resource.Policy, settings map[string]json.RawMessage) bool {
	if v == nil {
		return false
	}
	return labkit.SameProjection(workLoadProjection(v), settings)
}

func workLoadProjection(v *resource.Policy) map[string]any {
	return map[string]any{"AllowCalling": v.AllowCalling, "AllowCallingPinned": v.AllowCallingPinned, "AllowMeeting": v.AllowMeeting, "AllowMeetingPinned": v.AllowMeetingPinned, "AllowMessaging": v.AllowMessaging, "AllowMessagingPinned": v.AllowMessagingPinned, "Description": v.Description}
}
