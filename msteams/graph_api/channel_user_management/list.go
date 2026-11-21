package channelusermanagement

import (
	"context"
	"net/http"
	"net/url"
	"sort"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	member "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_member"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
	"github.com/go-resty/resty/v2"
)

func projectedRole(roles []member.Role) Role {
	guest := false
	for _, role := range roles {
		if strings.EqualFold(string(role), string(Owner)) {
			return Owner
		}
		guest = guest || strings.EqualFold(string(role), string(Guest))
	}
	if guest {
		return Guest
	}
	return Member
}

func roleOrder(role Role) int {
	switch role {
	case Owner:
		return 0
	case Guest:
		return 2
	default:
		return 1
	}
}

// List resolves a unique hosted channel, reads all bounded native memberships,
// projects/filter roles and enriches selected users through beta directory
// queries in groups of at most ten. Standard-channel reads are allowed. Failed
// enrichment is a failed/incomplete workflow; initial email values are not
// silently reported as a successful final result.
func (s *Service) List(ctx context.Context, teamID, name string, options ListOptions) (ListResult, error) {
	var result ListResult
	ctx, stop, err := s.begin(ctx, teamID, name, options.Traversal)
	if err != nil {
		return result, err
	}
	defer stop()
	role := Role(strings.ToLower(string(options.Role)))
	if role != "" && role != Owner && role != Member && role != Guest || options.UserLookupMaxPages < 0 {
		return result, client.ErrRequest
	}
	c, response, err := s.channels.ResolveByDisplayName(ctx, teamID, name, options.Traversal)
	record(&result.Stages, &result.Response, ResolveChannel, response, err)
	if err != nil {
		return result, err
	}
	result.ChannelID = c.ID
	listed, err := s.members.ListAll(ctx, teamID, c.ID, member.ListOptions{}, options.Traversal)
	result.Members = listed.Items
	record(&result.Stages, &result.Response, ReadMembers, listed.Response, err)
	if err != nil {
		return result, err
	}
	result.Items = []UserSummary{}
	for _, value := range listed.Items {
		category := projectedRole(value.Roles)
		if role != "" && role != category {
			continue
		}
		user := value.UserPrincipalName
		if user == nil {
			user = value.Email
		}
		result.Items = append(result.Items, UserSummary{MembershipID: value.ID, UserID: value.UserID, User: user,
			Name: value.DisplayName, Role: category, TenantID: value.TenantID, Membership: value})
	}
	if role == "" {
		sort.SliceStable(result.Items, func(i, j int) bool { return roleOrder(result.Items[i].Role) < roleOrder(result.Items[j].Role) })
	}
	response, err = s.enrich(ctx, result.Items, options.UserLookupMaxPages)
	record(&result.Stages, &result.Response, EnrichUsers, response, err)
	if err == nil {
		result.Complete = true
	}
	return result, err
}

func decodeUsers(data []byte) (client.PageContent[directoryuser.User], error) {
	return client.DecodeCollection[directoryuser.User](data, "value", "@odata.nextLink")
}

func (s *Service) enrich(ctx context.Context, items []UserSummary, maxPages int) (*resty.Response, error) {
	ids := []string{}
	seen := map[string]bool{}
	for _, value := range items {
		if value.UserID == nil {
			continue
		}
		id := strings.ToLower(string(*value.UserID))
		if !directoryID.MatchString(id) {
			return nil, client.ErrDecode
		}
		if !seen[id] {
			seen[id] = true
			ids = append(ids, id)
		}
	}
	var last *resty.Response
	for start := 0; start < len(ids); start += 10 {
		end := min(start+10, len(ids))
		selected := map[string]bool{}
		clauses := make([]string, 0, end-start)
		for _, id := range ids[start:end] {
			selected[id] = true
			clauses = append(clauses, "id eq '"+id+"'")
		}
		req := client.Request{Method: http.MethodGet, Path: []string{"beta", "users"}, Replay: client.ReplaySafe,
			Header: http.Header{"Accept": {"application/json"}},
			Query:  url.Values{"$select": {"id,userPrincipalName"}, "$filter": {strings.Join(clauses, " or ")}}}
		pager, err := client.NewPager(s.scope, req, decodeUsers, client.PagingOptions{Style: client.LinkPaging, MaxPages: maxPages, ExpectedStatus: http.StatusOK})
		if err != nil {
			return last, err
		}
		listed, err := pager.Collect(ctx)
		if listed.Response != nil {
			last = listed.Response
		}
		if err != nil {
			return last, err
		}
		names := map[string]*string{}
		for _, user := range listed.Items {
			key := strings.ToLower(user.ID)
			if !selected[key] {
				return last, client.ErrDecode
			}
			if _, exists := names[key]; exists {
				return last, directoryuser.ErrAmbiguous
			}
			names[key] = user.UserPrincipalName
		}
		for index, value := range items {
			if value.UserID != nil {
				if name, exists := names[strings.ToLower(string(*value.UserID))]; exists {
					items[index].User = name
				}
			}
		}
	}
	return last, nil
}
