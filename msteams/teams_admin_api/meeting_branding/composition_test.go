package meetingbranding_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/meeting_branding"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestUnitBrandingAssetInputsUseRealPreparationOnDeniedPatch(t *testing.T) {
	t.Parallel()
	runBrandingPreparedAssetPatch(t, 403)
}

func TestUnitBrandingAssetInputsUsePreparedSuccessfulPatch(t *testing.T) {
	t.Parallel()
	// Authored blank204 acceptance follows the pinned generic Set policy
	// parser's empty-success branch. The GET bytes are retained evidence;
	// no nonempty asset response or server persistence is invented.
	runBrandingPreparedAssetPatch(t, 204)
}

func runBrandingPreparedAssetPatch(t *testing.T, status int) {
	t.Helper()
	before := readResponseFixture(t, "get_global.json")
	var expected map[string]json.RawMessage
	if json.Unmarshal(before, &expected) != nil {
		t.Fatal("invalid real branding preparation")
	}
	for _, appendImage := range []bool{false, true} {
		label, images := "empty-edit", `[]`
		if appendImage {
			label = "append-image"
			images = `[{"Identity":null,"Id":"a","Order":0,"Name":"Fixture","ImageUri":"https://example.invalid/a","IsRequired":null,"IsHidden":null}]`
		}
		t.Run(label, func(t *testing.T) {
			body := json.RawMessage(`{"Identity":"Global","DefaultTheme":null,"MeetingBackgroundImages":` + images + `}`)
			want := contracttest.Expected{}
			if status == 403 {
				want.Error = auth.ErrForbidden
			}
			run(t, contracttest.Case{Name: label, Source: "msteams/teams_admin_api/meeting_branding/composition_test.go", Steps: []contracttest.Exchange{wire("GET", item, "Set-CsTeamsMeetingBrandingPolicy", nil, 200, string(before)), wire("PATCH", item, "Set-CsTeamsMeetingBrandingPolicy", body, status, "")}, Want: want, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				var add []*resource.BackgroundImage
				if appendImage {
					image, err := resource.NewBackgroundImage("a", 0, "Fixture", "https://example.invalid/a", resource.BackgroundImageOptions{})
					if err != nil {
						t.Fatal(err)
					}
					add = []*resource.BackgroundImage{image}
				}
				edit, err := resource.ChangeAssets(add, []*resource.BackgroundImage(nil))
				if err != nil {
					t.Fatal(err)
				}
				v, r, e := sdk.TeamsAdminAPI.MeetingBrandingPolicies.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Fields: resource.Fields{DefaultTheme: policy.Null[string]()}, MeetingBackgroundImages: edit})
				if v == nil || v.Before == nil || !v.PatchInvoked || v.HTTPAccepted != (status == 204) || v.After != nil || !reflect.DeepEqual(v.Before.RawFields, expected) {
					t.Error("asset update lost acceptance or modified preparation")
				}
				return contracttest.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitBrandingUpdatePreparationFailures(t *testing.T) {
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
			steps := []contracttest.Exchange{wire("GET", item, "Set-CsTeamsMeetingBrandingPolicy", nil, tc.read, tc.body)}
			if tc.patch != 0 {
				steps = append(steps, wire("PATCH", item, "Set-CsTeamsMeetingBrandingPolicy", json.RawMessage(`{"Identity":"Global","DefaultTheme":""}`), tc.patch, ""))
			}
			run(t, contracttest.Case{Name: tc.name, Source: "msteams/teams_admin_api/meeting_branding/composition_test.go", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.TeamsAdminAPI.MeetingBrandingPolicies.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Fields: resource.Fields{DefaultTheme: policy.Set("")}})
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
				return contracttest.Observation{Response: r}
			}})
		})
	}
}

func TestUnitBrandingCreateComparisonIsExplicitAndDoesNotCloneDefaultsOnDeniedSubmission(t *testing.T) {
	t.Parallel()
	for _, readStatus := range []int{200, 403} {
		t.Run(strconv.Itoa(readStatus), func(t *testing.T) {
			readBody := string(readResponseFixture(t, "get_global.json"))
			if readStatus != 200 {
				readBody = ""
			}
			run(t, contracttest.Case{Name: "comparison-denial", Source: "msteams/teams_admin_api/meeting_branding/composition_test.go", Steps: []contracttest.Exchange{wire("GET", item, "New-CsTeamsMeetingBrandingPolicy", nil, readStatus, readBody), wire("POST", collection, "New-CsTeamsMeetingBrandingPolicy", json.RawMessage(`{"Identity":"Fixture","EnableMeetingBackgroundImages":false}`), 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, e := sdk.TeamsAdminAPI.MeetingBrandingPolicies.CreateWithGlobalComparison(ctx, resource.CreateRequest{Identity: "Fixture", Settings: resource.Settings{Fields: resource.Fields{EnableMeetingBackgroundImages: policy.Set(false)}}})
				var apiError *client.APIError
				if !errors.As(e, &apiError) || apiError.StatusCode != 403 || v == nil || !v.CreateInvoked || v.HTTPAccepted || v.Created != nil || (v.ComparisonError != nil) != (readStatus != 200) {
					t.Error("comparison and denied submission outcomes were conflated")
				}
				return contracttest.Observation{Response: r}
			}})
		})
	}
}

func TestUnitBrandingCreateComparisonPreservesAcceptedSubmission(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	accepted := readResponseFixture(t, "create_201.body")
	var expected map[string]json.RawMessage
	if json.Unmarshal(before, &expected) != nil {
		t.Fatal("invalid retained Global comparison")
	}
	// Offline composition of retained primitive replies. A denied comparison
	// remains a separate observation even when the later create is accepted.
	for _, status := range []int{200, 403} {
		t.Run(strconv.Itoa(status), func(t *testing.T) {
			read := before
			if status == 403 {
				read = nil
			}
			run(t, contracttest.Case{Name: "accepted-comparison-" + strconv.Itoa(status), Source: "msteams/teams_admin_api/meeting_branding/composition_test.go", Steps: []contracttest.Exchange{wire("GET", item, "New-CsTeamsMeetingBrandingPolicy", nil, status, string(read)), wire("POST", collection, "New-CsTeamsMeetingBrandingPolicy", json.RawMessage(`{"Identity":"Fixture","EnableMeetingBackgroundImages":false}`), 201, string(accepted))}, Want: contracttest.Expected{Body: accepted}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.MeetingBrandingPolicies.CreateWithGlobalComparison(ctx, resource.CreateRequest{Identity: "Fixture", Settings: resource.Settings{Fields: resource.Fields{EnableMeetingBackgroundImages: policy.Set(false)}}})
				if value == nil || !value.CreateInvoked || !value.HTTPAccepted || value.Created != nil || value.ComparisonResponse == nil || value.ComparisonResponse.StatusCode() != status {
					t.Fatal("accepted create lost independent comparison and receipt stages")
				}
				if status == 200 && (value.ComparisonError != nil || value.Comparison == nil || value.Comparison.Identity == nil || *value.Comparison.Identity != "Global" || !reflect.DeepEqual(value.Comparison.RawFields, expected)) {
					t.Fatal("accepted create lost the typed Global comparison")
				}
				if status == 403 && (value.Comparison != nil || !errors.Is(value.ComparisonError, auth.ErrForbidden)) {
					t.Fatal("successful create erased the failed comparison")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
		})
	}
}

func TestUnitBrandingListFilteringPreservesExactGlobalOrderingFromRealResponses(t *testing.T) {
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
		{Name: "whitespace", Pattern: " ", Empty: true},
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
		result, err := contracttest.Run(t.Context(), contracttest.Case{
			Name: "real-filter-" + tc.Name, Family: "meeting_branding", Scope: "service", Source: "msteams/teams_admin_api/meeting_branding/testdata/list.json",
			Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/TeamsMeetingBrandingPolicy", 200, raw), Want: contracttest.Expected{Status: 200, JSON: expected, Body: raw},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				service := sdk.TeamsAdminAPI.MeetingBrandingPolicies
				if tc.Pattern == "" {
					value, response, err := service.List(ctx)
					return contracttest.Observation{Value: value, Response: response, Err: err}
				}
				value, response, err := service.ListFiltered(ctx, tc.Pattern)
				return contracttest.Observation{Value: value, Response: response, Err: err}
			},
		})
		if err != nil || !result.Passed() {
			t.Fatalf("real filtering response differs: %v %+v", err, result.Differences())
		}
	}

}
