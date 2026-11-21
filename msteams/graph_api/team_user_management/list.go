package teamusermanagement

import (
	"context"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	relation "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group_relationship"
)

func listRole(value Role) (Role, error) {
	if value == "" {
		return "", nil
	}
	for _, role := range []Role{Owner, Member, Guest} {
		if strings.EqualFold(string(value), string(role)) {
			return role, nil
		}
	}
	return "", client.ErrRequest
}

func summary(value relation.DirectoryObject, role Role) UserSummary {
	return UserSummary{UserID: value.ID, User: value.UserPrincipalName, Name: value.DisplayName, Role: role, DirectoryObject: value}
}

func project(owners, members []relation.DirectoryObject, role Role) []UserSummary {
	result := []UserSummary{}
	ownerIDs := map[string]bool{}
	for _, owner := range unique(owners) {
		ownerIDs[strings.ToLower(string(owner.ID))] = true
		if role == "" || role == Owner {
			result = append(result, summary(owner, Owner))
		}
	}
	ordinary, guests := []UserSummary{}, []UserSummary{}
	for _, member := range unique(members) {
		if ownerIDs[strings.ToLower(string(member.ID))] {
			continue
		}
		isGuest := member.UserType != nil && strings.EqualFold(*member.UserType, "guest")
		if isGuest {
			guests = append(guests, summary(member, Guest))
		} else {
			ordinary = append(ordinary, summary(member, Member))
		}
	}
	if role == "" || role == Member {
		result = append(result, ordinary...)
	}
	if role == "" || role == Guest {
		result = append(result, guests...)
	}
	return result
}

// List projects the pinned owner/member/guest categories and output order.
// Owner precedence deduplicates shared IDs. Guest-only skips owners; Owner-only
// skips members. The compatibility projection classifies nil userType as member,
// while preserving its unknown native directory data in each summary.
func (s *Service) List(ctx context.Context, groupID string, options ListOptions) (ListResult, error) {
	var result ListResult
	ctx, stop, err := s.begin(ctx, groupID)
	if err != nil {
		return result, err
	}
	defer stop()
	role, err := listRole(options.Role)
	if err != nil {
		return result, err
	}
	if !validTraversal(options.Traversal) {
		return result, client.ErrPagination
	}
	var observations ChangeResult
	var owners, members []relation.DirectoryObject
	selection := relation.ListOptions{Select: []string{"id", "displayName", "userPrincipalName", "userType"}}
	finish := func(err error) (ListResult, error) {
		result.Items = project(owners, members, role)
		result.Stages = observations.Stages
		result.Response = observations.Response
		result.Complete = err == nil
		return result, err
	}
	if role != Guest {
		owners, err = s.read(ctx, groupID, true, selection, options.Traversal, &observations)
		if err != nil {
			return finish(err)
		}
	}
	if role != Owner {
		members, err = s.read(ctx, groupID, false, selection, options.Traversal, &observations)
		if err != nil {
			return finish(err)
		}
	}
	return finish(nil)
}
