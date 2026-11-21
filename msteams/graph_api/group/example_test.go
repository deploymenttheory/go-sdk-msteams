package group_test

import (
	"context"
	"fmt"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
)

func ExampleNewMicrosoft365Group() {
	value, err := group.NewMicrosoft365Group("Fixture Group", "fixture-group")
	if err != nil {
		panic(err)
	}
	fmt.Println(value.GroupTypes[0], value.MailEnabled, value.SecurityEnabled, value.Visibility == nil)
	// Output: Unified true false true
}

// Compile-only: actual credentials, owner identity and lifecycle recovery belong
// to the application. This example performs no tenant calls in ordinary tests.
func ExampleService_Create() {
	create := func(ctx context.Context, sdk *msteams.Client, ownerID string) (string, error) {
		value, err := group.NewMicrosoft365Group("SDK lab fixture", "sdk-lab-fixture")
		if err != nil {
			return "", err
		}
		value.OwnerIDs = []string{ownerID}
		value.MemberIDs = []string{ownerID}
		created, _, err := sdk.GraphAPI.Groups.Create(ctx, value)
		if err != nil {
			return "", err
		}
		// Preserve this ID before a separate CreateFromGroup stage. Do not
		// automatically delete it after an ambiguous teamification outcome.
		return created.ID, nil
	}
	_ = create
}
