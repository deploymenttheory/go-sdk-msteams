package cloudcalldata_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"os"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	ccd "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/cloud_call_data"
	"github.com/go-resty/resty/v2"
)

const (
	configOrigin = "https://api.interfaces.records.teams.microsoft.com"
	targetForest = "https://admin.example.invalid/"

	route      = "/Teams.InternalSupport/CloudCallDataConnection"
	getCommand = "Get-CsCloudCallDataConnection"
	newCommand = "New-CsCloudCallDataConnection"
)

func discovery(t *testing.T) testutil.Exchange {
	t.Helper()
	body, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: http.MethodGet, URL: configOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience},
		Reply: testutil.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body}}
}

// Expected request metadata follows the selected Modern helper. Existing 403
// captures supply response bytes only; they do not prove a tenant request used
// these corrected private headers or rejected them.
func wire(method, command string, status int, reply []byte) testutil.Exchange {
	switch command {
	case getCommand:
		command = "Get-CsCloudCallDataConnectionModern_Get"
	case newCommand:
		command = "New-CsCloudCallDataConnectionModern_New"
	}
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {targetForest}}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: configOrigin + route, Audience: auth.TeamsAdminAudience, Headers: h},
		Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json; charset=utf-8"}}, Body: reply}}
}

func run(t *testing.T, c testutil.Case) {
	t.Helper()

	result, err := testutil.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Fatalf("cloud call data contract %s differed: %+v", c.Name, result.Differences())
	}
}

type invocation func(context.Context, *ccd.Service) (*ccd.Connection, *resty.Response, error)

func get(ctx context.Context, s *ccd.Service) (*ccd.Connection, *resty.Response, error) {
	return s.Get(ctx)
}

func create(ctx context.Context, s *ccd.Service) (*ccd.Connection, *resty.Response, error) {
	return s.Create(ctx)
}

func TestUnitCloudCallDataGetDenialUsesAcceptanceResponse(t *testing.T) {
	t.Parallel()
	denied, err := os.ReadFile("testdata/get_403.json")
	if err != nil {
		t.Fatal(err)
	}
	run(t, testutil.Case{Name: "captured-get-forbidden-reply-replay",
		Steps: []testutil.Exchange{discovery(t), wire("GET", getCommand, 403, denied)},
		Want:  testutil.Expected{Status: 403, Body: denied, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 403, Codes: []string{"Forbidden"}}},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.CloudCallData.Get(ctx)
			if !errors.Is(err, auth.ErrForbidden) {
				t.Fatal("403 was not classified as forbidden")
			}
			return testutil.Observation{Value: value, Response: response, Err: err}
		}})
}

func TestUnitCloudCallDataCreateDenialUsesAcceptanceResponse(t *testing.T) {
	t.Parallel()
	denied, err := os.ReadFile("testdata/create_403.json")
	if err != nil {
		t.Fatal(err)
	}
	run(t, testutil.Case{Name: "captured-create-forbidden-reply-replay",
		Steps: []testutil.Exchange{discovery(t), wire("POST", newCommand, 403, denied)},
		Want:  testutil.Expected{Status: 403, Body: denied, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 403, Codes: []string{"Forbidden"}}},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.CloudCallData.Create(ctx)
			if value != nil || !errors.Is(err, auth.ErrForbidden) {
				t.Fatal("create 403 was not classified as forbidden")
			}
			return testutil.Observation{Value: value, Response: response, Err: err}
		}})
}

func TestUnitCloudCallDataReturnsTokenWithoutPrintingIt(t *testing.T) {
	t.Parallel()
	body := []byte(`[{"Token":"synthetic-token-value","Future":1}]`)
	for index, c := range []struct {
		method, command string
		status          int
		invoke          invocation
	}{
		{"GET", getCommand, 200, get},
		{"POST", newCommand, 200, create},
	} {
		run(t, testutil.Case{Name: fmt.Sprintf("token-case-%d", index),
			Steps: []testutil.Exchange{discovery(t), wire(c.method, c.command, c.status, body)},
			Want:  testutil.Expected{Status: c.status, Body: body},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				value, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.CloudCallData)
				if err != nil || value.Token == nil || *value.Token != "synthetic-token-value" || value.RawFields["Future"] == nil {
					t.Fatal("connection token or unknown field was lost")
				}
				if fmt.Sprintf("%v %+v %#v", *value, *value, *value) != "cloud call data connection (redacted) cloud call data connection (redacted) cloud call data connection (redacted)" {
					t.Fatal("connection formatting exposed the token")
				}
				return testutil.Observation{Response: response, Err: err}
			}})
	}
	null := []byte(`[{"Token":null}]`)
	run(t, testutil.Case{Name: "null-token",
		Steps: []testutil.Exchange{discovery(t), wire("GET", getCommand, 200, null)},
		Want:  testutil.Expected{Status: 200, Body: null},
		Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			value, response, err := sdk.TeamsAdminAPI.CloudCallData.Get(ctx)
			if err != nil || value.Token != nil {
				t.Fatal("null token was not preserved as nil")
			}
			return testutil.Observation{Response: response, Err: err}
		}})
}

func TestUnitCloudCallDataFailuresStayTyped(t *testing.T) {
	t.Parallel()
	for name, c := range map[string]struct {
		method, command string
		status          int
		body            string
		want            error
		invoke          invocation
	}{
		"get-object-root": {"GET", getCommand, 200, `{}`, client.ErrDecode, get},
		"get-null":        {"GET", getCommand, 200, `null`, client.ErrDecode, get},

		"create-empty": {"POST", newCommand, 200, ``, ccd.ErrCardinality, create},
		"create-201":   {"POST", newCommand, 201, `[]`, ccd.ErrResponse, create},
		"create-202":   {"POST", newCommand, 202, `{}`, ccd.ErrResponse, create},
	} {
		run(t, testutil.Case{Name: name,
			Steps: []testutil.Exchange{discovery(t), wire(c.method, c.command, c.status, []byte(c.body))},
			Want:  testutil.Expected{Status: c.status, Error: c.want},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				value, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.CloudCallData)
				if value != nil {
					t.Fatal("failed response returned a value")
				}
				return testutil.Observation{Response: response, Err: err}
			}})
	}
}

func TestUnitCloudCallDataNilContextAndServiceAreRejected(t *testing.T) {
	t.Parallel()
	var s *ccd.Service
	//nolint:staticcheck // A nil context is the guarded input.
	if _, _, err := s.Get(nil); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, _, err := s.Create(t.Context()); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, _, err := ccd.NewService(nil).Get(ctx); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled context accepted")
	}
}

func TestUnitCloudCallDataListFollowsObservedModernArrayParser(t *testing.T) {
	t.Parallel()
	for _, c := range []struct {
		name, body string
		count      int
		nilArray   bool
	}{
		{"empty-array", `[]`, 0, false}, {"empty-text", ``, 0, true}, {"object-no-token", `[{}]`, 1, false}, {"null-token", `[{"Token":null}]`, 1, false}, {"empty-token", `[{"Token":""}]`, 1, false}, {"numeric-token", `[{"Token":7}]`, 1, false}, {"wrong-case-token", `[{"token":"not-the-typed-property"}]`, 1, false}, {"nonobject-entries", `[null,false,1,"text",[]]`, 5, false}, {"multiple", `[{"Token":"first"},{"Token":"second"}]`, 2, false},
	} {
		t.Run(c.name, func(t *testing.T) {
			run(t, testutil.Case{Name: "array-" + c.name, Steps: []testutil.Exchange{discovery(t), wire("GET", getCommand, 200, []byte(c.body))}, Want: testutil.Expected{Status: 200}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				values, response, e := sdk.TeamsAdminAPI.CloudCallData.List(ctx)
				if e == nil {
					if (values == nil) != c.nilArray || len(values) != c.count {
						t.Fatal("source null/empty collection cardinality changed")
					}
					switch c.name {
					case "object-no-token", "null-token", "numeric-token", "wrong-case-token":
						if values[0] == nil || values[0].Token != nil {
							t.Fatal("source nullable/case-sensitive token projection changed")
						}
					case "empty-token":
						if values[0] == nil || values[0].Token == nil || *values[0].Token != "" {
							t.Fatal("empty token became null")
						}
					case "nonobject-entries":
						for _, v := range values {
							if v != nil {
								t.Fatal("source nonobject entry became a model")
							}
						}
					case "multiple":
						if values[0].Token == nil || *values[0].Token != "first" || values[1].Token == nil || *values[1].Token != "second" {
							t.Fatal("returned collection was truncated or reordered")
						}
					}
					if c.name == "numeric-token" && string(values[0].RawFields["Token"]) != "7" {
						t.Fatal("native raw snapshot lost unprojected token value")
					}
				}
				return testutil.Observation{Response: response, Err: e}
			}})
		})
	}
}
func TestUnitCloudCallDataCardinalityRetainsReadAndAcceptedActionOutcomes(t *testing.T) {
	t.Parallel()
	for _, body := range []string{`[]`, `[{"Token":"first"},{"Token":"second"}]`, `[null]`} {
		for _, action := range []bool{false, true} {
			method, command, invoke := "GET", getCommand, invocation(get)
			if action {
				method, command, invoke = "POST", newCommand, create
			}
			run(t, testutil.Case{Name: fmt.Sprintf("cardinality-%t-%d", action, len(body)), Steps: []testutil.Exchange{discovery(t), wire(method, command, 200, []byte(body))}, Want: testutil.Expected{Status: 200, Error: ccd.ErrCardinality}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				value, response, e := invoke(ctx, sdk.TeamsAdminAPI.CloudCallData)
				var cardinality *ccd.CardinalityError
				if value != nil || !errors.Is(e, ccd.ErrCardinality) || !errors.As(e, &cardinality) || cardinality.Response != response || response == nil {
					t.Fatal("unique-result convenience lost outcome or selected an arbitrary model")
				}
				var expected []json.RawMessage
				if json.Unmarshal([]byte(body), &expected) != nil || len(cardinality.Returned) != len(expected) {
					t.Fatal("cardinality error discarded typed returned data")
				}
				if action && (cardinality.CreateOutcome == nil || !cardinality.CreateOutcome.CreateInvoked || !cardinality.CreateOutcome.HTTPAccepted || cardinality.CreateOutcome.Response != response || len(cardinality.CreateOutcome.Returned) != len(expected)) {
					t.Fatal("accepted action error discarded full receipt/outcome")
				}
				if !action && cardinality.CreateOutcome != nil {
					t.Fatal("read was labeled a creation")
				}
				for _, v := range []any{cardinality, *cardinality, cardinality.CreateOutcome} {
					for _, format := range []string{"%v", "%+v", "%#v"} {
						text := fmt.Sprintf(format, v)
						if strings.Contains(text, "first") || strings.Contains(text, "second") {
							t.Fatal("cardinality diagnostics exposed returned tokens")
						}
					}
				}
				return testutil.Observation{Response: response, Err: e}
			}})
		}
	}
}
func TestUnitCloudCallDataCreateResultSeparatesHTTPReceiptFromSourceSuccess(t *testing.T) {
	t.Parallel()
	for _, status := range []int{200, 201} {
		var want error
		if status == 201 {
			want = ccd.ErrResponse
		}
		run(t, testutil.Case{Name: fmt.Sprintf("action-receipt-%d", status), Steps: []testutil.Exchange{discovery(t), wire("POST", newCommand, status, []byte(`[{"Token":"first"},{"Token":"second"}]`))}, Want: testutil.Expected{Status: status, Error: want}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
			result, response, e := sdk.TeamsAdminAPI.CloudCallData.CreateWithResult(ctx)
			if result == nil || !result.CreateInvoked || !result.HTTPAccepted || result.Response != response || response == nil || (status == 200 && len(result.Returned) != 2) || (status == 201 && (result.Returned != nil || !errors.Is(e, ccd.ErrResponse))) {
				t.Fatal("2xx receipt was confused with declared parsed source success")
			}
			return testutil.Observation{Response: response, Err: e}
		}})
	}
}
