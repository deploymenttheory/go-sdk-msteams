package channelmember_test

import (
	"context"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	channelmember "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_member"
)

func TestUnitExternalTenantCreationPreservesExplicitTenant(t *testing.T) {
	t.Parallel()
	tenant := "33333333-3333-4333-8333-333333333333"
	path := "/beta/teams/" + teamID + "/channels/19:channel%2Fopaque@thread.tacv2/members"
	body := `{"@odata.type":"#microsoft.graph.aadUserConversationMember","roles":[],"tenantId":"` + tenant + `","user@odata.bind":"https://graph.microsoft.com/beta/users('` + userID + `')"}`
	raw := resourceJSON(`[]`)
	nativeCase(t, testutil.Case{Name: "external-tenant", Steps: []testutil.Exchange{wire("POST", auth.GraphAudience, path, body, 201, raw)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			return snapshot(sdk.GraphAPI.ChannelMembers.Create(ctx, teamID, channelID, channelmember.CreateRequest{UserID: userID, TenantID: &tenant}))
		}, Want: testutil.Expected{Status: 201, JSON: expectedSnapshot(`[]`, raw)}})
}

func TestUnitInvalidChannelAndTenantAreRejectedBeforeCredentials(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, channel, tenant string }{
		{"empty-channel", "", userID},
		{"dot-channel", "..", userID},
		{"control-channel", "private\n", userID},
		{"invalid-tenant", channelID, "not-a-uuid"},
		{"empty-tenant", channelID, ""},
	} {
		t.Run(tc.name, func(t *testing.T) {
			nativeCase(t, testutil.Case{Name: tc.name, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				return snapshot(sdk.GraphAPI.ChannelMembers.Create(ctx, teamID, tc.channel, channelmember.CreateRequest{UserID: userID, TenantID: &tc.tenant}))
			}, Want: testutil.Expected{Error: client.ErrRequest}})
		})
	}
}
