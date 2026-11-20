package auth_test

import (
	"fmt"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

func ExampleNewApplicationProvider_workloadIdentity() {
	provider, err := auth.NewApplicationProvider(&auth.Config{
		Method: auth.WorkloadIdentity, TenantID: "11111111-1111-4111-8111-111111111111", ClientID: "22222222-2222-4222-8222-222222222222",
		AssertionFile: "/var/run/secrets/identity/assertion.jwt",
	}, &http.Client{})
	if err != nil {
		panic(err)
	}
	defer func() { provider.Close(); provider.Wait() }()
	// Only AcquireToken reads the caller-managed assertion file. Each fresh
	// exchange reloads it, supporting rotation by Kubernetes or a CI runner.
	fmt.Println(provider.Binding().Mode)
	// Output: application
}

func ExampleNewManagedIdentityProvider() {
	provider, err := auth.NewManagedIdentityProvider(&auth.Config{
		Method: auth.ManagedIdentity, TenantID: "11111111-1111-4111-8111-111111111111",
		// Optional ClientID selects a user-assigned identity on supported hosts.
	}, &http.Client{})
	if err != nil {
		panic(err)
	}
	defer func() { provider.Close(); provider.Wait() }()
	// No endpoint discovery, filesystem read or token request occurs here.
	fmt.Println(provider.Binding().IdentityID)
	// Output: system-assigned
}
