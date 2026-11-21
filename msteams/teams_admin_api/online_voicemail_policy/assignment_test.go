package onlinevoicemailpolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/online_voicemail_policy"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

const generatedSource = "contracts/endpoints/8.1.0/fixtures/voicemail-generated-assignments.json"
const forest = "https://fixture-forest.example.invalid/"

func discovery() contracttest.Exchange {
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Body: []byte(`{"endpoints":{"adminServiceEndpoint":"fixture-forest.example.invalid:443/path?q=f#x"}}`)}}
}
func generatedWire(method, url, command string, body json.RawMessage, status int, reply string) contracttest.Exchange {
	h := http.Header{}
	h.Set("Accept", "application/json")
	h.Set("X-MS-CmdletName", command)
	h.Set("X-MS-Target-Uri", forest)
	h[http.CanonicalHeaderKey("MPACmdlet")] = nil
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: url, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}

func TestVoicemailGeneratedAssignmentRequests(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../" + generatedSource)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Cases []struct {
			Name    string          `json:"name"`
			Method  string          `json:"method"`
			URL     string          `json:"generated_url"`
			Reply   string          `json:"response_body"`
			Status  int             `json:"response_status"`
			Body    json.RawMessage `json:"body"`
			Failure *string         `json:"decoder_failure"`
			Sends   int             `json:"sends"`
		} `json:"cases"`
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Cases) != 14 {
		t.Fatal("invalid generated assignment capture")
	}
	for _, row := range fixture.Cases {
		if row.Sends != 1 {
			t.Fatal("invalid generated exchange")
		}
	}
	var results []contracttest.Result
	for _, f := range fixture.Cases {
		set, command := "Identity", "Grant-CsOnlineVoicemailPolicy"
		if strings.HasPrefix(f.Name, "tenant") {
			set = "GrantToTenant"
		}
		if strings.HasPrefix(f.Name, "group") || strings.HasPrefix(f.Name, "operation") {
			set = "GrantToGroup"
		}
		if strings.HasPrefix(f.Name, "operation") {
			command = "Get-CsBatchPolicyAssignmentOperation"
			f.Body = nil // A JSON null fixture body denotes absent content here.
		}
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Source: generatedSource, Mappings: []contracttest.Mapping{{Command: "Grant-CsOnlineVoicemailPolicy", ParameterSet: set}}, Steps: []contracttest.Exchange{discovery(), generatedWire(f.Method, f.URL, command, f.Body, f.Status, f.Reply)}}
			if f.Failure != nil {
				c.Want.Error = client.ErrDecode
			}
			if f.Status == 403 {
				c.Want.Error = auth.ErrForbidden
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.OnlineVoicemailPolicies
				if strings.HasPrefix(f.Name, "operation") {
					v, r, e := s.GetAssignmentOperation(ctx, "fixture/operation")
					if e == nil && (v == nil || v.OverallStatus == nil || *v.OverallStatus != "Future") {
						e = errors.New("open operation status lost")
					}
					return contracttest.Observation{Response: r, Err: e}
				}
				name := policy.Set("Fixture")
				if strings.Contains(f.Name, "null") || strings.HasPrefix(f.Name, "group-malformed") || f.Name == "group-empty-id" || f.Name == "group-forbidden" {
					name = policy.Null[string]()
				}
				if f.Name == "user-empty" {
					name = policy.Set("")
				}
				var v *resource.AssignmentResult
				if strings.HasPrefix(f.Name, "user") {
					value, r, e := s.AssignUser(ctx, "fixture/user", name)
					return contracttest.Observation{Value: value, Response: r, Err: e}
				}
				if strings.HasPrefix(f.Name, "tenant") {
					value, r, e := s.AssignTenantDefault(ctx, name, strings.Contains(f.Name, "force"))
					return contracttest.Observation{Value: value, Response: r, Err: e}
				}
				var priority *int32
				if f.Name == "group-zero-priority" {
					priority = new(int32)
				}
				if f.Name != "group-accepted" && f.Name != "group-zero-priority" {
					n := int32(-1)
					priority = &n
				}
				v, r, e := s.AssignGroup(ctx, "fixture/group", name, priority)
				if r != nil && r.StatusCode() == 202 && (v == nil || !v.Accepted) {
					return contracttest.Observation{Err: errors.New("group acceptance lost")}
				}
				if e == nil && f.Name == "group-empty-id" && v.OperationID != nil {
					e = errors.New("fabricated operation ID")
				}
				return contracttest.Observation{Response: r, Err: e}
			}
			results = append(results, run(t, c))
		}
	}
	evidence(t, results)
}
