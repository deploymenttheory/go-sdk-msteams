package cloudcalldata_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	ccd "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/cloud_call_data"
	"github.com/go-resty/resty/v2"
)

const (
	configOrigin = "https://api.interfaces.records.teams.microsoft.com"
	targetForest = "https://admin.example.invalid/"
	source       = "msteams/teams_admin_api/cloud_call_data/service_test.go"
	route        = "/Teams.InternalSupport/CloudCallDataConnection"
	getCommand   = "Get-CsCloudCallDataConnection"
	newCommand   = "New-CsCloudCallDataConnection"
)

func discovery(t *testing.T) contracttest.Exchange {
	t.Helper()
	body, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: http.MethodGet, URL: configOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience},
		Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body}}
}

func wire(method, command string, status int, reply []byte) contracttest.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {targetForest}}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: configOrigin + route, Audience: auth.TeamsAdminAudience, Headers: h},
		Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json; charset=utf-8"}}, Body: reply}}
}

func run(t *testing.T, c contracttest.Case) {
	t.Helper()
	c.Family = "cloud_call_data"
	if c.Source == "" {
		c.Source = source
	}
	if c.Scope == "" {
		c.Scope = "service"
	}
	result, err := contracttest.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Fatalf("cloud call data contract %s differed: %+v", c.Name, result.Differences())
	}
}

func mapping(command string) []contracttest.Mapping {
	return []contracttest.Mapping{{Command: command, ParameterSet: "__AllParameterSets"}}
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
	run(t, contracttest.Case{Name: "actual-get-forbidden", Source: "msteams/teams_admin_api/cloud_call_data/testdata/get_403.json", Mappings: mapping(getCommand),
		Steps: []contracttest.Exchange{discovery(t), wire("GET", getCommand, 403, denied)},
		Want:  contracttest.Expected{Status: 403, Body: denied, APIError: &client.APIError{Backend: client.TeamsAdmin, StatusCode: 403, Codes: []string{"Forbidden"}}},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.CloudCallData.Get(ctx)
			if !errors.Is(err, auth.ErrForbidden) {
				t.Fatal("403 was not classified as forbidden")
			}
			return contracttest.Observation{Value: value, Response: response, Err: err}
		}})
}

func TestUnitCloudCallDataReturnsTokenWithoutPrintingIt(t *testing.T) {
	t.Parallel()
	body := []byte(`{"Token":"synthetic-token-value","Future":1}`)
	for index, c := range []struct {
		method, command string
		status          int
		invoke          invocation
	}{
		{"GET", getCommand, 200, get},
		{"POST", newCommand, 200, create},
		{"POST", newCommand, 201, create},
	} {
		run(t, contracttest.Case{Name: fmt.Sprintf("token-case-%d", index), Mappings: mapping(c.command),
			Steps: []contracttest.Exchange{discovery(t), wire(c.method, c.command, c.status, body)},
			Want:  contracttest.Expected{Status: c.status, Body: body},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.CloudCallData)
				if err != nil || value.Token == nil || *value.Token != "synthetic-token-value" || value.RawFields["Future"] == nil {
					t.Fatal("connection token or unknown field was lost")
				}
				if fmt.Sprintf("%v %+v %#v", *value, *value, *value) != "cloud call data connection (redacted) cloud call data connection (redacted) cloud call data connection (redacted)" {
					t.Fatal("connection formatting exposed the token")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
	}
	null := []byte(`{"Token":null}`)
	run(t, contracttest.Case{Name: "null-token", Mappings: mapping(getCommand),
		Steps: []contracttest.Exchange{discovery(t), wire("GET", getCommand, 200, null)},
		Want:  contracttest.Expected{Status: 200, Body: null},
		Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			value, response, err := sdk.TeamsAdminAPI.CloudCallData.Get(ctx)
			if err != nil || value.Token != nil {
				t.Fatal("null token was not preserved as nil")
			}
			return contracttest.Observation{Response: response, Err: err}
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
		"get-array":     {"GET", getCommand, 200, `[]`, client.ErrDecode, get},
		"get-null":      {"GET", getCommand, 200, `null`, client.ErrDecode, get},
		"get-bad-token": {"GET", getCommand, 200, `{"Token":7}`, client.ErrDecode, get},
		"create-empty":  {"POST", newCommand, 200, ``, client.ErrDecode, create},
		"create-202":    {"POST", newCommand, 202, `{}`, ccd.ErrResponse, create},
	} {
		run(t, contracttest.Case{Name: name, Mappings: mapping(c.command),
			Steps: []contracttest.Exchange{discovery(t), wire(c.method, c.command, c.status, []byte(c.body))},
			Want:  contracttest.Expected{Status: c.status, Error: c.want},
			Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.CloudCallData)
				if value != nil {
					t.Fatal("failed response returned a value")
				}
				return contracttest.Observation{Response: response, Err: err}
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
