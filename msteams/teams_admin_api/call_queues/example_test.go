package callqueues_test

import (
	"context"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	cq "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/call_queues"
)

// This example compiles a real SDK workflow. It performs no network operation
// during go test; the application supplies its own auth configuration and context.
func ExampleService_Create() {
	create := func(ctx context.Context, cfg *auth.Config) (id string, err error) {
		sdk, err := msteams.NewClient(cfg)
		if err != nil {
			return "", err
		}
		defer func() { err = errors.Join(err, sdk.Close()) }()
		name, routing, alert := "Support", cq.RoutingLongestIdle, int32(30)
		queue, _, err := sdk.TeamsAdminAPI.CallQueues.Create(ctx, cq.Settings{
			Name:           &name,
			RoutingMethod:  &routing,
			AgentAlertTime: &alert,
		}, nil)
		if err != nil || queue.Identity == nil {
			return "", err
		}
		return *queue.Identity, nil
	}
	_ = create
}

// The service replaces the whole queue, so update from a fresh read as
// Set-CsCallQueue does.
func ExampleCallQueue_Settings() {
	update := func(ctx context.Context, sdk *msteams.Client, id string) error {
		s := sdk.TeamsAdminAPI.CallQueues
		queue, _, err := s.GetByID(ctx, id)
		if err != nil {
			return err
		}
		settings := queue.Settings()
		optOut := true
		settings.AllowOptOut = &optOut
		_, _, err = s.UpdateByID(ctx, id, settings, nil)
		return err
	}
	_ = update
}
