package teamstenantabuseconfiguration_test

import (
	"context"
	"encoding/json"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_tenant_abuse_configuration"
)

const collection = "/Skype.Policy/configurations/TeamsTenantAbuseConfiguration"

func TestUnitAuthorizationUsesActualTenantResponse(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "get_global_403.json")
	var envelope struct {
		Code      string `json:"code"`
		ErrorCode int    `json:"errorCode"`
	}
	if json.Unmarshal(raw, &envelope) != nil {
		t.Fatal("invalid actual denial fixture")
	}
	for _, operation := range []string{"get", "list", "patch", "update-preparation"} {
		t.Run(operation, func(t *testing.T) {
			method, path, command, body := "GET", collection+"/configuration/Global", "Get-CsTeamsTenantAbuseConfiguration", json.RawMessage(nil)
			if operation == "list" {
				path = collection
			}
			if operation == "update-preparation" {
				command = "Set-CsTeamsTenantAbuseConfiguration"
			}
			if operation == "patch" {
				method, command, body = "PATCH", "Set-CsTeamsTenantAbuseConfiguration", json.RawMessage(`{"Identity":"Global","LowSeatLimit":18446744073709551615}`)
			}
			// The actual Get denial envelope exercises the common PolicyRP error path;
			// list/patch injection does not claim successful or denied live writes.
			run(t, contracttest.Case{Name: operation + "-denied", Steps: []contracttest.Exchange{compiledWire(method, path, command, body, 403, string(raw))}, Want: contracttest.Expected{APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 403, Codes: []string{envelope.Code}, ServiceErrorCodes: []int{envelope.ErrorCode}}}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.TeamsTenantAbuseConfigurations
				if operation == "get" {
					_, r, e := s.GetByIdentity(ctx, "Global")
					return contracttest.Observation{Response: r, Err: e}
				}
				if operation == "list" {
					_, r, e := s.ListFiltered(ctx, "Global")
					return contracttest.Observation{Response: r, Err: e}
				}
				if operation == "update-preparation" {
					receipt, r, e := s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{LowSeatLimit: policy.Set(^uint64(0))}})
					if receipt == nil || receipt.ReadResponse == nil || receipt.PatchInvoked || receipt.HTTPAccepted {
						t.Fatal("failed preparation lost its receipt or submitted a write")
					}
					return contracttest.Observation{Response: r, Err: e}
				}
				_, r, e := s.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{ScalarSettings: resource.ScalarSettings{LowSeatLimit: policy.Set(^uint64(0))}}})
				return contracttest.Observation{Response: r, Err: e}
			}})
		})
	}
}
