package meetingbranding_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/meeting_branding"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

func TestUnitBrandingComposedUpdatePreservesPreparationAndBoundFields(t *testing.T) {
	t.Parallel()
	const before = `{"Identity":"read-model-identity","DefaultTheme":"preserved","EnableMeetingBackgroundImages":true,"MeetingBackgroundImages":[{"Id":"a","Order":0,"Name":"Fixture","ImageUri":"https://example.invalid/a"}],"Future":false}`
	const image = `{"Identity":null,"Id":"a","Order":0,"Name":"Fixture","ImageUri":"https://example.invalid/a","IsRequired":null,"IsHidden":null}`
	var results []contracttest.Result
	for _, mode := range []string{"append", "empty-edit", "remove"} {
		steps := []contracttest.Exchange{wire("GET", item, "Set-CsTeamsMeetingBrandingPolicy", nil, 200, before)}
		if mode != "remove" {
			body := `{"Identity":"read-model-identity","DefaultTheme":null,"MeetingBackgroundImages":[]}`
			if mode == "append" {
				body = `{"Identity":"read-model-identity","DefaultTheme":null,"MeetingBackgroundImages":[` + image + `,` + image + `]}`
			}
			steps = append(steps, wire("PATCH", item, "Set-CsTeamsMeetingBrandingPolicy", json.RawMessage(body), 204, ""))
		}
		results = append(results, run(t, contracttest.Case{Name: "composed-" + mode, Mappings: []contracttest.Mapping{{Command: "Set-CsTeamsMeetingBrandingPolicy", ParameterSet: "Identity"}}, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			image, e := resource.NewBackgroundImage("a", 0, "Fixture", "https://example.invalid/a", resource.BackgroundImageOptions{})
			if e != nil {
				return contracttest.Observation{Err: e}
			}
			var add, remove []*resource.BackgroundImage
			if mode == "append" {
				add = []*resource.BackgroundImage{image}
			}
			if mode == "remove" {
				remove = []*resource.BackgroundImage{image}
			}
			edit, e := resource.ChangeAssets(add, remove)
			if e != nil {
				return contracttest.Observation{Err: e}
			}
			value, response, e := sdk.TeamsAdminAPI.MeetingBrandingPolicies.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Fields: resource.Fields{DefaultTheme: policy.Null[string]()}, MeetingBackgroundImages: edit})
			if value == nil || value.Before == nil || value.ReadResponse == nil || len(value.Before.MeetingBackgroundImages) != 1 || value.PatchInvoked != (mode != "remove") || value.HTTPAccepted != (mode != "remove") {
				return contracttest.Observation{Err: errors.New("branding update preparation/acceptance lost")}
			}
			if mode == "remove" {
				if !errors.Is(e, resource.ErrCannotRemoveElements) {
					return contracttest.Observation{Err: errors.New("throwing comparator lost")}
				}
				return contracttest.Observation{Response: response}
			}
			return contracttest.Observation{Response: response, Err: e}
		}}))
	}
	evidence(t, results)
}

func TestUnitBrandingUpdateFailuresDoNotReplayAndRetainHTTPAcceptance(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, tc := range []struct {
		name              string
		read, patch       int
		body              string
		accepted, invoked bool
	}{
		{"read-denied", 403, 0, "", false, false},
		{"patch-throttled", 200, 429, `{"Message":"fixture-private"}`, false, true},
		{"patch-unavailable", 200, 503, `{"Message":"fixture-private"}`, false, true},
		{"accepted-malformed", 200, 200, "not-json", true, true},
	} {
		steps := []contracttest.Exchange{wire("GET", item, "Set-CsTeamsMeetingBrandingPolicy", nil, tc.read, `{"Identity":"read-model-identity"}`)}
		if tc.invoked {
			steps = append(steps, wire("PATCH", item, "Set-CsTeamsMeetingBrandingPolicy", json.RawMessage(`{"Identity":"read-model-identity","DefaultTheme":""}`), tc.patch, tc.body))
		}
		results = append(results, run(t, contracttest.Case{Name: tc.name, Mappings: []contracttest.Mapping{{Command: "Set-CsTeamsMeetingBrandingPolicy", ParameterSet: "Identity"}}, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.MeetingBrandingPolicies.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Fields: resource.Fields{DefaultTheme: policy.Set("")}})
			if e == nil || v == nil || v.PatchInvoked != tc.invoked || v.HTTPAccepted != tc.accepted || v.ReadResponse == nil {
				return contracttest.Observation{Err: errors.New("partial branding failure lost")}
			}
			if tc.accepted && !errors.Is(e, client.ErrDecode) {
				return contracttest.Observation{Err: errors.New("decode failure lost after acceptance")}
			}
			return contracttest.Observation{Response: r}
		}}))
	}
	evidence(t, results)
}

func TestUnitBrandingCreateComparisonIsExplicitAndDoesNotCloneDefaults(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, status := range []int{200, 403} {
		results = append(results, run(t, contracttest.Case{Name: fmt.Sprint("comparison-", status), Mappings: []contracttest.Mapping{{Command: "New-CsTeamsMeetingBrandingPolicy", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{wire("GET", item, "New-CsTeamsMeetingBrandingPolicy", nil, status, `{"Identity":"Global","DefaultTheme":"unbound-default"}`), wire("POST", collection, "New-CsTeamsMeetingBrandingPolicy", json.RawMessage(`{"Identity":"Fixture","EnableMeetingBackgroundImages":false}`), 201, `{"Identity":"Fixture"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.MeetingBrandingPolicies.CreateWithGlobalComparison(ctx, resource.CreateRequest{Identity: "Fixture", Settings: resource.Settings{Fields: resource.Fields{EnableMeetingBackgroundImages: policy.Set(false)}}})
			if e == nil && (v == nil || !v.CreateInvoked || !v.HTTPAccepted || v.Created == nil || (v.ComparisonError != nil) != (status == 403)) {
				e = errors.New("comparison and submission observations mixed")
			}
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	evidence(t, results)
}

func TestUnitBrandingListFilteringPreservesExactGlobalOrdering(t *testing.T) {
	t.Parallel()
	runResult := run(t, contracttest.Case{Name: "filter-local", Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsMeetingBrandingPolicy", ParameterSet: "Filter"}}, Steps: []contracttest.Exchange{wire("GET", collection, "Get-CsTeamsMeetingBrandingPolicy", nil, 200, `[{"Identity":"global"},{"Identity":"Tag:Fixture"},{"Identity":"Global"}]`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.MeetingBrandingPolicies.ListFiltered(ctx, "Fixture*")
		if e == nil && (len(v) != 1 || v[0].Identity == nil || *v[0].Identity != "Tag:Fixture") {
			e = errors.New("local branding wildcard selection changed")
		}
		return contracttest.Observation{Response: r, Err: e}
	}})

	order := run(t, contracttest.Case{Name: "global-order", Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsMeetingBrandingPolicy", ParameterSet: "Identity"}}, Want: contracttest.Expected{JSON: json.RawMessage(`[{"Identity":"Global"},{"Identity":"global"},{"Identity":"Tag:Fixture"}]`)}, Steps: []contracttest.Exchange{wire("GET", collection, "Get-CsTeamsMeetingBrandingPolicy", nil, 200, `[{"Identity":"global"},{"Identity":"Tag:Fixture"},{"Identity":"Global"}]`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, e := sdk.TeamsAdminAPI.MeetingBrandingPolicies.List(ctx)
		return contracttest.Observation{Value: v, Response: r, Err: e}
	}})
	evidence(t, []contracttest.Result{runResult, order})
}
