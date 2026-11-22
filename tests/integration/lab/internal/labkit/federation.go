package labkit

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"reflect"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/go-resty/resty/v2"
)

func federationReadOnly(t *testing.T) lab.Config {
	t.Helper()
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("federation reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid federation lab enablement")
	}
	return cfg
}

var federationProjectionFields = map[string][]string{
	"external-policy": {"Description", "EnableFederationAccess"},
	"acs":             {"EnableAcsUsers", "RequireAcsFederationForMeeting"},
	"external-users":  {"BlockExternalUserAccess"},
	"tenant":          {"AllowFederatedUsers", "AllowTeamsConsumer"},
}

var federationFixturePattern = regexp.MustCompile(`^msteams-sdk-[a-z2-7]{26,64}-external-access$`)

func safeFederationIdentity(kind, id string) bool {
	if strings.EqualFold(id, "Global") {
		return true
	}
	if kind != "external-policy" {
		return false
	}
	if len(id) > 4 && strings.EqualFold(id[:4], "tag:") {
		id = id[4:]
	}
	return federationFixturePattern.MatchString(strings.ToLower(id))
}

type FederationObservation struct {
	Success  bool                       `json:"success"`
	Kind     string                     `json:"kind"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

func decodeFederationOracle(raw []byte, kind string) (FederationObservation, error) {
	var v FederationObservation
	fields, ok := federationProjectionFields[kind]
	if !ok || len(raw) > MaxOracleBytes {
		return v, ErrOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || v.Kind != kind || !safeFederationIdentity(kind, v.Identity) || len(v.Settings) != len(fields) {
		return FederationObservation{}, ErrOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return FederationObservation{}, ErrOracle
	}
	for _, field := range fields {
		if _, ok := v.Settings[field]; !ok {
			return FederationObservation{}, ErrOracle
		}
	}
	return v, nil
}

func FederationOracleForIdentity(t *testing.T, ctx context.Context, kind, tenant, identity string) FederationObservation {
	t.Helper()
	if !safeFederationIdentity(kind, identity) {
		t.Fatal("invalid explicit oracle identity")
	}
	input, err := json.Marshal(map[string]string{"kind": kind, "tenant_id": tenant, "identity": identity})
	if err != nil {
		t.Fatal("invalid oracle selector")
	}
	defer clear(input)
	cmd := OracleCommand(ctx, "read-federation.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_MODULE_PATH="+os.Getenv("MSTEAMS_LAB_MODULE_PATH"), "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"))
	var output OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdin = bytes.NewReader(input)
	cmd.Stdout, cmd.Stderr = &output, io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned federation read oracle failed; no output retained")
	}
	v, err := decodeFederationOracle(output.Data, kind)
	if err != nil {
		t.Fatal("pinned federation read could not be verified")
	}
	if !SamePolicyName(v.Identity, identity) {
		t.Fatal("pinned federation identity differs")
	}
	return v
}

func CompareFederation(t *testing.T, kind string, raw map[string]json.RawMessage, observed FederationObservation) {
	t.Helper()
	want := map[string]json.RawMessage{}
	for _, field := range federationProjectionFields[kind] {
		want[field] = json.RawMessage("null")
		if v, ok := raw[field]; ok {
			want[field] = v
		}
	}
	var a, b any
	encodedA, _ := json.Marshal(want)
	encodedB, _ := json.Marshal(observed.Settings)
	if json.Unmarshal(encodedA, &a) != nil || json.Unmarshal(encodedB, &b) != nil || !reflect.DeepEqual(a, b) {
		t.Fatal("PowerShell and Go selected federation settings differ")
	}
}

// FederationRead reads one federation-family Global object and its filtered
// list through the SDK, writing response fixtures as it goes.
type FederationRead func(ctx context.Context, sdk *msteams.Client) (identity *string, raw map[string]json.RawMessage, listed []map[string]json.RawMessage, response *resty.Response, err error)

// RunFederationGlobal reads Global and a locally filtered list for one
// federation kind and compares the selected getters with pinned PowerShell.
// It never edits tenant configuration, grants assignments or creates/deletes a policy.
func RunFederationGlobal(t *testing.T, kind string, read FederationRead) {
	t.Helper()
	cfg := federationReadOnly(t)
	oracle := OracleEnabled(t)
	sdk, binding, _ := SDKClient(t)
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	identity, raw, listed, response, err := read(ctx, sdk)
	if err != nil {
		status := 0
		stage := "resource"
		if response != nil {
			status = response.StatusCode()
			if response.Request != nil && strings.HasSuffix(response.Request.URL, "/Teams.Tenant/serviceDiscovery") {
				stage = "discovery"
			}
		}
		t.Fatalf("selected federation read failed (HTTP %d, %s); no response payload retained", status, stage)
	}
	if identity == nil || !strings.EqualFold(*identity, "Global") || len(listed) != 1 || !SameConfigurationFields(raw, listed[0]) {
		if len(listed) == 1 {
			t.Log(ResponseComparisonFailure(raw, listed[0]))
			LogMetadataMismatches(t, raw, listed[0])
		}
		t.Fatal("Global federation read/list observations differ")
	}
	if oracle {
		tenant, _, ok := strings.Cut(binding, "|")
		if !ok || !DirectoryID.MatchString(tenant) {
			t.Fatal("canonical tenant binding required")
		}
		CompareFederation(t, kind, raw, FederationOracleForIdentity(t, ctx, kind, tenant, "Global"))
	}
	t.Log("selected federation read observations verified; effective policy and update/assignment behavior remain separately qualified")
}
