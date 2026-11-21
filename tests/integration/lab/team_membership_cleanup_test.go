package lab_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"os"
	"path/filepath"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

const cleanupTeamID = "11111111-1111-4111-8111-111111111111"
const cleanupOtherID = "22222222-2222-4222-8222-222222222222"
const cleanupFixture = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-membership"

func cleanupExchange(method, path string, status int, body string) contracttest.Exchange {
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: auth.GraphAudience + path, Audience: auth.GraphAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(body)}}
}

func runCleanupCase(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family = "membership_lab_cleanup"
	c.Source = "tests/integration/lab/team_membership_cleanup_test.go"
	if c.Scope == "" {
		c.Scope = "service"
	}
	result, err := contracttest.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Errorf("membership cleanup differs: %+v", result.Differences())
	}
}

func TestUnitMembershipCleanupResolveRequiresCompleteUniquePositiveMatch(t *testing.T) {
	t.Parallel()
	query := url.Values{"$filter": {"displayName eq '" + cleanupFixture + "'"}, "$select": {"id,displayName"}}.Encode()
	for _, tc := range []struct {
		name, body, id string
		status         int
		want           error
	}{
		{"unique", `{"value":[{"id":"` + cleanupTeamID + `","displayName":"` + cleanupFixture + `"}]}`, cleanupTeamID, 200, nil},
		{"absent-is-unresolved", `{"value":[]}`, "", 200, lab.ErrUnresolved},
		{"ambiguous", `{"value":[{"id":"` + cleanupTeamID + `","displayName":"` + cleanupFixture + `"},{"id":"` + cleanupOtherID + `","displayName":"` + cleanupFixture + `"}]}`, "", 200, lab.ErrUnresolved},
		{"unexpected-name", `{"value":[{"id":"` + cleanupTeamID + `","displayName":"different"}]}`, "", 200, lab.ErrUnresolved},
		{"denied", `{"error":{"code":"Forbidden"}}`, "", 403, auth.ErrForbidden},
		{"incomplete-cannot-resolve", `{"value":[{"id":"` + cleanupTeamID + `","displayName":"` + cleanupFixture + `"}],"@odata.nextLink":"https://foreign.invalid/private"}`, "", 200, client.ErrDestination},
	} {
		runCleanupCase(t, contracttest.Case{Name: tc.name, Steps: []contracttest.Exchange{cleanupExchange("GET", "/v1.0/teams?"+query, tc.status, tc.body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			id, err := (membershipCleanup{sdk: sdk}).Resolve(ctx, lab.Entry{Fixture: cleanupFixture})
			return contracttest.Observation{Value: id, Err: err}
		}, Want: contracttest.Expected{Error: tc.want, JSON: json.RawMessage(`"` + tc.id + `"`)}})
	}
	runCleanupCase(t, contracttest.Case{Name: "invalid-name-before-auth", Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, err := (membershipCleanup{sdk: sdk}).Resolve(ctx, lab.Entry{Fixture: "unowned' or true"})
		return contracttest.Observation{Err: err}
	}, Want: contracttest.Expected{Error: lab.ErrUnresolved}})
}

func TestUnitMembershipCleanupDeleteRequiresBothOwnedSnapshots(t *testing.T) {
	t.Parallel()
	groupQuery := url.Values{"$select": {"id,displayName,groupTypes"}}.Encode()
	for _, phase := range []string{"owned", "team-denied", "team-missing", "team-other-name", "team-other-id", "group-other-name", "group-not-unified", "group-other-id", "delete-denied"} {
		teamBody := `{"id":"` + cleanupTeamID + `","displayName":"` + cleanupFixture + `"}`
		groupBody := `{"id":"` + cleanupTeamID + `","displayName":"` + cleanupFixture + `","groupTypes":["Unified"]}`
		status := 200
		var want error
		switch phase {
		case "team-denied":
			status = 403
			teamBody = `{"error":{"code":"Forbidden"}}`
			want = auth.ErrForbidden
		case "team-missing":
			status = 404
			teamBody = `{"error":{"code":"Missing"}}`
		case "team-other-name":
			teamBody = `{"id":"` + cleanupTeamID + `","displayName":"unowned"}`
			want = lab.ErrUnresolved
		case "team-other-id":
			teamBody = `{"id":"` + cleanupOtherID + `","displayName":"` + cleanupFixture + `"}`
			want = lab.ErrUnresolved
		case "group-other-name":
			groupBody = `{"id":"` + cleanupTeamID + `","displayName":"unowned","groupTypes":["Unified"]}`
			want = lab.ErrUnresolved
		case "group-other-id":
			groupBody = `{"id":"` + cleanupOtherID + `","displayName":"` + cleanupFixture + `","groupTypes":["Unified"]}`
			want = lab.ErrUnresolved
		case "group-not-unified":
			groupBody = `{"id":"` + cleanupTeamID + `","displayName":"` + cleanupFixture + `","groupTypes":[]}`
			want = lab.ErrUnresolved
		case "delete-denied":
			want = auth.ErrForbidden
		}
		steps := []contracttest.Exchange{cleanupExchange("GET", "/v1.0/teams/"+cleanupTeamID, status, teamBody)}
		if phase == "owned" || phase == "delete-denied" || phase == "group-other-name" || phase == "group-not-unified" || phase == "group-other-id" {
			steps = append(steps, cleanupExchange("GET", "/v1.0/groups/"+cleanupTeamID+"?"+groupQuery, 200, groupBody))
		}
		if phase == "owned" || phase == "delete-denied" {
			deleteStatus, body := 204, ""
			if phase == "delete-denied" {
				deleteStatus = 403
				body = `{"error":{"code":"Forbidden"}}`
			}
			steps = append(steps, cleanupExchange("DELETE", "/v1.0/groups/"+cleanupTeamID, deleteStatus, body))
		}
		expected := contracttest.Expected{Error: want}
		if phase == "team-missing" {
			expected.APIError = &client.APIError{Backend: client.Graph, StatusCode: 404, Codes: []string{"Missing"}}
		}
		runCleanupCase(t, contracttest.Case{Name: phase, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			err := (membershipCleanup{sdk: sdk}).Delete(ctx, lab.Entry{ResourceID: cleanupTeamID, Fixture: cleanupFixture})
			return contracttest.Observation{Err: err}
		}, Want: expected})
	}
	for _, name := range []string{"bad-id", "bad-name"} {
		runCleanupCase(t, contracttest.Case{Name: name, Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			entry := lab.Entry{ResourceID: cleanupTeamID, Fixture: cleanupFixture}
			if name == "bad-id" {
				entry.ResourceID = "../private"
			} else {
				entry.Fixture = "unowned"
			}
			return contracttest.Observation{Err: (membershipCleanup{sdk: sdk}).Delete(ctx, entry)}
		}, Want: contracttest.Expected{Error: lab.ErrUnresolved}})
	}
}

func TestUnitMembershipUsersRequirePrivateSameTenantDistinctIDs(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"valid", "other-tenant", "reused-id", "invalid-id", "missing-file"} {
		t.Run(name, func(t *testing.T) {
			value := membershipUsers{Tenant: cleanupTeamID, Owner: cleanupTeamID, SecondOwner: cleanupOtherID, Member: "33333333-3333-4333-8333-333333333333"}
			switch name {
			case "other-tenant":
				value.Tenant = cleanupOtherID
			case "reused-id":
				value.Member = value.Owner
			case "invalid-id":
				value.Owner = "../private"
			}
			file := filepath.Join(t.TempDir(), "users.json")
			raw, err := json.Marshal(value)
			if err != nil {
				t.Fatal(err)
			}
			if name != "missing-file" {
				if err := os.WriteFile(file, raw, 0o600); err != nil {
					t.Fatal(err)
				}
			}
			_, err = loadMembershipUsers(file, cleanupTeamID+"|commercial|fixture-client")
			if (err == nil) != (name == "valid") {
				t.Fatal("membership fixture validation differs")
			}
		})
	}
}
