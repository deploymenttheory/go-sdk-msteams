package auth_test

import (
	"fmt"
	"net/http"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

func ExampleNewApplicationProvider() {
	provider, err := auth.NewApplicationProvider(&auth.Config{
		Method:       auth.ClientSecret,
		TenantID:     "11111111-1111-4111-8111-111111111111",
		ClientID:     "22222222-2222-4222-8222-222222222222",
		ClientSecret: "synthetic-example-secret",
	}, &http.Client{Timeout: 30 * time.Second})
	if err != nil {
		fmt.Println("invalid configuration")
		return
	}
	defer func() { provider.Close(); provider.Wait() }()
	// Acquisition is explicit: provider.AcquireToken(ctx,
	//     provider.Binding().Request(auth.GraphAudience)).
	// Never print a token or raw credential material.
	fmt.Println(provider.Binding().Mode)
	// Output: application
}
