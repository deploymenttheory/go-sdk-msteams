package graphacc

import (
	"context"
	"encoding/json"
	"errors"
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
	c.Family, c.Source = "team_lifecycle_lab", "tests/integration/lab/graph/team_lifecycle_cleanup_test.go"
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
			if tc.deleteStatus == 204 {
				steps = append(steps, lifecycleWire("GET", "/v1.0/groups/"+lifecycleTestID, 404, ""), lifecycleWire("GET", "/v1.0/teams/"+lifecycleTestID, 404, ""))
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

// A 204 receipt is not enough to seal recovery. If independent absence cannot
// be read, the persisted entry must remain actionable on the next recovery run.
func TestUnitCoreCleanupJournalRetainsUnverifiedDeletion(t *testing.T) {
	for _, kind := range []string{"lifecycle", "membership"} {
		t.Run(kind, func(t *testing.T) {
			journal, err := lab.NewJournal(t.TempDir(), "fixture-tenant|commercial|fixture-client")
			if err != nil {
				t.Fatal("offline recovery journal unavailable")
			}
			t.Cleanup(func() {
				if err := journal.Close(); err != nil {
					t.Error("offline recovery journal close failed")
				}
			})
			suffix, fixtureKind := "team", lifecycleFixtureKind
			if kind == "membership" {
				suffix, fixtureKind = "membership", membershipFixtureKind
			}
			fixture, err := journal.FixtureName(suffix)
			if err != nil {
				t.Fatal("offline recovery fixture unavailable")
			}
			intent, err := journal.PlanCreate(fixtureKind, fixture)
			if err != nil || journal.Applied(intent, lifecycleTestID) != nil {
				t.Fatal("offline owned identity was not journaled")
			}
			groupBody := `{"id":"` + lifecycleTestID + `","displayName":"` + fixture + `","mailNickname":"` + fixture + `","groupTypes":["Unified"]}`
			teamBody := `{"id":"` + lifecycleTestID + `","displayName":"` + fixture + `"}`
			steps := []contracttest.Exchange{}
			if kind == "lifecycle" {
				steps = append(steps,
					lifecycleWire("GET", "/v1.0/groups/"+lifecycleTestID, 200, groupBody),
					lifecycleWire("GET", "/v1.0/teams/"+lifecycleTestID, 200, teamBody),
				)
			} else {
				query := url.Values{"$select": {"id,displayName,groupTypes"}}.Encode()
				steps = append(steps,
					lifecycleWire("GET", "/v1.0/teams/"+lifecycleTestID, 200, teamBody),
					lifecycleWire("GET", "/v1.0/groups/"+lifecycleTestID+"?"+query, 200, groupBody),
				)
			}
			steps = append(steps,
				lifecycleWire("DELETE", "/v1.0/groups/"+lifecycleTestID, 204, ""),
				lifecycleWire("GET", "/v1.0/groups/"+lifecycleTestID, 404, ""),
				lifecycleWire("GET", "/v1.0/teams/"+lifecycleTestID, 403, ""),
			)
			// The next run must still execute the handler. Its initial missing
			// read is followed by independent paired absence, without re-deletion.
			initial := "/v1.0/groups/" + lifecycleTestID
			if kind == "membership" {
				initial = "/v1.0/teams/" + lifecycleTestID
			}
			steps = append(steps,
				lifecycleWire("GET", initial, 404, ""),
				lifecycleWire("GET", "/v1.0/groups/"+lifecycleTestID, 404, ""),
				lifecycleWire("GET", "/v1.0/teams/"+lifecycleTestID, 404, ""),
			)
			lifecycleOffline(t, contracttest.Case{Name: "journal-" + kind + "-absence", Steps: steps,
				Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					var handler lab.Handler = lifecycleCleanup{sdk: sdk}
					if kind == "membership" {
						handler = membershipCleanup{sdk: sdk}
					}
					handlers := map[string]lab.Handler{fixtureKind: handler}
					first, err := journal.Recover(ctx, handlers)
					if !errors.Is(err, lab.ErrCleanup) || len(first) != 1 || first[0].Outcome != "unresolved" {
						t.Error("delete receipt sealed an entry before independent absence")
					}
					second, err := journal.Recover(ctx, handlers)
					if err != nil || len(second) != 1 || second[0].Outcome != "cleaned" {
						t.Error("verified interrupted deletion could not be reconciled")
					}
					third, err := journal.Recover(ctx, handlers)
					if len(third) != 0 {
						t.Error("verified cleanup was not persisted")
					}
					return contracttest.Observation{Err: err}
				},
			})
		})
	}
}
