package tenantfederationconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	configuration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_federation_configuration"
)

func TestTenantFederationPinnedDispatch(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
		for _, f := range pinned(t).Cases {
			command := getCommand
			if f.Method == "PUT" {
				command = setCommand
			}
			body := f.Body
			if string(body) == "null" {
				body = nil
			}
			results = append(results, run(t, contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: command, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{discovery(), wire(f.Method, strings.TrimPrefix(f.URL, origin), command, body, f.Status, f.Response)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TenantFederationConfigurations
				switch f.Name {
				case "tenant-federation-list":
					v, r, e := s.List(ctx)
					if e == nil && (v == nil || len(v) != 0) {
						e = errors.New("empty list changed")
					}
					return contracttest.Observation{Response: r, Err: e}
				case "tenant-federation-read":
					v, r, e := s.GetByIdentity(ctx, "Global")
					if e == nil && (v.AllowFederatedUsers == nil || *v.AllowFederatedUsers || string(v.AllowedDomains) != "null") {
						e = errors.New("false/null read changed")
					}
					return contracttest.Observation{Response: r, Err: e}
				case "tenant-federation-update":
					no := false
					r, e := s.UpdateGlobal(ctx, configuration.UpdateRequest{Settings: configuration.Settings{AllowFederatedUsers: &no}, AllowedDomains: configuration.NewAllowList("fixture.example.invalid")})
					return contracttest.Observation{Response: r, Err: e}
				default:
					return contracttest.Observation{Err: errors.New("unknown pinned request")}
				}
			}}))
		}
	}
	evidence(t, results)
}

func inputFor(t *testing.T, name string) configuration.UpdateRequest {
	t.Helper()
	var v configuration.UpdateRequest
	a, b := "a.example.invalid", "b.example.invalid"
	object := json.RawMessage(`{"Domain":"object.example.invalid","Enabled":false,"DataSource":"fixture-private","Key":"fixture-key"}`)
	text := json.RawMessage(`"a.example.invalid"`)
	var err error
	switch name {
	case "omitted", "bound-null-scalar":
	case "bound-scalars":
		no := false
		future, empty := "FutureState", ""
		v.Settings = configuration.Settings{AllowFederatedUsers: &no, AllowTeamsConsumer: &no, AllowTeamsConsumerInbound: &no, RestrictTeamsConsumerToExternalUserProfiles: &no, SharedSipAddressSpace: &no, TreatDiscoveredPartnersAsUnverified: &no, BlockAllSubdomains: &no, EnableExternalAccessRestrictionsForChatParticipants: &no, EnableMutualFederationForChatParticipants: &no, ExternalAccessWithTrialTenants: &future, SecurityTeamAllowBlockListDelegation: &empty}
	case "allow-all":
		v.AllowedDomains = configuration.NewAllowAllKnownDomains()
	case "allow-list":
		v.AllowedDomains = configuration.NewAllowList(b, a, b, "")
	case "empty-allow-list":
		v.AllowedDomains = configuration.NewAllowList()
	case "allowed-replacement":
		v.AllowedDomainsAsAList = configuration.ReplaceDomains([]string{a, b})
	case "allowed-operators":
		v.AllowedDomainsAsAList = configuration.ChangeDomains([]string{a}, []string{b})
	case "allowed-empty-operators":
		v.AllowedDomainsAsAList = configuration.ChangeDomains(nil, nil)
	case "blocked-mixed":
		v.BlockedDomains, err = configuration.ReplaceDomainValues([]json.RawMessage{text, object})
	case "blocked-operators":
		v.BlockedDomains, err = configuration.ChangeDomainValues([]json.RawMessage{object, text}, []json.RawMessage{json.RawMessage(`"b.example.invalid"`)})
	case "trial-mixed":
		v.AllowedTrialTenantDomains, err = configuration.ReplaceDomainValues([]json.RawMessage{object, text})
	case "trial-null":
		v.AllowedTrialTenantDomains = configuration.ReplaceDomains(nil)
	default:
		t.Fatal("unknown accepted model")
	}
	if err != nil {
		t.Fatal(err)
	}
	return v
}

func TestTenantUpdateMatchesPinnedNormalization(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	rejected := 0
	for _, f := range pinned(t).Models {
		if f.Rejected {
			rejected++
			continue
		}
		input := inputFor(t, f.Name)
		results = append(results, run(t, contracttest.Case{Name: f.Name, Mode: auth.Application, Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{discovery(), wire("PUT", collection+"/configuration/Global", setCommand, f.Wire, 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			r, e := sdk.TeamsAdminAPI.TenantFederationConfigurations.UpdateGlobal(ctx, input)
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	if rejected != 3 {
		t.Fatal("pinned rejection count differs")
	}
	evidence(t, results)
}

func TestTenantListFilteringAndRawChoices(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	reply := `[{"Identity":"tag:Fixture","AllowedDomains":{"AllowedDomain":[{"Domain":"a.example.invalid"}],"Future":false},"BlockedDomains":null},{"Identity":"Global","AllowedDomains":null,"AllowedTrialTenantDomains":[],"AllowTeamsConsumer":false,"Future":{"x":null}},{"Identity":"Other","ExternalAccessWithTrialTenants":"FutureState"}]`
	for index, pattern := range []string{"", "Fixture*"} {
		results = append(results, run(t, contracttest.Case{Name: fmt.Sprint("list-", index), Mode: auth.Delegated, Mappings: []contracttest.Mapping{{Command: getCommand, ParameterSet: "Filter"}}, Steps: []contracttest.Exchange{discovery(), wire("GET", collection, getCommand, nil, 200, reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.TenantFederationConfigurations.ListFiltered(ctx, pattern)
			if e == nil {
				want := []string{"Global", "tag:Fixture", "Other"}
				if pattern != "" {
					want = []string{"tag:Fixture"}
				}
				var ids []string
				for _, item := range v {
					ids = append(ids, item.Identity)
					raw, err := json.Marshal(item)
					var fields map[string]json.RawMessage
					if err != nil || json.Unmarshal(raw, &fields) != nil || !reflect.DeepEqual(fields, item.RawFields) {
						e = errors.New("response snapshot changed")
					}
				}
				if !reflect.DeepEqual(ids, want) {
					e = errors.New("local filter/order changed")
				}
			}
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}

func TestTenantIdentityEscapingAndMutationErrors(t *testing.T) {
	t.Parallel()
	identity := "scope/Fixture / Config"
	run(t, contracttest.Case{Name: "escaped", Mode: auth.Application, Steps: []contracttest.Exchange{discovery(), wire("GET", collection+"/configuration/"+url.PathEscape(identity), getCommand, nil, 200, `{"Identity":"scope/Fixture / Config","AllowedDomains":["Future"]}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.TenantFederationConfigurations.GetByIdentity(ctx, identity)
		if e == nil && (v.Identity != identity || string(v.AllowedDomains) != `["Future"]`) {
			e = errors.New("escaped read changed")
		}
		return contracttest.Observation{Response: r, Err: e}
	}})
	for _, status := range []int{400, 403, 429, 503} {
		run(t, contracttest.Case{Name: fmt.Sprint("denied-", status), Mode: auth.Application, Steps: []contracttest.Exchange{discovery(), wire("PUT", collection+"/configuration/Global", setCommand, json.RawMessage(`{"Identity":"Global"}`), status, `{"Code":"FutureDenied","Message":"fixture private text"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			r, e := sdk.TeamsAdminAPI.TenantFederationConfigurations.UpdateGlobal(ctx, configuration.UpdateRequest{})
			var api *client.APIError
			if !errors.As(e, &api) || api.StatusCode != status || r == nil || !strings.Contains(string(r.Body()), "fixture private text") || strings.Contains(fmt.Sprint(e), "fixture private text") {
				return contracttest.Observation{Err: errors.New("error classification/metadata/redaction differs")}
			}
			return contracttest.Observation{Response: r}
		}})
	}
}

func TestTenantInvalidInputsDoNotAcquireCredentials(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "invalid-local", Mode: auth.Application, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.TenantFederationConfigurations
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, configuration.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid read identity accepted")}
			}
			if _, e := s.UpdateByIdentity(ctx, id, configuration.UpdateRequest{}); !errors.Is(e, configuration.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid update identity accepted")}
			}
		}
		for _, input := range []configuration.UpdateRequest{{AllowedDomains: configuration.NewAllowList(), AllowedDomainsAsAList: configuration.ReplaceDomains(nil)}, {AllowedDomains: &configuration.AllowedDomains{}}, {BlockedDomains: &configuration.ListChange{}}} {
			if _, e := s.UpdateGlobal(ctx, input); !errors.Is(e, configuration.ErrInput) {
				return contracttest.Observation{Err: errors.New("invalid domain input accepted")}
			}
		}
		if _, _, e := s.ListFiltered(ctx, "[broken"); !errors.Is(e, configuration.ErrInput) {
			return contracttest.Observation{Err: errors.New("invalid filter accepted")}
		}
		c, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.List(c); !errors.Is(e, context.Canceled) {
			return contracttest.Observation{Err: errors.New("cancellation changed")}
		}
		return contracttest.Observation{}
	}})
}
