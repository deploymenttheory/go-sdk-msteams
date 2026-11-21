package teamscustombannertext_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_custom_banner_text"
	"github.com/go-resty/resty/v2"
)

func TestUnitBannerComposedUpdateKeepsReadIDAndFailurePreparation(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	const readID = "00000000-0000-4000-8000-000000000032"
	for _, tc := range []struct {
		name            string
		read, patch     int
		readBody, reply string
		accepted        bool
	}{
		{"success", 200, 204, `{"Identity":"read-model-identity","Id":"` + readID + `","Description":"preserved-before","Future":false}`, "", true},
		{"read-denied", 403, 0, `{"Message":"fixture-private"}`, "", false},
		{"read-empty", 204, 0, "", "", false},
		{"read-malformed", 200, 0, "not-json", "", false},
		{"patch-throttled", 200, 429, `{"Id":"` + readID + `"}`, `{"Message":"fixture-private"}`, false},
		{"patch-unavailable", 200, 503, `{"Id":"` + readID + `"}`, "not-json", false},
		{"accepted-malformed", 200, 200, `{"Id":"` + readID + `"}`, "not-json", true},
	} {
		steps := []contracttest.Exchange{wire("GET", item, "Set-CsTeamsCustomBannerText", nil, tc.read, tc.readBody)}
		if tc.patch != 0 {
			steps = append(steps, wire("PATCH", item, "Set-CsTeamsCustomBannerText", json.RawMessage(`{"Id":"`+readID+`","Description":null}`), tc.patch, tc.reply))
		}
		results = append(results, run(t, contracttest.Case{Name: "update-" + tc.name, Mappings: []contracttest.Mapping{{Command: "Set-CsTeamsCustomBannerText", ParameterSet: "ParentAndRelativeKey"}}, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.TeamsCustomBannerTexts.UpdateByID(ctx, relative, resource.UpdateRequest{Settings: resource.Settings{Description: policy.Null[string]()}})
			if v == nil || v.ReadResponse == nil || v.PatchInvoked != (tc.patch != 0) || v.HTTPAccepted != tc.accepted {
				return contracttest.Observation{Err: errors.New("banner update preparation/acceptance lost")}
			}
			if tc.name == "success" {
				if v.Before == nil || v.Before.Description == nil || *v.Before.Description != "preserved-before" || v.After != nil {
					return contracttest.Observation{Err: errors.New("banner read model was modified or fabricated")}
				}
				return contracttest.Observation{Response: r, Err: e}
			}
			if e == nil || tc.accepted && !errors.Is(e, client.ErrDecode) {
				return contracttest.Observation{Err: errors.New("banner partial update error lost")}
			}
			return contracttest.Observation{Response: r}
		}}))
	}
	evidence(t, results)
}

func TestUnitBannerComposedUpdateUsesNullableReadSelector(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, tc := range []struct {
		name, read, body, set string
		byID                  bool
	}{
		{"identity-from-read", `{"Identity":"different-read-identity"}`, `{"Identity":"different-read-identity","Text":""}`, "Identity", false},
		{"null-read-identity", `{"Identity":null}`, `{"Identity":null,"Text":""}`, "Identity", false},
		{"null-read-id", `{"Identity":"ignored-identity","Id":null}`, `{"Id":null,"Text":""}`, "ParentAndRelativeKey", true},
	} {
		results = append(results, run(t, contracttest.Case{Name: tc.name, Mappings: []contracttest.Mapping{{Command: "Set-CsTeamsCustomBannerText", ParameterSet: tc.set}}, Steps: []contracttest.Exchange{
			wire("GET", item, "Set-CsTeamsCustomBannerText", nil, 200, tc.read),
			wire("PATCH", item, "Set-CsTeamsCustomBannerText", json.RawMessage(tc.body), 204, ""),
		}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			input := resource.UpdateRequest{Settings: resource.Settings{Text: policy.Set("")}}
			var v *resource.UpdateResult
			var r *resty.Response
			var e error
			if tc.byID {
				v, r, e = sdk.TeamsAdminAPI.TeamsCustomBannerTexts.UpdateByID(ctx, relative, input)
			} else {
				v, r, e = sdk.TeamsAdminAPI.TeamsCustomBannerTexts.UpdateByIdentity(ctx, "Global/"+relative, input)
			}
			if e != nil || v == nil || v.Before == nil || !v.PatchInvoked || !v.HTTPAccepted || v.After != nil {
				return contracttest.Observation{Err: errors.New("nullable read selector or empty Text was lost")}
			}
			return contracttest.Observation{Response: r}
		}}))
	}
	evidence(t, results)
}
