package targetingpolicy_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	policy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/targeting_policy"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/TeamsTargetingPolicy"

func discovery(t *testing.T) contracttest.Exchange {
	t.Helper()
	body, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Body: body}}
}

func wire(method, path, command string, body json.RawMessage, status int, reply string) contracttest.Exchange {
	headers := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {"https://admin.example.invalid/"}}
	if body != nil {
		headers.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: headers, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}
func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family, c.Source = "targeting_policy", "msteams/teams_admin_api/targeting_policy/service_test.go"
	if c.Scope == "" {
		c.Scope = "service"
	}
	r, err := contracttest.Run(t.Context(), c)
	if err != nil || !r.Passed() {
		t.Fatalf("targeting policy differs: setup=%v differences=%+v", err, r.Differences())
	}
	return r
}

func TestUnitTargetingRequestVectors(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, method, path, command string
		body                        json.RawMessage
	}{
		{"list", "GET", collection, "Get-CsTeamsTargetingPolicy", nil},
		{"read", "GET", collection + "/configuration/Global", "Get-CsTeamsTargetingPolicy", nil},
		{"update", "PUT", collection + "/configuration/Global", "Set-CsTeamsTargetingPolicy", json.RawMessage(`{"Identity":"Global","CustomTagsMode":"Enabled","Description":""}`)},
		{"delete", "DELETE", collection + "/configuration/Global", "Remove-CsTeamsTargetingPolicy", nil},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Steps: []contracttest.Exchange{discovery(t), wire(tc.method, tc.path, tc.command, tc.body, 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TargetingPolicies
				var o contracttest.Observation
				switch tc.name {
				case "list":
					o.Value, o.Response, o.Err = s.List(ctx)
				case "read":
					o.Value, o.Response, o.Err = s.GetByIdentity(ctx, "Global")
				case "update":
					enabled, empty := "Enabled", ""
					o.Response, o.Err = s.UpdateGlobal(ctx, policy.UpdateRequest{CustomTagsMode: &enabled, Description: &empty})
				case "delete":
					o.Response, o.Err = s.DeleteByIdentity(ctx, "Global")
				}
				return o
			}, Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden}})
		})
	}
}

func TestUnitSparseIdentityUpdateAndOpenValues(t *testing.T) {
	t.Parallel()
	identity := "tag:Fixture / Policy"
	mode, empty := "FutureServerMode", ""
	run(t, contracttest.Case{Name: "sparse-open-mode", Steps: []contracttest.Exchange{discovery(t), wire("PUT", collection+"/configuration/"+url.PathEscape(identity), "Set-CsTeamsTargetingPolicy", json.RawMessage(`{"Identity":"tag:Fixture / Policy","ManageTagsPermissionMode":"FutureServerMode","SuggestedPresetTags":""}`), 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		r, err := sdk.TeamsAdminAPI.TargetingPolicies.UpdateByIdentity(ctx, identity, policy.UpdateRequest{ManageTagsPermissionMode: &mode, SuggestedPresetTags: &empty})
		return contracttest.Observation{Response: r, Err: err}
	}, Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden}})
	for _, id := range []string{"", " ", "..", "bad\nidentity"} {
		run(t, contracttest.Case{Name: "invalid-identity", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			r, err := sdk.TeamsAdminAPI.TargetingPolicies.UpdateByIdentity(ctx, id, policy.UpdateRequest{})
			return contracttest.Observation{Response: r, Err: err}
		}, Want: contracttest.Expected{Error: policy.ErrInput}})
	}
}

func TestUnitTargetingLocalModelRetainsRawFields(t *testing.T) {
	t.Parallel()
	var value policy.Policy
	if err := json.Unmarshal([]byte(`{"Identity":"tag:Later","CustomTagsMode":"Future","Description":null,"future":{"v":5}}`), &value); err != nil || value.CustomTagsMode == nil || *value.CustomTagsMode != "Future" || string(value.RawFields["Description"]) != "null" || string(value.RawFields["future"]) != `{"v":5}` {
		t.Fatal("local model lost null/open/future values")
	}
}

func TestUnitCanceledPolicyDoesNotDiscoverOrAcquireCredentials(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "cancelled", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		ctx, stop := context.WithCancel(ctx)
		stop()
		r, err := sdk.TeamsAdminAPI.TargetingPolicies.UpdateGlobal(ctx, policy.UpdateRequest{})
		return contracttest.Observation{Response: r, Err: err}
	}, Want: contracttest.Expected{Error: context.Canceled}})
	var value policy.Policy
	if err := json.Unmarshal([]byte(`{"Identity":"Global","CustomTagsMode":17}`), &value); err == nil {
		t.Fatal("invalid typed setting accepted")
	}
}

func TestUnitTargetingFilterRequestAndLocalValidation(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "local-tag-filter", Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsTargetingPolicy", ParameterSet: "Filter"}}, Steps: []contracttest.Exchange{discovery(t), wire("GET", collection, "Get-CsTeamsTargetingPolicy", nil, 403, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, err := sdk.TeamsAdminAPI.TargetingPolicies.ListFiltered(ctx, "fIX*")
		return contracttest.Observation{Value: v, Response: r, Err: err}
	}, Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden}})

	run(t, contracttest.Case{Name: "bad-filter-local", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, err := sdk.TeamsAdminAPI.TargetingPolicies.ListFiltered(ctx, "[z-a]")
		return contracttest.Observation{Value: v, Response: r, Err: err}
	}, Want: contracttest.Expected{Error: policy.ErrInput}})
}

func TestUnitTargetingFailuresDoNotInventSuccessOrReplayWrites(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		status     int
		err        error
	}{
		{"accepted-read", `{"Identity":"Global"}`, 202, policy.ErrResponse},
		{"null-read", `null`, 200, client.ErrDecode},
		{"missing-identity", `{"Description":"fixture"}`, 200, client.ErrDecode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Steps: []contracttest.Exchange{discovery(t), wire("GET", collection+"/configuration/Global", "Get-CsTeamsTargetingPolicy", nil, tc.status, tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				v, r, err := sdk.TeamsAdminAPI.TargetingPolicies.GetByIdentity(ctx, "Global")
				return contracttest.Observation{Value: v, Response: r, Err: err}
			}, Want: contracttest.Expected{Status: tc.status, Error: tc.err}})
		})
	}
	for _, status := range []int{429, 503} {
		run(t, contracttest.Case{Name: "delete-no-replay", Steps: []contracttest.Exchange{discovery(t), wire("DELETE", collection+"/configuration/Global", "Remove-CsTeamsTargetingPolicy", nil, status, `{"code":"Unavailable","message":"private"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			r, err := sdk.TeamsAdminAPI.TargetingPolicies.DeleteByIdentity(ctx, "Global")
			return contracttest.Observation{Response: r, Err: err}
		}, Want: contracttest.Expected{Status: status, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: status, Codes: []string{"Unavailable"}}}})
	}
}

// Native 200 object/array and bodyless 204 vectors are authored from pinned
// ConfigAPI handlers. Null/future retention is SDK behavior, not pinned parity.
func TestUnitTargetingSuccessfulReadFilterAndSparseMutation(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, method, path, command, reply string
		status                             int
		body                               json.RawMessage
	}{
		{"read", "GET", collection + "/configuration/Global", "Get-CsTeamsTargetingPolicy", `{"Identity":"Global","CustomTagsMode":"Enabled","Description":null,"future":false}`, 200, nil},
		{"filter", "GET", collection, "Get-CsTeamsTargetingPolicy", `[{"Identity":"tag:Fixture","Description":"owned"},{"Identity":"Global"},{"Identity":"tag:Other"}]`, 200, nil},
		{"list", "GET", collection, "Get-CsTeamsTargetingPolicy", `[{"Identity":"tag:Fixture","Description":"owned"},{"Identity":"Global"},{"Identity":"tag:Other"}]`, 200, nil},
		{"update", "PUT", collection + "/configuration/Global", "Set-CsTeamsTargetingPolicy", "", 204, json.RawMessage(`{"Identity":"Global","Description":""}`)},
		{"delete", "DELETE", collection + "/configuration/Global", "Remove-CsTeamsTargetingPolicy", "", 204, nil},
	} {
		step := wire(tc.method, tc.path, tc.command, tc.body, tc.status, tc.reply)
		step.Reply.ProtocolSources = []string{"https://www.powershellgallery.com/packages/MicrosoftTeams/8.1.0"}
		set := "Identity"
		if tc.name == "filter" {
			set = "Filter"
		}
		run(t, contracttest.Case{Name: "positive-" + tc.name, Mappings: []contracttest.Mapping{{Command: tc.command, ParameterSet: set}}, Steps: []contracttest.Exchange{discovery(t), step}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TargetingPolicies
			switch tc.name {
			case "read":
				v, r, err := s.GetByIdentity(ctx, "Global")
				if err == nil && (v == nil || v.Identity != "Global" || v.CustomTagsMode == nil || *v.CustomTagsMode != "Enabled" || v.Description != nil || string(v.RawFields["Description"]) != "null" || string(v.RawFields["future"]) != "false") {
					t.Fatal("policy snapshot lost typed/raw fields")
				}
				return contracttest.Observation{Response: r, Err: err}
			case "list":
				v, r, err := s.List(ctx)
				if err == nil && (len(v) != 3 || v[0].Identity != "Global" || v[1].Identity != "tag:Fixture" || v[2].Identity != "tag:Other") {
					t.Fatal("native list sort or identities differ")
				}
				return contracttest.Observation{Response: r, Err: err}
			case "filter":
				v, r, err := s.ListFiltered(ctx, "fIX*")
				if err == nil && (len(v) != 1 || v[0].Identity != "tag:Fixture") {
					t.Fatal("case-insensitive tag filter did not select exact policy")
				}
				return contracttest.Observation{Response: r, Err: err}
			case "update":
				empty := ""
				r, err := s.UpdateGlobal(ctx, policy.UpdateRequest{Description: &empty})
				return contracttest.Observation{Response: r, Err: err}
			default:
				r, err := s.DeleteByIdentity(ctx, "Global")
				return contracttest.Observation{Response: r, Err: err}
			}
		}, Want: contracttest.Expected{Status: tc.status, Body: []byte(tc.reply)}})
	}
}
