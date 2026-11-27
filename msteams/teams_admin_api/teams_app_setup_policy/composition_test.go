package teamsappsetuppolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_app_setup_policy"
)

func TestUnitAppSetupUpdatePreparationFailures(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	var expectedBefore map[string]json.RawMessage
	if json.Unmarshal(before, &expectedBefore) != nil {
		t.Fatal("invalid real preparation fixture")
	}
	cases := []struct {
		name        string
		read, patch int
		body        string
	}{
		{"read-denied", 403, 0, ""},
		{"read-empty", 204, 0, ""},
		{"read-malformed", 200, 0, "not-json"},
	}
	cases = append(cases,
		struct {
			name        string
			read, patch int
			body        string
		}{"patch-denied", 200, 403, string(before)},
		struct {
			name        string
			read, patch int
			body        string
		}{"patch-throttled", 200, 429, string(before)},
		struct {
			name        string
			read, patch int
			body        string
		}{"patch-unavailable", 200, 503, string(before)},
	)
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			steps := []testutil.Exchange{wire("GET", item, "Set-CsTeamsAppSetupPolicy", nil, tc.read, tc.body)}
			if tc.patch != 0 {
				steps = append(steps, wire("PATCH", item, "Set-CsTeamsAppSetupPolicy", json.RawMessage(`{"Identity":"Global","AllowSideLoading":false}`), tc.patch, ""))
			}
			run(t, testutil.Case{Name: tc.name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsAppSetupPolicies.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{AllowSideLoading: policy.Set(false)}})
				if e == nil || v == nil || v.ReadResponse == nil || v.PatchInvoked != (tc.patch != 0) || v.HTTPAccepted {
					t.Error("failed update lost preparation state or invented acceptance")
				}
				if tc.patch != 0 && (v.Before == nil || !reflect.DeepEqual(v.Before.RawFields, expectedBefore)) {
					t.Error("failed update changed or lost the actual preparation snapshot")
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

func TestUnitAppSetupListReplacementsUseRealPreparationOnDeniedPatch(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	var expected map[string]json.RawMessage
	if json.Unmarshal(before, &expected) != nil {
		t.Fatal("invalid real setup preparation")
	}
	for _, field := range []string{"AppPresetList", "PinnedAppBarApps", "PinnedMessageBarApps", "AppPresetMeetingList", "AdditionalCustomizationApps", "PinnedCallingBarApps"} {
		t.Run(field, func(t *testing.T) {
			rows := `[{"Identity":null,"Id":"unit-app"}]`
			if field == "PinnedAppBarApps" || field == "PinnedMessageBarApps" {
				rows = `[{"Identity":null,"Id":"unit-app","Order":9007199254740993}]`
			}
			if field == "AdditionalCustomizationApps" {
				rows = `[{"Identity":null,"Id":"unit-app","AdditionalCustomizationId":"unit-custom"}]`
			}
			body := json.RawMessage(`{"Identity":"Global","` + field + `":` + rows + `}`)
			run(t, testutil.Case{Name: "list-replacement", Steps: []testutil.Exchange{wire("GET", item, "Set-CsTeamsAppSetupPolicy", nil, 200, string(before)), wire("PATCH", item, "Set-CsTeamsAppSetupPolicy", body, 403, "")}, Want: testutil.Expected{Error: auth.ErrForbidden}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				input := resource.UpdateRequest{}
				switch field {
				case "AppPresetList":
					input.AppPresetList = resource.ReplaceApps(resource.NewApp("unit-app"))
				case "AppPresetMeetingList":
					input.AppPresetMeetingList = resource.ReplaceApps(resource.NewApp("unit-app"))
				case "PinnedCallingBarApps":
					input.PinnedCallingBarApps = resource.ReplaceApps(resource.NewApp("unit-app"))
				case "PinnedAppBarApps":
					input.PinnedAppBarApps = resource.ReplacePins(resource.NewPinnedApp("unit-app", largeOrder))
				case "PinnedMessageBarApps":
					input.PinnedMessageBarApps = resource.ReplacePins(resource.NewPinnedApp("unit-app", largeOrder))
				case "AdditionalCustomizationApps":
					input.AdditionalCustomizationApps = resource.ReplaceCustomizations(resource.NewCustomizationApp("unit-app", "unit-custom"))
				}
				v, r, e := sdk.TeamsAdminAPI.TeamsAppSetupPolicies.UpdateByIdentity(ctx, "Global", input)
				if v == nil || v.Before == nil || !v.PatchInvoked || v.HTTPAccepted || !reflect.DeepEqual(v.Before.RawFields, expected) {
					t.Error("denied setup update lost or modified actual preparation")
				}
				return testutil.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitAppSetupBuildersNeverAcquireCredentialsOrHTTP(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "local-setup-builders", Invoke: func(_ context.Context, _ *msteams.Client) testutil.Observation {
		apps, e := resource.ReplaceApps(resource.NewApp("")).Apply(nil)
		if e != nil || len(apps) != 1 || *apps[0].ID != "" {
			return testutil.Observation{Err: errors.New("local app builder changed")}
		}
		pins, e := resource.ReplacePins(resource.NewPinnedApp("future-app", largeOrder)).Apply(nil)
		if e != nil || len(pins) != 1 || pins[0].Order != largeOrder {
			return testutil.Observation{Err: errors.New("local pin builder changed")}
		}
		customs, e := resource.ReplaceCustomizations(resource.NewCustomizationApp("future-app", "")).Apply(nil)
		if e != nil || len(customs) != 1 || *customs[0].AdditionalCustomizationID != "" {
			return testutil.Observation{Err: errors.New("local customization builder changed")}
		}
		return testutil.Observation{}
	}})
}
