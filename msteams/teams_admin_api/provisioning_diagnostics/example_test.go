package provisioningdiagnostics_test

import (
	"context"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	pd "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/provisioning_diagnostics"
)

// This example compiles a real SDK workflow. It performs no network operation
// during go test; the application supplies its own auth configuration and context.
// Tenant administrator credentials are denied with HTTP 403: these are
// Microsoft support diagnostics.
func ExampleService_ListObjectChangelog() {
	changes := func(ctx context.Context, cfg *auth.Config, userID string) (count int, err error) {
		sdk, err := msteams.NewClient(cfg)
		if err != nil {
			return 0, err
		}
		defer func() { err = errors.Join(err, sdk.Close()) }()
		items, _, err := sdk.TeamsAdminAPI.ProvisioningDiagnostics.ListObjectChangelog(ctx, pd.ObjectChangelogQuery{Identity: userID})
		if errors.Is(err, auth.ErrForbidden) {
			return 0, nil // Not a support identity.
		}
		return len(items), err
	}
	_ = changes
}

// Validation is the read-only first step of a hosted migration.
func ExampleService_ValidateUserMove() {
	validate := func(ctx context.Context, sdk *msteams.Client, tenantID, sipURI string) (*pd.DeploymentInfo, error) {
		deployment, _, err := sdk.TeamsAdminAPI.ProvisioningDiagnostics.ValidateUserMove(ctx, pd.ValidateUserMoveRequest{
			UserMove:        pd.UserMove{TenantID: tenantID, UserSipURI: sipURI},
			CmdletVersion:   "1.0",
			LocalDeployment: pd.LocalDeployment{PresenceFqdn: "pool.contoso.example", RegistrarFqdn: "pool.contoso.example", MajorVersion: 6},
			MoveToCloud:     true,
		})
		return deployment, err
	}
	_ = validate
}
