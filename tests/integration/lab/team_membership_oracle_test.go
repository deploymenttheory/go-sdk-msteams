package lab_test

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_user_management"
)

func sameTeamOracleUsers(want, actual []teamOracleUser) bool {
	index := func(values []teamOracleUser) (map[string]teamOracleUser, bool) {
		result := make(map[string]teamOracleUser, len(values))
		for _, value := range values {
			if !labDirectoryID.MatchString(value.ID) {
				return nil, false
			}
			value.ID, value.Role = strings.ToLower(value.ID), strings.ToLower(value.Role)
			if _, duplicate := result[value.ID]; duplicate {
				return nil, false
			}
			result[value.ID] = value
		}
		return result, true
	}
	left, ok := index(want)
	if !ok {
		return false
	}
	right, ok := index(actual)
	return ok && reflect.DeepEqual(left, right)
}

func comparePowerShellTeamUsers(t *testing.T, ctx context.Context, sdk *msteams.Client, id, fixture string) {
	t.Helper()
	if !teamOracleEnabled(t) {
		return
	}
	listed, err := sdk.GraphAPI.TeamUserManagement.List(ctx, id, management.ListOptions{})
	if err != nil || !listed.Complete {
		t.Fatal("complete Go team-user snapshot required before PowerShell comparison")
	}
	want := make([]teamOracleUser, 0, len(listed.Items))
	for _, value := range listed.Items {
		want = append(want, teamOracleUser{ID: string(value.UserID), User: value.User, Name: value.Name, Role: string(value.Role)})
	}
	observed := runTeamOracle(t, ctx, teamOracleRequest{Kind: "users", ID: id, Fixture: fixture})
	if !sameTeamOracleUsers(want, observed.Users) {
		t.Fatal("PowerShell and Go team-user identity, role or selected fields differ")
	}
	t.Log("PowerShell 8.1.0 / Go team-user projection: passed")
}

func TestTeamUserOracleComparesUniqueCollectionsWithoutAssumingServerOrder(t *testing.T) {
	t.Parallel()
	first := teamOracleUser{ID: "11111111-1111-4111-8111-111111111111", Role: "owner"}
	second := teamOracleUser{ID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", Role: "member"}
	for _, tc := range []struct {
		name   string
		actual []teamOracleUser
		equal  bool
	}{
		{name: "same", actual: []teamOracleUser{first, second}, equal: true},
		{name: "different-order", actual: []teamOracleUser{second, first}, equal: true},
		{name: "case-insensitive-id-role", actual: []teamOracleUser{first, {ID: strings.ToUpper(second.ID), Role: "Member"}}, equal: true},
		{name: "missing-member", actual: []teamOracleUser{first}},
		{name: "duplicate-member", actual: []teamOracleUser{first, second, second}},
		{name: "changed-role", actual: []teamOracleUser{first, {ID: second.ID, Role: "owner"}}},
		{name: "invalid-id", actual: []teamOracleUser{first, {ID: "../unowned", Role: "member"}}},
		{name: "null-not-empty", actual: []teamOracleUser{first, {ID: second.ID, Role: "member", User: new(string)}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if sameTeamOracleUsers([]teamOracleUser{first, second}, tc.actual) != tc.equal {
				t.Fatal("team-user oracle comparison differs")
			}
		})
	}
	if sameTeamOracleUsers([]teamOracleUser{first, first}, []teamOracleUser{first}) {
		t.Fatal("duplicate expected identities were collapsed")
	}
}
