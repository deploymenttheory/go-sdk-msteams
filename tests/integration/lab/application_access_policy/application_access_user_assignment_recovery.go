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

type assignmentUserFixture struct {
	TenantID      string `json:"tenant_id"`
	UserID        string `json:"user_id"`
	PrincipalName string `json:"user_principal_name"`
	DisplayName   string `json:"display_name"`
}

func loadAssignmentUserFixture(t *testing.T, binding string) assignmentUserFixture {
	t.Helper()
	raw, err := labkit.PrivateFile(os.Getenv("MSTEAMS_LAB_USER_FIXTURE_FILE"))
	if err != nil {
		t.Fatal("explicit private assignment-user fixture is required")
	}
	defer clear(raw)
	var fixture assignmentUserFixture
	tenant, _, ok := strings.Cut(binding, "|")
	if json.Unmarshal(raw, &fixture) != nil || !ok || !strings.EqualFold(fixture.TenantID, tenant) || !labkit.DirectoryID.MatchString(fixture.UserID) || !strings.HasPrefix(fixture.PrincipalName, "go-sdk-msteams-acceptance@") || fixture.DisplayName != "go-sdk-msteams acceptance user" {
		t.Fatal("assignment fixture ownership or tenant binding differs")
	}
	return fixture
}

type userAssignmentCleanup struct {
	sdk    *msteams.Client
	userID string
}

func (userAssignmentCleanup) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}

func (userAssignmentCleanup) Delete(context.Context, lab.Entry) error { return lab.ErrUnresolved }

func (h userAssignmentCleanup) Restore(ctx context.Context, entry lab.Entry) error {
	if entry.ResourceID != h.userID || !labkit.DirectoryID.MatchString(h.userID) {
		return lab.ErrUnresolved
	}
	var before []*assignments.EffectivePolicy
	if json.Unmarshal(entry.Snapshot, &before) != nil || before == nil || len(before) != 0 {
		return lab.ErrUnresolved
	}
	if _, err := h.sdk.TeamsAdminAPI.ApplicationAccessPolicies.AssignToUser(ctx, h.userID, ""); err != nil {
		return err
	}
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		values, _, err := h.sdk.TeamsAdminAPI.PolicyAssignments.GetUser(ctx, h.userID, "ApplicationAccessPolicy")
		return err == nil && len(values) == 0, err
	})
}
