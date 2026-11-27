package teamscortanapolicy_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_cortana_policy"
)

func TestUnitCortanaNativeRequestsForTypedInputs(t *testing.T) {
	t.Parallel()

	fixture := struct {
		Cases []struct {
			Command, Name string
			Wire          json.RawMessage
			Failure       *string `json:"failure_type"`
		}
	}{
		Cases: []struct {
			Command, Name string
			Wire          json.RawMessage
			Failure       *string `json:"failure_type"`
		}{

			{
				Command: "New-CsTeamsCortanaPolicy",
				Name:    "omitted-settings",
				Wire:    json.RawMessage("{\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "New-CsTeamsCortanaPolicy",
				Name:    "false-empty",
				Wire:    json.RawMessage("{\"AllowCortanaAmbientListening\":false,\"AllowCortanaInContextSuggestions\":false,\"AllowCortanaVoiceInvocation\":false,\"CortanaVoiceInvocationMode\":\"\",\"Description\":\"\",\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "New-CsTeamsCortanaPolicy",
				Name:    "true-future",
				Wire:    json.RawMessage("{\"AllowCortanaAmbientListening\":true,\"AllowCortanaInContextSuggestions\":true,\"AllowCortanaVoiceInvocation\":true,\"CortanaVoiceInvocationMode\":\"FutureMode\",\"Description\":\"" + strings.Repeat("x", 2000) + "\",\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "New-CsTeamsCortanaPolicy",
				Name:    "bound-null-fields",
				Failure: new("System.Management.Automation.PSInvalidCastException"),
			},

			{
				Command: "New-CsTeamsCortanaPolicy",
				Name:    "numeric-boolean-zero",
				Wire:    json.RawMessage("{\"AllowCortanaVoiceInvocation\":false,\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "New-CsTeamsCortanaPolicy",
				Name:    "boolean-string",
				Failure: new("System.Management.Automation.PSInvalidCastException"),
			},

			{
				Command: "New-CsTeamsCortanaPolicy",
				Name:    "null-Description",
				Wire:    json.RawMessage("{\"Description\":\"\",\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "New-CsTeamsCortanaPolicy",
				Name:    "null-CortanaVoiceInvocationMode",
				Wire:    json.RawMessage("{\"CortanaVoiceInvocationMode\":\"\",\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "New-CsTeamsCortanaPolicy",
				Name:    "null-AllowCortanaVoiceInvocation",
				Failure: new("System.Management.Automation.PSInvalidCastException"),
			},

			{
				Command: "New-CsTeamsCortanaPolicy",
				Name:    "null-AllowCortanaAmbientListening",
				Failure: new("System.Management.Automation.PSInvalidCastException"),
			},

			{
				Command: "New-CsTeamsCortanaPolicy",
				Name:    "null-AllowCortanaInContextSuggestions",
				Failure: new("System.Management.Automation.PSInvalidCastException"),
			},

			{
				Command: "Set-CsTeamsCortanaPolicy",
				Name:    "omitted-settings",
				Wire:    json.RawMessage("{\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "Set-CsTeamsCortanaPolicy",
				Name:    "false-empty",
				Wire:    json.RawMessage("{\"AllowCortanaAmbientListening\":false,\"AllowCortanaInContextSuggestions\":false,\"AllowCortanaVoiceInvocation\":false,\"CortanaVoiceInvocationMode\":\"\",\"Description\":\"\",\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "Set-CsTeamsCortanaPolicy",
				Name:    "true-future",
				Wire:    json.RawMessage("{\"AllowCortanaAmbientListening\":true,\"AllowCortanaInContextSuggestions\":true,\"AllowCortanaVoiceInvocation\":true,\"CortanaVoiceInvocationMode\":\"FutureMode\",\"Description\":\"" + strings.Repeat("x", 2000) + "\",\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "Set-CsTeamsCortanaPolicy",
				Name:    "bound-null-fields",
				Failure: new("System.Management.Automation.PSInvalidCastException"),
			},

			{
				Command: "Set-CsTeamsCortanaPolicy",
				Name:    "numeric-boolean-zero",
				Wire:    json.RawMessage("{\"AllowCortanaVoiceInvocation\":false,\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "Set-CsTeamsCortanaPolicy",
				Name:    "boolean-string",
				Failure: new("System.Management.Automation.PSInvalidCastException"),
			},

			{
				Command: "Set-CsTeamsCortanaPolicy",
				Name:    "null-Description",
				Wire:    json.RawMessage("{\"Description\":\"\",\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "Set-CsTeamsCortanaPolicy",
				Name:    "null-CortanaVoiceInvocationMode",
				Wire:    json.RawMessage("{\"CortanaVoiceInvocationMode\":\"\",\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "Set-CsTeamsCortanaPolicy",
				Name:    "null-AllowCortanaVoiceInvocation",
				Failure: new("System.Management.Automation.PSInvalidCastException"),
			},

			{
				Command: "Set-CsTeamsCortanaPolicy",
				Name:    "null-AllowCortanaAmbientListening",
				Failure: new("System.Management.Automation.PSInvalidCastException"),
			},

			{
				Command: "Set-CsTeamsCortanaPolicy",
				Name:    "null-AllowCortanaInContextSuggestions",
				Failure: new("System.Management.Automation.PSInvalidCastException"),
			},
		},
	}

	var results []testutil.Result
	for _, f := range fixture.Cases {
		if f.Failure != nil {
			continue
		} // Null Boolean/Boolean string have no typed native value.
		input := resource.Settings{}
		switch f.Name {
		case "omitted-settings":
		case "false-empty":
			input = resource.Settings{Description: new(""), CortanaVoiceInvocationMode: new(""), AllowCortanaVoiceInvocation: new(false), AllowCortanaAmbientListening: new(false), AllowCortanaInContextSuggestions: new(false)}
		case "true-future":
			input = resource.Settings{Description: new(strings.Repeat("x", 2000)), CortanaVoiceInvocationMode: new("FutureMode"), AllowCortanaVoiceInvocation: new(true), AllowCortanaAmbientListening: new(true), AllowCortanaInContextSuggestions: new(true)}
		case "numeric-boolean-zero":
			input.AllowCortanaVoiceInvocation = new(false)
		case "null-Description":
			input.Description = new("")
		case "null-CortanaVoiceInvocationMode":
			input.CortanaVoiceInvocationMode = new("")
		default:
			t.Fatal("unrecognized typed input")
		}
		method, path, status := "PUT", collection+"/configuration/Fixture", 400
		if f.Command == newCommand {
			method, path, status = "POST", collection, 400
		}
		results = append(results, run(t, testutil.Case{Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}, Name: strings.ToLower(f.Command + "-" + f.Name), Steps: []testutil.Exchange{discovery(t), wire(method, path, f.Command, f.Wire, status, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.TeamsCortanaPolicies
			if f.Command == newCommand {
				r, e := s.CreateByIdentity(ctx, "Fixture", input)
				return testutil.Observation{Response: r, Err: e}
			}
			r, e := s.UpdateByIdentity(ctx, "Fixture", input)
			return testutil.Observation{Response: r, Err: e}
		}}))
	}
	if len(results) != 12 {
		t.Fatal("source binding success count changed")
	}
}
