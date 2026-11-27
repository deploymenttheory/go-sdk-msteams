package group_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
	"github.com/go-resty/resty/v2"
)

const (
	groupID = "11111111-1111-4111-8111-111111111111"
	ownerID = "00000000-0000-0000-0000-000000000001"
)

func runCase(t *testing.T, c testutil.Case) testutil.Result {
	t.Helper()
	if c.Name == "" {
		c.Name = "native-case"
	}

	result, err := testutil.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Errorf("backing-group contract differs: %+v", result.Differences())
	}
	return result
}

func wire(method, origin, path, body string, status int, response string) testutil.Exchange {
	var payload json.RawMessage
	if body != "" {
		payload = json.RawMessage(body)
	}
	headers := http.Header{"Accept": {"application/json"}}
	if body != "" {
		headers.Set("Content-Type", "application/json")
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: origin + path, Audience: origin, Headers: headers, JSON: payload}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(response)}}
}

func observed(value *group.Group, response *resty.Response, err error) testutil.Observation {
	var data any
	if value != nil {
		data = map[string]any{"id": value.ID, "display_name": value.DisplayName, "security_enabled": value.SecurityEnabled, "raw": value.RawFields}
	}
	return testutil.Observation{Value: data, Response: response, Err: err}
}

func requestJSON(t *testing.T, name string) json.RawMessage {
	t.Helper()
	switch name {
	case "group-patch-explicit-empty":
		return json.RawMessage(`{"classification":"","description":"","displayName":"","mailNickname":"","visibility":""}`)
	case "group-patch-omitted":
		return json.RawMessage(`{}`)
	case "ordinary-group-create":
		return json.RawMessage(`{"creationOptions":["ExchangeProvisioningFlags:3552"],"displayName":"Fixture Team","groupTypes":["Unified"],"mailEnabled":true,"mailNickname":"fixture-team","members@odata.bind":["https://graph.microsoft.com/v1.0/users/00000000-0000-0000-0000-000000000001"],"owners@odata.bind":["https://graph.microsoft.com/v1.0/users/00000000-0000-0000-0000-000000000001"],"securityEnabled":false,"visibility":"Private"}`)
	default:
		t.Fatal("unknown group request vector")
		return nil
	}
}

func TestUnitGroupCreateRequestPayload(t *testing.T) {
	t.Parallel()
	body := requestJSON(t, "ordinary-group-create")
	response := `{"id":"` + groupID + `","displayName":"Fixture Team","securityEnabled":false,"future":{"value":7}}`
	runCase(t, testutil.Case{Name: "ordinary-create-request", Steps: []testutil.Exchange{wire("POST", auth.GraphAudience, "/v1.0/groups", string(body), 201, response)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, err := group.NewMicrosoft365Group("Fixture Team", "fixture-team")
			if err != nil {
				return testutil.Observation{Err: err}
			}
			visibility := group.Private
			value.Visibility = &visibility
			value.OwnerIDs = []string{ownerID}
			value.MemberIDs = []string{ownerID}
			flags := []string{"ExchangeProvisioningFlags:3552"}
			value.CreationOptions = &flags
			return observed(sdk.GraphAPI.Groups.Create(ctx, value))
		}, Want: testutil.Expected{Status: 201, JSON: json.RawMessage(`{"id":"` + groupID + `","display_name":"Fixture Team","security_enabled":false,"raw":` + response + `}`)}})
}

func TestUnitGroupNativeCRUDSparsePatchAndNoImplicitReads(t *testing.T) {
	t.Parallel()
	path := "/v1.0/groups/" + groupID
	runCase(t, testutil.Case{Name: "native-get", Steps: []testutil.Exchange{wire("GET", auth.GraphAudience, path+"?$select=id%2CdisplayName%2CsecurityEnabled", "", 200, `{"id":"`+groupID+`","displayName":null,"securityEnabled":false}`)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			return observed(sdk.GraphAPI.Groups.GetByID(ctx, groupID, group.ReadOptions{Select: []string{"id", "displayName", "securityEnabled"}}))
		},
		Want: testutil.Expected{Status: 200, JSON: json.RawMessage(`{"id":"` + groupID + `","display_name":null,"security_enabled":false,"raw":{"id":"` + groupID + `","displayName":null,"securityEnabled":false}}`)}})
	for _, name := range []string{"group-patch-omitted", "group-patch-explicit-empty"} {
		t.Run(name, func(t *testing.T) {
			value := group.UpdateRequest{}
			if name == "group-patch-explicit-empty" {
				empty := ""
				visibility := group.Visibility("")
				value = group.UpdateRequest{DisplayName: &empty, Description: &empty, MailNickname: &empty, Classification: &empty, Visibility: &visibility}
			}
			runCase(t, testutil.Case{Name: name, Steps: []testutil.Exchange{wire("PATCH", auth.GraphAudience, path, string(requestJSON(t, name)), 204, "")},
				Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					response, err := sdk.GraphAPI.Groups.UpdateByID(ctx, groupID, value)

					return testutil.Observation{Response: response, Err: err}
				}, Want: testutil.Expected{Status: 204}})
		})
	}
	runCase(t, testutil.Case{Name: "native-delete", Steps: []testutil.Exchange{wire("DELETE", auth.GraphAudience, path, "", 204, "")},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			response, err := sdk.GraphAPI.Groups.DeleteByID(ctx, groupID)

			return testutil.Observation{Response: response, Err: err}
		}, Want: testutil.Expected{Status: 204}})
}

func TestUnitGroupCreateCloudBindingsAndExplicitEmptyReferences(t *testing.T) {
	t.Parallel()
	for _, cloud := range []string{"commercial", "gcc", "gcch", "dod", "china", "france", "germany"} {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			t.Run(cloud+"/"+string(mode), func(t *testing.T) {
				env, err := auth.ResolveEnvironment(cloud, nil)
				if err != nil {
					t.Fatal(err)
				}
				origin := env.Endpoints.Graph
				body := `{"displayName":"fixture","mailNickname":"fixture","groupTypes":["Unified"],"mailEnabled":true,"securityEnabled":false,"owners@odata.bind":["` + origin + `/v1.0/users/` + ownerID + `"],"members@odata.bind":[],"creationOptions":[]}`
				runCase(t, testutil.Case{Cloud: cloud, Mode: mode, Steps: []testutil.Exchange{wire("POST", origin, "/v1.0/groups", body, 201, `{"id":"`+groupID+`"}`)},
					Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
						value, err := group.NewMicrosoft365Group("fixture", "fixture")
						if err != nil {
							return testutil.Observation{Err: err}
						}
						value.OwnerIDs = []string{ownerID}
						value.MemberIDs = []string{}
						flags := []string{}
						value.CreationOptions = &flags
						_, response, err := sdk.GraphAPI.Groups.Create(ctx, value)

						return testutil.Observation{Response: response, Err: err}
					}, Want: testutil.Expected{Status: 201}})
			})
		}
	}
}

func TestUnitGroupCreateRetainsExplicitEducationExtension(t *testing.T) {
	t.Parallel()
	const extension = "extension_fe2174665583431c953114ff7268b7b3_Education_ObjectType"
	body := `{"displayName":"fixture","mailNickname":"fixture","groupTypes":["Unified"],"mailEnabled":true,"securityEnabled":false,"` + extension + `":"Section"}`
	runCase(t, testutil.Case{Name: "explicit-extension", Steps: []testutil.Exchange{wire("POST", auth.GraphAudience, "/v1.0/groups", body, 201, `{"id":"`+groupID+`"}`)},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, err := group.NewMicrosoft365Group("fixture", "fixture")
			if err != nil {
				return testutil.Observation{Err: err}
			}
			value.Extensions = map[string]json.RawMessage{extension: json.RawMessage(`"Section"`)}
			_, response, err := sdk.GraphAPI.Groups.Create(ctx, value)

			return testutil.Observation{Response: response, Err: err}
		}, Want: testutil.Expected{Status: 201}})
}

func TestUnitGroupValidationIsLocal(t *testing.T) {
	t.Parallel()
	for _, action := range []string{"missing-name", "bad-owner", "too-many-links", "extension-overwrite", "bad-extension-json", "bad-group", "bad-select", "wrong-backend", "closed", "cancelled", "nil-service"} {
		t.Run(action, func(t *testing.T) {
			var want = client.ErrRequest
			switch action {
			case "wrong-backend":
				want = client.ErrDestination
			case "closed", "nil-service":
				want = client.ErrClosed
			case "cancelled":
				want = context.Canceled
			}
			runCase(t, testutil.Case{Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				value, _ := group.NewMicrosoft365Group("fixture", "fixture")
				var err error
				switch action {
				case "bad-group":
					_, _, err = sdk.GraphAPI.Groups.GetByID(ctx, "../private", group.ReadOptions{})
				case "bad-select":
					_, _, err = sdk.GraphAPI.Groups.GetByID(ctx, groupID, group.ReadOptions{Select: []string{""}})
				case "wrong-backend":
					_, _, err = group.NewService(sdk.TeamsAdminAPI.Scope()).GetByID(ctx, groupID, group.ReadOptions{})
				case "closed":
					if e := sdk.Close(); e != nil {
						return testutil.Observation{Err: e}
					}
					_, _, err = sdk.GraphAPI.Groups.GetByID(ctx, groupID, group.ReadOptions{})
				case "cancelled":
					cancelled, cancel := context.WithCancel(ctx)
					cancel()
					_, _, err = sdk.GraphAPI.Groups.Create(cancelled, value)
				case "nil-service":
					var service *group.Service
					_, _, err = service.GetByID(ctx, groupID, group.ReadOptions{})
				default:
					switch action {
					case "missing-name":
						value.DisplayName = nil
					case "bad-owner":
						value.OwnerIDs = []string{"../private"}
					case "too-many-links":
						for range 21 {
							value.OwnerIDs = append(value.OwnerIDs, ownerID)
						}
					case "extension-overwrite":
						value.Extensions = map[string]json.RawMessage{"displayName": json.RawMessage(`"private"`)}
					case "bad-extension-json":
						value.Extensions = map[string]json.RawMessage{"extension_fe2174665583431c953114ff7268b7b3_Education_ObjectType": json.RawMessage(`{`)}
					}
					_, _, err = sdk.GraphAPI.Groups.Create(ctx, value)
				}
				return testutil.Observation{Err: err}
			}, Want: testutil.Expected{Error: want}})
		})
	}
}

func TestUnitGroupFailuresPreserveMetadataAndDoNotReplayMutations(t *testing.T) {
	t.Parallel()
	for _, method := range []string{"POST", "PATCH", "DELETE"} {
		for _, status := range []int{403, 404, 429, 503} {
			t.Run(method+"/"+strconv.Itoa(status), func(t *testing.T) {
				path := "/v1.0/groups"
				body := ""
				if method != "POST" {
					path += "/" + groupID
				}
				if method == "POST" {
					body = `{"displayName":"fixture","mailNickname":"fixture","groupTypes":["Unified"],"mailEnabled":true,"securityEnabled":false}`
				}
				if method == "PATCH" {
					body = `{}`
				}
				runCase(t, testutil.Case{Steps: []testutil.Exchange{wire(method, auth.GraphAudience, path, body, status, `{"error":{"code":"FixtureError","message":"private"}}`)},
					Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
						var response *resty.Response
						var err error
						switch method {
						case "POST":
							value, _ := group.NewMicrosoft365Group("fixture", "fixture")
							_, response, err = sdk.GraphAPI.Groups.Create(ctx, value)
						case "PATCH":
							response, err = sdk.GraphAPI.Groups.UpdateByID(ctx, groupID, group.UpdateRequest{})
						case "DELETE":
							response, err = sdk.GraphAPI.Groups.DeleteByID(ctx, groupID)
						}
						return testutil.Observation{Response: response, Err: err}
					}, Want: testutil.Expected{Status: status, APIError: &client.APIError{Backend: client.Graph, StatusCode: status, Codes: []string{"FixtureError"}}}})
			})
		}
	}
	for _, tc := range []struct {
		name, body string
		status     int
		want       error
	}{{"missing-id", `{}`, 200, client.ErrDecode}, {"malformed", `{"id":`, 200, client.ErrDecode}, {"unexpected-status", `{"id":"` + groupID + `"}`, 202, group.ErrUnexpectedStatus}} {
		t.Run(tc.name, func(t *testing.T) {
			runCase(t, testutil.Case{Steps: []testutil.Exchange{wire("GET", auth.GraphAudience, "/v1.0/groups/"+groupID, "", tc.status, tc.body)},
				Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					return observed(sdk.GraphAPI.Groups.GetByID(ctx, groupID, group.ReadOptions{}))
				}, Want: testutil.Expected{Status: tc.status, Error: tc.want}})
		})
	}
}
