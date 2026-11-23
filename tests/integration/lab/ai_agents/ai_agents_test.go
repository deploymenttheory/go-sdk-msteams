package aiagentsacc

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"io"
	"os"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	aiagents "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/ai_agents"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const (
	fixtureService = "ai_agents"
	// Reserved fixture-form IDs that cannot name a real configuration, provider
	// or tag template.
	probeConfiguration = "00000000-0000-4000-8000-000000000001"
	probeProvider      = "00000000-0000-4000-8000-000000000003"
	probeAgent         = "00000000-0000-4000-8000-0000000000a1"
	probeTagTemplate   = "00000000-0000-4000-8000-000000000009"
	// A thread ID in the chat service's shape that names no real meeting, so the
	// knowledge probes cannot read or remove anything.
	probeThread = "19:msteams-sdk-nonexistent-0000@thread.v2"
)

func tenantReplacements(t *testing.T, binding string) map[string]string {
	t.Helper()
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok || !labkit.DirectoryID.MatchString(tenant) {
		t.Fatal("tenant binding required")
	}
	return map[string]string{tenant: "00000000-0000-4000-8000-0000000000aa", strings.ToLower(tenant): "00000000-0000-4000-8000-0000000000aa"}
}

type agentObservation struct {
	Success      bool   `json:"success"`
	Name         string `json:"name"`
	AgentID      string `json:"agent_id"`
	AgentType    string `json:"agent_type"`
	CatalogCount int    `json:"catalog_count"`
}

func powerShellAgent(t *testing.T, ctx context.Context, id string) agentObservation {
	t.Helper()
	if !configurationID.MatchString(id) {
		t.Fatal("owned agent selector required")
	}
	command := labkit.OracleCommand(ctx, "read-ai-agent.ps1")
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
	command.Env = append(command.Env, "MSTEAMS_LAB_READ_ONLY=1", "MSTEAMS_LAB_POWERSHELL_ORACLE=1", "MSTEAMS_LAB_AI_AGENT_ID="+id, "MSTEAMS_LAB_AI_AGENT_PROVIDER="+probeProvider)
	var output labkit.OracleBuffer
	defer func() { clear(output.Data) }()
	command.Stdout, command.Stderr = &output, io.Discard
	if command.Run() != nil || output.Exceeded {
		t.Fatal("pinned PowerShell AI agent read failed; no output retained")
	}
	var observed agentObservation
	decoder := json.NewDecoder(bytes.NewReader(output.Data))
	decoder.DisallowUnknownFields()
	if decoder.Decode(&observed) != nil || !observed.Success {
		t.Fatal("pinned PowerShell AI agent projection unavailable")
	}
	return observed
}

// TestAccAIAgentReads lists agent configurations, reads a missing one and
// queries the platform catalog for a provider that does not exist. It changes
// nothing.
func TestAccAIAgentReads(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.AIAgents
	replacements := tenantReplacements(t, binding)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()

	listed, _, err := service.List(ctx, nil)
	if err != nil || listed == nil || listed.AgentCount == nil || int(*listed.AgentCount) < len(listed.Agents) {
		t.Fatal("AI agent list failed")
	}
	first, exclude := int32(1), true
	page, _, err := service.List(ctx, &aiagents.ListOptions{First: &first, ExcludeContent: &exclude})
	if err != nil || len(page.Agents) > 1 {
		t.Fatal("paged AI agent list failed")
	}
	if _, response, err := service.GetByID(ctx, probeConfiguration); !labkit.IsHTTP(err, 404) {
		t.Fatal("missing AI agent was not a 404")
	} else {
		labkit.WriteResponseFixture(t, fixtureService, "get_missing", response, replacements)
	}
	t.Log("AI agent list/get: passed")

	limit, count := int32(5), true
	found, response, err := service.QueryPlatformAgents(ctx, aiagents.QueryRequest{ProviderID: probeProvider, MaxResults: &limit, ShowCount: &count})
	if err != nil || found == nil || len(found.Agents) != 0 || found.AgentCount == nil || *found.AgentCount != 0 {
		t.Fatal("platform AI agent query for an unknown provider changed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "query_empty", response, replacements)
	if _, response, err = service.QueryPlatformAgents(ctx, aiagents.QueryRequest{ProviderID: "msteams-sdk-probe"}); !labkit.IsHTTP(err, 400) {
		t.Fatal("non-GUID provider was not rejected")
	}
	labkit.WriteResponseFixture(t, fixtureService, "query_provider_not_guid", response, replacements)
	t.Log("platform AI agent query: passed")
}

// TestAccAIAgentRejectedWrites sends agent writes the service rejects, and
// the AI-generated knowledge requests for a thread that does not exist. Nothing
// in the tenant changes.
func TestAccAIAgentRejectedWrites(t *testing.T) {
	cfg := labkit.ReadOnlyEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.AIAgents
	replacements := tenantReplacements(t, binding)
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()

	request := aiagents.AgentRequest{Name: "msteams-sdk-probe", AgentID: probeAgent, AgentType: "msteams-sdk-unknown"}
	if _, response, err := service.Create(ctx, request); !labkit.IsHTTP(err, 400) {
		t.Fatal("unknown agent type create was not rejected")
	} else {
		labkit.WriteResponseFixture(t, fixtureService, "create_unknown_type", response, replacements)
	}
	if _, response, err := service.UpdateByID(ctx, probeConfiguration, request); !labkit.IsHTTP(err, 400) {
		t.Fatal("unknown agent type update was not rejected")
	} else {
		labkit.WriteResponseFixture(t, fixtureService, "update_unknown_type", response, replacements)
	}
	if response, err := service.DeleteByID(ctx, probeConfiguration); !labkit.IsHTTP(err, 404) {
		t.Fatal("missing agent delete was not a 404")
	} else {
		labkit.WriteResponseFixture(t, fixtureService, "delete_missing", response, replacements)
	}
	t.Log("AI agent write rejection: passed")

	// MCPS issues application tokens for its resource but rejects them with an
	// empty 401, so neither request can reach meeting data. The thread ID is
	// fabricated as well; a delete never targets a real meeting.
	knowledge := aiagents.KnowledgeRequest{ThreadID: probeThread}
	_, response, err := service.ListKnowledgeContainers(ctx, knowledge)
	if err == nil || response == nil || response.StatusCode() != 401 {
		t.Fatal("app-only knowledge container read was not rejected")
	}
	labkit.WriteResponseFixture(t, fixtureService, "knowledge_containers_app_only", response, replacements)
	response, err = service.DeleteKnowledge(ctx, knowledge)
	if err == nil || response == nil || response.StatusCode() != 401 {
		t.Fatal("app-only knowledge delete was not rejected")
	}
	t.Log("AI-generated knowledge app-only rejection: passed")
}

// TestAccAIAgentLifecycle creates one uniquely named agent configuration,
// reads, lists, renames and deletes it. Agent configurations carry no licence
// or billing; nothing else in the tenant is touched.
func TestAccAIAgentLifecycle(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("AI agent lifecycle requires explicit mutation opt-in and a private journal directory")
	} else if err != nil {
		t.Fatal(err)
	}
	sdk, binding, _ := labkit.SDKClient(t)
	service := sdk.TeamsAdminAPI.AIAgents
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private AI agent recovery journal unavailable")
	}
	cleanup := agentCleanup{service: service}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(ctx, map[string]lab.Handler{"ai-agent": cleanup}); err != nil {
			t.Error("AI agent cleanup incomplete; retained private journal requires recovery")
		}
		if err := journal.Close(); err != nil {
			t.Error("private AI agent journal close failed")
		}
	})
	fixture, err := journal.FixtureName("agent")
	if err != nil {
		t.Fatal("AI agent fixture name unavailable")
	}
	entry, err := journal.PlanCreate("ai-agent", fixture)
	if err != nil {
		t.Fatal("AI agent create was not journaled")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	replacements := tenantReplacements(t, binding)
	replacements[fixture] = "msteams-sdk-fixture-agent"

	request := aiagents.AgentRequest{Name: fixture, AgentID: probeAgent, AgentType: aiagents.AgentTypeMicrosoftCopilotStudio}
	created, response, err := service.Create(ctx, request)
	if err == nil && created != nil && created.ConfigurationID != nil {
		replacements[*created.ConfigurationID] = "00000000-0000-4000-8000-0000000000c1"
	}
	labkit.WriteResponseFixture(t, fixtureService, "create", response, replacements)
	if err != nil || response.StatusCode() != 201 || !ownedAgent(created, fixture) {
		t.Fatal("live AI agent create failed; recovery journal retained")
	}
	id := *created.ConfigurationID
	if err := journal.Applied(entry, id); err != nil {
		t.Fatal("created AI agent ID was not journaled")
	}
	if created.AgentID == nil || *created.AgentID != probeAgent || created.AgentType == nil || *created.AgentType != aiagents.AgentTypeMicrosoftCopilotStudio || created.AgentTargetTagTemplateID != nil {
		t.Fatal("create receipt lost the requested agent fields")
	}
	t.Log("AI agent create: passed")

	value, response, err := service.GetByID(ctx, id)
	if err != nil || !ownedAgent(value, fixture) || *value.ConfigurationID != id {
		t.Fatal("owned AI agent read-back failed")
	}
	labkit.WriteResponseFixture(t, fixtureService, "get", response, replacements)
	listed, response, err := service.List(ctx, nil)
	if err != nil {
		t.Fatal("AI agent list failed")
	}
	found := 0
	for _, item := range listed.Agents {
		if item.ConfigurationID != nil && *item.ConfigurationID == id && ownedAgent(item, fixture) {
			found++
		}
	}
	if found != 1 {
		t.Fatal("owned AI agent missing or ambiguous in list")
	}
	labkit.WriteResponseFixture(t, fixtureService, "list", response, replacements)
	t.Log("AI agent get/list: passed")

	if labkit.OracleEnabled(t) {
		observed := powerShellAgent(t, ctx, id)
		if observed.Name != fixture || observed.AgentID != probeAgent || observed.AgentType != aiagents.AgentTypeMicrosoftCopilotStudio || observed.CatalogCount != 0 {
			t.Fatal("PowerShell and Go AI agent observations differ")
		}
		t.Log("PowerShell 8.1.0 / Go AI agent parity: passed")
	}

	tag := probeTagTemplate
	renamed := aiagents.AgentRequest{Name: fixture + "-renamed", AgentID: probeAgent, AgentType: aiagents.AgentTypeMicrosoftCopilotStudio, AgentTargetTagTemplateID: &tag}
	if _, response, err = service.UpdateByID(ctx, id, renamed); !labkit.IsHTTP(err, 400) {
		t.Fatal("unknown tag template update was not rejected")
	}
	labkit.WriteResponseFixture(t, fixtureService, "update_unknown_tag_template", response, replacements)
	renamed.AgentTargetTagTemplateID = nil
	replacements[fixture+"-renamed"] = "msteams-sdk-fixture-agent-renamed"
	updated, response, err := service.UpdateByID(ctx, id, renamed)
	labkit.WriteResponseFixture(t, fixtureService, "update", response, replacements)
	if err != nil || !ownedAgent(updated, fixture) || *updated.Name != renamed.Name {
		t.Fatal("live AI agent update failed")
	}
	if value, _, err = service.GetByID(ctx, id); err != nil || value.Name == nil || *value.Name != renamed.Name {
		t.Fatal("AI agent rename was not read back")
	}
	t.Log("AI agent update/read-back: passed")

	response, err = service.DeleteByID(ctx, id)
	labkit.WriteResponseFixture(t, fixtureService, "delete", response, replacements)
	if err != nil {
		t.Fatal("live AI agent delete failed")
	}
	if err := cleanup.absent(ctx, lab.Entry{Fixture: fixture, ResourceID: id}); err != nil {
		t.Fatal("AI agent absence was not observed")
	}
	t.Log("AI agent delete/absence: passed")
}

// TestAccAIAgentRecover is an explicit operator-selected recovery path for an
// interrupted run. OpenJournal checks tenant/cloud/identity binding locally.
func TestAccAIAgentRecover(t *testing.T) {
	file := os.Getenv("MSTEAMS_LAB_RECOVERY_FILE")
	if file == "" {
		t.Skip("no explicit private recovery journal selected")
	}
	cfg, err := lab.Load()
	if err != nil || cfg.Require(lab.Mutation) != nil {
		t.Fatal("explicit mutation opt-in is required for AI agent recovery")
	}
	sdk, binding, _ := labkit.SDKClient(t)
	journal, err := lab.OpenJournal(file, binding)
	if err != nil {
		t.Fatal("private bound recovery journal unavailable")
	}
	defer func() {
		if err := journal.Close(); err != nil {
			t.Error("private recovery journal close failed")
		}
	}()
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	if _, err := journal.Recover(ctx, RecoveryHandlers(sdk)); err != nil {
		t.Fatal("AI agent recovery incomplete; private journal retained")
	}
	t.Log("owned AI agent recovery: passed")
}
