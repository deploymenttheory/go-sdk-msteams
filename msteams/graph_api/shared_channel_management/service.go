// Package sharedchannelmanagement composes the pinned shared-channel views
// through native relationship, associated-team and directory resource services.
package sharedchannelmanagement

import (
	"context"
	"errors"
	"sort"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	associatedteam "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/associated_team"
	member "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_member"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
	sharedchannel "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/shared_channel"
	"github.com/go-resty/resty/v2"
)

type Service struct {
	scope      *client.Scope
	shares     *sharedchannel.Service
	associated *associatedteam.Service
	users      *directoryuser.Resolver
}

// NewService constructs same-scope dependencies without credential acquisition.
func NewService(scope *client.Scope) *Service {
	return &Service{scope: scope, shares: sharedchannel.NewService(scope), associated: associatedteam.NewService(scope), users: directoryuser.NewResolver(scope)}
}

func (s *Service) begin(ctx context.Context) (context.Context, context.CancelFunc, error) {
	if ctx == nil {
		return nil, nil, client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return nil, nil, err
	}
	if s == nil || s.scope == nil || s.scope.Session() == nil {
		return nil, nil, client.ErrClosed
	}
	if s.scope.Backend() != client.Graph {
		return nil, nil, client.ErrDestination
	}
	return s.scope.Session().Begin(ctx)
}

// SharedTeamsOptions selects a single sharing identity or the complete list.
// Empty SharedTeamID retains the pinned command's collection behavior.
type SharedTeamsOptions struct {
	SharedTeamID string
	Traversal    client.PagingOptions
}

// ListSharedWithTeams preserves collection output for both command branches.
// A supplied identity is one native Get, without an unnecessary list/pre-read.
func (s *Service) ListSharedWithTeams(ctx context.Context, hostTeamID, channelID string, options SharedTeamsOptions) (client.ListResult[sharedchannel.SharedWithTeam], error) {
	ctx, stop, err := s.begin(ctx)
	if err != nil {
		return client.ListResult[sharedchannel.SharedWithTeam]{}, err
	}
	defer stop()
	if options.SharedTeamID == "" {
		return s.shares.ListAll(ctx, hostTeamID, channelID, options.Traversal)
	}
	value, response, err := s.shares.GetByID(ctx, hostTeamID, channelID, options.SharedTeamID)
	result := client.ListResult[sharedchannel.SharedWithTeam]{Response: response}
	if value != nil {
		result.Items = []sharedchannel.SharedWithTeam{*value}
		result.Pages = 1
	}
	return result, err
}

// UnshareWithTeam names the native relationship delete explicitly. It does not
// delete a team/channel, initiate sharing or repeat an ambiguous write.
func (s *Service) UnshareWithTeam(ctx context.Context, hostTeamID, channelID, sharedTeamID string) (*resty.Response, error) {
	ctx, stop, err := s.begin(ctx)
	if err != nil {
		return nil, err
	}
	defer stop()
	return s.shares.DeleteByID(ctx, hostTeamID, channelID, sharedTeamID)
}

type AssociatedOptions struct {
	Traversal       client.PagingOptions
	ResolveMaxPages int
}

type AssociatedResult struct {
	Items            []associatedteam.Team
	UserID           string
	ResolutionSource directoryuser.Source
	Response         *resty.Response
	Complete         bool
}

func (AssociatedResult) String() string     { return "associated-team result (redacted)" }
func (r AssociatedResult) GoString() string { return r.String() }

// ListAssociatedTeams implements omitted/empty User as the explicit delegated
// current-user view, or resolves an explicit UUID/UPN for the user route. The
// shared resolver uses its complete unique canonical mail fallback after 404.
func (s *Service) ListAssociatedTeams(ctx context.Context, user string, options AssociatedOptions) (AssociatedResult, error) {
	var result AssociatedResult
	ctx, stop, err := s.begin(ctx)
	if err != nil {
		return result, err
	}
	defer stop()
	if options.ResolveMaxPages < 0 {
		return result, client.ErrPagination
	}
	if !validTraversal(options.Traversal) {
		return result, client.ErrPagination
	}
	var listed client.ListResult[associatedteam.Team]
	if user == "" {
		listed, err = s.associated.ListMine(ctx, options.Traversal)
	} else {
		value, response, resolveErr := s.users.Resolve(ctx, user, directoryuser.ResolveOptions{MaxPages: options.ResolveMaxPages})
		result.Response = response
		if resolveErr != nil {
			return result, resolveErr
		}
		result.UserID, result.ResolutionSource = value.ID, value.Source
		listed, err = s.associated.ListForUser(ctx, value.ID, options.Traversal)
	}
	result.Items = listed.Items
	if listed.Response != nil {
		result.Response = listed.Response
	}
	result.Complete = err == nil
	return result, err
}

// Role is the projected allowed-user category, distinct from native qualifiers.
type Role string

const (
	Owner  Role = "owner"
	Member Role = "member"
	Guest  Role = "guest"
)

type AllowedUsersOptions struct {
	Role               Role
	Traversal          client.PagingOptions
	UserLookupMaxPages int
}

// UserSummary retains the native access record and its tenant/membership IDs.
// Allowed-access identities are not assumed to address the direct channel roster.
type UserSummary struct {
	MembershipID member.MembershipID
	UserID       *member.DirectoryUserID
	User         *string
	Name         *string
	Role         Role
	TenantID     *string
	Membership   member.Member
}

type AllowedUsersResult struct {
	Items    []UserSummary
	Members  []member.Member
	Response *resty.Response
	Complete bool
}

func (AllowedUsersResult) String() string     { return "shared-channel allowed-user result (redacted)" }
func (r AllowedUsersResult) GoString() string { return r.String() }

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

// ListAllowedUsers completely reads the selected sharing relationship's allowed
// members, projects/filter roles and enriches selected UUIDs through bounded
// directory reads. The pinned command allows only Owner/Member role filters;
// unfiltered future read data remains preserved in native membership records.
func (s *Service) ListAllowedUsers(ctx context.Context, hostTeamID, channelID, sharedTeamID string, options AllowedUsersOptions) (AllowedUsersResult, error) {
	var result AllowedUsersResult
	ctx, stop, err := s.begin(ctx)
	if err != nil {
		return result, err
	}
	defer stop()
	role := Role(strings.ToLower(string(options.Role)))
	if role != "" && role != Owner && role != Member || options.UserLookupMaxPages < 0 {
		return result, client.ErrRequest
	}
	listed, err := s.shares.ListAllowedMembers(ctx, hostTeamID, channelID, sharedTeamID, options.Traversal)
	result.Members, result.Response = listed.Items, listed.Response
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
		result.Items = append(result.Items, UserSummary{MembershipID: value.ID, UserID: value.UserID, User: user, Name: value.DisplayName, Role: category, TenantID: value.TenantID, Membership: value})
	}
	if role == "" {
		sort.SliceStable(result.Items, func(i, j int) bool { return roleOrder(result.Items[i].Role) < roleOrder(result.Items[j].Role) })
	}
	ids := []string{}
	for _, value := range result.Items {
		if value.UserID != nil {
			ids = append(ids, string(*value.UserID))
		}
	}
	names, response, err := s.users.LookupPrincipalNames(ctx, ids, directoryuser.PrincipalNamesOptions{MaxPages: options.UserLookupMaxPages})
	if errors.Is(err, client.ErrRequest) {
		// IDs originated in a server membership envelope, rather than an
		// invalid caller-supplied directory selector.
		err = client.ErrDecode
	}
	if response != nil {
		result.Response = response
	}
	for i, value := range result.Items {
		if value.UserID != nil {
			if name, exists := names[strings.ToLower(string(*value.UserID))]; exists {
				result.Items[i].User = name
			}
		}
	}
	result.Complete = err == nil
	return result, err
}

func validTraversal(value client.PagingOptions) bool {
	return (value.Style == 0 || value.Style == client.LinkPaging) && value.Parameter == "" &&
		(value.ExpectedStatus == 0 || value.ExpectedStatus == 200) && value.MaxPages >= 0 && value.MaxItems >= 0
}
