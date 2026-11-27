package privacyconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/privacy_configuration"
)

func TestUnitPrivacyUpdatePreparationFailures(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		read, patch int
		body        string
	}{
		{"read-denied", 403, 0, ""},
		{"read-empty", 204, 0, ""},
		{"read-malformed", 200, 0, "not-json"},
	}
	cases = append(cases, struct {
		name        string
		read, patch int
		body        string
	}{"actual-read-denial", 403, 0, string(readResponseFixture(t, "get_global_403.json"))})
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			steps := []testutil.Exchange{wire("GET", item, "Set-CsPrivacyConfiguration", nil, tc.read, tc.body)}
			if tc.patch != 0 {
				steps = append(steps, wire("PATCH", item, "Set-CsPrivacyConfiguration", json.RawMessage(`{"Identity":"Global","EnablePrivacyMode":false}`), tc.patch, ""))
			}
			run(t, testutil.Case{Name: tc.name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.PrivacyConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{EnablePrivacyMode: policy.Set(false)}})
				if e == nil || v == nil || v.ReadResponse == nil || v.PatchInvoked != (tc.patch != 0) || v.HTTPAccepted {
					t.Error("failed update lost preparation state or invented acceptance")
				}
				if tc.read == 403 || tc.patch != 0 {
					var apiError *client.APIError
					expected := tc.read
					if tc.patch != 0 {
						expected = tc.patch
					}
					if !errors.As(e, &apiError) || apiError.StatusCode != expected {
						t.Error("failure lost its HTTP status")
					}
				}
				return testutil.Observation{Response: r}
			}})
		})
	}
}

func TestUnitPrivacyRejectedFilterRequests(t *testing.T) {
	t.Parallel()
	// This retained fixture does not supply a successful list. These controlled
	// errors prove request construction only, not current live capability.
	for _, pattern := range []string{"*", "Global", "Fixture*"} {
		result, err := testutil.Run(t.Context(), testutil.Case{
			Name:  "request-only-filter",
			Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/PrivacyConfiguration", 400, nil),
			Want:  testutil.Expected{Status: 400, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				value, response, err := sdk.TeamsAdminAPI.PrivacyConfigurations.ListFiltered(ctx, pattern)
				return testutil.Observation{Value: value, Response: response, Err: err}
			},
		})
		if err != nil || !result.Passed() {
			t.Fatalf("request-only filter differs: %v %+v", err, result.Differences())
		}
	}

}

// Primary-authored Core PrivacyConfiguration DTOs, not a retained tenant list.
// Only Global is authored; no unsupported nondefault scope is inferred.
func TestUnitPrivacyAuthoredListAndLiteralFilterResponses(t *testing.T) {
	t.Parallel()
	const body = `[{"Identity":"Global","EnablePrivacyMode":false,"AutoInitiateContacts":null,"Future":false}]`
	for _, tc := range []struct {
		name, pattern string
		list          bool
		count         int
	}{
		{name: "omitted-list", list: true, count: 1},
		{name: "all", pattern: "*", count: 1},
		{name: "case-insensitive-global", pattern: "global", count: 1},
		{name: "missing", pattern: "__sdk_missing__"},
		{name: "literal-whitespace", pattern: " "},
	} {

		run(t, testutil.Case{Name: "authored-" + tc.name, Steps: []testutil.Exchange{wire("GET", collection, "Get-CsPrivacyConfiguration", nil, 200, body)}, Want: testutil.Expected{Body: []byte(body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			var values []*resource.Configuration
			var observed testutil.Observation
			if tc.list {
				values, observed.Response, observed.Err = sdk.TeamsAdminAPI.PrivacyConfigurations.List(ctx)
			} else {
				values, observed.Response, observed.Err = sdk.TeamsAdminAPI.PrivacyConfigurations.ListFiltered(ctx, tc.pattern)
			}
			if observed.Err == nil {
				if len(values) != tc.count {
					t.Fatal("successful list/filter did not execute its own selection")
				}
				if len(values) == 1 && (values[0] == nil || values[0].Identity == nil || *values[0].Identity != "Global" || values[0].EnablePrivacyMode == nil || *values[0].EnablePrivacyMode || string(values[0].RawFields["AutoInitiateContacts"]) != "null" || string(values[0].RawFields["Future"]) != "false") {
					t.Fatal("typed privacy list lost identity, false/null or future data")
				}
			}
			observed.Value = values
			return observed
		}})
	}
}

func TestUnitPrivacySparseInputStopsOnPreparationDenial(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "preparation-denial", Steps: []testutil.Exchange{wire("GET", item, "Set-CsPrivacyConfiguration", nil, 403, "")}, Want: testutil.Expected{Error: auth.ErrForbidden}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.PrivacyConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{AutoInitiateContacts: policy.Set(false)}})
		if v == nil || v.ReadResponse == nil || v.Before != nil || v.PatchInvoked || v.HTTPAccepted {
			t.Error("failed preparation invented a snapshot, patch or acceptance")
		}
		return testutil.Observation{Response: r, Err: e}
	}})
}
