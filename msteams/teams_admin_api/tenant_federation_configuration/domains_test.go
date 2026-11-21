package tenantfederationconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	federation "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_federation_configuration"
)

func TestUnitDomainConstructorsAreLocal(t *testing.T) {
	t.Parallel()

	artifact := struct {
		Cases []struct {
			Name string          `json:"name"`
			Wire json.RawMessage `json:"wire"`
		} `json:"cases"`
	}{
		Cases: []struct {
			Name string          `json:"name"`
			Wire json.RawMessage `json:"wire"`
		}{

			{
				Name: "domain-pattern",
				Wire: json.RawMessage("{\"Domain\":\"*.example.invalid\"}"),
			},

			{
				Name: "allow-list",
				Wire: json.RawMessage("[\"b.example.invalid\",\"a.example.invalid\",\"b.example.invalid\",\"\"]"),
			},

			{
				Name: "empty-allow-list",
				Wire: json.RawMessage("[]"),
			},

			{
				Name: "allow-all",
				Wire: json.RawMessage("[\"AllowAllKnownDomains\"]"),
			},
		},
	}

	observed := map[string]json.RawMessage{}
	for _, c := range artifact.Cases {
		observed[c.Name] = c.Wire
	}
	for _, tc := range []struct {
		name, command string
		build         func() (any, error)
	}{
		{"domain-pattern", "New-CsEdgeDomainPattern", func() (any, error) { return federation.NewDomainPattern("*.example.invalid") }},
		{"allow-list", "New-CsEdgeAllowList", func() (any, error) {
			return federation.NewAllowList("b.example.invalid", "a.example.invalid", "b.example.invalid", ""), nil
		}},
		{"empty-allow-list", "New-CsEdgeAllowList", func() (any, error) { return federation.NewAllowList(), nil }},
		{"allow-all", "New-CsEdgeAllowAllKnownDomains", func() (any, error) { return federation.NewAllowAllKnownDomains(), nil }},
	} {
		if len(observed[tc.name]) == 0 {
			t.Fatal("missing pinned case")
		}
		r, err := contracttest.Run(t.Context(), contracttest.Case{Name: tc.name, Family: "tenant_federation_configuration", Scope: "local", Source: "msteams/teams_admin_api/tenant_federation_configuration/testdata/get_global.json", Mappings: []contracttest.Mapping{{Command: tc.command, ParameterSet: "__AllParameterSets"}}, Invoke: func(context.Context, *msteams.Client) contracttest.Observation {
			v, e := tc.build()
			return contracttest.Observation{Value: v, Err: e}
		}, Want: contracttest.Expected{JSON: observed[tc.name]}})
		if err != nil || !r.Passed() {
			t.Fatalf("local builder differs: %v %+v", err, r.Differences())
		}

	}
}

func TestUnitAllowListOwnsOriginalOrderAndRejectsInvalidZeroChoice(t *testing.T) {
	t.Parallel()
	input := []string{"fixture.invalid", "fixture.invalid"}
	value := federation.NewAllowList(input...)
	input[0] = "changed.invalid"
	raw, err := json.Marshal(value)
	if err != nil || string(raw) != `["fixture.invalid","fixture.invalid"]` {
		t.Fatal("allow-list changed or borrowed inputs")
	}
	if _, err := json.Marshal(federation.AllowedDomains{}); !errors.Is(err, federation.ErrInput) {
		t.Fatal("invalid zero choice serialized")
	}
	if _, err := federation.NewDomainPattern(""); !errors.Is(err, federation.ErrInput) {
		t.Fatal("missing mandatory domain accepted")
	}
}
