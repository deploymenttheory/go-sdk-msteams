// Command client-construction demonstrates local client wiring without signing in.
package main

import (
	"fmt"
	"log"
	"time"

	msteams "github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

func main() {
	sdk, err := msteams.NewClient(&msteams.AuthConfig{
		Method: auth.ManagedIdentity, TenantID: "example-tenant",
	}, msteams.WithHTTPTimeout(30*time.Second))
	if err != nil {
		log.Fatal(err)
	}
	fmt.Println(sdk.GraphAPI.Scope().Backend(), sdk.TeamsAdminAPI.Scope().Backend())
	if err := sdk.Close(); err != nil {
		log.Fatal(err)
	}
}
