package recordingmessages_test

import (
	"encoding/json"
	"errors"
	"io"
	"net/http"
	"regexp"
	"strings"
	"testing"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/recording_messages"
)

type transportFunc func(*http.Request) (*http.Response, error)

func (f transportFunc) RoundTrip(r *http.Request) (*http.Response, error) { return f(r) }

func TestRecordingAutoCreateRetainsFreshIDWithoutReplay(t *testing.T) {
	t.Parallel()
	uuid := regexp.MustCompile(`^[a-f0-9]{8}-[a-f0-9]{4}-4[a-f0-9]{3}-[89ab][a-f0-9]{3}-[a-f0-9]{12}$`)
	for _, status := range []int{204, 403, 429, 200} {
		calls := 0
		seen := map[string]bool{}
		correlations := map[string]bool{}
		var sentIdentity string
		transport := transportFunc(func(r *http.Request) (*http.Response, error) {
			calls++
			correlation := r.Header.Get("X-MS-Correlation-Id")
			if r.Method != "POST" || r.URL.String() != origin+collection || r.Header.Get("Authorization") != "Bearer fixture-only" || r.Header.Get("X-MS-CmdletName") != "New-CsTeamsRecordingAndTranscriptionCustomMessage" || r.Header.Get("MPACmdlet") != "true" || r.Header.Get("X-MS-Target-Uri") != "" || !uuid.MatchString(correlation) || correlations[correlation] {
				t.Error("auto-create protocol/destination changed")
			}
			correlations[correlation] = true
			raw, e := io.ReadAll(r.Body)
			var body map[string]json.RawMessage
			if e != nil || json.Unmarshal(raw, &body) != nil || len(body) != 1 || json.Unmarshal(body["Identity"], &sentIdentity) != nil {
				t.Error("auto-create bound an unexpected body field")
			}
			id, ok := strings.CutPrefix(sentIdentity, "Global/")
			if !ok || !uuid.MatchString(id) || seen[id] {
				t.Error("auto-create ID missing, duplicated or malformed")
			}
			seen[id] = true
			reply := ""
			if status == 200 {
				reply = "not-json"
			}
			return &http.Response{StatusCode: status, Header: http.Header{"Content-Type": {"text/plain"}}, Body: io.NopCloser(strings.NewReader(reply)), Request: r}, nil
		})
		sdk, e := msteams.NewClient(&auth.Config{Method: auth.SuppliedTokens, TenantID: "fixture-tenant", IdentityID: "fixture-identity", TokenMode: auth.Application, AccessTokens: map[string]auth.Token{auth.TeamsAdminAudience: {AccessToken: "fixture-only", ExpiresAt: time.Now().Add(time.Hour)}}}, msteams.WithHTTPClient(&http.Client{Transport: transport}))
		if e != nil {
			t.Fatal(e)
		}
		for range 2 {
			v, r, e := sdk.TeamsAdminAPI.RecordingMessages.Create(t.Context(), resource.CreateRequest{})
			if v == nil || r == nil || v.Identity != sentIdentity || "Global/"+v.ID != sentIdentity || v.HTTPAccepted != (status == 204 || status == 200) || v.Value != nil {
				t.Fatal("generated ID or acceptance was lost")
			}
			if status == 204 && e != nil || status != 204 && e == nil || status == 200 && !errors.Is(e, client.ErrDecode) {
				t.Fatal("auto-create error/acceptance changed")
			}
		}
		if calls != 2 || len(seen) != 2 || len(correlations) != 2 {
			t.Fatal("auto-create retried or reused an ID")
		}
		if sdk.Close() != nil {
			t.Fatal("fixture SDK close failed")
		}
	}
}
