package teamtemplate_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
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
	coverage := collectServiceCoverage(t, "New-CsTeamsHiddenTemplate", "New-CsTeamsHiddenMeetingTemplate")
	for _, tc := range []struct {
		name, command, source string
		invoke                func(context.Context, *msteams.Client) contracttest.Observation
	}{
		{name: "hidden-team-local", command: "New-CsTeamsHiddenTemplate",
			source: "contracts/powershell/8.1.0/runtime.json#/commands/362/parameter_sets/0",
			invoke: func(context.Context, *msteams.Client) contracttest.Observation {
				value, err := template.NewHiddenTemplate("fixture-template")
				return contracttest.Observation{Value: value, Err: err}
			}},
		{name: "hidden-meeting-local", command: "New-CsTeamsHiddenMeetingTemplate",
			source: "contracts/powershell/8.1.0/runtime.json#/commands/361/parameter_sets/0",
			invoke: func(context.Context, *msteams.Client) contracttest.Observation {
				value, err := template.NewHiddenMeetingTemplate("fixture-template")
				return contracttest.Observation{Value: value, Err: err}
			}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := contracttest.Case{Name: tc.name,
				Family: "team_templates", Scope: "local", Source: tc.source,
				Mappings: []contracttest.Mapping{{Command: tc.command, ParameterSet: "Identity"}},
				Invoke:   tc.invoke, Want: contracttest.Expected{JSON: json.RawMessage(`{"ID":"fixture-template"}`)}}
			result, err := contracttest.Run(t.Context(), c)
			if err != nil {
				t.Fatal(err)
			}
			if !result.Passed() {
				t.Fatalf("local builder contract differed: %+v", result.Differences())
			}
			coverage.add(c)
		})
	}
}
