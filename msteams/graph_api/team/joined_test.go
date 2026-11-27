package team_test

import (
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func TestUnitJoinedTeamsRequestHasNoInitialODataOnDenial(t *testing.T) {
	t.Parallel()
	sdk, _, calls := fixture(t, "commercial", auth.Application, exchange{method: "GET", path: "/v1.0/users/" + userID + "/joinedTeams", status: 403, body: `{"error":{"code":"UnitRequestDenied"}}`})
	result, err := sdk.GraphAPI.Teams.ListJoined(t.Context(), userID, client.PagingOptions{})
	if !errors.Is(err, auth.ErrForbidden) || *calls != 1 || result.Pages != 0 || len(result.Items) != 0 {
		t.Fatal("denied joined-team read fabricated projections or pagination")
	}
}

func TestUnitJoinedTeamsRejectInvalidBoundsAndSuccessfulNon200(t *testing.T) {
	t.Parallel()
	sdk, provider, _ := fixture(t, "commercial", auth.Application)
	if _, err := sdk.GraphAPI.Teams.NewJoinedPager("../private", client.PagingOptions{}); !errors.Is(err, client.ErrRequest) {
		t.Fatal("invalid directory user ID accepted")
	}
	if _, err := sdk.GraphAPI.Teams.NewJoinedPager(userID, client.PagingOptions{MaxPages: -1}); !errors.Is(err, client.ErrPagination) {
		t.Fatal("invalid paging bounds accepted")
	}
	if provider.calls != 0 {
		t.Fatal("joined pager construction acquired credentials")
	}
	for _, status := range []int{202, 206} {
		sdk, _, _ := fixture(t, "commercial", auth.Application, exchange{method: "GET", path: "/v1.0/users/" + userID + "/joinedTeams", status: status, body: `{"value":[]}`})
		result, err := sdk.GraphAPI.Teams.ListJoined(t.Context(), userID, client.PagingOptions{})
		if !errors.Is(err, client.ErrPageStatus) || len(result.Items) != 0 || result.Response == nil || result.Response.StatusCode() != status {
			t.Fatal("unexpected success established a complete joined collection")
		}
	}
}
