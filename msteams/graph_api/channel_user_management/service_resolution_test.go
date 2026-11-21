package channelusermanagement_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	member "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_member"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_user_management"
)

func TestPrivateUPNFailurePreventsMembershipReadsAndWrites(t *testing.T) {
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
			run(t, contracttest.Case{Name: "upn-" + tc.name, Steps: []contracttest.Exchange{channelRead("private"), wire("GET", "/v1.0/users/member@fixture.invalid?$select=id", "", tc.status, tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return changed(sdk.GraphAPI.ChannelUserManagement.Add(ctx, teamID, "fixture", "member@fixture.invalid", management.ChangeOptions{Role: management.Owner}))
			}, Want: contracttest.Expected{Status: tc.status, Error: tc.err}})
		})
	}
}

func TestSharedNullEmailAndTenantSelectionUseReturnedMembership(t *testing.T) {
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
			steps := []contracttest.Exchange{channelRead("shared"), wire("GET", membersPath(), "", 200, tc.body)}
			status := 200
			var want json.RawMessage
			if tc.err == nil {
				steps = append(steps, wire("PATCH", membersPath()+"/"+tc.id, `{"roles":["owner"]}`, 200, `{"id":"`+tc.id+`","roles":["owner"]}`))
				want = json.RawMessage(`{"membership":"` + tc.id + `"}`)
			}
			run(t, contracttest.Case{Name: tc.name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, e := sdk.GraphAPI.ChannelUserManagement.Add(ctx, teamID, "fixture", userID, management.ChangeOptions{Role: management.Owner, TenantID: tc.tenant})
				return contracttest.Observation{Value: map[string]string{"membership": string(v.MembershipID)}, Response: v.Response, Err: e}
			}, Want: contracttest.Expected{Status: status, JSON: want, Error: tc.err}})
		})
	}
}
