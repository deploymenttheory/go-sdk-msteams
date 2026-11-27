package teamscallparkpolicy_test

import (
	"context"
	"encoding/json"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_call_park_policy"
)

func TestUnitCallParkNativeRequestsForTypedInputs(t *testing.T) {
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
				Command: "New-CsTeamsCallParkPolicy",
				Name:    "omitted-settings",
				Wire:    json.RawMessage("{\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "New-CsTeamsCallParkPolicy",
				Name:    "false-zero-empty",
				Wire:    json.RawMessage("{\"AllowCallPark\":false,\"Description\":\"\",\"Identity\":\"Fixture\",\"ParkTimeoutSeconds\":0,\"PickupRangeEnd\":0,\"PickupRangeStart\":0}"),
			},

			{
				Command: "New-CsTeamsCallParkPolicy",
				Name:    "signed-and-large",
				Wire:    json.RawMessage("{\"AllowCallPark\":true,\"Identity\":\"Fixture\",\"ParkTimeoutSeconds\":-9223372036854775808,\"PickupRangeEnd\":9223372036854775807,\"PickupRangeStart\":9007199254740993}"),
			},

			{
				Command: "New-CsTeamsCallParkPolicy",
				Name:    "bound-null-fields",
				Failure: new("System.Management.Automation.PSInvalidCastException"),
			},

			{
				Command: "New-CsTeamsCallParkPolicy",
				Name:    "numeric-string",
				Wire:    json.RawMessage("{\"Identity\":\"Fixture\",\"PickupRangeStart\":9007199254740993}"),
			},

			{
				Command: "New-CsTeamsCallParkPolicy",
				Name:    "overflow",
				Failure: new("System.OverflowException"),
			},

			{
				Command: "New-CsTeamsCallParkPolicy",
				Name:    "null-Description",
				Wire:    json.RawMessage("{\"Description\":\"\",\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "New-CsTeamsCallParkPolicy",
				Name:    "null-AllowCallPark",
				Failure: new("System.Management.Automation.PSInvalidCastException"),
			},

			{
				Command: "New-CsTeamsCallParkPolicy",
				Name:    "null-ParkTimeoutSeconds",
				Wire:    json.RawMessage("{\"Identity\":\"Fixture\",\"ParkTimeoutSeconds\":0}"),
			},

			{
				Command: "New-CsTeamsCallParkPolicy",
				Name:    "null-PickupRangeStart",
				Wire:    json.RawMessage("{\"Identity\":\"Fixture\",\"PickupRangeStart\":0}"),
			},

			{
				Command: "New-CsTeamsCallParkPolicy",
				Name:    "null-PickupRangeEnd",
				Wire:    json.RawMessage("{\"Identity\":\"Fixture\",\"PickupRangeEnd\":0}"),
			},

			{
				Command: "Set-CsTeamsCallParkPolicy",
				Name:    "omitted-settings",
				Wire:    json.RawMessage("{\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "Set-CsTeamsCallParkPolicy",
				Name:    "false-zero-empty",
				Wire:    json.RawMessage("{\"AllowCallPark\":false,\"Description\":\"\",\"Identity\":\"Fixture\",\"ParkTimeoutSeconds\":0,\"PickupRangeEnd\":0,\"PickupRangeStart\":0}"),
			},

			{
				Command: "Set-CsTeamsCallParkPolicy",
				Name:    "signed-and-large",
				Wire:    json.RawMessage("{\"AllowCallPark\":true,\"Identity\":\"Fixture\",\"ParkTimeoutSeconds\":-9223372036854775808,\"PickupRangeEnd\":9223372036854775807,\"PickupRangeStart\":9007199254740993}"),
			},

			{
				Command: "Set-CsTeamsCallParkPolicy",
				Name:    "bound-null-fields",
				Failure: new("System.Management.Automation.PSInvalidCastException"),
			},

			{
				Command: "Set-CsTeamsCallParkPolicy",
				Name:    "numeric-string",
				Wire:    json.RawMessage("{\"Identity\":\"Fixture\",\"PickupRangeStart\":9007199254740993}"),
			},

			{
				Command: "Set-CsTeamsCallParkPolicy",
				Name:    "overflow",
				Failure: new("System.OverflowException"),
			},

			{
				Command: "Set-CsTeamsCallParkPolicy",
				Name:    "null-Description",
				Wire:    json.RawMessage("{\"Description\":\"\",\"Identity\":\"Fixture\"}"),
			},

			{
				Command: "Set-CsTeamsCallParkPolicy",
				Name:    "null-AllowCallPark",
				Failure: new("System.Management.Automation.PSInvalidCastException"),
			},

			{
				Command: "Set-CsTeamsCallParkPolicy",
				Name:    "null-ParkTimeoutSeconds",
				Wire:    json.RawMessage("{\"Identity\":\"Fixture\",\"ParkTimeoutSeconds\":0}"),
			},

			{
				Command: "Set-CsTeamsCallParkPolicy",
				Name:    "null-PickupRangeStart",
				Wire:    json.RawMessage("{\"Identity\":\"Fixture\",\"PickupRangeStart\":0}"),
			},

			{
				Command: "Set-CsTeamsCallParkPolicy",
				Name:    "null-PickupRangeEnd",
				Wire:    json.RawMessage("{\"Identity\":\"Fixture\",\"PickupRangeEnd\":0}"),
			},
		},
	}

	var results []testutil.Result
	for _, f := range fixture.Cases {
		if f.Failure != nil {
			continue
		} // Null Boolean/overflow have no typed native value.
		input := resource.Settings{}
		switch f.Name {
		case "omitted-settings":
		case "false-zero-empty":
			input = resource.Settings{Description: new(""), AllowCallPark: new(false), ParkTimeoutSeconds: new(int64(0)), PickupRangeStart: new(int64(0)), PickupRangeEnd: new(int64(0))}
		case "signed-and-large":
			input = resource.Settings{AllowCallPark: new(true), ParkTimeoutSeconds: new(int64(-9223372036854775808)), PickupRangeStart: new(int64(9007199254740993)), PickupRangeEnd: new(int64(9223372036854775807))}
		case "numeric-string":
			input.PickupRangeStart = new(int64(9007199254740993))
		case "null-Description":
			input.Description = new("")
		case "null-ParkTimeoutSeconds":
			input.ParkTimeoutSeconds = new(int64(0))
		case "null-PickupRangeStart":
			input.PickupRangeStart = new(int64(0))
		case "null-PickupRangeEnd":
			input.PickupRangeEnd = new(int64(0))
		default:
			t.Fatal("unrecognized typed input")
		}
		method, path, status := "PUT", collection+"/configuration/Fixture", 400
		if f.Command == newCommand {
			method, path, status = "POST", collection, 400
		}
		results = append(results, run(t, testutil.Case{Want: testutil.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}}, Name: strings.ToLower(f.Command + "-" + f.Name), Steps: []testutil.Exchange{discovery(t), wire(method, path, f.Command, f.Wire, status, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			s := sdk.TeamsAdminAPI.TeamsCallParkPolicies
			if f.Command == newCommand {
				r, e := s.CreateByIdentity(ctx, "Fixture", input)
				return testutil.Observation{Response: r, Err: e}
			}
			r, e := s.UpdateByIdentity(ctx, "Fixture", input)
			return testutil.Observation{Response: r, Err: e}
		}}))
	}
	if len(results) != 16 {
		t.Fatal("source binding success count changed")
	}
}
