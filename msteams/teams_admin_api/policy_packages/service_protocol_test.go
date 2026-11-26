package policypackages_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy_packages"
)

// These are primary-authored offline responses, not tenant captures. Pinned
// private.dll SHA256 8b7ccef9361ee1ffeaf05b79a2638f6a6b8ace39bba034edc630e0415e95ccb6:
// Generated.Models.FormattedPackage::ToJson IL67/130/193/241/304;
// PackageServiceModelsPostPackageResponse::ToJson IL52/186/312/438/491;
// BatchPostPackageResponse IL138/191/254, CustomPackageResponse IL64.
// Exact command callback statuses and nested model provenance are recorded in
// docs/implementation/policies/pol90-current-test-audit.md. Raw future fields
// and missing/null distinctions are explicitly SDK extensions to those models.
const authoredPackage = `{"Name":"Fixture/Package:@","Description":"","PackageType":"Fixture","RecommendationType":null,"Policies":{"TeamsMeetingPolicy":"FixtureMeeting"},"Future":false}`

func verifyAuthoredPackage(t *testing.T, value *resource.Package) {
	t.Helper()
	if value == nil || value.Name == nil || *value.Name != "Fixture/Package:@" || value.Description == nil || *value.Description != "" || value.RecommendationType != nil || string(value.RawFields["RecommendationType"]) != "null" || string(value.RawFields["Future"]) != "false" || string(value.Policies["TeamsMeetingPolicy"]) != `"FixtureMeeting"` {
		t.Fatal("typed package lost identity, empty/null, dictionary or future data")
	}
}

func TestUnitPackageAuthoredTypedResponses(t *testing.T) {
	t.Parallel()
	for _, vector := range requestCases() {
		if vector.Name == "create-empty-dto" {
			continue
		}
		t.Run(vector.Name, func(t *testing.T) {
			var kind string
			switch {
			case vector.Name == "list" || vector.Name == "user-recommendations":
				kind, vector.Status, vector.Reply = "list", 200, "["+authoredPackage+"]"
			case vector.Name == "get" || vector.Name == "user-get":
				kind, vector.Status, vector.Reply = "package", 200, authoredPackage
			case vector.Name == "create" || vector.Name == "update":
				kind, vector.Status, vector.Reply = "message", 200, `{"ResponseMessage":"received","Future":false}`
				if vector.Name == "create" {
					vector.Status = 201
				}
			case vector.Name == "delete":
				kind, vector.Status, vector.Reply = "delete", 204, ""
			case strings.HasPrefix(vector.Name, "group"):
				kind, vector.Status, vector.Reply = "assignment", 200, `{"ResponseMessage":"received","FailedAssignPolicies":[],"FailedCreatePolicies":[]}`
			case vector.Name == "user-grant":
				kind, vector.Status, vector.Reply = "user", 201, `{"ResponseMessage":"received","FailedUserIds":["fixture/b"],"FailedUserErrorMessages":{"LegacyUser":["fixture/b"],"Future":["fixture-future"]},"FailedAssignPolicies":[],"FailedCreatePolicies":[]}`
			case vector.Name == "batch-grant":
				kind, vector.Status, vector.Reply = "batch", 202, `{"ResponseMessage":"received","OperationId":"opaque/job:@","FailedCreatePolicies":[]}`
			default:
				t.Fatal("request vector lacks an individually reviewed response contract")
			}
			mappings := mapping(vector.Name)
			want := contracttest.Expected{Status: vector.Status, Body: []byte(vector.Reply)}
			if kind != "delete" {
				want.JSON = json.RawMessage(vector.Reply)
			}
			r, err := contracttest.Run(t.Context(), contracttest.Case{Name: "typed-" + vector.Name, Family: "policy_packages", Scope: "service", Mode: auth.Delegated, Source: "msteams/teams_admin_api/policy_packages/service_protocol_test.go", Mappings: mappings, Steps: []contracttest.Exchange{discovery(t), wire(vector, mappings[0].Command)}, Want: want, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				observed := invoke(vector.Name, ctx, sdk)
				if observed.Err != nil {
					return observed
				}
				switch kind {
				case "list":
					values, ok := observed.Value.([]*resource.Package)
					if !ok || len(values) != 1 {
						t.Fatal("package collection did not produce a nonempty typed slice")
					}
					verifyAuthoredPackage(t, values[0])
				case "package":
					value, ok := observed.Value.(*resource.Package)
					if !ok {
						t.Fatal("package read returned the wrong type")
					}
					verifyAuthoredPackage(t, value)
				case "message", "delete":
					value, ok := observed.Value.(*resource.MessageResult)
					if !ok || value == nil || !value.HTTPAccepted || (kind == "message" && (value.ResponseMessage == nil || *value.ResponseMessage != "received" || string(value.RawFields["Future"]) != "false")) || (kind == "delete" && (value.ResponseMessage != nil || value.RawFields != nil)) {
						t.Fatal("custom package HTTP receipt was lost or fabricated from empty204")
					}
				default:
					value, ok := observed.Value.(*resource.AssignmentResult)
					if !ok || value == nil || !value.HTTPAccepted || value.ResponseMessage == nil || *value.ResponseMessage != "received" {
						t.Fatal("assignment did not retain a typed HTTP receipt")
					}
					if kind == "user" && (len(value.FailedUserIDs) != 1 || value.FailedUserIDs[0] == nil || *value.FailedUserIDs[0] != "fixture/b" || value.FailedUserErrorMessages == nil || len(value.FailedUserErrorMessages.LegacyUser) != 1 || string(value.FailedUserErrorMessages.RawFields["Future"]) != `["fixture-future"]`) {
						t.Fatal("HTTP201 erased a per-user failure or nested future detail")
					}
					if kind == "batch" && (value.OperationID == nil || *value.OperationID != "opaque/job:@") {
						t.Fatal("batch receipt lost its opaque handle")
					}
					if kind != "batch" && value.OperationID != nil {
						t.Fatal("synchronous response fabricated a batch operation handle")
					}
				}
				return observed
			}})
			if err != nil || !r.Passed() {
				t.Fatalf("authored package response differs: %v %+v", err, r.Differences())
			}
		})
	}
}

func TestUnitPackageAuthoredDomainFailureAndMalformedReceipt(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, request, reply string
		status               int
		decode               bool
	}{
		{"user-domain-error", "user-grant", `{"FailedUserIds":["fixture/b"],"ResponseMessage":"failed"}`, 520, false},
		{"group-domain-error", "group-assign", `{"FailedAssignPolicies":["FixturePolicy"],"ResponseMessage":"failed"}`, 521, false},
		{"batch-error", "batch-grant", `{"FailedCreatePolicies":["FixturePolicy"],"ResponseMessage":"failed"}`, 400, false},
		{"batch-malformed", "batch-grant", `{"OperationId":`, 202, true},
		{"create-malformed", "create", `{"ResponseMessage":`, 201, true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			var vector requestCase
			for _, candidate := range requestCases() {
				if candidate.Name == tc.request {
					vector = candidate
				}
			}
			vector.Status, vector.Reply = tc.status, tc.reply
			mappings := mapping(tc.request)
			c := contracttest.Case{Name: tc.name, Family: "policy_packages", Scope: "service", Mode: auth.Delegated, Source: "msteams/teams_admin_api/policy_packages/service_protocol_test.go", Mappings: mappings, Steps: []contracttest.Exchange{discovery(t), wire(vector, mappings[0].Command)}, Want: contracttest.Expected{Status: tc.status, Body: []byte(tc.reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				observed := invoke(tc.request, ctx, sdk)
				if tc.decode {
					if !errors.Is(observed.Err, client.ErrDecode) {
						t.Fatal("malformed accepted receipt did not retain decode failure")
					}
				} else {
					var api *client.APIError
					if !errors.As(observed.Err, &api) || api.StatusCode != tc.status {
						t.Fatal("domain failure lost its HTTP error")
					}
				}
				if tc.request == "create" {
					value, ok := observed.Value.(*resource.MessageResult)
					if !ok || value == nil || !value.HTTPAccepted {
						t.Fatal("malformed accepted create lost its HTTP receipt")
					}
				} else {
					value, ok := observed.Value.(*resource.AssignmentResult)
					if !ok || value == nil || value.HTTPAccepted != tc.decode || (!tc.decode && value.RawFields == nil) || (tc.name == "user-domain-error" && len(value.FailedUserIDs) != 1) || (tc.name == "group-domain-error" && len(value.FailedAssignPolicies) != 1) || (tc.name == "batch-error" && len(value.FailedCreatePolicies) != 1) {
						t.Fatal("typed failure data or HTTP acceptance distinction lost")
					}
				}
				// Error assertions above precede normalization for the shared runner.
				observed.Err = nil
				return observed
			}}
			r, err := contracttest.Run(t.Context(), c)
			if err != nil || !r.Passed() {
				t.Fatalf("authored package failure differs: %v %+v", err, r.Differences())
			}
		})
	}
}

// Source GetViaIdentity/RemoveViaIdentity read IConfigApiBasedCmdletsIdentity
// InputObject.PackageName at IL793/798, guard null at819, re-read at880 and use
// the same scalar-name operation at975. Getter onOk200 returns FormattedPackage;
// removal onNoContent204 has no response body task. These replies are authored.
func TestUnitPackageInputObjectAuthoredSuccessfulGetAndRemove(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name   string
		remove bool
	}{{"get", false}, {"remove", true}} {
		t.Run(tc.name, func(t *testing.T) {
			vector := requestCase{Name: "input-object-" + tc.name, Method: "GET", URL: origin + "/Teams.VerticalPackaging/package/Fixture%2FPackage%3A%40", Status: 200, Reply: authoredPackage}
			command, set := "Get-CsPolicyPackage", "GetViaIdentity"
			want := contracttest.Expected{JSON: json.RawMessage(authoredPackage), Body: []byte(authoredPackage)}
			if tc.remove {
				vector.Method, vector.URL, vector.Status, vector.Reply = "DELETE", origin+"/Teams.VerticalPackaging/package/custompackage/Fixture%2FPackage%3A%40", 204, ""
				command, set = "Remove-CsCustomPolicyPackage", "RemoveViaIdentity"
				want = contracttest.Expected{Body: []byte{}}
			}
			run(t, contracttest.Case{Name: vector.Name, Source: "msteams/teams_admin_api/policy_packages/service_protocol_test.go", Mappings: []contracttest.Mapping{{Command: command, ParameterSet: set}}, Steps: []contracttest.Exchange{discovery(t), wire(vector, command)}, Want: want, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				input := resource.Identity{PackageName: "Fixture/Package:@"}
				if tc.remove {
					value, response, err := sdk.TeamsAdminAPI.PolicyPackages.DeleteByIdentity(ctx, input)
					if err == nil && (value == nil || !value.HTTPAccepted || value.ResponseMessage != nil || value.RawFields != nil) {
						t.Fatal("InputObject removal lost its empty receipt or fabricated business data")
					}
					return contracttest.Observation{Response: response, Err: err}
				}
				value, response, err := sdk.TeamsAdminAPI.PolicyPackages.GetByIdentity(ctx, input)
				if err == nil {
					verifyAuthoredPackage(t, value)
				}
				return contracttest.Observation{Value: value, Response: response, Err: err}
			}})
		})
	}
}

// Deliberately missing/null/empty operation IDs exercise the SDK's documented
// optional-handle behavior, not a claim of a valid completed upstream job.
// The pinned batch callback is202 and its OperationId projection is a string;
// no polling route or guaranteed populated handle is inferred from that shape.
func TestUnitPackageBatchUnusableHandlesDoNotInventOperationID(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, reply string
		id          *string
		present     bool
	}{
		{"missing", `{}`, nil, false},
		{"null", `{"OperationId":null}`, nil, true},
		{"empty", `{"OperationId":""}`, new(""), true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			vector := requestCase{Method: "POST", URL: origin + "/Teams.VerticalPackaging/package/operation/batchpackage", Body: json.RawMessage(`{"PackageType":"FixturePackage","UserIds":["fixture/a","fixture/a","fixture/b"]}`), Status: 202, Reply: tc.reply}
			run(t, contracttest.Case{Name: "unusable-handle-" + tc.name, Source: "msteams/teams_admin_api/policy_packages/service_protocol_test.go", Mappings: mapping("batch-grant"), Steps: []contracttest.Exchange{discovery(t), wire(vector, "New-CsBatchPolicyPackageAssignmentOperation")}, Want: contracttest.Expected{JSON: json.RawMessage(tc.reply), Body: []byte(tc.reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.PolicyPackages.SubmitBatch(ctx, userInput())
				if err == nil {
					if value == nil || !value.HTTPAccepted || (tc.id == nil && value.OperationID != nil) || (tc.id != nil && (value.OperationID == nil || *value.OperationID != *tc.id)) {
						t.Fatal("unusable batch handle was invented or conflated with HTTP acceptance")
					}
					_, present := value.RawFields["OperationId"]
					if present != tc.present || (tc.name == "null" && string(value.RawFields["OperationId"]) != "null") {
						t.Fatal("optional-handle raw presence was lost")
					}
				}
				return contracttest.Observation{Value: value, Response: response, Err: err}
			}})
		})
	}
}
