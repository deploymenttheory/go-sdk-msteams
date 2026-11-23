package graphacc

import (
	"context"
	"reflect"
	"strings"
	"testing"

	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_management"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func comparePowerShellTeam(t *testing.T, ctx context.Context, value *management.ManagedTeam, fixture, owner string) {
	t.Helper()
	if !labkit.OracleEnabled(t) {
		return
	}
	if value == nil || value.TeamSnapshot == nil {
		t.Fatal("complete Go snapshot required before PowerShell comparison")
	}
	want := labkit.TeamOracleProjection{ID: value.ID, DisplayName: value.DisplayName, Description: value.Description, Alias: value.MailNickname,
		Classification: value.Classification, Archived: value.IsArchived}
	if value.Visibility != nil {
		text := strings.ToLower(string(*value.Visibility))
		want.Visibility = &text
	}
	if value.TeamSnapshot.FunSettings != nil {
		want.AllowGiphy = value.TeamSnapshot.FunSettings.AllowGiphy
	}
	observed := labkit.RunTeamOracle(t, ctx, labkit.TeamOracleRequest{Kind: "team", ID: value.ID, Fixture: fixture, Owner: owner})
	for _, actual := range []*labkit.TeamOracleProjection{observed.Team, observed.Tenant, observed.User} {
		actual.ID = strings.ToLower(actual.ID)
		want.ID = strings.ToLower(want.ID)
		if actual.Visibility != nil {
			text := strings.ToLower(*actual.Visibility)
			actual.Visibility = &text
		}
		if !reflect.DeepEqual(*actual, want) {
			t.Fatal("PowerShell and Go selected team identity/list fields differ")
		}
	}
	t.Log("PowerShell 8.1.0 / Go selected team identity and list observations: passed")
}
