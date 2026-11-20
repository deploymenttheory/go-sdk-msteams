package applicationaccesspolicy_test

import (
	"context"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	policy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/application_access_policy"
)

func ExampleService_Create() {
	// Construction is local. Supply real credentials explicitly when running this
	// workflow in an approved tenant; never embed credentials in source.
	workflow := func(ctx context.Context, cfg *auth.Config) (err error) {
		sdk, err := msteams.NewClient(cfg)
		if err != nil {
			return err
		}
		defer func() { err = errors.Join(err, sdk.Close()) }()
		description := "Application access policy managed by the SDK"
		_, _, err = sdk.TeamsAdminAPI.ApplicationAccessPolicies.Create(ctx, policy.CreateRequest{
			Identity: "sdk-example-policy", AppIDs: policy.ReplaceAppIDs([]string{"application-client-id"}), Description: &description,
		})
		return err
	}
	_ = workflow // Defining the example does not connect to a tenant.
}
