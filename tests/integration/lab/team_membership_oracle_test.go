package lab_test

import (
	"context"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_user_management"
)

func sameTeamOracleUsers(want, actual []labkit.TeamOracleUser) bool {
	index := func(values []labkit.TeamOracleUser) (map[string]labkit.TeamOracleUser, bool) {
		result := make(map[string]labkit.TeamOracleUser, len(values))
		for _, value := range values {
			if !labkit.DirectoryID.MatchString(value.ID) {
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
	if !labkit.OracleEnabled(t) {
		return
	}
	listed, err := sdk.GraphAPI.TeamUserManagement.List(ctx, id, management.ListOptions{})
	if err != nil || !listed.Complete {
		t.Fatal("complete Go team-user snapshot required before PowerShell comparison")
	}
	want := make([]labkit.TeamOracleUser, 0, len(listed.Items))
	for _, value := range listed.Items {
		want = append(want, labkit.TeamOracleUser{ID: string(value.UserID), User: value.User, Name: value.Name, Role: string(value.Role)})
	}
	observed := labkit.RunTeamOracle(t, ctx, labkit.TeamOracleRequest{Kind: "users", ID: id, Fixture: fixture})
	if !sameTeamOracleUsers(want, observed.Users) {
		t.Fatal("PowerShell and Go team-user identity, role or selected fields differ")
	}
	t.Log("PowerShell 8.1.0 / Go team-user projection: passed")
}

func TestUnitTeamUserOracleComparesUniqueCollectionsWithoutAssumingServerOrder(t *testing.T) {
	t.Parallel()
	first := labkit.TeamOracleUser{ID: "11111111-1111-4111-8111-111111111111", Role: "owner"}
	second := labkit.TeamOracleUser{ID: "aaaaaaaa-aaaa-4aaa-8aaa-aaaaaaaaaaaa", Role: "member"}
	for _, tc := range []struct {
		name   string
		actual []labkit.TeamOracleUser
		equal  bool
	}{
		{name: "same", actual: []labkit.TeamOracleUser{first, second}, equal: true},
		{name: "different-order", actual: []labkit.TeamOracleUser{second, first}, equal: true},
		{name: "case-insensitive-id-role", actual: []labkit.TeamOracleUser{first, {ID: strings.ToUpper(second.ID), Role: "Member"}}, equal: true},
		{name: "missing-member", actual: []labkit.TeamOracleUser{first}},
		{name: "duplicate-member", actual: []labkit.TeamOracleUser{first, second, second}},
		{name: "changed-role", actual: []labkit.TeamOracleUser{first, {ID: second.ID, Role: "owner"}}},
		{name: "invalid-id", actual: []labkit.TeamOracleUser{first, {ID: "../unowned", Role: "member"}}},
		{name: "null-not-empty", actual: []labkit.TeamOracleUser{first, {ID: second.ID, Role: "member", User: new(string)}}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if sameTeamOracleUsers([]labkit.TeamOracleUser{first, second}, tc.actual) != tc.equal {
				t.Fatal("team-user oracle comparison differs")
			}
		})
	}
	if sameTeamOracleUsers([]labkit.TeamOracleUser{first, first}, []labkit.TeamOracleUser{first}) {
		t.Fatal("duplicate expected identities were collapsed")
	}
}
