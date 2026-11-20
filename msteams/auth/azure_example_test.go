package auth_test

import (
	"context"
	"fmt"

	"github.com/Azure/azure-sdk-for-go/sdk/azidentity"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

func ExampleNewAzureTokenProvider() {
	const tenantID = "11111111-1111-4111-8111-111111111111"
	credential, err := azidentity.NewDeviceCodeCredential(&azidentity.DeviceCodeCredentialOptions{
		TenantID: tenantID,
		ClientID: "22222222-2222-4222-8222-222222222222",
		UserPrompt: func(ctx context.Context, message azidentity.DeviceCodeMessage) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			// Present only to the signing-in user, not to application logs.
			fmt.Printf("Visit %s and enter %s\n", message.VerificationURL, message.UserCode)
			return nil
		},
	})
	if err != nil {
		fmt.Println("invalid credential configuration")
		return
	}
	binding := auth.Binding{TenantID: tenantID, Cloud: "commercial", IdentityID: "fixture-user", Mode: auth.Delegated}
	provider, err := auth.NewAzureTokenProvider(binding, credential)
	if err != nil {
		fmt.Println("invalid provider configuration")
		return
	}
	// Pass binding and provider to msteams.WithTokenProvider. Construction does not
	// authenticate; an explicit token request or SDK operation starts acquisition.
	fmt.Println(provider)
	// Output: msteams bound provider (redacted)
}
