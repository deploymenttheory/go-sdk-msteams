package audioconferencingprompts_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const source = "contracts/endpoints/8.1.0/fixtures/audio-conferencing-prompts-provider.json"
const sourceModels = "contracts/endpoints/8.1.0/fixtures/audio-conferencing-prompts-models.json"
const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/TeamsAudioConferencingCustomPromptsConfiguration"
const item = collection + "/configuration/Global"
const scoped = "scope /Fixture / Child"
const getCommand = "Get-CsTeamsAudioConferencingCustomPromptsConfiguration"
const setCommand = "Set-CsTeamsAudioConferencingCustomPromptsConfiguration"
const firstID = "11111111-1111-1111-1111-111111111111"
const secondID = "22222222-2222-2222-2222-222222222222"

type capture struct {
	Name, Operation, Fields, Response string
	Status                            int
	ForceMode                         int `json:"force_mode"`
	Capture                           struct {
		Method, Url, Body, CmdletName, MpaCmdlet string
		FailureType                              *string
		Sends, TokenRequests                     int
		CorrelationIsGuid, HasTargetUri          bool
	}
}

func captures(t *testing.T) []capture {
	t.Helper()
	raw, e := os.ReadFile("../../../" + source)
	if e != nil {
		t.Fatal(e)
	}
	var v struct{ Cases []capture }
	if json.Unmarshal(raw, &v) != nil || len(v.Cases) != 15 {
		t.Fatal("incomplete prompts provider fixture")
	}
	for _, c := range v.Cases {
		x := c.Capture
		if x.Sends != 1 || x.TokenRequests != 0 || !x.CorrelationIsGuid || x.HasTargetUri || x.MpaCmdlet != "true" {
			t.Fatal("unsafe/incomplete source fixture")
		}
	}
	return v.Cases
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
	c.Family = "teams_audio_conferencing_custom_prompts_configuration"
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
		t.Fatalf("prompts contract differs: %v %+v", e, r.Differences())
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
	f, e := root.Create("audio-conferencing-prompts-service.json")
	if e != nil {
		t.Fatal(e)
	}
	a, b := contracttest.WriteEvidence(f, revision, results), f.Close()
	if a != nil || b != nil {
		t.Fatal("prompts evidence write failed")
	}
}
