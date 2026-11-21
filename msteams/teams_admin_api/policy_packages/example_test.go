package policypackages_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_packages"
)

func ExampleService_Create() {
	use := func(ctx context.Context, sdk *msteams.Client) error {
		policies, e := resource.ParsePolicyList([]string{"TeamsMeetingPolicy,LabMeetingPolicy"})
		if e != nil {
			return e
		}
		input := resource.CustomPackageRequest{Name: new("LabPackage"), Description: new("Lab description"), Policies: policies}
		service := sdk.TeamsAdminAPI.PolicyPackages
		if _, _, e = service.Create(ctx, input); e != nil {
			return e
		}
		if _, _, e = service.GetByName(ctx, *input.Name); e != nil {
			return e
		}
		input.Description = new("Updated lab description")
		if _, _, e = service.Update(ctx, input); e != nil {
			return e
		}
		if _, _, e = service.List(ctx); e != nil {
			return e
		}
		_, _, e = service.DeleteByName(ctx, *input.Name)
		return e
	}
	_ = use // Compiling example; no operation or delegated sign-in runs here.
}

func ExampleService_SubmitBatch() {
	use := func(ctx context.Context, sdk *msteams.Client, users []string) error {
		result, _, err := sdk.TeamsAdminAPI.PolicyPackages.SubmitBatch(ctx, resource.UserAssignmentRequest{PackageName: new("LabPackage"), UserIDs: users})
		if result != nil {
			_ = result.HTTPAccepted
			_ = result.OperationID
			_ = result.FailedCreatePolicies
		}
		// Inspect partial failures explicitly. A returned ID does not establish
		// a polling route or mean that user policies have become effective.
		return err
	}
	_ = use
}
