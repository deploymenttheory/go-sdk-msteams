package teamsupdatemanagementpolicyacc

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

	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_update_management_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

var updateManagementPolicyScalars = []string{"Description", "AllowManagedUpdates", "AllowPreview", "UpdateDayOfWeek", "UpdateTime", "UpdateTimeOfDay", "AllowPublicPreview", "UseNewTeamsClient", "BlockLegacyAuthorization", "DisabledInProductMessages"}

type updateManagementPolicyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

var updateManagementPolicyOwnedIdentity = regexp.MustCompile(`(?i)^(?:tag:)?msteams-sdk-[a-z2-7]{26,64}-update-management-policy$`)

func decodeUpdateManagementPolicyOracleForIdentity(raw []byte, tenant, identity string) (updateManagementPolicyObservation, error) {
	var v updateManagementPolicyObservation
	if len(raw) > labkit.MaxOracleBytes || !labkit.DirectoryID.MatchString(tenant) || !strings.EqualFold(identity, "Global") && !updateManagementPolicyOwnedIdentity.MatchString(identity) {
		return v, labkit.ErrOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil {
		return updateManagementPolicyObservation{}, labkit.ErrOracle
	}
	identityMatches := strings.EqualFold(v.Identity, "Global")
	if !strings.EqualFold(identity, "Global") {
		identityMatches = labkit.SamePolicyName(v.Identity, identity)
	}
	if !v.Success || !strings.EqualFold(v.Tenant, tenant) || !identityMatches || len(v.Settings) != len(updateManagementPolicyScalars) {
		return updateManagementPolicyObservation{}, labkit.ErrOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return updateManagementPolicyObservation{}, labkit.ErrOracle
	}
	for _, name := range updateManagementPolicyScalars {
		if _, ok := v.Settings[name]; !ok {
			return updateManagementPolicyObservation{}, labkit.ErrOracle
		}
	}
	settings, err := json.Marshal(v.Settings)
	var model resource.Policy
	if err != nil || json.Unmarshal(settings, &model) != nil {
		return updateManagementPolicyObservation{}, labkit.ErrOracle
	}
	return v, nil
}

func updateManagementPolicyOracleForIdentity(t *testing.T, ctx context.Context, tenant, identity string) updateManagementPolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant, "identity": identity})
	if err != nil {
		t.Fatal("invalid update-management policy oracle selector")
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "read-update-management-policy.ps1")
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
		t.Fatal("pinned update-management policy read failed; no output retained")
	}
	if stage := labkit.ReadOracleFailureStage(output.Data); stage != "" {
		t.Log("oracle failure stage: " + stage)
	}
	value, err := decodeUpdateManagementPolicyOracleForIdentity(output.Data, tenant, identity)
	if err != nil {
		t.Fatal("pinned update-management policy projection unavailable")
	}
	return value
}
