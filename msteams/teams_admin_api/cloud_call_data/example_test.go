package cloudcalldata_test

import (
	"context"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

// This example compiles a real SDK workflow. It performs no network operation
// during go test; the application supplies its own auth configuration and context.
// The token is a credential: pass it to the on-premises deployment, never log it.
func ExampleService_Get() {
	token := func(ctx context.Context, cfg *auth.Config) (value string, err error) {
		sdk, err := msteams.NewClient(cfg)
		if err != nil {
			return "", err
		}
		defer func() { err = errors.Join(err, sdk.Close()) }()
		connection, _, err := sdk.TeamsAdminAPI.CloudCallData.Get(ctx)
		if err != nil || connection.Token == nil {
			return "", err
		}
		return *connection.Token, nil
	}
	_ = token
}

// Create sets up the tenant's connection once; the module cannot remove it.
func ExampleService_Create() {
	connect := func(ctx context.Context, sdk *msteams.Client) (*string, error) {
		connection, _, err := sdk.TeamsAdminAPI.CloudCallData.Create(ctx)
		if err != nil {
			return nil, err
		}
		return connection.Token, nil
	}
	_ = connect
}
