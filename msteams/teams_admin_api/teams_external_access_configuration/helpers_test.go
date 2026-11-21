package teamsexternalaccessconfiguration_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const source = "contracts/endpoints/8.1.0/fixtures/policy-provider-dispatch.json"
const origin = "https://api.interfaces.records.teams.microsoft.com"
const scoped = "scope /Fixture / Child"

type capturedCase struct {
	Name     string `json:"name"`
	Status   int    `json:"status"`
	Response string `json:"response"`
	Capture  struct {
		Method            string
		Url               string
		Body              string
		ContentType       string
		CmdletName        string
		FailureType       *string
		CorrelationIsGuid bool
		HasTargetUri      bool
		MpaCmdlet         string
		Sends             int
		TokenRequests     int
		DecodedNull       *bool
	} `json:"capture"`
}

func captured(t *testing.T) []capturedCase {
	t.Helper()
	raw, err := os.ReadFile("../../../" + source)
	if err != nil {
		t.Fatal(err)
	}
	var v struct {
		Cases []capturedCase `json:"cases"`
	}
	if json.Unmarshal(raw, &v) != nil || len(v.Cases) != 24 {
		t.Fatal("invalid provider capture")
	}
	for _, c := range v.Cases {
		if c.Capture.Sends != 1 || c.Capture.TokenRequests != 0 || !c.Capture.CorrelationIsGuid || c.Capture.HasTargetUri || c.Capture.MpaCmdlet != "true" {
			t.Fatal("unsafe or incomplete provider observation")
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
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"text/plain; charset=utf-8"}, "X-Ms-Configapi-Powershell-Warningmessage": {"fixture-warning"}}, Body: []byte(reply)}}
}
func exchange(c capturedCase) contracttest.Exchange {
	var body json.RawMessage
	if c.Capture.Body != "" {
		body = json.RawMessage(c.Capture.Body)
	}
	return wire(c.Capture.Method, c.Capture.Url[len(origin):], c.Capture.CmdletName, body, c.Status, c.Response)
}
func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family = "teams_external_access_configuration"
	c.Source = source
	if len(c.Steps) > 0 {
		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	} else {
		c.Scope = "local"
	}
	if c.Scope == "" {
		c.Scope = "service"
	}
	r, err := contracttest.Run(t.Context(), c)
	if err != nil || !r.Passed() {
		t.Fatalf("provider differs: %v %+v", err, r.Differences())
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
	f, err := root.Create("teams-external-access-configuration-service.json")
	if err != nil {
		t.Fatal(err)
	}
	writeErr, closeErr := contracttest.WriteEvidence(f, revision, results), f.Close()
	if writeErr != nil || closeErr != nil {
		t.Fatal("provider evidence write failed")
	}
}
