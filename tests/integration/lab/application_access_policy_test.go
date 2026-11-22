package lab_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	policy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/application_access_policy"
)

type policyCleanup struct{ service *policy.Service }

func (h policyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	items, _, err := h.service.List(ctx)
	if err != nil {
		return "", err
	}
	count := 0
	for _, item := range items {
		if labkit.SamePolicyName(item.Identity, entry.Fixture) {
			count++
		}
	}
	if count != 1 {
		return "", lab.ErrUnresolved
	}
	return entry.Fixture, nil
}

func (h policyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if entry.Fixture == "" || !strings.HasPrefix(entry.Fixture, "msteams-sdk-") || !labkit.SamePolicyName(entry.ResourceID, entry.Fixture) {
		return lab.ErrUnresolved
	}
	_, err := h.service.DeleteByIdentity(ctx, entry.ResourceID)
	if err != nil && !labkit.IsHTTP(err, 404) {
		return err
	}
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		_, _, err := h.service.GetByIdentity(ctx, entry.ResourceID)
		if err == nil {
			return false, nil
		}
		if !labkit.IsHTTP(err, 404) {
			return false, err
		}
		items, _, err := h.service.List(ctx)
		if err != nil {
			return false, err
		}
		for _, item := range items {
			if labkit.SamePolicyName(item.Identity, entry.Fixture) {
				return false, nil
			}
		}
		return true, nil
	})
}

func (policyCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }

func comparePowerShellPolicy(t *testing.T, ctx context.Context, name, description, appID string) {
	t.Helper()
	if os.Getenv("MSTEAMS_LAB_POWERSHELL_ORACLE") != "1" {
		return
	}
	module := os.Getenv("MSTEAMS_LAB_MODULE_PATH")
	if module == "" {
		t.Fatal("explicit pinned PowerShell module path is required")
	}
	command := labkit.OracleCommand(ctx, "read-application-access-policy.ps1")
	command.WaitDelay = time.Second
	command.Env = []string{}
	for _, key := range []string{"PATH", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR"} {
		if value, ok := os.LookupEnv(key); ok {
			command.Env = append(command.Env, key+"="+value)
		}
	}
	command.Env = append(command.Env, "MSTEAMS_LAB_AUTH_METHOD="+os.Getenv("MSTEAMS_LAB_AUTH_METHOD"), "MSTEAMS_LAB_MODULE_PATH="+module, "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"), "MSTEAMS_LAB_POLICY_FIXTURE="+name)
	var output labkit.OracleBuffer
	command.Stdout, command.Stderr = &output, io.Discard
	defer func() { clear(output.Data) }()
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned PowerShell read oracle failed; no output retained")
	}
	var observed struct {
		Success     bool     `json:"success"`
		Identity    string   `json:"identity"`
		AppIDs      []string `json:"app_ids"`
		Description string   `json:"description"`
	}
	if json.Unmarshal(output.Data, &observed) != nil || !observed.Success {
		t.Fatal("pinned PowerShell could not read the owned fixture")
	}
	if !labkit.SamePolicyName(observed.Identity, name) || observed.Description != description || len(observed.AppIDs) != 1 || observed.AppIDs[0] != appID {
		t.Fatal("pinned PowerShell and Go policy fields differ")
	}
	t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
}

// TestAccApplicationAccessPolicyCRUD creates only a uniquely named, unassigned
// policy. It never grants that policy to a user/group/tenant or edits Global.
func TestAccApplicationAccessPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("policy CRUD requires explicit lab read-only/mutation opt-in and private journal directory")
	} else if err != nil {
		t.Fatal(err)
	}
	sdk, binding, appID := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.ApplicationAccessPolicies
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private recovery journal unavailable")
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		_, cleanupErr := journal.Recover(ctx, map[string]lab.Handler{"application-access-policy": policyCleanup{service: service}})
		if cleanupErr != nil {
			t.Error("policy cleanup incomplete; retained private journal requires recovery")
		}
		if err := journal.Close(); err != nil {
			t.Error("private journal close failed")
		}
	})
	name, err := journal.FixtureName("app-access")
	if err != nil {
		t.Fatal("fixture name unavailable")
	}
	entry, err := journal.PlanCreate("application-access-policy", name)
	if err != nil {
		t.Fatal("creation was not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	initial := "SDK lab non-default initial description"
	created, createResponse, err := service.Create(ctx, policy.CreateRequest{Identity: name, AppIDs: policy.ReplaceAppIDs([]string{appID}), Description: &initial})
	if err != nil {
		t.Fatal("live policy creation/read-back failed; recovery journal retained")
	}
	if !labkit.SamePolicyName(created.Identity, name) || created.Description == nil || *created.Description != initial || len(created.AppIDs) != 1 || created.AppIDs[0] != appID {
		t.Fatal("live create read-back did not preserve non-default fields")
	}
	replacements := map[string]string{name: "fixture-policy", appID: "00000000-0000-4000-8000-000000000001"}
	labkit.WriteResponseFixture(t, "application_access_policy", "create_readback", createResponse, replacements)
	if err := journal.Applied(entry, name); err != nil {
		t.Fatal("created policy receipt was not journaled")
	}
	t.Log("policy create/read-back: passed")
	comparePowerShellPolicy(t, ctx, name, initial, appID)
	updated := "SDK lab non-default updated description"
	updateResponse, err := service.UpdateByIdentity(ctx, name, policy.UpdateRequest{Description: &updated})
	if err != nil {
		t.Fatal("live policy update failed")
	}
	labkit.WriteResponseFixture(t, "application_access_policy", "update", updateResponse, replacements)
	if err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		read, response, err := service.GetByIdentity(ctx, name)
		if err != nil {
			return false, err
		}
		matches := read != nil && read.Description != nil && *read.Description == updated
		preserved := read != nil && len(read.AppIDs) == 1 && read.AppIDs[0] == appID
		if matches && preserved {
			labkit.WriteResponseFixture(t, "application_access_policy", "update_readback", response, replacements)
		}
		return matches && preserved, nil
	}); err != nil {
		t.Fatal("non-default policy update and omitted-field preservation did not converge")
	}
	t.Log("policy update/read-back and omission: passed")
	comparePowerShellPolicy(t, ctx, name, updated, appID)
	items, listResponse, err := service.List(ctx)
	if err != nil {
		t.Fatal("live policy list failed")
	}
	found := 0
	for _, item := range items {
		if labkit.SamePolicyName(item.Identity, name) {
			found++
		}
	}
	if found != 1 {
		t.Fatal("owned policy was missing/ambiguous in live list")
	}
	labkit.WriteResponseFixture(t, "application_access_policy", "list", listResponse, replacements)
	t.Log("policy list: passed")
	if err := (policyCleanup{service}).Delete(ctx, lab.Entry{Fixture: name, ResourceID: name}); err != nil {
		t.Fatal("policy deletion and independent read/list absence were not verified")
	}
	_, deletedResponse, err := service.GetByIdentity(ctx, name)
	if !labkit.IsHTTP(err, 404) {
		t.Fatal("deleted policy did not return a typed 404")
	}
	labkit.WriteResponseFixture(t, "application_access_policy", "get_deleted", deletedResponse, replacements)
	t.Log("policy delete/absence: passed")
	// Cleanup deliberately runs again; its handler accepts a verified 404. The
	// private journal remains available if any earlier stage was interrupted.
}
