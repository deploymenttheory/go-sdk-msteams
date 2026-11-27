package sharedchannelmanagement_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"

	"strconv"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	associated "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/associated_team"
	directory "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/shared_channel_management"
)

const (
	hostID    = "11111111-1111-4111-8111-111111111111"
	targetID  = "22222222-2222-4222-8222-222222222222"
	userID    = "33333333-3333-4333-8333-333333333333"
	channelID = "19:fixture/opaque@thread.tacv2"
)

func base() string {
	return "/beta/teams/" + hostID + "/channels/" + url.PathEscape(channelID) + "/sharedWithTeams"
}
func allowed() string { return base() + "/" + targetID + "/allowedMembers" }
func wire(method, path string, status int, body string) testutil.Exchange {
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: auth.GraphAudience + path, Audience: auth.GraphAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(body)}}
}
func run(t *testing.T, c testutil.Case) testutil.Result {
	t.Helper()

	r, err := testutil.Run(t.Context(), c)
	if err != nil || !r.Passed() {
		t.Fatalf("shared-channel composition differs: setup=%v differences=%+v", err, r.Differences())
	}
	return r
}

func usersPath(ids ...string) string {
	clauses := make([]string, len(ids))
	for i, id := range ids {
		clauses[i] = "id eq '" + strings.ToLower(id) + "'"
	}
	return "/beta/users?" + url.Values{"$select": {"id,userPrincipalName"}, "$filter": {strings.Join(clauses, " or ")}}.Encode()
}
func observed(v management.AllowedUsersResult, err error) testutil.Observation {
	rows := []map[string]any{}
	for _, row := range v.Items {
		rows = append(rows, map[string]any{"id": row.MembershipID, "uid": row.UserID, "user": row.User, "role": row.Role, "tenant": row.TenantID})
	}
	return testutil.Observation{Value: map[string]any{"complete": v.Complete, "members": len(v.Members), "items": rows}, Response: v.Response, Err: err}
}

func TestUnitSharedTeamOptionalSelectorAndUnshare(t *testing.T) {
	t.Parallel()

	for _, single := range []bool{false, true} {
		name, path := "collection", base()
		options := management.SharedTeamsOptions{}
		if single {
			name, path = "single", base()+"/"+targetID
			options.SharedTeamID = targetID
		}
		body := `{"value":[{"id":"` + targetID + `","isHostTeam":false}]}`
		if single {
			body = `{"id":"` + targetID + `","isHostTeam":false}`
		}
		run(t, testutil.Case{Name: name, Steps: []testutil.Exchange{wire("GET", path, 200, body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, err := sdk.GraphAPI.SharedChannelManagement.ListSharedWithTeams(ctx, hostID, channelID, options)

			return testutil.Observation{Value: map[string]any{"items": len(v.Items), "pages": v.Pages, "target": v.Items[0].ID, "host": v.Items[0].IsHostTeam}, Response: v.Response, Err: err}
		}, Want: testutil.Expected{Status: 200, JSON: json.RawMessage(`{"items":1,"pages":1,"target":"` + targetID + `","host":false}`)}})
	}
	run(t, testutil.Case{Name: "unshare-once", Steps: []testutil.Exchange{wire("DELETE", base()+"/"+targetID, 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		r, err := sdk.GraphAPI.SharedChannelManagement.UnshareWithTeam(ctx, hostID, channelID, targetID)

		return testutil.Observation{Response: r, Err: err}
	}, Want: testutil.Expected{Status: 204}})
}

func TestUnitAssociatedTeamsSelectionAndResolution(t *testing.T) {
	t.Parallel()
	userRoute := "/beta/users/" + userID + "/teamwork/associatedTeams"
	direct := "/v1.0/users/fixture@invalid?" + url.Values{"$select": {"id"}}.Encode()
	fallback := "/v1.0/users?" + url.Values{"$select": {"id"}, "$top": {"2"}, "$filter": {"mail eq 'fixture@invalid'"}}.Encode()

	for _, tc := range []struct {
		name, user, path, source string
		mode                     auth.Mode
		reads                    []testutil.Exchange
	}{
		{"explicit-id", userID, userRoute, "identifier", auth.Application, nil},
		{"explicit-upn", "fixture@invalid", userRoute, "direct", auth.Application, []testutil.Exchange{wire("GET", direct, 200, `{"id":"`+userID+`"}`)}},
		{"canonical-mail-fallback", "fixture@invalid", userRoute, "mail", auth.Application, []testutil.Exchange{wire("GET", direct, 404, `{"error":{"code":"NotFound"}}`), wire("GET", fallback, 200, `{"value":[{"id":"`+userID+`"}]}`)}},
		{"delegated-me", "", "/beta/me/teamwork/associatedTeams", "", auth.Delegated, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			steps := append(tc.reads, wire("GET", tc.path, 200, `{"value":[{"id":"`+hostID+`","tenantId":"`+targetID+`","displayName":"Associated fixture","future":false}]}`))
			run(t, testutil.Case{Name: tc.name, Mode: tc.mode, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, err := sdk.GraphAPI.SharedChannelManagement.ListAssociatedTeams(ctx, tc.user, management.AssociatedOptions{})

				return testutil.Observation{Value: map[string]any{"items": len(v.Items), "complete": v.Complete, "source": v.ResolutionSource, "id": v.Items[0].ID, "tenant": v.Items[0].TenantID, "future": v.Items[0].RawFields["future"]}, Response: v.Response, Err: err}
			}, Want: testutil.Expected{Status: 200, JSON: json.RawMessage(fmt.Sprintf(`{"items":1,"complete":true,"source":%q,"id":"`+hostID+`","tenant":"`+targetID+`","future":false}`, tc.source))}})
		})
	}
	run(t, testutil.Case{Name: "application-me-local", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		_, err := sdk.GraphAPI.SharedChannelManagement.ListAssociatedTeams(ctx, "", management.AssociatedOptions{})

		return testutil.Observation{Err: err}
	}, Want: testutil.Expected{Error: associated.ErrCurrentUser}})
}

func TestUnitAllowedUsersProjectionFilterAndNullableEnrichment(t *testing.T) {
	t.Parallel()
	members := `{"value":[{"id":"owner","userId":"` + hostID + `","email":"old-owner@invalid","roles":["owner"]},{"id":"member","userId":"` + userID + `","email":"old-member@invalid","roles":[]},{"id":"future","userId":null,"email":"future@invalid","roles":["future"],"tenantId":"external"}]}`

	run(t, testutil.Case{Name: "allowed-users-projection", Steps: []testutil.Exchange{wire("GET", allowed(), 200, members), wire("GET", usersPath(hostID, userID), 200, `{"value":[{"id":"`+hostID+`","userPrincipalName":"owner@invalid"},{"id":"`+userID+`","userPrincipalName":null}]}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		return observed(sdk.GraphAPI.SharedChannelManagement.ListAllowedUsers(ctx, hostID, channelID, targetID, management.AllowedUsersOptions{}))
	}, Want: testutil.Expected{Status: 200, JSON: json.RawMessage(`{"complete":true,"members":3,"items":[{"id":"owner","uid":"` + hostID + `","user":"owner@invalid","role":"owner","tenant":null},{"id":"member","uid":"` + userID + `","user":null,"role":"member","tenant":null},{"id":"future","uid":null,"user":"future@invalid","role":"member","tenant":"external"}]}`)}})
	run(t, testutil.Case{Name: "filter-before-enrichment", Steps: []testutil.Exchange{wire("GET", allowed(), 200, members), wire("GET", usersPath(hostID), 200, `{"value":[{"id":"`+hostID+`","userPrincipalName":"owner@invalid"}]}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, err := sdk.GraphAPI.SharedChannelManagement.ListAllowedUsers(ctx, hostID, channelID, targetID, management.AllowedUsersOptions{Role: "OwNeR"})

		return testutil.Observation{Value: map[string]any{"complete": v.Complete, "items": len(v.Items), "members": len(v.Members)}, Response: v.Response, Err: err}
	}, Want: testutil.Expected{Status: 200, JSON: json.RawMessage(`{"complete":true,"items":1,"members":3}`)}})
}

func TestUnitAllowedUserFailuresRetainNativeState(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name, body string
		status     int
		err        error
	}{
		{"denied", `{"error":{"code":"Denied"}}`, 403, auth.ErrForbidden},
		{"duplicate", `{"value":[{"id":"` + userID + `"},{"id":"` + userID + `"}]}`, 200, directory.ErrAmbiguous},
		{"unexpected-user", `{"value":[{"id":"` + hostID + `"}]}`, 200, client.ErrDecode},
		{"malformed", `{"value":null}`, 200, client.ErrDecode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: "enrich-" + tc.name, Steps: []testutil.Exchange{wire("GET", allowed(), 200, `{"value":[{"id":"opaque","userId":"`+userID+`","email":"fallback@invalid"}]}`), wire("GET", usersPath(userID), tc.status, tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				return observed(sdk.GraphAPI.SharedChannelManagement.ListAllowedUsers(ctx, hostID, channelID, targetID, management.AllowedUsersOptions{}))
			}, Want: testutil.Expected{Status: tc.status, Error: tc.err, JSON: json.RawMessage(`{"complete":false,"members":1,"items":[{"id":"opaque","uid":"` + userID + `","user":"fallback@invalid","role":"member","tenant":null}]}`)}})
		})
	}
}

func TestUnitCompositionsRejectOptionsAndCancellationBeforeCredentials(t *testing.T) {
	t.Parallel()
	for i, options := range []management.AllowedUsersOptions{{Role: management.Guest}, {Role: "future"}, {UserLookupMaxPages: -1}, {Traversal: client.PagingOptions{MaxPages: -1}}} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			want := client.ErrRequest
			if i == 3 {
				want = client.ErrPagination
			}
			run(t, testutil.Case{Name: "bad-allowed-options-" + strconv.Itoa(i), Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				_, err := sdk.GraphAPI.SharedChannelManagement.ListAllowedUsers(ctx, hostID, channelID, targetID, options)

				return testutil.Observation{Err: err}
			}, Want: testutil.Expected{Error: want}})
		})
	}
	for i, options := range []management.AssociatedOptions{{ResolveMaxPages: -1}, {Traversal: client.PagingOptions{MaxPages: -1}}, {Traversal: client.PagingOptions{MaxItems: -1}}, {Traversal: client.PagingOptions{Style: client.OffsetPaging}}} {
		run(t, testutil.Case{Name: "bad-associated-options-" + strconv.Itoa(i), Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			_, err := sdk.GraphAPI.SharedChannelManagement.ListAssociatedTeams(ctx, "fixture@invalid", options)

			return testutil.Observation{Err: err}
		}, Want: testutil.Expected{Error: client.ErrPagination}})
	}
	run(t, testutil.Case{Name: "cancelled", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		cancelled, stop := context.WithCancel(ctx)
		stop()
		_, err := sdk.GraphAPI.SharedChannelManagement.UnshareWithTeam(cancelled, hostID, channelID, targetID)

		return testutil.Observation{Err: err}
	}, Want: testutil.Expected{Error: context.Canceled}})
}
