package applicationaccesspolicy_test

import (
	"encoding/json"
	"errors"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	policy "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/application_access_policy"
)

const responsePolicyID = "fixture-policy"
const responseApplicationID = "00000000-0000-4000-8000-000000000001"

func TestUnitGroupGrantUsesAcceptanceResponse(t *testing.T) {
	t.Parallel()
	const group = "00000000-0000-4000-8000-000000000020"
	sdk, calls := scriptedClient(t, []wireCase{{Method: "POST", URL: policyOrigin + "/Skype.Policy/groupPolicyAssignments/" + group + "/policyTypes/ApplicationAccessPolicy",
		Body: json.RawMessage(`{"policyName":"fixture-policy-a","priority":1}`), Status: 204, Response: applicationResponseFixture(t, "assign_group_204.body")}})
	rank := 1
	value, response, err := sdk.TeamsAdminAPI.ApplicationAccessPolicies.AssignToGroup(t.Context(), group, "fixture-policy-a", &rank)
	if err != nil || response == nil || response.StatusCode() != 204 || value == nil || value.Accepted || value.OperationID != "" || *calls != 2 {
		t.Fatal("actual HTTP204 group grant fabricated asynchronous operation data or changed its request")
	}
}

func applicationResponseFixture(t *testing.T, name string) string {
	t.Helper()
	root, err := os.OpenRoot("testdata")
	if err != nil {
		t.Fatal(err)
	}
	defer func() {
		if err := root.Close(); err != nil {
			t.Error(err)
		}
	}()
	raw, err := root.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func discoveryResponseFixture(t *testing.T) string {
	t.Helper()
	raw, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return string(raw)
}

func responsePolicyIdentity(value string) bool {
	return strings.EqualFold(strings.TrimPrefix(value, "Tag:"), responsePolicyID)
}

func TestUnitCreateUsesAcceptanceReadBackResponse(t *testing.T) {
	t.Parallel()
	initial := "SDK lab non-default initial description"
	read := applicationResponseFixture(t, "create_readback.json")
	sdk, calls := scriptedClient(t, []wireCase{
		{Method: "POST", URL: policyOrigin + "/Skype.Policy/configurations/ApplicationAccessPolicy",
			Body: json.RawMessage(`{"Identity":"fixture-policy","AppIds":["` + responseApplicationID + `"],"Description":"` + initial + `"}`), Status: 201},
		{Method: "GET", URL: policyOrigin + "/Skype.Policy/configurations/ApplicationAccessPolicy/configuration/" + responsePolicyID,
			Status: 200, Response: read},
	})
	value, response, err := sdk.TeamsAdminAPI.ApplicationAccessPolicies.Create(t.Context(), policy.CreateRequest{
		Identity: responsePolicyID, AppIDs: policy.ReplaceAppIDs([]string{responseApplicationID}), Description: &initial,
	})
	if err != nil || response == nil || response.StatusCode() != 200 || value == nil {
		t.Fatal("create/read-back failed")
	}
	if !responsePolicyIdentity(value.Identity) || value.Description == nil || *value.Description != initial ||
		len(value.AppIDs) != 1 || value.AppIDs[0] != responseApplicationID {
		t.Fatal("acceptance response fields were not decoded")
	}
	if value.RawFields["Key"] == nil || value.RawFields["ConfigMetadata"] == nil || *calls != 3 {
		t.Fatal("server metadata was lost or request sequence changed")
	}
}

func TestUnitUpdateUsesAcceptanceResponseAndPreservesOmittedAppIDs(t *testing.T) {
	t.Parallel()
	updated := "SDK lab non-default updated description"
	sdk, calls := scriptedClient(t, []wireCase{
		{Method: "PUT", URL: policyOrigin + "/Skype.Policy/configurations/ApplicationAccessPolicy/configuration/" + responsePolicyID,
			Body:   json.RawMessage(`{"Identity":"fixture-policy","Description":"` + updated + `"}`),
			Status: 204, Response: applicationResponseFixture(t, "update_204.body")},
		{Method: "GET", URL: policyOrigin + "/Skype.Policy/configurations/ApplicationAccessPolicy/configuration/" + responsePolicyID,
			Status: 200, Response: applicationResponseFixture(t, "update_readback.json")},
	})
	if _, err := sdk.TeamsAdminAPI.ApplicationAccessPolicies.UpdateByIdentity(t.Context(), responsePolicyID, policy.UpdateRequest{Description: &updated}); err != nil {
		t.Fatal(err)
	}
	value, _, err := sdk.TeamsAdminAPI.ApplicationAccessPolicies.GetByIdentity(t.Context(), responsePolicyID)
	if err != nil || value == nil || value.Description == nil || *value.Description != updated ||
		len(value.AppIDs) != 1 || value.AppIDs[0] != responseApplicationID || *calls != 3 {
		t.Fatal("updated response or omitted AppIds differ")
	}
}

func TestUnitListUsesAcceptanceResponse(t *testing.T) {
	t.Parallel()
	raw := applicationResponseFixture(t, "list.json")
	var expected []json.RawMessage
	if json.Unmarshal([]byte(raw), &expected) != nil {
		t.Fatal("invalid acceptance list fixture")
	}
	sdk, _ := scriptedClient(t, []wireCase{{Method: "GET", URL: policyOrigin + "/Skype.Policy/configurations/ApplicationAccessPolicy", Status: 200, Response: raw}})
	values, _, err := sdk.TeamsAdminAPI.ApplicationAccessPolicies.List(t.Context())
	if err != nil || len(values) != len(expected) {
		t.Fatal("acceptance list was not decoded completely")
	}
	found := 0
	for _, value := range values {
		if responsePolicyIdentity(value.Identity) {
			found++
		}
		if value.RawFields["ConfigId"] == nil {
			t.Fatal("original response metadata lost")
		}
	}
	if found != 1 {
		t.Fatal("owned acceptance policy missing or duplicated")
	}
}

func TestUnitDeleteUsesAcceptanceNotFoundResponse(t *testing.T) {
	t.Parallel()
	sdk, calls := scriptedClient(t, []wireCase{
		{Method: "DELETE", URL: policyOrigin + "/Skype.Policy/configurations/ApplicationAccessPolicy/configuration/" + responsePolicyID, Status: 204},
		{Method: "GET", URL: policyOrigin + "/Skype.Policy/configurations/ApplicationAccessPolicy/configuration/" + responsePolicyID,
			Status: 404, Response: applicationResponseFixture(t, "get_deleted_404.json")},
	})
	if _, err := sdk.TeamsAdminAPI.ApplicationAccessPolicies.DeleteByIdentity(t.Context(), responsePolicyID); err != nil {
		t.Fatal(err)
	}
	_, response, err := sdk.TeamsAdminAPI.ApplicationAccessPolicies.GetByIdentity(t.Context(), responsePolicyID)
	var apiError *client.APIError
	if !errors.As(err, &apiError) || apiError.StatusCode != 404 || response == nil || response.StatusCode() != 404 || *calls != 3 {
		t.Fatal("real deleted-policy response was not retained and classified")
	}
}
