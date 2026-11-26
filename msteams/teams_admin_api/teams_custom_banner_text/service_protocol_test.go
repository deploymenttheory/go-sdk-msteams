package teamscustombannertext_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_custom_banner_text"
)

// Primary-authored Core Banner model and selector grammar, not a tenant capture.
// Current public New examples also bind nonempty Text/Description; pinned 8.1.0
// metadata/setter rules remain authoritative over the drifting help signature.
// Text is this resource's string, not recording-message localization data.
const authoredBannerIdentity = "Global/" + relative
const authoredBanner = `{"Identity":"` + authoredBannerIdentity + `","Id":"` + relative + `","Text":"fixture text","Description":null,"Future":false}`

func verifyAuthoredBanner(t *testing.T, value *resource.Banner) {
	t.Helper()
	if value == nil || value.Identity == nil || *value.Identity != authoredBannerIdentity || value.ID == nil || *value.ID != relative || value.Text == nil || *value.Text != "fixture text" || value.Description != nil || string(value.RawFields["Description"]) != "null" || string(value.RawFields["Future"]) != "false" {
		t.Fatal("banner lost typed identity/text, nullable description or raw extension")
	}
}

func TestUnitBannerAuthoredSelectorResponses(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, operation, selector, method, command, path string
		body                                             json.RawMessage
		status                                           int
	}{
		{"get-identity", "get", "identity", "GET", "Get-CsTeamsCustomBannerText", item, nil, 200},
		{"get-id", "get", "id", "GET", "Get-CsTeamsCustomBannerText", item, nil, 200},
		{"create-identity", "create", "identity", "POST", newCommand, collection, json.RawMessage(`{"Identity":"` + authoredBannerIdentity + `","Id":"` + relative + `","Text":"fixture text","Description":"fixture description"}`), 201},
		{"create-id", "create", "id", "POST", newCommand, collection, json.RawMessage(`{"Identity":"` + authoredBannerIdentity + `","Id":null,"Text":"fixture text","Description":"fixture description"}`), 201},
		{"patch-identity", "patch", "identity", "PATCH", setCommand, item, json.RawMessage(`{"Identity":"` + authoredBannerIdentity + `","Text":"","Description":null}`), 204},
		{"patch-id", "patch", "id", "PATCH", setCommand, item, json.RawMessage(`{"Id":"` + relative + `","Text":"","Description":null}`), 204},
		{"remove-identity", "remove", "identity", "DELETE", deleteCommand, item, json.RawMessage(`{"Identity":"` + authoredBannerIdentity + `"}`), 204},
		{"remove-id", "remove", "id", "DELETE", deleteCommand, item, json.RawMessage(`{"Id":"` + relative + `"}`), 204},
	} {
		t.Run(tc.name, func(t *testing.T) {
			set, reply := "Identity", ""
			if tc.selector == "id" {
				set = "ParentAndRelativeKey"
			}
			want := contracttest.Expected{Body: []byte{}}
			if tc.operation == "get" {
				reply = authoredBanner
				want = contracttest.Expected{JSON: json.RawMessage(reply), Body: []byte(reply)}
			}
			run(t, contracttest.Case{Name: "typed-" + tc.name, Source: "msteams/teams_admin_api/teams_custom_banner_text/service_protocol_test.go", Mappings: []contracttest.Mapping{{Command: tc.command, ParameterSet: set}}, Steps: []contracttest.Exchange{wire(tc.method, tc.path, tc.command, tc.body, tc.status, reply)}, Want: want, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TeamsCustomBannerTexts
				var value *resource.Banner
				var observed contracttest.Observation
				switch tc.operation {
				case "get":
					if tc.selector == "id" {
						value, observed.Response, observed.Err = s.GetByID(ctx, relative)
					} else {
						value, observed.Response, observed.Err = s.GetByIdentity(ctx, authoredBannerIdentity)
					}
				case "create":
					input := resource.CreateRequest{Text: "fixture text", Description: policy.Set("fixture description")}
					if tc.selector == "id" {
						value, observed.Response, observed.Err = s.CreateByID(ctx, relative, input)
					} else {
						value, observed.Response, observed.Err = s.CreateByIdentity(ctx, authoredBannerIdentity, input)
					}
				case "patch":
					input := resource.PatchRequest{Settings: resource.Settings{Text: policy.Set(""), Description: policy.Null[string]()}}
					if tc.selector == "id" {
						input.BodyID = policy.Set(relative)
						value, observed.Response, observed.Err = s.PatchByID(ctx, relative, input)
					} else {
						input.BodyIdentity = policy.Set(authoredBannerIdentity)
						value, observed.Response, observed.Err = s.PatchByIdentity(ctx, authoredBannerIdentity, input)
					}
				default:
					if tc.selector == "id" {
						value, observed.Response, observed.Err = s.DeleteByID(ctx, relative, policy.Controls{})
					} else {
						value, observed.Response, observed.Err = s.DeleteByIdentity(ctx, authoredBannerIdentity, policy.Controls{})
					}
				}
				if observed.Err == nil {
					if tc.operation == "get" {
						verifyAuthoredBanner(t, value)
					} else if value != nil {
						t.Fatal("empty banner receipt fabricated a policy or readback")
					}
				}
				observed.Value = value
				return observed
			}})
		})
	}
}

func TestUnitBannerAuthoredNonemptyFilterPreservesTypedItem(t *testing.T) {
	t.Parallel()
	other := strings.ReplaceAll(authoredBanner, relative, "01234567-89ab-cdef-0123-456789abcdee")
	reply := "[" + other + "," + authoredBanner + "]"
	want := "[" + authoredBanner + "]"
	run(t, contracttest.Case{Name: "nonempty-banner-filter", Source: "msteams/teams_admin_api/teams_custom_banner_text/service_protocol_test.go", Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsCustomBannerText", ParameterSet: "Filter"}}, Steps: []contracttest.Exchange{wire("GET", collection, "Get-CsTeamsCustomBannerText", nil, 200, reply)}, Want: contracttest.Expected{JSON: json.RawMessage(want), Body: []byte(reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		values, response, err := sdk.TeamsAdminAPI.TeamsCustomBannerTexts.ListFiltered(ctx, "*"+relative)
		if err == nil {
			if len(values) != 1 {
				t.Fatal("nonempty banner filter selected the wrong resource")
			}
			verifyAuthoredBanner(t, values[0])
		}
		return contracttest.Observation{Value: values, Response: response, Err: err}
	}})
}

func TestUnitBannerAuthoredMalformed200RetainsResponse(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, reply string }{
		{"invalid-id", `{"Id":"not-a-guid"}`},
		{"null-text", `{"Text":null}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Source: "msteams/teams_admin_api/teams_custom_banner_text/service_protocol_test.go", Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsCustomBannerText", ParameterSet: "ParentAndRelativeKey"}}, Steps: []contracttest.Exchange{wire("GET", item, "Get-CsTeamsCustomBannerText", nil, 200, tc.reply)}, Want: contracttest.Expected{Body: []byte(tc.reply), Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.TeamsCustomBannerTexts.GetByID(ctx, relative)
				if value != nil {
					t.Fatal("malformed banner reply exposed a fabricated result")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
		})
	}
}
