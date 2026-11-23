package compliancerecordingacc

import (
	"bytes"
	"context"
	"crypto/rand"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"os"
	"slices"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	cr "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/compliance_recording"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const fixtureService = "compliance_recording"

type policyObservation struct {
	Success                  bool    `json:"success"`
	Identity                 string  `json:"identity"`
	Enabled                  *bool   `json:"enabled"`
	WarnUserOnRemoval        *bool   `json:"warn_user_on_removal"`
	DisableAudioNotification *bool   `json:"disable_audio_notification"`
	Description              *string `json:"description"`
	RecordReroutedCalls      *bool   `json:"record_rerouted_calls"`
	CustomPromptsEnabled     *bool   `json:"custom_prompts_enabled"`
	CustomPromptsPackageID   *string `json:"custom_prompts_package_id"`
	CustomBanner             *string `json:"custom_banner"`
	ApplicationCount         int     `json:"application_count"`
}

func powerShellPolicy(t *testing.T, ctx context.Context, fixture string) policyObservation {
	t.Helper()
	command := labkit.OracleCommand(ctx, "read-compliance-recording-policy.ps1")
	command.WaitDelay = time.Second
	command.Env = []string{}
	for _, key := range []string{"PATH", "HOME", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR", "PSModuleAnalysisCachePath"} {
		if value, ok := os.LookupEnv(key); ok {
			command.Env = append(command.Env, key+"="+value)
		}
	}
	for _, key := range []string{"MSTEAMS_LAB_IDENTITY_FILE", "MSTEAMS_LAB_MODULE_PATH", "MSTEAMS_LAB_AUTH_METHOD"} {
		command.Env = append(command.Env, key+"="+os.Getenv(key))
	}
	command.Env = append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_POLICY_FIXTURE="+fixture)
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	command.Stdout, command.Stderr = &output, io.Discard
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned PowerShell compliance recording read failed; no output retained")
	}
	var observed policyObservation
	decoder := json.NewDecoder(bytes.NewReader(output.Data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&observed) != nil || !observed.Success {
		t.Fatal("pinned PowerShell compliance recording projection unavailable")
	}
	return observed
}

func equalPtr[T comparable](a, b *T) bool {
	return (a == nil && b == nil) || (a != nil && b != nil && *a == *b)
}

func sameProjection(value *cr.Policy, observed policyObservation) bool {
	return value != nil && value.Identity != nil && labkit.SamePolicyName(observed.Identity, *value.Identity) &&
		equalPtr(value.Enabled, observed.Enabled) && equalPtr(value.WarnUserOnRemoval, observed.WarnUserOnRemoval) &&
		equalPtr(value.DisableComplianceRecordingAudioNotificationForCalls, observed.DisableAudioNotification) &&
		equalPtr(value.Description, observed.Description) && equalPtr(value.RecordReroutedCalls, observed.RecordReroutedCalls) &&
		equalPtr(value.CustomPromptsEnabled, observed.CustomPromptsEnabled) && equalPtr(value.CustomPromptsPackageID, observed.CustomPromptsPackageID) &&
		equalPtr(value.CustomBanner, observed.CustomBanner) && len(value.ComplianceRecordingApplications) == observed.ApplicationCount
}

func randomGUID() string {
	var b [16]byte
	_, _ = rand.Read(b[:])
	b[6] = b[6]&0x0f | 0x40
	b[8] = b[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", b[:4], b[4:6], b[6:8], b[8:10], b[10:])
}

// TestAccComplianceRecordingPolicyCRUD manages one uniquely named, disabled,
// unassigned policy. It never changes Global or grants a policy. Applications
// need a real resource account, which the lab cannot create, so the
// application operations are proven through their service rejections.
func TestAccComplianceRecordingPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("compliance recording CRUD requires explicit mutation opt-in and recovery directory")
	} else if err != nil {
		t.Fatal("invalid mutation enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.ComplianceRecording
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private compliance recording recovery journal unavailable")
	}
	cleanup := policyCleanup{service}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"compliance-recording-policy": cleanup}); err != nil {
			t.Error("compliance recording cleanup unresolved; private journal retained")
		}
		if err := journal.Close(); err != nil {
			t.Error("compliance recording journal close failed")
		}
	})
	name, err := journal.FixtureName("compliance-policy")
	if err != nil {
		t.Fatal(err)
	}
	entry, err := journal.PlanCreate("compliance-recording-policy", name)
	if err != nil {
		t.Fatal("compliance recording create intent not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := map[string]string{name: "fixture-policy"}
	compare := func(value *cr.Policy) {
		if !oracle {
			return
		}
		if !sameProjection(value, powerShellPolicy(t, ctx, name)) {
			t.Fatal("SDK and PowerShell compliance recording policy projections differ")
		}
		t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
	}

	initial := "SDK lab non-default initial description"
	_, response, err := service.CreatePolicyByIdentity(ctx, name, cr.PolicySettings{Enabled: policy.Set(false), WarnUserOnRemoval: policy.Set(false),
		RecordReroutedCalls: policy.Set(true), Description: policy.Set(initial)}, policy.Controls{})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("compliance recording policy create failed; recovery journal retained")
	}
	labkit.WriteResponseFixture(t, fixtureService, "create", response, replacements)
	var value *cr.Policy
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetPolicyByIdentity(ctx, name)
		if labkit.IsHTTP(err, 404) {
			return false, nil
		}
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Identity != nil && labkit.SamePolicyName(*read.Identity, name) && read.Description != nil && *read.Description == initial &&
			read.WarnUserOnRemoval != nil && !*read.WarnUserOnRemoval && read.RecordReroutedCalls != nil && *read.RecordReroutedCalls && read.Enabled != nil && !*read.Enabled
		if matched {
			value = read
			labkit.WriteResponseFixture(t, fixtureService, "create_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("compliance recording non-default create read-back did not converge")
	}
	identity := *value.Identity
	if err := journal.Applied(entry, identity); err != nil {
		t.Fatal("compliance recording creation receipt not journaled")
	}
	t.Log("policy create/read-back: passed")
	compare(value)

	// Applications: an unknown application instance ID is rejected (40013);
	// reads, updates and removes of a missing application are 404s.
	app := randomGUID()
	_, response, err = service.CreateApplicationByID(ctx, identity, app, cr.ApplicationSettings{RequiredDuringCall: policy.Set(true)}, policy.Controls{})
	var api *client.APIError
	if !errors.As(err, &api) || api.StatusCode != 400 || !slices.Contains(api.ServiceErrorCodes, 40013) {
		t.Fatal("unknown application instance was not rejected")
	}
	labkit.WriteResponseFixture(t, fixtureService, "application_create_invalid", response, map[string]string{name: "fixture-policy", app: "00000000-0000-4000-8000-000000000001"})
	if _, response, err = service.GetApplicationByID(ctx, identity, app); !labkit.IsHTTP(err, 404) {
		t.Fatal("missing application read was not a 404")
	}
	labkit.WriteResponseFixture(t, fixtureService, "application_missing", response, map[string]string{name: "fixture-policy", app: "00000000-0000-4000-8000-000000000001"})
	if _, _, err = service.UpdateApplicationByID(ctx, identity, app, cr.ApplicationUpdate{ApplicationSettings: cr.ApplicationSettings{RequiredDuringCall: policy.Set(false)}}); !labkit.IsHTTP(err, 404) {
		t.Fatal("missing application update was not a 404")
	}
	if _, _, err = service.DeleteApplicationByID(ctx, identity, app, policy.Controls{}); !labkit.IsHTTP(err, 404) {
		t.Fatal("missing application remove was not a 404")
	}
	apps, response, err := service.ListApplications(ctx)
	if err != nil {
		t.Fatal("application list failed")
	}
	for _, item := range apps {
		if item != nil && item.Identity != nil && labkit.SamePolicyName(*item.Identity, identity+"/"+app) {
			t.Fatal("rejected application was listed")
		}
	}
	labkit.WriteResponseFixture(t, fixtureService, "application_list", response, replacements)
	t.Log("application rejection and absence: passed")

	updated := "SDK lab non-default updated description"
	_, response, err = service.UpdatePolicyByIdentity(ctx, identity, cr.PolicyUpdate{PolicySettings: cr.PolicySettings{Description: policy.Set(updated)}})
	if err != nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("compliance recording sparse update failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "update", response, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetPolicyByIdentity(ctx, identity)
		if err != nil {
			return false, err
		}
		matched := read != nil && read.Description != nil && *read.Description == updated && equalPtr(read.WarnUserOnRemoval, value.WarnUserOnRemoval) &&
			equalPtr(read.RecordReroutedCalls, value.RecordReroutedCalls) && equalPtr(read.Enabled, value.Enabled)
		if matched {
			value = read
			labkit.WriteResponseFixture(t, fixtureService, "update_readback", response, replacements)
		}
		return matched, nil
	}); err != nil {
		t.Fatal("compliance recording update or omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	compare(value)

	values, listResponse, err := service.ListPolicies(ctx)
	if err != nil {
		t.Fatal("compliance recording policy list failed")
	}
	count := 0
	for _, item := range values {
		if item != nil && item.Identity != nil && labkit.SamePolicyName(*item.Identity, name) {
			count++
		}
	}
	if count != 1 || len(values) == 0 || values[0].Identity == nil || *values[0].Identity != "Global" {
		t.Fatal("owned compliance recording policy missing, ambiguous or Global not first")
	}
	labkit.WriteResponseFixture(t, fixtureService, "owned_list", listResponse, replacements)
	t.Log("policy list: passed")

	_, deleted, err := service.DeletePolicyByIdentity(ctx, identity, policy.Controls{})
	if err != nil {
		labkit.LogAccessResponse(t, deleted, err)
		t.Fatal("owned policy delete failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "delete", deleted, replacements)
	if err := cleanup.Delete(ctx, lab.Entry{Fixture: name, ResourceID: identity}); err != nil {
		t.Fatal("compliance recording deletion and independent absence were not verified")
	}
	_, response, err = service.GetPolicyByIdentity(ctx, identity)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("deleted policy did not return typed 404")
	}
	labkit.WriteResponseFixture(t, fixtureService, "get_deleted", response, replacements)
	t.Log("policy delete/absence: passed")
}

// TestAccComplianceRecordingReadOnly reads Global, the policy list and the
// application list without changing the tenant.
func TestAccComplianceRecordingReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab configuration")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("compliance recording reads require explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid read-only enablement")
	}
	sdk, _, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.ComplianceRecording
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	global, response, err := service.GetPolicyByIdentity(ctx, "Global")
	if err != nil || global == nil || global.Identity == nil || *global.Identity != "Global" || global.Enabled == nil {
		labkit.LogAccessResponse(t, response, err)
		t.Fatal("Global compliance recording policy read failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "get_global", response, nil)
	values, response, err := service.ListPolicies(ctx)
	if err != nil || len(values) == 0 || values[0].Identity == nil || *values[0].Identity != "Global" {
		t.Fatal("compliance recording policy list failed or Global not first")
	}
	labkit.WriteResponseFixture(t, fixtureService, "list", response, nil)
	if _, _, err := service.ListApplications(ctx); err != nil {
		t.Fatal("compliance recording application list failed")
	}
	t.Log("compliance recording reads: passed")
}
