package teamusermanagement

import (
	"context"
	"errors"
	"regexp"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
	relation "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group_relationship"
)

// ErrLastOwner is a local precheck rejection, not a fabricated server HTTP 412.
// It is not a concurrency guarantee; Graph still enforces its owner restrictions.
var ErrLastOwner = errors.New("cannot remove the last returned group owner")
var directoryID = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// Service composes same-scope directory resolution and group reference services.
// It does not create directory users or native conversation memberships.
type Service struct {
	scope    *client.Scope
	resolver *directoryuser.Resolver
	members  *relation.MemberService
	owners   *relation.OwnerService
}

// NewService constructs local dependencies on one tenant/cloud/identity binding.
func NewService(scope *client.Scope) *Service {
	return &Service{scope: scope, resolver: directoryuser.NewResolver(scope), members: relation.NewMemberService(scope), owners: relation.NewOwnerService(scope)}
}

func (s *Service) begin(ctx context.Context, groupID string) (context.Context, context.CancelFunc, error) {
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
	if !directoryID.MatchString(groupID) {
		return nil, nil, client.ErrRequest
	}
	return s.scope.Session().Begin(ctx)
}

func changeRole(value Role) (Role, error) {
	switch {
	case value == "" || strings.EqualFold(string(value), string(Member)):
		return Member, nil
	case strings.EqualFold(string(value), string(Owner)):
		return Owner, nil
	default:
		return "", client.ErrRequest
	}
}

func validTraversal(value client.PagingOptions) bool {
	return (value.Style == 0 || value.Style == client.LinkPaging) && value.Parameter == "" &&
		(value.ExpectedStatus == 0 || value.ExpectedStatus == 200) && value.MaxPages >= 0 && value.MaxItems >= 0
}

func (s *Service) resolve(ctx context.Context, user string, maxPages int, result *ChangeResult) error {
	if strings.TrimSpace(user) == "" {
		return client.ErrRequest
	}
	resolved, response, err := s.resolver.Resolve(ctx, user, directoryuser.ResolveOptions{MaxPages: maxPages})
	state := Succeeded
	if err != nil {
		state = Unconfirmed
	}
	result.record(ResolveUser, state, response)
	if err != nil {
		return err
	}
	result.UserID = relation.DirectoryObjectID(resolved.ID)
	result.ResolutionSource = resolved.Source
	return nil
}

func contains(items []relation.DirectoryObject, id relation.DirectoryObjectID) bool {
	for _, item := range items {
		if strings.EqualFold(string(item.ID), string(id)) {
			return true
		}
	}
	return false
}

func unique(items []relation.DirectoryObject) []relation.DirectoryObject {
	seen := map[string]bool{}
	result := make([]relation.DirectoryObject, 0, len(items))
	for _, item := range items {
		key := strings.ToLower(string(item.ID))
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}
	return result
}

func (s *Service) read(ctx context.Context, groupID string, owners bool, options relation.ListOptions, limits client.PagingOptions, result *ChangeResult) ([]relation.DirectoryObject, error) {
	var list client.ListResult[relation.DirectoryObject]
	var err error
	action := ReadMembers
	if owners {
		action = ReadOwners
		list, err = s.owners.ListAll(ctx, groupID, options, limits)
	} else {
		list, err = s.members.ListAll(ctx, groupID, options, limits)
	}
	state := Succeeded
	if err != nil {
		state = Unconfirmed
	}
	result.record(action, state, list.Response)
	return unique(list.Items), err
}

// Add ensures membership, then ownership only when Role Owner is selected.
// Complete preflight enumeration distinguishes absent from present. Any read
// failure stops writes; 200 + [] permits creation, explicitly correcting the
// pinned status-only no-op. Later failures retain completed stages; no rollback.
func (s *Service) Add(ctx context.Context, groupID, user string, options ChangeOptions) (ChangeResult, error) {
	var result ChangeResult
	ctx, stop, err := s.begin(ctx, groupID)
	if err != nil {
		return result, err
	}
	defer stop()
	role, err := changeRole(options.Role)
	if err != nil {
		return result, err
	}
	if !validTraversal(options.Traversal) || options.ResolveMaxPages < 0 {
		return result, client.ErrPagination
	}
	if err = s.resolve(ctx, user, options.ResolveMaxPages, &result); err != nil {
		return result, err
	}
	items, err := s.read(ctx, groupID, false, relation.ListOptions{Select: []string{"id"}}, options.Traversal, &result)
	if err != nil {
		return result, err
	}
	if contains(items, result.UserID) {
		result.record(AddMemberReference, Skipped, nil)
	} else {
		response, err := s.members.AddReference(ctx, groupID, relation.Reference{ID: result.UserID})
		state := Succeeded
		if err != nil {
			state = Unconfirmed
		}
		result.record(AddMemberReference, state, response)
		if err != nil {
			return result, err
		}
	}
	if role != Owner {
		return result, nil
	}
	items, err = s.read(ctx, groupID, true, relation.ListOptions{Select: []string{"id"}}, options.Traversal, &result)
	if err != nil {
		return result, err
	}
	if contains(items, result.UserID) {
		result.record(AddOwnerReference, Skipped, nil)
		return result, nil
	}
	response, err := s.owners.AddReference(ctx, groupID, relation.Reference{ID: result.UserID})
	state := Succeeded
	if err != nil {
		state = Unconfirmed
	}
	result.record(AddOwnerReference, state, response)
	return result, err
}

// Remove reads all owners, blocks the pinned last-returned-owner case, then
// removes ownership first. Role Owner retains membership; omitted/Member removes
// the member reference afterwards. A non-owner Owner request is a no-op.
func (s *Service) Remove(ctx context.Context, groupID, user string, options ChangeOptions) (ChangeResult, error) {
	var result ChangeResult
	ctx, stop, err := s.begin(ctx, groupID)
	if err != nil {
		return result, err
	}
	defer stop()
	role, err := changeRole(options.Role)
	if err != nil {
		return result, err
	}
	if !validTraversal(options.Traversal) || options.ResolveMaxPages < 0 {
		return result, client.ErrPagination
	}
	if err = s.resolve(ctx, user, options.ResolveMaxPages, &result); err != nil {
		return result, err
	}
	owners, err := s.read(ctx, groupID, true, relation.ListOptions{Select: []string{"id"}}, options.Traversal, &result)
	if err != nil {
		return result, err
	}
	isOwner := contains(owners, result.UserID)
	if isOwner && len(owners) == 1 {
		result.record(CheckLastOwner, Blocked, nil)
		return result, ErrLastOwner
	}
	result.record(CheckLastOwner, Succeeded, nil)
	if isOwner {
		response, err := s.owners.RemoveReference(ctx, groupID, result.UserID)
		state := Succeeded
		if err != nil {
			state = Unconfirmed
		}
		result.record(RemoveOwnerReference, state, response)
		if err != nil {
			return result, err
		}
	} else {
		result.record(RemoveOwnerReference, Skipped, nil)
	}
	if role == Owner {
		return result, nil
	}
	response, err := s.members.RemoveReference(ctx, groupID, result.UserID)
	state := Succeeded
	if err != nil {
		state = Unconfirmed
	}
	result.record(RemoveMemberReference, state, response)
	return result, err
}
