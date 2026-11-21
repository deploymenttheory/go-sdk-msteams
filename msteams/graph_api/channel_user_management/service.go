package channelusermanagement

import (
	"context"
	"errors"
	"regexp"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel"
	member "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel_member"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
	"github.com/go-resty/resty/v2"
)

// ErrStandardChannel is a local composition rejection, not a received HTTP 403.
var ErrStandardChannel = errors.New("standard-channel membership is managed through the team")
var directoryID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Service composes native services bound to one tenant/cloud/identity scope.
type Service struct {
	scope    *client.Scope
	channels *channel.Service
	members  *member.Service
	users    *directoryuser.Resolver
}

// NewService constructs dependencies locally, without credentials or HTTP.
func NewService(scope *client.Scope) *Service {
	return &Service{scope: scope, channels: channel.NewService(scope), members: member.NewService(scope), users: directoryuser.NewResolver(scope)}
}

func (s *Service) begin(ctx context.Context, teamID, name string, limits client.PagingOptions) (context.Context, context.CancelFunc, error) {
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
	if !directoryID.MatchString(teamID) || strings.TrimSpace(name) == "" {
		return nil, nil, client.ErrRequest
	}
	if !validTraversal(limits) {
		return nil, nil, client.ErrPagination
	}
	return s.scope.Session().Begin(ctx)
}

func validTraversal(value client.PagingOptions) bool {
	return (value.Style == 0 || value.Style == client.LinkPaging) && value.Parameter == "" &&
		(value.ExpectedStatus == 0 || value.ExpectedStatus == 200) && value.MaxPages >= 0 && value.MaxItems >= 0
}

func validChange(user string, options ChangeOptions) bool {
	role := strings.ToLower(string(options.Role))
	return strings.TrimSpace(user) != "" && user != "." && user != ".." && !strings.ContainsRune(user, '\\') && !strings.ContainsFunc(user, unicode.IsControl) &&
		(role == "" || role == string(Owner)) && options.ResolveMaxPages >= 0 &&
		(options.TenantID == nil || directoryID.MatchString(*options.TenantID))
}

func (s *Service) resolveChannel(ctx context.Context, teamID, name string, limits client.PagingOptions, result *ChangeResult) (*channel.Channel, error) {
	value, response, err := s.channels.ResolveByDisplayName(ctx, teamID, name, limits)
	record(&result.Stages, &result.Response, ResolveChannel, response, err)
	if err != nil {
		return nil, err
	}
	result.ChannelID = value.ID
	if value.MembershipType == nil {
		return nil, client.ErrDecode
	}
	if *value.MembershipType == channel.Standard {
		return nil, ErrStandardChannel
	}
	if *value.MembershipType != channel.Private && *value.MembershipType != channel.Shared {
		return nil, client.ErrRequest
	}
	return value, nil
}

func (s *Service) resolveUser(ctx context.Context, selector string, maxPages int, result *ChangeResult) error {
	value, response, err := s.users.Resolve(ctx, selector, directoryuser.ResolveOptions{MaxPages: maxPages})
	record(&result.Stages, &result.Response, ResolveUser, response, err)
	if err == nil {
		result.UserID = member.DirectoryUserID(value.ID)
	}
	return err
}

func (s *Service) resolveMembership(ctx context.Context, teamID string, c *channel.Channel, user string, options ChangeOptions, result *ChangeResult) error {
	if *c.MembershipType == channel.Private {
		if err := s.resolveUser(ctx, user, options.ResolveMaxPages, result); err != nil {
			return err
		}
		value, response, err := s.members.ResolveByUser(ctx, teamID, c.ID, result.UserID, options.TenantID, options.Traversal)
		record(&result.Stages, &result.Response, ResolveMembership, response, err)
		if err == nil {
			result.MembershipID, result.Member = value.ID, value
		}
		return err
	}
	// Shared members may belong to another tenant. Match returned ID/email
	// without substituting a host-directory lookup or synthesizing membership IDs.
	listed, err := s.members.ListAll(ctx, teamID, c.ID, member.ListOptions{}, options.Traversal)
	if err == nil {
		for _, value := range listed.Items {
			if options.TenantID != nil && (value.TenantID == nil || !strings.EqualFold(*value.TenantID, *options.TenantID)) {
				continue
			}
			matches := value.UserID != nil && strings.EqualFold(string(*value.UserID), user) || value.Email != nil && strings.EqualFold(*value.Email, user)
			if !matches {
				continue
			}
			if result.Member != nil {
				err = member.ErrAmbiguous
				break
			}
			copyValue := value
			result.Member, result.MembershipID = &copyValue, value.ID
			if value.UserID != nil {
				result.UserID = *value.UserID
			}
		}
		if err == nil && result.Member == nil {
			err = member.ErrNotFound
		}
	}
	if err != nil {
		result.Member, result.MembershipID = nil, ""
		result.UserID = ""
	}
	record(&result.Stages, &result.Response, ResolveMembership, listed.Response, err)
	return err
}

// Add adds ordinary membership when Role is omitted, or promotes an existing
// membership when Owner is selected. It never silently adds a missing owner.
// Native payloads use documented empty/owner roles and selected-cloud bindings.
func (s *Service) Add(ctx context.Context, teamID, name, user string, options ChangeOptions) (ChangeResult, error) {
	return s.change(ctx, teamID, name, user, options, true)
}

// Remove deletes membership when Role is omitted, or demotes an existing owner
// while retaining membership when Owner is selected. Graph enforces last-owner
// constraints; no local check claims a concurrent ownership guarantee.
func (s *Service) Remove(ctx context.Context, teamID, name, user string, options ChangeOptions) (ChangeResult, error) {
	return s.change(ctx, teamID, name, user, options, false)
}

func (s *Service) change(ctx context.Context, teamID, name, user string, options ChangeOptions, add bool) (ChangeResult, error) {
	var result ChangeResult
	ctx, stop, err := s.begin(ctx, teamID, name, options.Traversal)
	if err != nil {
		return result, err
	}
	defer stop()
	if !validChange(user, options) {
		return result, client.ErrRequest
	}
	if add && options.Role == "" && options.TenantID != nil && !directoryID.MatchString(user) {
		// External creation addresses an explicit user object in the selected
		// tenant; a host-directory UPN lookup cannot establish that identity.
		return result, client.ErrRequest
	}
	c, err := s.resolveChannel(ctx, teamID, name, options.Traversal, &result)
	if err != nil {
		return result, err
	}
	owner := strings.EqualFold(string(options.Role), string(Owner))
	if add && !owner {
		if err := s.resolveUser(ctx, user, options.ResolveMaxPages, &result); err != nil {
			return result, err
		}
		value, response, err := s.members.Create(ctx, teamID, c.ID, member.CreateRequest{UserID: result.UserID, TenantID: options.TenantID})
		record(&result.Stages, &result.Response, CreateMember, response, err)
		if value != nil {
			result.Member, result.MembershipID = value, value.ID
		}
		return result, err
	}
	if err := s.resolveMembership(ctx, teamID, c, user, options, &result); err != nil {
		return result, err
	}
	if owner {
		var value *member.Member
		var response *resty.Response
		action := DemoteOwner
		if add {
			action = PromoteOwner
			value, response, err = s.members.PromoteOwner(ctx, teamID, c.ID, result.MembershipID)
		} else {
			value, response, err = s.members.DemoteOwner(ctx, teamID, c.ID, result.MembershipID)
		}
		record(&result.Stages, &result.Response, action, response, err)
		if value != nil {
			result.Member = value
		}
		return result, err
	}
	response, err := s.members.DeleteByID(ctx, teamID, c.ID, result.MembershipID)
	record(&result.Stages, &result.Response, DeleteMember, response, err)
	return result, err
}
