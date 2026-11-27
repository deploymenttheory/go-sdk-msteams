package teamsenhancedencryptionpolicy_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_enhanced_encryption_policy"
)

func TestUnitEnhancedEncryptionRequestsMatchLayeredBinding(t *testing.T) {
	t.Parallel()

	fixture := struct {
		Cases []struct {
			Command, Name string
			Wire          json.RawMessage
			Failure       *string `json:"failure_type"`
		}
		PrivateIdentities []struct {
			Command, Name string
			Failure       *string `json:"failure_type"`
		} `json:"private_identity_bindings"`
		Metadata []struct {
			CallingModes []string `json:"calling_mode_valid_set"`
			Unchanged    bool     `json:"future_value_unchanged"`
			Validator    string   `json:"validator_body"`
		} `json:"private_metadata"`
	}{
		Cases: []struct {
			Command, Name string
			Wire          json.RawMessage
			Failure       *string `json:"failure_type"`
		}{

			{
				Command: "New-CsTeamsEnhancedEncryptionPolicy",
				Name:    "omitted-settings",
				Wire:    json.RawMessage("{\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "New-CsTeamsEnhancedEncryptionPolicy",
				Name:    "empty-strings",
				Wire:    json.RawMessage("{\"CallingEndtoEndEncryptionEnabledType\":\"\",\"Description\":\"\",\"Identity\":\"Fixture\",\"MeetingEndToEndEncryption\":\"\"}"),
			},

			{
				Command: "New-CsTeamsEnhancedEncryptionPolicy",
				Name:    "future-strings",
				Wire:    json.RawMessage("{\"CallingEndtoEndEncryptionEnabledType\":\"FutureCalling\",\"Description\":\"" + strings.Repeat("x", 2000) + "\",\"Identity\":\"Fixture\",\"MeetingEndToEndEncryption\":\"FutureMeeting\"}"),
			},

			{
				Command: "New-CsTeamsEnhancedEncryptionPolicy",
				Name:    "bound-null-fields",
				Wire:    json.RawMessage("{\"CallingEndtoEndEncryptionEnabledType\":\"\",\"Description\":\"\",\"Identity\":\"Fixture\",\"MeetingEndToEndEncryption\":\"\"}"),
			},

			{
				Command: "New-CsTeamsEnhancedEncryptionPolicy",
				Name:    "numeric-string",
				Wire:    json.RawMessage("{\"CallingEndtoEndEncryptionEnabledType\":\"0\",\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "New-CsTeamsEnhancedEncryptionPolicy",
				Name:    "boolean-string",
				Wire:    json.RawMessage("{\"Identity\":\"Fixture\",\"MeetingEndToEndEncryption\":\"False\"}"),
			},

			{
				Command: "New-CsTeamsEnhancedEncryptionPolicy",
				Name:    "empty-identity",
				Failure: new("System.Management.Automation.ParameterBindingValidationException"),
			},

			{
				Command: "New-CsTeamsEnhancedEncryptionPolicy",
				Name:    "null-identity",
				Failure: new("System.Management.Automation.ParameterBindingValidationException"),
			},

			{
				Command: "New-CsTeamsEnhancedEncryptionPolicy",
				Name:    "null-Description",
				Wire:    json.RawMessage("{\"Description\":\"\",\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "New-CsTeamsEnhancedEncryptionPolicy",
				Name:    "null-CallingEndtoEndEncryptionEnabledType",
				Wire:    json.RawMessage("{\"CallingEndtoEndEncryptionEnabledType\":\"\",\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "New-CsTeamsEnhancedEncryptionPolicy",
				Name:    "null-MeetingEndToEndEncryption",
				Wire:    json.RawMessage("{\"Identity\":\"Fixture\",\"MeetingEndToEndEncryption\":\"\"}"),
			},

			{
				Command: "Set-CsTeamsEnhancedEncryptionPolicy",
				Name:    "omitted-settings",
				Wire:    json.RawMessage("{\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "Set-CsTeamsEnhancedEncryptionPolicy",
				Name:    "empty-strings",
				Wire:    json.RawMessage("{\"CallingEndtoEndEncryptionEnabledType\":\"\",\"Description\":\"\",\"Identity\":\"Fixture\",\"MeetingEndToEndEncryption\":\"\"}"),
			},

			{
				Command: "Set-CsTeamsEnhancedEncryptionPolicy",
				Name:    "future-strings",
				Wire:    json.RawMessage("{\"CallingEndtoEndEncryptionEnabledType\":\"FutureCalling\",\"Description\":\"" + strings.Repeat("x", 2000) + "\",\"Identity\":\"Fixture\",\"MeetingEndToEndEncryption\":\"FutureMeeting\"}"),
			},

			{
				Command: "Set-CsTeamsEnhancedEncryptionPolicy",
				Name:    "bound-null-fields",
				Wire:    json.RawMessage("{\"CallingEndtoEndEncryptionEnabledType\":\"\",\"Description\":\"\",\"Identity\":\"Fixture\",\"MeetingEndToEndEncryption\":\"\"}"),
			},

			{
				Command: "Set-CsTeamsEnhancedEncryptionPolicy",
				Name:    "numeric-string",
				Wire:    json.RawMessage("{\"CallingEndtoEndEncryptionEnabledType\":\"0\",\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "Set-CsTeamsEnhancedEncryptionPolicy",
				Name:    "boolean-string",
				Wire:    json.RawMessage("{\"Identity\":\"Fixture\",\"MeetingEndToEndEncryption\":\"False\"}"),
			},

			{
				Command: "Set-CsTeamsEnhancedEncryptionPolicy",
				Name:    "empty-identity",
				Wire:    json.RawMessage("{\"Description\":\"\",\"Identity\":\"\"}"),
			},

			{
				Command: "Set-CsTeamsEnhancedEncryptionPolicy",
				Name:    "null-identity",
				Wire:    json.RawMessage("{\"Description\":\"\",\"Identity\":\"\"}"),
			},

			{
				Command: "Set-CsTeamsEnhancedEncryptionPolicy",
				Name:    "null-Description",
				Wire:    json.RawMessage("{\"Description\":\"\",\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "Set-CsTeamsEnhancedEncryptionPolicy",
				Name:    "null-CallingEndtoEndEncryptionEnabledType",
				Wire:    json.RawMessage("{\"CallingEndtoEndEncryptionEnabledType\":\"\",\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "Set-CsTeamsEnhancedEncryptionPolicy",
				Name:    "null-MeetingEndToEndEncryption",
				Wire:    json.RawMessage("{\"Identity\":\"Fixture\",\"MeetingEndToEndEncryption\":\"\"}"),
			},
		},
		PrivateIdentities: []struct {
			Command, Name string
			Failure       *string `json:"failure_type"`
		}{

			{
				Command: "New-CsTeamsEnhancedEncryptionPolicy",
				Name:    "null",
				Failure: new("System.Management.Automation.ValidationMetadataException"),
			},

			{
				Command: "New-CsTeamsEnhancedEncryptionPolicy",
				Name:    "empty",
				Failure: new("System.Management.Automation.ValidationMetadataException"),
			},

			{
				Command: "New-CsTeamsEnhancedEncryptionPolicy",
				Name:    "whitespace",
			},

			{
				Command: "New-CsTeamsEnhancedEncryptionPolicy",
				Name:    "selected",
			},

			{
				Command: "Set-CsTeamsEnhancedEncryptionPolicy",
				Name:    "null",
				Failure: new("System.Management.Automation.ValidationMetadataException"),
			},

			{
				Command: "Set-CsTeamsEnhancedEncryptionPolicy",
				Name:    "empty",
				Failure: new("System.Management.Automation.ValidationMetadataException"),
			},

			{
				Command: "Set-CsTeamsEnhancedEncryptionPolicy",
				Name:    "whitespace",
			},

			{
				Command: "Set-CsTeamsEnhancedEncryptionPolicy",
				Name:    "selected",
			},
		},
		Metadata: []struct {
			CallingModes []string `json:"calling_mode_valid_set"`
			Unchanged    bool     `json:"future_value_unchanged"`
			Validator    string   `json:"validator_body"`
		}{

			{
				CallingModes: []string{
					"Disabled",
					"DisabledUserOverride",
				},
				Unchanged: true,
				Validator: "verified-empty-il-nop-branch-ret",
			},

			{
				CallingModes: []string{
					"Disabled",
					"DisabledUserOverride",
				},
				Unchanged: true,
				Validator: "verified-empty-il-nop-branch-ret",
			},
		},
	}

	privateRejected := map[string]bool{}
	for _, x := range fixture.PrivateIdentities {
		if (x.Failure != nil) != (x.Name == "null" || x.Name == "empty") {
			t.Fatal("private identity validation differs")
		}
		if x.Failure != nil {
			privateRejected[x.Command+"/"+x.Name] = true
		}
	}
	for _, x := range fixture.Metadata {
		if len(x.CallingModes) != 2 || x.CallingModes[0] != "Disabled" || x.CallingModes[1] != "DisabledUserOverride" || !x.Unchanged || x.Validator != "verified-empty-il-nop-branch-ret" {
			t.Fatal("private metadata/modern validator distinction differs")
		}
	}

	for _, f := range fixture.Cases {
		var bound struct {
			Identity    string
			Description *string
			Calling     *string `json:"CallingEndtoEndEncryptionEnabledType"`
			Meeting     *string `json:"MeetingEndToEndEncryption"`
		}
		if f.Failure == nil && json.Unmarshal(f.Wire, &bound) != nil {
			t.Fatal("invalid source typed wire projection")
		}
		input := resource.Settings{Description: bound.Description, CallingEndToEndEncryptionEnabledType: bound.Calling, MeetingEndToEndEncryption: bound.Meeting}
		c := testutil.Case{Name: strings.ToLower(f.Command + "-" + f.Name)}
		if bound.Identity == "" {
			kind := "empty"
			if f.Name == "null-identity" {
				kind = "null"
			}
			if !privateRejected[f.Command+"/"+kind] {
				t.Fatal("source private selector rejection is missing")
			}
			c.Want.Error = resource.ErrInput
		} else {
			method, path, status := "PUT", collection+"/configuration/Fixture", 400
			if f.Command == newCommand {
				method, path, status = "POST", collection, 400
			}
			c.Steps = []testutil.Exchange{discovery(t), wire(method, path, f.Command, f.Wire, status, "")}
		}
		if len(c.Steps) != 0 {
			c.Want = testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}
		}
		c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.TeamsEnhancedEncryptionPolicies
			if f.Command == newCommand {
				r, e := s.CreateByIdentity(ctx, bound.Identity, input)
				return testutil.Observation{Response: r, Err: e}
			}
			r, e := s.UpdateByIdentity(ctx, bound.Identity, input)
			return testutil.Observation{Response: r, Err: e}
		}
		run(t, c)
	}
}
