package teamseducationconfiguration_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_education_configuration"
)

func projection(v resource.Configuration) map[string]*string {
	return map[string]*string{"Identity": v.Identity, "ParentGuardianPreferredContactMethod": v.ParentGuardianPreferredContactMethod, "EduGenerativeAIEnhancements": v.EduGenerativeAIEnhancements, "UpdateParentInformation": v.UpdateParentInformation}
}

func TestUnitEducationConfigurationModelsMatchTypedSetters(t *testing.T) {
	t.Parallel()

	fixture := struct {
		Models []struct {
			Name, Input string
			Observation struct {
				Failure *string
				Value   map[string]*string
			}
		}
	}{
		Models: []struct {
			Name, Input string
			Observation struct {
				Failure *string
				Value   map[string]*string
			}
		}{

			{
				Name:  "constructor",
				Input: "constructor",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"EduGenerativeAIEnhancements":          nil,
						"Identity":                             nil,
						"ParentGuardianPreferredContactMethod": nil,
						"UpdateParentInformation":              nil,
					},
				},
			},

			{
				Name:  "empty-object",
				Input: "{}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"EduGenerativeAIEnhancements":          nil,
						"Identity":                             nil,
						"ParentGuardianPreferredContactMethod": nil,
						"UpdateParentInformation":              nil,
					},
				},
			},

			{
				Name:  "all-settings",
				Input: "{\"ParentGuardianPreferredContactMethod\":\"FutureContact\",\"EduGenerativeAIEnhancements\":\"FutureEnhancement\",\"UpdateParentInformation\":\"FutureUpdate\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"EduGenerativeAIEnhancements":          new("FutureEnhancement"),
						"Identity":                             nil,
						"ParentGuardianPreferredContactMethod": new("FutureContact"),
						"UpdateParentInformation":              new("FutureUpdate"),
					},
				},
			},

			{
				Name:  "ParentGuardianPreferredContactMethod-null",
				Input: "{\"ParentGuardianPreferredContactMethod\":null}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"EduGenerativeAIEnhancements":          nil,
						"Identity":                             nil,
						"ParentGuardianPreferredContactMethod": nil,
						"UpdateParentInformation":              nil,
					},
				},
			},

			{
				Name:  "ParentGuardianPreferredContactMethod-empty",
				Input: "{\"ParentGuardianPreferredContactMethod\":\"\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"EduGenerativeAIEnhancements":          nil,
						"Identity":                             nil,
						"ParentGuardianPreferredContactMethod": new(""),
						"UpdateParentInformation":              nil,
					},
				},
			},

			{
				Name:  "ParentGuardianPreferredContactMethod-future",
				Input: "{\"ParentGuardianPreferredContactMethod\":\"Future-ParentGuardianPreferredContactMethod\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"EduGenerativeAIEnhancements":          nil,
						"Identity":                             nil,
						"ParentGuardianPreferredContactMethod": new("Future-ParentGuardianPreferredContactMethod"),
						"UpdateParentInformation":              nil,
					},
				},
			},

			{
				Name:  "EduGenerativeAIEnhancements-null",
				Input: "{\"EduGenerativeAIEnhancements\":null}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"EduGenerativeAIEnhancements":          nil,
						"Identity":                             nil,
						"ParentGuardianPreferredContactMethod": nil,
						"UpdateParentInformation":              nil,
					},
				},
			},

			{
				Name:  "EduGenerativeAIEnhancements-empty",
				Input: "{\"EduGenerativeAIEnhancements\":\"\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"EduGenerativeAIEnhancements":          new(""),
						"Identity":                             nil,
						"ParentGuardianPreferredContactMethod": nil,
						"UpdateParentInformation":              nil,
					},
				},
			},

			{
				Name:  "EduGenerativeAIEnhancements-future",
				Input: "{\"EduGenerativeAIEnhancements\":\"Future-EduGenerativeAIEnhancements\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"EduGenerativeAIEnhancements":          new("Future-EduGenerativeAIEnhancements"),
						"Identity":                             nil,
						"ParentGuardianPreferredContactMethod": nil,
						"UpdateParentInformation":              nil,
					},
				},
			},

			{
				Name:  "UpdateParentInformation-null",
				Input: "{\"UpdateParentInformation\":null}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"EduGenerativeAIEnhancements":          nil,
						"Identity":                             nil,
						"ParentGuardianPreferredContactMethod": nil,
						"UpdateParentInformation":              nil,
					},
				},
			},

			{
				Name:  "UpdateParentInformation-empty",
				Input: "{\"UpdateParentInformation\":\"\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"EduGenerativeAIEnhancements":          nil,
						"Identity":                             nil,
						"ParentGuardianPreferredContactMethod": nil,
						"UpdateParentInformation":              new(""),
					},
				},
			},

			{
				Name:  "UpdateParentInformation-future",
				Input: "{\"UpdateParentInformation\":\"Future-UpdateParentInformation\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"EduGenerativeAIEnhancements":          nil,
						"Identity":                             nil,
						"ParentGuardianPreferredContactMethod": nil,
						"UpdateParentInformation":              new("Future-UpdateParentInformation"),
					},
				},
			},

			{
				Name:  "long-contact",
				Input: "{\"ParentGuardianPreferredContactMethod\":\"" + strings.Repeat("x", 2000) + "\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"EduGenerativeAIEnhancements":          nil,
						"Identity":                             nil,
						"ParentGuardianPreferredContactMethod": new("" + strings.Repeat("x", 2000)),
						"UpdateParentInformation":              nil,
					},
				},
			},

			{
				Name:  "case-duplicate",
				Input: "{\"UpdateParentInformation\":\"first\",\"updateparentinformation\":\"second\"}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"EduGenerativeAIEnhancements":          nil,
						"Identity":                             nil,
						"ParentGuardianPreferredContactMethod": nil,
						"UpdateParentInformation":              new("second"),
					},
				},
			},

			{
				Name:  "duplicate-null",
				Input: "{\"EduGenerativeAIEnhancements\":\"earlier\",\"edugenerativeaienhancements\":null}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Value: map[string]*string{
						"EduGenerativeAIEnhancements":          nil,
						"Identity":                             nil,
						"ParentGuardianPreferredContactMethod": nil,
						"UpdateParentInformation":              nil,
					},
				},
			},

			{
				Name:  "invalid-string-array",
				Input: "{\"ParentGuardianPreferredContactMethod\":[]}",
				Observation: struct {
					Failure *string
					Value   map[string]*string
				}{
					Failure: new("Newtonsoft.Json.JsonReaderException"),
				},
			},
		},
	}

	for _, c := range fixture.Models {
		t.Run(c.Name, func(t *testing.T) {
			var v resource.Configuration
			var err error
			if c.Input != "constructor" {
				err = json.Unmarshal([]byte(c.Input), &v)
			}
			if c.Observation.Failure != nil {
				if !errors.Is(err, client.ErrDecode) {
					t.Fatal("source setter failure accepted")
				}
				return
			}
			if err != nil || !reflect.DeepEqual(projection(v), c.Observation.Value) {
				t.Fatal("known model projection differs")
			}
		})
	}
}

func TestUnitEducationConfigurationSnapshotOwnershipAndRedaction(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"UpdateParentInformation":"fixture-private","Future":9007199254740993}`)
	var v resource.Configuration
	if json.Unmarshal(raw, &v) != nil {
		t.Fatal("opaque source setting rejected")
	}
	clear(raw)
	if !strings.Contains(string(v.RawJSON), "9007199254740993") || string(v.RawFields["Future"]) != "9007199254740993" {
		t.Fatal("owned future snapshot lost")
	}
	encoded, e := json.Marshal(v)
	if e != nil || !strings.Contains(string(encoded), "fixture-private") {
		t.Fatal("caller-owned original snapshot lost")
	}
	for _, format := range []string{"%v", "%+v", "%#v"} {
		if strings.Contains(fmt.Sprintf(format, v), "fixture-private") {
			t.Fatal("ordinary configuration formatting exposed content")
		}
	}
}
