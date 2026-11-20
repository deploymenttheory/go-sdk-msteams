package graphapi_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	graphapi "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api"
	"github.com/go-resty/resty/v2"
)

func ExampleNewTeamOperation() {
	sdk, err := msteams.NewClient(&auth.Config{Method: auth.SuppliedTokens, TenantID: "fixture.example", IdentityID: "fixture", TokenMode: auth.Application, AccessTokens: map[string]auth.Token{auth.GraphAudience: {AccessToken: "placeholder", ExpiresAt: time.Now().Add(time.Hour)}}})
	if err != nil {
		panic(err)
	}
	defer func() { _ = sdk.Close() }()
	// In a service, use the actual response returned by the initiating request.
	accepted := &resty.Response{RawResponse: &http.Response{StatusCode: 202, Header: http.Header{"Location": {`/teams('fixture')/operations('job')`}}}}
	op, err := graphapi.NewTeamOperation(sdk.GraphAPI.Scope(), accepted)
	if err != nil {
		panic(err)
	}
	// Construction makes no requests. With real credentials, call op.Wait(ctx,
	// client.WaitOptions{}); retain its snapshot even when the local wait fails.
	fmt.Println(op.Snapshot().Status.State)
	// Output: accepted
}
