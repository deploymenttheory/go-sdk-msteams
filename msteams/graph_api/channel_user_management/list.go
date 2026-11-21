package channelusermanagement

import (
	"context"
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

func (s *Service) enrich(ctx context.Context, items []UserSummary, maxPages int) (*resty.Response, error) {
	ids := make([]string, 0, len(items))
	for _, value := range items {
		if value.UserID == nil {
			continue
		}
		id := string(*value.UserID)
		if !directoryID.MatchString(id) {
			return nil, client.ErrDecode
		}
		ids = append(ids, id)
	}
	names, response, err := s.users.LookupPrincipalNames(ctx, ids, directoryuser.PrincipalNamesOptions{MaxPages: maxPages})
	for index, value := range items {
		if value.UserID != nil {
			if name, exists := names[strings.ToLower(string(*value.UserID))]; exists {
				items[index].User = name
			}
		}
	}
	return response, err
}
