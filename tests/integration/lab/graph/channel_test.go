package graphacc

import (
	"context"
	"errors"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_management"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// TestAccChannelCRUD changes only channels in its own newly journalled team.
// The shared lifecycle recovery handler positively verifies and deletes that
// parent team/group, including child channels after an interrupted run.
func TestAccChannelCRUD(t *testing.T) {
	cfg := lifecycleLabConfig(t) // All three opt-ins precede any credential read.
	sdk, binding, _ := labkit.SDKClient(t)
	owner, err := loadLifecycleOwner(os.Getenv("MSTEAMS_LAB_TEAM_OWNER_FILE"), binding)
	if err != nil {
		t.Fatal("explicit private same-tenant owner fixture is required")
	}
	journal, fixture, intent := lifecycleJournal(t, cfg, sdk, binding, "team")
	ctx, cancel := context.WithTimeout(t.Context(), cfg.Timeout)
	defer cancel()
	created, err := sdk.GraphAPI.TeamManagement.Create(ctx, management.CreateRequest{
		DisplayName: fixture, MailNickname: &fixture, Owner: owner.UserID,
		RetainCreatedGroup: true, Options: management.CreationOptions{Timeout: cfg.Timeout},
	})
	if created.CreatedGroupOwned && created.ID != "" {
		if err := journal.Applied(intent, created.ID); err != nil {
			t.Fatal("parent group identity could not be journalled; retain journal")
		}
	}
	if err != nil || created.Value == nil || !lifecycleDirectoryID.MatchString(created.ID) {
		t.Fatal("owned parent team creation was not confirmed; retain journal")
	}
	for _, kind := range []channel.MembershipType{channel.Standard, channel.Private, channel.Shared} {
		t.Run(string(kind), func(t *testing.T) {
			stage, stop := context.WithTimeout(t.Context(), cfg.Timeout)
			defer stop()
			name, description := "SDK "+string(kind), "SDK channel smoke fixture"
			request := channel.CreateRequest{DisplayName: name, Description: &description, MembershipType: kind}
			if kind != channel.Standard {
				request.OwnerUserID = &owner.UserID
			}
			result, err := sdk.GraphAPI.Channels.Create(stage, created.ID, request)
			if err != nil {
				t.Fatal("channel creation failed; parent journal retains cleanup ownership")
			}
			value := result.Channel
			if kind == channel.Shared {
				if result.Operation == nil || result.Channel != nil {
					t.Fatal("shared channel was not returned as an accepted operation")
				}
				if _, err := result.Operation.Wait(stage, client.WaitOptions{MaxPolls: 8, Timeout: cfg.Timeout}); err != nil {
					t.Fatal("shared channel operation did not succeed")
				}
				value, _, err = sdk.GraphAPI.Channels.ResolveByDisplayName(stage, created.ID, name, client.PagingOptions{MaxPages: 16})
				if err != nil {
					t.Fatal("completed shared channel could not be uniquely resolved")
				}
			}
			if value == nil || value.ID == "" {
				t.Fatal("channel identity was not returned")
			}
			read, _, err := sdk.GraphAPI.Channels.GetByID(stage, created.ID, value.ID, channel.ReadOptions{IncludeUnknownEnumMembers: true})
			if err != nil || read.DisplayName == nil || *read.DisplayName != name || read.MembershipType == nil || *read.MembershipType != kind {
				t.Fatal("channel creation readback differed")
			}
			changedName, empty := name+" renamed", ""
			if _, err := sdk.GraphAPI.Channels.UpdateByID(stage, created.ID, value.ID, channel.UpdateRequest{DisplayName: &changedName, Description: &empty}); err != nil {
				t.Fatal("channel update failed")
			}
			read, _, err = sdk.GraphAPI.Channels.GetByID(stage, created.ID, value.ID, channel.ReadOptions{})
			if err != nil || read.DisplayName == nil || *read.DisplayName != changedName || read.Description == nil || *read.Description != "" {
				t.Fatal("channel update readback differed")
			}
			if _, err := sdk.GraphAPI.Channels.ListAll(stage, created.ID, channel.ListOptions{}, client.PagingOptions{MaxPages: 16}); err != nil {
				t.Fatal("complete all-channel listing failed")
			}
			if _, err := sdk.GraphAPI.Channels.ListIncoming(stage, created.ID, channel.ReadOptions{}, client.PagingOptions{MaxPages: 16}); err != nil {
				t.Fatal("complete incoming-channel listing failed")
			}
			if _, err := sdk.GraphAPI.Channels.DeleteByDisplayName(stage, created.ID, changedName, client.PagingOptions{MaxPages: 16}); err != nil {
				t.Fatal("unique named-channel deletion failed")
			}
			_, _, err = sdk.GraphAPI.Channels.GetByID(stage, created.ID, value.ID, channel.ReadOptions{})
			var apiErr *client.APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != 404 {
				t.Fatal("channel deletion was not confirmed by a typed 404")
			}
		})
	}
	t.Log("owned channel smoke scenarios completed; PowerShell comparison, external sharing and cloud qualification remain separate")
}
