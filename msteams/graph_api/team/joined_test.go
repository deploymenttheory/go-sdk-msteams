package team_test

import (
	"errors"
	"net/url"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

func TestJoinedTeamsArePagedBasicProjectionsWithoutInitialOData(t *testing.T) {
	t.Parallel()
	path := "/v1.0/users/" + userID + "/joinedTeams"
	next := "https://graph.microsoft.com" + path + "?$skiptoken=opaque%2B%2F%3D"
	sdk, _, calls := fixture(t, "commercial", auth.Application,
		exchange{method: "GET", path: path, status: 200, body: `{"value":[{"id":"` + teamID + `","displayName":null,"isArchived":false,"future":7}],"@odata.nextLink":"` + next + `"}`},
		exchange{method: "GET", path: path, query: url.Values{"$skiptoken": {"opaque+/="}}, status: 200, body: `{"value":[]}`})
	result, err := sdk.GraphAPI.Teams.ListJoined(t.Context(), userID, client.PagingOptions{})
	if err != nil || *calls != 2 || result.Pages != 2 || len(result.Items) != 1 || result.Items[0].DisplayName != nil || result.Items[0].IsArchived == nil || *result.Items[0].IsArchived || string(result.Items[0].RawFields["future"]) != "7" {
		t.Fatal("joined projections, nullability or opaque paging changed")
	}
}

func TestJoinedTeamsRejectInvalidBoundsAndSuccessfulNon200(t *testing.T) {
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
