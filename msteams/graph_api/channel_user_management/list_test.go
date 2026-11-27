package channelusermanagement_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_user_management"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
)

func usersPath(ids ...string) string {
	clauses := make([]string, len(ids))
	for i, id := range ids {
		clauses[i] = "id eq '" + strings.ToLower(id) + "'"
	}
	return "/beta/users?" + url.Values{"$select": {"id,userPrincipalName"}, "$filter": {strings.Join(clauses, " or ")}}.Encode()
}

func listed(value management.ListResult, err error) testutil.Observation {
	rows := []map[string]any{}
	for _, v := range value.Items {
		rows = append(rows, map[string]any{"id": v.MembershipID, "user_id": v.UserID, "user": v.User, "name": v.Name, "role": v.Role, "tenant": v.TenantID})
	}
	return testutil.Observation{Value: map[string]any{"complete": value.Complete, "members": len(value.Members), "items": rows}, Response: value.Response, Err: err}
}

func TestUnitListRoleOrderPresenceAndPagedEnrichment(t *testing.T) {
	t.Parallel()
	ownerID, guestID := teamID, tenantID

	query := usersPath(ownerID, userID, guestID)
	members := `{"value":[{"id":"guest","userId":"` + guestID + `","roles":["guest"],"tenantId":"` + tenantID + `"},{"id":"member","userId":"` + userID + `","email":"old@fixture.invalid","roles":[]},{"id":"owner","userId":"` + ownerID + `","displayName":"Owner","roles":["owner"]},{"id":"future","userId":null,"email":"unknown@fixture.invalid","roles":["future"]}]}`

	run(t, testutil.Case{Name: "list-role-order-paged-upns",
		Steps: []testutil.Exchange{channelRead("standard"), wire("GET", membersPath(), "", 200, members), wire("GET", query, "", 200, `{"value":[{"id":"`+ownerID+`","userPrincipalName":"owner@fixture.invalid"},{"id":"`+userID+`","userPrincipalName":null}],"@odata.nextLink":"https://graph.microsoft.com/beta/users?$skiptoken=opaque%2Bvalue"}`), wire("GET", "/beta/users?$skiptoken=opaque%2Bvalue", "", 200, `{"value":[{"id":"`+guestID+`","userPrincipalName":"guest@external.invalid"}]}`)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			return listed(sdk.GraphAPI.ChannelUserManagement.List(ctx, teamID, "fixture", management.ListOptions{}))
		},
		Want: testutil.Expected{Status: 200, JSON: json.RawMessage(`{"complete":true,"members":4,"items":[` +
			`{"id":"owner","user_id":"` + ownerID + `","user":"owner@fixture.invalid","name":"Owner","role":"owner","tenant":null},` +
			`{"id":"member","user_id":"` + userID + `","user":null,"name":null,"role":"member","tenant":null},` +
			`{"id":"future","user_id":null,"user":"unknown@fixture.invalid","name":null,"role":"member","tenant":null},` +
			`{"id":"guest","user_id":"` + guestID + `","user":"guest@external.invalid","name":null,"role":"guest","tenant":"` + tenantID + `"}]}`)}})
}

func TestUnitRoleFilterPrecedesEnrichment(t *testing.T) {
	t.Parallel()
	members := `{"value":[{"id":"owner","userId":"` + teamID + `","roles":["owner"]},{"id":"member","userId":"` + userID + `","roles":[]},{"id":"guest","userId":"` + tenantID + `","roles":["guest"]}]}`

	for _, tc := range []struct{ name, role, id string }{
		{"owner", "OwNeR", teamID}, {"member", "Member", userID}, {"guest", "GUEST", tenantID},
	} {
		t.Run(tc.name, func(t *testing.T) {

			want := fmt.Sprintf(`{"complete":true,"members":3,"items":[{"id":%q,"user_id":%q,"user":null,"name":null,"role":%q,"tenant":null}]}`, tc.name, tc.id, tc.name)
			run(t, testutil.Case{Name: "role-" + tc.name, Steps: []testutil.Exchange{channelRead("private"), wire("GET", membersPath(), "", 200, members), wire("GET", usersPath(tc.id), "", 200, `{"value":[{"id":"`+tc.id+`","userPrincipalName":null}]}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				return listed(sdk.GraphAPI.ChannelUserManagement.List(ctx, teamID, "fixture", management.ListOptions{Role: management.Role(tc.role)}))
			}, Want: testutil.Expected{Status: 200, JSON: json.RawMessage(want)}})
		})
	}
}

func TestUnitEnrichmentChunksDistinctIDsAndStillRefreshesExistingUPNs(t *testing.T) {
	t.Parallel()
	ids := make([]string, 11)
	rows := make([]string, 12)
	for i := range ids {
		ids[i] = fmt.Sprintf("00000000-0000-4000-8000-%012x", i+1)
		rows[i] = fmt.Sprintf(`{"id":"membership-%d","userId":%q,"userPrincipalName":"existing@fixture.invalid","roles":[]}`, i, ids[i])
	}
	rows[11] = fmt.Sprintf(`{"id":"duplicate-user-membership","userId":%q,"userPrincipalName":"existing@fixture.invalid","roles":[]}`, ids[0])
	run(t, testutil.Case{Name: "enrich-eleven-distinct-twelve-members", Steps: []testutil.Exchange{channelRead("shared"), wire("GET", membersPath(), "", 200, `{"value":[`+strings.Join(rows, ",")+`]}`), wire("GET", usersPath(ids[:10]...), "", 200, `{"value":[]}`), wire("GET", usersPath(ids[10]), "", 200, `{"value":[]}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		value, err := sdk.GraphAPI.ChannelUserManagement.List(ctx, teamID, "fixture", management.ListOptions{})

		for _, row := range value.Items {
			if row.User == nil || *row.User != "existing@fixture.invalid" {
				t.Error("missing host directory user did not retain initial UPN")
			}
		}
		return testutil.Observation{Value: map[string]any{"complete": value.Complete, "items": len(value.Items)}, Response: value.Response, Err: err}
	}, Want: testutil.Expected{Status: 200, JSON: json.RawMessage(`{"complete":true,"items":12}`)}})
}

func TestUnitEnrichmentFailureNeverProducesACompleteList(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, body string
		status     int
		maxPages   int
		err        error
	}{
		{"forbidden", `{"error":{"code":"Denied"}}`, 403, 0, auth.ErrForbidden},
		{"malformed", `{"value":null}`, 200, 0, client.ErrDecode},
		{"unexpected-status", `{"value":[]}`, 202, 0, client.ErrPageStatus},
		{"duplicate-users", `{"value":[{"id":"` + userID + `"},{"id":"` + userID + `"}]}`, 200, 0, directoryuser.ErrAmbiguous},
		{"unrequested-user", `{"value":[{"id":"` + teamID + `"}]}`, 200, 0, client.ErrDecode},
		{"foreign-continuation", `{"value":[],"@odata.nextLink":"https://foreign.invalid/users"}`, 200, 0, client.ErrDestination},
		{"bounded-continuation", `{"value":[],"@odata.nextLink":"https://graph.microsoft.com/beta/users?$skiptoken=next"}`, 200, 1, client.ErrPageLimit},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: "enrich-" + tc.name, Steps: []testutil.Exchange{channelRead("private"), roster(), wire("GET", usersPath(userID), "", tc.status, tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				value, err := sdk.GraphAPI.ChannelUserManagement.List(ctx, teamID, "fixture", management.ListOptions{UserLookupMaxPages: tc.maxPages})

				if value.Complete || len(value.Members) != 1 || len(value.Stages) != 3 || value.Stages[2].State != management.Unconfirmed {
					t.Error("failed enrichment lost native state or appeared complete")
				}
				return testutil.Observation{Value: map[string]any{"complete": value.Complete, "members": len(value.Members)}, Response: value.Response, Err: err}
			}, Want: testutil.Expected{Status: tc.status, Error: tc.err, JSON: json.RawMessage(`{"complete":false,"members":1}`)}})
		})
	}
}

func TestUnitEmptyFilteredListDoesNotQueryDirectoryAndInvalidOptionsAreLocal(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "empty-filter", Steps: []testutil.Exchange{channelRead("private"), roster()}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		return listed(sdk.GraphAPI.ChannelUserManagement.List(ctx, teamID, "fixture", management.ListOptions{Role: management.Owner}))
	}, Want: testutil.Expected{Status: 200, JSON: json.RawMessage(`{"complete":true,"members":1,"items":[]}`)}})
	for i, options := range []management.ListOptions{{Role: "future"}, {UserLookupMaxPages: -1}, {Traversal: client.PagingOptions{MaxPages: -1}}} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			expected := client.ErrRequest
			if i == 2 {
				expected = client.ErrPagination
			}
			run(t, testutil.Case{Name: "invalid-list-" + strconv.Itoa(i), Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				return listed(sdk.GraphAPI.ChannelUserManagement.List(ctx, teamID, "fixture", options))
			}, Want: testutil.Expected{Error: expected}})
		})
	}
}
