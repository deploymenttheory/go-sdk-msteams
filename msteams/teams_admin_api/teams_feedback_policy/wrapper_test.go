package teamsfeedbackpolicy_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_feedback_policy"
)

func TestUnitFeedbackPolicyMappedGrantForwarding(t *testing.T) {
	t.Parallel()

	fixture := struct {
		Cases []struct {
			Name    string
			Failure *string `json:"failure_type"`
			Calls   []struct {
				Operation  string
				Parameters struct {
					Identity, PolicyName, PolicyType, GroupID string
					Rank                                      *int32
					Force                                     bool `json:"forceSwitchPresent"`
				}
			}
		}
	}{
		Cases: []struct {
			Name    string
			Failure *string `json:"failure_type"`
			Calls   []struct {
				Operation  string
				Parameters struct {
					Identity, PolicyName, PolicyType, GroupID string
					Rank                                      *int32
					Force                                     bool `json:"forceSwitchPresent"`
				}
			}
		}{

			{
				Name: "grant-user",
				Calls: []struct {
					Operation  string
					Parameters struct {
						Identity, PolicyName, PolicyType, GroupID string
						Rank                                      *int32
						Force                                     bool `json:"forceSwitchPresent"`
					}
				}{

					{
						Operation: "assign-user",
						Parameters: struct {
							Identity, PolicyName, PolicyType, GroupID string
							Rank                                      *int32
							Force                                     bool `json:"forceSwitchPresent"`
						}{
							Identity:   "fixture/user",
							PolicyName: "Fixture",
							PolicyType: "TeamsFeedbackPolicy",
						},
					},
				},
			},

			{
				Name: "clear-user",
				Calls: []struct {
					Operation  string
					Parameters struct {
						Identity, PolicyName, PolicyType, GroupID string
						Rank                                      *int32
						Force                                     bool `json:"forceSwitchPresent"`
					}
				}{

					{
						Operation: "assign-user",
						Parameters: struct {
							Identity, PolicyName, PolicyType, GroupID string
							Rank                                      *int32
							Force                                     bool `json:"forceSwitchPresent"`
						}{
							Identity:   "fixture/user",
							PolicyName: "",
							PolicyType: "TeamsFeedbackPolicy",
						},
					},
				},
			},

			{
				Name: "grant-tenant",
				Calls: []struct {
					Operation  string
					Parameters struct {
						Identity, PolicyName, PolicyType, GroupID string
						Rank                                      *int32
						Force                                     bool `json:"forceSwitchPresent"`
					}
				}{

					{
						Operation: "assign-tenant",
						Parameters: struct {
							Identity, PolicyName, PolicyType, GroupID string
							Rank                                      *int32
							Force                                     bool `json:"forceSwitchPresent"`
						}{
							PolicyName: "Fixture",
							PolicyType: "TeamsFeedbackPolicy",
							Force:      false,
						},
					},
				},
			},

			{
				Name: "grant-group-rank-omitted",
				Calls: []struct {
					Operation  string
					Parameters struct {
						Identity, PolicyName, PolicyType, GroupID string
						Rank                                      *int32
						Force                                     bool `json:"forceSwitchPresent"`
					}
				}{

					{
						Operation: "assign-group",
						Parameters: struct {
							Identity, PolicyName, PolicyType, GroupID string
							Rank                                      *int32
							Force                                     bool `json:"forceSwitchPresent"`
						}{
							PolicyName: "Fixture",
							PolicyType: "TeamsFeedbackPolicy",
							GroupID:    "fixture/group",
						},
					},
				},
			},

			{
				Name: "grant-group-rank-zero",
				Calls: []struct {
					Operation  string
					Parameters struct {
						Identity, PolicyName, PolicyType, GroupID string
						Rank                                      *int32
						Force                                     bool `json:"forceSwitchPresent"`
					}
				}{

					{
						Operation: "assign-group",
						Parameters: struct {
							Identity, PolicyName, PolicyType, GroupID string
							Rank                                      *int32
							Force                                     bool `json:"forceSwitchPresent"`
						}{
							PolicyName: "Fixture",
							PolicyType: "TeamsFeedbackPolicy",
							GroupID:    "fixture/group",
						},
					},
				},
			},

			{
				Name: "grant-group-rank-two",
				Calls: []struct {
					Operation  string
					Parameters struct {
						Identity, PolicyName, PolicyType, GroupID string
						Rank                                      *int32
						Force                                     bool `json:"forceSwitchPresent"`
					}
				}{

					{
						Operation: "assign-group",
						Parameters: struct {
							Identity, PolicyName, PolicyType, GroupID string
							Rank                                      *int32
							Force                                     bool `json:"forceSwitchPresent"`
						}{
							PolicyName: "Fixture",
							PolicyType: "TeamsFeedbackPolicy",
							GroupID:    "fixture/group",
							Rank:       new(int32(2)),
						},
					},
				},
			},

			{
				Name: "grant-group-rank-negative",
				Calls: []struct {
					Operation  string
					Parameters struct {
						Identity, PolicyName, PolicyType, GroupID string
						Rank                                      *int32
						Force                                     bool `json:"forceSwitchPresent"`
					}
				}{

					{
						Operation: "assign-group",
						Parameters: struct {
							Identity, PolicyName, PolicyType, GroupID string
							Rank                                      *int32
							Force                                     bool `json:"forceSwitchPresent"`
						}{
							PolicyName: "Fixture",
							PolicyType: "TeamsFeedbackPolicy",
							GroupID:    "fixture/group",
							Rank:       new(int32(-1)),
						},
					},
				},
			},

			{
				Name:    "grant-missing-policy-name",
				Failure: new("System.Management.Automation.ParameterBindingException"),
				Calls: []struct {
					Operation  string
					Parameters struct {
						Identity, PolicyName, PolicyType, GroupID string
						Rank                                      *int32
						Force                                     bool `json:"forceSwitchPresent"`
					}
				}{},
			},

			{
				Name:    "grant-missing-scope",
				Failure: new("System.Management.Automation.ParameterBindingException"),
				Calls: []struct {
					Operation  string
					Parameters struct {
						Identity, PolicyName, PolicyType, GroupID string
						Rank                                      *int32
						Force                                     bool `json:"forceSwitchPresent"`
					}
				}{},
			},

			{
				Name:    "grant-passthru-true-rejected",
				Failure: new("System.Management.Automation.ParameterBindingException"),
				Calls: []struct {
					Operation  string
					Parameters struct {
						Identity, PolicyName, PolicyType, GroupID string
						Rank                                      *int32
						Force                                     bool `json:"forceSwitchPresent"`
					}
				}{},
			},

			{
				Name:    "grant-passthru-false-rejected",
				Failure: new("System.Management.Automation.ParameterBindingException"),
				Calls: []struct {
					Operation  string
					Parameters struct {
						Identity, PolicyName, PolicyType, GroupID string
						Rank                                      *int32
						Force                                     bool `json:"forceSwitchPresent"`
					}
				}{},
			},
		},
	}

	for _, f := range fixture.Cases {
		if f.Failure != nil {
			if *f.Failure != "System.Management.Automation.ParameterBindingException" || len(f.Calls) != 0 {
				t.Fatal("source binding boundary changed")
			}
			if f.Name != "grant-missing-scope" {
				continue
			}
			run(t, contracttest.Case{Name: f.Name, Source: wrapperSource, Mappings: []contracttest.Mapping{{Command: grantCommand, ParameterSet: "Identity"}}, Want: contracttest.Expected{Error: resource.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				_, r, e := sdk.TeamsAdminAPI.TeamsFeedbackPolicies.Assign(ctx, resource.Assignment{})
				return contracttest.Observation{Response: r, Err: e}
			}})
			continue
		}
		if len(f.Calls) != 1 || f.Calls[0].Parameters.PolicyType != "TeamsFeedbackPolicy" {
			t.Fatal("unexpected mapped forwarding")
		}
		call := f.Calls[0]
		p := call.Parameters
		set, method, path := "Identity", "PATCH", "/Skype.Policy/users/fixture%2Fuser/policies/TeamsFeedbackPolicy"
		status := 200
		body := map[string]any{"policyName": p.PolicyName}
		input := resource.Assignment{User: &resource.UserAssignment{Identity: p.Identity, PolicyName: p.PolicyName}}
		switch call.Operation {
		case "assign-user":
		case "assign-tenant":
			set, path = "GrantToTenant", "/Skype.Policy/tenants/policies/TeamsFeedbackPolicy"
			body["forceSwitchPresent"] = p.Force
			input = resource.Assignment{Tenant: &resource.TenantAssignment{PolicyName: p.PolicyName}}
		case "assign-group":
			set, method, path, status = "GrantToGroup", "POST", "/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsFeedbackPolicy", 204
			if p.Rank != nil {
				body["priority"] = *p.Rank
			}
			rank := p.Rank
			if f.Name == "grant-group-rank-zero" {
				rank = new(int32(0))
			}
			input = resource.Assignment{Group: &resource.GroupGrant{ID: p.GroupID, PolicyName: p.PolicyName, Rank: rank}}
		default:
			t.Fatal("unknown mapped grant operation")
		}
		encoded, e := json.Marshal(body)
		if e != nil {
			t.Fatal(e)
		}
		run(t, contracttest.Case{Name: f.Name, Source: wrapperSource, Mappings: []contracttest.Mapping{{Command: grantCommand, ParameterSet: set}}, Steps: []contracttest.Exchange{discovery(t), modernWire(method, path, encoded, status, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			_, r, e := sdk.TeamsAdminAPI.TeamsFeedbackPolicies.Assign(ctx, input)
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
}
