package teamscallholdpolicy_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const source = "contracts/endpoints/8.1.0/fixtures/call-hold-provider.json"
const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/TeamsCallHoldPolicy"
const item = collection + "/configuration/Global"
const scoped = "scope /Fixture / Child"

type capture struct {
	Name, Operation, Fields, Response string
	Status                            int
	ForceMode                         int     `json:"force_mode"`
	PolicyName                        *string `json:"policy_name"`
	Capture                           struct {
		Method, Url, Body, ContentType, CmdletName, MpaCmdlet string
		FailureType                                           *string
		Sends, TokenRequests                                  int
		CorrelationIsGuid, HasTargetUri                       bool
	}
}

func captures(t *testing.T) []capture {
	t.Helper()
	raw, e := os.ReadFile("../../../" + source)
	if e != nil {
		t.Fatal(e)
	}
	var v struct {
		Cases []capture `json:"cases"`
	}
	if json.Unmarshal(raw, &v) != nil || len(v.Cases) != 15 {
		t.Fatal("invalid call hold fixture")
	}
	for _, c := range v.Cases {
		r := c.Capture
		if r.Sends != 1 || r.TokenRequests != 0 || !r.CorrelationIsGuid || r.HasTargetUri || r.MpaCmdlet != "true" {
			t.Fatal("unsafe/incomplete call hold fixture")
		}
	}
	return v.Cases
}
func compiledWire(method, path, command string, body json.RawMessage, status int, reply string) contracttest.Exchange {
	h := http.Header{}
	h.Set("MPACmdlet", "true")
	h.Set("X-MS-CmdletName", command)
	h[http.CanonicalHeaderKey("X-MS-Target-Uri")] = nil
	if body != nil {
		h.Set("Content-Type", "application/json; charset=utf-8")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"text/plain; charset=utf-8"}}, Body: []byte(reply)}}
}
func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family = "teams_call_hold_policy"
	if c.Source == "" {
		c.Source = source
	}
	if len(c.Steps) == 0 {
		c.Scope = "local"
	} else {
		c.Scope = "service"
		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	}
	r, e := contracttest.Run(t.Context(), c)
	if e != nil || !r.Passed() {
		t.Fatalf("call hold contract differs: %v %+v", e, r.Differences())
	}
	return r
}
func evidence(t *testing.T, results []contracttest.Result) {
	t.Helper()
	revision := os.Getenv("MSTEAMS_CONTRACT_REVISION")
	if revision == "" {
		return
	}
	root, e := os.OpenRoot(t.ArtifactDir())
	if e != nil {
		t.Fatal(e)
	}
	defer func() { _ = root.Close() }()
	f, e := root.Create("call-hold-service.json")
	if e != nil {
		t.Fatal(e)
	}
	a, b := contracttest.WriteEvidence(f, revision, results), f.Close()
	if a != nil || b != nil {
		t.Fatal("call hold evidence write failed")
	}
}

const getCommand = "Get-CsTeamsCallHoldPolicy"
const newCommand = "New-CsTeamsCallHoldPolicy"
const setCommand = "Set-CsTeamsCallHoldPolicy"
const deleteCommand = "Remove-CsTeamsCallHoldPolicy"
const grantCommand = "Grant-CsTeamsCallHoldPolicy"
const forest = "https://forest.fixture.invalid/"
const grantSource = "contracts/endpoints/8.1.0/fixtures/call-hold-grant-requests.json"
const wrapperSource = "contracts/endpoints/8.1.0/fixtures/call-hold-grant-wrappers.json"

func discovery() contracttest.Exchange {
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(`{"Endpoints":{"AdminServiceEndpoint":"forest.fixture.invalid:444/ignored?x=y#ignored"}}`)}}
}
func modernWire(method, path string, body json.RawMessage, status int, reply string) contracttest.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {grantCommand}, "X-MS-Target-Uri": {forest}}
	h[http.CanonicalHeaderKey("MPACmdlet")] = nil
	if body != nil && string(body) != "null" {
		h.Set("Content-Type", "application/json")
	} else {
		body = nil
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}
