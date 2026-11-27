package tenantidentity_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	ti "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_identity"
)

type groupProjectionObservation struct {
	Models []struct {
		JSON      string `json:"input_json"`
		Exception json.RawMessage
		Projected struct {
			ID          *string `json:"Id"`
			DisplayName *string
		}
	}
	Envelopes []struct {
		JSON      string `json:"input_json"`
		Exception json.RawMessage
		Projected json.RawMessage
	}
}

func groupObservations(t *testing.T) groupProjectionObservation {
	t.Helper()
	var fixture groupProjectionObservation
	if json.Unmarshal(readFile(t, "testdata/pinned-find-group.json"), &fixture) != nil || len(fixture.Models) != 31 || len(fixture.Envelopes) != 11 {
		t.Fatal("pinned Find group observations unavailable")
	}
	return fixture
}

func TestUnitFindGroupPublicProjectionUsesPinnedModelAndGUIDRules(t *testing.T) {
	for index, item := range groupObservations(t).Models {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			var group ti.Group
			err := json.Unmarshal([]byte(item.JSON), &group)
			if string(item.Exception) != "null" {
				if !errors.Is(err, client.ErrDecode) {
					t.Fatal("source projection failure accepted or misclassified")
				}
				return
			}
			if err != nil || group.ID == nil || item.Projected.ID == nil || *group.ID != *item.Projected.ID || (group.DisplayName == nil) != (item.Projected.DisplayName == nil) || group.DisplayName != nil && *group.DisplayName != *item.Projected.DisplayName {
				t.Fatal("actual source public projection changed")
			}
			var wire map[string]json.RawMessage
			if json.Unmarshal([]byte(item.JSON), &wire) != nil || string(group.RawFields["Id"]) != string(wire["Id"]) {
				t.Fatal("raw wire identity lost while canonicalizing typed ID")
			}
			group.RawFields["Id"] = json.RawMessage(`"different"`)
			if *group.ID != *item.Projected.ID {
				t.Fatal("raw-field edit changed the source-projected ID")
			}
		})
	}
}

func TestUnitFindGroupRootClientProjectionRetainsEveryReceipt(t *testing.T) {
	for index, item := range groupObservations(t).Envelopes {
		t.Run(fmt.Sprint(index), func(t *testing.T) {
			want := testutil.Expected{Status: 200, Body: []byte(item.JSON), JSON: item.Projected}
			if string(item.Exception) != "null" {
				want.Error = client.ErrDecode
				want.JSON = nil
			}
			run(t, testutil.Case{Name: fmt.Sprintf("find-projection-%d", index),
				Steps: []testutil.Exchange{discovery(t), wire("GET", "/Teams.VoiceApps/groups?SearchQuery=abc", "Find-CsGroup", nil, 200, []byte(item.JSON))}, Want: want,
				Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					value, response, err := sdk.TeamsAdminAPI.TenantIdentity.SearchGroups(ctx, ti.GroupSearchRequest{SearchQuery: "abc"})
					if response == nil || response.StatusCode() != 200 || string(response.Body()) != item.JSON {
						t.Fatal("projection lost accepted receipt")
					}
					if err != nil && value != nil {
						t.Fatal("buffered source failure leaked partial group results")
					}
					return testutil.Observation{Value: value, Response: response, Err: err}
				}})
		})
	}
}

func TestUnitFindGroupSourceQueryBindingAndExactStatus(t *testing.T) {
	for _, item := range []struct {
		name, query string
		request     ti.GroupSearchRequest
	}{
		{"whitespace", "SearchQuery=+", ti.GroupSearchRequest{SearchQuery: " "}},
		{"encoded-control", "SearchQuery=a%0Ab", ti.GroupSearchRequest{SearchQuery: "a\nb"}},
		{"explicit-zero-and-false", "ExactMatchOnly=False&MailEnabledOnly=False&MaxResults=0&SearchQuery=abc", ti.GroupSearchRequest{SearchQuery: "abc", MaxResults: ptr(uint32(0)), ExactMatchOnly: ptr(false), MailEnabledOnly: ptr(false)}},
		{"signed-boundary", "MaxResults=2147483647&SearchQuery=abc", ti.GroupSearchRequest{SearchQuery: "abc", MaxResults: ptr(uint32(2147483647))}},
	} {
		t.Run(item.name, func(t *testing.T) {
			reply := []byte(`{"Groups":[]}`)
			run(t, testutil.Case{Name: "find-query-" + item.name, Steps: []testutil.Exchange{discovery(t), wire("GET", "/Teams.VoiceApps/groups?"+item.query, "Find-CsGroup", nil, 200, reply)}, Want: testutil.Expected{Status: 200, Body: reply, JSON: json.RawMessage(`[]`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.TenantIdentity.SearchGroups(ctx, item.request)
				return testutil.Observation{Value: v, Response: r, Err: e}
			}})
		})
	}
	for _, status := range []int{201, 204} {
		t.Run(fmt.Sprint(status), func(t *testing.T) {
			run(t, testutil.Case{Name: fmt.Sprintf("find-undeclared-%d", status), Steps: []testutil.Exchange{discovery(t), wire("GET", "/Teams.VoiceApps/groups?SearchQuery=abc", "Find-CsGroup", nil, status, nil)}, Want: testutil.Expected{Status: status, Error: ti.ErrResponse}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.TenantIdentity.SearchGroups(ctx, ti.GroupSearchRequest{SearchQuery: "abc"})
				return testutil.Observation{Value: v, Response: r, Err: e}
			}})
		})
	}
	// Exact duplicates fail even in unprojected nested diagnostic/unknown data.
	for _, reply := range []string{`{"Groups":[],"Groups":null}`, `{"future":{"x":1,"x":2}}`, `{"Groups":[{"Id":"00000000-0000-0000-0000-000000000000","DisplayName":1,"DisplayName":"x"}]}`} {
		run(t, testutil.Case{Name: "find-exact-duplicate", Steps: []testutil.Exchange{discovery(t), wire("GET", "/Teams.VoiceApps/groups?SearchQuery=abc", "Find-CsGroup", nil, 200, []byte(reply))}, Want: testutil.Expected{Status: 200, Body: []byte(reply), Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			v, r, e := sdk.TeamsAdminAPI.TenantIdentity.SearchGroups(ctx, ti.GroupSearchRequest{SearchQuery: "abc"})
			return testutil.Observation{Value: v, Response: r, Err: e}
		}})
	}
}
