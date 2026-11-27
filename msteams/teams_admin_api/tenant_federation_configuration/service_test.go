package tenantfederationconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/url"

	"slices"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	configuration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_federation_configuration"
)

func TestUnitTenantFederationTypedDispatch(t *testing.T) {
	t.Parallel()

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
			run(t, testutil.Case{Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}, Name: f.Name + "-" + string(mode), Mode: mode, Steps: []testutil.Exchange{discovery(t), wire(f.Method, strings.TrimPrefix(f.URL, origin), command, body, f.Status, f.Response)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				s := sdk.TeamsAdminAPI.TenantFederationConfigurations
				switch f.Name {
				case "tenant-federation-list":
					v, r, e := s.List(ctx)
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "tenant-federation-read":
					v, r, e := s.GetByIdentity(ctx, "Global")
					return testutil.Observation{Value: v, Response: r, Err: e}
				case "tenant-federation-update":
					no := false
					r, e := s.UpdateGlobal(ctx, configuration.UpdateRequest{Settings: configuration.Settings{AllowFederatedUsers: &no}, AllowedDomains: configuration.NewAllowList("fixture.example.invalid")})
					return testutil.Observation{Response: r, Err: e}
				default:
					return testutil.Observation{Err: errors.New("unknown pinned request")}
				}
			}})
		}
	}
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

func TestUnitTenantUpdateMatchesTypedNormalization(t *testing.T) {
	t.Parallel()

	rejected := 0
	for _, f := range pinned(t).Models {
		if f.Rejected {
			rejected++
			continue
		}
		input := inputFor(t, f.Name)
		run(t, testutil.Case{Name: f.Name, Mode: auth.Application, Steps: []testutil.Exchange{discovery(t), wire("PUT", collection+"/configuration/Global", setCommand, f.Wire, 400, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			r, e := sdk.TeamsAdminAPI.TenantFederationConfigurations.UpdateGlobal(ctx, input)
			return testutil.Observation{Response: r, Err: e}
		}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.
			TeamsAdmin, StatusCode: 400}},
		})
	}
	if rejected != 3 {
		t.Fatal("pinned rejection count differs")
	}
}

func TestUnitTenantFiltersUseRealResponses(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "list.json")
	var rows []json.RawMessage
	if json.Unmarshal(raw, &rows) != nil {
		t.Fatal("invalid real list fixture")
	}
	identities := func(row json.RawMessage) string {
		var value struct {
			Identity string `json:"Identity"`
		}
		if json.Unmarshal(row, &value) != nil {
			t.Fatal("invalid real list identity")
		}
		return value.Identity
	}
	slices.SortStableFunc(rows, func(a, b json.RawMessage) int {
		left, right := strings.EqualFold(identities(a), "Global"), strings.EqualFold(identities(b), "Global")
		if left == right {
			return 0
		}
		if left {
			return -1
		}
		return 1
	})
	cases := []struct {
		Name, Pattern, OnlyID string
		Empty                 bool
	}{
		{Name: "list"}, {Name: "all", Pattern: "*"}, {Name: "global", Pattern: "Global", OnlyID: "Global"},
		{Name: "missing", Pattern: "__sdk_missing_fixture__", Empty: true},
		{Name: "whitespace", Pattern: " ", Empty: false},
	}
	for _, row := range rows {
		id := identities(row)
		if strings.HasPrefix(id, "Tag:") && !strings.ContainsAny(id, "*?[]`") {
			cases = append(cases, struct {
				Name, Pattern, OnlyID string
				Empty                 bool
			}{Name: "tag-prefix", Pattern: strings.TrimPrefix(id, "Tag:"), OnlyID: id})
			break
		}
	}
	for _, tc := range cases {
		want := make([]json.RawMessage, 0, len(rows))
		for _, row := range rows {
			if !tc.Empty && (tc.OnlyID == "" || strings.EqualFold(identities(row), tc.OnlyID)) {
				want = append(want, row)
			}
		}
		expected, err := json.Marshal(want)
		if err != nil {
			t.Fatal(err)
		}
		result, err := testutil.Run(t.Context(), testutil.Case{
			Name:  "real-filter-" + tc.Name,
			Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/TenantFederationSettings", 200, raw), Want: testutil.Expected{Status: 200, JSON: expected, Body: raw},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				service := sdk.TeamsAdminAPI.TenantFederationConfigurations
				if tc.Pattern == "" {
					value, response, err := service.List(ctx)
					return testutil.Observation{Value: value, Response: response, Err: err}
				}
				value, response, err := service.ListFiltered(ctx, tc.Pattern)
				return testutil.Observation{Value: value, Response: response, Err: err}
			},
		})
		if err != nil || !result.Passed() {
			t.Fatalf("real filtering response differs: %v %+v", err, result.Differences())
		}
	}

}

func TestUnitTenantIdentityEscapingAndMutationErrors(t *testing.T) {
	t.Parallel()

	var snapshot configuration.Configuration
	if json.Unmarshal([]byte(`{"Identity":"scope/Fixture / Config","AllowedDomains":["Future"]}`), &snapshot) != nil || snapshot.Identity != "scope/Fixture / Config" || string(snapshot.AllowedDomains) != `["Future"]` {
		t.Fatal("local compatibility model input lost raw fields")
	}
	identity := "scope/Fixture / Config"
	run(t, testutil.Case{Name: "escaped", Mode: auth.Application, Steps: []testutil.Exchange{discovery(t), wire("GET", collection+"/configuration/"+url.PathEscape(identity), getCommand, nil, 400, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.TenantFederationConfigurations.GetByIdentity(ctx, identity)
		if e == nil && (v.Identity != identity || string(v.AllowedDomains) != `["Future"]`) {
			e = errors.New("escaped read changed")
		}
		return testutil.Observation{Response: r, Err: e}
	}, Want: testutil.Expected{APIError: &client.APIError{Backend: client.
		TeamsAdmin, StatusCode: 400}},
	})
	for _, status := range []int{400, 403, 429, 503} {
		run(t, testutil.Case{Name: fmt.Sprint("denied-", status), Mode: auth.Application, Steps: []testutil.Exchange{discovery(t), wire("PUT", collection+"/configuration/Global", setCommand, json.RawMessage(`{"Identity":"Global"}`), status, `{"Code":"FutureDenied","Message":"fixture private text"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			r, e := sdk.TeamsAdminAPI.TenantFederationConfigurations.UpdateGlobal(ctx, configuration.UpdateRequest{})
			var api *client.APIError
			if !errors.As(e, &api) || api.StatusCode != status || r == nil || !strings.Contains(string(r.Body()), "fixture private text") || strings.Contains(fmt.Sprint(e), "fixture private text") {
				return testutil.Observation{Err: errors.New("error classification/metadata/redaction differs")}
			}
			return testutil.Observation{Response: r}
		}})
	}
}

func TestUnitTenantInvalidInputsDoNotAcquireCredentials(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "invalid-local", Mode: auth.Application, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.TenantFederationConfigurations
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, configuration.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid read identity accepted")}
			}
			if _, e := s.UpdateByIdentity(ctx, id, configuration.UpdateRequest{}); !errors.Is(e, configuration.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid update identity accepted")}
			}
		}
		for _, input := range []configuration.UpdateRequest{{AllowedDomains: configuration.NewAllowList(), AllowedDomainsAsAList: configuration.ReplaceDomains(nil)}, {AllowedDomains: &configuration.AllowedDomains{}}, {BlockedDomains: &configuration.ListChange{}}} {
			if _, e := s.UpdateGlobal(ctx, input); !errors.Is(e, configuration.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid domain input accepted")}
			}
		}
		if _, _, e := s.ListFiltered(ctx, "[broken"); !errors.Is(e, configuration.ErrInput) {
			return testutil.Observation{Err: errors.New("invalid filter accepted")}
		}
		c, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.List(c); !errors.Is(e, context.Canceled) {
			return testutil.Observation{Err: errors.New("cancellation changed")}
		}
		return testutil.Observation{}
	}})
}
