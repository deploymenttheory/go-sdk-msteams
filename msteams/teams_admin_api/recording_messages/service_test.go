package recordingmessages_test

import (
	"context"
	"encoding/json"

	"slices"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/recording_messages"
)

func fixtureLocalization() *resource.Localization {
	return &resource.Localization{Language: new("en-US"), InitiatorImplicit: new("fixture initiator"), ParticipantImplicit: new(""), ParticipantExplicitRequested: new("fixture requested"), AgreementDialogue: new("fixture agreement")}
}
func TestUnitRecordingTypedProviderRequests(t *testing.T) {
	t.Parallel()

	for _, f := range requestInputs(t) {
		if f.Selector == "auto" {
			continue
		} // Random native IDs are checked with a memory sender separately.
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			x := f.Request.Request
			var body json.RawMessage
			if x.Body != "" {
				body = json.RawMessage(x.Body)
			}
			set := "ParentAndRelativeKey"
			if f.Selector == "identity" {
				set = "Identity"
			}
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: x.CmdletName, ParameterSet: set}}, Steps: []contracttest.Exchange{wire(x.Method, strings.TrimPrefix(x.Url, origin), x.CmdletName, body, f.Status, f.Response)}}
			// These retained vectors prove rejected request bytes, not decoding.
			if f.Status != 400 || f.Response != "" {
				t.Fatal("rejected-wire vector unexpectedly contains a response contract")
			}
			c.Want = contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.RecordingMessages
				controls := policy.Controls{}
				if f.Force {
					controls.Force = new(false)
				}
				id := relative
				if f.Selector == "invalid-id" {
					id = "fixture-non-guid"
				}
				if f.Operation == "list" {
					v, r, e := s.List(ctx)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				if f.Operation == "read" {
					if f.Selector == "identity" {
						v, r, e := s.GetByIdentity(ctx, scoped)
						return contracttest.Observation{Value: v, Response: r, Err: e}
					}
					v, r, e := s.GetByID(ctx, id)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				if f.Operation == "delete" {
					if f.Selector == "identity" {
						v, r, e := s.DeleteByIdentity(ctx, scoped, controls)
						return contracttest.Observation{Value: v, Response: r, Err: e}
					}
					v, r, e := s.DeleteByID(ctx, id, controls)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				settings := resource.Settings{}
				input := resource.CreateRequest{Controls: controls}
				if f.Fields != "none" {
					settings.Description = policy.Set("")
					input.Description = policy.Set("")
				}
				if f.Fields == "all" {
					settings.Localizations = policy.Set([]*resource.Localization{fixtureLocalization()})
					input.Localizations = resource.ReplaceLocalizations(fixtureLocalization())
				}
				if f.Operation == "create" {
					if f.Selector == "identity" {
						v, r, e := s.CreateByIdentity(ctx, scoped, input)
						return contracttest.Observation{Value: v, Response: r, Err: e}
					}
					v, r, e := s.CreateByID(ctx, id, input)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				patch := resource.PatchRequest{Settings: settings, Controls: controls, BodyIdentity: policy.Set("read-model-identity")}
				if f.Selector == "identity" {
					v, r, e := s.PatchByIdentity(ctx, scoped, patch)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				patch.BodyID = policy.Set(relative)
				v, r, e := s.PatchByID(ctx, id, patch)
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}
			run(t, c)
		}
	}
}

func TestUnitRecordingListFiltersUseRealResponses(t *testing.T) {
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
			Name: "real-filter-" + tc.Name, Family: "recording_messages", Scope: "service", Source: "msteams/teams_admin_api/recording_messages/testdata/list.json",
			Steps: responseFixtureSteps(t, "/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage", 200, raw), Want: contracttest.Expected{Status: 200, JSON: expected, Body: raw},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				service := sdk.TeamsAdminAPI.RecordingMessages
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
