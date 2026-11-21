package tenantfederationconfiguration_test

import (
	"encoding/json"

	"net/http"
	"os"

	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/TenantFederationSettings"
const getCommand = "Get-CsTenantFederationConfiguration"
const setCommand = "Set-CsTenantFederationConfiguration"
const source = "msteams/teams_admin_api/tenant_federation_configuration/testdata/get_global.json"

func discovery(t *testing.T) contracttest.Exchange {
	t.Helper()
	raw, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience, Headers: http.Header{"Accept": {"application/json"}}}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: raw}}
}

func wire(method, path, command string, body json.RawMessage, status int, reply string) contracttest.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {"https://admin.example.invalid/"}}
	if body != nil {
		h.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}
func run(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	if len(c.Steps) > 0 {
		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	} else {
		c.Scope = "local"
	}
	c.Family = "tenant_federation_configuration"
	c.Source = source
	if c.Scope == "" {
		c.Scope = "service"
	}
	r, err := contracttest.Run(t.Context(), c)
	if err != nil || !r.Passed() {
		t.Fatalf("configuration differs: %v %+v", err, r.Differences())
	}
	return r
}

type fixture struct {
	Cases []struct {
		Name     string          `json:"name"`
		Method   string          `json:"method"`
		URL      string          `json:"generated_url"`
		Body     json.RawMessage `json:"body"`
		Status   int             `json:"response_status"`
		Response string          `json:"response_body"`
	} `json:"cases"`
	Models []struct {
		Name     string          `json:"name"`
		Rejected bool            `json:"rejected"`
		Wire     json.RawMessage `json:"wire"`
	} `json:"models"`
}

func pinned(t *testing.T) fixture {
	t.Helper()

	f := fixture{
		Cases: []struct {
			Name     string          `json:"name"`
			Method   string          `json:"method"`
			URL      string          `json:"generated_url"`
			Body     json.RawMessage `json:"body"`
			Status   int             `json:"response_status"`
			Response string          `json:"response_body"`
		}{

			{
				Name:   "tenant-federation-list",
				Method: "GET",
				URL:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TenantFederationSettings",
				Status: 400,
			},

			{
				Name:   "tenant-federation-read",
				Method: "GET",
				URL:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TenantFederationSettings/configuration/Global",
				Status: 400,
			},

			{
				Name:   "tenant-federation-update",
				Method: "PUT",
				URL:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TenantFederationSettings/configuration/Global",
				Body:   json.RawMessage("{\"AllowFederatedUsers\":false,\"AllowedDomains\":[\"fixture.example.invalid\"],\"Identity\":\"Global\"}"),
				Status: 400,
			},
		},
		Models: []struct {
			Name     string          `json:"name"`
			Rejected bool            `json:"rejected"`
			Wire     json.RawMessage `json:"wire"`
		}{

			{
				Name:     "omitted",
				Rejected: false,
				Wire:     json.RawMessage("{\"Identity\":\"Global\"}"),
			},

			{
				Name:     "bound-scalars",
				Rejected: false,
				Wire:     json.RawMessage("{\"AllowFederatedUsers\":false,\"AllowTeamsConsumer\":false,\"AllowTeamsConsumerInbound\":false,\"BlockAllSubdomains\":false,\"EnableExternalAccessRestrictionsForChatParticipants\":false,\"EnableMutualFederationForChatParticipants\":false,\"ExternalAccessWithTrialTenants\":\"FutureState\",\"Identity\":\"Global\",\"RestrictTeamsConsumerToExternalUserProfiles\":false,\"SecurityTeamAllowBlockListDelegation\":\"\",\"SharedSipAddressSpace\":false,\"TreatDiscoveredPartnersAsUnverified\":false}"),
			},

			{
				Name:     "allow-all",
				Rejected: false,
				Wire:     json.RawMessage("{\"AllowedDomains\":[\"AllowAllKnownDomains\"],\"Identity\":\"Global\"}"),
			},

			{
				Name:     "allow-list",
				Rejected: false,
				Wire:     json.RawMessage("{\"AllowedDomains\":[\"b.example.invalid\",\"a.example.invalid\",\"b.example.invalid\",\"\"],\"Identity\":\"Global\"}"),
			},

			{
				Name:     "empty-allow-list",
				Rejected: false,
				Wire:     json.RawMessage("{\"AllowedDomains\":[],\"Identity\":\"Global\"}"),
			},

			{
				Name:     "allowed-replacement",
				Rejected: false,
				Wire:     json.RawMessage("{\"AllowedDomains\":[\"a.example.invalid\",\"b.example.invalid\"],\"Identity\":\"Global\"}"),
			},

			{
				Name:     "allowed-operators",
				Rejected: false,
				Wire:     json.RawMessage("{\"AllowedDomains\":{\"Add\":[\"a.example.invalid\"],\"Remove\":[\"b.example.invalid\"]},\"Identity\":\"Global\"}"),
			},

			{
				Name:     "allowed-empty-operators",
				Rejected: false,
				Wire:     json.RawMessage("{\"AllowedDomains\":{},\"Identity\":\"Global\"}"),
			},

			{
				Name:     "blocked-mixed",
				Rejected: false,
				Wire:     json.RawMessage("{\"BlockedDomains\":[\"a.example.invalid\",{\"Domain\":\"object.example.invalid\",\"Enabled\":false}],\"Identity\":\"Global\"}"),
			},

			{
				Name:     "blocked-operators",
				Rejected: false,
				Wire:     json.RawMessage("{\"BlockedDomains\":{\"Add\":[{\"Domain\":\"object.example.invalid\",\"Enabled\":false},\"a.example.invalid\"],\"Remove\":[\"b.example.invalid\"]},\"Identity\":\"Global\"}"),
			},

			{
				Name:     "trial-mixed",
				Rejected: false,
				Wire:     json.RawMessage("{\"AllowedTrialTenantDomains\":[{\"Domain\":\"object.example.invalid\",\"Enabled\":false},\"a.example.invalid\"],\"Identity\":\"Global\"}"),
			},

			{
				Name:     "trial-null",
				Rejected: false,
				Wire:     json.RawMessage("{\"AllowedTrialTenantDomains\":[],\"Identity\":\"Global\"}"),
			},

			{
				Name:     "bound-null-scalar",
				Rejected: false,
				Wire:     json.RawMessage("{\"Identity\":\"Global\"}"),
			},

			{
				Name:     "conflicting-choices",
				Rejected: true,
			},

			{
				Name:     "null-choice",
				Rejected: true,
			},

			{
				Name:     "replace-operator",
				Rejected: true,
			},
		},
	}

	return f
}
