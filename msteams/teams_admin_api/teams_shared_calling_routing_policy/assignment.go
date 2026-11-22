package teamssharedcallingroutingpolicy

import (
	"context"
	"net/http"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

// Assign requires one explicit scope. The parameterless Grant/Identity set has
// no matching source operation; missing/mixed selection rejects before HTTP.
// Nil/empty PolicyName remain distinct; backend clear/effect semantics are separate.
func (s *Service) Assign(ctx context.Context, input policy.Assignment) (*Policy, *resty.Response, error) {
	if e := s.ready(ctx); e != nil {
		return nil, nil, e
	}
	count := 0
	if input.User != nil {
		count++
	}
	if input.Group != nil {
		count++
	}
	if input.Tenant != nil {
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
		path = []string{"Skype.Policy", "users", input.User.Identity, "policies", "TeamsSharedCallingRoutingPolicy"}
		body = map[string]any{"policyName": input.User.PolicyName, "additionalParameters": nil}
	case input.Group != nil:
		if !policyrp.ValidIdentity(input.Group.ID) {
			return nil, nil, ErrInput
		}
		method = http.MethodPost
		path = []string{"Skype.Policy", "groupPolicyAssignments", input.Group.ID, "policyTypes", "TeamsSharedCallingRoutingPolicy"}
		body = map[string]any{"policyName": input.Group.PolicyName, "priority": input.Group.Rank}
	case input.Tenant != nil:
		method = http.MethodPatch
		path = []string{"Skype.Policy", "tenants", "policies", "TeamsSharedCallingRoutingPolicy"}
		body = map[string]any{"PolicyName": input.Tenant.PolicyName, "additionalParameters": nil, "forceSwitchPresent": input.Tenant.Force}
	}
	r, e := s.resource.Do(ctx, "Grant-CsTeamsSharedCallingRoutingPolicy", method, path, body, client.ReplayNever)
	if e != nil {
		return nil, r, e
	}
	v, e := policyrp.Decode[Policy](r)
	return v, r, e
}
func (s *Service) AssignUser(ctx context.Context, id string, name *string) (*Policy, *resty.Response, error) {
	return s.Assign(ctx, policy.Assignment{User: &policy.UserAssignment{Identity: id, PolicyName: name}})
}
func (s *Service) AssignGroup(ctx context.Context, id string, name *string, rank int32) (*Policy, *resty.Response, error) {
	return s.Assign(ctx, policy.Assignment{Group: &policy.GroupAssignment{ID: id, PolicyName: name, Rank: rank}})
}
func (s *Service) AssignTenantDefault(ctx context.Context, name *string, force bool) (*Policy, *resty.Response, error) {
	return s.Assign(ctx, policy.Assignment{Tenant: &policy.TenantAssignment{PolicyName: name, Force: force}})
}
