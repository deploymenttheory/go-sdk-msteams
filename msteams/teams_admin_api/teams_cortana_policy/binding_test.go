package teamscortanapolicy_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_cortana_policy"
)

func TestCortanaNativeRequestsMatchSuccessfulIsolatedBinding(t *testing.T) {
	t.Parallel()
	raw, e := os.ReadFile("../../../" + modelSource)
	if e != nil {
		t.Fatal(e)
	}
	var fixture struct {
		Cases []struct {
			Command, Name string
			Wire          json.RawMessage
			Failure       *string `json:"failure_type"`
		}
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Cases) != 22 {
		t.Fatal("incomplete binding fixture")
	}
	var results []contracttest.Result
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
			t.Fatal("unrecognized successful source binding")
		}
		method, path, status := "PUT", collection+"/configuration/Fixture", 204
		if f.Command == newCommand {
			method, path, status = "POST", collection, 201
		}
		results = append(results, run(t, contracttest.Case{Name: strings.ToLower(f.Command + "-" + f.Name), Source: modelSource, Mappings: []contracttest.Mapping{{Command: f.Command, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{discovery(), wire(method, path, f.Command, f.Wire, status, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsCortanaPolicies
			if f.Command == newCommand {
				r, e := s.CreateByIdentity(ctx, "Fixture", input)
				return contracttest.Observation{Response: r, Err: e}
			}
			r, e := s.UpdateByIdentity(ctx, "Fixture", input)
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	if len(results) != 12 {
		t.Fatal("source binding success count changed")
	}
	evidence(t, results)
}
