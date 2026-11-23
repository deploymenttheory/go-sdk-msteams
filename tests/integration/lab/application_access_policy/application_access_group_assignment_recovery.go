package applicationaccesspolicyacc

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	assignments "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_assignments"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

type assignmentGroupFixture struct {
	TenantID                string `json:"tenant_id"`
	GroupID                 string `json:"group_id"`
	DisplayName             string `json:"display_name"`
	MailNickname            string `json:"mail_nickname"`
	SecurityEnabled         bool   `json:"security_enabled"`
	MailEnabled             bool   `json:"mail_enabled"`
	RoleAssignable          bool   `json:"role_assignable"`
	EmptyMembershipVerified bool   `json:"empty_membership_verified"`
}

func loadAssignmentGroupFixture(t *testing.T, binding string) assignmentGroupFixture {
	t.Helper()
	raw, err := labkit.PrivateFile(os.Getenv("MSTEAMS_LAB_GROUP_FIXTURE_FILE"))
	if err != nil {
		t.Fatal("explicit private assignment-group fixture is required")
	}
	defer clear(raw)
	var fixture assignmentGroupFixture
	tenant, _, ok := strings.Cut(binding, "|")
	if json.Unmarshal(raw, &fixture) != nil || !ok || !strings.EqualFold(fixture.TenantID, tenant) || !labkit.DirectoryID.MatchString(fixture.GroupID) || fixture.DisplayName != "go-sdk-msteams acceptance policy group" || fixture.MailNickname != "go-sdk-msteams-acceptance-policy" || !fixture.SecurityEnabled || fixture.MailEnabled || fixture.RoleAssignable || !fixture.EmptyMembershipVerified {
		t.Fatal("assignment group ownership or tenant binding differs")
	}
	return fixture
}

type groupAssignmentCleanup struct {
	sdk     *msteams.Client
	groupID string
}

func (groupAssignmentCleanup) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}

func (groupAssignmentCleanup) Delete(context.Context, lab.Entry) error { return lab.ErrUnresolved }

func (h groupAssignmentCleanup) absent(ctx context.Context) (bool, error) {
	_, _, err := h.sdk.TeamsAdminAPI.PolicyAssignments.GetGroup(ctx, h.groupID, "ApplicationAccessPolicy")
	if !labkit.IsHTTP(err, 404) {
		return false, err
	}
	values, _, err := h.sdk.TeamsAdminAPI.PolicyAssignments.ListGroup(ctx, h.groupID)
	if err != nil {
		return false, err
	}
	for _, value := range values {
		if value == nil || value.PolicyType == nil {
			return false, lab.ErrUnresolved
		}
		if strings.EqualFold(*value.PolicyType, "ApplicationAccessPolicy") {
			return false, nil
		}
	}
	return true, nil
}

func (h groupAssignmentCleanup) Restore(ctx context.Context, entry lab.Entry) error {
	var before []*assignments.GroupAssignment
	if entry.ResourceID != h.groupID || !labkit.DirectoryID.MatchString(h.groupID) || json.Unmarshal(entry.Snapshot, &before) != nil || before == nil || len(before) != 0 {
		return lab.ErrUnresolved
	}
	absent, err := h.absent(ctx)
	if err != nil {
		return err
	}
	if !absent {
		if _, _, err := h.sdk.TeamsAdminAPI.PolicyAssignments.DeleteGroup(ctx, h.groupID, "ApplicationAccessPolicy"); err != nil && !labkit.IsHTTP(err, 404) {
			return err
		}
	}
	return labkit.ObserveUntil(ctx, h.absent)
}

// Never delete a policy while this fixture still has an assignment: recovery
// continues independent entries after a failure, so enforce the dependency here.
type groupOwnedPolicyCleanup struct {
	policyCleanup
	assignment groupAssignmentCleanup
}

func (h groupOwnedPolicyCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	absent, err := h.assignment.absent(ctx)
	if err != nil {
		return err
	}
	if !absent {
		return lab.ErrUnresolved
	}
	return h.policyCleanup.Delete(ctx, entry)
}
