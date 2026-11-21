package directoryuser

import (
	"context"
	"net/http"
	"net/url"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// PrincipalNamesOptions bounds each opaque user-query traversal. Zero MaxPages
// selects the shared pager default. The caller's context spans every group.
type PrincipalNamesOptions struct{ MaxPages int }

// LookupPrincipalNames reads selected directory UUIDs in distinct groups of at
// most ten on the pinned beta user-query family. Keys are lowercase UUIDs.
// A present nil value means the server returned that user with an absent/null
// UPN; a missing key means no directory match. Empty input performs no HTTP.
// On failure, only names from earlier fully validated groups are retained.
func (r *Resolver) LookupPrincipalNames(ctx context.Context, userIDs []string, options PrincipalNamesOptions) (map[string]*string, *resty.Response, error) {
	if err := r.ready(ctx); err != nil {
		return nil, nil, err
	}
	if options.MaxPages < 0 {
		return nil, nil, client.ErrPagination
	}
	ids := make([]string, 0, len(userIDs))
	seen := map[string]bool{}
	for _, raw := range userIDs {
		if !directoryID.MatchString(raw) {
			return nil, nil, client.ErrRequest
		}
		id := strings.ToLower(raw)
		if !seen[id] {
			ids = append(ids, id)
			seen[id] = true
		}
	}
	ctx, stop, err := r.scope.Session().Begin(ctx)
	if err != nil {
		return nil, nil, err
	}
	defer stop()
	names := map[string]*string{}
	var last *resty.Response
	for start := 0; start < len(ids); start += 10 {
		end := min(start+10, len(ids))
		selected := map[string]bool{}
		clauses := make([]string, 0, end-start)
		for _, id := range ids[start:end] {
			selected[id] = true
			clauses = append(clauses, "id eq '"+id+"'")
		}
		req := client.Request{Method: http.MethodGet, Path: []string{"beta", "users"}, Replay: client.ReplaySafe, Header: http.Header{"Accept": {"application/json"}}, Query: url.Values{"$select": {"id,userPrincipalName"}, "$filter": {strings.Join(clauses, " or ")}}}
		pager, err := client.NewPager(r.scope, req, decodePrincipalNameUsers, client.PagingOptions{Style: client.LinkPaging, MaxPages: options.MaxPages, ExpectedStatus: http.StatusOK})
		if err != nil {
			return names, last, err
		}
		listed, err := pager.Collect(ctx)
		if listed.Response != nil {
			last = listed.Response
		}
		if err != nil {
			return names, last, err
		}
		group := map[string]*string{}
		for _, user := range listed.Items {
			key := strings.ToLower(user.ID)
			if !selected[key] {
				return names, last, client.ErrDecode
			}
			if _, exists := group[key]; exists {
				return names, last, ErrAmbiguous
			}
			group[key] = user.UserPrincipalName
		}
		for key, value := range group {
			names[key] = value
		}
	}
	return names, last, nil
}

func decodePrincipalNameUsers(data []byte) (client.PageContent[User], error) {
	return client.DecodeCollection[User](data, "value", "@odata.nextLink")
}
