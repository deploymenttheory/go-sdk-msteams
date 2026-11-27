package teamseducationconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_education_configuration"
)

// Pinned Core's three plain String setters and bound-key guard preserve null,
// empty and omitted fields. These write replies are authored protocol vectors;
// only the prepared read below replays retained tenant response bytes.
func TestUnitEducationSparsePatchAndPreparedUpdateAcceptedResponses(t *testing.T) {
	t.Parallel()
	const source = "msteams/teams_admin_api/teams_education_configuration/provider_success_test.go"
	const typed = `{"Identity":"body-identity","ParentGuardianPreferredContactMethod":null,"EduGenerativeAIEnhancements":"","Future":false}`
	for _, malformed := range []bool{false, true} {
		name, reply := "typed", typed
		if malformed {
			name, reply = "malformed", `{"Identity":"body-identity","EduGenerativeAIEnhancements":false}`
		}
		run(t, contracttest.Case{Name: name, Source: source, Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{
			compiledWire("PATCH", item, setCommand, json.RawMessage(`{"Identity":"body-identity","ParentGuardianPreferredContactMethod":null,"EduGenerativeAIEnhancements":""}`), 200, reply),
		}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamsEducationConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{ParentGuardianPreferredContactMethod: policy.Null[string](), EduGenerativeAIEnhancements: policy.Set("")}, BodyIdentity: policy.Set("body-identity")})
			if response == nil || response.StatusCode() != 200 || string(response.Body()) != reply {
				t.Fatal("PATCH lost its successful HTTP receipt")
			}
			if malformed {
				if !errors.Is(err, client.ErrDecode) || value != nil {
					t.Fatal("malformed accepted body fabricated a typed value or lost decode error")
				}
				return contracttest.Observation{Response: response}
			}
			if err != nil || value == nil || value.Identity == nil || *value.Identity != "body-identity" || value.ParentGuardianPreferredContactMethod != nil || value.EduGenerativeAIEnhancements == nil || *value.EduGenerativeAIEnhancements != "" || value.UpdateParentInformation != nil || string(value.RawFields["ParentGuardianPreferredContactMethod"]) != "null" || string(value.RawFields["Future"]) != "false" || string(value.RawJSON) != typed {
				t.Fatal("typed PATCH lost distinct body identity, null, empty, omission or raw fields")
			}
			return contracttest.Observation{Response: response, Err: err}
		}})
	}

	before := readResponseFixture(t, "get_global.json")
	run(t, contracttest.Case{Name: "prepared-accepted", Source: source, Mappings: []contracttest.Mapping{{Command: setCommand, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{
		compiledWire("GET", item, setCommand, nil, 200, string(before)),
		compiledWire("PATCH", item, setCommand, json.RawMessage(`{"Identity":"Global","UpdateParentInformation":null}`), 204, ""),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		value, response, err := sdk.TeamsAdminAPI.TeamsEducationConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Settings: resource.Settings{UpdateParentInformation: policy.Null[string]()}})
		if err != nil || response == nil || response.StatusCode() != 204 || value == nil || !value.PatchInvoked || !value.HTTPAccepted || value.After != nil || value.Before == nil || value.Before.Identity == nil || *value.Before.Identity != "Global" || value.Before.UpdateParentInformation == nil || *value.Before.UpdateParentInformation != "Enabled" || value.ReadResponse == nil || value.ReadResponse.StatusCode() != 200 || string(value.ReadResponse.Body()) != string(before) {
			t.Fatal("accepted Update lost preparation or invented typed persistence/readback")
		}
		return contracttest.Observation{Response: response, Err: err}
	}})
}
