package teamsmeetingconfiguration_test

import (
	"encoding/json"
	"net/http"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const source = "msteams/teams_admin_api/teams_meeting_configuration/testdata/get_global.json"
const family = "teams_meeting_configuration"
const collection = "/Skype.Policy/configurations/TeamsMeetingConfiguration"
const selected = collection + "/configuration/Global"
const get = "Get-CsTeamsMeetingConfiguration"
const set = "Set-CsTeamsMeetingConfiguration"

func compiledWire(method, path, command string, body json.RawMessage, status int, reply string) contracttest.Exchange {
	h := http.Header{}
	h.Set("MPACmdlet", "true")
	h.Set("X-MS-CmdletName", command)
	h[http.CanonicalHeaderKey("X-MS-Target-Uri")] = nil
	if body != nil {
		h.Set("Content-Type", "application/json; charset=utf-8")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json; charset=utf-8"}}, Body: []byte(reply)}}
}

func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family = family
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

// withContexts returns the actual Global response with its context list
// replaced, for local PSListModifier tests that need a nonempty current list.
func withContexts(t *testing.T, contexts string) string {
	t.Helper()
	var body map[string]json.RawMessage
	if json.Unmarshal(readResponseFixture(t, "get_global.json"), &body) != nil {
		t.Fatal("invalid actual response")
	}
	body["PublishedEntraAuthenticationContexts"] = json.RawMessage(contexts)
	raw, err := json.Marshal(body)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func jsonEqual(a, b []byte) bool {
	var x, y any
	return json.Unmarshal(a, &x) == nil && json.Unmarshal(b, &y) == nil && reflect.DeepEqual(x, y)
}
