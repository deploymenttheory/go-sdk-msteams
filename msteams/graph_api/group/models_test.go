package group_test

import (
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
)

func TestGroupPatchMatchesPinnedOmissionAndExplicitEmpty(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/core-team-serialization.json")
	if err != nil {
		t.Fatal(err)
	}
	var source struct {
		Version string `json:"baseline_version"`
		Cases   []struct {
			Name string          `json:"name"`
			Wire json.RawMessage `json:"wire"`
		} `json:"cases"`
	}
	if json.Unmarshal(raw, &source) != nil || source.Version != "8.1.0" {
		t.Fatal("pinned group serializer capture unavailable")
	}
	empty := ""
	visibility := group.Visibility("")
	requests := map[string]group.UpdateRequest{
		"group-patch-omitted":        {},
		"group-patch-explicit-empty": {DisplayName: &empty, Description: &empty, Visibility: &visibility, MailNickname: &empty, Classification: &empty},
	}
	checked := 0
	for _, capture := range source.Cases {
		request, ok := requests[capture.Name]
		if !ok {
			continue
		}
		checked++
		encoded, err := json.Marshal(request)
		if err != nil {
			t.Fatal(err)
		}
		var actual, expected map[string]any
		if json.Unmarshal(encoded, &actual) != nil || json.Unmarshal(capture.Wire, &expected) != nil || !reflect.DeepEqual(actual, expected) {
			t.Fatal("group patch changed pinned omitted/empty intent")
		}
	}
	if checked != 2 {
		t.Fatal("pinned group patch captures disappeared")
	}
}

func TestGroupReadPreservesNullableAndFutureFields(t *testing.T) {
	t.Parallel()
	var value group.Group
	raw := `{"id":"fixture-group","displayName":"","description":null,"visibility":"futureVisibility","mailEnabled":true,"securityEnabled":false,"groupTypes":["Unified"],"resourceProvisioningOptions":[],"createdDateTime":"2026-01-02T03:04:05Z","extension_future":{"value":false}}`
	if err := json.Unmarshal([]byte(raw), &value); err != nil {
		t.Fatal(err)
	}
	if value.ID != "fixture-group" || value.DisplayName == nil || *value.DisplayName != "" || value.Description != nil {
		t.Fatal("group identity or nullable/empty strings changed")
	}
	if value.MailEnabled == nil || !*value.MailEnabled || value.SecurityEnabled == nil || *value.SecurityEnabled {
		t.Fatal("explicit group booleans were lost")
	}
	if value.Visibility == nil || *value.Visibility != "futureVisibility" || value.ResourceProvisioningOptions == nil || value.CreatedDateTime == nil {
		t.Fatal("future visibility, empty provisioning array or date was lost")
	}
	if string(value.RawFields["extension_future"]) != `{"value":false}` || string(value.RawFields["description"]) != "null" {
		t.Fatal("raw future or presence information was lost")
	}
	if json.Unmarshal([]byte(`{"id":"limited","mailEnabled":null,"groupTypes":null}`), &value) != nil || value.MailEnabled != nil || value.GroupTypes != nil {
		t.Fatal("unknown group properties were defaulted")
	}
}

func TestGroupReadRequiresIdentity(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`null`, `[]`, `{}`, `{"id":7}`, `{"id":"fixture","mailEnabled":"false"}`} {
		var value group.Group
		if err := json.Unmarshal([]byte(raw), &value); !errors.Is(err, client.ErrDecode) {
			t.Fatal("invalid group snapshot accepted")
		}
	}
}

func TestMicrosoft365GroupConstructorDoesNotApplyPowerShellDefaults(t *testing.T) {
	t.Parallel()
	value, err := group.NewMicrosoft365Group("Fixture Group", "fixture-group")
	if err != nil || value.DisplayName == nil || *value.DisplayName != "Fixture Group" || value.MailNickname == nil || *value.MailNickname != "fixture-group" {
		t.Fatal("required native group inputs incorrect")
	}
	if !reflect.DeepEqual(value.GroupTypes, []string{"Unified"}) || !value.MailEnabled || value.SecurityEnabled {
		t.Fatal("native Unified-group type and flags incorrect")
	}
	if value.Visibility != nil || value.OwnerIDs != nil || value.MemberIDs != nil || value.CreationOptions != nil || value.ResourceBehaviorOptions != nil {
		t.Fatal("native construction injected composition defaults or references")
	}
	for _, inputs := range [][2]string{{"", "alias"}, {"name", ""}, {" ", "alias"}, {"name", " "}} {
		if _, err := group.NewMicrosoft365Group(inputs[0], inputs[1]); !errors.Is(err, client.ErrRequest) {
			t.Fatal("empty required group property accepted")
		}
	}
}
