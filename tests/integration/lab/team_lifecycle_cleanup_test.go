package lab_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

const lifecycleTestID = "11111111-1111-4111-8111-111111111111"
const lifecycleOtherID = "22222222-2222-4222-8222-222222222222"
const lifecycleTestName = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-team"

func lifecycleWire(method, path string, status int, body string) contracttest.Exchange {
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method,
		URL: auth.GraphAudience + path, Audience: auth.GraphAudience, Headers: http.Header{"Accept": {"application/json"}}},
		Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(body)}}
}

func lifecycleOffline(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family, c.Source = "team_lifecycle_lab", "contracts/endpoints/8.1.0/fixtures/core-team-compositions.json"
	if c.Scope == "" {
		c.Scope = "service"
	}
	result, err := contracttest.Run(t.Context(), c)
	if err != nil || !result.Passed() {
		t.Fatalf("lifecycle recovery contract differs: %+v", result.Differences())
	}
}

func TestUnitLifecycleResolveRequiresCompleteUniquePositiveMatch(t *testing.T) {
	t.Parallel()
	row := `{"id":"` + lifecycleTestID + `","displayName":"` + lifecycleTestName + `","mailNickname":"` + lifecycleTestName + `"}`
	query := url.Values{"$filter": {"displayName eq '" + lifecycleTestName + "'"}, "$select": {"id,displayName,mailNickname"}}.Encode()
	for _, tc := range []struct {
		name, body, id string
		status         int
		want           error
	}{
		{name: "unique", status: 200, body: `{"value":[` + row + `]}`, id: lifecycleTestID},
		{name: "absent-unresolved", status: 200, body: `{"value":[]}`, want: lab.ErrUnresolved},
		{name: "ambiguous", status: 200, body: `{"value":[` + row + `,` + strings.ReplaceAll(row, lifecycleTestID, lifecycleOtherID) + `]}`, want: lab.ErrUnresolved},
		{name: "wrong-alias", status: 200, body: `{"value":[{"id":"` + lifecycleTestID + `","displayName":"` + lifecycleTestName + `","mailNickname":"unowned"}]}`, want: lab.ErrUnresolved},
		{name: "denied", status: 403, body: `{"error":{"code":"Forbidden"}}`, want: auth.ErrForbidden},
		{name: "incomplete", status: 200, body: `{"value":[` + row + `],"@odata.nextLink":"https://foreign.invalid/private"}`, want: client.ErrDestination},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			lifecycleOffline(t, contracttest.Case{Name: tc.name, Steps: []contracttest.Exchange{lifecycleWire("GET", "/v1.0/groups?"+query, tc.status, tc.body)},
				Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					id, err := (lifecycleCleanup{sdk: sdk}).Resolve(ctx, lab.Entry{Fixture: lifecycleTestName})
					return contracttest.Observation{Value: id, Err: err}
				}, Want: contracttest.Expected{Error: tc.want, JSON: json.RawMessage(`"` + tc.id + `"`)}})
		})
	}
}

func TestUnitLifecycleDeletionRequiresOwnedPairAndNeverReplays(t *testing.T) {
	t.Parallel()
	groupBody := `{"id":"` + lifecycleTestID + `","displayName":"` + lifecycleTestName + `","mailNickname":"` + lifecycleTestName + `","groupTypes":["Unified"]}`
	teamBody := `{"id":"` + lifecycleTestID + `","displayName":"` + lifecycleTestName + `"}`
	for _, tc := range []struct {
		name, group, team string
		status            int
		deleteStatus      int
		want              error
	}{
		{name: "owned", group: groupBody, team: teamBody, status: 200, deleteStatus: 204},
		{name: "wrong-group-name", group: strings.ReplaceAll(groupBody, `"displayName":"`+lifecycleTestName+`"`, `"displayName":"unowned"`), team: teamBody, status: 200, want: lab.ErrUnresolved},
		{name: "wrong-team-name", group: groupBody, team: strings.ReplaceAll(teamBody, lifecycleTestName, "unowned"), status: 200, want: lab.ErrUnresolved},
		{name: "not-unified", group: strings.ReplaceAll(groupBody, `"Unified"`, `"Security"`), team: teamBody, status: 200, want: lab.ErrUnresolved},
		{name: "group-denied", group: `{"error":{"code":"Forbidden"}}`, status: 403, want: auth.ErrForbidden},
		{name: "delete-denied", group: groupBody, team: teamBody, status: 200, deleteStatus: 403, want: auth.ErrForbidden},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			steps := []contracttest.Exchange{lifecycleWire("GET", "/v1.0/groups/"+lifecycleTestID, tc.status, tc.group)}
			if tc.team != "" {
				steps = append(steps, lifecycleWire("GET", "/v1.0/teams/"+lifecycleTestID, 200, tc.team))
			}
			if tc.deleteStatus != 0 {
				steps = append(steps, lifecycleWire("DELETE", "/v1.0/groups/"+lifecycleTestID, tc.deleteStatus, ""))
			}
			lifecycleOffline(t, contracttest.Case{Name: tc.name, Steps: steps,
				Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					return contracttest.Observation{Err: (lifecycleCleanup{sdk: sdk}).Delete(ctx, lab.Entry{ResourceID: lifecycleTestID, Fixture: lifecycleTestName})}
				}, Want: contracttest.Expected{Error: tc.want}})
		})
	}
	for _, name := range []string{"invalid-id", "invalid-fixture"} {
		t.Run(name, func(t *testing.T) {
			lifecycleOffline(t, contracttest.Case{Name: name, Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				entry := lab.Entry{ResourceID: lifecycleTestID, Fixture: lifecycleTestName}
				if name == "invalid-id" {
					entry.ResourceID = "../unowned"
				} else {
					entry.Fixture = "unowned' or true"
				}
				return contracttest.Observation{Err: (lifecycleCleanup{sdk: sdk}).Delete(ctx, entry)}
			}, Want: contracttest.Expected{Error: lab.ErrUnresolved}})
		})
	}
}

func TestUnitLifecycleOwnerRequiresPrivateSameTenantFixture(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"valid", "other-tenant", "invalid-user", "missing-file"} {
		t.Run(name, func(t *testing.T) {
			t.Parallel()
			value := lifecycleOwner{Tenant: lifecycleTestID, UserID: lifecycleOtherID}
			switch name {
			case "other-tenant":
				value.Tenant = lifecycleOtherID
			case "invalid-user":
				value.UserID = "../unowned"
			}
			file := filepath.Join(t.TempDir(), "owner.json")
			data, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			if name != "missing-file" {
				if err := os.WriteFile(file, data, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			_, err = loadLifecycleOwner(file, lifecycleTestID+"|commercial|fixture-client")
			if (err == nil) != (name == "valid") {
				t.Fatal("owner fixture validation differs")
			}
		})
	}
}

func TestUnitLifecycleDeletionObservationRequiresBoth404sAndRejectsDenial(t *testing.T) {
	for _, denied := range []bool{false, true} {
		name := "both-missing"
		if denied {
			name = "team-denied"
		}
		t.Run(name, func(t *testing.T) {
			synctest.Test(t, func(t *testing.T) {
				steps := []contracttest.Exchange{lifecycleWire("GET", "/v1.0/groups/"+lifecycleTestID, 200, `{"id":"`+lifecycleTestID+`"}`), lifecycleWire("GET", "/v1.0/teams/"+lifecycleTestID, 200, `{"id":"`+lifecycleTestID+`"}`), lifecycleWire("GET", "/v1.0/groups/"+lifecycleTestID, 404, "")}
				status := 404
				var want error
				if denied {
					status, want = 403, auth.ErrForbidden
				}
				steps = append(steps, lifecycleWire("GET", "/v1.0/teams/"+lifecycleTestID, status, ""))
				started := time.Now()
				lifecycleOffline(t, contracttest.Case{Name: name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
					defer cancel()
					return contracttest.Observation{Err: lifecycleObserveAbsent(ctx, sdk, lifecycleTestID)}
				}, Want: contracttest.Expected{Error: want}})
				if time.Since(started) != 2*time.Second {
					t.Fatal("observation did not use bounded read-only cadence")
				}
			})
		})
	}
}
