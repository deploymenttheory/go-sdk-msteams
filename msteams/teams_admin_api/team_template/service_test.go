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

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	template "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/team_template"
)

const opaqueID = "/api/teamtemplates/v1.0/fixture/Tenant/en-US"
const configOrigin = "https://api.interfaces.records.teams.microsoft.com"
const targetForest = "https://admin.example.invalid/"

type requestVector struct {
	Name, Method, URL string
	Body              json.RawMessage
}

func requestFixture(t *testing.T, name string) requestVector {
	t.Helper()
	for _, tc := range []requestVector{
		{Name: "template-read-opaque-id", Method: "GET", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.Templates/%2Fapi%2Fteamtemplates%2Fv1.0%2Ffixture%2FTenant%2Fen-US", Body: nil},
		{Name: "template-list", Method: "GET", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.Templates/api/TeamTemplates/v1.0/en-US", Body: nil},
		{Name: "template-delete-empty", Method: "DELETE", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.Templates/%2Fapi%2Fteamtemplates%2Fv1.0%2Ffixture%2FTenant%2Fen-US", Body: nil},
		{Name: "template-create", Method: "POST", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.Templates/api/TeamTemplates/v1.0/Tenant/en-US", Body: json.RawMessage("{\"displayName\":\"Fixture\",\"shortDescription\":\"Summary\"}")},
		{Name: "template-update", Method: "PUT", URL: "https://api.interfaces.records.teams.microsoft.com/Teams.Templates/%2Fapi%2Fteamtemplates%2Fv1.0%2Ffixture%2FTenant%2Fen-US", Body: json.RawMessage("{\"displayName\":\"Fixture\",\"shortDescription\":\"Summary\"}")},
	} {
		if tc.Name == name {
			return tc
		}
	}
	t.Fatal("unknown request vector")
	return requestVector{}
}

func serviceDiscoveryFixture(t *testing.T) []byte {
	t.Helper()
	body, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func serviceCase(t *testing.T, name, fixtureName, command string, captured ...string) testutil.Case {
	t.Helper()
	fixture := requestFixture(t, fixtureName)
	requestHeaders := http.Header{}
	requestHeaders.Set("Accept", "application/json")
	requestHeaders.Set("X-MS-CmdletName", command)
	requestHeaders.Set("X-MS-Target-Uri", targetForest)
	var body json.RawMessage
	if fixture.Body != nil {
		body = fixture.Body
		requestHeaders.Set("Content-Type", "application/json")
	}
	c := testutil.Case{Name: name,

		Steps: []testutil.Exchange{
			{Request: testutil.WireRequest{Method: http.MethodGet, URL: configOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience},
				Reply: testutil.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}},
					Body: serviceDiscoveryFixture(t)}},
			{Request: testutil.WireRequest{Method: fixture.Method, URL: fixture.URL, Audience: auth.TeamsAdminAudience, Headers: requestHeaders, JSON: body},
				Reply: testutil.Reply{Status: http.StatusForbidden, Headers: http.Header{"Content-Type": {"application/json"}}, Body: nil}},
		}, Want: testutil.Expected{Status: http.StatusForbidden, Error: auth.ErrForbidden}}
	if len(captured) == 0 {
		return c
	}
	// Permanent primitive fixtures were captured by the owned template lab.
	// Their replay proves decoding/result behavior, not a new tenant execution.
	response, err := os.ReadFile("testdata/" + captured[0])
	if err != nil {
		t.Fatal(err)
	}
	status := http.StatusOK
	switch fixture.Method {
	case http.MethodPost:
		status = http.StatusCreated
	case http.MethodDelete:
		status = http.StatusNoContent
	}
	c.Steps[1].Reply.Status, c.Steps[1].Reply.Body = status, response
	c.Want = testutil.Expected{Status: status}
	if strings.HasSuffix(captured[0], ".json") {

		expected := json.RawMessage(response)
		if fixtureName == "template-read-opaque-id" {
			var envelope struct {
				Value json.RawMessage `json:"value"`
			}
			if json.Unmarshal(response, &envelope) != nil || len(envelope.Value) == 0 {
				t.Fatal("captured detail envelope unavailable")
			}
			expected = envelope.Value
		}
		c.Want.JSON = expected
	} else {
		c.Want.Body = response
	}
	return c
}

func passCase(t *testing.T, c testutil.Case) testutil.Result {
	t.Helper()
	result, err := testutil.Run(t.Context(), c)
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
	for _, fixture := range definitionCases() {
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
	// This expectation composes the definition input with the custom wrapper's
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

	for _, set := range []string{"Get", "GetViaIdentity"} {
		t.Run(set, func(t *testing.T) {
			t.Parallel()
			c := serviceCase(t, "get-"+strings.ToLower(set), "template-read-opaque-id", "Get-CsTeamTemplate", "expanded_create_readback.json")

			c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				id := opaqueID
				if set == "GetViaIdentity" {
					identity := template.Identity{ODataID: opaqueID}
					id = identity.ODataID
				}
				value, response, err := sdk.TeamsAdminAPI.TeamTemplates.GetByID(ctx, id)
				if err == nil && (value.DisplayName == nil || *value.DisplayName != "fixture-template" || value.Description == nil || *value.Description != "SDK lab non-default initial description" || value.ShortDescription == nil || *value.ShortDescription != "SDK lab non-default initial summary" || len(value.Channels) != 1 || value.Channels[0].DisplayName == nil || *value.Channels[0].DisplayName != "General") {
					t.Error("captured template detail lost typed flattened fields")
				}
				if err == nil && (value.DiscoverySettings == nil || value.DiscoverySettings.ShowInTeamsSearchAndSuggestions) {
					t.Error("detail projection lost the generated default settings")
				}
				return testutil.Observation{Value: value, Response: response, Err: err}
			}
			passCase(t, c)

		})
	}
}

func TestUnitServiceList(t *testing.T) {
	t.Parallel()

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
			c := serviceCase(t, strings.ReplaceAll(tc.name, " ", "-"), "template-list", "Get-CsTeamTemplateList", "expanded_owned_list.json")

			var options *template.ListOptions
			if tc.locale != "" {
				options = &template.ListOptions{PublicTemplateLocale: &tc.locale}
				c.Steps[1].Request.URL = strings.Replace(c.Steps[1].Request.URL, "en-US", tc.locale, 1)
			}
			if tc.incomplete {
				c.Steps[1].Reply.Headers.Set("X-Teams-TemplateSvc-Diagnostics", "fixture; templatesRETRIEVALfailure")
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				value, response, err := sdk.TeamsAdminAPI.TeamTemplates.List(ctx, options)
				if err == nil && (len(value.Items) != 16 || value.Items[0].Name == nil || *value.Items[0].Name != "fixture-template" || value.Items[0].ChannelCount == nil || *value.Items[0].ChannelCount != 1) {
					t.Error("captured template list lost typed summary fields")
				}
				if err == nil && (value.IsIncomplete != tc.incomplete || (tc.incomplete && len(value.Diagnostics) != 1)) {
					t.Error("partial collection diagnostic was lost")
				}
				return testutil.Observation{Value: value, Response: response, Err: err}
			}
			passCase(t, c)

		})
	}
}

func TestUnitServiceCreate(t *testing.T) {
	t.Parallel()

	for _, set := range []string{"New", "NewExpanded", "NewViaIdentity", "NewViaIdentityExpanded"} {
		t.Run(set, func(t *testing.T) {
			t.Parallel()
			body := set == "New" || set == "NewViaIdentity"
			definition, expected := inputDefinition(t, body)
			prefix := "expanded_"
			if body {
				prefix = "json_body_"
			}
			c := serviceCase(t, "create-"+strings.ToLower(set), "template-create", "New-CsTeamTemplate", prefix+"create_201.json")
			c.Steps[1].Request.JSON = expectedCreateBody(t, expected)

			c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				locale := "en-US"
				if strings.Contains(set, "ViaIdentity") {
					identity := template.Identity{Locale: locale}
					locale = identity.Locale
				}
				value, response, err := sdk.TeamsAdminAPI.TeamTemplates.Create(ctx, locale, definition)
				return testutil.Observation{Value: value, Response: response, Err: err}
			}
			passCase(t, c)

		})
	}
}

func TestUnitServiceUpdateByID(t *testing.T) {
	t.Parallel()

	for _, set := range []string{"Update", "UpdateExpanded", "UpdateViaIdentity", "UpdateViaIdentityExpanded"} {
		t.Run(set, func(t *testing.T) {
			t.Parallel()
			for _, status := range []int{200, 201, 403} {
				t.Run(strconv.Itoa(status), func(t *testing.T) {
					definition, expected := inputDefinition(t, set == "Update" || set == "UpdateViaIdentity")
					c := serviceCase(t, "update-"+strings.ToLower(set)+"-"+strconv.Itoa(status), "template-update", "Update-CsTeamTemplate")
					if status != 403 {
						prefix := "expanded_"
						if set == "Update" || set == "UpdateViaIdentity" {
							prefix = "json_body_"
						}
						c = serviceCase(t, c.Name, "template-update", "Update-CsTeamTemplate", prefix+"update.json")

						// Pinned TeamTemplateUpdate_Call handles200/201 with the same
						// ICreateTemplateResponse; only200 is a live capture here.

					}
					c.Steps[1].Request.JSON = expected
					c.Steps[1].Reply.Status, c.Want.Status = status, status

					c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
						id := opaqueID
						if strings.Contains(set, "ViaIdentity") {
							identity := template.Identity{ODataID: id}
							id = identity.ODataID
						}
						value, response, err := sdk.TeamsAdminAPI.TeamTemplates.UpdateByID(ctx, id, definition)
						return testutil.Observation{Value: value, Response: response, Err: err}
					}
					passCase(t, c)

				})
			}
		})
	}
}

func TestUnitServiceDeleteByID(t *testing.T) {
	t.Parallel()

	for _, set := range []string{"Delete", "DeleteViaIdentity"} {
		t.Run(set, func(t *testing.T) {
			t.Parallel()
			for _, status := range []int{200, 204, 403} {
				t.Run(strconv.Itoa(status), func(t *testing.T) {
					c := serviceCase(t, "delete-"+strings.ToLower(set)+"-"+strconv.Itoa(status), "template-delete-empty", "Remove-CsTeamTemplate")
					if status == 204 {
						c = serviceCase(t, c.Name, "template-delete-empty", "Remove-CsTeamTemplate", "expanded_delete_204.body")
					}
					c.Steps[1].Reply.Status, c.Want.Status = status, status
					if status == 200 {
						c.Steps[1].Reply.Body = []byte(`{}`)

						c.Want.Error = nil
					}
					c.Want.Body = c.Steps[1].Reply.Body
					c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
						id := opaqueID
						if set == "DeleteViaIdentity" {
							id = (template.Identity{ODataID: id}).ODataID
						}
						response, err := sdk.TeamsAdminAPI.TeamTemplates.DeleteByID(ctx, id)
						return testutil.Observation{Response: response, Err: err}
					}
					passCase(t, c)

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
			c := serviceCase(t, strings.ReplaceAll(tc.name, " ", "-"), "template-create", "New-CsTeamTemplate")
			c.Steps[1].Request.JSON = json.RawMessage(`{"displayName":"Fixture","shortDescription":"Fixture summary","channels":` + tc.expected + `}`)

			c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				value, response, err := sdk.TeamsAdminAPI.TeamTemplates.Create(ctx, "en-US", definition)
				if backing[1] != preserved || len(definition.Channels) != 1 || definition.MemberSettings != nil {
					t.Error("create mutated caller state or shared slice backing storage")
				}
				return testutil.Observation{Value: value, Response: response, Err: err}
			}
			passCase(t, c)
		})
	}
}

func operationCase(t *testing.T, operation, name string) testutil.Case {
	t.Helper()
	fixture, command := "template-read-opaque-id", "Get-CsTeamTemplate"
	switch operation {
	case "list":
		fixture, command = "template-list", "Get-CsTeamTemplateList"
	case "create":
		fixture, command = "template-create", "New-CsTeamTemplate"
	case "update":
		fixture, command = "template-update", "Update-CsTeamTemplate"
	case "delete":
		fixture, command = "template-delete-empty", "Remove-CsTeamTemplate"
	}
	c := serviceCase(t, name, fixture, command)
	c.Want.Error = nil
	if operation == "create" {
		c.Steps[1].Request.JSON = expectedCreateBody(t, c.Steps[1].Request.JSON)
	}
	c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		service := sdk.TeamsAdminAPI.TeamTemplates
		display, short := "Fixture", "Summary"
		definition := template.Definition{DisplayName: &display, ShortDescription: &short}
		switch operation {
		case "list":
			value, response, err := service.List(ctx, nil)
			return testutil.Observation{Value: value, Response: response, Err: err}
		case "create":
			value, response, err := service.Create(ctx, "en-US", definition)
			return testutil.Observation{Value: value, Response: response, Err: err}
		case "update":
			value, response, err := service.UpdateByID(ctx, opaqueID, definition)
			return testutil.Observation{Value: value, Response: response, Err: err}
		case "delete":
			response, err := service.DeleteByID(ctx, opaqueID)
			return testutil.Observation{Response: response, Err: err}
		default:
			value, response, err := service.GetByID(ctx, opaqueID)
			return testutil.Observation{Value: value, Response: response, Err: err}
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
					c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
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
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
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
		call func(context.Context, *msteams.Client) testutil.Observation
	}{
		{name: "empty id", want: template.ErrInput, call: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamTemplates.GetByID(ctx, "")
			return testutil.Observation{Value: value, Response: response, Err: err}
		}},
		{name: "dot segment", want: template.ErrInput, call: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			response, err := sdk.TeamsAdminAPI.TeamTemplates.DeleteByID(ctx, "/api/../fixture")
			return testutil.Observation{Response: response, Err: err}
		}},
		{name: "control segment", want: template.ErrInput, call: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamTemplates.UpdateByID(ctx, "fixture\n", template.Definition{})
			return testutil.Observation{Value: value, Response: response, Err: err}
		}},
		{name: "empty locale", want: template.ErrInput, call: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.TeamTemplates.Create(ctx, "", template.Definition{})
			return testutil.Observation{Value: value, Response: response, Err: err}
		}},
		{name: "bound whitespace locale", want: template.ErrInput, call: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			blank := " "
			value, response, err := sdk.TeamsAdminAPI.TeamTemplates.List(ctx, &template.ListOptions{PublicTemplateLocale: &blank})
			return testutil.Observation{Value: value, Response: response, Err: err}
		}},
		{name: "nil context", want: client.ErrNilContext, call: func(_ context.Context, sdk *msteams.Client) testutil.Observation {
			//nolint:contextcheck,staticcheck // Deliberately verifies nil context rejection before any I/O.
			value, response, err := sdk.TeamsAdminAPI.TeamTemplates.GetByID(nil, opaqueID)
			return testutil.Observation{Value: value, Response: response, Err: err}
		}},
		{name: "canceled", want: context.Canceled, call: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			canceled, cancel := context.WithCancel(ctx)
			cancel()
			value, response, err := sdk.TeamsAdminAPI.TeamTemplates.Create(canceled, "en-US", template.Definition{})
			return testutil.Observation{Value: value, Response: response, Err: err}
		}},
		{name: "closed client", want: client.ErrClosed, call: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			if err := sdk.Close(); err != nil {
				return testutil.Observation{Err: err}
			}
			value, response, err := sdk.TeamsAdminAPI.TeamTemplates.List(ctx, nil)
			return testutil.Observation{Value: value, Response: response, Err: err}
		}},
	} {
		t.Run(tc.name, func(t *testing.T) {
			t.Parallel()
			c := testutil.Case{Name: strings.ReplaceAll(tc.name, " ", "-"),
				Invoke: tc.call, Want: testutil.Expected{Error: tc.want}}
			passCase(t, c)
		})
	}
}

func TestUnitServiceListRequestUsesDefaultLocale(t *testing.T) {
	t.Parallel()
	c := serviceCase(t, "list-default-locale", "template-list", "Get-CsTeamTemplateList")
	c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.TeamTemplates.List(ctx, nil)
		return testutil.Observation{Value: v, Response: r, Err: e}
	}
	passCase(t, c)
}

func TestUnitServiceUsesSharedCollectionDiscovery(t *testing.T) {
	t.Parallel()
	c := serviceCase(t, "shared-discovery", "template-list", "Get-CsTeamTemplateList")
	policyRead := testutil.Exchange{Request: testutil.WireRequest{Method: http.MethodGet, URL: configOrigin + "/Skype.Policy/configurations/ApplicationAccessPolicy", Audience: auth.TeamsAdminAudience, Headers: http.Header{"X-MS-Target-Uri": {targetForest}}}, Reply: testutil.Reply{Status: 403}}
	c.Steps = []testutil.Exchange{c.Steps[0], policyRead, c.Steps[1]}
	c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		if _, _, e := sdk.TeamsAdminAPI.ApplicationAccessPolicies.List(ctx); !errors.Is(e, auth.ErrForbidden) {
			return testutil.Observation{Err: fmt.Errorf("expected policy denial: %w", e)}
		}
		v, r, e := sdk.TeamsAdminAPI.TeamTemplates.List(ctx, nil)
		return testutil.Observation{Value: v, Response: r, Err: e}
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
					c := serviceCase(t, "route-"+cloud+"-"+string(mode), "template-read-opaque-id", "Get-CsTeamTemplate")
					c.Cloud, c.Mode = cloud, mode
					environment, err := auth.ResolveEnvironment(cloud, nil)
					if err != nil {
						t.Fatal(err)
					}
					for index := range c.Steps {
						c.Steps[index].Request.URL = strings.Replace(c.Steps[index].Request.URL, configOrigin, environment.Endpoints.TeamsAdmin, 1)
						c.Steps[index].Request.Audience = environment.Endpoints.TeamsAdminResource
					}

					c.Invoke = func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
						value, response, err := sdk.TeamsAdminAPI.TeamTemplates.GetByID(ctx, opaqueID)
						return testutil.Observation{Value: value, Response: response, Err: err}
					}
					// This proves routing and binding, not service availability or
					// app/delegated eligibility in an actual sovereign tenant.
					passCase(t, c)
				})
			}
		})
	}
}
