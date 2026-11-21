package policypackages_test

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_packages"
)

func TestPackageDocumentedApplicationRestrictionBeforeCredentials(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, cloud := range []string{"commercial", "gcc", "gcch", "dod"} {
		c := contracttest.Case{Name: "application-mode-" + cloud, Mode: auth.Application, Cloud: cloud, Source: "contracts/endpoints/8.1.0/fixtures/policy-package-mode.json", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			for _, name := range []string{"list", "get", "user-get", "user-recommendations", "create", "update", "delete", "group-assign", "user-grant", "batch-grant"} {
				o := invoke(name, ctx, sdk)
				if !errors.Is(o.Err, client.ErrAuthenticationMode) || o.Response != nil {
					return contracttest.Observation{Err: errors.New("documented app exception acquired credentials or allowed a request")}
				}
			}
			return contracttest.Observation{}
		}}
		results = append(results, run(t, c))
	}
	evidence(t, results)
}

func TestPackageInputContextAndBackendBeforeCredentials(t *testing.T) {
	t.Parallel()
	r := run(t, contracttest.Case{Name: "input-context-before-tokens", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.PolicyPackages
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nid"} {
			if _, _, e := s.GetByName(ctx, id); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid name accepted")}
			}
			if _, _, e := s.DeleteByIdentity(ctx, resource.Identity{PackageName: id}); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid InputObject accepted")}
			}
		}
		for _, id := range []string{"", " ", "user\r\nfixture: injection"} {
			if _, _, e := s.GetForUser(ctx, id); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid user header accepted")}
			}
		}
		if _, _, e := s.Create(ctx, resource.CustomPackageRequest{}); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("unbound create accepted")}
		}
		if _, _, e := s.AssignUsers(ctx, resource.UserAssignmentRequest{}); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("unbound user collection accepted")}
		}
		if _, _, e := s.AssignGroup(ctx, resource.GroupAssignmentRequest{}); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("unbound group accepted")}
		}
		graph, e := client.NewScope(client.Graph, sdk.TeamsAdminAPI.Scope().Session())
		if e != nil {
			return contracttest.Observation{Err: e}
		}
		if _, _, e := resource.NewService(graph).List(ctx); !errors.Is(e, client.ErrDestination) {
			return contracttest.Observation{Err: errors.New("wrong backend accepted")}
		}
		for _, missing := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := missing.List(ctx); !errors.Is(e, client.ErrClosed) {
				return contracttest.Observation{Err: errors.New("missing service accepted")}
			}
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.Create(canceled, customInput(false)); !errors.Is(e, context.Canceled) {
			return contracttest.Observation{Err: errors.New("cancellation lost")}
		}
		return contracttest.Observation{}
	}})
	evidence(t, []contracttest.Result{r})
}

func TestPackageMutationErrorsKeepDataAndNeverReplay(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, row := range captures(t) {
		if row.Name != "create" && row.Name != "update" && row.Name != "delete" && row.Name != "group-assign" && row.Name != "user-grant" && row.Name != "batch-grant" {
			continue
		}
		for _, status := range []int{401, 403, 429, 503} {
			f := row
			f.Status, f.Reply = status, `{"ResponseMessage":"fixture-private"}`
			c := contracttest.Case{Name: fmt.Sprintf("%s-denial-%d", f.Name, status), Mappings: mapping(f.Name), Steps: []contracttest.Exchange{discovery(), wire(f, mapping(f.Name)[0].Command)}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				o := invoke(f.Name, ctx, sdk)
				if o.Err == nil || strings.Contains(fmt.Sprintf("%v %#v", o.Err, o.Err), "fixture-private") {
					o.Err = errors.New("denial lost or leaked a business message")
				}
				switch v := o.Value.(type) {
				case *resource.MessageResult:
					if v == nil || v.HTTPAccepted || v.ResponseMessage == nil {
						o.Err = errors.New("typed custom-package failure lost")
					}
				case *resource.AssignmentResult:
					if v == nil || v.HTTPAccepted || v.ResponseMessage == nil {
						o.Err = errors.New("typed assignment failure lost")
					}
				}
				return o
			}}
			results = append(results, run(t, c))
		}
	}
	evidence(t, results)
}

func TestPackageDiscoveryFailureIsNotBusinessResponse(t *testing.T) {
	t.Parallel()
	step := discovery()
	step.Reply.Status = 403
	step.Reply.Body = []byte(`{"ResponseMessage":"fixture-not-business"}`)
	c := contracttest.Case{Name: "discovery-denial", Steps: []contracttest.Exchange{step}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 403}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.PolicyPackages.SubmitBatch(ctx, userInput())
		if v != nil {
			e = errors.New("discovery error was misclassified as package assignment data")
		}
		return contracttest.Observation{Response: r, Err: e}
	}}
	evidence(t, []contracttest.Result{run(t, c)})
}

func TestPackageGroupDefaultRankingsMatchExportedWrapper(t *testing.T) {
	t.Parallel()
	f := captured{Method: "POST", URL: origin + "/Teams.VerticalPackaging/package/group", Body: []byte(`{"GroupId":"fixture/group","PolicyRankings":[]}`), Status: 200, Reply: `{}`}
	c := contracttest.Case{Name: "group-default-rankings", Source: "contracts/endpoints/8.1.0/fixtures/policy-package-binding.json", Mappings: mapping("group-assign"), Steps: []contracttest.Exchange{discovery(), wire(f, "Grant-CsGroupPolicyPackageAssignment")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.PolicyPackages.AssignGroup(ctx, resource.GroupAssignmentRequest{GroupID: "fixture/group"})
		return contracttest.Observation{Value: v, Response: r, Err: e}
	}}
	evidence(t, []contracttest.Result{run(t, c)})
}
