package graphapi_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	graphapi "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api"
)

func ExampleNewPager() {
	sdk, err := msteams.NewClient(&auth.Config{Method: auth.SuppliedTokens, TenantID: "fixture.example", IdentityID: "fixture", TokenMode: auth.Application, AccessTokens: map[string]auth.Token{auth.GraphAudience: {AccessToken: "placeholder", ExpiresAt: time.Now().Add(time.Hour)}}})
	if err != nil {
		panic(err)
	}
	defer func() { _ = sdk.Close() }()
	type group struct {
		ID string `json:"id"`
	}
	pager, err := graphapi.NewPager[group](sdk.GraphAPI.Scope(), client.Request{Method: http.MethodGet, Path: []string{"v1.0", "groups"}}, client.PagingOptions{MaxPages: 100, MaxItems: 10000})
	if err != nil {
		panic(err)
	}
	// Construction performs no requests. With real credentials and permissions,
	// call Next(ctx) or Collect(ctx), and check errors before treating data as complete.
	fmt.Println(pager)
	// Output: msteams pager (redacted)
}
