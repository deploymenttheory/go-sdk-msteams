package extintegration

import (
	"context"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	agents "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/ai_agents"
	templates "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/call_templates"
	agentsacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/ai_agents"
	templatesacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/call_templates"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

type agentObservation struct {
	Success         bool   `json:"success"`
	AgentName       string `json:"agent_name"`
	TemplateName    string `json:"template_name"`
	TemplateID      string `json:"template_id"`
	AgentTemplateID string `json:"agent_template_id"`
	TagCount        int    `json:"tag_count"`
}

func compareAgent(t *testing.T, ctx context.Context, agentID, agentName, templateID, templateName string) {
	t.Helper()
	var observed agentObservation
	oracle(t, ctx, "read-ext-agent-tags.ps1", []string{"MSTEAMS_LAB_AGENT_ID=" + agentID, "MSTEAMS_LAB_AGENT_FIXTURE=" + agentName, "MSTEAMS_LAB_TAG_TEMPLATE_ID=" + templateID, "MSTEAMS_LAB_TAG_TEMPLATE_FIXTURE=" + templateName}, &observed)
	if !observed.Success || observed.AgentName != agentName || observed.TemplateName != templateName || observed.TemplateID != templateID || observed.AgentTemplateID != templateID || observed.TagCount != 1 {
		t.Fatal("PowerShell and Go agent/template linkage differs")
	}
	t.Log("PowerShell 8.1.0 / Go EXT agent-template parity: passed")
}

// TestAccAgentTargetTemplateComposition creates a tag template and an agent
// configuration referring to it. The agent is not attached to an account or
// call flow, so no call can reach its reserved external target/provider.
func TestAccAgentTargetTemplateComposition(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Require(lab.Mutation); errors.Is(err, lab.ErrDisabled) {
		t.Skip("EXT agent-template integration needs explicit mutation opt-in and private journaling")
	} else if err != nil {
		t.Fatal(err)
	}
	if !labkit.OracleEnabled(t) {
		t.Fatal("EXT agent-template integration requires the pinned oracle")
	}
	sdk, binding, _ := labkit.SDKClient(t)
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private agent-template journal unavailable")
	}
	handlers := agentsacc.RecoveryHandlers(sdk)
	for kind, handler := range templatesacc.RecoveryHandlers(sdk) {
		handlers[kind] = handler
	}
	t.Cleanup(func() {
		ctx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		results, err := journal.Recover(ctx, handlers)
		if err != nil {
			t.Error("EXT agent-template cleanup incomplete; retained journal needs recovery")
		}
		for _, result := range results {
			if result.Outcome != "cleaned" {
				t.Error("EXT agent-template cleanup unresolved")
			}
		}
		if err := journal.Close(); err != nil {
			t.Error("private agent-template journal close failed")
		}
		if !t.Failed() {
			t.Log("EXT agent-template independent cleanup: passed")
		}
	})
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	templateName, err := journal.FixtureName("tags")
	if err != nil {
		t.Fatal("owned tag template name unavailable")
	}
	templateEntry, err := journal.PlanCreate("tags-template", templateName)
	if err != nil {
		t.Fatal("tag template create was not journaled")
	}
	template, response, err := sdk.TeamsAdminAPI.CallTemplates.Tags.Create(ctx, &templates.TagsTemplate{Name: &templateName, Description: ptr("Owned EXT agent target template"), Tags: []templates.Tag{{TagName: ptr("support"), TagDetails: &templates.TagTarget{ID: ptr("tel:+15550001111"), Type: ptr("ExternalPstn")}}}})
	if err == nil && template != nil && template.ID != nil {
		err = journal.Applied(templateEntry, *template.ID)
	}
	if err != nil || template == nil || template.ID == nil {
		t.Fatal("owned agent target template create failed; journal retained")
	}
	replacements := map[string]string{templateName: "msteams-sdk-fixture-tags", *template.ID: "00000000-0000-4000-8000-000000000001"}
	labkit.WriteResponseFixture(t, "call_templates", "ext_integration_create", response, replacements)
	agentName, err := journal.FixtureName("agent")
	if err != nil {
		t.Fatal("owned agent name unavailable")
	}
	agentEntry, err := journal.PlanCreate("ai-agent", agentName)
	if err != nil {
		t.Fatal("agent create was not journaled")
	}
	request := agents.AgentRequest{Name: agentName, AgentID: "00000000-0000-4000-8000-000000000003", AgentType: agents.AgentTypeMicrosoftCopilotStudio, AgentTargetTagTemplateID: template.ID}
	agent, response, err := sdk.TeamsAdminAPI.AIAgents.Create(ctx, request)
	if err == nil && agent != nil && agent.ConfigurationID != nil {
		err = journal.Applied(agentEntry, *agent.ConfigurationID)
	}
	if err != nil || agent == nil || agent.ConfigurationID == nil {
		t.Fatal("owned agent/template composition create failed; journal retained")
	}
	replacements[agentName] = "msteams-sdk-fixture-agent"
	replacements[*agent.ConfigurationID] = "00000000-0000-4000-8000-000000000002"
	labkit.WriteResponseFixture(t, "ai_agents", "ext_integration_create", response, replacements)
	check := func(expectedName string) {
		value, response, err := sdk.TeamsAdminAPI.AIAgents.GetByID(ctx, *agent.ConfigurationID)
		if err != nil || value == nil || value.Name == nil || *value.Name != expectedName || value.AgentTargetTagTemplateID == nil || *value.AgentTargetTagTemplateID != *template.ID {
			t.Fatal("agent lost its owned tag template reference")
		}
		read, _, err := sdk.TeamsAdminAPI.CallTemplates.Tags.GetByID(ctx, *template.ID)
		if err != nil || read == nil || read.Name == nil || *read.Name != templateName || len(read.Tags) != 1 {
			t.Fatal("agent target template read-back differs")
		}
		labkit.WriteResponseFixture(t, "ai_agents", "ext_integration_read", response, replacements)
		compareAgent(t, ctx, *agent.ConfigurationID, expectedName, *template.ID, templateName)
	}
	check(agentName)
	t.Log("EXT agent-template create/read-back: passed")
	request.Name = agentName + "-renamed"
	replacements[request.Name] = "msteams-sdk-fixture-agent-renamed"
	if _, _, err := sdk.TeamsAdminAPI.AIAgents.UpdateByID(ctx, *agent.ConfigurationID, request); err != nil {
		t.Fatal("owned agent/template replacement failed")
	}
	check(request.Name)
	t.Log("EXT agent-template update/read-back: passed")
}
