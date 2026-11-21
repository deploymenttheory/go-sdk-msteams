package lab_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
	relationship "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group_relationship"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
	member "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_member"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_user_management"
)

const membershipFixtureKind = "membership-team"

var labDirectoryID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)
var membershipFixtureName = regexp.MustCompile(`^msteams-sdk-[a-z2-7]{26,64}-membership$`)

// User fixtures are existing lab accounts. The test changes relationships only
// inside its own new team; it never creates users or assigns tenant-wide roles.
type membershipUsers struct {
	Tenant      string `json:"tenant_id"`
	Owner       string `json:"owner_user_id"`
	SecondOwner string `json:"second_owner_user_id"`
	Member      string `json:"member_user_id"`
}

func loadMembershipUsers(file, binding string) (membershipUsers, error) {
	var value membershipUsers
	raw, err := privateLabFile(file)
	if err != nil {
		return value, lab.ErrConfig
	}
	defer clear(raw)
	if json.Unmarshal(raw, &value) != nil {
		return membershipUsers{}, lab.ErrConfig
	}
	tenant, _, ok := strings.Cut(binding, "|")
	if !ok || !strings.EqualFold(value.Tenant, tenant) {
		return membershipUsers{}, lab.ErrConfig
	}
	seen := map[string]bool{}
	for _, id := range []string{value.Owner, value.SecondOwner, value.Member} {
		key := strings.ToLower(id)
		if !labDirectoryID.MatchString(id) || seen[key] {
			return membershipUsers{}, lab.ErrConfig
		}
		seen[key] = true
	}
	return value, nil
}

// membershipCleanup never treats a denied/missing/ambiguous read as proof that
// creation did not happen. Positive team+group ownership checks precede delete.
type membershipCleanup struct{ sdk *msteams.Client }

func (h membershipCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	if !membershipFixtureName.MatchString(entry.Fixture) {
		return "", lab.ErrUnresolved
	}
	list, err := h.sdk.GraphAPI.Teams.List(ctx, team.ListOptions{Filter: "displayName eq '" + entry.Fixture + "'", Select: []string{"id", "displayName"}, MaxPages: 16})
	if err != nil {
		return "", err
	}
	id := ""
	for _, item := range list.Items {
		if item.DisplayName == nil || *item.DisplayName != entry.Fixture || !labDirectoryID.MatchString(item.ID) {
			return "", lab.ErrUnresolved
		}
		if id != "" && !strings.EqualFold(id, item.ID) {
			return "", lab.ErrUnresolved
		}
		id = item.ID
	}
	if id == "" {
		return "", lab.ErrUnresolved
	}
	return id, nil
}

func (h membershipCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !membershipFixtureName.MatchString(entry.Fixture) || !labDirectoryID.MatchString(entry.ResourceID) {
		return lab.ErrUnresolved
	}
	teamData, _, err := h.sdk.GraphAPI.Teams.GetByID(ctx, entry.ResourceID)
	if err != nil {
		return err
	}
	if !strings.EqualFold(teamData.ID, entry.ResourceID) || teamData.DisplayName == nil || *teamData.DisplayName != entry.Fixture {
		return lab.ErrUnresolved
	}
	groupData, _, err := h.sdk.GraphAPI.Groups.GetByID(ctx, entry.ResourceID, group.ReadOptions{Select: []string{"id", "displayName", "groupTypes"}})
	if err != nil {
		return err
	}
	unified := false
	for _, kind := range groupData.GroupTypes {
		if kind == "Unified" {
			unified = true
		}
	}
	if !strings.EqualFold(groupData.ID, entry.ResourceID) || groupData.DisplayName == nil || *groupData.DisplayName != entry.Fixture || !unified {
		return lab.ErrUnresolved
	}
	_, err = h.sdk.GraphAPI.Teams.DeleteByID(ctx, entry.ResourceID)
	return err
}

func (membershipCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }

func membershipLabConfig(t *testing.T) lab.Config {
	t.Helper()
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.Consequential); errors.Is(err, lab.ErrDisabled) {
		t.Skip("membership lab requires explicit read-only, mutation and consequential opt-ins")
	} else if err != nil {
		t.Fatal("invalid membership lab enablement")
	}
	return cfg
}

// TestLabTeamMembershipCRUD creates one disposable team. It is never run by
// normal CI and cannot accept a pre-existing team ID as its mutation target.
func TestLabTeamMembershipCRUD(t *testing.T) {
	cfg := membershipLabConfig(t)
	_ = teamOracleEnabled(t) // Validate the optional oracle before fixture writes.
	sdk, binding, _ := labSDKClient(t)
	users, err := loadMembershipUsers(os.Getenv("MSTEAMS_LAB_MEMBERSHIP_USERS_FILE"), binding)
	if err != nil {
		t.Fatal("explicit private same-tenant membership user fixtures are required")
	}
	journal, err := lab.NewJournal(cfg.JournalDirectory, binding)
	if err != nil {
		t.Fatal("private membership journal unavailable")
	}
	t.Logf("private recovery journal=%s", journal.FileName())
	handler := membershipCleanup{sdk: sdk}
	t.Cleanup(func() {
		cleanupCtx, cancel := context.WithTimeout(context.WithoutCancel(t.Context()), cfg.Timeout)
		defer cancel()
		if _, err := journal.Recover(cleanupCtx, map[string]lab.Handler{membershipFixtureKind: handler}); err != nil {
			t.Error("membership cleanup remains unresolved; retain private journal for recovery")
		}
		if err := journal.Close(); err != nil {
			t.Error("private membership journal close failed")
		}
	})
	fixture, err := journal.FixtureName("membership")
	if err != nil {
		t.Fatal("membership fixture allocation failed")
	}
	intent, err := journal.PlanCreate(membershipFixtureKind, fixture)
	if err != nil {
		t.Fatal("membership creation intent could not be persisted")
	}
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	accepted, err := sdk.GraphAPI.Teams.Create(ctx, team.CreateRequest{DisplayName: fixture, Members: []team.InitialMember{{UserID: users.Owner, Owner: true}, {UserID: users.SecondOwner, Owner: true}}})
	// Persist every known accepted identity before waiting. When extraction yields
	// no identity, retain the original unique-name intent for positive discovery.
	if accepted.TeamID != "" {
		if persistErr := journal.Applied(intent, accepted.TeamID); persistErr != nil {
			t.Fatal("accepted team identity could not be journalled")
		}
	}
	if err != nil || accepted.Operation == nil {
		t.Fatal("disposable team was not accepted; recovery journal retained")
	}
	if _, err := accepted.Operation.Wait(ctx, client.WaitOptions{Timeout: cfg.Timeout}); err != nil {
		t.Fatal("disposable team creation did not complete; recovery journal retained")
	}
	teamID := accepted.TeamID
	created, _, err := sdk.GraphAPI.TeamMembers.Create(ctx, teamID, member.CreateRequest{UserID: member.DirectoryUserID(users.Member)})
	if err != nil || created == nil || created.ID == "" {
		t.Fatal("native membership create failed")
	}
	if err := observeNativeMember(ctx, sdk, teamID, created.ID, users.Member, false); err != nil {
		t.Fatal("native membership read/list did not confirm creation")
	}
	if _, _, err := sdk.GraphAPI.TeamMembers.PromoteOwner(ctx, teamID, created.ID); err != nil {
		t.Fatal("native owner promotion failed")
	}
	if err := observeNativeMember(ctx, sdk, teamID, created.ID, users.Member, true); err != nil {
		t.Fatal("native membership read/list did not confirm promotion")
	}
	if _, _, err := sdk.GraphAPI.TeamMembers.DemoteOwner(ctx, teamID, created.ID); err != nil {
		t.Fatal("native owner demotion failed")
	}
	if err := observeNativeMember(ctx, sdk, teamID, created.ID, users.Member, false); err != nil {
		t.Fatal("native membership read/list did not confirm demotion")
	}
	if _, err := sdk.GraphAPI.TeamMembers.DeleteByID(ctx, teamID, created.ID); err != nil {
		t.Fatal("native membership delete failed")
	}
	if err := observeNativeAbsence(ctx, sdk, teamID, created.ID); err != nil {
		t.Fatal("native membership read/list did not confirm deletion")
	}
	if err := observeReferences(ctx, sdk, teamID, users.Member, false, false); err != nil {
		t.Fatal("backing references did not converge after native membership deletion")
	}
	if _, err := sdk.GraphAPI.TeamUserManagement.Add(ctx, teamID, users.Member, management.ChangeOptions{Role: management.Owner}); err != nil {
		t.Fatal("team-user owner add failed; partial stages remain journal-owned")
	}
	if err := observeReferences(ctx, sdk, teamID, users.Member, true, true); err != nil {
		t.Fatal("group owner/member references did not confirm composed add")
	}
	listed, err := sdk.GraphAPI.TeamUserManagement.List(ctx, teamID, management.ListOptions{})
	if err != nil || !listed.Complete {
		t.Fatal("team-user list did not complete")
	}
	found := false
	for _, user := range listed.Items {
		if strings.EqualFold(string(user.UserID), users.Member) && user.Role == management.Owner {
			found = true
		}
	}
	if !found {
		t.Fatal("team-user list did not project added owner")
	}
	comparePowerShellTeamUsers(t, ctx, sdk, teamID, fixture)
	if _, err := sdk.GraphAPI.TeamUserManagement.Remove(ctx, teamID, users.Member, management.ChangeOptions{Role: management.Owner}); err != nil {
		t.Fatal("team-user owner demotion failed")
	}
	if err := observeReferences(ctx, sdk, teamID, users.Member, true, false); err != nil {
		t.Fatal("owner-only removal did not preserve membership")
	}
	comparePowerShellTeamUsers(t, ctx, sdk, teamID, fixture)
	if _, err := sdk.GraphAPI.TeamUserManagement.Remove(ctx, teamID, users.Member, management.ChangeOptions{}); err != nil {
		t.Fatal("team-user member removal failed")
	}
	if err := observeReferences(ctx, sdk, teamID, users.Member, false, false); err != nil {
		t.Fatal("team-user removal did not converge in backing references")
	}
	comparePowerShellTeamUsers(t, ctx, sdk, teamID, fixture)
	t.Log("native membership CRUD/role changes and backing-group team-user compositions passed; guest and Teams-client convergence qualification remain outstanding")
}

// Observation polling is test-owned, read-only and explicitly bounded by ctx.
// Mutations are never retried; received read errors fail immediately.
func observeUntil(ctx context.Context, read func(context.Context) (bool, error)) error {
	for {
		ok, err := read(ctx)
		if err != nil || ok {
			return err
		}
		timer := time.NewTimer(2 * time.Second)
		select {
		case <-ctx.Done():
			timer.Stop()
			return ctx.Err()
		case <-timer.C:
		}
	}
}

func observeNativeMember(ctx context.Context, sdk *msteams.Client, teamID string, id member.MembershipID, userID string, owner bool) error {
	return observeUntil(ctx, func(ctx context.Context) (bool, error) {
		value, _, err := sdk.GraphAPI.TeamMembers.GetByID(ctx, teamID, id)
		if err != nil {
			return false, err
		}
		if value.UserID == nil || !strings.EqualFold(string(*value.UserID), userID) {
			return false, client.ErrDecode
		}
		isOwner := false
		for _, role := range value.Roles {
			if role == member.Owner {
				isOwner = true
			}
		}
		if isOwner != owner {
			return false, nil
		}
		list, err := sdk.GraphAPI.TeamMembers.ListAll(ctx, teamID, member.ListOptions{}, client.PagingOptions{MaxPages: 16})
		if err != nil {
			return false, err
		}
		for _, listed := range list.Items {
			if listed.ID == id {
				if listed.UserID == nil || !strings.EqualFold(string(*listed.UserID), userID) {
					return false, client.ErrDecode
				}
				listedOwner := false
				for _, role := range listed.Roles {
					if role == member.Owner {
						listedOwner = true
					}
				}
				return listedOwner == owner, nil
			}
		}
		return false, nil
	})
}

func observeNativeAbsence(ctx context.Context, sdk *msteams.Client, teamID string, id member.MembershipID) error {
	return observeUntil(ctx, func(ctx context.Context) (bool, error) {
		_, _, err := sdk.GraphAPI.TeamMembers.GetByID(ctx, teamID, id)
		var apiError *client.APIError
		if err == nil {
			return false, nil
		}
		if !errors.As(err, &apiError) || apiError.Backend != client.Graph || apiError.StatusCode != 404 {
			return false, err
		}
		list, err := sdk.GraphAPI.TeamMembers.ListAll(ctx, teamID, member.ListOptions{}, client.PagingOptions{MaxPages: 16})
		if err != nil {
			return false, err
		}
		for _, item := range list.Items {
			if item.ID == id {
				return false, nil
			}
		}
		return true, nil
	})
}

func observeReferences(ctx context.Context, sdk *msteams.Client, teamID, userID string, wantMember, wantOwner bool) error {
	return observeUntil(ctx, func(ctx context.Context) (bool, error) {
		members, err := sdk.GraphAPI.GroupMembers.ListAll(ctx, teamID, relationship.ListOptions{Select: []string{"id"}}, client.PagingOptions{MaxPages: 16})
		if err != nil {
			return false, err
		}
		owners, err := sdk.GraphAPI.GroupOwners.ListAll(ctx, teamID, relationship.ListOptions{Select: []string{"id"}}, client.PagingOptions{MaxPages: 16})
		if err != nil {
			return false, err
		}
		hasMember, hasOwner := false, false
		for _, item := range members.Items {
			if strings.EqualFold(string(item.ID), userID) {
				hasMember = true
			}
		}
		for _, item := range owners.Items {
			if strings.EqualFold(string(item.ID), userID) {
				hasOwner = true
			}
		}
		return hasMember == wantMember && hasOwner == wantOwner, nil
	})
}

// TestLabRecoverTeamMembership only reopens an explicit private journal under the
// same tenant/cloud/app binding. Denied/absent/ambiguous reads keep it unresolved.
func TestLabRecoverTeamMembership(t *testing.T) {
	cfg := membershipLabConfig(t)
	sdk, binding, _ := labSDKClient(t)
	file := os.Getenv("MSTEAMS_LAB_RECOVERY_JOURNAL")
	if file == "" {
		t.Skip("explicit private recovery journal is required")
	}
	journal, err := lab.OpenJournal(file, binding)
	if err != nil {
		t.Fatal("private journal unavailable for this binding")
	}
	defer func() {
		if err := journal.Close(); err != nil {
			t.Error("private recovery journal close failed")
		}
	}()
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	if _, err := journal.Recover(ctx, map[string]lab.Handler{membershipFixtureKind: membershipCleanup{sdk: sdk}}); err != nil {
		t.Fatal("membership recovery remains unresolved")
	}
	t.Log("membership journal recovery completed")
}
