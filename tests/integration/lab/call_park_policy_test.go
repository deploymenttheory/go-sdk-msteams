package lab_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"reflect"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_call_park_policy"
)

var callParkPolicyFields = []string{"Description", "AllowCallPark", "ParkTimeoutSeconds", "PickupRangeStart", "PickupRangeEnd"}

type callParkPolicyObservation struct {
	Success  bool                       `json:"success"`
	Tenant   string                     `json:"tenant_id"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

func decodeCallParkPolicyOracle(raw []byte, tenant string) (callParkPolicyObservation, error) {
	var v callParkPolicyObservation
	if len(raw) > maxTeamOracleBytes || !labDirectoryID.MatchString(tenant) {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || !strings.EqualFold(v.Tenant, tenant) || !strings.EqualFold(v.Identity, "Global") || len(v.Settings) != len(callParkPolicyFields) {
		return callParkPolicyObservation{}, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return callParkPolicyObservation{}, errTeamOracle
	}
	var text *string
	if raw, ok := v.Settings["Description"]; !ok || json.Unmarshal(raw, &text) != nil {
		return callParkPolicyObservation{}, errTeamOracle
	}
	var boolean *bool
	if raw, ok := v.Settings["AllowCallPark"]; !ok || json.Unmarshal(raw, &boolean) != nil {
		return callParkPolicyObservation{}, errTeamOracle
	}
	for _, key := range []string{"ParkTimeoutSeconds", "PickupRangeStart", "PickupRangeEnd"} {
		var number *int64
		if raw, ok := v.Settings[key]; !ok || json.Unmarshal(raw, &number) != nil {
			return callParkPolicyObservation{}, errTeamOracle
		}
	}
	return v, nil
}

func callParkPolicyOracle(t *testing.T, ctx context.Context, tenant string) callParkPolicyObservation {
	t.Helper()
	input, err := json.Marshal(map[string]string{"tenant_id": tenant})
	if err != nil {
		t.Fatal("invalid call park policy oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-call-park-policy.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output teamOracleBuffer
	defer func() { clear(output.data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.exceeded {
		t.Fatal("pinned call park policy read failed; no output retained")
	}
	value, err := decodeCallParkPolicyOracle(output.data, tenant)
	if err != nil {
		t.Fatal("pinned call park policy projection unavailable")
	}
	return value
}

// TestAccCallParkPolicyReadOnly never edits Global, creates resources or grants a
// policy. Tenant writes and effective assignments need separate acceptance.
func TestAccCallParkPolicyReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("call park policy reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid lab enablement")
	}
	oracle := teamOracleEnabled(t)
	sdk, binding, _ := labSDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	value, response, err := sdk.TeamsAdminAPI.TeamsCallParkPolicies.GetByIdentity(ctx, "Global")
	if err != nil {
		status := 0
		if response != nil {
			status = response.StatusCode()
		}
		t.Fatalf("call park policy Global read failed (HTTP %d); no response payload retained", status)
	}
	listed, _, err := sdk.TeamsAdminAPI.TeamsCallParkPolicies.ListFiltered(ctx, "Global")
	selected, selectionErr := globalCallParkPolicy(listed)
	if err != nil || selectionErr != nil || value == nil || !strings.EqualFold(value.Identity, "Global") {
		t.Fatal("call park policy Global selection differs")
	}
	for _, key := range callParkPolicyFields {
		var a, b any
		left, right := json.NewDecoder(bytes.NewReader(value.RawFields[key])), json.NewDecoder(bytes.NewReader(selected.RawFields[key]))
		left.UseNumber()
		right.UseNumber()
		if left.Decode(&a) != nil || right.Decode(&b) != nil || !reflect.DeepEqual(a, b) {
			t.Fatal("call park policy Global read/list projections differ")
		}
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok {
			t.Fatal("canonical tenant binding required")
		}
		actual := callParkPolicyOracle(t, ctx, tenant)
		for _, key := range callParkPolicyFields {
			want := value.RawFields[key]
			var left, right any
			if decodeCallParkValue(want, &left) != nil || decodeCallParkValue(actual.Settings[key], &right) != nil || !reflect.DeepEqual(left, right) {
				t.Fatal("Go and pinned call park policy known projections differ")
			}
		}
	}
	t.Log("selected call park policy reads verified; persistence, feature effects and assignments remain separately qualified")
}

func decodeCallParkValue(raw []byte, out *any) error {
	d := json.NewDecoder(bytes.NewReader(raw))
	d.UseNumber()
	return d.Decode(out)
}
func globalCallParkPolicy(values []resource.Policy) (*resource.Policy, error) {
	var found *resource.Policy
	for i := range values {
		if strings.EqualFold(values[i].Identity, "Global") {
			if found != nil {
				return nil, errTeamOracle
			}
			found = &values[i]
		}
	}
	if found == nil {
		return nil, errTeamOracle
	}
	return found, nil
}
func TestUnitCallParkPolicyOracleRejectsIncompleteWrongTenantMixedOutputAndInt64Overflow(t *testing.T) {
	t.Parallel()
	const tenant = "00000000-0000-4000-8000-000000000003"
	settings := `{"Description":"","AllowCallPark":false,"ParkTimeoutSeconds":-9223372036854775808,"PickupRangeStart":9007199254740993,"PickupRangeEnd":9223372036854775807}`
	valid := `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":` + settings + `}`
	for _, raw := range []string{`{"success":false}`, strings.Replace(valid, tenant, "00000000-0000-4000-8000-000000000004", 1), strings.Replace(valid, `"identity":"Global"`, `"identity":"Other"`, 1), valid + "{}", strings.Replace(valid, `"AllowCallPark":false`, `"AllowCallPark":0`, 1), strings.Replace(valid, "9223372036854775807", "9223372036854775808", 1), strings.Replace(valid, "9007199254740993", "0.5", 1), `{"success":true,"tenant_id":"` + tenant + `","identity":"Global","settings":{}}`} {
		if _, e := decodeCallParkPolicyOracle([]byte(raw), tenant); e == nil {
			t.Fatal("invalid call park oracle accepted")
		}
	}
	if _, e := decodeCallParkPolicyOracle([]byte(valid), tenant); e != nil {
		t.Fatal("complete exact Int64 projection rejected")
	}
	if _, e := globalCallParkPolicy([]resource.Policy{{Identity: "Tag:Global"}, {Identity: "Global"}}); e != nil {
		t.Fatal("tagged match hid Global")
	}
	if _, e := globalCallParkPolicy([]resource.Policy{{Identity: "Global"}, {Identity: "global"}}); e == nil {
		t.Fatal("ambiguous Global accepted")
	}
}
