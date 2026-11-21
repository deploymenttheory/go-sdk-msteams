package grouprelationship_test

import (
	"context"
	"encoding/json"
	"net/http"
	"strconv"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
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

func runCase(t *testing.T, c contracttest.Case) contracttest.Result {
	t.Helper()
	if c.Name == "" {
		c.Name = "reference-case"
	}
	c.Family = "group_relationships"
	if c.Scope == "" {
		c.Scope = "service"
	}
	if c.Source == "" {
		c.Source = "contracts/endpoints/8.1.0/core.json"
	}
	result, err := contracttest.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Errorf("group reference contract differs: %+v", result.Differences())
	}
	return result
}

func wire(method, origin, path, body string, status int, response string) contracttest.Exchange {
	var payload json.RawMessage
	if body != "" {
		payload = json.RawMessage(body)
	}
	headers := http.Header{"Accept": {"application/json"}}
	if body != "" {
		headers.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: origin, JSON: payload, Headers: headers}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(response)}}
}

func TestReferenceWritesAndFixedDeleteSuffix(t *testing.T) {
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
				runCase(t, contracttest.Case{Name: "add-reference", Steps: []contracttest.Exchange{wire("POST", auth.GraphAudience, path+"/$ref", `{"@odata.id":"`+auth.GraphAudience+`/v1.0/`+collection+`/`+string(objectID)+`"}`, 204, "")},
					Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
						response, err := service(sdk, kind).AddReference(ctx, groupID, relation.Reference{ID: objectID, Kind: target})
						return contracttest.Observation{Response: response, Err: err}
					}, Want: contracttest.Expected{Status: 204}})
			}
			runCase(t, contracttest.Case{Name: "remove-reference", Steps: []contracttest.Exchange{wire("DELETE", auth.GraphAudience, path+"/"+string(objectID)+"/$ref", "", 204, "")},
				Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					response, err := service(sdk, kind).RemoveReference(ctx, groupID, objectID)
					return contracttest.Observation{Response: response, Err: err}
				}, Want: contracttest.Expected{Status: 204}})
		})
	}
}

func TestReferenceBindingsUseSelectedGraphCloud(t *testing.T) {
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
					runCase(t, contracttest.Case{Cloud: cloud, Mode: mode, Steps: []contracttest.Exchange{wire("POST", origin, "/v1.0/groups/"+groupID+"/"+kind+"/$ref", `{"@odata.id":"`+origin+`/v1.0/`+collection+`/`+string(objectID)+`"}`, 204, "")},
						Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
							response, err := service(sdk, kind).AddReference(ctx, groupID, relation.Reference{ID: objectID})
							return contracttest.Observation{Response: response, Err: err}
						}, Want: contracttest.Expected{Status: 204}})
				})
			}
		}
	}
}

func TestReferenceFailuresAreActualOutcomesAndNeverReplayed(t *testing.T) {
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
					runCase(t, contracttest.Case{Steps: []contracttest.Exchange{wire(method, auth.GraphAudience, path, body, status, `{"error":{"code":"FixtureFailure","message":"private"}}`)},
						Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
							var response *resty.Response
							var err error
							if method == "POST" {
								response, err = service(sdk, kind).AddReference(ctx, groupID, relation.Reference{ID: objectID})
							} else {
								response, err = service(sdk, kind).RemoveReference(ctx, groupID, objectID)
							}
							return contracttest.Observation{Response: response, Err: err}
						}, Want: contracttest.Expected{Status: status, APIError: &client.APIError{Backend: client.Graph, StatusCode: status, Codes: []string{"FixtureFailure"}}}})
				})
			}
		}
	}
}

func TestReferenceListPreservesLimitedAndFutureObjects(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"members", "owners"} {
		t.Run(kind, func(t *testing.T) {
			body := `{"value":[{"id":"` + string(objectID) + `","@odata.type":"#microsoft.graph.user","displayName":null,"userPrincipalName":null,"userType":null,"future":{"flag":false}}],"@odata.nextLink":"https://foreign.invalid/opaque","futurePage":7}`
			runCase(t, contracttest.Case{Steps: []contracttest.Exchange{wire("GET", auth.GraphAudience, "/v1.0/groups/"+groupID+"/"+kind+"?$select=id%2CdisplayName%2CuserPrincipalName%2CuserType", "", 200, body)},
				Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
					page, response, err := service(sdk, kind).List(ctx, groupID, relation.ListOptions{Select: []string{"id", "displayName", "userPrincipalName", "userType"}})
					var data any
					if page != nil {
						data = map[string]any{"id": page.Items[0].ID, "user_type": page.Items[0].UserType, "upn": page.Items[0].UserPrincipalName, "next": page.NextLink, "raw": page.RawFields}
					}
					return contracttest.Observation{Value: data, Response: response, Err: err}
				}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"id":"` + string(objectID) + `","user_type":null,"upn":null,"next":"https://foreign.invalid/opaque","raw":` + body + `}`)}})
		})
	}
}

func aggregate(result client.ListResult[relation.DirectoryObject], err error) contracttest.Observation {
	ids := []relation.DirectoryObjectID{}
	for _, object := range result.Items {
		ids = append(ids, object.ID)
	}
	return contracttest.Observation{Value: map[string]any{"ids": ids, "pages": result.Pages}, Response: result.Response, Err: err}
}

func TestReferencePagingIsOpaqueAndRetainsHeader(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"members", "owners"} {
		path := "/v1.0/groups/" + groupID + "/" + kind
		first := wire("GET", auth.GraphAudience, path+"?$count=true&$filter=displayName+eq+%27fixture%27&$top=1", "", 200, `{"value":[{"id":"`+string(objectID)+`"}],"@odata.nextLink":"`+auth.GraphAudience+path+`?$skiptoken=opaque%2B%2F%3D"}`)
		second := wire("GET", auth.GraphAudience, path+"?$skiptoken=opaque%2B%2F%3D", "", 200, `{"value":[]}`)
		first.Request.Headers.Set("ConsistencyLevel", "eventual")
		second.Request.Headers.Set("ConsistencyLevel", "eventual")
		runCase(t, contracttest.Case{Name: "opaque-pages", Steps: []contracttest.Exchange{first, second}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			count := true
			pager, err := service(sdk, kind).NewPager(groupID, relation.ListOptions{Filter: "displayName eq 'fixture'", Count: &count, Top: 1, EventualConsistency: true}, client.PagingOptions{MaxPages: 2, MaxItems: 1})
			if err != nil {
				return contracttest.Observation{Err: err}
			}
			count = false
			return aggregate(pager.Collect(ctx))
		}, Want: contracttest.Expected{Status: 200, JSON: json.RawMessage(`{"ids":["` + string(objectID) + `"],"pages":2}`)}})
	}
}

func TestReferencePartialTraversalAndEnvelopeFailures(t *testing.T) {
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
			steps := []contracttest.Exchange{wire("GET", auth.GraphAudience, path, "", 200, `{"value":[{"id":"`+string(objectID)+`"}],"@odata.nextLink":"`+tc.next+`"}`)}
			if tc.second {
				steps = append(steps, wire("GET", auth.GraphAudience, path+"?$skiptoken=next", "", tc.status, `{"value":[],"error":{"code":"Forbidden"}}`))
			}
			runCase(t, contracttest.Case{Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				return aggregate(service(sdk, kind).ListAll(ctx, groupID, relation.ListOptions{}, client.PagingOptions{MaxPages: tc.limit}))
			},
				Want: contracttest.Expected{Status: tc.status, Error: tc.want, JSON: json.RawMessage(`{"ids":["` + string(objectID) + `"],"pages":1}`)}})
		}
		for _, body := range []string{`{}`, `{"value":null}`, `{"value":[null]}`, `{"value":[{"displayName":"private"}]}`, `{"value":[],"@odata.nextLink":7}`} {
			runCase(t, contracttest.Case{Steps: []contracttest.Exchange{wire("GET", auth.GraphAudience, path, "", 200, body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				_, response, err := service(sdk, kind).List(ctx, groupID, relation.ListOptions{})
				return contracttest.Observation{Response: response, Err: err}
			}, Want: contracttest.Expected{Status: 200, Error: client.ErrDecode}})
		}
	}
}

func TestReferenceValidationDoesNotAcquireCredentials(t *testing.T) {
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
			runCase(t, contracttest.Case{Scope: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
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
						return contracttest.Observation{Err: e}
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
				return contracttest.Observation{Err: err}
			}, Want: contracttest.Expected{Error: want}})
		}
	}
}

func TestReferenceUnexpectedSuccessPreservesMetadata(t *testing.T) {
	t.Parallel()
	for _, kind := range []string{"members", "owners"} {
		runCase(t, contracttest.Case{Steps: []contracttest.Exchange{wire("DELETE", auth.GraphAudience, "/v1.0/groups/"+groupID+"/"+kind+"/"+string(objectID)+"/$ref", "", 200, `{}`)},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				response, err := service(sdk, kind).RemoveReference(ctx, groupID, objectID)
				return contracttest.Observation{Response: response, Err: err}
			}, Want: contracttest.Expected{Status: 200, Error: relation.ErrUnexpectedStatus}})
	}
}
