package grouprelationship_test

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
	relation "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group_relationship"
	"github.com/go-resty/resty/v2"
)

const (
	groupID                             = "11111111-1111-4111-8111-111111111111"
	objectID relation.DirectoryObjectID = "22222222-2222-4222-8222-222222222222"
)

type referenceService interface {
	AddReference(context.Context, string, relation.Reference) (*resty.Response, error)
	RemoveReference(context.Context, string, relation.DirectoryObjectID) (*resty.Response, error)
	List(context.Context, string, relation.ListOptions) (*relation.ListPage, *resty.Response, error)
	NewPager(string, relation.ListOptions, client.PagingOptions) (*client.Pager[relation.DirectoryObject], error)
	ListAll(context.Context, string, relation.ListOptions, client.PagingOptions) (client.ListResult[relation.DirectoryObject], error)
}

func service(sdk *msteams.Client, kind string) referenceService {
	if kind == "members" {
		return sdk.GraphAPI.GroupMembers
	}
	return sdk.GraphAPI.GroupOwners
}

func runCase(t *testing.T, c testutil.Case) testutil.Result {
	t.Helper()
	if c.Name == "" {
		c.Name = "reference-case"
	}

	result, err := testutil.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Errorf("group reference contract differs: %+v", result.Differences())
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
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: origin + path, Audience: origin, JSON: payload, Headers: headers}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(response)}}
}

func TestUnitReferenceWritesAndFixedDeleteSuffix(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"members", "owners"} {
		defaultTarget := "directoryObjects"
		if kind == "owners" {
			defaultTarget = "users"
		}
		t.Run(kind, func(t *testing.T) {
			path := "/v1.0/groups/" + groupID + "/" + kind
			for _, target := range []relation.ReferenceKind{"", relation.Users, relation.DirectoryObjects, relation.ServicePrincipals} {
				collection := string(target)
				if collection == "" {
					collection = defaultTarget
				}
				runCase(t, testutil.Case{Name: "add-reference", Steps: []testutil.Exchange{wire("POST", auth.GraphAudience, path+"/$ref", `{"@odata.id":"`+auth.GraphAudience+`/v1.0/`+collection+`/`+string(objectID)+`"}`, 204, "")},
					Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
						response, err := service(sdk, kind).AddReference(ctx, groupID, relation.Reference{ID: objectID, Kind: target})
						return testutil.Observation{Response: response, Err: err}
					}, Want: testutil.Expected{Status: 204}})
			}
			runCase(t, testutil.Case{Name: "remove-reference", Steps: []testutil.Exchange{wire("DELETE", auth.GraphAudience, path+"/"+string(objectID)+"/$ref", "", 204, "")},
				Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					response, err := service(sdk, kind).RemoveReference(ctx, groupID, objectID)
					return testutil.Observation{Response: response, Err: err}
				}, Want: testutil.Expected{Status: 204}})
		})
	}
}

func TestUnitReferenceBindingsUseSelectedGraphCloud(t *testing.T) {
	t.Parallel()
	for _, cloud := range []string{"commercial", "gcc", "gcch", "dod", "china", "france", "germany"} {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			for _, kind := range []string{"members", "owners"} {
				t.Run(cloud+"/"+string(mode)+"/"+kind, func(t *testing.T) {
					env, err := auth.ResolveEnvironment(cloud, nil)
					if err != nil {
						t.Fatal(err)
					}
					origin := env.Endpoints.Graph
					collection := "directoryObjects"
					if kind == "owners" {
						collection = "users"
					}
					runCase(t, testutil.Case{Cloud: cloud, Mode: mode, Steps: []testutil.Exchange{wire("POST", origin, "/v1.0/groups/"+groupID+"/"+kind+"/$ref", `{"@odata.id":"`+origin+`/v1.0/`+collection+`/`+string(objectID)+`"}`, 204, "")},
						Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
							response, err := service(sdk, kind).AddReference(ctx, groupID, relation.Reference{ID: objectID})
							return testutil.Observation{Response: response, Err: err}
						}, Want: testutil.Expected{Status: 204}})
				})
			}
		}
	}
}

func TestUnitReferenceFailuresAreActualOutcomesAndNeverReplayed(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"members", "owners"} {
		for _, method := range []string{"POST", "DELETE"} {
			for _, status := range []int{400, 403, 404, 412, 429, 503} {
				t.Run(kind+"/"+method+"/"+strconv.Itoa(status), func(t *testing.T) {
					path := "/v1.0/groups/" + groupID + "/" + kind
					body := ""
					if method == "DELETE" {
						path += "/" + string(objectID)
					} else {
						collection := "directoryObjects"
						if kind == "owners" {
							collection = "users"
						}
						body = `{"@odata.id":"` + auth.GraphAudience + `/v1.0/` + collection + `/` + string(objectID) + `"}`
					}
					path += "/$ref"
					runCase(t, testutil.Case{Steps: []testutil.Exchange{wire(method, auth.GraphAudience, path, body, status, `{"error":{"code":"FixtureFailure","message":"private"}}`)},
						Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
							var response *resty.Response
							var err error
							if method == "POST" {
								response, err = service(sdk, kind).AddReference(ctx, groupID, relation.Reference{ID: objectID})
							} else {
								response, err = service(sdk, kind).RemoveReference(ctx, groupID, objectID)
							}
							return testutil.Observation{Response: response, Err: err}
						}, Want: testutil.Expected{Status: status, APIError: &client.APIError{Backend: client.Graph, StatusCode: status, Codes: []string{"FixtureFailure"}}}})
				})
			}
		}
	}
}

func TestUnitReferenceListPreservesLimitedAndFutureObjects(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"members", "owners"} {
		t.Run(kind, func(t *testing.T) {
			body := `{"value":[{"id":"` + string(objectID) + `","@odata.type":"#microsoft.graph.user","displayName":null,"userPrincipalName":null,"userType":null,"future":{"flag":false}}],"@odata.nextLink":"https://foreign.invalid/opaque","futurePage":7}`
			runCase(t, testutil.Case{Steps: []testutil.Exchange{wire("GET", auth.GraphAudience, "/v1.0/groups/"+groupID+"/"+kind+"?$select=id%2CdisplayName%2CuserPrincipalName%2CuserType", "", 200, body)},
				Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					page, response, err := service(sdk, kind).List(ctx, groupID, relation.ListOptions{Select: []string{"id", "displayName", "userPrincipalName", "userType"}})
					var data any
					if page != nil {
						data = map[string]any{"id": page.Items[0].ID, "user_type": page.Items[0].UserType, "upn": page.Items[0].UserPrincipalName, "next": page.NextLink, "raw": page.RawFields}
					}
					return testutil.Observation{Value: data, Response: response, Err: err}
				}, Want: testutil.Expected{Status: 200, JSON: json.RawMessage(`{"id":"` + string(objectID) + `","user_type":null,"upn":null,"next":"https://foreign.invalid/opaque","raw":` + body + `}`)}})
		})
	}
}

func aggregate(result client.ListResult[relation.DirectoryObject], err error) testutil.Observation {
	ids := []relation.DirectoryObjectID{}
	for _, object := range result.Items {
		ids = append(ids, object.ID)
	}
	return testutil.Observation{Value: map[string]any{"ids": ids, "pages": result.Pages}, Response: result.Response, Err: err}
}

func TestUnitReferencePagingIsOpaqueAndRetainsHeader(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"members", "owners"} {
		path := "/v1.0/groups/" + groupID + "/" + kind
		first := wire("GET", auth.GraphAudience, path+"?$count=true&$filter=displayName+eq+%27fixture%27&$top=1", "", 200, `{"value":[{"id":"`+string(objectID)+`"}],"@odata.nextLink":"`+auth.GraphAudience+path+`?$skiptoken=opaque%2B%2F%3D"}`)
		second := wire("GET", auth.GraphAudience, path+"?$skiptoken=opaque%2B%2F%3D", "", 200, `{"value":[]}`)
		first.Request.Headers.Set("ConsistencyLevel", "eventual")
		second.Request.Headers.Set("ConsistencyLevel", "eventual")
		runCase(t, testutil.Case{Name: "opaque-pages", Steps: []testutil.Exchange{first, second}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			count := true
			pager, err := service(sdk, kind).NewPager(groupID, relation.ListOptions{Filter: "displayName eq 'fixture'", Count: &count, Top: 1, EventualConsistency: true}, client.PagingOptions{MaxPages: 2, MaxItems: 1})
			if err != nil {
				return testutil.Observation{Err: err}
			}
			count = false
			return aggregate(pager.Collect(ctx))
		}, Want: testutil.Expected{Status: 200, JSON: json.RawMessage(`{"ids":["` + string(objectID) + `"],"pages":2}`)}})
	}
}

func TestUnitReferencePartialTraversalAndEnvelopeFailures(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"members", "owners"} {
		path := "/v1.0/groups/" + groupID + "/" + kind
		for _, tc := range []struct {
			next   string
			limit  int
			second bool
			status int
			want   error
		}{
			{"https://foreign.invalid/private", 0, false, 200, client.ErrDestination},
			{auth.GraphAudience + path + "?$skiptoken=next", 1, false, 200, client.ErrPageLimit},
			{auth.GraphAudience + path + "?$skiptoken=next", 0, true, 403, auth.ErrForbidden},
			{auth.GraphAudience + path + "?$skiptoken=next", 0, true, 206, client.ErrPageStatus},
		} {
			steps := []testutil.Exchange{wire("GET", auth.GraphAudience, path, "", 200, `{"value":[{"id":"`+string(objectID)+`"}],"@odata.nextLink":"`+tc.next+`"}`)}
			if tc.second {
				steps = append(steps, wire("GET", auth.GraphAudience, path+"?$skiptoken=next", "", tc.status, `{"value":[],"error":{"code":"Forbidden"}}`))
			}
			runCase(t, testutil.Case{Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				return aggregate(service(sdk, kind).ListAll(ctx, groupID, relation.ListOptions{}, client.PagingOptions{MaxPages: tc.limit}))
			},
				Want: testutil.Expected{Status: tc.status, Error: tc.want, JSON: json.RawMessage(`{"ids":["` + string(objectID) + `"],"pages":1}`)}})
		}
		for _, body := range []string{`{}`, `{"value":null}`, `{"value":[null]}`, `{"value":[{"displayName":"private"}]}`, `{"value":[],"@odata.nextLink":7}`} {
			runCase(t, testutil.Case{Steps: []testutil.Exchange{wire("GET", auth.GraphAudience, path, "", 200, body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				_, response, err := service(sdk, kind).List(ctx, groupID, relation.ListOptions{})
				return testutil.Observation{Response: response, Err: err}
			}, Want: testutil.Expected{Status: 200, Error: client.ErrDecode}})
		}
	}
}

func TestUnitReferenceValidationDoesNotAcquireCredentials(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"members", "owners"} {
		for _, action := range []string{"bad-group", "bad-object", "bad-kind", "bad-top", "bad-select", "bad-status", "wrong-backend", "closed", "cancelled", "nil-service"} {
			var want = client.ErrRequest
			switch action {
			case "bad-status":
				want = client.ErrPagination
			case "wrong-backend":
				want = client.ErrDestination
			case "closed", "nil-service":
				want = client.ErrClosed
			case "cancelled":
				want = context.Canceled
			}
			runCase(t, testutil.Case{Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				var err error
				switch action {
				case "bad-group":
					_, err = service(sdk, kind).AddReference(ctx, "../private", relation.Reference{ID: objectID})
				case "bad-object":
					_, err = service(sdk, kind).RemoveReference(ctx, groupID, "../private")
				case "bad-kind":
					_, err = service(sdk, kind).AddReference(ctx, groupID, relation.Reference{ID: objectID, Kind: "https://foreign.invalid"})
				case "bad-top":
					_, _, err = service(sdk, kind).List(ctx, groupID, relation.ListOptions{Top: -1})
				case "bad-select":
					_, _, err = service(sdk, kind).List(ctx, groupID, relation.ListOptions{Select: []string{""}})
				case "bad-status":
					_, err = service(sdk, kind).NewPager(groupID, relation.ListOptions{}, client.PagingOptions{ExpectedStatus: 206})
				case "wrong-backend":
					_, err = relation.NewMemberService(sdk.TeamsAdminAPI.Scope()).RemoveReference(ctx, groupID, objectID)
				case "closed":
					if e := sdk.Close(); e != nil {
						return testutil.Observation{Err: e}
					}
					_, err = service(sdk, kind).NewPager(groupID, relation.ListOptions{}, client.PagingOptions{})
				case "cancelled":
					cancelled, cancel := context.WithCancel(ctx)
					cancel()
					_, err = service(sdk, kind).RemoveReference(cancelled, groupID, objectID)
				case "nil-service":
					if kind == "members" {
						var s *relation.MemberService
						_, err = s.RemoveReference(ctx, groupID, objectID)
					} else {
						var s *relation.OwnerService
						_, err = s.RemoveReference(ctx, groupID, objectID)
					}
				}
				return testutil.Observation{Err: err}
			}, Want: testutil.Expected{Error: want}})
		}
	}
}

func TestUnitReferenceUnexpectedSuccessPreservesMetadata(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"members", "owners"} {
		runCase(t, testutil.Case{Steps: []testutil.Exchange{wire("DELETE", auth.GraphAudience, "/v1.0/groups/"+groupID+"/"+kind+"/"+string(objectID)+"/$ref", "", 200, "")},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				response, err := service(sdk, kind).RemoveReference(ctx, groupID, objectID)
				return testutil.Observation{Response: response, Err: err}
			}, Want: testutil.Expected{Status: 200, Error: relation.ErrUnexpectedStatus}})
	}
}
