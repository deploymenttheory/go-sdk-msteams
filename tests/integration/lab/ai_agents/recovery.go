package aiagentsacc

import (
	"context"
	"regexp"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	aiagents "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/ai_agents"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

var configurationID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

// RecoveryHandlers returns journal recovery handlers for the owned ai_agents
// resources these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"ai-agent": agentCleanup{sdk.TeamsAdminAPI.AIAgents},
	}
}

type agentCleanup struct{ service *aiagents.Service }

// ownedAgentFixture accepts only journal fixture names for agent creates.
func ownedAgentFixture(name string) bool {
	return strings.HasPrefix(name, "msteams-sdk-") && strings.HasSuffix(name, "-agent")
}

// ownedAgent matches the create name or the update name derived from it.
func ownedAgent(value *aiagents.Agent, fixture string) bool {
	return value != nil && value.Name != nil && (*value.Name == fixture || *value.Name == fixture+"-renamed") &&
		value.ConfigurationID != nil && configurationID.MatchString(*value.ConfigurationID)
}

// Resolve reconciles a planned create by its unique agent name.
func (h agentCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	if !ownedAgentFixture(entry.Fixture) {
		return "", lab.ErrUnresolved
	}
	listed, _, err := h.service.List(ctx, nil)
	if err != nil {
		return "", err
	}
	if listed == nil {
		return "", lab.ErrUnresolved
	}
	id := ""
	for _, value := range listed.Agents {
		if !ownedAgent(value, entry.Fixture) {
			continue
		}
		if id != "" {
			return "", lab.ErrUnresolved
		}
		id = *value.ConfigurationID
	}
	if id == "" {
		// An empty list cannot prove a lost create never happened.
		return "", lab.ErrUnresolved
	}
	return id, nil
}

// Delete removes the owned agent after confirming its name, then observes 404.
func (h agentCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !ownedAgentFixture(entry.Fixture) || !configurationID.MatchString(entry.ResourceID) {
		return lab.ErrUnresolved
	}
	value, _, err := h.service.GetByID(ctx, entry.ResourceID)
	if labkit.IsHTTP(err, 404) {
		return nil
	}
	if err != nil {
		return err
	}
	if !ownedAgent(value, entry.Fixture) {
		return lab.ErrUnresolved
	}
	if _, err := h.service.DeleteByID(ctx, entry.ResourceID); err != nil && !labkit.IsHTTP(err, 404) {
		return err
	}
	return h.absent(ctx, entry)
}

func (h agentCleanup) absent(ctx context.Context, entry lab.Entry) error {
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		_, _, err := h.service.GetByID(ctx, entry.ResourceID)
		if labkit.IsHTTP(err, 404) {
			return true, nil
		}
		return false, err
	})
}

func (agentCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }
