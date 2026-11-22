package lab_test

import (
	"context"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/shared_channel_management"
)

type sharedChannelFixture struct {
	TenantID         string `json:"tenant_id"`
	HostTeamID       string `json:"host_team_id"`
	ChannelID        string `json:"channel_id"`
	SharedTeamID     string `json:"shared_team_id"`
	AssociatedUserID string `json:"associated_user_id"`
}

// TestAccSharedChannelReadOnly reads an explicitly selected existing sharing
// relationship and optional associated-user view. It never establishes/removes
// sharing or changes membership, a channel, a team, or cross-tenant policy.
func TestAccSharedChannelReadOnly(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal("invalid lab enablement")
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("shared-channel lab requires explicit read-only opt-in")
	} else if err != nil {
		t.Fatal("invalid shared-channel lab enablement")
	}
	oracle := labkit.OracleEnabled(t)
	sdk, binding, _ := labkit.SDKClient(t)
	raw, err := labkit.PrivateFile(os.Getenv("MSTEAMS_LAB_SHARED_CHANNEL_FILE"))
	if err != nil {
		t.Fatal("explicit private shared-channel fixture is required")
	}
	defer clear(raw)
	var fixture sharedChannelFixture
	tenant, _, ok := strings.Cut(binding, "|")
	if json.Unmarshal(raw, &fixture) != nil || !ok || !strings.EqualFold(fixture.TenantID, tenant) || !labkit.DirectoryID.MatchString(fixture.HostTeamID) || !labkit.DirectoryID.MatchString(fixture.SharedTeamID) || strings.TrimSpace(fixture.ChannelID) == "" || fixture.AssociatedUserID != "" && !labkit.DirectoryID.MatchString(fixture.AssociatedUserID) {
		t.Fatal("invalid shared-channel fixture binding")
	}
	ctx, stop := context.WithTimeout(t.Context(), cfg.Timeout)
	defer stop()
	bounds := client.PagingOptions{MaxPages: 16, MaxItems: 1000}
	selected, _, err := sdk.GraphAPI.ChannelShares.GetByID(ctx, fixture.HostTeamID, fixture.ChannelID, fixture.SharedTeamID)
	if err != nil || selected == nil || !strings.EqualFold(selected.ID, fixture.SharedTeamID) {
		t.Fatal("selected sharing relationship read failed")
	}
	listed, err := sdk.GraphAPI.SharedChannelManagement.ListSharedWithTeams(ctx, fixture.HostTeamID, fixture.ChannelID, management.SharedTeamsOptions{Traversal: bounds})
	if err != nil {
		t.Fatal("complete sharing relationship list failed")
	}
	want := make([]labkit.SharedTeamObservation, 0, len(listed.Items))
	found := false
	for _, value := range listed.Items {
		found = found || strings.EqualFold(value.ID, fixture.SharedTeamID)
		want = append(want, labkit.SharedTeamObservation{ID: value.ID, DisplayName: value.DisplayName, TenantID: value.TenantID, IsHost: value.IsHostTeam})
	}
	if !found {
		t.Fatal("selected relationship missing from complete list")
	}
	if oracle {
		observed := labkit.RunTeamOracle(t, ctx, labkit.TeamOracleRequest{Kind: "shared-teams", ID: fixture.HostTeamID, Channel: fixture.ChannelID}).SharedTeams
		if !sameSharedTeamObservations(want, observed) {
			t.Fatal("PowerShell and Go sharing collections differ")
		}
		single := labkit.RunTeamOracle(t, ctx, labkit.TeamOracleRequest{Kind: "shared-teams", ID: fixture.HostTeamID, Channel: fixture.ChannelID, SharedTeamID: fixture.SharedTeamID}).SharedTeams
		if !sameSharedTeamObservations([]labkit.SharedTeamObservation{{ID: selected.ID, DisplayName: selected.DisplayName, TenantID: selected.TenantID, IsHost: selected.IsHostTeam}}, single) {
			t.Fatal("PowerShell and Go selected sharing fields differ")
		}
	}
	allowed, err := sdk.GraphAPI.SharedChannelManagement.ListAllowedUsers(ctx, fixture.HostTeamID, fixture.ChannelID, fixture.SharedTeamID, management.AllowedUsersOptions{Traversal: bounds, UserLookupMaxPages: 16})
	if err != nil || !allowed.Complete {
		t.Fatal("complete sharing allowed-user projection failed")
	}
	if oracle {
		users := make([]labkit.TeamOracleUser, 0, len(allowed.Items))
		for _, value := range allowed.Items {
			if value.UserID == nil {
				t.Fatal("selected allowed-user subtype cannot be compared by the pinned projection")
			}
			users = append(users, labkit.TeamOracleUser{ID: string(*value.UserID), User: value.User, Name: value.Name, Role: string(value.Role), TenantID: value.TenantID})
		}
		actual := labkit.RunTeamOracle(t, ctx, labkit.TeamOracleRequest{Kind: "shared-users", ID: fixture.HostTeamID, Channel: fixture.ChannelID, SharedTeamID: fixture.SharedTeamID}).Users
		if !sameTeamOracleUsers(users, actual) {
			t.Fatal("PowerShell and Go allowed-user projection differs")
		}
	}
	if fixture.AssociatedUserID != "" {
		associated, err := sdk.GraphAPI.SharedChannelManagement.ListAssociatedTeams(ctx, fixture.AssociatedUserID, management.AssociatedOptions{Traversal: bounds})
		if err != nil || !associated.Complete {
			t.Fatal("complete associated-user view failed")
		}
		if oracle {
			rows := make([]labkit.SharedTeamObservation, 0, len(associated.Items))
			for _, value := range associated.Items {
				rows = append(rows, labkit.SharedTeamObservation{ID: value.ID, DisplayName: value.DisplayName, TenantID: value.TenantID})
			}
			actual := labkit.RunTeamOracle(t, ctx, labkit.TeamOracleRequest{Kind: "associated-teams", ID: fixture.AssociatedUserID}).AssociatedTeams
			if !sameSharedTeamObservations(rows, actual) {
				t.Fatal("PowerShell and Go associated-user view differs")
			}
		}
	}
	t.Log("selected sharing reads observed; unshare, delegated me, cross-tenant setup and permission scenarios remain separately qualified")
}

func sameSharedTeamObservations(a, b []labkit.SharedTeamObservation) bool {
	index := func(rows []labkit.SharedTeamObservation) (map[string]labkit.SharedTeamObservation, bool) {
		out := make(map[string]labkit.SharedTeamObservation, len(rows))
		for _, value := range rows {
			if !labkit.DirectoryID.MatchString(value.ID) {
				return nil, false
			}
			value.ID = strings.ToLower(value.ID)
			if _, exists := out[value.ID]; exists {
				return nil, false
			}
			out[value.ID] = value
		}
		return out, true
	}
	left, ok := index(a)
	if !ok {
		return false
	}
	right, ok := index(b)
	return ok && reflect.DeepEqual(left, right)
}

func TestUnitSharedTeamOracleRetainsHostPresenceAndRejectsDuplicates(t *testing.T) {
	t.Parallel()
	row := labkit.SharedTeamObservation{ID: "11111111-1111-4111-8111-111111111111", IsHost: new(bool)}
	if !sameSharedTeamObservations([]labkit.SharedTeamObservation{row}, []labkit.SharedTeamObservation{row}) {
		t.Fatal("equal sharing observations differ")
	}
	if sameSharedTeamObservations([]labkit.SharedTeamObservation{row}, []labkit.SharedTeamObservation{{ID: row.ID}}) {
		t.Fatal("unknown and false host values collapsed")
	}
	if sameSharedTeamObservations([]labkit.SharedTeamObservation{row}, []labkit.SharedTeamObservation{row, row}) {
		t.Fatal("duplicate sharing identities collapsed")
	}
}
