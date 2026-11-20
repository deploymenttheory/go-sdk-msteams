package contracttest

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

func TestWriteEvidence(t *testing.T) {
	t.Parallel()
	c := graphCase()
	c.Name = "retains-unknown-fields"
	c.Steps[0].Reply.Body = []byte(`{"private-field-name":"private-value-canary","token":"private-token-canary"}`)
	c.Want.JSON = json.RawMessage(`{"private-field-name":"private-value-canary","token":"private-token-canary"}`)
	passed := requirePass(t, c)
	c.Name = "detects-unknown-fields"
	c.Want.JSON = json.RawMessage(`{}`)
	failed, err := Run(t.Context(), c)
	if err != nil || failed.Passed() {
		t.Fatal("negative fixture did not fail")
	}
	first, second := &bytes.Buffer{}, &bytes.Buffer{}
	revision := strings.Repeat("a", 40)
	if err := WriteEvidence(first, revision, []Result{passed, failed}); err != nil {
		t.Fatal(err)
	}
	if err := WriteEvidence(second, revision, []Result{failed, passed}); err != nil || !bytes.Equal(first.Bytes(), second.Bytes()) {
		t.Fatal("evidence depends on input order")
	}
	text := first.String()
	for _, secret := range []string{"private-field-name", "private-value-canary", "private-token-canary", fixtureToken("https://graph.microsoft.com"), "graph.microsoft.com", "fixture-tenant", "fixture-identity"} {
		if strings.Contains(text, secret) {
			t.Fatal("evidence retained wire, identity or payload content")
		}
	}
	var artifact struct {
		Outcome       string         `json:"outcome"`
		Revision      string         `json:"revision"`
		CoverageClaim bool           `json:"coverage_claim"`
		Live          string         `json:"live"`
		Cases         []caseEvidence `json:"cases"`
	}
	if json.Unmarshal(first.Bytes(), &artifact) != nil || artifact.Outcome != "failed" || artifact.Revision != revision || artifact.CoverageClaim || artifact.Live != "not-run" || len(artifact.Cases) != 2 {
		t.Fatal("evidence overstated qualification or lost run identity")
	}
	if artifact.Cases[0].Outcome != "failed" || artifact.Cases[1].Outcome != "passed" || artifact.Cases[0].Scope != "primitive" {
		t.Fatal("evidence lost case outcome or primitive scope")
	}
	if artifact.Cases[0].Mappings[0] != c.Mappings[0] {
		t.Fatal("command/parameter-set mapping was lost")
	}
	if strings.Contains(fmt.Sprintf("%+v %#v %+v", c, passed, Observation{Value: "private-token-canary"}), "private-token-canary") {
		t.Fatal("formatted runner values leaked observation content")
	}
}

type rejectedWriter struct{}

func (rejectedWriter) Write([]byte) (int, error) { return 0, errors.New("private-writer-diagnostic") }

func TestWriteEvidenceRejectsInvalidOrUnexecutedRuns(t *testing.T) {
	t.Parallel()
	result := requirePass(t, graphCase())
	for _, tc := range []struct {
		name     string
		revision string
		results  []Result
	}{
		{name: "missing revision", results: []Result{result}},
		{name: "dirty revision", revision: strings.Repeat("a", 40) + "-dirty", results: []Result{result}},
		{name: "no results", revision: strings.Repeat("a", 40)},
		{name: "unexecuted", revision: strings.Repeat("a", 40), results: []Result{{}}},
		{name: "duplicate", revision: strings.Repeat("a", 40), results: []Result{result, result}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			if err := WriteEvidence(&bytes.Buffer{}, tc.revision, tc.results); !errors.Is(err, ErrEvidence) {
				t.Fatal("invalid evidence accepted")
			}
		})
	}
	if err := WriteEvidence(rejectedWriter{}, strings.Repeat("a", 40), []Result{result}); !errors.Is(err, ErrEvidence) {
		t.Fatal("writer failure exposed an uncontrolled diagnostic")
	}
}

func TestRunDoesNotReadEnvironmentCredentials(t *testing.T) {
	// Deliberately non-parallel: environment changes are process-scoped. The
	// callback invokes the real SDK scope, not just a no-op constructor.
	t.Setenv("MSTEAMS_AUTH_METHOD", "not-a-method")
	t.Setenv("MSTEAMS_TENANT_ID", "private-env-tenant-canary")
	t.Setenv("MSTEAMS_CLIENT_SECRET", "private-env-secret-canary")
	result := requirePass(t, graphCase())
	var output bytes.Buffer
	if err := WriteEvidence(&output, strings.Repeat("a", 40), []Result{result}); err != nil {
		t.Fatal(err)
	}
	if strings.Contains(output.String(), "private-env-") {
		t.Fatal("ambient credentials reached evidence")
	}
}

func TestRunLocalCaseRejectsCredentialAcquisition(t *testing.T) {
	t.Parallel()
	c := graphCase()
	c.Name, c.Scope, c.Steps = "invalid-local-auth", "local", []Exchange{}
	c.Want = Expected{}
	c.Invoke = func(ctx context.Context, sdk *msteams.Client) Observation {
		_, _ = sdk.GraphAPI.Scope().Session().AcquireToken(ctx, "https://graph.microsoft.com")
		return Observation{}
	}
	result, err := Run(t.Context(), c)
	if err != nil || result.Passed() {
		t.Fatal("local helper acquired a token without failing")
	}
	for _, difference := range result.Differences() {
		if difference.Path == "tokens" {
			return
		}
	}
	t.Fatal("missing local credential acquisition difference")
}
