package aiagents_test

import (
	"context"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	aiagents "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/ai_agents"
)

// This example compiles a real SDK workflow. It performs no network operation
// during go test; the application supplies its own auth configuration and context.
// The agent ID is the one Microsoft Copilot Studio issued for the agent.
func ExampleService_Create() {
	create := func(ctx context.Context, cfg *auth.Config, copilotAgentID string) (configurationID string, err error) {
		sdk, err := msteams.NewClient(cfg)
		if err != nil {
			return "", err
		}
		defer func() { err = errors.Join(err, sdk.Close()) }()
		agent, _, err := sdk.TeamsAdminAPI.AIAgents.Create(ctx, aiagents.AgentRequest{
			Name:      "Front desk agent",
			AgentID:   copilotAgentID,
			AgentType: aiagents.AgentTypeMicrosoftCopilotStudio,
		})
		if err != nil || agent.ConfigurationID == nil {
			return "", err
		}
		return *agent.ConfigurationID, nil
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
		if current.AgentID == nil || current.AgentType == nil {
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
		for {
			page, _, err := sdk.TeamsAdminAPI.AIAgents.QueryPlatformAgents(ctx, request)
			if err != nil {
				return nil, err
			}
			out = append(out, page.Agents...)
			if page.ContinuationToken == nil || *page.ContinuationToken == "" {
				return out, nil
			}
			request.ContinuationToken = page.ContinuationToken
		}
	}
	_ = all
}
