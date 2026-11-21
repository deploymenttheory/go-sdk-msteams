package teamsaudioconferencingpolicy_test

import (
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const source = "contracts/endpoints/8.1.0/fixtures/audio-conferencing-policy-requests.json"
const modelSource = "contracts/endpoints/8.1.0/fixtures/audio-conferencing-policy-models.json"
const wrapperSource = "contracts/endpoints/8.1.0/fixtures/audio-conferencing-policy-wrappers.json"
const origin = "https://api.interfaces.records.teams.microsoft.com"
const forest = "https://forest.fixture.invalid/"
const collection = "/Skype.Policy/configurations/TeamsAudioConferencingPolicy"
const item = collection + "/configuration/Global"
const identity = "tag:Fixture / Policy"
const getCommand = "Get-CsTeamsAudioConferencingPolicy"
const newCommand = "New-CsTeamsAudioConferencingPolicy"
const setCommand = "Set-CsTeamsAudioConferencingPolicy"
const grantCommand = "Grant-CsTeamsAudioConferencingPolicy"
const deleteCommand = "Remove-CsTeamsAudioConferencingPolicy"

type requestCapture struct {
	Name, Operation, Method string
	URL                     string `json:"generated_url"`
	Status                  int    `json:"response_status"`
	Reply                   string `json:"response_body"`
	Sends                   int
	Canceled                bool
	Body                    json.RawMessage
}

func captures(t *testing.T) []requestCapture {
	t.Helper()
	raw, e := os.ReadFile("../../../" + source)
	if e != nil {
		t.Fatal(e)
	}
	var v struct{ Cases []requestCapture }
	if json.Unmarshal(raw, &v) != nil || len(v.Cases) != 15 {
		t.Fatal("incomplete generated audio policy fixture")
	}
	for _, c := range v.Cases {
		if c.Canceled && c.Sends != 0 || !c.Canceled && c.Sends != 1 {
			t.Fatal("unexpected source send count")
		}
	}
	return v.Cases
}
func discovery() contracttest.Exchange {
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(`{"Endpoints":{"AdminServiceEndpoint":"forest.fixture.invalid:444/ignored?x=y#ignored"}}`)}}
}
func wire(method, path, command string, body json.RawMessage, status int, reply string) contracttest.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {forest}}
	h[http.CanonicalHeaderKey("MPACmdlet")] = nil
	if body != nil && string(body) != "null" {
		h.Set("Content-Type", "application/json")
	} else {
		body = nil
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}
func capturedWire(t *testing.T, c requestCapture, command string) contracttest.Exchange {
	t.Helper()
	u, e := url.Parse(c.URL)
	if e != nil {
		t.Fatal(e)
	}
	parts := strings.Split(u.EscapedPath(), "/")
	for i, p := range parts {
		v, e := url.PathUnescape(p)
		if e != nil {
			t.Fatal(e)
		}
		parts[i] = url.PathEscape(v)
	}
	return wire(c.Method, strings.Join(parts, "/"), command, c.Body, c.Status, c.Reply)
}
func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family = "teams_audio_conferencing_policy"
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
		t.Fatalf("audio policy contract differs: %v %+v", e, r.Differences())
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
	f, e := root.Create("audio-conferencing-policy-service.json")
	if e != nil {
		t.Fatal(e)
	}
	a, b := contracttest.WriteEvidence(f, revision, results), f.Close()
	if a != nil || b != nil {
		t.Fatal("audio policy evidence write failed")
	}
}
