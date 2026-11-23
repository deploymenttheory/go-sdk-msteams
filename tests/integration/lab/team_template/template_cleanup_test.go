package teamtemplateacc

import (
	"context"
	"encoding/json"
	"errors"
	"net/url"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

func TestUnitTemplateCleanupRequiresOwnedReferencesAndListAbsence(t *testing.T) {
	t.Parallel()
	const name = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-template"
	for _, entry := range []lab.Entry{
		{Fixture: "Global", ResourceID: "/api/teamtemplates/v1.0/fixture/Tenant/en-US"},
		{Fixture: name, ResourceID: "/api/teamtemplates/v1.0/fixture/Public/en-US"},
		{Fixture: name, ResourceID: "https://foreign.example/fixture"},
		{Fixture: name, ResourceID: "/api/teamtemplates/v1.0/fixture/Tenant/en-GB"},
	} {
		if !errors.Is((templateCleanup{}).Delete(t.Context(), entry), lab.ErrUnresolved) {
			t.Fatal("template cleanup accepted a foreign selector")
		}
	}
	root, err := os.OpenRoot(labkit.Root + "/msteams/teams_admin_api")
	if err != nil {
		t.Fatal(err)
	}
	t.Cleanup(func() {
		if err := root.Close(); err != nil {
			t.Error(err)
		}
	})
	read := func(path string) []byte {
		t.Helper()
		raw, err := root.ReadFile(path)
		if err != nil {
			t.Fatal(err)
		}
		return raw
	}
	created := read("team_template/testdata/expanded_create_201.json")
	var receipt struct {
		Reference string `json:"@odata.id"`
	}
	if json.Unmarshal(created, &receipt) != nil || receipt.Reference == "" {
		t.Fatal("actual template selector unavailable")
	}
	missing := read("team_template/testdata/expanded_get_deleted_404.json")
	listed := read("team_template/testdata/expanded_list_deleted.json")
	discovery := read("internal/configapi/testdata/service_discovery.json")
	steps := []contracttest.Exchange{
		labkit.CleanupExchange("GET", "/Teams.Tenant/serviceDiscovery", 200, string(discovery)),
		labkit.CleanupExchange("GET", "/Teams.Templates/"+url.PathEscape(receipt.Reference), 404, string(missing)),
		labkit.CleanupExchange("GET", "/Teams.Templates/api/TeamTemplates/v1.0/en-US", 200, string(listed)),
	}
	result, err := contracttest.Run(t.Context(), contracttest.Case{Name: "template-independent-absence", Family: "team_templates_cleanup", Scope: "service", Source: "tests/integration/lab/team_template/template_cleanup_test.go", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		return contracttest.Observation{Err: (templateCleanup{sdk.TeamsAdminAPI.TeamTemplates}).absent(ctx, lab.Entry{Fixture: "fixture-template", ResourceID: receipt.Reference})}
	}})
	if err != nil || !result.Passed() {
		t.Fatalf("template cleanup omitted selected/list absence: %v %+v", err, result.Differences())
	}
}
