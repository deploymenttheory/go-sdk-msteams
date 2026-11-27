package channelusermanagement_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	member "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_member"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_user_management"
)

func TestUnitPrivateUPNFailurePreventsMembershipReadsAndWrites(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		status     int
		err        error
	}{
		{"forbidden", `{"error":{"code":"Denied"}}`, 403, auth.ErrForbidden},
		{"malformed", `{}`, 200, client.ErrDecode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: "upn-" + tc.name, Steps: []testutil.Exchange{channelRead("private"), wire("GET", "/v1.0/users/member@fixture.invalid?$select=id", "", tc.status, tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				return changed(sdk.GraphAPI.ChannelUserManagement.Add(ctx, teamID, "fixture", "member@fixture.invalid", management.ChangeOptions{Role: management.Owner}))
			}, Want: testutil.Expected{Status: tc.status, Error: tc.err}})
		})
	}
}

func TestUnitSharedNullEmailAndTenantSelectionUseReturnedMembership(t *testing.T) {
	t.Parallel()
	first := `{"id":"first","userId":"` + userID + `","email":null,"tenantId":"` + teamID + `","roles":[]}`
	second := `{"id":"second","userId":"` + userID + `","email":null,"tenantId":"` + tenantID + `","roles":[]}`
	for _, tc := range []struct {
		name, body string
		tenant     *string
		id         string
		err        error
	}{
		{name: "null-email-id", body: `{"value":[` + first + `]}`, id: "first"},
		{name: "cross-tenant-ambiguous", body: `{"value":[` + first + `,` + second + `]}`, err: member.ErrAmbiguous},
		{name: "explicit-tenant", body: `{"value":[` + first + `,` + second + `]}`, tenant: stringPointer(tenantID), id: "second"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			steps := []testutil.Exchange{channelRead("shared"), wire("GET", membersPath(), "", 200, tc.body)}
			status := 200
			var want json.RawMessage
			if tc.err == nil {
				reply := first
				if tc.id == "second" {
					reply = second
				}
				reply = strings.Replace(reply, `"roles":[]`, `"roles":["owner"]`, 1)
				steps = append(steps, wire("PATCH", membersPath()+"/"+tc.id, `{"roles":["owner"]}`, 200, reply))
				want = json.RawMessage(`{"membership":"` + tc.id + `"}`)
			}
			run(t, testutil.Case{Name: tc.name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, e := sdk.GraphAPI.ChannelUserManagement.Add(ctx, teamID, "fixture", userID, management.ChangeOptions{Role: management.Owner, TenantID: tc.tenant})

				return testutil.Observation{Value: map[string]string{"membership": string(v.MembershipID)}, Response: v.Response, Err: e}
			}, Want: testutil.Expected{Status: status, JSON: want, Error: tc.err}})
		})
	}
}
