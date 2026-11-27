package teamtemplate_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	template "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/team_template"
)

func TestUnitNewHiddenTemplate(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, id string
	}{
		{name: "empty"},
		{name: "opaque", id: "fixture-template"},
		{name: "whitespace is not invented validation", id: " "},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			value, err := template.NewHiddenTemplate(tc.id)
			if tc.id == "" {
				if !errors.Is(err, template.ErrInput) {
					t.Fatal("missing required ID was accepted")
				}
				return
			}
			if err != nil || value.ID != tc.id {
				t.Fatal("local template ID was changed")
			}
		})
	}
}

func TestUnitNewHiddenMeetingTemplate(t *testing.T) {
	t.Parallel()
	if _, err := template.NewHiddenMeetingTemplate(""); !errors.Is(err, template.ErrInput) {
		t.Fatal("missing required meeting-template ID was accepted")
	}
	value, err := template.NewHiddenMeetingTemplate("fixture-meeting-template")
	if err != nil || value.ID != "fixture-meeting-template" {
		t.Fatal("local meeting-template ID was changed")
	}
}

func TestUnitHiddenBuildersHaveNoManagementOrAuthIO(t *testing.T) {
	t.Parallel()

	for _, tc := range []struct {
		name   string
		invoke func(context.Context, *msteams.Client) testutil.Observation
	}{
		{name: "hidden-team-local",
			invoke: func(context.Context, *msteams.Client) testutil.Observation {
				value, err := template.NewHiddenTemplate("fixture-template")
				return testutil.Observation{Value: value, Err: err}
			}},
		{name: "hidden-meeting-local",
			invoke: func(context.Context, *msteams.Client) testutil.Observation {
				value, err := template.NewHiddenMeetingTemplate("fixture-template")
				return testutil.Observation{Value: value, Err: err}
			}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := testutil.Case{Name: tc.name,

				Invoke: tc.invoke, Want: testutil.Expected{JSON: json.RawMessage(`{"ID":"fixture-template"}`)}}
			result, err := testutil.Run(t.Context(), c)
			if err != nil {
				t.Fatal(err)
			}
			if !result.Passed() {
				t.Fatalf("local builder contract differed: %+v", result.Differences())
			}

		})
	}
}
