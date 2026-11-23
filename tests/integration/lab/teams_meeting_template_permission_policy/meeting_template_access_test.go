package teamsmeetingtemplatepermissionpolicyacc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"slices"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	permission "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_template_permission_policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func populatedTemplateUnauthorized(raw []byte) bool {
	var body struct {
		Code      string `json:"code"`
		ErrorCode int    `json:"errorCode"`
		Inner     struct {
			Metadata struct {
				Parameters []string `json:"ParameterList"`
			} `json:"ErrorCodeMetadata"`
		} `json:"innerError"`
	}
	if json.Unmarshal(raw, &body) != nil || body.Code != "ClientError" || body.ErrorCode != 40013 {
		return false
	}
	for _, parameter := range body.Inner.Metadata.Parameters {
		var detail struct {
			Status int `json:"status"`
		}
		if json.Unmarshal([]byte(parameter), &detail) == nil && detail.Status == 401 {
			return true
		}
	}
	return false
}

type populatedTemplateBaseline struct {
	Success  bool   `json:"success"`
	Tenant   string `json:"tenant_id"`
	Identity string `json:"identity"`
	Outcome  string `json:"outcome"`
	Code     int    `json:"service_error_code"`
	Status   int    `json:"downstream_status"`
}

func comparePopulatedTemplateCreate(t *testing.T, ctx context.Context, tenant, name string, ids []string) {
	t.Helper()
	input, err := json.Marshal(map[string]any{"tenant_id": tenant, "identity": name, "kind": "permission", "template_ids": ids})
	if err != nil {
		t.Fatal("invalid populated policy baseline selector")
	}
	defer clear(input)
	cmd := labkit.OracleCommand(ctx, "create-meeting-template-permission.ps1")
	cmd.WaitDelay = time.Second
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath", "MSTEAMS_LAB_AUTH_METHOD", "MSTEAMS_LAB_MODULE_PATH", "MSTEAMS_LAB_IDENTITY_FILE"} {
		if value, ok := os.LookupEnv(key); ok {
			cmd.Env = append(cmd.Env, key+"="+value)
		}
	}
	cmd.Env = append(cmd.Env, "MSTEAMS_LAB_MUTATIONS=1", "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1")
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	cmd.Stdin, cmd.Stdout, cmd.Stderr = bytes.NewReader(input), &output, io.Discard
	if cmd.Run() != nil || output.Exceeded {
		t.Fatal("pinned populated policy create process failed")
	}
	var actual populatedTemplateBaseline
	decoder := json.NewDecoder(bytes.NewReader(output.Data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&actual) != nil || decoder.Decode(new(any)) != io.EOF || !actual.Success || !strings.EqualFold(actual.Tenant, tenant) || !labkit.SamePolicyName(actual.Identity, name) || actual.Outcome != "downstream-unauthorized" || actual.Code != 40013 || actual.Status != 401 {
		t.Fatal("pinned populated policy authorization outcome differs")
	}
}

// TestAccMeetingTemplatePermissionPopulatedAccess compares a verified app-only
// rejection. Passing this test does not qualify populated-list CRUD.
func TestAccMeetingTemplatePermissionPopulatedAccess(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid populated policy lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("populated policy authorization comparison requires mutation opt-in")
	} else if err != nil {
		t.Fatal("invalid populated policy mutation enablement")
	}
	if !labkit.OracleEnabled(t) {
		t.Fatal("pinned populated policy comparison is required")
	}
	sdk, binding, _ := labkit.SDKClient(t)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok {
		t.Fatal("populated policy tenant binding unavailable")
	}
	ids := existingMeetingTemplateIDs(t, ctx, tenant)
	service := sdk.TeamsAdminAPI.TeamsMeetingTemplatePermissionPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("populated policy recovery journal unavailable")
	}
	handlers := map[string]lab.Handler{"meeting-template-policy": meetingTemplatePolicyCleanup{service}, "template-baseline-policy": meetingTemplatePolicyCleanup{service}}
	t.Cleanup(func() {
		cleanup, stop := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer stop()
		if _, err := journal.Recover(cleanup, handlers); err != nil {
			t.Error("populated policy comparison cleanup unresolved")
		}
		if err := journal.Close(); err != nil {
			t.Error("populated policy comparison journal close failed")
		}
	})
	names := make([]string, 0, 2)
	for _, kind := range []string{"meeting-template-policy", "template-baseline-policy"} {
		name, err := journal.FixtureName(kind)
		if err != nil {
			t.Fatal(err)
		}
		if _, err := journal.PlanCreate(kind, name); err != nil {
			t.Fatal("populated policy comparison intent not journaled")
		}
		names = append(names, name)
	}
	_, response, err := service.CreateByIdentity(ctx, names[0], permission.CreateRequest{ScalarSettings: permission.ScalarSettings{Description: policy.Set("SDK lab non-default initial description")}, HiddenMeetingTemplates: permission.ReplaceHiddenTemplates(permission.NewHiddenTemplate(ids[0]))})
	labkit.WriteResponseFixture(t, "teams_meeting_template_permission_policy", "create_populated_denied", response, map[string]string{names[0]: "fixture-policy"})
	var api *client.APIError
	if !errors.As(err, &api) || response == nil || api.StatusCode != 400 || !slices.Contains(api.ServiceErrorCodes, 40013) || !populatedTemplateUnauthorized(response.Body()) {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("native populated policy authorization outcome changed")
	}
	t.Log("native populated-list create denial: verified")
	comparePopulatedTemplateCreate(t, ctx, tenant, names[1], ids[:1])
	t.Log("PowerShell 8.1.0 / Go populated-list authorization parity: passed")
	if _, err := journal.Recover(ctx, handlers); err != nil {
		t.Fatal("populated policy comparison absence not verified")
	}
	t.Log("populated-list comparison cleanup: passed")
}

func TestUnitPopulatedAuthorizationUsesAcceptanceEnvelope(t *testing.T) {
	t.Parallel()
	directory, err := os.OpenRoot(labkit.Root + "/msteams/teams_admin_api/teams_meeting_template_permission_policy/testdata")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := directory.Close(); err != nil {
			t.Error("permission fixture directory close failed")
		}
	})
	raw, err := directory.ReadFile("create_populated_denied_400.json")
	if err != nil {
		t.Fatal(err)
	}
	if !populatedTemplateUnauthorized(raw) {
		t.Fatal("actual downstream authorization envelope not recognized")
	}
	for _, bad := range [][]byte{[]byte("null"), []byte(strings.Replace(string(raw), `"errorCode": 40013`, `"errorCode": 40003`, 1)), []byte(strings.ReplaceAll(string(raw), `\"status\":401`, `\"status\":403`))} {
		if bytes.Equal(raw, bad) {
			t.Fatal("invalid authorization test did not alter its fixture")
		}
		if populatedTemplateUnauthorized(bad) {
			t.Fatal("different or malformed service outcome accepted")
		}
	}
}
