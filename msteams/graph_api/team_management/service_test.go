package teammanagement_test

import (
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
	management "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team_management"
)

const groupID = "00000000-0000-4000-8000-000000000002"
const ownerID = "00000000-0000-4000-8000-000000000001"

func creationRequest() management.CreateRequest {
	alias := "fixture"
	return management.CreateRequest{DisplayName: "Fixture", MailNickname: &alias, Owner: ownerID}
}
func runCase(t *testing.T, command string, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Family = "team_management"
	c.Source = "msteams/graph_api/team_management/service_test.go"
	if c.Name == "" {
		c.Name = "request-case"
	}
	if c.Scope == "" {
		c.Scope = "service"
	}
	if c.Mode == "" {
		c.Mode = auth.Delegated
	}
	if len(c.Mappings) == 0 {
		set := "__AllParameterSets"
		if command == "Get-Team" {
			set = "Identity"
		}
		c.Mappings = []contracttest.Mapping{{Command: command, ParameterSet: set}}
	}
	result, err := contracttest.Run(t.Context(), c)
	if err != nil || !result.Passed() {
		t.Fatalf("SDK request/validation differed: setup=%v differences=%+v", err, result.Differences())
	}
	return result
}
func creationCase(t *testing.T, set string, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Mappings = []contracttest.Mapping{{Command: "New-Team", ParameterSet: set}}
	return runCase(t, "New-Team", c)
}
func listCase(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	c.Mappings = []contracttest.Mapping{{Command: "Get-Team", ParameterSet: "Filters"}}
	return runCase(t, "Get-Team", c)
}
func denied(method, path, body string) contracttest.Exchange {
	headers := http.Header{"Accept": {"application/json"}}
	var request json.RawMessage
	if body != "" {
		request = json.RawMessage(body)
		headers.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: auth.GraphAudience + path, Audience: auth.GraphAudience, Headers: headers, JSON: request}, Reply: contracttest.Reply{Status: http.StatusForbidden, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(`{"error":{"code":"UnitRequestDenied"}}`)}}
}

// These cases prove the outgoing first request and denial propagation. Successful
// paired reads, hydration, provisioning, compensation and archive workflows need
// real Graph TestAcc responses; none are inferred from a fabricated 2xx body.
func TestUnitManagementFirstRequestDenials(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, command string
		request       contracttest.Exchange
		invoke        func(context.Context, *msteams.Client) error
	}{
		{name: "get-group-before-team", command: "Get-Team", request: denied("GET", "/v1.0/groups/"+groupID, ""), invoke: func(ctx context.Context, sdk *msteams.Client) error {
			_, err := sdk.GraphAPI.TeamManagement.GetByID(ctx, groupID)
			return err
		}},
		{name: "patch-group-before-team", command: "Set-Team", request: denied("PATCH", "/v1.0/groups/"+groupID, `{}`), invoke: func(ctx context.Context, sdk *msteams.Client) error {
			_, err := sdk.GraphAPI.TeamManagement.UpdateByID(ctx, groupID, management.UpdateRequest{})
			return err
		}},
		{name: "migrate-supplied-group", command: "New-Team", request: denied("GET", "/v1.0/groups/"+groupID, ""), invoke: func(ctx context.Context, sdk *msteams.Client) error {
			_, err := sdk.GraphAPI.TeamManagement.MigrateGroup(ctx, groupID, management.CreationOptions{})
			return err
		}},
		{name: "archive-preread", command: "Set-TeamArchivedState", request: denied("GET", "/v1.0/groups/"+groupID, ""), invoke: func(ctx context.Context, sdk *msteams.Client) error {
			_, err := sdk.GraphAPI.TeamManagement.SetArchivedState(ctx, groupID, true, management.ArchiveStateOptions{})
			return err
		}},
		{name: "unarchive-preread", command: "Set-TeamArchivedState", request: denied("GET", "/v1.0/groups/"+groupID, ""), invoke: func(ctx context.Context, sdk *msteams.Client) error {
			_, err := sdk.GraphAPI.TeamManagement.SetArchivedState(ctx, groupID, false, management.ArchiveStateOptions{})
			return err
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			c := contracttest.Case{Name: tc.name, Steps: []contracttest.Exchange{tc.request}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return contracttest.Observation{Err: tc.invoke(ctx, sdk)}
			}, Want: contracttest.Expected{Error: auth.ErrForbidden}}
			if tc.command == "New-Team" {
				creationCase(t, "MigrateGroup", c)
			} else {
				runCase(t, tc.command, c)
			}
		})
	}
}
func TestUnitManagementListDiscoveryDenial(t *testing.T) {
	t.Parallel()
	for _, search := range []string{"", "Fixture"} {
		t.Run(map[bool]string{true: "tenant", false: "search"}[search == ""], func(t *testing.T) {
			q := url.Values{"$filter": {"resourceProvisioningOptions/Any(x:x eq 'Team')"}, "$top": {"999"}, "$count": {"true"}}
			if search != "" {
				q.Set("$search", `"displayName:`+search+`"`)
			}
			step := denied("GET", "/beta/groups?"+q.Encode(), "")
			step.Request.Headers.Set("ConsistencyLevel", "eventual")
			listCase(t, contracttest.Case{Name: "discovery-denial", Steps: []contracttest.Exchange{step}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				result, err := sdk.GraphAPI.TeamManagement.List(ctx, management.ListOptions{DisplayName: search})
				if result.Complete || result.DiscoveryComplete || len(result.Items) != 0 {
					t.Error("denied discovery was reported complete")
				}
				return contracttest.Observation{Err: err}
			}, Want: contracttest.Expected{Error: auth.ErrForbidden}})
		})
	}
}
func TestUnitManagementCreateRequestVariantsDenied(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name            string
		template        management.GroupTemplate
		visibility      string
		options         []string
		resourceOptions []string
		education       bool
	}{
		{name: "ordinary", visibility: "Private", options: []string{"ExchangeProvisioningFlags:3552"}},
		{name: "education-class", template: management.EducationClass, visibility: "HiddenMembership", options: []string{"ExchangeProvisioningFlags:461", "classAssignments"}, resourceOptions: []string{"appRoleForSite:8f348934-64be-4bb2-bc16-c54c96789f43:fullcontrol", "appRoleForSite:2d4d3d8e-2be3-4bef-9f87-7875a61c29de:fullcontrol", "appRoleForSite:13291f5a-59ac-4c59-b0fa-d1632e8f3292:fullcontrol", "appRoleForSite:c9a559d2-7aab-4f13-a6ed-e7e9c52aec87:fullcontrol", "appRoleForSite:22d27567-b3f0-4dc2-9ec2-46ed368ba538:fullcontrol"}, education: true},
		{name: "education-plc", template: management.EducationPLC, visibility: "Private", options: []string{"PLC"}, resourceOptions: []string{"appRoleForSite:13291f5a-59ac-4c59-b0fa-d1632e8f3292:fullcontrol", "appRoleForSite:2d4d3d8e-2be3-4bef-9f87-7875a61c29de:fullcontrol"}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			body := map[string]any{"displayName": "Fixture", "mailNickname": "fixture", "groupTypes": []string{"Unified"}, "mailEnabled": true, "securityEnabled": false, "visibility": tc.visibility, "creationOptions": tc.options, "owners@odata.bind": []string{auth.GraphAudience + "/v1.0/users/" + ownerID}, "members@odata.bind": []string{auth.GraphAudience + "/v1.0/users/" + ownerID}}
			if tc.resourceOptions != nil {
				body["resourceBehaviorOptions"] = tc.resourceOptions
			}
			if tc.education {
				body["extension_fe2174665583431c953114ff7268b7b3_Education_ObjectType"] = "Section"
			}
			payload, err := json.Marshal(body)
			if err != nil {
				t.Fatal(err)
			}
			creationCase(t, "CreateTeam", contracttest.Case{Name: tc.name, Steps: []contracttest.Exchange{denied("POST", "/v1.0/groups", string(payload))}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				request := creationRequest()
				request.Template = tc.template
				result, err := sdk.GraphAPI.TeamManagement.Create(ctx, request)
				if result.ID != "" || result.CreatedGroupOwned || result.Cleanup.Disposition == management.CleanupDeleted {
					t.Error("denied creation fabricated ownership or cleanup")
				}
				return contracttest.Observation{Err: err}
			}, Want: contracttest.Expected{Error: auth.ErrForbidden}})
		})
	}
}
func TestUnitManagementSparsePropertyAndSettingsInput(t *testing.T) {
	t.Parallel()
	empty, disabled := "", false
	request := management.UpdateRequest{Properties: group.UpdateRequest{Description: &empty}, Settings: team.Settings{MemberSettings: &team.MemberSettings{AllowCreateUpdateChannels: &disabled}}}
	runCase(t, "Set-Team", contracttest.Case{Name: "sparse-property-denied", Steps: []contracttest.Exchange{denied("PATCH", "/v1.0/groups/"+groupID, `{"description":""}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		result, err := sdk.GraphAPI.TeamManagement.UpdateByID(ctx, groupID, request)
		if result.Value != nil {
			t.Error("denied update fabricated a read-back")
		}
		if request.Settings.GuestSettings != nil || request.Settings.MessagingSettings != nil || request.Settings.DiscoverySettings != nil || *request.Settings.MemberSettings.AllowCreateUpdateChannels {
			t.Error("failed update changed caller settings")
		}
		return contracttest.Observation{Err: err}
	}, Want: contracttest.Expected{Error: auth.ErrForbidden}})
}
