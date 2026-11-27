package externalaccesspolicy

import (
	"context"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// Assign requires exactly one explicit scope. The compiled Grant default set
// has no resource operation; native Go rejects absent scope before credentials.
// PolicyName nil/empty remain distinct; their server clear/default effects are
// unqualified. A group response remains an optional policy snapshot, not a job.
func (s *Service) Assign(ctx context.Context, input policy.Assignment) (*Policy, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	count := 0
	if input.User != nil {
		count++
	}
	if input.Tenant != nil {
		count++
	}
	if input.Group != nil {
		count++
	}
	if count != 1 {
		return nil, nil, ErrInput
	}
	var method string
	var path []string
	var body map[string]any
	switch {
	case input.User != nil:
		if !policyrp.ValidIdentity(input.User.Identity) {
			return nil, nil, ErrInput
		}
		method = http.MethodPatch
		path = []string{"Skype.Policy", "users", input.User.Identity, "policies", "ExternalAccessPolicy"}
		body = map[string]any{"policyName": input.User.PolicyName, "additionalParameters": nil}
	case input.Tenant != nil:
		method = http.MethodPatch
		path = []string{"Skype.Policy", "tenants", "policies", "ExternalAccessPolicy"}
		body = map[string]any{"PolicyName": input.Tenant.PolicyName, "additionalParameters": nil, "forceSwitchPresent": input.Tenant.Force}
	case input.Group != nil:
		if !policyrp.ValidIdentity(input.Group.ID) {
			return nil, nil, ErrInput
		}
		method = http.MethodPost
		path = []string{"Skype.Policy", "groupPolicyAssignments", input.Group.ID, "policyTypes", "ExternalAccessPolicy"}
		body = map[string]any{"policyName": input.Group.PolicyName, "priority": input.Group.Rank}
	}
	response, err := s.resource.Do(ctx, "Grant-CsExternalAccessPolicy", method, path, body, client.ReplayNever)
	if err != nil {
		return nil, response, err
	}
	value, err := policyrp.Decode[Policy](response)
	return value, response, err
}

func (s *Service) AssignUser(ctx context.Context, id string, name *string) (*Policy, *resty.Response, error) {
	return s.Assign(ctx, policy.Assignment{User: &policy.UserAssignment{Identity: id, PolicyName: name}})
}
func (s *Service) AssignTenantDefault(ctx context.Context, name *string, force bool) (*Policy, *resty.Response, error) {
	return s.Assign(ctx, policy.Assignment{Tenant: &policy.TenantAssignment{PolicyName: name, Force: force}})
}
func (s *Service) AssignGroup(ctx context.Context, id string, name *string, rank int32) (*Policy, *resty.Response, error) {
	return s.Assign(ctx, policy.Assignment{Group: &policy.GroupAssignment{ID: id, PolicyName: name, Rank: rank}})
}
