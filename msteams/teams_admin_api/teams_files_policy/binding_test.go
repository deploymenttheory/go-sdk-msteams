package teamsfilespolicy_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_files_policy"
)

func TestUnitFilesPolicyTypedInputsAndScopeRequests(t *testing.T) {
	t.Parallel()
	cases := []struct {
		Name, Command, Set, Input string
		Invalid                   bool
	}{
		{Name: "list", Command: "Get-CsTeamsFilesPolicy", Set: "Identity", Input: "{}", Invalid: false},
		{Name: "filter", Command: "Get-CsTeamsFilesPolicy", Set: "Filter", Input: "{\"Filter\":\"Fixture*\"}", Invalid: false},
		{Name: "empty-filter", Command: "Get-CsTeamsFilesPolicy", Set: "Filter", Input: "{\"Filter\":\"\"}", Invalid: true},
		{Name: "empty-identity", Command: "Get-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"\"}", Invalid: true},
		{Name: "identity-only", Command: "New-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\"}", Invalid: false},
		{Name: "empty-identity", Command: "New-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"\"}", Invalid: true},
		{Name: "identity-only", Command: "Set-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\"}", Invalid: true},
		{Name: "force-only", Command: "Set-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\",\"Force\":{\"IsPresent\":false}}", Invalid: true},
		{Name: "empty-identity", Command: "Set-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"\",\"NativeFileEntryPoints\":\"\"}", Invalid: true},
		{Name: "force-false", Command: "Remove-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\",\"Force\":{\"IsPresent\":false}}", Invalid: false},
		{Name: "no-scope", Command: "Grant-CsTeamsFilesPolicy", Set: "Identity", Input: "{}", Invalid: true},
		{Name: "user-omitted-name", Command: "Grant-CsTeamsFilesPolicy", Set: "GrantToUser", Input: "{\"Identity\":\"user/fixture\"}", Invalid: false},
		{Name: "user-null-name", Command: "Grant-CsTeamsFilesPolicy", Set: "GrantToUser", Input: "{\"Identity\":\"user/fixture\",\"PolicyName\":null}", Invalid: false},
		{Name: "user-empty-name", Command: "Grant-CsTeamsFilesPolicy", Set: "GrantToUser", Input: "{\"Identity\":\"user/fixture\",\"PolicyName\":\"\"}", Invalid: false},
		{Name: "tenant-false-selector", Command: "Grant-CsTeamsFilesPolicy", Set: "GrantToTenant", Input: "{\"Force\":{\"IsPresent\":false},\"PolicyName\":null,\"Global\":{\"IsPresent\":false}}", Invalid: false},
		{Name: "tenant-true-force", Command: "Grant-CsTeamsFilesPolicy", Set: "GrantToTenant", Input: "{\"Force\":{\"IsPresent\":true},\"PolicyName\":\"Fixture\",\"Global\":{\"IsPresent\":true}}", Invalid: false},
		{Name: "group-rank-zero", Command: "Grant-CsTeamsFilesPolicy", Set: "GrantToGroup", Input: "{\"Group\":\"group/fixture\",\"Rank\":0,\"PolicyName\":null}", Invalid: false},
		{Name: "group-rank-negative", Command: "Grant-CsTeamsFilesPolicy", Set: "GrantToGroup", Input: "{\"Group\":\"group/fixture\",\"Rank\":-1,\"PolicyName\":\"\"}", Invalid: false},
		{Name: "group-rank-two", Command: "Grant-CsTeamsFilesPolicy", Set: "GrantToGroup", Input: "{\"Group\":\"group/fixture\",\"Rank\":2,\"PolicyName\":\"Fixture\"}", Invalid: false},
		{Name: "NativeFileEntryPoints-null", Command: "New-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\",\"NativeFileEntryPoints\":null}", Invalid: false},
		{Name: "NativeFileEntryPoints-empty", Command: "New-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\",\"NativeFileEntryPoints\":\"\"}", Invalid: false},
		{Name: "SPChannelFilesTab-null", Command: "New-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\",\"SPChannelFilesTab\":null}", Invalid: false},
		{Name: "SPChannelFilesTab-empty", Command: "New-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\",\"SPChannelFilesTab\":\"\"}", Invalid: false},
		{Name: "DefaultFileUploadAppId-null", Command: "New-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\",\"DefaultFileUploadAppId\":null}", Invalid: false},
		{Name: "DefaultFileUploadAppId-empty", Command: "New-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\",\"DefaultFileUploadAppId\":\"\"}", Invalid: false},
		{Name: "FileSharingInChatswithExternalUsers-null", Command: "New-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\",\"FileSharingInChatswithExternalUsers\":null}", Invalid: false},
		{Name: "FileSharingInChatswithExternalUsers-empty", Command: "New-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\",\"FileSharingInChatswithExternalUsers\":\"\"}", Invalid: false},
		{Name: "all-settings", Command: "New-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"DefaultFileUploadAppId\":\"Future-DefaultFileUploadAppId\",\"Identity\":\"Fixture\",\"FileSharingInChatswithExternalUsers\":\"Future-FileSharingInChatswithExternalUsers\",\"SPChannelFilesTab\":\"Future-SPChannelFilesTab\",\"NativeFileEntryPoints\":\"Future-NativeFileEntryPoints\"}", Invalid: false},
		{Name: "NativeFileEntryPoints-null", Command: "Set-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\",\"NativeFileEntryPoints\":null}", Invalid: false},
		{Name: "NativeFileEntryPoints-empty", Command: "Set-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\",\"NativeFileEntryPoints\":\"\"}", Invalid: false},
		{Name: "SPChannelFilesTab-null", Command: "Set-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\",\"SPChannelFilesTab\":null}", Invalid: false},
		{Name: "SPChannelFilesTab-empty", Command: "Set-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\",\"SPChannelFilesTab\":\"\"}", Invalid: false},
		{Name: "DefaultFileUploadAppId-null", Command: "Set-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\",\"DefaultFileUploadAppId\":null}", Invalid: false},
		{Name: "DefaultFileUploadAppId-empty", Command: "Set-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\",\"DefaultFileUploadAppId\":\"\"}", Invalid: false},
		{Name: "FileSharingInChatswithExternalUsers-null", Command: "Set-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\",\"FileSharingInChatswithExternalUsers\":null}", Invalid: false},
		{Name: "FileSharingInChatswithExternalUsers-empty", Command: "Set-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"Identity\":\"Fixture\",\"FileSharingInChatswithExternalUsers\":\"\"}", Invalid: false},
		{Name: "all-settings", Command: "Set-CsTeamsFilesPolicy", Set: "Identity", Input: "{\"DefaultFileUploadAppId\":\"Future-DefaultFileUploadAppId\",\"Identity\":\"Fixture\",\"FileSharingInChatswithExternalUsers\":\"Future-FileSharingInChatswithExternalUsers\",\"SPChannelFilesTab\":\"Future-SPChannelFilesTab\",\"NativeFileEntryPoints\":\"Future-NativeFileEntryPoints\"}", Invalid: false},
	}
	for _, f := range cases {
		var bound map[string]json.RawMessage
		if err := json.Unmarshal([]byte(f.Input), &bound); err != nil {
			t.Fatal(err)
		}
		set := f.Set

		encoded, e := json.Marshal(bound)
		if e != nil {
			t.Fatal(e)
		}
		var input resource.Settings
		if json.Unmarshal(encoded, &input) != nil {
			t.Fatal("ordinary bound input values are not typed settings")
		}
		var identity, pattern string
		var name *string
		if json.Unmarshal(bound["Identity"], &identity) != nil {
			identity = ""
		}
		if json.Unmarshal(bound["Filter"], &pattern) != nil {
			pattern = ""
		}
		if value, ok := bound["PolicyName"]; ok && json.Unmarshal(value, &name) != nil {
			t.Fatal("invalid typed policy name")
		}
		controls := policy.Controls{}
		if value, ok := bound["Force"]; ok {
			var v struct{ IsPresent bool }
			if json.Unmarshal(value, &v) != nil {
				t.Fatal("invalid typed switch")
			}
			controls.Force = new(v.IsPresent)
		}
		c := contracttest.Case{Name: strings.ToLower(f.Command + "-" + f.Name), Mappings: []contracttest.Mapping{{Command: f.Command, ParameterSet: set}}}
		invalid := f.Invalid
		if invalid {
			c.Want.Error = resource.ErrInput
		} else {
			method, path, reply, status := "GET", collection, "", 400
			var body json.RawMessage
			switch f.Command {
			case getCommand:
				if identity != "" {
					path = collection + "/configuration/Fixture"

				}

			case "New-CsTeamsFilesPolicy", setCommand, "Remove-CsTeamsFilesPolicy":
				method, path, reply, status = "POST", collection, "", 400
				fields := map[string]json.RawMessage{}
				if f.Command != "Remove-CsTeamsFilesPolicy" {
					settings, _ := json.Marshal(input)
					if json.Unmarshal(settings, &fields) != nil {
						t.Fatal("invalid settings projection")
					}
				}
				fields["Identity"], _ = json.Marshal(identity)
				if controls.Force != nil {
					value := "False"
					if *controls.Force {
						value = "True"
					}
					fields["Force"], _ = json.Marshal(value)
				}
				body, _ = json.Marshal(fields)
				if f.Command == setCommand {
					method, path = "PATCH", collection+"/configuration/Fixture"
				}
				if f.Command == "Remove-CsTeamsFilesPolicy" {
					method, path = "DELETE", collection+"/configuration/Fixture"
				}
			case grantCommand:
				reply, status = "", 400
				var fields map[string]any
				switch set {
				case "GrantToUser":
					method, path = "PATCH", "/Skype.Policy/users/user%2Ffixture/policies/TeamsFilesPolicy"
					fields = map[string]any{"policyName": name, "additionalParameters": nil}
				case "GrantToTenant":
					method, path = "PATCH", "/Skype.Policy/tenants/policies/TeamsFilesPolicy"
					fields = map[string]any{"PolicyName": name, "additionalParameters": nil, "forceSwitchPresent": controls.Force != nil && *controls.Force}
				case "GrantToGroup":
					var rank int32
					if json.Unmarshal(bound["Rank"], &rank) != nil {
						t.Fatal("required typed Rank missing")
					}
					method, path = "POST", "/Skype.Policy/groupPolicyAssignments/group%2Ffixture/policyTypes/TeamsFilesPolicy"
					fields = map[string]any{"policyName": name, "priority": rank}
				default:
					t.Fatal("valid no-scope grant")
				}
				body, _ = json.Marshal(fields)
			default:
				t.Fatal("unknown test command")
			}
			c.Steps = []contracttest.Exchange{compiledWire(method, path, f.Command, body, status, reply)}
			c.Want.APIError = &client.APIError{Backend: client.TeamsAdmin, StatusCode: 400}
		}
		c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.TeamsFilesPolicies
			switch f.Command {
			case getCommand:
				if _, ok := bound["Filter"]; ok {
					v, r, e := s.ListFiltered(ctx, pattern)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				if _, ok := bound["Identity"]; ok {
					v, r, e := s.GetByIdentity(ctx, identity)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				v, r, e := s.List(ctx)
				return contracttest.Observation{Value: v, Response: r, Err: e}
			case "New-CsTeamsFilesPolicy":
				v, r, e := s.CreateByIdentity(ctx, identity, resource.CreateRequest{Settings: input, Controls: controls})
				return contracttest.Observation{Value: v, Response: r, Err: e}
			case setCommand:
				v, r, e := s.PatchByIdentity(ctx, identity, resource.PatchRequest{Settings: input, Controls: controls})
				return contracttest.Observation{Value: v, Response: r, Err: e}
			case "Remove-CsTeamsFilesPolicy":
				v, r, e := s.DeleteByIdentity(ctx, identity, controls)
				return contracttest.Observation{Value: v, Response: r, Err: e}
			case grantCommand:
				switch set {
				case "GrantToUser":
					v, r, e := s.AssignUser(ctx, identity, name)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "GrantToTenant":
					v, r, e := s.AssignTenantDefault(ctx, name, controls.Force != nil && *controls.Force)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				case "GrantToGroup":
					var rank int32
					var id string
					if json.Unmarshal(bound["Rank"], &rank) != nil || json.Unmarshal(bound["Group"], &id) != nil {
						return contracttest.Observation{Err: errors.New("invalid typed grant")}
					}
					v, r, e := s.AssignGroup(ctx, id, name, rank)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				default:
					v, r, e := s.Assign(ctx, policy.Assignment{})
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
			}
			return contracttest.Observation{Err: errors.New("unknown test command")}
		}
		run(t, c)
	}
}
