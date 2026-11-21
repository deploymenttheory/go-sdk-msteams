package teamseducationassignmentsapppolicy_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"os"
	"reflect"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_education_assignments_app_policy"
)

func projection(v resource.Policy) map[string]*string {
	return map[string]*string{"Identity": v.Identity, "ParentDigestEnabledType": v.ParentDigestEnabledType, "MakeCodeEnabledType": v.MakeCodeEnabledType, "TurnItInEnabledType": v.TurnItInEnabledType, "TurnItInApiUrl": v.TurnItInAPIURL, "TurnItInApiKey": v.TurnItInAPIKey}
}

func TestEducationAssignmentsModelsMatchPinnedSetters(t *testing.T) {
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
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Models) != 25 {
		t.Fatal("incomplete source models")
	}
	for _, c := range fixture.Models {
		t.Run(c.Name, func(t *testing.T) {
			var v resource.Policy
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

func TestEducationAssignmentsSnapshotOwnershipAndRedaction(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"TurnItInApiKey":"fixture-private","Future":9007199254740993,"TurnItInApiUrl":"https:// "}`)
	var v resource.Policy
	if json.Unmarshal(raw, &v) != nil {
		t.Fatal("source-compatible opaque URL rejected")
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
			t.Fatal("ordinary policy formatting exposed key")
		}
	}
	for _, bad := range []string{`{"TurnItInApiUrl":"bad","turnitinapiurl":"https://fixture.invalid"}`, `{"TurnItInApiUrl":"https://x\n\n"}`, `{"TurnItInApiUrl":"httpſ://fixture.invalid"}`} {
		if !errors.Is(json.Unmarshal([]byte(bad), &v), client.ErrDecode) {
			t.Fatal("source-invalid URL accepted")
		}
	}
}
