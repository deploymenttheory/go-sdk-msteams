package teamsmedialoggingpolicy_test

import (
	"encoding/json"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const source = "msteams/teams_admin_api/teams_media_logging_policy/testdata/get_global.json"

func compiledWire(method, path, command string, body json.RawMessage, status int, reply string) contracttest.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {"https://admin.example.invalid/"}, "MPACmdlet": nil}
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}
func discovery(t *testing.T) contracttest.Exchange {
	t.Helper()
	raw, e := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if e != nil {
		t.Fatal(e)
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: raw}}
}
func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family = "teams_media_logging_policy"
	if c.Source == "" {
		c.Source = source
	}
	if len(c.Steps) == 0 {
		c.Scope = "local"
	} else {
		c.Scope = "service"
		c.Steps = append([]contracttest.Exchange{discovery(t)}, c.Steps...)
		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	}
	r, e := contracttest.Run(t.Context(), c)
	if e != nil || !r.Passed() {
		t.Fatalf("configuration contract %s differs: %v %+v", c.Name, e, r.Differences())
	}
	return r
}

func runExplicit(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family = "teams_media_logging_policy"
	c.Source = source
	if len(c.Steps) == 0 {
		c.Scope = "local"
	} else {
		c.Scope = "service"
		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	}
	r, e := contracttest.Run(t.Context(), c)
	if e != nil || !r.Passed() {
		t.Fatalf("media logging contract differs: %v %+v", e, r.Differences())
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
