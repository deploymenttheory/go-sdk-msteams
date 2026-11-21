package teamscustombannertext_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const source = "contracts/endpoints/8.1.0/fixtures/banner-text-provider.json"
const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/CustomBannerText"
const relative = "01234567-89ab-cdef-0123-456789abcdef"
const item = collection + "/configuration/Global/" + relative

const binderSource = "contracts/endpoints/8.1.0/fixtures/banner-text-binding.json"
const getCommand = "Get-CsTeamsCustomBannerText"
const newCommand = "New-CsTeamsCustomBannerText"
const setCommand = "Set-CsTeamsCustomBannerText"
const deleteCommand = "Remove-CsTeamsCustomBannerText"

type capture struct {
	Name, Operation, Selector, Identity, ID, Input, Response string
	Status                                                   int
	ForceMode                                                int    `json:"force_mode"`
	ReadID                                                   string `json:"read_model_id"`
	Observed                                                 struct {
		Identity string `json:"normalized_identity"`
		Request  struct {
			Method, Url, Body, CmdletName   string
			FailureType                     *string
			Sends, TokenRequests            int
			CorrelationIsGuid, HasTargetUri bool
			MpaCmdlet                       string
		} `json:"capture"`
	} `json:"observed"`
}

func captures(t *testing.T) []capture {
	t.Helper()
	raw, err := os.ReadFile("../../../" + source)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct{ Cases []capture }
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Cases) != 24 {
		t.Fatal("invalid banner provider fixture")
	}
	for _, c := range fixture.Cases {
		v := c.Observed.Request
		if v.TokenRequests != 0 || (v.Sends != 0 && (v.Sends != 1 || !v.CorrelationIsGuid || v.HasTargetUri || v.MpaCmdlet != "true")) {
			t.Fatal("unsafe/incomplete banner fixture")
		}
	}
	return fixture.Cases
}

func wire(method, path, command string, body json.RawMessage, status int, reply string) contracttest.Exchange {
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
	c.Family = "teams_custom_banner_text"
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
		t.Fatalf("banner contract %s differs: %v %+v", c.Name, e, r.Differences())
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
	f, e := root.Create("banner-text-service.json")
	if e != nil {
		t.Fatal(e)
	}
	a, b := contracttest.WriteEvidence(f, revision, results), f.Close()
	if a != nil || b != nil {
		t.Fatal("banner evidence write failed")
	}
}
