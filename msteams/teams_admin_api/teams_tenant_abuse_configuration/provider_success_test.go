package teamstenantabuseconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_tenant_abuse_configuration"
)

// Authored from the pinned Core nullable scalar/list DTO, not a tenant capture.
const authoredRead = `{"Identity":"Global","LowSeatLimit":18446744073709551615,"TrialTenantValidation":null,"SkipValidation":"","CreateThreadThresholdPerSeat":0,"AddMemberThresholdPerSeat":null,"SendMessageThresholdPerSeat":9007199254740993,"CreateThreadThresholdForTrialTenants":1,"AddMemberThresholdForTrialTenants":2,"SendMessageThresholdForTrialTenants":3,"BlockedTenants":["entry",null,"entry",null],"Future":false}`

const getCommand = "Get-CsTeamsTenantAbuseConfiguration"
const setCommand = "Set-CsTeamsTenantAbuseConfiguration"
const selected = collection + "/configuration/Global"

func assertAuthoredRead(t *testing.T, value *resource.Configuration) {
	t.Helper()
	if value == nil || value.Identity == nil || *value.Identity != "Global" || value.LowSeatLimit == nil || *value.LowSeatLimit != ^uint64(0) || value.TrialTenantValidation != nil || value.SkipValidation == nil || *value.SkipValidation != "" || value.CreateThreadThresholdPerSeat == nil || *value.CreateThreadThresholdPerSeat != 0 || value.AddMemberThresholdPerSeat != nil || value.SendMessageThresholdPerSeat == nil || *value.SendMessageThresholdPerSeat != 9007199254740993 || value.CreateThreadThresholdForTrialTenants == nil || *value.CreateThreadThresholdForTrialTenants != 1 || value.AddMemberThresholdForTrialTenants == nil || *value.AddMemberThresholdForTrialTenants != 2 || value.SendMessageThresholdForTrialTenants == nil || *value.SendMessageThresholdForTrialTenants != 3 || !reflect.DeepEqual(value.BlockedTenants, []*string{new("entry"), nil, new("entry"), nil}) || string(value.RawFields["Future"]) != "false" || string(value.RawJSON) != authoredRead {
		t.Fatal("authored read lost typed width, zero/null/empty, nullable entries or raw fields")
	}
}

func TestUnitTenantAbuseAuthoredTypedReadListAndLiteralFilter(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "typed-read", Steps: []testutil.Exchange{compiledWire("GET", selected, getCommand, nil, 200, authoredRead)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		value, response, err := sdk.TeamsAdminAPI.TeamsTenantAbuseConfigurations.GetByIdentity(ctx, "Global")
		if err == nil {
			assertAuthoredRead(t, value)
		}
		return testutil.Observation{Response: response, Err: err}
	}})
	run(t, testutil.Case{Name: "typed-collection", Steps: []testutil.Exchange{compiledWire("GET", collection, getCommand, nil, 200, "["+authoredRead+"]")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		values, response, err := sdk.TeamsAdminAPI.TeamsTenantAbuseConfigurations.List(ctx)
		if err == nil && len(values) != 1 {
			t.Fatal("typed Global-only collection changed")
		}
		if err == nil {
			assertAuthoredRead(t, values[0])
		}
		return testutil.Observation{Response: response, Err: err}
	}})
	for _, pattern := range []string{"Glo?al", " "} {
		run(t, testutil.Case{Name: "literal-filter", Steps: []testutil.Exchange{compiledWire("GET", collection, getCommand, nil, 200, "["+authoredRead+"]")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			values, response, err := sdk.TeamsAdminAPI.TeamsTenantAbuseConfigurations.ListFiltered(ctx, pattern)
			if err == nil && (pattern == " " && len(values) != 0 || pattern != " " && len(values) != 1) {
				t.Fatal("literal whitespace or wildcard selection changed")
			}
			if err == nil && len(values) != 0 {
				assertAuthoredRead(t, values[0])
			}
			return testutil.Observation{Response: response, Err: err}
		}})
	}
}

func TestUnitTenantAbuseNullablePreparedEditRetainsLaterFailureAndExplicitRecovery(t *testing.T) {
	t.Parallel()
	const accepted = `{"Identity":"Global","LowSeatLimit":0,"BlockedTenants":["entry",null,null,"ENTRY"]}`
	run(t, testutil.Case{Name: "nullable-edit-failure-recovery", Steps: []testutil.Exchange{
		compiledWire("GET", selected, setCommand, nil, 200, authoredRead),
		compiledWire("PATCH", selected, setCommand, json.RawMessage(accepted), 204, ""),
		compiledWire("GET", selected, setCommand, nil, 200, accepted),
		compiledWire("PATCH", selected, setCommand, json.RawMessage(`{"Identity":"Global","SkipValidation":null}`), 403, `{"Message":"controlled denial"}`),
		compiledWire("PATCH", selected, setCommand, json.RawMessage(`{"Identity":"Global","LowSeatLimit":18446744073709551615,"BlockedTenants":["entry",null,"entry",null]}`), 204, ""),
		compiledWire("GET", selected, getCommand, nil, 200, authoredRead),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		service := sdk.TeamsAdminAPI.TeamsTenantAbuseConfigurations
		first, response, err := service.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{LowSeatLimit: policy.Set(uint64(0))}, BlockedTenants: policy.ChangeNullableStrings([]*string{nil, new("ENTRY")}, []*string{nil, new("entry")})})
		if err != nil || first == nil || !first.PatchInvoked || !first.HTTPAccepted || first.After != nil || first.ReadResponse == nil || string(first.ReadResponse.Body()) != authoredRead || response == nil || response.StatusCode() != 204 {
			t.Fatal("nullable edit lost preparation/acceptance or invented persistence")
		}
		assertAuthoredRead(t, first.Before)
		later, response, err := service.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{SkipValidation: policy.Null[string]()}})
		var apiError *client.APIError
		if !errors.As(err, &apiError) || apiError.StatusCode != 403 || response == nil || response.StatusCode() != 403 || later == nil || later.Before == nil || later.Before.LowSeatLimit == nil || *later.Before.LowSeatLimit != 0 || !reflect.DeepEqual(later.Before.BlockedTenants, []*string{new("entry"), nil, nil, new("ENTRY")}) || later.ReadResponse == nil || later.ReadResponse.StatusCode() != 200 || string(later.ReadResponse.Body()) != accepted || !later.PatchInvoked || later.HTTPAccepted || later.After != nil || first.ReadResponse.StatusCode() != 200 || !first.HTTPAccepted {
			t.Fatal("later denial lost its partial receipt or earlier accepted result")
		}
		assertAuthoredRead(t, first.Before)
		// Restoration is an explicit caller operation; no automatic rollback runs.
		if value, response, err := service.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{ScalarSettings: resource.ScalarSettings{LowSeatLimit: policy.Set(*first.Before.LowSeatLimit)}, BlockedTenants: policy.NullableStrings(first.Before.BlockedTenants...)}}); err != nil || value != nil || response == nil || response.StatusCode() != 204 {
			t.Fatal("caller restoration lost null entries or its accepted empty response")
		}
		value, response, err := service.GetByIdentity(ctx, "Global")
		if err == nil {
			assertAuthoredRead(t, value)
		}
		return testutil.Observation{Response: response, Err: err}
	}})
}

func TestUnitTenantAbuseMalformedSuccessRetainsPreparationReceiptWithoutPatch(t *testing.T) {
	t.Parallel()
	const invalid = `{"LowSeatLimit":18446744073709551616}`
	run(t, testutil.Case{Name: "overflow-preparation", Steps: []testutil.Exchange{compiledWire("GET", selected, setCommand, nil, 200, invalid)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		value, response, err := sdk.TeamsAdminAPI.TeamsTenantAbuseConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{BlockedTenants: policy.ReplaceNullableStrings(nil)})
		if !errors.Is(err, client.ErrDecode) || response == nil || string(response.Body()) != invalid || value == nil || value.ReadResponse != response || value.Before != nil || value.After != nil || value.PatchInvoked || value.HTTPAccepted {
			t.Fatal("malformed successful preparation lost receipt or allowed a later PATCH")
		}
		return testutil.Observation{Response: response}
	}})
}
