package assignmentmigration_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/assignment_migration"
)

func ExampleService_Create() {
	create := func(ctx context.Context, sdk *msteams.Client, groupID string) (*resource.Job, error) {
		job, _, err := sdk.TeamsAdminAPI.AssignmentMigrations.Create(ctx, resource.CreateRequest{Authority: "Tenant", PolicyType: "TeamsMeetingPolicy", PolicyName: "Fixture", GroupID: groupID})

		return job, err
	}
	_ = create
}

func ExampleService_List() {
	read := func(ctx context.Context, sdk *msteams.Client) ([]*resource.Job, error) {
		jobs, _, err := sdk.TeamsAdminAPI.AssignmentMigrations.List(ctx)
		return jobs, err
	}
	_ = read
}

func ExampleService_RecommendationsByPolicyType() {
	read := func(ctx context.Context, sdk *msteams.Client) ([]*resource.PolicyTypeRecommendation, error) {
		items, _, err := sdk.TeamsAdminAPI.AssignmentMigrations.RecommendationsByPolicyType(ctx, "groups", resource.RecommendationOptions{})
		return items, err
	}
	_ = read
}
