package extintegration

import (
	"context"
	"encoding/json"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	agents "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/ai_agents"
	templates "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/call_templates"
	cd "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/configuration_diagnostics"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// Exact response provenance: docs/implementation/ext/evidence/response-fixtures-25ad385.json.
// The harness source is the mixed-response test vector; each body is loaded unchanged.
const captureSource = "tests/integration/lab/ext_integration/captured_test.go"

func captured(t *testing.T, service, name string) []byte {
	t.Helper()
	root, err := os.OpenRoot(labkit.Root + "/msteams/teams_admin_api")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	raw, err := root.ReadFile(service + "/testdata/" + name)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func runCaptured(t *testing.T, name string, steps []contracttest.Exchange, invoke func(context.Context, *msteams.Client) contracttest.Observation) {
	t.Helper()
	result, err := contracttest.Run(t.Context(), contracttest.Case{Name: name, Family: "ext_integration", Source: captureSource, Scope: "service", Steps: append([]contracttest.Exchange{discovery(t)}, steps...), Invoke: invoke})
	if err != nil || !result.Passed() {
		t.Fatalf("captured EXT response replay differed: %v %+v", err, result.Differences())
	}
}

// Response bytes are unchanged sanitized captures. Request selectors are
// controlled surrogates; the original private requests are not reconstructed.
func TestUnitCapturedGenericConfigurationPreservesIdentityAndSettings(t *testing.T) {
	t.Parallel()
	raw := captured(t, "configuration_diagnostics", "integration_generic_get.json")
	var observed cd.Configuration
	if json.Unmarshal(raw, &observed) != nil || observed.Identity == nil {
		t.Fatal("captured generic identity missing")
	}
	identity := *observed.Identity
	receipt := captured(t, "configuration_diagnostics", "integration_create_201.body")
	if len(receipt) != 0 {
		t.Fatal("captured creation receipt was not empty")
	}
	runCaptured(t, "captured-generic-create-read", []contracttest.Exchange{
		exchange(false, "POST", configPath, "New-CsConfiguration", `{"Identity":"`+identity+`","UserInitiatedMode":"Disabled","AllowLogCollection":false}`, 201, string(receipt)),
		exchange(false, "GET", configPath+"/configuration/"+identity, "Get-CsConfiguration", "", 200, string(raw)),
	}, func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		generic := sdk.TeamsAdminAPI.ConfigurationDiagnostics
		if _, err := generic.CreateConfiguration(ctx, configType, cd.ConfigurationRequest{Identity: identity, Settings: map[string]any{"UserInitiatedMode": "Disabled", "AllowLogCollection": false}}, nil); err != nil {
			return contracttest.Observation{Err: err}
		}
		value, _, err := generic.GetConfiguration(ctx, configType, identity, nil)
		if err == nil && (value == nil || value.Identity == nil || *value.Identity != identity || value.ConfigID == nil || observed.ConfigID == nil || *value.ConfigID != *observed.ConfigID || string(value.RawFields["AllowLogCollection"]) != "false" || value.RawFields["Key"] == nil || value.RawFields["ConfigMetadata"] == nil) {
			t.Fatal("captured generic result lost identity, false or metadata")
		}
		return contracttest.Observation{Err: err}
	})
}

func TestUnitCapturedTypedFeedbackPreservesFullResponseSnapshots(t *testing.T) {
	t.Parallel()
	for _, fixture := range []struct {
		name, mode string
		logs       bool
	}{{"integration_typed_get.json", "Disabled", true}, {"integration_replaced_get.json", "Enabled", false}} {
		t.Run(fixture.name, func(t *testing.T) {
			raw := captured(t, "teams_feedback_policy", fixture.name)
			var body struct{ Identity string }
			if json.Unmarshal(raw, &body) != nil || body.Identity == "" {
				t.Fatal("captured typed identity missing")
			}
			// Direct compiled PolicyRP needs no generated-client discovery.
			result, err := contracttest.Run(t.Context(), contracttest.Case{Name: "captured-feedback-" + strings.ToLower(fixture.mode), Family: "ext_integration", Scope: "service", Source: captureSource, Steps: []contracttest.Exchange{exchange(true, "GET", configPath+"/configuration/"+body.Identity, "Get-CsTeamsFeedbackPolicy", "", 200, string(raw))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, _, err := sdk.TeamsAdminAPI.TeamsFeedbackPolicies.GetByIdentity(ctx, body.Identity)
				if err != nil {
					return contracttest.Observation{Err: err}
				}
				if value == nil || value.Identity == nil || *value.Identity != body.Identity || value.UserInitiatedMode == nil || *value.UserInitiatedMode != fixture.mode || value.AllowLogCollection == nil || *value.AllowLogCollection != fixture.logs || value.EnableFeatureSuggestions == nil || *value.EnableFeatureSuggestions || value.RawFields["Key"] == nil || value.RawFields["ConfigId"] == nil {
					t.Fatal("captured typed setting/presence projection changed")
				}
				encoded, err := json.Marshal(value)
				var roundTrip, original any
				if err != nil || json.Unmarshal(encoded, &roundTrip) != nil || json.Unmarshal(raw, &original) != nil || !reflect.DeepEqual(roundTrip, original) {
					t.Fatal("captured typed response snapshot lost metadata")
				}
				return contracttest.Observation{}
			}})
			if err != nil || !result.Passed() {
				t.Fatalf("captured typed feedback replay differed: %v %+v", err, result.Differences())
			}
		})
	}
}

func TestUnitCapturedTemplateAgentHandoffPreservesNestedDefaults(t *testing.T) {
	t.Parallel()
	tags := captured(t, "call_templates", "ext_integration_create_201.json")
	create := captured(t, "ai_agents", "ext_integration_create_201.json")
	read := captured(t, "ai_agents", "ext_integration_read.json")
	const templateID = "00000000-0000-4000-8000-000000000001"
	const agentID = "00000000-0000-4000-8000-000000000002"
	const providerID = "00000000-0000-4000-8000-000000000003"
	runCaptured(t, "captured-template-agent-handoff", []contracttest.Exchange{
		exchange(false, "POST", "/Teams.VoiceApps/ivr-tags-template", "New-CsTagsTemplate", `{"Name":"Fixture name","Description":"Fixture description","Tags":[{"TagName":"support","TagDetails":{"Id":"tel:+15550001111","Type":"ExternalPstn"}}]}`, 201, string(tags)),
		exchange(false, "POST", "/Teams.VoiceApps/aiagents", "New-CsAgent", `{"Name":"Fixture name","AgentId":"`+providerID+`","AgentType":"MicrosoftCopilotStudio","AgentTargetTagTemplateId":"`+templateID+`"}`, 201, string(create)),
		exchange(false, "GET", "/Teams.VoiceApps/aiagents/"+agentID, "Get-CsAgent", "", 200, string(read)),
	}, func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		value, _, err := sdk.TeamsAdminAPI.CallTemplates.Tags.Create(ctx, &templates.TagsTemplate{Name: ptr("Fixture name"), Description: ptr("Fixture description"), Tags: []templates.Tag{{TagName: ptr("support"), TagDetails: &templates.TagTarget{ID: ptr("tel:+15550001111"), Type: ptr("ExternalPstn")}}}})
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		if value == nil || value.ID == nil || *value.ID != templateID || len(value.Tags) != 1 || value.Tags[0].TagDetails == nil {
			t.Fatal("captured template identity or nested target missing")
		}
		target := value.Tags[0].TagDetails
		if target.EnableTranscription == nil || *target.EnableTranscription || target.EnableSharedVoicemailSystemPromptSuppression == nil || *target.EnableSharedVoicemailSystemPromptSuppression || string(target.RawFields["DisplayName"]) != "null" {
			t.Fatal("captured nested false/null target fields were lost")
		}
		agent, _, err := sdk.TeamsAdminAPI.AIAgents.Create(ctx, agents.AgentRequest{Name: "Fixture name", AgentID: providerID, AgentType: agents.AgentTypeMicrosoftCopilotStudio, AgentTargetTagTemplateID: value.ID})
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		if agent == nil || agent.ConfigurationID == nil || *agent.ConfigurationID != agentID || agent.AgentTargetTagTemplateID == nil || *agent.AgentTargetTagTemplateID != *value.ID {
			t.Fatal("captured agent lost created template reference")
		}
		observed, _, err := sdk.TeamsAdminAPI.AIAgents.GetByID(ctx, *agent.ConfigurationID)
		if err == nil && (observed == nil || observed.AgentTargetTagTemplateID == nil || *observed.AgentTargetTagTemplateID != *value.ID || observed.AgentID == nil || *observed.AgentID != providerID) {
			t.Fatal("captured agent read lost provider or template identity")
		}
		return contracttest.Observation{Err: err}
	})
}
