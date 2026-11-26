package appcatalog_test

import (
	"context"
	"os"
	"path/filepath"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	appcatalog "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/app_catalog"
	"github.com/go-resty/resty/v2"
)

// This example compiles a real SDK workflow. It performs no network operation
// during go test; the application supplies its own client and context.
func ExampleService_List() {
	organizationApps := func(ctx context.Context, sdk *msteams.Client) ([]*appcatalog.TeamsApp, error) {
		result, err := sdk.GraphAPI.AppCatalog.List(ctx, appcatalog.ListOptions{DistributionMethod: appcatalog.OrganizationDistribution})
		return result.Items, err
	}
	_ = organizationApps
}

// Create uploads an app package (.zip) to the organization catalog; the
// returned nullable ID can address later mutations only when present and valid.
// Keep the receipt when an accepted response supplies no model or identity.
func ExampleService_Create() {
	publish := func(ctx context.Context, sdk *msteams.Client, path string) (*string, *resty.Response, error) {
		pkg, err := os.ReadFile(filepath.Clean(path))
		if err != nil {
			return nil, nil, err
		}
		app, response, err := sdk.GraphAPI.AppCatalog.Create(ctx, appcatalog.CreateRequest{DistributionMethod: appcatalog.OrganizationDistribution, Package: pkg})
		if err != nil || app == nil {
			return nil, response, err
		}
		return app.ID, response, nil
	}
	_ = publish
}
