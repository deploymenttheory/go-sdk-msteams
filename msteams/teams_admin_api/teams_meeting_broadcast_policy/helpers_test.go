package teamsmeetingbroadcastpolicy_test

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"

func compiledWire(method, path, command string, body json.RawMessage, status int, reply string) testutil.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {"https://admin.example.invalid/"}, "MPACmdlet": nil}
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}
func discovery(t *testing.T) testutil.Exchange {
	t.Helper()
	raw, e := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if e != nil {
		t.Fatal(e)
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: testutil.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: raw}}
}
func run(t *testing.T, c testutil.Case) testutil.Result {
	t.Helper()

	if len(c.Steps) == 0 {

	} else {

		c.Steps = append([]testutil.Exchange{discovery(t)}, c.Steps...)
		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	}
	r, e := testutil.Run(t.Context(), c)
	if e != nil || !r.Passed() {
		t.Fatalf("configuration contract %s differs: %v %+v", c.Name, e, r.Differences())
	}
	return r
}

func readResponseFixture(t *testing.T, name string) []byte {
	t.Helper()
	directory, e := os.OpenRoot("testdata")
	if e != nil {
		t.Fatal(e)
	}
	t.Cleanup(func() {
		if e := directory.Close(); e != nil {
			t.Error(e)
		}
	})
	raw, e := directory.ReadFile(name)
	if e != nil {
		t.Fatal(e)
	}
	if (len(raw) != 0 || !strings.HasSuffix(name, ".body")) && !json.Valid(raw) {
		t.Fatal("invalid acceptance fixture")
	}
	return raw
}
