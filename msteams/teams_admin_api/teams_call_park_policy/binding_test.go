package teamscallparkpolicy_test

import (
	"context"
	"encoding/json"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_call_park_policy"
)

func TestUnitCallParkNativeRequestsMatchSuccessfulIsolatedBinding(t *testing.T) {
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
			t.Fatal("unrecognized successful source binding")
		}
		method, path, status := "PUT", collection+"/configuration/Fixture", 204
		if f.Command == newCommand {
			method, path, status = "POST", collection, 201
		}
		results = append(results, run(t, contracttest.Case{Name: strings.ToLower(f.Command + "-" + f.Name), Source: modelSource, Mappings: []contracttest.Mapping{{Command: f.Command, ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{discovery(), wire(method, path, f.Command, f.Wire, status, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsCallParkPolicies
			if f.Command == newCommand {
				r, e := s.CreateByIdentity(ctx, "Fixture", input)
				return contracttest.Observation{Response: r, Err: e}
			}
			r, e := s.UpdateByIdentity(ctx, "Fixture", input)
			return contracttest.Observation{Response: r, Err: e}
		}}))
	}
	if len(results) != 16 {
		t.Fatal("source binding success count changed")
	}
	evidence(t, results)
}
