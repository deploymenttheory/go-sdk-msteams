package teamsmedialoggingpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_media_logging_policy"
)

const collection = "/Skype.Policy/configurations/TeamsMediaLoggingPolicy"

func TestUnitActualModernReadAndCollection(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ file, path, identity string }{
		{"get_global.json", collection + "/configuration/Global", "Global"},
		{"get_tag_enabled.json", collection + "/configuration/Tag:Enabled", "Tag:Enabled"},
		{"list.json", collection, ""},
	} {
		raw := readResponseFixture(t, tc.file)
		run(t, contracttest.Case{Name: strings.TrimSuffix(tc.file, ".json"), Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsMediaLoggingPolicy", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{compiledWire("GET", tc.path, "Get-CsTeamsMediaLoggingPolicy", nil, 200, string(raw))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsMediaLoggingPolicies
			if tc.identity != "" {
				v, r, e := s.GetByIdentity(ctx, tc.identity)
				if e == nil && (v.Identity != tc.identity || v.AllowMediaLogging == nil || *v.AllowMediaLogging != (tc.identity != "Global")) {
					t.Fatal("actual getter projection lost")
				}
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}
			values, r, e := s.List(ctx)
			var expected []map[string]json.RawMessage
			if json.Unmarshal(raw, &expected) != nil || len(values) != len(expected) || len(values) < 2 || values[0].Identity != "Global" {
				t.Fatal("actual collection or Global ordering lost")
			}
			return contracttest.Observation{Value: values, Response: r, Err: e}
		}})
	}
}

func TestUnitModernFiltersUseActualCollection(t *testing.T) {
	t.Parallel()
	var values []resource.Policy
	if json.Unmarshal(readResponseFixture(t, "list.json"), &values) != nil {
		t.Fatal("invalid actual list")
	}
	for _, tc := range []struct {
		pattern string
		count   int
	}{{"Global", 1}, {"Glo?al", 1}, {"Enabled", 1}, {"Tag:Enabled", 1}, {"*", len(values)}, {"missing", 0}, {" ", len(values)}, {"", len(values)}} {
		run(t, contracttest.Case{Name: "filter", Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsMediaLoggingPolicy", ParameterSet: "Filter"}}, Steps: []contracttest.Exchange{compiledWire("GET", collection, "Get-CsTeamsMediaLoggingPolicy", nil, 200, string(readResponseFixture(t, "list.json")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			values, r, e := sdk.TeamsAdminAPI.TeamsMediaLoggingPolicies.ListFiltered(ctx, tc.pattern)
			if e == nil && len(values) != tc.count {
				t.Fatalf("modern filter %q differs", tc.pattern)
			}
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
}

func TestUnitReadFailuresRetainResponses(t *testing.T) {
	t.Parallel()
	missing := readResponseFixture(t, "get_missing_404.json")
	var envelope struct {
		Code      string `json:"code"`
		ErrorCode int    `json:"errorCode"`
	}
	if json.Unmarshal(missing, &envelope) != nil || envelope.Code == "" {
		t.Fatal("invalid actual missing response")
	}
	for _, tc := range []struct {
		name   string
		status int
		body   string
		want   contracttest.Expected
	}{
		{"missing", 404, string(missing), contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 404, Codes: []string{envelope.Code}, ServiceErrorCodes: []int{envelope.ErrorCode}}}},
		{"denied", 403, "", contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 403}}},
		{"malformed", 200, `[`, contracttest.Expected{Error: client.ErrDecode}},
		{"wrong-shape", 200, `"text"`, contracttest.Expected{Error: client.ErrDecode}},
	} {
		run(t, contracttest.Case{Name: tc.name, Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsMediaLoggingPolicy", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{compiledWire("GET", collection+"/configuration/Global", "Get-CsTeamsMediaLoggingPolicy", nil, tc.status, tc.body)}, Want: tc.want, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsMediaLoggingPolicies.GetByIdentity(ctx, "Global")
			if e == nil || v != nil {
				t.Fatal("read failure returned a value")
			}
			return contracttest.Observation{Response: r, Err: e}
		}})
	}
}

func TestUnitGrantScopeIdentityAndCancellationRejectBeforeHTTP(t *testing.T) {
	t.Parallel()
	runExplicit(t, contracttest.Case{Name: "local-validation", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.TeamsMediaLoggingPolicies
		for _, input := range []resource.Assignment{{}, {User: &resource.UserAssignment{Identity: "fixture"}, Tenant: &resource.TenantAssignment{}}, {Tenant: &resource.TenantAssignment{}, Group: &resource.GroupGrant{ID: "group"}}} {
			if _, _, e := s.Assign(ctx, input); !errors.Is(e, resource.ErrInput) {
				t.Fatal("ambiguous scope reached HTTP")
			}
		}
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid identity reached HTTP")
			}
			if _, _, e := s.Assign(ctx, resource.Assignment{User: &resource.UserAssignment{Identity: id, PolicyName: "Enabled"}}); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid user reached HTTP")
			}
			if _, _, e := s.Assign(ctx, resource.Assignment{Group: &resource.GroupGrant{ID: id, PolicyName: "Enabled"}}); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid group reached HTTP")
			}
		}
		if _, _, e := s.ListFiltered(ctx, "[broken"); !errors.Is(e, resource.ErrInput) {
			t.Fatal("malformed filter reached HTTP")
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.Assign(canceled, resource.Assignment{Tenant: &resource.TenantAssignment{}}); !errors.Is(e, context.Canceled) {
			t.Fatal("cancellation lost")
		}
		if _, e := s.AssignTenantDefault(canceled, "", false); !errors.Is(e, context.Canceled) {
			t.Fatal("tenant cancellation lost")
		}
		if _, _, e := s.GetByIdentity(canceled, "Global"); !errors.Is(e, context.Canceled) {
			t.Fatal("read cancellation lost")
		}
		for _, value := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := value.List(ctx); !errors.Is(e, client.ErrClosed) {
				t.Fatal("nil backend reached HTTP")
			}
		}
		return contracttest.Observation{}
	}})
}
