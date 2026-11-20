package lab_test

import (
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"os/exec"
	"runtime"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	policy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/application_access_policy"
)

func privateLabFile(path string) ([]byte, error) {
	//nolint:gosec // G304: explicitly selected private lab files, after live opt-in.
	file, err := os.Open(path)
	if err != nil {
		return nil, lab.ErrConfig
	}
	defer func() { _ = file.Close() }()
	info, err := file.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > 1<<20 {
		return nil, lab.ErrConfig
	}
	if runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		return nil, lab.ErrConfig
	}
	data, err := io.ReadAll(io.LimitReader(file, (1<<20)+1))
	if err != nil || len(data) > 1<<20 {
		return nil, lab.ErrConfig
	}
	return data, nil
}

func policyLabClient(t *testing.T) (*msteams.Client, string, string) {
	t.Helper()
	file := os.Getenv("MSTEAMS_LAB_IDENTITY_FILE")
	if file == "" {
		t.Fatal("explicit private lab identity file is required")
	}
	raw, err := privateLabFile(file)
	if err != nil {
		t.Fatal("private lab identity unavailable")
	}
	defer clear(raw)
	var identity struct {
		Tenant      string `json:"tenant_id"`
		Client      string `json:"client_id"`
		Secret      string `json:"client_secret"`
		Certificate string `json:"certificate_file"`
		Key         string `json:"private_key_file"`
	}
	if json.Unmarshal(raw, &identity) != nil {
		t.Fatal("invalid lab identity")
	}
	method := os.Getenv("MSTEAMS_LAB_AUTH_METHOD")
	cfg := &auth.Config{Method: auth.Certificate, TenantID: identity.Tenant, ClientID: identity.Client}
	switch method {
	case "", "certificate":
		cfg.Certificate, err = privateLabFile(identity.Certificate)
		if err != nil {
			t.Fatal("private lab certificate unavailable")
		}
		cfg.PrivateKey, err = privateLabFile(identity.Key)
		if err != nil {
			t.Fatal("private lab key unavailable")
		}
		defer clear(cfg.PrivateKey)
	case "client-secret":
		cfg.Method = auth.ClientSecret
		cfg.ClientSecret = identity.Secret
	default:
		t.Fatal("unsupported explicit lab auth method")
	}
	sdk, err := msteams.NewClient(cfg)
	identity.Secret = ""
	cfg.ClientSecret = ""
	if err != nil {
		t.Fatal("lab SDK construction failed")
	}
	t.Cleanup(func() {
		if err := sdk.Close(); err != nil {
			t.Error("lab SDK cleanup failed")
		}
	})
	return sdk, identity.Tenant + "|commercial|" + identity.Client, identity.Client
}

type policyCleanup struct{ service *policy.Service }

func samePolicyName(identity, fixture string) bool {
	if len(identity) > 4 && strings.EqualFold(identity[:4], "tag:") {
		identity = identity[4:]
	}
	return strings.EqualFold(identity, fixture)
}

func (h policyCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	items, _, err := h.service.List(ctx)
	if err != nil {
		return "", err
	}
	count := 0
	for _, item := range items {
		if samePolicyName(item.Identity, entry.Fixture) {
			count++
		}
	}
	if count > 1 {
		return "", lab.ErrUnresolved
	}
	if count == 1 {
		return entry.Fixture, nil
	}
	return "", nil
}

func (h policyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if entry.Fixture == "" || !strings.HasPrefix(entry.Fixture, "msteams-sdk-") || !samePolicyName(entry.ResourceID, entry.Fixture) {
		return lab.ErrUnresolved
	}
	_, err := h.service.DeleteByIdentity(ctx, entry.ResourceID)
	var apiError *client.APIError
	if errors.As(err, &apiError) && apiError.StatusCode == 404 {
		return nil
	}
	return err
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
	command := exec.CommandContext(ctx, "pwsh", "-NoLogo", "-NoProfile", "-NonInteractive", "-File", "oracle/read-application-access-policy.ps1")
	command.Env = []string{}
	for _, key := range []string{"PATH", "TMPDIR", "TMP", "TEMP", "LANG", "DOTNET_ROOT", "SystemRoot", "WINDIR"} {
		if value, ok := os.LookupEnv(key); ok {
			command.Env = append(command.Env, key+"="+value)
		}
	}
	command.Env = append(command.Env, "MSTEAMS_LAB_MODULE_PATH="+module, "MSTEAMS_LAB_IDENTITY_FILE="+os.Getenv("MSTEAMS_LAB_IDENTITY_FILE"), "MSTEAMS_LAB_POLICY_FIXTURE="+name)
	output, err := command.Output()
	defer clear(output)
	if err != nil {
		t.Fatal("pinned PowerShell read oracle failed; no output retained")
	}
	var observed struct {
		Success     bool     `json:"success"`
		Identity    string   `json:"identity"`
		AppIDs      []string `json:"app_ids"`
		Description string   `json:"description"`
	}
	if json.Unmarshal(output, &observed) != nil || !observed.Success {
		t.Fatal("pinned PowerShell could not read the owned fixture")
	}
	if !samePolicyName(observed.Identity, name) || observed.Description != description || len(observed.AppIDs) != 1 || observed.AppIDs[0] != appID {
		t.Fatal("pinned PowerShell and Go policy fields differ")
	}
	t.Log("PowerShell 8.1.0 / Go policy read parity: passed")
}

// TestLabApplicationAccessPolicyCRUD creates only a uniquely named, unassigned
// policy. It never grants that policy to a user/group/tenant or edits Global.
func TestLabApplicationAccessPolicyCRUD(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("policy CRUD requires explicit lab read-only/mutation opt-in and private journal directory")
	} else if err != nil {
		t.Fatal(err)
	}
	sdk, binding, appID := policyLabClient(t)
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
	created, _, err := service.Create(ctx, policy.CreateRequest{Identity: name, AppIDs: policy.ReplaceAppIDs([]string{appID}), Description: &initial})
	if err != nil {
		t.Fatal("live policy creation/read-back failed; recovery journal retained")
	}
	if !samePolicyName(created.Identity, name) || created.Description == nil || *created.Description != initial || len(created.AppIDs) != 1 || created.AppIDs[0] != appID {
		t.Fatal("live create read-back did not preserve non-default fields")
	}
	if err := journal.Applied(entry, name); err != nil {
		t.Fatal("created policy receipt was not journaled")
	}
	t.Log("policy create/read-back: passed")
	comparePowerShellPolicy(t, ctx, name, initial, appID)
	updated := "SDK lab non-default updated description"
	if _, err := service.UpdateByIdentity(ctx, name, policy.UpdateRequest{Description: &updated}); err != nil {
		t.Fatal("live policy update failed")
	}
	read, _, err := service.GetByIdentity(ctx, name)
	if err != nil {
		t.Fatal("live policy update read-back failed")
	}
	if read.Description == nil || *read.Description != updated || len(read.AppIDs) != 1 || read.AppIDs[0] != appID {
		t.Fatal("sparse update changed an omitted field or lost the non-default update")
	}
	t.Log("policy update/read-back and omission: passed")
	comparePowerShellPolicy(t, ctx, name, updated, appID)
	items, _, err := service.List(ctx)
	if err != nil {
		t.Fatal("live policy list failed")
	}
	found := 0
	for _, item := range items {
		if samePolicyName(item.Identity, name) {
			found++
		}
	}
	if found != 1 {
		t.Fatal("owned policy was missing/ambiguous in live list")
	}
	t.Log("policy list: passed")
	if _, err := service.DeleteByIdentity(ctx, name); err != nil {
		t.Fatal("live policy delete failed")
	}
	_, _, err = service.GetByIdentity(ctx, name)
	var apiError *client.APIError
	if !errors.As(err, &apiError) || apiError.StatusCode != 404 {
		t.Fatal("deleted policy absence was not verified")
	}
	t.Log("policy delete/absence: passed")
	// Cleanup deliberately runs again; its handler accepts a verified 404. The
	// private journal remains available if any earlier stage was interrupted.
}
