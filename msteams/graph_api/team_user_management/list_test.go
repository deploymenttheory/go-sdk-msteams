package teamusermanagement_test

import (
	"context"
	"encoding/json"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_user_management"
)

const unknownID = "55555555-5555-4555-8555-555555555555"

func listRead(kind, body string) contracttest.Exchange {
	return wire("GET", "/v1.0/groups/"+groupID+"/"+kind+"?$select=id%2CdisplayName%2CuserPrincipalName%2CuserType", "", 200, body)
}

func listed(value management.ListResult, err error) contracttest.Observation {
	items := []map[string]any{}
	stages := []string{}
	for _, item := range value.Items {
		items = append(items, map[string]any{"id": item.UserID, "role": item.Role, "native_user_type": item.DirectoryObject.UserType, "name": item.Name, "user": item.User})
	}
	for _, stage := range value.Stages {
		stages = append(stages, string(stage.Action)+":"+string(stage.State))
	}
	return contracttest.Observation{Value: map[string]any{"items": items, "complete": value.Complete, "stages": stages}, Response: value.Response, Err: err}
}

func TestUnitListCompositionOwnerPrecedenceProjectionAndFilters(t *testing.T) {
	t.Parallel()

	owners := `{"value":[{"id":"` + userID + `","displayName":"Owner","userPrincipalName":"owner@example.invalid","userType":"Member"}]}`
	members := `{"value":[{"id":"` + guestID + `","displayName":"Guest","userPrincipalName":"guest@example.invalid","userType":"gUeSt"},{"id":"` + userID + `","displayName":"Owner","userPrincipalName":"owner@example.invalid","userType":"Member"},{"id":"` + otherID + `","displayName":"Member","userPrincipalName":"member@example.invalid","userType":"Member"},{"id":"` + unknownID + `","displayName":null,"userPrincipalName":null,"userType":null}]}`
	owner := `{"id":"` + userID + `","role":"owner","native_user_type":"Member","name":"Owner","user":"owner@example.invalid"}`
	member := `{"id":"` + otherID + `","role":"member","native_user_type":"Member","name":"Member","user":"member@example.invalid"}`
	unknown := `{"id":"` + unknownID + `","role":"member","native_user_type":null,"name":null,"user":null}`
	guest := `{"id":"` + guestID + `","role":"guest","native_user_type":"gUeSt","name":"Guest","user":"guest@example.invalid"}`
	for _, tc := range []struct {
		name            string
		role            management.Role
		owners, members bool
		items, stages   string
	}{
		{"all", "", true, true, `[` + owner + `,` + member + `,` + unknown + `,` + guest + `]`, `["read-owners:succeeded","read-members:succeeded"]`},
		{"owner-only", management.Owner, true, false, `[` + owner + `]`, `["read-owners:succeeded"]`},
		{"member-only", management.Role("MEMBER"), true, true, `[` + member + `,` + unknown + `]`, `["read-owners:succeeded","read-members:succeeded"]`},
		{"guest-only", management.Guest, false, true, `[` + guest + `]`, `["read-members:succeeded"]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			steps := []contracttest.Exchange{}
			if tc.owners {
				steps = append(steps, listRead("owners", owners))
			}
			if tc.members {
				steps = append(steps, listRead("members", members))
			}
			runCase(t, "Get-TeamUser", contracttest.Case{Name: tc.name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return listed(sdk.GraphAPI.TeamUserManagement.List(ctx, groupID, management.ListOptions{Role: tc.role}))
			},
				Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"items":` + tc.items + `,"complete":true,"stages":` + tc.stages + `}`)}})
		})
	}
}

func TestUnitListCompositionPreservesOwnerPrefixOnMemberReadFailure(t *testing.T) {
	t.Parallel()
	owner := `{"id":"` + userID + `","role":"owner","native_user_type":null,"name":null,"user":null}`
	runCase(t, "Get-TeamUser", contracttest.Case{Name: "partial-after-owners", Steps: []contracttest.Exchange{listRead("owners", `{"value":[{"id":"`+userID+`"}]}`), wire("GET", "/v1.0/groups/"+groupID+"/members?$select=id%2CdisplayName%2CuserPrincipalName%2CuserType", "", 403, `{"error":{"code":"Forbidden"}}`)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			return listed(sdk.GraphAPI.TeamUserManagement.List(ctx, groupID, management.ListOptions{}))
		},
		Want: contracttest.Expected{Status: 403, Error: auth.ErrForbidden, JSON: json.RawMessage(`{"items":[` + owner + `],"complete":false,"stages":["read-owners:succeeded","read-members:unconfirmed"]}`)}})
}

func TestUnitCompositionMappingsExistInPinnedRuntime(t *testing.T) {
	t.Parallel()
	// Source files are curated contract metadata. This guards report mappings
	// against accidental command/set changes without promoting their coverage.
	raw, err := os.ReadFile("../../../contracts/powershell/8.1.0/runtime.json")
	if err != nil {
		t.Fatal("pinned runtime metadata unavailable")
	}
	var runtime struct {
		Commands []struct {
			Name string `json:"name"`
			Sets []struct {
				Name string `json:"name"`
			} `json:"parameter_sets"`
		} `json:"commands"`
	}
	if json.Unmarshal(raw, &runtime) != nil {
		t.Fatal("invalid pinned runtime metadata")
	}
	wanted := map[string]bool{"Add-TeamUser": false, "Get-TeamUser": false, "Remove-TeamUser": false}
	for _, command := range runtime.Commands {
		if _, ok := wanted[command.Name]; ok {
			for _, set := range command.Sets {
				if set.Name == "__AllParameterSets" {
					wanted[command.Name] = true
				}
			}
		}
	}
	for _, found := range wanted {
		if !found {
			t.Fatal("composition command/set mapping absent from pinned runtime")
		}
	}
}
