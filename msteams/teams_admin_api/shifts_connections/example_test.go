package shiftsconnections_test

import (
	"context"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	sc "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/shifts_connections"
)

// This example compiles a real SDK workflow. It performs no network operation
// during go test. The Shifts connector service denies app-only Teams
// Administrator credentials (HTTP 403), so the application needs credentials
// the service accepts. Load WFM secrets from a secret store, never source code.
func ExampleService_CreateConnection() {
	connect := func(ctx context.Context, sdk *msteams.Client, connectorID string, settings sc.Settings) (string, error) {
		if _, err := sdk.TeamsAdminAPI.ShiftsConnections.Validate(ctx, sc.ValidateRequest{
			Name: "Store WFM", ConnectorID: connectorID, ConnectorSpecificSettings: settings,
		}); err != nil {
			return "", err
		}
		connection, _, err := sdk.TeamsAdminAPI.ShiftsConnections.CreateConnection(ctx, sc.CreateConnectionRequest{
			Name: "Store WFM", ConnectorID: connectorID, State: "Active", ConnectorSpecificSettings: settings,
		})
		if err != nil {
			return "", err
		}
		if connection.ID == nil {
			return "", errors.New("connection has no ID")
		}
		return *connection.ID, nil
	}
	_ = connect
}

// Instance writes need the current etag in If-Match.
func ExampleService_UpdateInstanceByID() {
	slowDown := func(ctx context.Context, sdk *msteams.Client, id string) error {
		instance, _, err := sdk.TeamsAdminAPI.ShiftsConnections.GetInstanceByID(ctx, id)
		if err != nil {
			return err
		}
		if instance.Etag == nil {
			return errors.New("instance has no etag")
		}
		minutes := int32(30)
		_, _, err = sdk.TeamsAdminAPI.ShiftsConnections.UpdateInstanceByID(ctx, id, sc.InstanceUpdate{
			IfMatch:         *instance.Etag,
			InstanceRequest: sc.InstanceRequest{SyncFrequencyInMin: &minutes},
		})
		return err
	}
	_ = slowDown
}
