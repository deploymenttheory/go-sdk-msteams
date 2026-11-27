package teamstemplatepermissionpolicy_test

import (
	"context"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_template_permission_policy"
)

func TestUnitHiddenTemplateEditsAreLocalAndOwned(t *testing.T) {
	t.Parallel()
	a := resource.NewHiddenTemplate("a")
	b := resource.NewHiddenTemplate("b")
	upper := resource.NewHiddenTemplate("A")
	edit := resource.ChangeHiddenTemplates([]*resource.HiddenTemplate{upper}, []*resource.HiddenTemplate{a})
	values, err := edit.Apply([]*resource.HiddenTemplate{a, a, b})
	if err != nil || len(values) != 3 || *values[0].ID != "a" || *values[1].ID != "b" || *values[2].ID != "A" {
		t.Fatal("list edit lost first-match ordinal semantics")
	}
	*values[0].ID = "changed"
	*values[2].ID = "changed"
	if *a.ID != "a" || *upper.ID != "A" {
		t.Fatal("list result aliases caller data")
	}
	cleared, err := resource.ClearHiddenTemplates().Apply([]*resource.HiddenTemplate{a})
	if err != nil || cleared == nil || len(cleared) != 0 {
		t.Fatal("bound clear is not an owned empty array")
	}
	empty, err := resource.ChangeHiddenTemplates(nil, nil).Apply([]*resource.HiddenTemplate{a})
	if err != nil || len(empty) != 0 {
		t.Fatal("source empty modifier did not clear")
	}
}
func TestUnitHiddenTemplateRequestValidationPrecedesHTTP(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "hidden-template-validation", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		service := sdk.TeamsAdminAPI.TeamsTemplatePermissionPolicies
		if _, _, err := service.CreateByIdentity(ctx, "fixture-policy", resource.CreateRequest{HiddenTemplates: &resource.HiddenTemplateEdit{}}); !errors.Is(err, resource.ErrInput) {
			t.Fatal("invalid list edit acquired credentials")
		}
		_, r, e := service.PatchByIdentity(ctx, "fixture-policy", resource.PatchRequest{Settings: resource.Settings{HiddenTemplates: policy.Null[[]*resource.HiddenTemplate]()}})
		if !errors.Is(e, resource.ErrInput) {
			t.Fatal("source DTO accepted a null array setting")
		}
		return testutil.Observation{Response: r}
	}})
}
