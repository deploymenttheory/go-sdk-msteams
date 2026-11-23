package graphacc

import (
	"context"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel"
	member "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_member"
	userservice "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_user_management"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_management"
	teammember "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_member"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccChannelMembershipCRUD changes memberships only in a private channel of
// its own journalled parent team. Existing fixtures are directory users only.
func TestAccChannelMembershipCRUD(t *testing.T) {
	cfg := lifecycleLabConfig(t)
	_ = labkit.OracleEnabled(t) // Validate optional oracle prerequisites before writes.
	sdk, binding, _ := labkit.SDKClient(t)
	users, err := loadMembershipUsers(os.Getenv("MSTEAMS_LAB_MEMBERSHIP_USERS_FILE"), binding)
	if err != nil {
		t.Fatal("explicit private same-tenant membership users are required")
	}
	journal, fixture, intent := lifecycleJournal(t, cfg, sdk, binding, "team")
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	parent, err := sdk.GraphAPI.TeamManagement.Create(ctx, management.CreateRequest{DisplayName: fixture, MailNickname: &fixture,
		Owner: users.Owner, RetainCreatedGroup: true, Options: management.CreationOptions{Timeout: cfg.Timeout}})
	if parent.CreatedGroupOwned && parent.ID != "" {
		if err := journal.Applied(intent, parent.ID); err != nil {
			t.Fatal("parent identity could not be journalled; retain private journal")
		}
	}
	if err != nil || parent.Value == nil || !lifecycleDirectoryID.MatchString(parent.ID) {
		t.Fatal("owned parent team was not confirmed; retain private journal")
	}
	for _, id := range []string{users.SecondOwner, users.Member} {
		if _, _, err := sdk.GraphAPI.TeamMembers.Create(ctx, parent.ID, teammember.CreateRequest{UserID: teammember.DirectoryUserID(id)}); err != nil {
			t.Fatal("owned-team membership prerequisite failed")
		}
	}
	name := "SDK membership"
	created, err := sdk.GraphAPI.Channels.Create(ctx, parent.ID, channel.CreateRequest{DisplayName: name, MembershipType: channel.Private, OwnerUserID: &users.Owner})
	if err != nil || created.Channel == nil || created.Channel.ID == "" {
		t.Fatal("owned private channel creation failed")
	}
	channelID := created.Channel.ID
	native, _, err := sdk.GraphAPI.ChannelMembers.Create(ctx, parent.ID, channelID, member.CreateRequest{UserID: member.DirectoryUserID(users.SecondOwner)})
	if err != nil || native == nil || native.ID == "" {
		t.Fatal("native channel membership create failed")
	}
	observeChannelMember(t, ctx, sdk, parent.ID, channelID, native.ID, users.SecondOwner, false)
	if _, _, err := sdk.GraphAPI.ChannelMembers.PromoteOwner(ctx, parent.ID, channelID, native.ID); err != nil {
		t.Fatal("native channel owner promotion failed")
	}
	observeChannelMember(t, ctx, sdk, parent.ID, channelID, native.ID, users.SecondOwner, true)
	if _, _, err := sdk.GraphAPI.ChannelMembers.DemoteOwner(ctx, parent.ID, channelID, native.ID); err != nil {
		t.Fatal("native channel owner demotion failed")
	}
	observeChannelMember(t, ctx, sdk, parent.ID, channelID, native.ID, users.SecondOwner, false)
	if _, err := sdk.GraphAPI.ChannelMembers.DeleteByID(ctx, parent.ID, channelID, native.ID); err != nil {
		t.Fatal("native channel membership deletion failed")
	}
	observeChannelMemberAbsent(t, ctx, sdk, parent.ID, channelID, native.ID)
	added, err := sdk.GraphAPI.ChannelUserManagement.Add(ctx, parent.ID, name, users.Member, userservice.ChangeOptions{})
	if err != nil || added.MembershipID == "" {
		t.Fatal("channel-user add failed")
	}
	observeChannelMember(t, ctx, sdk, parent.ID, channelID, added.MembershipID, users.Member, false)
	comparePowerShellChannelUsers(t, ctx, sdk, parent.ID, fixture, name)
	if _, err := sdk.GraphAPI.ChannelUserManagement.Add(ctx, parent.ID, name, users.Member, userservice.ChangeOptions{Role: userservice.Owner}); err != nil {
		t.Fatal("channel-user owner promotion failed")
	}
	observeChannelMember(t, ctx, sdk, parent.ID, channelID, added.MembershipID, users.Member, true)
	comparePowerShellChannelUsers(t, ctx, sdk, parent.ID, fixture, name)
	if _, err := sdk.GraphAPI.ChannelUserManagement.Remove(ctx, parent.ID, name, users.Member, userservice.ChangeOptions{Role: userservice.Owner}); err != nil {
		t.Fatal("channel-user owner demotion failed")
	}
	observeChannelMember(t, ctx, sdk, parent.ID, channelID, added.MembershipID, users.Member, false)
	comparePowerShellChannelUsers(t, ctx, sdk, parent.ID, fixture, name)
	if _, err := sdk.GraphAPI.ChannelUserManagement.Remove(ctx, parent.ID, name, users.Member, userservice.ChangeOptions{}); err != nil {
		t.Fatal("channel-user membership deletion failed")
	}
	observeChannelMemberAbsent(t, ctx, sdk, parent.ID, channelID, added.MembershipID)
	comparePowerShellChannelUsers(t, ctx, sdk, parent.ID, fixture, name)
	t.Log("owned private-channel membership observations completed; shared/external, guest, cloud and client-convergence qualification remain separate")
}

func observeChannelMember(t *testing.T, ctx context.Context, sdk *msteams.Client, teamID, channelID string, id member.MembershipID, userID string, owner bool) {
	t.Helper()
	err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		value, _, err := sdk.GraphAPI.ChannelMembers.GetByID(ctx, teamID, channelID, id)
		if err != nil {
			return false, err
		}
		if value.UserID == nil || !strings.EqualFold(string(*value.UserID), userID) {
			return false, client.ErrDecode
		}
		isOwner := false
		for _, role := range value.Roles {
			isOwner = isOwner || strings.EqualFold(string(role), string(member.Owner))
		}
		if isOwner != owner {
			return false, nil
		}
		listed, err := sdk.GraphAPI.ChannelMembers.ListAll(ctx, teamID, channelID, member.ListOptions{}, client.PagingOptions{MaxPages: 16})
		if err != nil {
			return false, err
		}
		for _, listedMember := range listed.Items {
			if listedMember.ID != id {
				continue
			}
			if listedMember.UserID == nil || !strings.EqualFold(string(*listedMember.UserID), userID) {
				return false, client.ErrDecode
			}
			listedOwner := false
			for _, role := range listedMember.Roles {
				listedOwner = listedOwner || strings.EqualFold(string(role), string(member.Owner))
			}
			return listedOwner == owner, nil
		}
		return false, nil
	})
	if err != nil {
		t.Fatal("channel membership identity/role readback did not converge")
	}
}

func observeChannelMemberAbsent(t *testing.T, ctx context.Context, sdk *msteams.Client, teamID, channelID string, id member.MembershipID) {
	t.Helper()
	err := labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		_, _, err := sdk.GraphAPI.ChannelMembers.GetByID(ctx, teamID, channelID, id)
		if err == nil {
			return false, nil
		}
		var apiErr *client.APIError
		if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
			return false, err
		}
		listed, err := sdk.GraphAPI.ChannelMembers.ListAll(ctx, teamID, channelID, member.ListOptions{}, client.PagingOptions{MaxPages: 16})
		if err != nil {
			return false, err
		}
		for _, value := range listed.Items {
			if value.ID == id {
				return false, nil
			}
		}
		return true, nil
	})
	if err != nil {
		t.Fatal("channel membership deletion did not converge across typed 404 and complete list")
	}
}

func comparePowerShellChannelUsers(t *testing.T, ctx context.Context, sdk *msteams.Client, teamID, fixture, channelName string) {
	t.Helper()
	listed, err := sdk.GraphAPI.ChannelUserManagement.List(ctx, teamID, channelName, userservice.ListOptions{Traversal: client.PagingOptions{MaxPages: 16}})
	if err != nil || !listed.Complete {
		t.Fatal("complete enriched Go channel-user snapshot is required")
	}
	if !labkit.OracleEnabled(t) {
		return
	}
	want := make([]labkit.TeamOracleUser, 0, len(listed.Items))
	for _, value := range listed.Items {
		if value.UserID == nil {
			t.Fatal("owned channel member has no comparable directory identity")
		}
		want = append(want, labkit.TeamOracleUser{ID: string(*value.UserID), User: value.User, Name: value.Name, Role: string(value.Role)})
	}
	observed := labkit.RunTeamOracle(t, ctx, labkit.TeamOracleRequest{Kind: "channel-users", ID: teamID, Fixture: fixture, Channel: channelName})
	if !sameTeamOracleUsers(want, observed.Users) {
		t.Fatal("PowerShell and Go channel-user projection differed")
	}
	t.Log("PowerShell 8.1.0 / Go private channel-user projection: passed")
}
