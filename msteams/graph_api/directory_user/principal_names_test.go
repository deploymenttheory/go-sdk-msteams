package directoryuser_test

import (
	"context"
	"encoding/json"
	"fmt"
	"net/url"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
)

func namesCase(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family, c.Source = "directory_principal_names", "contracts/endpoints/8.1.0/core.json"
	if c.Scope == "" {
		c.Scope = "service"
	}
	r, e := contracttest.Run(t.Context(), c)
	if e != nil || !r.Passed() {
		t.Fatalf("principal-name contract differs: setup=%v differences=%+v", e, r.Differences())
	}
}

func namePath(ids []string) string {
	clauses := make([]string, len(ids))
	for i, id := range ids {
		clauses[i] = "id eq '" + id + "'"
	}
	return "/beta/users?" + url.Values{"$select": {"id,userPrincipalName"}, "$filter": {strings.Join(clauses, " or ")}}.Encode()
}

func TestUnitPrincipalNamesEmptyAndInvalidInputsAreLocal(t *testing.T) {
	t.Parallel()
	namesCase(t, contracttest.Case{Name: "empty", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.GraphAPI.UserResolver.LookupPrincipalNames(ctx, nil, directoryuser.PrincipalNamesOptions{})
		return contracttest.Observation{Value: v, Response: r, Err: e}
	}, Want: contracttest.Expected{JSON: json.RawMessage(`{}`)}})
	namesCase(t, contracttest.Case{Name: "invalid-uuid", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.GraphAPI.UserResolver.LookupPrincipalNames(ctx, []string{userID, "id' or true"}, directoryuser.PrincipalNamesOptions{})
		return contracttest.Observation{Value: v, Response: r, Err: e}
	}, Want: contracttest.Expected{Error: client.ErrRequest}})
	namesCase(t, contracttest.Case{Name: "invalid-limit", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.GraphAPI.UserResolver.LookupPrincipalNames(ctx, nil, directoryuser.PrincipalNamesOptions{MaxPages: -1})
		return contracttest.Observation{Value: v, Response: r, Err: e}
	}, Want: contracttest.Expected{Error: client.ErrPagination}})
}

func TestUnitPrincipalNamesRetainOnlyCompletedGroupsOnLaterFailure(t *testing.T) {
	t.Parallel()
	ids := make([]string, 11)
	for i := range ids {
		ids[i] = fmt.Sprintf("00000000-0000-4000-8000-%012x", i+1)
	}
	first := `{"value":[{"id":"` + ids[0] + `","userPrincipalName":"first@fixture.invalid"},{"id":"` + ids[1] + `","userPrincipalName":null}]}`
	namesCase(t, contracttest.Case{Name: "partial-groups", Steps: []contracttest.Exchange{wire(auth.GraphAudience, namePath(ids[:10]), 200, first), wire(auth.GraphAudience, namePath(ids[10:]), 403, `{"error":{"code":"Denied"}}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.GraphAPI.UserResolver.LookupPrincipalNames(ctx, ids, directoryuser.PrincipalNamesOptions{})
		return contracttest.Observation{Value: v, Response: r, Err: e}
	}, Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden, JSON: json.RawMessage(`{"` + ids[0] + `":"first@fixture.invalid","` + ids[1] + `":null}`)}})
	namesCase(t, contracttest.Case{Name: "ambiguous-group-atomic", Steps: []contracttest.Exchange{wire(auth.GraphAudience, namePath(ids[:1]), 200, `{"value":[{"id":"`+ids[0]+`","userPrincipalName":"unconfirmed@fixture.invalid"},{"id":"`+ids[0]+`"}]}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.GraphAPI.UserResolver.LookupPrincipalNames(ctx, ids[:1], directoryuser.PrincipalNamesOptions{})
		return contracttest.Observation{Value: v, Response: r, Err: e}
	}, Want: contracttest.Expected{Status: 200, Error: directoryuser.ErrAmbiguous, JSON: json.RawMessage(`{}`)}})
}
