package auth_test

import (
	"context"
	"fmt"
	"net/http"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

func ExampleNewDelegatedProvider() {
	provider, err := auth.NewDelegatedProvider(&auth.Config{
		Method:       auth.DeviceCode,
		TenantID:     "11111111-1111-4111-8111-111111111111",
		ClientID:     "22222222-2222-4222-8222-222222222222",
		LoginTimeout: 5 * time.Minute,
		DevicePrompt: func(ctx context.Context, message auth.DeviceCodeMessage) error {
			if err := ctx.Err(); err != nil {
				return err
			}
			// Present to the signing-in user; do not send codes to application logs.
			fmt.Printf("Visit %s and enter %s\n", message.VerificationURL, message.UserCode)
			return nil
		},
	}, &http.Client{Timeout: 30 * time.Second})
	if err != nil {
		fmt.Println("invalid configuration")
		return
	}
	defer func() { provider.Close(); provider.Wait() }()
	// Only AcquireToken starts sign-in. Both resource audiences subsequently use
	// the same selected account. No credential chain or interactive fallback exists.
	fmt.Println(provider.Binding().Mode)
	// Output: delegated
}
