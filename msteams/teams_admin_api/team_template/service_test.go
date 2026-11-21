package teamtemplate_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	template "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/team_template"
)

const opaqueID = "/api/teamtemplates/v1.0/fixture/Tenant/en-US"
const configOrigin = "https://api.interfaces.records.teams.microsoft.com"
const targetForest = "https://forest.fixture.invalid/"

type requestObservation struct {
	Name     string          `json:"name"`
	Method   string          `json:"method"`
	URL      string          `json:"generated_url"`
	Body     json.RawMessage `json:"body"`
	Status   int             `json:"response_status"`
	Response string          `json:"response_body"`
	ref      string
}

func requestFixture(t *testing.T, name string) requestObservation {
	t.Helper()
	raw, err := os.ReadFile("../../../contracts/endpoints/8.1.0/fixtures/generated-request-dispatch.json")
	if err != nil {
		t.Fatal(err)
	}
	var artifact struct {
		Version string               `json:"baseline_version"`
		Cases   []requestObservation `json:"cases"`
	}
	if json.Unmarshal(raw, &artifact) != nil || artifact.Version != "8.1.0" {
		t.Fatal("pinned request observations unavailable")
	}
	for index, item := range artifact.Cases {
		if item.Name == name {
			item.ref = "contracts/endpoints/8.1.0/fixtures/generated-request-dispatch.json#/cases/" + strconv.Itoa(index)
			return item
		}
	}
	t.Fatal("required request fixture disappeared")
	return requestObservation{}
}

func serviceCase(t *testing.T, name, fixtureName, command, set string) contracttest.Case {
	t.Helper()
	fixture := requestFixture(t, fixtureName)
	requestHeaders := http.Header{}
	requestHeaders.Set("Accept", "application/json")
	requestHeaders.Set("X-MS-CmdletName", command)
	requestHeaders.Set("X-MS-Target-Uri", targetForest)
	var body json.RawMessage
	if string(fixture.Body) != "null" {
		body = fixture.Body
		requestHeaders.Set("Content-Type", "application/json")
	}
	return contracttest.Case{Name: name, Family: "team_templates", Scope: "service", Source: fixture.ref,
		Mappings: []contracttest.Mapping{{Command: command, ParameterSet: set}},
		Steps: []contracttest.Exchange{
			{Request: contracttest.WireRequest{Method: http.MethodGet, URL: configOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience},
				Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}},
					Body: []byte(`{"Endpoints":{"AdminServiceEndpoint":"forest.fixture.invalid:444/ignored?x=y#ignored"}}`)}},
			{Request: contracttest.WireRequest{Method: fixture.Method, URL: fixture.URL, Audience: auth.TeamsAdminAudience, Headers: requestHeaders, JSON: body},
				Reply: contracttest.Reply{Status: fixture.Status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(fixture.Response)}},
		}, Want: contracttest.Expected{Status: fixture.Status}}
}

func passCase(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	result, err := contracttest.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Fatalf("service contract differed: %+v", result.Differences())
	}
	return result
}

func inputDefinition(t *testing.T, body bool) (template.Definition, json.RawMessage) {
	t.Helper()
	name := "template-expanded-sparse"
	if body {
		name = "template-json-body-default-settings"
	}
	for _, fixture := range modelObservations(t) {
		if fixture.Name != name {
			continue
		}
		if body {
			var input string
			if json.Unmarshal(fixture.Input, &input) != nil {
				t.Fatal("pinned Body input is invalid")
			}
			definition, err := template.ParseDefinition([]byte(input))
			if err != nil {
				t.Fatal(err)
			}
			return *definition, fixture.Wire
		}
		definition, err := template.NewDefinition("Fixture", "Fixture summary")
		if err != nil {
			t.Fatal(err)
		}
		return definition, fixture.Wire
	}
	t.Fatal("required model fixture disappeared")
	return template.Definition{}, nil
}

func expectedCreateBody(t *testing.T, definition json.RawMessage) json.RawMessage {
	t.Helper()
	var fields map[string]json.RawMessage
	if json.Unmarshal(definition, &fields) != nil {
		t.Fatal("expected model JSON is invalid")
	}
	// This expectation composes the captured model with the custom wrapper's
	// source-established append; it does not call the SDK normalization helper.
	fields["channels"] = json.RawMessage(`[{"id":"General","displayName":"General","isFavoriteByDefault":true}]`)
	raw, err := json.Marshal(fields)
	if err != nil {
		t.Fatal(err)
	}
	return raw
}

func TestUnitServiceGetByID(t *testing.T) {
	t.Parallel()
	evidence := collectServiceEvidence(t, "Get-CsTeamTemplate")
	for _, set := range []string{"Get", "GetViaIdentity"} {
		t.Run(set, func(t *testing.T) {
			t.Parallel()
			c := serviceCase(t, "get-"+strings.ToLower(set), "template-read-opaque-id", "Get-CsTeamTemplate", set)
			c.Want.JSON = json.RawMessage(`{"displayName":"Fixture"}`)
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				id := opaqueID
				if set == "GetViaIdentity" {
					identity := template.Identity{ODataID: opaqueID}
					id = identity.ODataID
				}
				value, response, err := sdk.TeamsAdminAPI.TeamTemplates.GetByID(ctx, id)
				if err == nil && (value.DiscoverySettings == nil || value.DiscoverySettings.ShowInTeamsSearchAndSuggestions) {
					t.Error("detail projection lost the generated default settings")
				}
				return contracttest.Observation{Value: value, Response: response, Err: err}
			}
			evidence.add(c, passCase(t, c))
		})
	}
}

func TestUnitServiceList(t *testing.T) {
	t.Parallel()
	evidence := collectServiceEvidence(t, "Get-CsTeamTemplateList")
	for _, tc := range []struct {
		name, locale string
		incomplete   bool
	}{
		{name: "default locale"},
		{name: "explicit locale", locale: "fr-FR"},
		{name: "partial collection", incomplete: true},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := serviceCase(t, strings.ReplaceAll(tc.name, " ", "-"), "template-list", "Get-CsTeamTemplateList", "DefaultLocaleOverride")
			c.Want.JSON = json.RawMessage(`{"value":[]}`)
			var options *template.ListOptions
			if tc.locale != "" {
				options = &template.ListOptions{PublicTemplateLocale: &tc.locale}
				c.Steps[1].Request.URL = strings.Replace(c.Steps[1].Request.URL, "en-US", tc.locale, 1)
			}
			if tc.incomplete {
				c.Steps[1].Reply.Headers.Set("X-Teams-TemplateSvc-Diagnostics", "fixture; templatesRETRIEVALfailure")
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.TeamTemplates.List(ctx, options)
				if err == nil && (value.IsIncomplete != tc.incomplete || (tc.incomplete && len(value.Diagnostics) != 1)) {
					t.Error("partial collection diagnostic was lost")
				}
				return contracttest.Observation{Value: value, Response: response, Err: err}
			}
			evidence.add(c, passCase(t, c))
		})
	}
}

func TestUnitServiceCreate(t *testing.T) {
	t.Parallel()
	evidence := collectServiceEvidence(t, "New-CsTeamTemplate")
	for _, set := range []string{"New", "NewExpanded", "NewViaIdentity", "NewViaIdentityExpanded"} {
		t.Run(set, func(t *testing.T) {
			t.Parallel()
			body := set == "New" || set == "NewViaIdentity"
			definition, expected := inputDefinition(t, body)
			c := serviceCase(t, "create-"+strings.ToLower(set), "template-create", "New-CsTeamTemplate", set)
			c.Steps[1].Request.JSON = expectedCreateBody(t, expected)
			c.Want.JSON = json.RawMessage(c.Steps[1].Reply.Body)
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				locale := "en-US"
				if strings.Contains(set, "ViaIdentity") {
					identity := template.Identity{Locale: locale}
					locale = identity.Locale
				}
				value, response, err := sdk.TeamsAdminAPI.TeamTemplates.Create(ctx, locale, definition)
				return contracttest.Observation{Value: value, Response: response, Err: err}
			}
			evidence.add(c, passCase(t, c))
		})
	}
}

func TestUnitServiceUpdateByID(t *testing.T) {
	t.Parallel()
	evidence := collectServiceEvidence(t, "Update-CsTeamTemplate")
	for _, set := range []string{"Update", "UpdateExpanded", "UpdateViaIdentity", "UpdateViaIdentityExpanded"} {
		t.Run(set, func(t *testing.T) {
			t.Parallel()
			for _, status := range []int{200, 201} {
				t.Run(strconv.Itoa(status), func(t *testing.T) {
					definition, expected := inputDefinition(t, set == "Update" || set == "UpdateViaIdentity")
					c := serviceCase(t, "update-"+strings.ToLower(set)+"-"+strconv.Itoa(status), "template-update", "Update-CsTeamTemplate", set)
					c.Steps[1].Request.JSON = expected
					c.Steps[1].Reply.Status, c.Want.Status = status, status
					c.Want.JSON = json.RawMessage(c.Steps[1].Reply.Body)
					c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
						id := opaqueID
						if strings.Contains(set, "ViaIdentity") {
							identity := template.Identity{ODataID: id}
							id = identity.ODataID
						}
						value, response, err := sdk.TeamsAdminAPI.TeamTemplates.UpdateByID(ctx, id, definition)
						return contracttest.Observation{Value: value, Response: response, Err: err}
					}
					evidence.add(c, passCase(t, c))
				})
			}
		})
	}
}

func TestUnitServiceDeleteByID(t *testing.T) {
	t.Parallel()
	evidence := collectServiceEvidence(t, "Remove-CsTeamTemplate")
	for _, set := range []string{"Delete", "DeleteViaIdentity"} {
		t.Run(set, func(t *testing.T) {
			t.Parallel()
			for _, status := range []int{200, 204} {
				t.Run(strconv.Itoa(status), func(t *testing.T) {
					c := serviceCase(t, "delete-"+strings.ToLower(set)+"-"+strconv.Itoa(status), "template-delete-empty", "Remove-CsTeamTemplate", set)
					c.Steps[1].Reply.Status, c.Want.Status = status, status
					if status == 200 {
						c.Steps[1].Reply.Body = []byte(`{"futureReceipt":false}`)
					}
					c.Want.Body = c.Steps[1].Reply.Body
					c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
						id := opaqueID
						if set == "DeleteViaIdentity" {
							id = (template.Identity{ODataID: id}).ODataID
						}
						response, err := sdk.TeamsAdminAPI.TeamTemplates.DeleteByID(ctx, id)
						return contracttest.Observation{Response: response, Err: err}
					}
					evidence.add(c, passCase(t, c))
				})
			}
		})
	}
}

func TestUnitServiceCreateGeneralChannelDoesNotMutateCaller(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, channel, expected string
	}{
		{name: "append on copy", channel: "Custom", expected: `[{"displayName":"Custom"},{"id":"General","displayName":"General","isFavoriteByDefault":true}]`},
		{name: "existing case insensitive general", channel: "gEnErAl", expected: `[{"displayName":"gEnErAl"}]`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			definition, _ := inputDefinition(t, false)
			backing := []*template.Channel{{DisplayName: &tc.channel}, {}}
			preserved := backing[1]
			definition.Channels = backing[:1]
			c := serviceCase(t, strings.ReplaceAll(tc.name, " ", "-"), "template-create", "New-CsTeamTemplate", "NewExpanded")
			c.Steps[1].Request.JSON = json.RawMessage(`{"displayName":"Fixture","shortDescription":"Fixture summary","channels":` + tc.expected + `}`)
			c.Want.JSON = json.RawMessage(c.Steps[1].Reply.Body)
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.TeamTemplates.Create(ctx, "en-US", definition)
				if backing[1] != preserved || len(definition.Channels) != 1 || definition.MemberSettings != nil {
					t.Error("create mutated caller state or shared slice backing storage")
				}
				return contracttest.Observation{Value: value, Response: response, Err: err}
			}
			passCase(t, c)
		})
	}
}

func operationCase(t *testing.T, operation, name string) contracttest.Case {
	t.Helper()
	fixture, command, set := "template-read-opaque-id", "Get-CsTeamTemplate", "Get"
	switch operation {
	case "list":
		fixture, command, set = "template-list", "Get-CsTeamTemplateList", "DefaultLocaleOverride"
	case "create":
		fixture, command, set = "template-create", "New-CsTeamTemplate", "NewExpanded"
	case "update":
		fixture, command, set = "template-update", "Update-CsTeamTemplate", "UpdateExpanded"
	case "delete":
		fixture, command, set = "template-delete-empty", "Remove-CsTeamTemplate", "Delete"
	}
	c := serviceCase(t, name, fixture, command, set)
	if operation == "create" {
		c.Steps[1].Request.JSON = expectedCreateBody(t, c.Steps[1].Request.JSON)
	}
	c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		service := sdk.TeamsAdminAPI.TeamTemplates
		display, short := "Fixture", "Summary"
		definition := template.Definition{DisplayName: &display, ShortDescription: &short}
		switch operation {
		case "list":
			value, response, err := service.List(ctx, nil)
			return contracttest.Observation{Value: value, Response: response, Err: err}
		case "create":
			value, response, err := service.Create(ctx, "en-US", definition)
			return contracttest.Observation{Value: value, Response: response, Err: err}
		case "update":
			value, response, err := service.UpdateByID(ctx, opaqueID, definition)
			return contracttest.Observation{Value: value, Response: response, Err: err}
		case "delete":
			response, err := service.DeleteByID(ctx, opaqueID)
			return contracttest.Observation{Response: response, Err: err}
		default:
			value, response, err := service.GetByID(ctx, opaqueID)
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}
	}
	return c
}

func TestUnitServiceErrorAndMutationReplayContracts(t *testing.T) {
	t.Parallel()
	for _, operation := range []string{"get", "list", "create", "update", "delete"} {
		t.Run(operation, func(t *testing.T) {
			t.Parallel()
			statuses := []int{403}
			if operation == "create" || operation == "update" || operation == "delete" {
				statuses = append(statuses, 503)
			}
			for _, status := range statuses {
				t.Run(strconv.Itoa(status), func(t *testing.T) {
					c := operationCase(t, operation, operation+"-failure-"+strconv.Itoa(status))
					c.Steps[1].Reply.Status, c.Want.Status = status, status
					c.Steps[1].Reply.Headers.Set("Request-Id", "fixture-request")
					c.Steps[1].Reply.Body = []byte(`{"operationId":"fixture-operation","errors":[{"code":"FixtureDenied","message":"private-message-canary","action":"private-action-canary","futureDetail":false}],"futureEnvelope":[]}`)
					c.Want.APIError = &client.APIError{Backend: client.TeamsAdmin, StatusCode: status,
						Codes: []string{"FixtureDenied"}, RequestIDs: []string{"fixture-request"}, OperationIDs: []string{"fixture-operation"}}
					if status == 403 {
						c.Want.Error = auth.ErrForbidden
					}
					invoke := c.Invoke
					c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
						observed := invoke(ctx, sdk)
						var typed *template.Error
						if !errors.As(observed.Err, &typed) || typed.Envelope == nil || len(typed.Envelope.Errors) != 1 || typed.Envelope.Errors[0].Action == nil {
							t.Error("typed template failure details were lost")
						} else {
							if _, exists := typed.Envelope.RawFields["futureEnvelope"]; !exists {
								t.Error("unknown envelope field was dropped")
							}
							formatted := fmt.Sprintf("%v %+v %#v %+v %#v", typed, typed, typed, *typed, *typed)
							for _, private := range []string{"private-message-canary", "private-action-canary", "fixture-operation", "fixture-request"} {
								if strings.Contains(formatted, private) {
									t.Error("formatted error exposed service diagnostics")
								}
							}
						}
						return observed
					}
					// Exactly discovery + one request. A replayed mutation fails the
					// rejecting transport's attempt contract, even after HTTP 503.
					passCase(t, c)
				})
			}
		})
	}
}

func TestUnitServiceMalformedSuccessAndUndeclaredStatuses(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, operation, body string
		status                int
		want                  error
	}{
		{name: "get missing value", operation: "get", status: 200, body: `{}`, want: client.ErrDecode},
		{name: "get null value", operation: "get", status: 200, body: `{"value":null}`, want: client.ErrDecode},
		{name: "get array value", operation: "get", status: 200, body: `{"value":[]}`, want: client.ErrDecode},
		{name: "list missing value", operation: "list", status: 200, body: `{}`, want: client.ErrDecode},
		{name: "list null value", operation: "list", status: 200, body: `{"value":null}`, want: client.ErrDecode},
		{name: "list null item", operation: "list", status: 200, body: `{"value":[null]}`, want: client.ErrDecode},
		{name: "create empty receipt", operation: "create", status: 201, want: client.ErrDecode},
		{name: "update invalid receipt", operation: "update", status: 200, body: `null`, want: client.ErrDecode},
		{name: "delete invalid any", operation: "delete", status: 200, body: `not-json`, want: client.ErrDecode},
		{name: "get undeclared", operation: "get", status: 202, body: `not-json`, want: template.ErrResponse},
		{name: "list undeclared", operation: "list", status: 201, body: `not-json`, want: template.ErrResponse},
		{name: "create undeclared", operation: "create", status: 200, body: `not-json`, want: template.ErrResponse},
		{name: "update undeclared", operation: "update", status: 204, want: template.ErrResponse},
		{name: "delete undeclared", operation: "delete", status: 201, body: `not-json`, want: template.ErrResponse},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := operationCase(t, tc.operation, strings.ReplaceAll(tc.name, " ", "-"))
			c.Steps[1].Reply.Status, c.Want.Status = tc.status, tc.status
			c.Steps[1].Reply.Body, c.Want.Error = []byte(tc.body), tc.want
			passCase(t, c)
		})
	}
}

func TestUnitServiceMalformedErrorsKeepHTTPClassification(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, body string }{
		{name: "empty"}, {name: "text", body: "not-json"}, {name: "array", body: "[]"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := operationCase(t, "get", "malformed-error-"+tc.name)
			c.Steps[1].Reply.Status, c.Want.Status, c.Want.Error = 403, 403, auth.ErrForbidden
			c.Steps[1].Reply.Body = []byte(tc.body)
			invoke := c.Invoke
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				observed := invoke(ctx, sdk)
				var typed *template.Error
				if !errors.As(observed.Err, &typed) || typed.Envelope != nil {
					t.Error("malformed error body replaced HTTP classification or became a valid envelope")
				}
				return observed
			}
			passCase(t, c)
		})
	}
}

func TestUnitServiceValidationCancellationAndCloseHaveNoIO(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name string
		want error
		call func(context.Context, *msteams.Client) contracttest.Observation
	}{
		{name: "empty id", want: template.ErrInput, call: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamTemplates.GetByID(ctx, "")
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}},
		{name: "dot segment", want: template.ErrInput, call: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			response, err := sdk.TeamsAdminAPI.TeamTemplates.DeleteByID(ctx, "/api/../fixture")
			return contracttest.Observation{Response: response, Err: err}
		}},
		{name: "control segment", want: template.ErrInput, call: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamTemplates.UpdateByID(ctx, "fixture\n", template.Definition{})
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}},
		{name: "empty locale", want: template.ErrInput, call: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamTemplates.Create(ctx, "", template.Definition{})
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}},
		{name: "bound whitespace locale", want: template.ErrInput, call: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			blank := " "
			value, response, err := sdk.TeamsAdminAPI.TeamTemplates.List(ctx, &template.ListOptions{PublicTemplateLocale: &blank})
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}},
		{name: "nil context", want: client.ErrNilContext, call: func(_ context.Context, sdk *msteams.Client) contracttest.Observation {
			//nolint:contextcheck,staticcheck // Deliberately verifies nil context rejection before any I/O.
			value, response, err := sdk.TeamsAdminAPI.TeamTemplates.GetByID(nil, opaqueID)
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}},
		{name: "canceled", want: context.Canceled, call: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			canceled, cancel := context.WithCancel(ctx)
			cancel()
			value, response, err := sdk.TeamsAdminAPI.TeamTemplates.Create(canceled, "en-US", template.Definition{})
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}},
		{name: "closed client", want: client.ErrClosed, call: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			if err := sdk.Close(); err != nil {
				return contracttest.Observation{Err: err}
			}
			value, response, err := sdk.TeamsAdminAPI.TeamTemplates.List(ctx, nil)
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := contracttest.Case{Name: strings.ReplaceAll(tc.name, " ", "-"), Family: "team_templates", Scope: "service",
				Source: "contracts/endpoints/8.1.0/core.json#/commands", Invoke: tc.call, Want: contracttest.Expected{Error: tc.want}}
			passCase(t, c)
		})
	}
}

func TestUnitServiceListRetainsUnknownEnvelopeAndSummaryFields(t *testing.T) {
	t.Parallel()
	c := serviceCase(t, "list-unknown-fields", "template-list", "Get-CsTeamTemplateList", "DefaultLocaleOverride")
	c.Steps[1].Reply.Body = []byte(`{"value":[{"id":"fixture","name":"Fixture","appCount":0,"futureSummary":false}],"futureEnvelope":{"values":[]}}`)
	c.Want.JSON = json.RawMessage(c.Steps[1].Reply.Body)
	c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		value, response, err := sdk.TeamsAdminAPI.TeamTemplates.List(ctx, nil)
		return contracttest.Observation{Value: value, Response: response, Err: err}
	}
	passCase(t, c)
}

func TestUnitServiceUsesSharedCollectionDiscovery(t *testing.T) {
	t.Parallel()
	c := serviceCase(t, "collection-discovery-sharing", "template-list", "Get-CsTeamTemplateList", "DefaultLocaleOverride")
	policyRead := contracttest.Exchange{Request: contracttest.WireRequest{Method: http.MethodGet,
		URL: configOrigin + "/Skype.Policy/configurations/ApplicationAccessPolicy", Audience: auth.TeamsAdminAudience,
		Headers: http.Header{"X-MS-Target-Uri": {targetForest}}},
		Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(`[]`)}}
	c.Steps = []contracttest.Exchange{c.Steps[0], policyRead, c.Steps[1]}
	c.Want.JSON = json.RawMessage(`{"value":[]}`)
	c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		if _, response, err := sdk.TeamsAdminAPI.ApplicationAccessPolicies.List(ctx); err != nil {
			return contracttest.Observation{Response: response, Err: err}
		}
		value, response, err := sdk.TeamsAdminAPI.TeamTemplates.List(ctx, nil)
		return contracttest.Observation{Value: value, Response: response, Err: err}
	}
	passCase(t, c)
}

func TestUnitServiceCloudAndModeRouting(t *testing.T) {
	t.Parallel()
	for _, cloud := range []string{"commercial", "gcc", "gcch", "dod", "china", "france", "germany"} {
		t.Run(cloud, func(t *testing.T) {
			t.Parallel()
			for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
				t.Run(string(mode), func(t *testing.T) {
					c := serviceCase(t, "route-"+cloud+"-"+string(mode), "template-read-opaque-id", "Get-CsTeamTemplate", "Get")
					c.Cloud, c.Mode = cloud, mode
					environment, err := auth.ResolveEnvironment(cloud, nil)
					if err != nil {
						t.Fatal(err)
					}
					for index := range c.Steps {
						c.Steps[index].Request.URL = strings.Replace(c.Steps[index].Request.URL, configOrigin, environment.Endpoints.TeamsAdmin, 1)
						c.Steps[index].Request.Audience = environment.Endpoints.TeamsAdminResource
					}
					c.Want.JSON = json.RawMessage(`{"displayName":"Fixture"}`)
					c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
						value, response, err := sdk.TeamsAdminAPI.TeamTemplates.GetByID(ctx, opaqueID)
						return contracttest.Observation{Value: value, Response: response, Err: err}
					}
					// This proves routing and binding, not service availability or
					// app/delegated eligibility in an actual sovereign tenant.
					passCase(t, c)
				})
			}
		})
	}
}
