package compliancerecording_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	cr "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/compliance_recording"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

func apiError(t *testing.T, status int, body []byte) *client.APIError {
	t.Helper()
	var envelope struct {
		Code      string `json:"code"`
		ErrorCode int    `json:"errorCode"`
	}
	if json.Unmarshal(body, &envelope) != nil {
		t.Fatal("invalid error fixture")
	}
	return &client.APIError{Backend: client.TeamsAdmin, StatusCode: status, Codes: []string{envelope.Code}, ServiceErrorCodes: []int{envelope.ErrorCode}}
}

func TestUnitComplianceRecordingPolicyCRUDUsesAcceptanceResponses(t *testing.T) {
	t.Parallel()
	const initial = "SDK lab non-default initial description"
	const updated = "SDK lab non-default updated description"
	before := readResponseFixture(t, "create_readback.json")
	after := readResponseFixture(t, "update_readback.json")
	list := readResponseFixture(t, "owned_list.json")
	missing := readResponseFixture(t, "get_deleted_404.json")
	steps := []contracttest.Exchange{
		compiledWire("POST", policies, "New-CsTeamsComplianceRecordingPolicy", json.RawMessage(`{"Identity":"fixture-policy","Enabled":false,"WarnUserOnRemoval":false,"RecordReroutedCalls":true,"Description":"`+initial+`"}`), 201, string(readResponseFixture(t, "create_201.body"))),
		compiledWire("GET", policies+"/configuration/fixture-policy", "Get-CsTeamsComplianceRecordingPolicy", nil, 200, string(before)),
		compiledWire("GET", tagged, "Set-CsTeamsComplianceRecordingPolicy", nil, 200, string(before)),
		compiledWire("PATCH", tagged, "Set-CsTeamsComplianceRecordingPolicy", json.RawMessage(`{"Identity":"Tag:fixture-policy","Description":"`+updated+`"}`), 204, string(readResponseFixture(t, "update_204.body"))),
		compiledWire("GET", tagged, "Get-CsTeamsComplianceRecordingPolicy", nil, 200, string(after)),
		compiledWire("GET", policies, "Get-CsTeamsComplianceRecordingPolicy", nil, 200, string(list)),
		compiledWire("DELETE", tagged, "Remove-CsTeamsComplianceRecordingPolicy", json.RawMessage(`{"Identity":"Tag:fixture-policy","Force":"True"}`), 204, string(readResponseFixture(t, "delete_204.body"))),
		compiledWire("GET", tagged, "Get-CsTeamsComplianceRecordingPolicy", nil, 404, string(missing)),
	}
	run(t, contracttest.Case{Name: "owned-crud", Source: "msteams/teams_admin_api/compliance_recording/testdata/create_readback.json", Steps: steps,
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			s := sdk.TeamsAdminAPI.ComplianceRecording
			_, response, err := s.CreatePolicyByIdentity(ctx, "fixture-policy", cr.PolicySettings{Enabled: policy.Set(false), WarnUserOnRemoval: policy.Set(false), RecordReroutedCalls: policy.Set(true), Description: policy.Set(initial)}, policy.Controls{})
			if err != nil || response.StatusCode() != 201 {
				return contracttest.Observation{Err: errors.New("create request or acceptance changed")}
			}
			value, _, err := s.GetPolicyByIdentity(ctx, "fixture-policy")
			if err != nil || value == nil || value.Identity == nil || *value.Identity != "Tag:fixture-policy" || value.Description == nil || *value.Description != initial ||
				value.WarnUserOnRemoval == nil || *value.WarnUserOnRemoval || value.RecordReroutedCalls == nil || !*value.RecordReroutedCalls || value.ComplianceRecordingApplications == nil || value.RawFields["ConfigMetadata"] == nil {
				return contracttest.Observation{Err: errors.New("actual create response was decoded incorrectly")}
			}
			changed, response, err := s.UpdatePolicyByIdentity(ctx, "Tag:fixture-policy", cr.PolicyUpdate{PolicySettings: cr.PolicySettings{Description: policy.Set(updated)}})
			if err != nil || response.StatusCode() != 204 || changed == nil || !changed.PatchInvoked || !changed.HTTPAccepted || changed.Before == nil || *changed.Before.Description != initial {
				return contracttest.Observation{Err: errors.New("sparse update lost preparation or acceptance")}
			}
			value, _, err = s.GetPolicyByIdentity(ctx, "Tag:fixture-policy")
			if err != nil || value == nil || value.Description == nil || *value.Description != updated || value.RecordReroutedCalls == nil || !*value.RecordReroutedCalls {
				return contracttest.Observation{Err: errors.New("update read-back lost the changed or omitted setting")}
			}
			values, _, err := s.ListPolicies(ctx)
			if err != nil || len(values) != 2 || values[0].Identity == nil || *values[0].Identity != "Global" || values[1].Identity == nil || *values[1].Identity != "Tag:fixture-policy" {
				return contracttest.Observation{Err: errors.New("owned list decoded incorrectly")}
			}
			if _, response, err := s.DeletePolicyByIdentity(ctx, "Tag:fixture-policy", policy.Controls{Force: new(true)}); err != nil || response.StatusCode() != 204 {
				return contracttest.Observation{Err: errors.New("delete acceptance changed")}
			}
			_, response, err = s.GetPolicyByIdentity(ctx, "Tag:fixture-policy")
			return contracttest.Observation{Response: response, Err: err}
		}, Want: contracttest.Expected{Body: missing, APIError: apiError(t, 404, missing)}})
}

func TestUnitComplianceRecordingPolicyReadsAndFilters(t *testing.T) {
	t.Parallel()
	global := readResponseFixture(t, "get_global.json")
	list := readResponseFixture(t, "owned_list.json")
	run(t, contracttest.Case{Name: "global-and-filter", Steps: []contracttest.Exchange{
		compiledWire("GET", policies+"/configuration/Global", "Get-CsTeamsComplianceRecordingPolicy", nil, 200, string(global)),
		compiledWire("GET", policies, "Get-CsTeamsComplianceRecordingPolicy", nil, 200, string(list)),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.ComplianceRecording
		value, _, err := s.GetPolicyByIdentity(ctx, "Global")
		if err != nil || value == nil || value.Identity == nil || *value.Identity != "Global" || value.Enabled == nil || *value.Enabled || value.WarnUserOnRemoval == nil || !*value.WarnUserOnRemoval ||
			value.Description != nil || value.CustomBanner != nil || value.CustomPromptsPackageID != nil || len(value.ComplianceRecordingApplications) != 0 || len(value.RawJSON) == 0 {
			return contracttest.Observation{Err: errors.New("Global decoded incorrectly")}
		}
		if fmt.Sprint(value) != "compliance recording policy (redacted)" || fmt.Sprintf("%#v", *value) != "compliance recording policy (redacted)" {
			return contracttest.Observation{Err: errors.New("policy formatting leaked fields")}
		}
		values, response, err := s.ListPoliciesFiltered(ctx, "Tag:fixture*")
		if err == nil && (len(values) != 1 || values[0].Identity == nil || *values[0].Identity != "Tag:fixture-policy") {
			err = errors.New("filter selected the wrong policies")
		}
		return contracttest.Observation{Response: response, Err: err}
	}})
}

// TestUnitComplianceRecordingApplicationRejections replays the lab's
// rejection of an application whose Id is not a resource account and the
// absence responses for application reads, updates and removes.
func TestUnitComplianceRecordingApplicationRejections(t *testing.T) {
	t.Parallel()
	invalid := readResponseFixture(t, "application_create_invalid_400.json")
	missing := readResponseFixture(t, "application_missing_404.json")
	empty := readResponseFixture(t, "application_list.json")
	run(t, contracttest.Case{Name: "application-create-rejected", Source: "msteams/teams_admin_api/compliance_recording/testdata/application_create_invalid_400.json", Steps: []contracttest.Exchange{
		compiledWire("POST", applications, "New-CsTeamsComplianceRecordingApplication", json.RawMessage(`{"Identity":"Tag:fixture-policy/`+appID+`","Id":"`+appID+`","RequiredDuringCall":true,"ConcurrentInvitationCount":1}`), 400, string(invalid)),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, response, err := sdk.TeamsAdminAPI.ComplianceRecording.CreateApplicationByID(ctx, "Tag:fixture-policy", appID, cr.ApplicationSettings{RequiredDuringCall: policy.Set(true), ConcurrentInvitationCount: policy.Set[uint32](1)}, policy.Controls{})
		return contracttest.Observation{Response: response, Err: err}
	}, Want: contracttest.Expected{Body: invalid, APIError: apiError(t, 400, invalid)}})

	run(t, contracttest.Case{Name: "application-identity-create-paired", Source: "msteams/teams_admin_api/compliance_recording/testdata/application_create_invalid_400.json", Steps: []contracttest.Exchange{
		compiledWire("POST", applications, "New-CsTeamsComplianceRecordingApplication", json.RawMessage(`{"Identity":"Tag:fixture-policy/`+appID+`","Id":"`+appID+`","ComplianceRecordingPairedApplications":[{"Id":"`+pairedID+`"}],"Force":"False"}`), 400, string(invalid)),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		paired, err := cr.NewPairedApplication(pairedID)
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		_, response, err := sdk.TeamsAdminAPI.ComplianceRecording.CreateApplicationByIdentity(ctx, "Tag:fixture-policy/"+appID, cr.ApplicationSettings{ComplianceRecordingPairedApplications: policy.Set([]*cr.PairedApplication{paired})}, policy.Controls{Force: new(false)})
		return contracttest.Observation{Response: response, Err: err}
	}, Want: contracttest.Expected{Body: invalid, APIError: apiError(t, 400, invalid)}})

	run(t, contracttest.Case{Name: "application-absent", Source: "msteams/teams_admin_api/compliance_recording/testdata/application_missing_404.json", Steps: []contracttest.Exchange{
		compiledWire("GET", applications, "Get-CsTeamsComplianceRecordingApplication", nil, 200, string(empty)),
		compiledWire("GET", appPath, "Get-CsTeamsComplianceRecordingApplication", nil, 404, string(missing)),
		compiledWire("GET", appPath, "Set-CsTeamsComplianceRecordingApplication", nil, 404, string(missing)),
		compiledWire("DELETE", appPath, "Remove-CsTeamsComplianceRecordingApplication", json.RawMessage(`{"Identity":"Tag:fixture-policy/`+appID+`"}`), 404, string(missing)),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.ComplianceRecording
		if values, _, err := s.ListApplications(ctx); err != nil || len(values) != 0 {
			return contracttest.Observation{Err: errors.New("empty application list decoded incorrectly")}
		}
		var api *client.APIError
		if _, _, err := s.GetApplicationByID(ctx, "Tag:fixture-policy", appID); !errors.As(err, &api) || api.StatusCode != 404 {
			return contracttest.Observation{Err: errors.New("missing application read lost its typed error")}
		}
		result, _, err := s.UpdateApplicationByID(ctx, "Tag:fixture-policy", appID, cr.ApplicationUpdate{ApplicationSettings: cr.ApplicationSettings{RequiredDuringCall: policy.Set(false)}})
		if !errors.As(err, &api) || api.StatusCode != 404 || result == nil || result.PatchInvoked {
			return contracttest.Observation{Err: errors.New("failed preparation read sent a PATCH")}
		}
		_, response, err := s.DeleteApplicationByID(ctx, "Tag:fixture-policy", appID, policy.Controls{})
		return contracttest.Observation{Response: response, Err: err}
	}, Want: contracttest.Expected{Body: missing, APIError: apiError(t, 404, missing)}})
}

// TestUnitComplianceRecordingListEditsApplyToReadValue checks the
// PSListModifier composition for both list parameters.
func TestUnitComplianceRecordingListEditsApplyToReadValue(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "create_readback.json")
	run(t, contracttest.Case{Name: "policy-application-add", Steps: []contracttest.Exchange{
		compiledWire("GET", tagged, "Set-CsTeamsComplianceRecordingPolicy", nil, 200, string(before)),
		compiledWire("PATCH", tagged, "Set-CsTeamsComplianceRecordingPolicy", json.RawMessage(`{"Identity":"Tag:fixture-policy","Enabled":true,"ComplianceRecordingApplications":[{"Id":"`+appID+`","RequiredDuringCall":true}]}`), 204, ""),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		id := appID
		result, response, err := sdk.TeamsAdminAPI.ComplianceRecording.UpdatePolicyByIdentity(ctx, "Tag:fixture-policy", cr.PolicyUpdate{PolicySettings: cr.PolicySettings{Enabled: policy.Set(true)},
			Applications: &cr.ListEdit[cr.Application]{Add: []*cr.Application{{ID: &id, RequiredDuringCall: new(true)}}}})
		if err == nil && (result == nil || !result.HTTPAccepted) {
			err = errors.New("list edit acceptance lost")
		}
		return contracttest.Observation{Response: response, Err: err}
	}})

	const third = "00000000-0000-4000-8000-000000000003"
	read := `{"Identity":"Tag:fixture-policy/` + appID + `","Id":"` + appID + `","ComplianceRecordingPairedApplications":[{"Id":"` + strings.ToUpper(pairedID) + `"}],"RequiredDuringCall":true,"ConcurrentInvitationCount":1}`
	run(t, contracttest.Case{Name: "application-paired-replace", Steps: []contracttest.Exchange{
		compiledWire("GET", appPath, "Set-CsTeamsComplianceRecordingApplication", nil, 200, read),
		compiledWire("PATCH", appPath, "Set-CsTeamsComplianceRecordingApplication", json.RawMessage(`{"Identity":"Tag:fixture-policy/`+appID+`","ComplianceRecordingPairedApplications":[{"Id":"`+third+`"}]}`), 204, ""),
		compiledWire("GET", appPath, "Set-CsTeamsComplianceRecordingApplication", nil, 200, read),
		compiledWire("PATCH", appPath, "Set-CsTeamsComplianceRecordingApplication", json.RawMessage(`{"Identity":"Tag:fixture-policy/`+appID+`","ComplianceRecordingPairedApplications":[]}`), 204, ""),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.ComplianceRecording
		paired, _ := cr.NewPairedApplication(third)
		result, _, err := s.UpdateApplicationByID(ctx, "Tag:fixture-policy", appID, cr.ApplicationUpdate{PairedApplications: &cr.ListEdit[cr.PairedApplication]{Add: []*cr.PairedApplication{paired}, Remove: []string{pairedID}}})
		if err != nil || result == nil || result.Before == nil || result.Before.ConcurrentInvitationCount == nil || *result.Before.ConcurrentInvitationCount != 1 || len(result.Before.ComplianceRecordingPairedApplications) != 1 {
			return contracttest.Observation{Err: errors.New("application read or paired edit changed")}
		}
		if fmt.Sprint(*result.Before) != "compliance recording application (redacted)" || fmt.Sprint(*result) != "compliance recording update result (redacted)" {
			return contracttest.Observation{Err: errors.New("application formatting leaked fields")}
		}
		_, response, err := s.UpdateApplicationByIdentity(ctx, "Tag:fixture-policy/"+appID, cr.ApplicationUpdate{PairedApplications: &cr.ListEdit[cr.PairedApplication]{Replace: []*cr.PairedApplication{}}})
		return contracttest.Observation{Response: response, Err: err}
	}})
}

func TestUnitComplianceRecordingModernGrant(t *testing.T) {
	t.Parallel()
	const accepted = `{"OperationId":"fixture-operation"}`
	cases := []struct {
		name, method, path, set string
		body                    json.RawMessage
		status                  int
		reply                   string
		invoke                  func(context.Context, *cr.Service) (*resty.Response, error)
	}{
		{"user", "PATCH", "/Skype.Policy/users/fixture%2Fuser/policies/TeamsComplianceRecordingPolicy", "Identity", json.RawMessage(`{"policyName":"Fixture"}`), 200, "{}", func(ctx context.Context, s *cr.Service) (*resty.Response, error) {
			return s.AssignToUser(ctx, "fixture/user", "Fixture")
		}},
		{"tenant", "PATCH", "/Skype.Policy/tenants/policies/TeamsComplianceRecordingPolicy", "GrantToTenant", json.RawMessage(`{"policyName":"Fixture","forceSwitchPresent":false}`), 200, "{}", func(ctx context.Context, s *cr.Service) (*resty.Response, error) {
			_, r, err := s.Assign(ctx, cr.Assignment{Tenant: &cr.TenantAssignment{PolicyName: "Fixture"}})
			return r, err
		}},
		{"group", "POST", "/Skype.Policy/groupPolicyAssignments/fixture%2Fgroup/policyTypes/TeamsComplianceRecordingPolicy", "GrantToGroup", json.RawMessage(`{"policyName":"Fixture","priority":2}`), 202, accepted, func(ctx context.Context, s *cr.Service) (*resty.Response, error) {
			v, r, err := s.Assign(ctx, cr.Assignment{Group: &cr.GroupGrant{ID: "fixture/group", PolicyName: "Fixture", Rank: new(int32(2))}})
			if err == nil && (v == nil || !v.Accepted || v.OperationID != "fixture-operation" || fmt.Sprint(*v) != "compliance recording group assignment (redacted)") {
				err = errors.New("group acceptance changed")
			}
			return r, err
		}},
	}
	for _, c := range cases {
		run(t, contracttest.Case{Name: "grant-" + c.name, Mappings: []contracttest.Mapping{{Command: grantCommand, ParameterSet: c.set}}, Steps: []contracttest.Exchange{discovery(t), modernWire(c.method, c.path, c.body, c.status, c.reply)},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				r, err := c.invoke(ctx, sdk.TeamsAdminAPI.ComplianceRecording)
				return contracttest.Observation{Response: r, Err: err}
			}})
	}
}

func TestUnitComplianceRecordingInvalidInputHasNoHTTP(t *testing.T) {
	t.Parallel()
	run(t, contracttest.Case{Name: "local-input", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		s := sdk.TeamsAdminAPI.ComplianceRecording
		fail := func(message string) contracttest.Observation {
			return contracttest.Observation{Err: errors.New(message)}
		}
		for _, id := range []string{"", "not-a-guid", "00000000-0000-4000-8000-00000000000"} {
			if v, err := cr.NewPairedApplication(id); v != nil || !errors.Is(err, cr.ErrInput) {
				return fail("invalid paired application accepted")
			}
		}
		if v, err := cr.NewPairedApplication(pairedID); err != nil || v.ID == nil || *v.ID != pairedID || v.Identity != nil {
			return fail("paired application builder changed")
		}
		for _, parent := range []string{"", "Tag:a/b", "bad\\parent"} {
			if _, _, err := s.GetApplicationByID(ctx, parent, appID); !errors.Is(err, cr.ErrInput) {
				return fail("invalid parent accepted")
			}
		}
		if _, _, err := s.CreateApplicationByID(ctx, "Global", "bad", cr.ApplicationSettings{}, policy.Controls{}); !errors.Is(err, cr.ErrInput) {
			return fail("non-GUID application Id accepted")
		}
		if _, _, err := s.CreateApplicationByIdentity(ctx, "Global", cr.ApplicationSettings{}, policy.Controls{}); !errors.Is(err, cr.ErrInput) {
			return fail("application identity without Id accepted")
		}
		bad := "bad"
		if _, _, err := s.CreatePolicyByIdentity(ctx, "fixture", cr.PolicySettings{ComplianceRecordingApplications: policy.Set([]*cr.Application{{ID: &bad}})}, policy.Controls{}); !errors.Is(err, cr.ErrInput) {
			return fail("invalid nested application accepted")
		}
		if _, _, err := s.PatchApplicationByIdentity(ctx, "Global/"+appID, cr.ApplicationSettings{ComplianceRecordingPairedApplications: policy.Set([]*cr.PairedApplication{nil})}, policy.Controls{}); !errors.Is(err, cr.ErrInput) {
			return fail("nil paired application accepted")
		}
		if _, _, err := s.PatchPolicyByIdentity(ctx, "Global", cr.PolicySettings{}, policy.Controls{Force: new(true)}); !errors.Is(err, cr.ErrInput) {
			return fail("empty policy patch accepted")
		}
		if _, _, err := s.UpdatePolicyByIdentity(ctx, "Global", cr.PolicyUpdate{}); !errors.Is(err, cr.ErrInput) {
			return fail("empty policy update accepted")
		}
		if _, _, err := s.UpdatePolicyByIdentity(ctx, "Global", cr.PolicyUpdate{PolicySettings: cr.PolicySettings{ComplianceRecordingApplications: policy.Set([]*cr.Application{})}, Applications: &cr.ListEdit[cr.Application]{}}); !errors.Is(err, cr.ErrInput) {
			return fail("conflicting application list inputs accepted")
		}
		if _, _, err := s.UpdateApplicationByIdentity(ctx, "Global/"+appID, cr.ApplicationUpdate{}); !errors.Is(err, cr.ErrInput) {
			return fail("empty application update accepted")
		}
		for _, pattern := range []string{"", "[broken"} {
			if _, _, err := s.ListPoliciesFiltered(ctx, pattern); !errors.Is(err, cr.ErrInput) {
				return fail("invalid policy filter accepted")
			}
			if _, _, err := s.ListApplicationsFiltered(ctx, pattern); !errors.Is(err, cr.ErrInput) {
				return fail("invalid application filter accepted")
			}
		}
		for _, input := range []cr.Assignment{{}, {User: &cr.UserAssignment{Identity: "fixture"}, Tenant: &cr.TenantAssignment{}}} {
			if _, _, err := s.Assign(ctx, input); !errors.Is(err, cr.ErrInput) {
				return fail("invalid assignment scope accepted")
			}
		}
		graph, err := client.NewScope(client.Graph, sdk.TeamsAdminAPI.Scope().Session())
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		if _, _, err := cr.NewService(graph).ListPolicies(ctx); !errors.Is(err, client.ErrDestination) {
			return fail("wrong backend accepted")
		}
		for _, missing := range []*cr.Service{nil, {}, cr.NewService(nil)} {
			if _, _, err := missing.ListApplications(ctx); !errors.Is(err, client.ErrClosed) {
				return fail("missing service accepted")
			}
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, err := s.AssignTenantDefault(canceled, "", false); !errors.Is(err, context.Canceled) {
			return fail("canceled grant acquired credentials")
		}
		return contracttest.Observation{}
	}})
}
