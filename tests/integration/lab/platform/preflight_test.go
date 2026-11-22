package platformacc

import (
	"encoding/json"
	"errors"
	"io"
	"os"
	"runtime"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

// TestAccPreflight stays discoverable in ordinary go test and explicitly skips
// before loading credentials. All requests are the harness's two read-only probes.
func TestAccPreflight(t *testing.T) {
	cfg, err := lab.Load()
	if err != nil {
		t.Fatal(err)
	}
	if err := cfg.Require(lab.ReadOnly); errors.Is(err, lab.ErrDisabled) {
		t.Skip("lab preflight is opt-in: MSTEAMS_LAB_READ_ONLY=1")
	} else if err != nil {
		t.Fatal(err)
	}
	expectedUsersOutcome := "forbidden"
	switch os.Getenv("MSTEAMS_LAB_ACCESS_PROFILE") {
	case "", "organization-only":
	case "teams-administrator":
		expectedUsersOutcome = "available"
	default:
		t.Fatal("unsupported explicit lab access profile")
	}
	file := os.Getenv("MSTEAMS_LAB_IDENTITY_FILE")
	if file == "" {
		t.Fatal("explicit private lab identity file is required")
	}
	//nolint:gosec // G304: operator-selected private lab file, only read after explicit live opt-in.
	stream, err := os.Open(file)
	if err != nil {
		t.Fatal("private lab identity unavailable")
	}
	defer func() { _ = stream.Close() }()
	info, err := stream.Stat()
	if err != nil || !info.Mode().IsRegular() || info.Size() > 1<<20 || runtime.GOOS != "windows" && info.Mode().Perm()&0o077 != 0 {
		t.Fatal("lab identity requires a private regular file")
	}
	raw, err := io.ReadAll(io.LimitReader(stream, (1<<20)+1))
	if err != nil || len(raw) > 1<<20 {
		t.Fatal("lab identity could not be read")
	}
	defer clear(raw)
	var identity struct {
		Tenant string `json:"tenant_id"`
		Client string `json:"client_id"`
		Secret string `json:"client_secret"`
	}
	if json.Unmarshal(raw, &identity) != nil {
		t.Fatal("invalid lab identity")
	}
	sdk, err := msteams.NewClient(&auth.Config{Method: auth.ClientSecret, TenantID: identity.Tenant, ClientID: identity.Client, ClientSecret: identity.Secret})
	identity.Secret = ""
	if err != nil {
		t.Fatal("lab client configuration failed")
	}
	t.Cleanup(func() {
		if err := sdk.Close(); err != nil {
			t.Error("lab client cleanup failed")
		}
	})
	observations, err := lab.Preflight(t.Context(), cfg, sdk.GraphAPI.Scope())
	if err != nil {
		t.Fatal("lab preflight interrupted")
	}
	for _, observation := range observations {
		t.Logf("probe=%s status=%d outcome=%s", observation.Probe, observation.Status, observation.Outcome)
	}
	// Select the expected profile explicitly after an approved role change. Never
	// infer a new expected outcome from a response or automatically change grants.
	if len(observations) != 2 || observations[0].Outcome != "available" || observations[1].Outcome != expectedUsersOutcome {
		t.Fatal("dedicated lab app access inventory differs from its explicit profile")
	}
}
