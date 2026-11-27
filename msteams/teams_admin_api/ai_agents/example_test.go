package aiagents_test

import (
	"context"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	aiagents "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/ai_agents"
	"github.com/go-resty/resty/v2"
)

// This example compiles a real SDK workflow. It performs no network operation
// during go test; the application supplies its own auth configuration and context.
// The agent ID is the one Microsoft Copilot Studio issued for the agent.
func ExampleService_Create() {
	create := func(ctx context.Context, cfg *auth.Config, copilotAgentID string) (agent *aiagents.Agent, response *resty.Response, err error) {
		sdk, err := msteams.NewClient(cfg)
		if err != nil {
			return nil, nil, err
		}
		defer func() { err = errors.Join(err, sdk.Close()) }()
		// Preserve the receipt even when an accepted write has no domain model.
		// Never replay Create merely because the output or ID is absent.
		return sdk.TeamsAdminAPI.AIAgents.Create(ctx, aiagents.AgentRequest{
			Name: "Front desk agent", AgentID: copilotAgentID,
			AgentType: aiagents.AgentTypeMicrosoftCopilotStudio,
		})
	}
	_ = create
}

// UpdateByID replaces the whole configuration, so read it first and send every
// field back, as Set-CsAgent does with the instance it is given.
func ExampleService_UpdateByID() {
	rename := func(ctx context.Context, sdk *msteams.Client, id, name string) error {
		agents := sdk.TeamsAdminAPI.AIAgents
		current, _, err := agents.GetByID(ctx, id)
		if err != nil {
			return err
		}
		if current == nil || current.AgentID == nil || current.AgentType == nil {
			return errors.New("agent configuration is incomplete")
		}
		_, _, err = agents.UpdateByID(ctx, id, aiagents.AgentRequest{
			Name:                     name,
			AgentID:                  *current.AgentID,
			AgentType:                *current.AgentType,
			AgentTargetTagTemplateID: current.AgentTargetTagTemplateID,
		})
		return err
	}
	_ = rename
}

// QueryPlatformAgents pages with the continuation token.
func ExampleService_QueryPlatformAgents() {
	all := func(ctx context.Context, sdk *msteams.Client, providerID string) ([]*aiagents.PlatformAgent, error) {
		var out []*aiagents.PlatformAgent
		request := aiagents.QueryRequest{ProviderID: providerID}
		seen := map[string]bool{}
		for pages := 0; pages < 100; pages++ {
			page, _, err := sdk.TeamsAdminAPI.AIAgents.QueryPlatformAgents(ctx, request)
			if err != nil || page == nil {
				return out, err
			}
			out = append(out, page.Agents...)
			if page.ContinuationToken == nil || *page.ContinuationToken == "" {
				return out, nil
			}
			if seen[*page.ContinuationToken] {
				return out, errors.New("catalog continuation cycle")
			}
			seen[*page.ContinuationToken] = true
			request.ContinuationToken = page.ContinuationToken
		}
		return out, errors.New("catalog page limit reached")
	}
	_ = all
}
