package recordingmessages_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const source = "contracts/endpoints/8.1.0/fixtures/recording-message-provider.json"
const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage"
const relative = "00000000-0000-4000-8000-000000000031"
const scoped = "scope /" + relative + "/ Child"
const item = collection + "/configuration/Global/" + relative

type capture struct {
	Name, Operation, Selector, Fields, Response string
	Status                                      int
	Force                                       bool
	Capture                                     struct {
		PreparedIdentity, RelativeId, Metadata string
		ConstructorIdWasGuid                   *bool
		Request                                struct {
			Method, Url, Body, ContentType, CmdletName, MpaCmdlet string
			FailureType                                           *string
			Sends, TokenRequests                                  int
			CorrelationIsGuid, HasTargetUri                       bool
		}
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
	if json.Unmarshal(raw, &v) != nil || len(v.Cases) != 16 {
		t.Fatal("invalid recording provider fixture")
	}
	for _, c := range v.Cases {
		r := c.Capture.Request
		if r.Sends != 1 || r.TokenRequests != 0 || !r.CorrelationIsGuid || r.HasTargetUri || r.MpaCmdlet != "true" {
			t.Fatal("unsafe/incomplete recording capture")
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
	c.Family = "recording_messages"
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
		t.Fatalf("recording contract differs: %v %+v", e, r.Differences())
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
	f, e := root.Create("recording-message-service.json")
	if e != nil {
		t.Fatal(e)
	}
	a, b := contracttest.WriteEvidence(f, revision, results), f.Close()
	if a != nil || b != nil {
		t.Fatal("recording evidence write failed")
	}
}
