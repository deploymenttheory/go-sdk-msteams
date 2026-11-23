package appcatalog_test

import (
	"context"
	"os"
	"path/filepath"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	appcatalog "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/app_catalog"
)

// This example compiles a real SDK workflow. It performs no network operation
// during go test; the application supplies its own client and context.
func ExampleService_List() {
	organizationApps := func(ctx context.Context, sdk *msteams.Client) ([]appcatalog.TeamsApp, error) {
		result, err := sdk.GraphAPI.AppCatalog.List(ctx, appcatalog.ListOptions{DistributionMethod: appcatalog.OrganizationDistribution})
		return result.Items, err
	}
	_ = organizationApps
}

// Create uploads an app package (.zip) to the organization catalog; the
// returned ID addresses later UpdateByID and DeleteByID calls.
func ExampleService_Create() {
	publish := func(ctx context.Context, sdk *msteams.Client, path string) (string, error) {
		pkg, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return "", err
		}
		app, _, err := sdk.GraphAPI.AppCatalog.Create(ctx, appcatalog.CreateRequest{DistributionMethod: appcatalog.OrganizationDistribution, Package: pkg})
		if err != nil {
			return "", err
		}
		return app.ID, nil
	}
	_ = publish
}
