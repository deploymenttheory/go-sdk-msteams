package testutil

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams"
)

func TestUnitRunnerFormattingRedactsObservationContent(t *testing.T) {
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
	if strings.Contains(fmt.Sprintf("%+v %#v %+v", c, passed, Observation{Value: "private-token-canary"}), "private-token-canary") {
		t.Fatal("formatted runner values leaked observation content")
	}
}

func TestUnitRunnerRejectsDifferentBinaryPayload(t *testing.T) {
	t.Parallel()
	c := binaryCase([]byte("private-binary-canary\x00\xff"))
	requirePass(t, c)
	c.Steps[0].Request.Binary = []byte("private-expected-canary")
	failed, err := Run(t.Context(), c)
	if err != nil || failed.Passed() {
		t.Fatal("binary difference did not fail")
	}
}

func TestUnitRunDoesNotReadEnvironmentCredentials(t *testing.T) {
	// Environment changes are process-scoped. The callback invokes the real SDK
	// scope, whose provider must retain the explicit fixture binding.
	t.Setenv("MSTEAMS_AUTH_METHOD", "not-a-method")
	t.Setenv("MSTEAMS_TENANT_ID", "private-env-tenant-canary")
	t.Setenv("MSTEAMS_CLIENT_SECRET", "private-env-secret-canary")
	requirePass(t, graphCase())
}

func TestUnitRunLocalCaseRejectsCredentialAcquisition(t *testing.T) {
	t.Parallel()
	c := graphCase()
	c.Name, c.Steps = "invalid-local-auth", []Exchange{}
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
