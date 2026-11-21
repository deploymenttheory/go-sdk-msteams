package onlinevoicemailvalidationconfiguration_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const source = "contracts/endpoints/8.1.0/fixtures/voicemail-validation-provider.json"
const origin = "https://api.interfaces.records.teams.microsoft.com"
const scoped = "scope /Fixture / Child"
const collection = "/Skype.Policy/configurations/OnlineVoicemailValidationConfiguration"
const item = collection + "/configuration/Global"

type capture struct {
	Name, Operation, Response string
	Status                    int
	Force                     bool
	NullURI                   bool `json:"null_uri"`
	Capture                   struct {
		Method, Url, Body, ContentType, CmdletName, MpaCmdlet string
		FailureType                                           *string
		Sends, TokenRequests                                  int
		CorrelationIsGuid, HasTargetUri                       bool
	}
}

func captured(t *testing.T) []capture {
	t.Helper()
	raw, err := os.ReadFile("../../../" + source)
	if err != nil {
		t.Fatal(err)
	}
	var value struct {
		Cases []capture `json:"cases"`
	}
	if json.Unmarshal(raw, &value) != nil || len(value.Cases) != 10 {
		t.Fatal("invalid voicemail provider capture")
	}
	for _, c := range value.Cases {
		if c.Capture.Sends != 1 || c.Capture.TokenRequests != 0 || !c.Capture.CorrelationIsGuid || c.Capture.HasTargetUri || c.Capture.MpaCmdlet != "true" {
			t.Fatal("unsafe/incomplete voicemail capture")
		}
	}
	return value.Cases
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
	c.Family = "online_voicemail_validation_configuration"
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
		t.Fatalf("voicemail contract differs: %v %+v", e, r.Differences())
	}
	return r
}
func evidence(t *testing.T, results []contracttest.Result) {
	t.Helper()
	revision := os.Getenv("MSTEAMS_CONTRACT_REVISION")
	if revision == "" {
		return
	}
	root, err := os.OpenRoot(t.ArtifactDir())
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = root.Close() }()
	f, err := root.Create("voicemail-validation-service.json")
	if err != nil {
		t.Fatal(err)
	}
	writeErr, closeErr := contracttest.WriteEvidence(f, revision, results), f.Close()
	if writeErr != nil || closeErr != nil {
		t.Fatal("voicemail evidence write failed")
	}
}
