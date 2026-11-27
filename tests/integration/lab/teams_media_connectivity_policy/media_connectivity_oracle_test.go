package teamsmediaconnectivitypolicyacc

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

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_media_connectivity_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

var mediaConnectivityPolicyScalars = []string{"DirectConnection"}

type mediaConnectivityPolicyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

var mediaConnectivityPolicyOwnedIdentity = regexp.MustCompile(`(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-media-connectivity-policy$`)

func decodeMediaConnectivityPolicyOracleForIdentity(raw []byte, tenant, identity string) (mediaConnectivityPolicyObservation, error) {
	var v mediaConnectivityPolicyObservation
	if len(raw) > labkit.MaxOracleBytes || !labkit.DirectoryID.MatchString(tenant) || !strings.EqualFold(identity, "Global") && !mediaConnectivityPolicyOwnedIdentity.MatchString(identity) {
		return v, labkit.ErrOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil {
		return mediaConnectivityPolicyObservation{}, labkit.ErrOracle
	}
	identityMatches := strings.EqualFold(v.Identity, "Global")
	if !strings.EqualFold(identity, "Global") {
		identityMatches = labkit.SamePolicyName(v.Identity, identity)
	}
	if !v.Success || !strings.EqualFold(v.Tenant, tenant) || !identityMatches || len(v.Settings) != len(mediaConnectivityPolicyScalars) {
		return mediaConnectivityPolicyObservation{}, labkit.ErrOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return mediaConnectivityPolicyObservation{}, labkit.ErrOracle
	}
	for _, name := range mediaConnectivityPolicyScalars {
		if _, ok := v.Settings[name]; !ok {
			return mediaConnectivityPolicyObservation{}, labkit.ErrOracle
		}
	}
	settings, err := json.Marshal(v.Settings)
	var model resource.Policy
	if err != nil || json.Unmarshal(settings, &model) != nil {
		return mediaConnectivityPolicyObservation{}, labkit.ErrOracle
	}
	return v, nil
}

// mediaConnectivityPolicyOracleForIdentity runs the dedicated pinned
// PowerShell read for Global or one owned msteams-sdk-* policy.
func mediaConnectivityPolicyOracleForIdentity(t *testing.T, ctx context.Context, tenant, identity string) mediaConnectivityPolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": identity})
	if err != nil {
		t.Fatal("invalid media connectivity policy oracle selector")
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-media-connectivity-policy.ps1")
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
		t.Fatal("pinned media connectivity policy read failed; no output retained")
	}
	if stage := labkit.ReadOracleFailureStage(output.Data); stage != "" {
		t.Log("oracle failure stage: " + stage)
	}
	value, err := decodeMediaConnectivityPolicyOracleForIdentity(output.Data, tenant, identity)
	if err != nil {
		t.Fatal("pinned media connectivity policy projection unavailable")
	}
	return value
}

func matchesMediaConnectivityPolicyGetterProjection(v *resource.Policy, settings map[string]json.RawMessage) bool {
	if v == nil {
		return false
	}
	return labkit.SameProjection(map[string]any{"DirectConnection": v.DirectConnection}, settings)
}
