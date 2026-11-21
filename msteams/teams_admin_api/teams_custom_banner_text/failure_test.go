package teamscustombannertext_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_custom_banner_text"
)

func TestBannerMutationsNeverReplayInEitherSelector(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, byID := range []bool{false, true} {
		for _, op := range []string{"create", "patch", "delete"} {
			for _, status := range []int{400, 401, 403, 429, 503} {
				method, command, path, set, body := "POST", newCommand, collection, "Identity", `{"Identity":"Global/`+relative+`","Id":"`+relative+`","Text":"fixture"}`
				if byID {
					set = "ParentAndRelativeKey"
					body = `{"Identity":"Global/` + relative + `","Id":null,"Text":"fixture"}`
				}
				if op == "patch" {
					method, command, path = "PATCH", setCommand, item
					body = `{"Identity":"Global/` + relative + `","Description":null}`
					if byID {
						body = `{"Id":null,"Description":null}`
					}
				}
				if op == "delete" {
					method, command, path = "DELETE", deleteCommand, item
					body = `{"Identity":"Global/` + relative + `"}`
					if byID {
						body = `{"Id":"` + relative + `"}`
					}
				}
				results = append(results, run(t, contracttest.Case{Name: fmt.Sprintf("%s-%s-%d", strings.ToLower(set), op, status), Mappings: []contracttest.Mapping{{Command: command, ParameterSet: set}}, Steps: []contracttest.Exchange{wire(method, path, command, json.RawMessage(body), status, `{"Message":"fixture-private"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					s := sdk.TeamsAdminAPI.TeamsCustomBannerTexts
					var o contracttest.Observation
					switch op {
					case "create":
						input := resource.CreateRequest{Text: "fixture"}
						if byID {
							_, r, e := s.CreateByID(ctx, relative, input)
							o = contracttest.Observation{Response: r, Err: e}
						} else {
							_, r, e := s.CreateByIdentity(ctx, "Global/"+relative, input)
							o = contracttest.Observation{Response: r, Err: e}
						}
					case "patch":
						input := resource.PatchRequest{Settings: resource.Settings{Description: policy.Null[string]()}}
						if byID {
							input.BodyID = policy.Null[string]()
							_, r, e := s.PatchByID(ctx, relative, input)
							o = contracttest.Observation{Response: r, Err: e}
						} else {
							_, r, e := s.PatchByIdentity(ctx, "Global/"+relative, input)
							o = contracttest.Observation{Response: r, Err: e}
						}
					case "delete":
						if byID {
							_, r, e := s.DeleteByID(ctx, relative, policy.Controls{})
							o = contracttest.Observation{Response: r, Err: e}
						} else {
							_, r, e := s.DeleteByIdentity(ctx, "Global/"+relative, policy.Controls{})
							o = contracttest.Observation{Response: r, Err: e}
						}
					}
					var api *client.APIError
					if !errors.As(o.Err, &api) || api.StatusCode != status || strings.Contains(fmt.Sprint(o.Err), "fixture-private") {
						return contracttest.Observation{Err: errors.New("mutation status/redaction changed")}
					}
					o.Err = nil
					return o
				}}))
			}
		}
	}
	evidence(t, results)
}
func TestBannerLocalInputContextAndBackendGuards(t *testing.T) {
	t.Parallel()
	r := run(t, contracttest.Case{Name: "local-input-context-backend", Source: binderSource, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.TeamsCustomBannerTexts
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, resource.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid identity accepted")}
			}
		}
		if _, _, e := s.ListFiltered(ctx, ""); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("empty Filter accepted")}
		}
		if _, _, e := s.Create(ctx, resource.CreateRequest{}); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("mandatory New text missing")}
		}
		settings := resource.Settings{Text: policy.Null[string]()}
		if _, _, e := s.UpdateByID(ctx, relative, resource.UpdateRequest{Settings: settings}); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("null Text performed a read")}
		}
		if _, _, e := s.UpdateByID(ctx, relative, resource.UpdateRequest{}); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("empty Set performed a read")}
		}
		valid := resource.Settings{Text: policy.Set("")}
		if _, _, e := s.PatchByID(ctx, relative, resource.PatchRequest{Settings: valid}); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("missing read-model Id accepted")}
		}
		if _, _, e := s.PatchByID(ctx, relative, resource.PatchRequest{Settings: valid, BodyID: policy.Set("invalid")}); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("invalid model Id accepted")}
		}
		if _, _, e := s.PatchByID(ctx, relative, resource.PatchRequest{Settings: valid, BodyID: policy.Null[string](), BodyIdentity: policy.Null[string]()}); !errors.Is(e, resource.ErrInput) {
			return contracttest.Observation{Err: errors.New("mixed body identifiers accepted")}
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
		if _, _, e := s.Create(canceled, resource.CreateRequest{Text: "fixture"}); !errors.Is(e, context.Canceled) {
			return contracttest.Observation{Err: errors.New("canceled create acquired credentials")}
		}
		return contracttest.Observation{}
	}})
	evidence(t, []contracttest.Result{r})
}
