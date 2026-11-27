package teamsfirstpartymeetingtemplateconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_first_party_meeting_template_configuration"
)

func names(values []*resource.Template) []string {
	out := make([]string, len(values))
	for i, v := range values {
		out[i] = *v.Name
	}
	return out
}

func TestUnitTemplateEditsAreLocalOrdinalAndOwned(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "template-edits", Invoke: func(context.Context, *msteams.Client) testutil.Observation {
		a, b, upper := resource.NewTemplate("a"), resource.NewTemplate("b"), resource.NewTemplate("A")
		a.TeamsMeetingOptions = append(a.TeamsMeetingOptions, resource.NewOption("AutoAdmittedUsers"))
		values, err := resource.ChangeTemplates([]*resource.Template{upper}, []*resource.Template{resource.NewTemplate("a")}).Apply([]*resource.Template{a, a, b})
		if err != nil || len(values) != 3 || names(values)[0] != "a" || names(values)[1] != "b" || names(values)[2] != "A" {
			t.Fatal("template edit lost first ordinal Name-match semantics")
		}
		*values[0].Name = "changed"
		*values[0].TeamsMeetingOptions[0].Name = "changed"
		*values[2].Name = "changed"
		if *a.Name != "a" || *a.TeamsMeetingOptions[0].Name != "AutoAdmittedUsers" || *upper.Name != "A" {
			t.Fatal("template edit result aliases caller data")
		}
		missing, err := resource.ChangeTemplates(nil, []*resource.Template{resource.NewTemplate("absent")}).Apply([]*resource.Template{a})
		if err != nil || len(missing) != 1 {
			t.Fatal("unmatched Remove must leave the list unchanged")
		}
		cleared, err := resource.ClearTemplates().Apply([]*resource.Template{a})
		if err != nil || cleared == nil || len(cleared) != 0 {
			t.Fatal("bound clear is not an owned empty array")
		}
		empty, err := resource.ChangeTemplates(nil, nil).Apply([]*resource.Template{a})
		if err != nil || len(empty) != 0 {
			t.Fatal("source empty modifier did not clear")
		}
		for _, bad := range []func() error{
			func() error { _, e := (&resource.TemplateEdit{}).Apply(nil); return e },
			func() error { _, e := (*resource.TemplateEdit)(nil).Apply(nil); return e },
			func() error {
				_, e := resource.ChangeTemplates(nil, []*resource.Template{nil}).Apply([]*resource.Template{a})
				return e
			},
			func() error {
				_, e := resource.ChangeTemplates(nil, []*resource.Template{a}).Apply([]*resource.Template{nil})
				return e
			},
		} {
			if !errors.Is(bad(), resource.ErrInput) {
				t.Fatal("invalid template edit accepted")
			}
		}
		raw, err := json.Marshal(resource.NewTemplate("x"))
		if err != nil || !jsonEqual(raw, []byte(`{"Identity":null,"Description":null,"Name":"x","Category":null,"TeamsMeetingOptions":[]}`)) {
			t.Fatal("local constructor differs from the pinned name constructor")
		}
		return testutil.Observation{}
	}})
}
