package appinstallation_test

import (
	"context"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	ai "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/app_installation"
)

// The application supplies an authorized client and owned target. This example
// compiles without executing the closure or installing an app during go test.
func ExampleService_Create() {
	create := func(ctx context.Context, sdk *msteams.Client, teamID, catalogID string) error {
		_, err := sdk.GraphAPI.AppInstallations.Create(ctx, ai.Target{TeamID: teamID}, ai.CreateRequest{CatalogID: catalogID})
		// A successful empty 201 response contains no installation ID. A lost
		// response requires reconciliation by catalog ID before another create.
		return err
	}
	_ = create
}

// List preserves partial items when a later Graph page fails.
func ExampleService_List() {
	read := func(ctx context.Context, sdk *msteams.Client, userID string) ([]ai.Installation, error) {
		result, err := sdk.GraphAPI.AppInstallations.List(ctx, ai.Target{UserID: userID}, ai.ListOptions{})
		return result.Items, err
	}
	_ = read
}

// Resolution distinguishes the catalog ID from the installed resource ID and
// refuses to remove anything when the mapping is absent or ambiguous.
func ExampleService_DeleteByCatalogID() {
	remove := func(ctx context.Context, sdk *msteams.Client, teamID, catalogID string) error {
		_, err := sdk.GraphAPI.AppInstallations.DeleteByCatalogID(ctx, ai.Target{TeamID: teamID}, catalogID)
		if errors.Is(err, ai.ErrNotFound) {
			return nil
		}
		return err
	}
	_ = remove
}

// Consent is native Graph v1.0 data. This does not assume equivalence with the
// legacy Permissions property emitted by the pinned preview cmdlet.
func ExampleService_UpgradeByID() {
	upgrade := func(ctx context.Context, sdk *msteams.Client, teamID, installationID string) error {
		consent := &ai.PermissionSet{ResourceSpecificPermissions: []ai.ResourcePermission{{PermissionValue: "ChannelMessage.Read.Group", PermissionType: ai.PermissionApplication}}}
		_, err := sdk.GraphAPI.AppInstallations.UpgradeByID(ctx, ai.Target{TeamID: teamID}, installationID, consent)
		return err
	}
	_ = upgrade
}
