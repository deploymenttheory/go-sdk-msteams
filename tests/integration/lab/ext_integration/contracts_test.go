// Package extintegration checks compositions across extension services and the
// typed resources they share. Synthetic replies do not claim tenant success.
package extintegration

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	agents "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/ai_agents"
	templates "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/call_templates"
	cd "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/configuration_diagnostics"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	feedback "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_feedback_policy"
	templatesacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/call_templates"
	diagnosticsacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/configuration_diagnostics"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const source = "tests/integration/lab/ext_integration/contracts_test.go"
const configType = "TeamsFeedbackPolicy"
const fixture = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-config"
const configPath = "/Skype.Policy/configurations/TeamsFeedbackPolicy"
const taggedPath = configPath + "/configuration/Tag%3A" + fixture

func discovery(t *testing.T) contracttest.Exchange {
	t.Helper()
	raw, err := os.ReadFile(labkit.Root + "/msteams/teams_admin_api/internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: raw}}
}

func exchange(compiled bool, method, path, command, body string, status int, reply string) contracttest.Exchange {
	// AI response fixtures remain observed bytes; these request headers follow
	// the independently traced current generated helper defaults.
	switch command {
	case "New-CsAgent":
		command = "New-CsAgent_New"
	case "Get-CsAgent":
		if strings.Contains(path, "/aiagents/") {
			command = "Get-CsAgent_Get1"
		} else {
			command = "Get-CsAgent_Get"
		}
	case "Set-CsAgent":
		command = "Set-CsAgent_SetExpanded"
	case "Remove-CsAgent":
		command = "Remove-CsAgent_Remove"
	}
	headers := http.Header{"X-MS-CmdletName": {command}}
	if compiled {
		headers.Set("MPACmdlet", "true")
		headers[http.CanonicalHeaderKey("X-MS-Target-Uri")] = nil
	} else {
		// The generated ConfigAPI adapter uses standard segment escaping; the
		// compiled PolicyRP transport additionally escapes the identity colon.
		path = strings.ReplaceAll(path, "Tag%3A", "Tag:")
		headers.Set("Accept", "application/json")
		headers.Set("X-MS-Target-Uri", "https://admin.example.invalid/")
		headers[http.CanonicalHeaderKey("MPACmdlet")] = nil
	}
	var input json.RawMessage
	if body != "" {
		input = json.RawMessage(body)
		contentType := "application/json"
		if compiled {
			contentType += "; charset=utf-8"
		}
		headers.Set("Content-Type", contentType)
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: headers, JSON: input}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}

func run(t *testing.T, name string, steps []contracttest.Exchange, want error, invoke func(context.Context, *msteams.Client) contracttest.Observation) {
	t.Helper()
	result, err := contracttest.Run(t.Context(), contracttest.Case{Name: name, Family: "ext_integration", Source: source, Scope: "service", Steps: append([]contracttest.Exchange{discovery(t)}, steps...), Want: contracttest.Expected{Error: want}, Invoke: invoke})
	if err != nil || !result.Passed() {
		t.Fatalf("EXT composition differed: %v %+v", err, result.Differences())
	}
}

func TestUnitGenericAndTypedConfigurationShareIdentityAndPresence(t *testing.T) {
	t.Parallel()
	initial := `{"Identity":"Tag:` + fixture + `","UserInitiatedMode":"Disabled","ReceiveSurveysMode":null,"AllowLogCollection":true,"EnableFeatureSuggestions":false,"FutureField":{"Items":[]}}`
	patched := `{"Identity":"Tag:` + fixture + `","UserInitiatedMode":"Disabled","ReceiveSurveysMode":null,"AllowLogCollection":false,"EnableFeatureSuggestions":false,"FutureField":{"Items":[]}}`
	replaced := `{"Identity":"Tag:` + fixture + `","UserInitiatedMode":"Enabled","ReceiveSurveysMode":"","AllowLogCollection":false}`
	run(t, "generic-typed-configuration", []contracttest.Exchange{
		exchange(false, "POST", configPath, "New-CsConfiguration", `{"Identity":"`+fixture+`","UserInitiatedMode":"Disabled","AllowLogCollection":true}`, 201, ""),
		exchange(true, "GET", taggedPath, "Get-CsTeamsFeedbackPolicy", "", 200, initial),
		exchange(true, "PATCH", taggedPath, "Set-CsTeamsFeedbackPolicy", `{"Identity":"Tag:`+fixture+`","AllowLogCollection":false}`, 204, ""),
		exchange(false, "GET", taggedPath, "Get-CsConfiguration", "", 200, patched),
		exchange(false, "PUT", taggedPath, "Set-CsConfiguration", replaced, 204, ""),
		exchange(true, "GET", taggedPath, "Get-CsTeamsFeedbackPolicy", "", 200, replaced),
	}, nil, func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		generic, typed := sdk.TeamsAdminAPI.ConfigurationDiagnostics, sdk.TeamsAdminAPI.TeamsFeedbackPolicies
		if _, err := generic.CreateConfiguration(ctx, configType, cd.ConfigurationRequest{Identity: fixture, Settings: map[string]any{"UserInitiatedMode": "Disabled", "AllowLogCollection": true}}, nil); err != nil {
			return contracttest.Observation{Err: err}
		}
		value, _, err := typed.GetByIdentity(ctx, "Tag:"+fixture)
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		if value.Identity == nil || *value.Identity != "Tag:"+fixture || value.AllowLogCollection == nil || !*value.AllowLogCollection || value.EnableFeatureSuggestions == nil || *value.EnableFeatureSuggestions || value.RawFields["FutureField"] == nil || string(value.RawFields["ReceiveSurveysMode"]) != "null" {
			t.Fatal("generic-to-typed identity/presence was lost")
		}
		if _, _, err := typed.PatchByIdentity(ctx, *value.Identity, feedback.PatchRequest{Settings: feedback.Settings{AllowLogCollection: policy.Set(false)}}); err != nil {
			return contracttest.Observation{Err: err}
		}
		raw, _, err := generic.GetConfiguration(ctx, configType, *value.Identity, nil)
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		if string(raw.RawFields["AllowLogCollection"]) != "false" || raw.RawFields["FutureField"] == nil {
			t.Fatal("typed patch did not preserve generic omitted fields")
		}
		if _, err := generic.SetConfiguration(ctx, configType, *value.Identity, cd.ConfigurationRequest{Identity: *value.Identity, Settings: map[string]any{"UserInitiatedMode": "Enabled", "ReceiveSurveysMode": "", "AllowLogCollection": false}}, nil); err != nil {
			return contracttest.Observation{Err: err}
		}
		last, _, err := typed.GetByIdentity(ctx, *value.Identity)
		if err == nil && (last.ReceiveSurveysMode == nil || *last.ReceiveSurveysMode != "" || last.AllowLogCollection == nil || *last.AllowLogCollection) {
			t.Fatal("generic replacement lost explicit empty/false at typed read")
		}
		return contracttest.Observation{Err: err}
	})
}

func TestUnitTypedPreparationDenialRetainsErrorAndGenericCleanup(t *testing.T) {
	t.Parallel()
	run(t, "typed-denial-generic-recovery", []contracttest.Exchange{
		exchange(false, "POST", configPath, "New-CsConfiguration", `{"Identity":"`+fixture+`","UserInitiatedMode":"Disabled"}`, 201, ""),
		exchange(true, "GET", taggedPath, "Set-CsTeamsFeedbackPolicy", "", 403, `{"code":"Forbidden"}`),
		exchange(false, "DELETE", taggedPath, "Remove-CsConfiguration", "", 204, ""),
		exchange(false, "GET", taggedPath, "Get-CsConfiguration", "", 404, `{"code":"NotFound"}`),
		exchange(false, "GET", configPath, "Get-CsConfiguration", "", 200, `[]`),
	}, auth.ErrForbidden, func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		if _, err := sdk.TeamsAdminAPI.ConfigurationDiagnostics.CreateConfiguration(ctx, configType, cd.ConfigurationRequest{Identity: fixture, Settings: map[string]any{"UserInitiatedMode": "Disabled"}}, nil); err != nil {
			return contracttest.Observation{Err: err}
		}
		result, response, operationErr := sdk.TeamsAdminAPI.TeamsFeedbackPolicies.UpdateByIdentity(ctx, "Tag:"+fixture, feedback.UpdateRequest{Settings: feedback.Settings{AllowLogCollection: policy.Set(false)}})
		if !errors.Is(operationErr, auth.ErrForbidden) || response == nil || response.StatusCode() != 403 || result == nil || result.PatchInvoked || result.HTTPAccepted {
			t.Fatal("denied typed preparation lost error or sent a mutation")
		}
		handler := diagnosticsacc.RecoveryHandlers(sdk)["generic-configuration"]
		if err := handler.Delete(context.WithoutCancel(ctx), lab.Entry{Fixture: fixture, ResourceID: "Tag:" + fixture}); err != nil {
			t.Fatal("independent generic cleanup did not prove absence")
		}
		return contracttest.Observation{Err: operationErr}
	})
}

func TestUnitAgentRetainsCreatedTagTemplateIdentity(t *testing.T) {
	t.Parallel()
	const templateID = "00000000-0000-4000-8000-000000000001"
	const agentID = "00000000-0000-4000-8000-000000000002"
	const providerID = "00000000-0000-4000-8000-000000000003"
	tags := `{"Name":"owned-tags","Description":"owned targets","Tags":[{"TagName":"support","TagDetails":{"Id":"tel:+15550001111","Type":"ExternalPstn"}}]}`
	run(t, "agent-tag-template-identity", []contracttest.Exchange{
		exchange(false, "POST", "/Teams.VoiceApps/ivr-tags-template", "New-CsTagsTemplate", tags, 201, `{"IvrTagsTemplate":{"Id":"`+templateID+`","Name":"owned-tags","Description":"owned targets","Tags":[{"TagName":"support","TagDetails":{"Id":"tel:+15550001111","Type":"ExternalPstn"}}]}}`),
		exchange(false, "POST", "/Teams.VoiceApps/aiagents", "New-CsAgent", `{"Name":"owned-agent","AgentId":"`+providerID+`","AgentType":"MicrosoftCopilotStudio","AgentTargetTagTemplateId":"`+templateID+`"}`, 201, `{"AIAgent":{"ConfigurationId":"`+agentID+`","Name":"owned-agent","AgentId":"`+providerID+`","AgentType":"MicrosoftCopilotStudio","AgentTargetTagTemplateId":"`+templateID+`"}}`),
	}, nil, func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		template, _, err := sdk.TeamsAdminAPI.CallTemplates.Tags.Create(ctx, &templates.TagsTemplate{Name: ptr("owned-tags"), Description: ptr("owned targets"), Tags: []templates.Tag{{TagName: ptr("support"), TagDetails: &templates.TagTarget{ID: ptr("tel:+15550001111"), Type: ptr("ExternalPstn")}}}})
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		agent, _, err := sdk.TeamsAdminAPI.AIAgents.Create(ctx, agents.AgentRequest{Name: "owned-agent", AgentID: providerID, AgentType: agents.AgentTypeMicrosoftCopilotStudio, AgentTargetTagTemplateID: template.ID})
		if err == nil && (agent.AgentTargetTagTemplateID == nil || *agent.AgentTargetTagTemplateID != *template.ID || agent.ConfigurationID == nil || *agent.ConfigurationID != agentID) {
			t.Fatal("agent lost the created target template identity")
		}
		return contracttest.Observation{Err: err}
	})
}

// A later service's authorization failure must not prevent independent cleanup
// of the already created template, or replace the original operation error.
func TestUnitAgentDenialRetainsErrorAndIndependentlyCleansTemplate(t *testing.T) {
	t.Parallel()
	const templateID = "00000000-0000-4000-8000-000000000001"
	const providerID = "00000000-0000-4000-8000-000000000003"
	const name = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-tags"
	template := `{"IvrTagsTemplate":{"Id":"` + templateID + `","Name":"` + name + `","Description":"owned targets","Tags":[{"TagName":"support","TagDetails":{"Id":"tel:+15550001111","Type":"ExternalPstn"}}]}}`
	path := "/Teams.VoiceApps/ivr-tags-template/" + templateID
	run(t, "agent-denial-template-recovery", []contracttest.Exchange{
		exchange(false, "POST", "/Teams.VoiceApps/ivr-tags-template", "New-CsTagsTemplate", `{"Name":"`+name+`","Description":"owned targets","Tags":[{"TagName":"support","TagDetails":{"Id":"tel:+15550001111","Type":"ExternalPstn"}}]}`, 201, template),
		exchange(false, "POST", "/Teams.VoiceApps/aiagents", "New-CsAgent", `{"Name":"owned-agent","AgentId":"`+providerID+`","AgentType":"MicrosoftCopilotStudio","AgentTargetTagTemplateId":"`+templateID+`"}`, 403, `{"code":"Forbidden"}`),
		exchange(false, "GET", path, "Get-CsTagsTemplate", "", 200, template),
		exchange(false, "DELETE", path, "Remove-CsTagsTemplate", "", 204, ""),
		exchange(false, "GET", path, "Get-CsTagsTemplate", "", 404, `{"code":"100004"}`),
	}, auth.ErrForbidden, func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		value, _, err := sdk.TeamsAdminAPI.CallTemplates.Tags.Create(ctx, &templates.TagsTemplate{Name: ptr(name), Description: ptr("owned targets"), Tags: []templates.Tag{{TagName: ptr("support"), TagDetails: &templates.TagTarget{ID: ptr("tel:+15550001111"), Type: ptr("ExternalPstn")}}}})
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		_, response, operationErr := sdk.TeamsAdminAPI.AIAgents.Create(ctx, agents.AgentRequest{Name: "owned-agent", AgentID: providerID, AgentType: agents.AgentTypeMicrosoftCopilotStudio, AgentTargetTagTemplateID: value.ID})
		if !errors.Is(operationErr, auth.ErrForbidden) || response == nil || response.StatusCode() != 403 {
			t.Fatal("agent authorization failure lost its operation outcome")
		}
		handler := templatesacc.RecoveryHandlers(sdk)["tags-template"]
		if err := handler.Delete(context.WithoutCancel(ctx), lab.Entry{Fixture: name, ResourceID: *value.ID}); err != nil {
			t.Fatal("independent target template cleanup did not prove absence")
		}
		return contracttest.Observation{Err: operationErr}
	})
}

func ptr[T any](value T) *T { return &value }
