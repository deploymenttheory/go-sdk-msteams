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
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
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

type federationObservation struct {
	Success  bool                       `json:"success"`
	Kind     string                     `json:"kind"`
	Identity string                     `json:"identity"`
	Settings map[string]json.RawMessage `json:"settings"`
}

func decodeFederationOracle(raw []byte, kind string) (federationObservation, error) {
	var v federationObservation
	fields, ok := federationProjectionFields[kind]
	if !ok || len(raw) > maxTeamOracleBytes {
		return v, errTeamOracle
	}
	d := json.NewDecoder(bytes.NewReader(raw))
	d.DisallowUnknownFields()
	if d.Decode(&v) != nil || !v.Success || v.Kind != kind || !safeFederationIdentity(kind, v.Identity) || len(v.Settings) != len(fields) {
		return federationObservation{}, errTeamOracle
	}
	var extra any
	if d.Decode(&extra) != io.EOF {
		return federationObservation{}, errTeamOracle
	}
	for _, field := range fields {
		if _, ok := v.Settings[field]; !ok {
			return federationObservation{}, errTeamOracle
		}
	}
	return v, nil
}

func federationOracle(t *testing.T, ctx context.Context, kind, tenant string) federationObservation {
	return federationOracleForIdentity(t, ctx, kind, tenant, "Global")
}

func federationOracleForIdentity(t *testing.T, ctx context.Context, kind, tenant, identity string) federationObservation {
	t.Helper()
	if !safeFederationIdentity(kind, identity) {
		t.Fatal("invalid explicit oracle identity")
	}
	input, err := json.Marshal(map[string]string{"kind": kind, "tenant_id": tenant, "identity": identity})
	if err != nil {
		t.Fatal("invalid oracle selector")
	}
	defer clear(input)
	cmd := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-federation.ps1")
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
		t.Fatal("pinned federation read oracle failed; no output retained")
	}
	v, err := decodeFederationOracle(output.data, kind)
	if err != nil {
		t.Fatal("pinned federation read could not be verified")
	}
	if !samePolicyName(v.Identity, identity) {
		t.Fatal("pinned federation identity differs")
	}
	return v
}

func compareFederation(t *testing.T, kind string, raw map[string]json.RawMessage, observed federationObservation) {
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

// TestAccFederationReadOnly reads Global and a locally filtered list. It never
// edits tenant configuration, grants assignments or creates/deletes a policy.
func TestAccFederationReadOnly(t *testing.T) {
	cfg := federationReadOnly(t)
	oracle := teamOracleEnabled(t)
	for _, kind := range []string{"external-policy", "acs", "external-users", "tenant"} {
		t.Run(kind, func(t *testing.T) {
			sdk, binding, _ := labSDKClient(t)
			ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
			defer stop()
			var identity *string
			var raw map[string]json.RawMessage
			var response *resty.Response
			var err error
			var listed []map[string]json.RawMessage
			switch kind {
			case "external-policy":
				v, r, e := sdk.TeamsAdminAPI.ExternalAccessPolicies.GetByIdentity(ctx, "Global")
				writeResponseFixture(t, "external_access_policy", "get_global", r, nil)
				response, err = r, e
				if v != nil {
					identity, raw = v.Identity, v.RawFields
				}
				if err == nil {
					values, listResponse, e := sdk.TeamsAdminAPI.ExternalAccessPolicies.ListFiltered(ctx, "Global")
					writeResponseFixture(t, "external_access_policy", "list", listResponse, nil)
					err = e
					for _, v := range values {
						if v != nil {
							listed = append(listed, v.RawFields)
						}
					}
				}
			case "acs":
				v, r, e := sdk.TeamsAdminAPI.ACSFederationConfigurations.GetByIdentity(ctx, "Global")
				writeResponseFixture(t, "acs_federation_configuration", "get_global", r, nil)
				response, err = r, e
				if v != nil {
					identity, raw = v.Identity, v.RawFields
				}
				if err == nil {
					values, listResponse, e := sdk.TeamsAdminAPI.ACSFederationConfigurations.ListFiltered(ctx, "Global")
					writeResponseFixture(t, "acs_federation_configuration", "list", listResponse, nil)
					err = e
					for _, v := range values {
						if v != nil {
							listed = append(listed, v.RawFields)
						}
					}
				}
			case "external-users":
				v, r, e := sdk.TeamsAdminAPI.TeamsExternalAccessConfigurations.GetByIdentity(ctx, "Global")
				writeResponseFixture(t, "teams_external_access_configuration", "get_global", r, nil)
				response, err = r, e
				if v != nil {
					identity, raw = v.Identity, v.RawFields
				}
				if err == nil {
					values, listResponse, e := sdk.TeamsAdminAPI.TeamsExternalAccessConfigurations.ListFiltered(ctx, "Global")
					writeResponseFixture(t, "teams_external_access_configuration", "list", listResponse, nil)
					err = e
					for _, v := range values {
						if v != nil {
							listed = append(listed, v.RawFields)
						}
					}
				}
			case "tenant":
				v, r, e := sdk.TeamsAdminAPI.TenantFederationConfigurations.GetByIdentity(ctx, "Global")
				writeResponseFixture(t, "tenant_federation_configuration", "get_global", r, nil)
				response, err = r, e
				if v != nil {
					identity, raw = &v.Identity, v.RawFields
				}
				if err == nil {
					values, listResponse, e := sdk.TeamsAdminAPI.TenantFederationConfigurations.ListFiltered(ctx, "Global")
					writeResponseFixture(t, "tenant_federation_configuration", "list", listResponse, nil)
					err = e
					for _, v := range values {
						listed = append(listed, v.RawFields)
					}
				}
			}
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
			if identity == nil || !strings.EqualFold(*identity, "Global") || len(listed) != 1 || !sameResponseFields(raw, listed[0]) {
				if len(listed) == 1 {
					t.Log(responseComparisonFailure(raw, listed[0]))
					logMetadataMismatches(t, raw, listed[0])
				}
				t.Fatal("Global federation read/list observations differ")
			}
			if oracle {
				tenant, _, ok := strings.Cut(binding, "|")
				if !ok || !labDirectoryID.MatchString(tenant) {
					t.Fatal("canonical tenant binding required")
				}
				compareFederation(t, kind, raw, federationOracle(t, ctx, kind, tenant))
			}
			t.Log("selected federation read observations verified; effective policy and update/assignment behavior remain separately qualified")
		})
	}
}

func TestUnitFederationOracleRejectsMixedAndIncompleteOutput(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"success":false}`, `{"success":true,"kind":"acs","identity":"Global","settings":{}}`, `{"success":true,"kind":"acs","identity":"Global","settings":{"EnableAcsUsers":false,"RequireAcsFederationForMeeting":null,"extra":false}}`, `{"success":true,"kind":"tenant","identity":"Global","settings":{"EnableAcsUsers":false,"RequireAcsFederationForMeeting":null}}`} {
		if _, err := decodeFederationOracle([]byte(raw), "acs"); err == nil {
			t.Fatal("invalid oracle projection accepted")
		}
	}
	if _, err := decodeFederationOracle([]byte(`{"success":true,"kind":"acs","identity":"Global","settings":{"EnableAcsUsers":false,"RequireAcsFederationForMeeting":null}}`), "acs"); err != nil {
		t.Fatal("valid selected nullable projection rejected")
	}
}
