package compliancerecording_test

import (
	"encoding/json"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const (
	source       = "msteams/teams_admin_api/compliance_recording/testdata/get_global.json"
	origin       = "https://api.interfaces.records.teams.microsoft.com"
	policies     = "/Skype.Policy/configurations/TeamsComplianceRecordingPolicy"
	applications = "/Skype.Policy/configurations/TeamsComplianceRecordingApplication"
	tagged       = policies + "/configuration/Tag%3Afixture-policy"
	appID        = "00000000-0000-4000-8000-000000000001"
	pairedID     = "00000000-0000-4000-8000-000000000002"
	appPath      = applications + "/configuration/Tag%3Afixture-policy/" + appID
	grantCommand = "Grant-CsTeamsComplianceRecordingPolicy"
	forest       = "https://admin.example.invalid/"
)

func readResponseFixture(t *testing.T, name string) []byte {
	t.Helper()
	directory, err := os.OpenRoot("testdata")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := directory.Close(); err != nil {
			t.Error("testdata directory close failed")
		}
	}()
	raw, err := directory.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func compiledWire(method, path, command string, body json.RawMessage, status int, reply string) contracttest.Exchange {
	h := http.Header{}
	h.Set("MPACmdlet", "true")
	h.Set("X-MS-CmdletName", command)
	h[http.CanonicalHeaderKey("X-MS-Target-Uri")] = nil
	if body != nil {
		h.Set("Content-Type", "application/json; charset=utf-8")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}

func discovery(t *testing.T) contracttest.Exchange {
	t.Helper()
	raw, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: raw}}
}

func modernWire(method, path string, body json.RawMessage, status int, reply string) contracttest.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {grantCommand}, "X-MS-Target-Uri": {forest}}
	h[http.CanonicalHeaderKey("MPACmdlet")] = nil
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}

func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family = "compliance_recording"
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
		t.Fatalf("compliance recording contract differs: %v %+v", e, r.Differences())
	}
	return r
}
