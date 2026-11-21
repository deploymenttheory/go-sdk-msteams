package teamseducationconfiguration_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_education_configuration"
)

func projection(v resource.Configuration) map[string]*string {
	return map[string]*string{"Identity": v.Identity, "ParentGuardianPreferredContactMethod": v.ParentGuardianPreferredContactMethod, "EduGenerativeAIEnhancements": v.EduGenerativeAIEnhancements, "UpdateParentInformation": v.UpdateParentInformation}
}

func TestEducationConfigurationModelsMatchPinnedSetters(t *testing.T) {
	t.Parallel()
	raw, e := os.ReadFile("../../../" + source)
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct {
		Models []struct {
			Name, Input string
			Observation struct {
				Failure *string
				Value   map[string]*string
			}
		}
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Models) != 16 {
		t.Fatal("incomplete source models")
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

func TestEducationConfigurationSnapshotOwnershipAndRedaction(t *testing.T) {
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
