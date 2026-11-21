package policyassignments_test

import (
	"context"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
)

func ExampleService_CreateGroup() {
	use := func(ctx context.Context, sdk *msteams.Client, groupID string) error {
		service := sdk.TeamsAdminAPI.PolicyAssignments
		created, _, err := service.CreateGroup(ctx, groupID, "TeamsMeetingPolicy", resource.GroupSettings{PolicyName: policy.Set("LabMeetingPolicy"), Priority: policy.Set(int32(0))})
		if err != nil {
			return err
		}
		// Acceptance is independent of effective policy propagation.
		_ = created.HTTPAccepted
		_, _, err = service.GetGroup(ctx, groupID, "TeamsMeetingPolicy")
		if err != nil {
			return err
		}
		_, _, err = service.UpdateGroup(ctx, groupID, "TeamsMeetingPolicy", resource.GroupSettings{Priority: policy.Set(int32(1))})
		if err != nil {
			return err
		}
		_, _, err = service.ListGroup(ctx, groupID)
		if err != nil {
			return err
		}
		_, _, err = service.DeleteGroup(ctx, groupID, "TeamsMeetingPolicy")
		return err
	}
	_ = use // Compiling example; no tenant operation runs automatically.
}

func ExampleService_CreateBatch() {
	use := func(ctx context.Context, sdk *msteams.Client, users []string) error {
		service := sdk.TeamsAdminAPI.PolicyAssignments
		result, _, err := service.CreateBatch(ctx, resource.BatchCreateRequest{Identities: users, PolicyType: policy.Set("TeamsMeetingPolicy"), PolicyName: policy.Set("LabMeetingPolicy"), OperationName: "Lab assignment"})
		if err != nil {
			return err
		}
		if result.OperationID != nil && *result.OperationID != "" {
			status, _, err := service.GetBatch(ctx, *result.OperationID)
			if err != nil {
				return err
			}
			_ = status.UserState // Inspect each user; one snapshot is not a wait.
		}
		return nil
	}
	_ = use
}
