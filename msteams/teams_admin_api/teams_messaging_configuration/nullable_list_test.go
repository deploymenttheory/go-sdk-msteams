package teamsmessagingconfiguration_test

import (
	"context"
	"encoding/json"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_messaging_configuration"
)

// The pinned DTO/resolver and actual PSListModifier permit null entries. These
// replies are authored protocol vectors, not captured null-member tenant writes.
func TestUnitMessagingNullableListPreparedEditRetainsSnapshotAndExplicitReadback(t *testing.T) {
	t.Parallel()
	const before = `{"Identity":"Global","CustomEmojis":true,"Storyline":"Enabled","FileTypeBlockList":["entry",null,"entry",null],"Future":false}`
	const after = `{"Identity":"Global","CustomEmojis":false,"Storyline":"Enabled","FileTypeBlockList":["entry",null,null,"ENTRY"],"Future":false}`
	run(t, contracttest.Case{Name: "nullable-list-prepared-edit", Source: "msteams/teams_admin_api/teams_messaging_configuration/nullable_list_test.go", Mappings: []contracttest.Mapping{{Command: set, ParameterSet: "Identity"}, {Command: "Get-CsTeamsMessagingConfiguration", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{
		compiledWire("GET", collection, "Get-CsTeamsMessagingConfiguration", nil, 200, "["+before+"]"),
		compiledWire("GET", selected, set, nil, 200, before),
		compiledWire("PATCH", selected, set, json.RawMessage(`{"Identity":"Global","CustomEmojis":false,"FileTypeBlockList":["entry",null,null,"ENTRY"]}`), 204, ""),
		compiledWire("GET", selected, "Get-CsTeamsMessagingConfiguration", nil, 200, after),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		service := sdk.TeamsAdminAPI.TeamsMessagingConfigurations
		// Omitted Identity is the public default collection path, independent of Filter.
		listed, listResponse, listErr := service.List(ctx)
		if listErr != nil || listResponse == nil || listResponse.StatusCode() != 200 || string(listResponse.Body()) != "["+before+"]" || len(listed) != 1 || listed[0] == nil || listed[0].Identity == nil || *listed[0].Identity != "Global" || !reflect.DeepEqual(listed[0].FileTypeBlockList, []*string{new("entry"), nil, new("entry"), nil}) {
			t.Fatal("default collection read lost typed identity, nullable entries or raw receipt")
		}
		result, response, err := service.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{CustomEmojis: policy.Set(false)}, FileTypeBlockList: policy.ChangeNullableStrings([]*string{nil, new("ENTRY")}, []*string{nil, new("entry")})})
		if err != nil || response == nil || response.StatusCode() != 204 || result == nil || result.Before == nil || result.Before.Identity == nil || *result.Before.Identity != "Global" || result.Before.CustomEmojis == nil || !*result.Before.CustomEmojis || result.Before.Storyline == nil || *result.Before.Storyline != resource.StorylineEnabled || !reflect.DeepEqual(result.Before.FileTypeBlockList, []*string{new("entry"), nil, new("entry"), nil}) || string(result.Before.RawJSON) != before || result.ReadResponse == nil || string(result.ReadResponse.Body()) != before || !result.PatchInvoked || !result.HTTPAccepted || result.After != nil {
			t.Fatal("nullable edit lost the typed prepared snapshot/acceptance or invented readback")
		}
		// Readback is caller-requested and remains separate from the empty204 receipt.
		value, response, err := service.GetByIdentity(ctx, "Global")
		if err == nil && (value == nil || value.CustomEmojis == nil || *value.CustomEmojis || !reflect.DeepEqual(value.FileTypeBlockList, []*string{new("entry"), nil, nil, new("ENTRY")}) || string(value.RawJSON) != after || string(value.RawFields["Future"]) != "false" || !reflect.DeepEqual(result.Before.FileTypeBlockList, []*string{new("entry"), nil, new("entry"), nil})) {
			t.Fatal("explicit readback lost null entries or mutated earlier preparation")
		}
		return contracttest.Observation{Response: response, Err: err}
	}})
}
