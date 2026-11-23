// Package teamsmedialoggingpolicyacc qualifies media logging policy reads and
// reversible grants of the built-in Enabled instance to the dedicated SDK
// user and group fixtures.
package teamsmedialoggingpolicyacc

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

const policyType = "TeamsMediaLoggingPolicy"

// builtInPolicy is the service-provided non-default instance. The family has no
// public New/Set/Remove, so grants target it rather than an owned policy.
const builtInPolicy = "Enabled"

// FixtureRecoveryHandlers returns the handlers that restore the dedicated user
// and group fixtures. Each is registered only when its private fixture file is
// configured, because recovery must bind to that exact identity.
func FixtureRecoveryHandlers(t *testing.T, sdk *msteams.Client, binding string) map[string]lab.Handler {
	t.Helper()
	handlers := map[string]lab.Handler{}
	if os.Getenv("MSTEAMS_LAB_USER_FIXTURE_FILE") != "" {
		handlers["media-logging-user-assignment"] = userGrantCleanup{sdk, loadUserFixture(t, binding).UserID}
	}
	if os.Getenv("MSTEAMS_LAB_GROUP_FIXTURE_FILE") != "" {
		handlers["media-logging-group-assignment"] = groupGrantCleanup{sdk, loadGroupFixture(t, binding).GroupID}
	}
	return handlers
}

type userFixture struct {
	TenantID      string `json:"tenant_id"`
	UserID        string `json:"user_id"`
	PrincipalName string `json:"user_principal_name"`
	DisplayName   string `json:"display_name"`
}

func loadUserFixture(t *testing.T, binding string) userFixture {
	t.Helper()
	raw, err := labkit.PrivateFile(os.Getenv("MSTEAMS_LAB_USER_FIXTURE_FILE"))
	if err != nil {
		t.Fatal("explicit private assignment-user fixture is required")
	}
	defer clear(raw)
	var fixture userFixture
	tenant, _, ok := strings.Cut(binding, "|")
	if json.Unmarshal(raw, &fixture) != nil || !ok || !strings.EqualFold(fixture.TenantID, tenant) || !labkit.DirectoryID.MatchString(fixture.UserID) || !strings.HasPrefix(fixture.PrincipalName, "go-sdk-msteams-acceptance@") || fixture.DisplayName != "go-sdk-msteams acceptance user" {
		t.Fatal("assignment fixture ownership or tenant binding differs")
	}
	return fixture
}

type groupFixture struct {
	TenantID                string `json:"tenant_id"`
	GroupID                 string `json:"group_id"`
	DisplayName             string `json:"display_name"`
	MailNickname            string `json:"mail_nickname"`
	SecurityEnabled         bool   `json:"security_enabled"`
	MailEnabled             bool   `json:"mail_enabled"`
	RoleAssignable          bool   `json:"role_assignable"`
	EmptyMembershipVerified bool   `json:"empty_membership_verified"`
}

func loadGroupFixture(t *testing.T, binding string) groupFixture {
	t.Helper()
	raw, err := labkit.PrivateFile(os.Getenv("MSTEAMS_LAB_GROUP_FIXTURE_FILE"))
	if err != nil {
		t.Fatal("explicit private assignment-group fixture is required")
	}
	defer clear(raw)
	var fixture groupFixture
	tenant, _, ok := strings.Cut(binding, "|")
	if json.Unmarshal(raw, &fixture) != nil || !ok || !strings.EqualFold(fixture.TenantID, tenant) || !labkit.DirectoryID.MatchString(fixture.GroupID) || fixture.DisplayName != "go-sdk-msteams acceptance policy group" || fixture.MailNickname != "go-sdk-msteams-acceptance-policy" || !fixture.SecurityEnabled || fixture.MailEnabled || fixture.RoleAssignable || !fixture.EmptyMembershipVerified {
		t.Fatal("assignment group ownership or tenant binding differs")
	}
	return fixture
}

// userGrantCleanup restores the journaled empty direct assignment through the
// family's own Grant clear form and verifies it independently.
type userGrantCleanup struct {
	sdk    *msteams.Client
	userID string
}

func (userGrantCleanup) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}

func (userGrantCleanup) Delete(context.Context, lab.Entry) error { return lab.ErrUnresolved }

func (h userGrantCleanup) Restore(ctx context.Context, entry lab.Entry) error {
	if entry.ResourceID != h.userID || !labkit.DirectoryID.MatchString(h.userID) {
		return lab.ErrUnresolved
	}
	var before []*assignments.EffectivePolicy
	if json.Unmarshal(entry.Snapshot, &before) != nil || before == nil || len(before) != 0 {
		return lab.ErrUnresolved
	}
	if _, err := h.sdk.TeamsAdminAPI.TeamsMediaLoggingPolicies.AssignToUser(ctx, h.userID, ""); err != nil {
		return err
	}
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		values, _, err := h.sdk.TeamsAdminAPI.PolicyAssignments.GetUser(ctx, h.userID, policyType)
		return err == nil && len(values) == 0, err
	})
}

// groupGrantCleanup removes only this family's assignment from the owned group
// and requires the journaled empty baseline.
type groupGrantCleanup struct {
	sdk     *msteams.Client
	groupID string
}

func (groupGrantCleanup) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}

func (groupGrantCleanup) Delete(context.Context, lab.Entry) error { return lab.ErrUnresolved }

func (h groupGrantCleanup) absent(ctx context.Context) (bool, error) {
	_, _, err := h.sdk.TeamsAdminAPI.PolicyAssignments.GetGroup(ctx, h.groupID, policyType)
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
		if strings.EqualFold(*value.PolicyType, policyType) {
			return false, nil
		}
	}
	return true, nil
}

func (h groupGrantCleanup) Restore(ctx context.Context, entry lab.Entry) error {
	var before []*assignments.GroupAssignment
	if entry.ResourceID != h.groupID || !labkit.DirectoryID.MatchString(h.groupID) || json.Unmarshal(entry.Snapshot, &before) != nil || before == nil || len(before) != 0 {
		return lab.ErrUnresolved
	}
	absent, err := h.absent(ctx)
	if err != nil {
		return err
	}
	if !absent {
		if _, _, err := h.sdk.TeamsAdminAPI.PolicyAssignments.DeleteGroup(ctx, h.groupID, policyType); err != nil && !labkit.IsHTTP(err, 404) {
			return err
		}
	}
	return labkit.ObserveUntil(ctx, h.absent)
}
