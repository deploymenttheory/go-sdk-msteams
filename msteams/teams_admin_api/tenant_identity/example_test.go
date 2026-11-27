package tenantidentity_test

import (
	"context"
	"errors"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	ti "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_identity"
)

// This example compiles a real SDK workflow. It performs no network operation
// during go test; the application supplies its own auth configuration and context.
func ExampleService_GetTenant() {
	domains := func(ctx context.Context, cfg *auth.Config) (names []string, err error) {
		sdk, err := msteams.NewClient(cfg)
		if err != nil {
			return nil, err
		}
		defer func() { err = errors.Join(err, sdk.Close()) }()
		tenant, _, err := sdk.TeamsAdminAPI.TenantIdentity.GetTenant(ctx)
		if err != nil {
			return nil, err
		}
		for _, domain := range tenant.VerifiedDomains {
			if domain.Name != nil {
				names = append(names, *domain.Name)
			}
		}
		return names, nil
	}
	_ = domains
}

// SearchGroups finds groups for voice application routing. The service needs
// at least three characters.
func ExampleService_SearchGroups() {
	find := func(ctx context.Context, sdk *msteams.Client, name string) (string, error) {
		exact := true
		groups, _, err := sdk.TeamsAdminAPI.TenantIdentity.SearchGroups(ctx, ti.GroupSearchRequest{SearchQuery: name, ExactMatchOnly: &exact})
		if err != nil {
			return "", err
		}
		if len(groups) != 1 || groups[0] == nil || groups[0].ID == nil {
			return "", errors.New("group not found or ambiguous")
		}
		return *groups[0].ID, nil
	}
	_ = find
}
