package teamusermanagement_test

import (
	"context"
	"encoding/json"
	"net/http"

	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_user_management"
)

const (
	groupID = "11111111-1111-4111-8111-111111111111"
	userID  = "22222222-2222-4222-8222-222222222222"
	otherID = "33333333-3333-4333-8333-333333333333"
	guestID = "44444444-4444-4444-8444-444444444444"
)

func runCase(t *testing.T, command string, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family = "team_user_management"
	if c.Name == "" {
		c.Name = "composition-case"
	}
	if c.Scope == "" {
		c.Scope = "service"
	}
	c.Source = "msteams/graph_api/team_user_management/service_test.go"
	c.Mode = auth.Delegated
	c.Mappings = []contracttest.Mapping{{Command: command, ParameterSet: "__AllParameterSets"}}
	result, err := contracttest.Run(t.Context(), requestBoundary(t, c))
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Errorf("team-user composition differs: %+v", result.Differences())
	}
	return result
}

func wire(method, path, body string, status int, response string) contracttest.Exchange {
	var payload json.RawMessage
	if body != "" {
		payload = json.RawMessage(body)
	}
	headers := http.Header{"Accept": {"application/json"}}
	if body != "" {
		headers.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: auth.GraphAudience + path, Audience: auth.GraphAudience, Headers: headers, JSON: payload}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(response)}}
}

func read(kind, body string) contracttest.Exchange {
	return wire("GET", "/v1.0/groups/"+groupID+"/"+kind+"?$select=id", "", 200, body)
}
func add(kind string, status int) contracttest.Exchange {
	collection := "directoryObjects"
	if kind == "owners" {
		collection = "users"
	}
	return wire("POST", "/v1.0/groups/"+groupID+"/"+kind+"/$ref", `{"@odata.id":"`+auth.GraphAudience+`/v1.0/`+collection+`/`+userID+`"}`, status, "")
}
func remove(kind string, status int) contracttest.Exchange {
	return wire("DELETE", "/v1.0/groups/"+groupID+"/"+kind+"/"+userID+"/$ref", "", status, "")
}

func changed(result management.ChangeResult, err error) contracttest.Observation {
	stages := []string{}
	for _, stage := range result.Stages {
		stages = append(stages, string(stage.Action)+":"+string(stage.State))
	}
	return contracttest.Observation{Value: map[string]any{"id": result.UserID, "resolution": result.ResolutionSource, "stages": stages}, Response: result.Response, Err: err}
}

func TestUnitAddCompositionStagesAndCorrectedPreflight(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		role   management.Role
		steps  []contracttest.Exchange
		stages string
		status int
		want   error
	}{
		{"empty-preflight-adds-member", "", []contracttest.Exchange{read("members", `{"value":[]}`), add("members", 204)}, `["resolve-user:succeeded","read-members:succeeded","add-member-reference:succeeded"]`, 204, nil},
		{"present-member-skips-post", management.Member, []contracttest.Exchange{read("members", `{"value":[{"id":"`+userID+`"}]}`)}, `["resolve-user:succeeded","read-members:succeeded","add-member-reference:skipped"]`, 200, nil},
		{"owner-adds-member-before-owner", management.Role("OWNER"), []contracttest.Exchange{read("members", `{"value":[]}`), add("members", 204), read("owners", `{"value":[]}`), add("owners", 204)}, `["resolve-user:succeeded","read-members:succeeded","add-member-reference:succeeded","read-owners:succeeded","add-owner-reference:succeeded"]`, 204, nil},
		{"existing-both-skips-writes", management.Owner, []contracttest.Exchange{read("members", `{"value":[{"id":"`+userID+`"}]}`), read("owners", `{"value":[{"id":"`+userID+`"}]}`)}, `["resolve-user:succeeded","read-members:succeeded","add-member-reference:skipped","read-owners:succeeded","add-owner-reference:skipped"]`, 200, nil},
		{"forbidden-preflight-no-write", "", []contracttest.Exchange{wire("GET", "/v1.0/groups/"+groupID+"/members?$select=id", "", 403, `{"error":{"code":"Forbidden"}}`)}, `["resolve-user:succeeded","read-members:unconfirmed"]`, 403, auth.ErrForbidden},
		{"member-survives-owner-failure", management.Owner, []contracttest.Exchange{read("members", `{"value":[]}`), add("members", 204), read("owners", `{"value":[]}`), add("owners", 403)}, `["resolve-user:succeeded","read-members:succeeded","add-member-reference:succeeded","read-owners:succeeded","add-owner-reference:unconfirmed"]`, 403, auth.ErrForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runCase(t, "Add-TeamUser", contracttest.Case{Name: tc.name, Steps: tc.steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return changed(sdk.GraphAPI.TeamUserManagement.Add(ctx, groupID, userID, management.ChangeOptions{Role: tc.role}))
			},
				Want: contracttest.Expected{Status: tc.status, Error: tc.want, JSON: json.RawMessage(`{"id":"` + userID + `","resolution":"identifier","stages":` + tc.stages + `}`)}})
		})
	}
}

func TestUnitAddCompletePreflightAndResolvedIdentity(t *testing.T) {
	t.Parallel()
	next := auth.GraphAudience + "/v1.0/groups/" + groupID + "/members?$skiptoken=next"
	for _, tc := range []struct {
		name   string
		limit  int
		second bool
		want   error
		stages string
	}{
		{"later-member-skips", 0, true, nil, `["resolve-user:succeeded","read-members:succeeded","add-member-reference:skipped"]`},
		{"partial-cannot-establish-absence", 1, false, client.ErrPageLimit, `["resolve-user:succeeded","read-members:unconfirmed"]`},
	} {
		steps := []contracttest.Exchange{read("members", `{"value":[],"@odata.nextLink":"`+next+`"}`)}
		if tc.second {
			steps = append(steps, wire("GET", "/v1.0/groups/"+groupID+"/members?$skiptoken=next", "", 200, ""))
		}
		runCase(t, "Add-TeamUser", contracttest.Case{Name: tc.name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			return changed(sdk.GraphAPI.TeamUserManagement.Add(ctx, groupID, userID, management.ChangeOptions{Traversal: client.PagingOptions{MaxPages: tc.limit}}))
		},
			Want: contracttest.Expected{Status: 200, Error: tc.want, JSON: json.RawMessage(`{"id":"` + userID + `","resolution":"identifier","stages":` + tc.stages + `}`)}})
	}
	runCase(t, "Add-TeamUser", contracttest.Case{Name: "resolved-upn", Steps: []contracttest.Exchange{wire("GET", "/v1.0/users/fixture@example.invalid?$select=id", "", 200, ""), read("members", `{"value":[{"id":"`+userID+`"}]}`)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			return changed(sdk.GraphAPI.TeamUserManagement.Add(ctx, groupID, "fixture@example.invalid", management.ChangeOptions{}))
		},
		Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"id":"` + userID + `","resolution":"direct","stages":["resolve-user:succeeded","read-members:succeeded","add-member-reference:skipped"]}`)}})
}

func TestUnitRemoveCompositionPreservesLastOwnerAndPartialWrites(t *testing.T) {
	t.Parallel()

	owners := `{"value":[{"id":"` + userID + `"},{"id":"` + otherID + `"}]}`
	for _, tc := range []struct {
		name   string
		role   management.Role
		steps  []contracttest.Exchange
		stages string
		status int
		want   error
	}{
		{"last-owner-blocks-all-deletes", "", []contracttest.Exchange{read("owners", `{"value":[{"id":"`+userID+`"}]}`)}, `["resolve-user:succeeded","read-owners:succeeded","check-last-owner:blocked"]`, 200, management.ErrLastOwner},
		{"owner-role-demotes-only", management.Owner, []contracttest.Exchange{read("owners", owners), remove("owners", 204)}, `["resolve-user:succeeded","read-owners:succeeded","check-last-owner:succeeded","remove-owner-reference:succeeded"]`, 204, nil},
		{"default-removes-owner-then-member", "", []contracttest.Exchange{read("owners", owners), remove("owners", 204), remove("members", 204)}, `["resolve-user:succeeded","read-owners:succeeded","check-last-owner:succeeded","remove-owner-reference:succeeded","remove-member-reference:succeeded"]`, 204, nil},
		{"nonowner-owner-request-noop", management.Owner, []contracttest.Exchange{read("owners", `{"value":[{"id":"`+otherID+`"}]}`)}, `["resolve-user:succeeded","read-owners:succeeded","check-last-owner:succeeded","remove-owner-reference:skipped"]`, 200, nil},
		{"nonowner-member-only", management.Member, []contracttest.Exchange{read("owners", `{"value":[{"id":"`+otherID+`"}]}`), remove("members", 204)}, `["resolve-user:succeeded","read-owners:succeeded","check-last-owner:succeeded","remove-owner-reference:skipped","remove-member-reference:succeeded"]`, 204, nil},
		{"owner-removal-survives-member-failure", management.Member, []contracttest.Exchange{read("owners", owners), remove("owners", 204), remove("members", 403)}, `["resolve-user:succeeded","read-owners:succeeded","check-last-owner:succeeded","remove-owner-reference:succeeded","remove-member-reference:unconfirmed"]`, 403, auth.ErrForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			runCase(t, "Remove-TeamUser", contracttest.Case{Name: tc.name, Steps: tc.steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return changed(sdk.GraphAPI.TeamUserManagement.Remove(ctx, groupID, userID, management.ChangeOptions{Role: tc.role}))
			},
				Want: contracttest.Expected{Status: tc.status, Error: tc.want, JSON: json.RawMessage(`{"id":"` + userID + `","resolution":"identifier","stages":` + tc.stages + `}`)}})
		})
	}
}

func TestUnitRemoveReadsAllOwnersBeforeLastOwnerGuard(t *testing.T) {
	t.Parallel()
	next := auth.GraphAudience + "/v1.0/groups/" + groupID + "/owners?$skiptoken=next"
	runCase(t, "Remove-TeamUser", contracttest.Case{Name: "last-owner-after-all-pages", Steps: []contracttest.Exchange{
		read("owners", `{"value":[{"id":"`+userID+`"}],"@odata.nextLink":"`+next+`"}`),
		wire("GET", "/v1.0/groups/"+groupID+"/owners?$skiptoken=next", "", 200, ""), remove("owners", 204)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			return changed(sdk.GraphAPI.TeamUserManagement.Remove(ctx, groupID, userID, management.ChangeOptions{Role: management.Owner}))
		},
		Want: contracttest.Expected{Status: 204, JSON: json.RawMessage(`{"id":"` + userID + `","resolution":"identifier","stages":["resolve-user:succeeded","read-owners:succeeded","check-last-owner:succeeded","remove-owner-reference:succeeded"]}`)}})
}

func TestUnitCompositionInvalidInputAndCancellationAreLocal(t *testing.T) {
	t.Parallel()
	for _, method := range []string{"add", "remove", "list"} {
		for _, condition := range []string{"bad-role", "bad-group", "bad-limits", "cancelled", "closed", "nil-service"} {
			want := client.ErrRequest
			switch condition {
			case "bad-limits":
				want = client.ErrPagination
			case "cancelled":
				want = context.Canceled
			case "closed", "nil-service":
				want = client.ErrClosed
			}
			runCase(t, map[string]string{"add": "Add-TeamUser", "remove": "Remove-TeamUser", "list": "Get-TeamUser"}[method], contracttest.Case{Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				id := groupID
				role := management.Role("")
				limits := client.PagingOptions{}
				service := sdk.GraphAPI.TeamUserManagement
				switch condition {
				case "bad-role":
					role = "private"
				case "bad-group":
					id = "../private"
				case "bad-limits":
					limits.MaxPages = -1
				case "cancelled":
					cancelled, cancel := context.WithCancel(ctx)
					cancel()
					ctx = cancelled
				case "closed":
					if e := sdk.Close(); e != nil {
						return contracttest.Observation{Err: e}
					}
				case "nil-service":
					service = nil
				}
				var err error
				switch method {
				case "add":
					_, err = service.Add(ctx, id, userID, management.ChangeOptions{Role: role, Traversal: limits})
				case "remove":
					_, err = service.Remove(ctx, id, userID, management.ChangeOptions{Role: role, Traversal: limits})
				case "list":
					_, err = service.List(ctx, id, management.ListOptions{Role: role, Traversal: limits})
				}
				return contracttest.Observation{Err: err}
			}, Want: contracttest.Expected{Error: want}})
		}
	}
}
