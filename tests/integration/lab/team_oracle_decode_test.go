package lab_test

import (
	"errors"
	"strings"
	"testing"
)

func TestTeamOracleRejectsIncompleteOrPollutedReports(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, kind, body string
		valid            bool
	}{
		{name: "team", kind: "team", body: `{"success":true,"stage":"complete","team":{"id":"fixture","archived":false},"tenant":{"id":"fixture"},"user":{"id":"fixture"}}`, valid: true},
		{name: "empty-user-collection", kind: "users", body: `{"success":true,"stage":"complete","users":[]}`, valid: true},
		{name: "channel-users", kind: "channel-users", body: `{"success":true,"stage":"complete","users":[]}`, valid: true},
		{name: "batch-status", kind: "batch-status", body: `{"success":true,"stage":"complete","batch_status":"Completed; opaque text"}`, valid: true},
		{name: "empty-batch-status", kind: "batch-status", body: `{"success":true,"stage":"complete","batch_status":""}`, valid: true},
		{name: "missing-batch-status", kind: "batch-status", body: `{"success":true,"stage":"complete"}`},
		{name: "private-migration", kind: "private-migration", body: `{"success":true,"stage":"complete","private_migration":{"tenant_id":"fixture","status":"Future","details":null}}`, valid: true},
		{name: "mixed-batch-status", kind: "batch-status", body: `{"success":true,"stage":"complete","batch_status":"raw","users":[]}`},
		{name: "unexpected-private-migration", kind: "hierarchy-status", body: `{"success":true,"stage":"complete","private_migration":{},"hierarchy_status":{"id":"fixture"}}`},
		{name: "targeting-policy", kind: "targeting-policy", body: `{"success":true,"stage":"complete","targeting_policy":{"identity":"Global","description":""}}`, valid: true},
		{name: "hierarchy-status", kind: "hierarchy-status", body: `{"success":true,"stage":"complete","hierarchy_status":{"id":"fixture","status":"FutureStatus"}}`, valid: true},
		{name: "missing-targeting", kind: "targeting-policy", body: `{"success":true,"stage":"complete"}`},
		{name: "mixed-targeting", kind: "targeting-policy", body: `{"success":true,"stage":"complete","targeting_policy":{"identity":"Global"},"hierarchy_status":{"id":"fixture"}}`},
		{name: "mixed-hierarchy", kind: "hierarchy-status", body: `{"success":true,"stage":"complete","hierarchy_status":{"id":"fixture"},"users":[]}`},
		{name: "unexpected-hierarchy", kind: "shared-teams", body: `{"success":true,"stage":"complete","shared_teams":[],"hierarchy_status":{"id":"fixture"}}`},
		{name: "shared-users", kind: "shared-users", body: `{"success":true,"stage":"complete","users":[{"id":"fixture","role":"owner","tenant_id":"external"}]}`, valid: true},
		{name: "shared-teams", kind: "shared-teams", body: `{"success":true,"stage":"complete","shared_teams":[]}`, valid: true},
		{name: "associated-teams", kind: "associated-teams", body: `{"success":true,"stage":"complete","associated_teams":[]}`, valid: true},
		{name: "missing-shared-teams", kind: "shared-teams", body: `{"success":true,"stage":"complete"}`},
		{name: "mixed-sharing", kind: "shared-teams", body: `{"success":true,"stage":"complete","shared_teams":[],"associated_teams":[]}`},
		{name: "mixed-sharing-users", kind: "shared-users", body: `{"success":true,"stage":"complete","shared_teams":[],"users":[]}`},
		{name: "unexpected-sharing", kind: "team", body: `{"success":true,"stage":"complete","shared_teams":[],"team":{},"tenant":{},"user":{}}`},
		{name: "operation", kind: "operation", body: `{"success":true,"stage":"complete","operation":{"id":"fixture","attempts":0}}`, valid: true},
		{name: "missing-operation", kind: "operation", body: `{"success":true,"stage":"complete"}`},
		{name: "mixed-operation-users", kind: "operation", body: `{"success":true,"stage":"complete","operation":{"id":"fixture"},"users":[]}`},
		{name: "unexpected-operation", kind: "users", body: `{"success":true,"stage":"complete","operation":{"id":"fixture"},"users":[]}`},
		{name: "missing-channel-users", kind: "channel-users", body: `{"success":true,"stage":"complete"}`},
		{name: "nullable-user-fields", kind: "users", body: `{"success":true,"stage":"complete","users":[{"id":"fixture","user":null,"name":null,"role":"owner"}]}`, valid: true},
		{name: "unknown-output", kind: "team", body: `{"success":true,"stage":"complete","private-token-canary":"secret"}`},
		{name: "failed-command", kind: "team", body: `{"success":false,"stage":"read"}`},
		{name: "partial-success", kind: "team", body: `{"success":true,"stage":"read","team":{},"tenant":{},"user":{}}`},
		{name: "missing-collection", kind: "users", body: `{"success":true,"stage":"complete"}`},
		{name: "missing-user-list", kind: "team", body: `{"success":true,"stage":"complete","team":{},"tenant":{}}`},
		{name: "mixed-report", kind: "team", body: `{"success":true,"stage":"complete","team":{},"tenant":{},"user":{},"users":[]}`},
		{name: "trailing-json", kind: "users", body: `{"success":true,"stage":"complete","users":[]} {"success":true}`},
		{name: "banner", kind: "users", body: `private-banner-canary {"success":true,"stage":"complete","users":[]}`},
		{name: "wrong-bool-type", kind: "team", body: `{"success":true,"stage":"complete","team":{"archived":"false"},"tenant":{},"user":{}}`},
		{name: "wrong-kind", kind: "unknown", body: `{"success":true,"stage":"complete","users":[]}`},
		{name: "oversize", kind: "users", body: strings.Repeat("x", maxTeamOracleBytes+1)},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			value, err := decodeTeamOracle([]byte(tc.body), tc.kind)
			if (err == nil) != tc.valid {
				t.Fatal("oracle report validity differs")
			}
			if err != nil && (!errors.Is(err, errTeamOracle) || strings.Contains(err.Error(), "canary")) {
				t.Fatal("oracle error leaked uncontrolled output")
			}
			if tc.name == "team" && (value.Team.Archived == nil || *value.Team.Archived || value.Team.AllowGiphy != nil) {
				t.Fatal("false or unknown team fields were collapsed")
			}
			if tc.name == "nullable-user-fields" && (value.Users[0].User != nil || value.Users[0].Name != nil) {
				t.Fatal("limited user fields were collapsed into empty strings")
			}
		})
	}
}

func TestTeamOracleOutputIsBoundedAcrossWrites(t *testing.T) {
	t.Parallel()
	var buffer teamOracleBuffer
	first := []byte(strings.Repeat("a", maxTeamOracleBytes-1))
	if n, err := buffer.Write(first); err != nil || n != len(first) {
		t.Fatal("bounded output was rejected")
	}
	if n, err := buffer.Write([]byte("b")); err != nil || n != 1 {
		t.Fatal("exact bound was rejected")
	}
	if _, err := buffer.Write([]byte("private-overflow-canary")); !errors.Is(err, errTeamOracle) {
		t.Fatal("overflow accepted")
	}
	if len(buffer.data) != maxTeamOracleBytes || !buffer.exceeded || strings.Contains(string(buffer.data), "canary") {
		t.Fatal("overflow was retained or hidden")
	}
	if _, err := buffer.Write(nil); !errors.Is(err, errTeamOracle) {
		t.Fatal("overflow state was reset")
	}
	clear(buffer.data)
}
