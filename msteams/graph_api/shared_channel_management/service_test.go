package sharedchannelmanagement_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
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
func wire(method, path string, status int, body string) contracttest.Exchange {
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: auth.GraphAudience + path, Audience: auth.GraphAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(body)}}
}
func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family, c.Source = "shared_channel_management", "contracts/endpoints/8.1.0/core.json"
	if c.Scope == "" {
		c.Scope = "service"
	}
	r, err := contracttest.Run(t.Context(), c)
	if err != nil || !r.Passed() {
		t.Fatalf("shared-channel composition differs: setup=%v differences=%+v", err, r.Differences())
	}
	return r
}
func mapping(command string) []contracttest.Mapping {
	return []contracttest.Mapping{{Command: command, ParameterSet: "__AllParameterSets"}}
}
func evidence(t *testing.T, results []contracttest.Result) {
	t.Helper()
	rev := os.Getenv("MSTEAMS_CONTRACT_REVISION")
	if rev == "" {
		return
	}
	root, err := os.OpenRoot(t.ArtifactDir())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	f, err := root.Create("shared-channel-management-service.json")
	if err != nil {
		t.Fatal(err)
	}
	writeErr, closeErr := contracttest.WriteEvidence(f, rev, results), f.Close()
	if writeErr != nil || closeErr != nil {
		t.Fatal("shared-channel composition evidence write failed")
	}
}
func usersPath(ids ...string) string {
	clauses := make([]string, len(ids))
	for i, id := range ids {
		clauses[i] = "id eq '" + strings.ToLower(id) + "'"
	}
	return "/beta/users?" + url.Values{"$select": {"id,userPrincipalName"}, "$filter": {strings.Join(clauses, " or ")}}.Encode()
}
func observed(v management.AllowedUsersResult, err error) contracttest.Observation {
	rows := []map[string]any{}
	for _, row := range v.Items {
		rows = append(rows, map[string]any{"id": row.MembershipID, "uid": row.UserID, "user": row.User, "role": row.Role, "tenant": row.TenantID})
	}
	return contracttest.Observation{Value: map[string]any{"complete": v.Complete, "members": len(v.Members), "items": rows}, Response: v.Response, Err: err}
}

func TestSharedTeamOptionalSelectorAndUnshare(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, single := range []bool{false, true} {
		name, path, body := "collection", base(), `{"value":[{"id":"`+targetID+`","isHostTeam":false}]}`
		options := management.SharedTeamsOptions{}
		if single {
			name, path, body = "single", base()+"/"+targetID, `{"id":"`+targetID+`","isHostTeam":false}`
			options.SharedTeamID = targetID
		}
		results = append(results, run(t, contracttest.Case{Name: name, Mappings: mapping("Get-SharedWithTeam"), Steps: []contracttest.Exchange{wire("GET", path, 200, body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, err := sdk.GraphAPI.SharedChannelManagement.ListSharedWithTeams(ctx, hostID, channelID, options)
			return contracttest.Observation{Value: map[string]any{"items": len(v.Items), "pages": v.Pages, "target": v.Items[0].ID, "host": v.Items[0].IsHostTeam}, Response: v.Response, Err: err}
		}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"items":1,"pages":1,"target":"` + targetID + `","host":false}`)}}))
	}
	results = append(results, run(t, contracttest.Case{Name: "unshare-once", Mappings: mapping("Remove-SharedWithTeam"), Steps: []contracttest.Exchange{wire("DELETE", base()+"/"+targetID, 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		r, err := sdk.GraphAPI.SharedChannelManagement.UnshareWithTeam(ctx, hostID, channelID, targetID)
		return contracttest.Observation{Response: r, Err: err}
	}, Want: contracttest.Expected{Status: 204}}))
	evidence(t, results)
}

func TestAssociatedTeamsSelectionAndResolution(t *testing.T) {
	t.Parallel()
	userRoute := "/beta/users/" + userID + "/teamwork/associatedTeams"
	direct := "/v1.0/users/fixture@invalid?" + url.Values{"$select": {"id"}}.Encode()
	fallback := "/v1.0/users?" + url.Values{"$select": {"id"}, "$top": {"2"}, "$filter": {"mail eq 'fixture@invalid'"}}.Encode()
	var results []contracttest.Result
	for _, tc := range []struct {
		name, user, path, source string
		mode                     auth.Mode
		reads                    []contracttest.Exchange
	}{
		{"explicit-id", userID, userRoute, "identifier", auth.Application, nil},
		{"explicit-upn", "fixture@invalid", userRoute, "direct", auth.Application, []contracttest.Exchange{wire("GET", direct, 200, `{"id":"`+userID+`"}`)}},
		{"canonical-mail-fallback", "fixture@invalid", userRoute, "mail", auth.Application, []contracttest.Exchange{wire("GET", direct, 404, `{"error":{"code":"NotFound"}}`), wire("GET", fallback, 200, `{"value":[{"id":"`+userID+`"}]}`)}},
		{"delegated-me", "", "/beta/me/teamwork/associatedTeams", "", auth.Delegated, nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			steps := append(tc.reads, wire("GET", tc.path, 200, `{"value":[{"id":"`+targetID+`","tenantId":"external-tenant"}]}`))
			results = append(results, run(t, contracttest.Case{Name: tc.name, Mode: tc.mode, Mappings: mapping("Get-AssociatedTeam"), Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, err := sdk.GraphAPI.SharedChannelManagement.ListAssociatedTeams(ctx, tc.user, management.AssociatedOptions{})
				return contracttest.Observation{Value: map[string]any{"items": len(v.Items), "complete": v.Complete, "source": v.ResolutionSource}, Response: v.Response, Err: err}
			}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(fmt.Sprintf(`{"items":1,"complete":true,"source":%q}`, tc.source))}}))
		})
	}
	results = append(results, run(t, contracttest.Case{Name: "application-me-local", Scope: "local", Mappings: mapping("Get-AssociatedTeam"), Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, err := sdk.GraphAPI.SharedChannelManagement.ListAssociatedTeams(ctx, "", management.AssociatedOptions{})
		return contracttest.Observation{Err: err}
	}, Want: contracttest.Expected{Error: associated.ErrCurrentUser}}))
	evidence(t, results)
}

func TestAllowedUsersProjectionFilterAndNullableEnrichment(t *testing.T) {
	t.Parallel()
	roster := `{"value":[{"id":"member","userId":"` + userID + `","email":"fallback@invalid","roles":[]},{"id":"owner","userId":"` + hostID + `","roles":["guest","OWNER"]},{"id":"future","email":"future@invalid","roles":["future"],"tenantId":"external"}]}`
	result := run(t, contracttest.Case{Name: "allowed-users-projection", Mappings: mapping("Get-SharedWithTeamUser"), Steps: []contracttest.Exchange{wire("GET", allowed(), 200, roster), wire("GET", usersPath(hostID, userID), 200, `{"value":[{"id":"`+hostID+`","userPrincipalName":"owner@invalid"},{"id":"`+userID+`","userPrincipalName":null}]}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		return observed(sdk.GraphAPI.SharedChannelManagement.ListAllowedUsers(ctx, hostID, channelID, targetID, management.AllowedUsersOptions{}))
	}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"complete":true,"members":3,"items":[{"id":"owner","uid":"` + hostID + `","user":"owner@invalid","role":"owner","tenant":null},{"id":"member","uid":"` + userID + `","user":null,"role":"member","tenant":null},{"id":"future","uid":null,"user":"future@invalid","role":"member","tenant":"external"}]}`)}})
	filtered := run(t, contracttest.Case{Name: "filter-before-enrichment", Mappings: mapping("Get-SharedWithTeamUser"), Steps: []contracttest.Exchange{wire("GET", allowed(), 200, roster), wire("GET", usersPath(hostID), 200, `{"value":[]}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, err := sdk.GraphAPI.SharedChannelManagement.ListAllowedUsers(ctx, hostID, channelID, targetID, management.AllowedUsersOptions{Role: "OwNeR"})
		return contracttest.Observation{Value: map[string]any{"complete": v.Complete, "items": len(v.Items), "members": len(v.Members)}, Response: v.Response, Err: err}
	}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"complete":true,"items":1,"members":3}`)}})
	evidence(t, []contracttest.Result{result, filtered})
}

func TestAllowedUserFailuresRetainNativeState(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
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
			results = append(results, run(t, contracttest.Case{Name: "enrich-" + tc.name, Mappings: mapping("Get-SharedWithTeamUser"), Steps: []contracttest.Exchange{wire("GET", allowed(), 200, `{"value":[{"id":"opaque","userId":"`+userID+`","email":"fallback@invalid"}]}`), wire("GET", usersPath(userID), tc.status, tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return observed(sdk.GraphAPI.SharedChannelManagement.ListAllowedUsers(ctx, hostID, channelID, targetID, management.AllowedUsersOptions{}))
			}, Want: contracttest.Expected{Status: tc.status, Error: tc.err, JSON: json.RawMessage(`{"complete":false,"members":1,"items":[{"id":"opaque","uid":"` + userID + `","user":"fallback@invalid","role":"member","tenant":null}]}`)}}))
		})
	}
	evidence(t, results)
}

func TestCompositionsRejectOptionsAndCancellationBeforeCredentials(t *testing.T) {
	t.Parallel()
	for i, options := range []management.AllowedUsersOptions{{Role: management.Guest}, {Role: "future"}, {UserLookupMaxPages: -1}, {Traversal: client.PagingOptions{MaxPages: -1}}} {
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			want := client.ErrRequest
			if i == 3 {
				want = client.ErrPagination
			}
			run(t, contracttest.Case{Name: "bad-allowed-options-" + strconv.Itoa(i), Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				_, err := sdk.GraphAPI.SharedChannelManagement.ListAllowedUsers(ctx, hostID, channelID, targetID, options)
				return contracttest.Observation{Err: err}
			}, Want: contracttest.Expected{Error: want}})
		})
	}
	for i, options := range []management.AssociatedOptions{{ResolveMaxPages: -1}, {Traversal: client.PagingOptions{MaxPages: -1}}, {Traversal: client.PagingOptions{MaxItems: -1}}, {Traversal: client.PagingOptions{Style: client.OffsetPaging}}} {
		run(t, contracttest.Case{Name: "bad-associated-options-" + strconv.Itoa(i), Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, err := sdk.GraphAPI.SharedChannelManagement.ListAssociatedTeams(ctx, "fixture@invalid", options)
			return contracttest.Observation{Err: err}
		}, Want: contracttest.Expected{Error: client.ErrPagination}})
	}
	run(t, contracttest.Case{Name: "cancelled", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		cancelled, stop := context.WithCancel(ctx)
		stop()
		_, err := sdk.GraphAPI.SharedChannelManagement.UnshareWithTeam(cancelled, hostID, channelID, targetID)
		return contracttest.Observation{Err: err}
	}, Want: contracttest.Expected{Error: context.Canceled}})
}
