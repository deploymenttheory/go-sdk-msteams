package teamsenhancedencryptionpolicy_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_enhanced_encryption_policy"
)

func TestUnitEnhancedEncryptionRequestsMatchLayeredBinding(t *testing.T) {
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
		PrivateIdentities []struct {
			Command, Name string
			Failure       *string `json:"failure_type"`
		} `json:"private_identity_bindings"`
		Metadata []struct {
			CallingModes []string `json:"calling_mode_valid_set"`
			Unchanged    bool     `json:"future_value_unchanged"`
			Validator    string   `json:"validator_body"`
		} `json:"private_metadata"`
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Cases) != 22 || len(fixture.PrivateIdentities) != 8 || len(fixture.Metadata) != 2 {
		t.Fatal("incomplete layered binding fixture")
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
	var results []contracttest.Result
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
		c := contracttest.Case{Name: strings.ToLower(f.Command + "-" + f.Name), Source: modelSource, Mappings: []contracttest.Mapping{{Command: f.Command, ParameterSet: "Identity"}}}
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
			method, path, status := "PUT", collection+"/configuration/Fixture", 204
			if f.Command == newCommand {
				method, path, status = "POST", collection, 201
			}
			c.Steps = []contracttest.Exchange{discovery(), wire(method, path, f.Command, f.Wire, status, "")}
		}
		c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsEnhancedEncryptionPolicies
			if f.Command == newCommand {
				r, e := s.CreateByIdentity(ctx, bound.Identity, input)
				return contracttest.Observation{Response: r, Err: e}
			}
			r, e := s.UpdateByIdentity(ctx, bound.Identity, input)
			return contracttest.Observation{Response: r, Err: e}
		}
		results = append(results, run(t, c))
	}
	evidence(t, results)
}
