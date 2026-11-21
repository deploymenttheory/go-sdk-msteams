package group_test

import (
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
)

func TestUnitGroupPatchPreservesOmissionAndExplicitEmpty(t *testing.T) {
	t.Parallel()
	empty := ""
	visibility := group.Visibility("")
	for _, tc := range []struct {
		name    string
		request group.UpdateRequest
		want    string
	}{
		{name: "omitted", request: group.UpdateRequest{}, want: `{}`},
		{name: "explicit-empty", request: group.UpdateRequest{DisplayName: &empty, Description: &empty, Visibility: &visibility, MailNickname: &empty, Classification: &empty}, want: `{"classification":"","description":"","displayName":"","mailNickname":"","visibility":""}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			actual, err := json.Marshal(tc.request)
			if err != nil {
				t.Fatal(err)
			}
			var got, want map[string]any
			if json.Unmarshal(actual, &got) != nil || json.Unmarshal([]byte(tc.want), &want) != nil || !reflect.DeepEqual(got, want) {
				t.Fatal("group patch lost omitted or explicit-empty input")
			}
		})
	}
}

func TestUnitGroupReadPreservesNullableAndFutureFields(t *testing.T) {
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

func TestUnitGroupReadRequiresIdentity(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`null`, `[]`, `{}`, `{"id":7}`, `{"id":"fixture","mailEnabled":"false"}`} {
		var value group.Group
		if err := json.Unmarshal([]byte(raw), &value); !errors.Is(err, client.ErrDecode) {
			t.Fatal("invalid group snapshot accepted")
		}
	}
}

func TestUnitMicrosoft365GroupConstructorDoesNotApplyPowerShellDefaults(t *testing.T) {
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
