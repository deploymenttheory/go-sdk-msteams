package resourceaccounts_test

import (
	"context"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	ra "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/resource_accounts"
)

// This example compiles a real SDK workflow. It performs no network operation
// during go test; the application supplies its own auth configuration and context.
// Creating a resource account needs directory write rights, which app-only
// Teams Administrator credentials do not have.
func ExampleService_Create() {
	create := func(ctx context.Context, cfg *auth.Config, upn string) (objectID string, err error) {
		sdk, err := msteams.NewClient(cfg)
		if err != nil {
			return "", err
		}
		defer func() { err = errors.Join(err, sdk.Close()) }()
		app, name := ra.CallQueueApplicationID, "Support line"
		account, _, err := sdk.TeamsAdminAPI.ResourceAccounts.Create(ctx, ra.CreateRequest{
			UserPrincipalName: upn,
			ApplicationID:     &app,
			DisplayName:       &name,
		})
		if err != nil || account.ObjectID == nil {
			return "", err
		}
		return *account.ObjectID, nil
	}
	_ = create
}

// CreateAssociations reports per-account failures in each StatusRecord.
func ExampleService_CreateAssociations() {
	associate := func(ctx context.Context, sdk *msteams.Client, accountID, callQueueID string) error {
		results, _, err := sdk.TeamsAdminAPI.ResourceAccounts.CreateAssociations(ctx, ra.AssociationRequest{
			EndpointIDs:       []string{accountID},
			ConfigurationID:   callQueueID,
			ConfigurationType: ra.ConfigurationTypeCallQueue,
		})
		if err != nil {
			return err
		}
		for _, result := range results {
			if result.Status.Failed() {
				return errors.New(*result.Status.ErrorCode)
			}
		}
		return nil
	}
	_ = associate
}
