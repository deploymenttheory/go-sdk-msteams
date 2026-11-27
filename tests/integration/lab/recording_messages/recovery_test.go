package recordingmessagesacc

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

func TestUnitOwnedRecordRecoveryRequiresOwnershipAndIndependentAbsence(t *testing.T) {
	t.Parallel()
	const fixture = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-recording-message"
	id, err := ownedRecordID(fixture)
	if err != nil {
		t.Fatal(err)
	}
	identity := "Global/" + id
	owned, err := os.ReadFile("testdata/owned_cleanup.json")
	if err != nil {
		t.Fatal(err)
	}
	foreign := []byte(strings.ReplaceAll(string(owned), fixture, "unrelated-record"))
	misplaced := []byte(strings.Replace(string(owned), identity, "Global/00000000-0000-4000-8000-000000000099", 1))
	wire := func(method string, status int, body []byte) testutil.Exchange {
		path := "/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage"
		var input json.RawMessage
		if method != "LIST" {
			path += "/configuration/" + identity
		} else {
			method = "GET"
		}
		if method == "DELETE" {
			input = json.RawMessage(`{"Identity":"` + identity + `"}`)
		}
		headers := http.Header{}
		if input != nil {
			headers.Set("Content-Type", "application/json; charset=utf-8")
		}
		return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: "https://api.interfaces.records.teams.microsoft.com" + path, Audience: auth.TeamsAdminAudience, Headers: headers, JSON: input}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body}}
	}
	for _, tc := range []struct {
		name   string
		steps  []testutil.Exchange
		entry  lab.Entry
		remove bool
		want   error
	}{
		{name: "invalid-fixture", entry: lab.Entry{Fixture: "foreign"}, want: lab.ErrUnresolved},
		{name: "wrong-resource-id", entry: lab.Entry{Fixture: fixture, ResourceID: "Global/00000000-0000-4000-8000-000000000099"}, want: lab.ErrUnresolved},
		{name: "foreign-description", entry: lab.Entry{Fixture: fixture}, steps: []testutil.Exchange{wire("GET", 200, foreign)}, want: lab.ErrUnresolved},
		{name: "denied-read", entry: lab.Entry{Fixture: fixture}, steps: []testutil.Exchange{wire("GET", 403, nil)}, want: auth.ErrForbidden},
		{name: "verified-absence", entry: lab.Entry{Fixture: fixture}, steps: []testutil.Exchange{wire("GET", 404, nil), wire("LIST", 200, []byte(`[]`))}},
		{name: "denied-collection", entry: lab.Entry{Fixture: fixture}, steps: []testutil.Exchange{wire("GET", 404, nil), wire("LIST", 403, nil)}, want: auth.ErrForbidden},
		{name: "inconsistent-absence", entry: lab.Entry{Fixture: fixture}, steps: []testutil.Exchange{wire("GET", 404, nil), wire("LIST", 200, append(append([]byte{'['}, owned...), ']'))}, want: lab.ErrUnresolved},
		{name: "misplaced-owned-marker", entry: lab.Entry{Fixture: fixture}, steps: []testutil.Exchange{wire("GET", 404, nil), wire("LIST", 200, append(append([]byte{'['}, misplaced...), ']'))}, want: lab.ErrUnresolved},
		{name: "owned-delete", entry: lab.Entry{Fixture: fixture, ResourceID: identity}, remove: true, steps: []testutil.Exchange{wire("GET", 200, owned), wire("DELETE", 204, nil), wire("GET", 404, nil), wire("LIST", 200, []byte(`[]`))}},
		{name: "foreign-delete-refused", entry: lab.Entry{Fixture: fixture, ResourceID: identity}, remove: true, steps: []testutil.Exchange{wire("GET", 200, foreign)}, want: lab.ErrUnresolved},
		{name: "denied-delete", entry: lab.Entry{Fixture: fixture, ResourceID: identity}, remove: true, steps: []testutil.Exchange{wire("GET", 200, owned), wire("DELETE", 403, nil)}, want: auth.ErrForbidden},
		{name: "denied-delete-confirmation", entry: lab.Entry{Fixture: fixture, ResourceID: identity}, remove: true, steps: []testutil.Exchange{wire("GET", 200, owned), wire("DELETE", 204, nil), wire("GET", 403, nil)}, want: auth.ErrForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			result, err := testutil.Run(t.Context(), testutil.Case{Name: tc.name, Steps: tc.steps, Want: testutil.Expected{Error: tc.want}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				h := RecoveryHandlers(sdk)["recording-message"]
				if tc.remove {
					return testutil.Observation{Err: h.Delete(ctx, tc.entry)}
				}
				resolved, err := h.Resolve(ctx, tc.entry)
				if err == nil && resolved != "" {
					return testutil.Observation{Err: errors.New("unexpected resource resolution")}
				}
				return testutil.Observation{Err: err}
			}})
			if err != nil || !result.Passed() {
				t.Fatalf("cleanup accepted unverified ownership or absence: %v %+v", err, result.Differences())
			}
		})
	}
}
