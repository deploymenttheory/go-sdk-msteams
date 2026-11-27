package sipdevicesgateway_test

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
	sdg "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/sip_devices_gateway"
	"github.com/go-resty/resty/v2"
)

const (
	configOrigin = "https://api.interfaces.records.teams.microsoft.com"
	targetForest = "https://admin.example.invalid/"

	root        = "/Teams.SipDevicesGatewayManagement"
	probe       = "00000000-0000-4000-8000-000000000001"
	tenant      = "00000000-0000-4000-8000-0000000000aa"
	summaryCmd  = "Get-CsSdgBulkSignInRequestsSummary"
	statusCmd   = "Get-CsSdgBulkSignInRequestStatus"
	bulkCmd     = "New-CsSdgBulkSignInRequest"
	taggingCmd  = "New-CsSdgDeviceTaggingRequest"
	transferCmd = "New-CsSdgDeviceTransferRequest"
	region      = "msteams-sdk-invalid"
	hardware    = "00-00-5E-00-53-01"
	user        = "msteams-sdk-probe@example.invalid"
)

func readFile(t *testing.T, name string) []byte {
	t.Helper()
	directory, err := os.OpenRoot(".")
	if err != nil {
		t.Fatal(err)
	}
	defer func() { _ = directory.Close() }()
	body, err := directory.ReadFile(name)
	if err != nil {
		t.Fatal(err)
	}
	return body
}

func discovery(t *testing.T) testutil.Exchange {
	t.Helper()
	body, err := os.ReadFile("../internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: http.MethodGet, URL: configOrigin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience},
		Reply: testutil.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: body}}
}

func wire(method, path, command string, body json.RawMessage, status int, reply []byte) testutil.Exchange {
	h := http.Header{"Accept": {"application/json"}, "X-MS-CmdletName": {command}, "X-MS-Target-Uri": {targetForest}}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: configOrigin + root + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body},
		Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json; charset=utf-8"}}, Body: reply}}
}

func run(t *testing.T, c testutil.Case) {
	t.Helper()

	result, err := testutil.Run(t.Context(), c)
	if err != nil {
		t.Fatal(err)
	}
	if !result.Passed() {
		t.Fatalf("SIP devices gateway contract %s differed: %+v", c.Name, result.Differences())
	}
}

func ptr[T any](v T) *T { return &v }

type call struct {
	name, method, path, command, fixture string
	body                                 json.RawMessage
	invoke                               func(context.Context, *sdg.Service) (any, *resty.Response, error)
}

func calls() []call {
	return []call{
		{"summary", "GET", "/v1/batchsummaries", summaryCmd, "summary_403.json", nil,
			func(ctx context.Context, s *sdg.Service) (any, *resty.Response, error) {
				return s.ListBulkSignInRequests(ctx)
			}},
		{"status", "GET", "/v1/batchstatus?batchid=" + probe, statusCmd, "status_403.json", nil,
			func(ctx context.Context, s *sdg.Service) (any, *resty.Response, error) {
				return s.GetBulkSignInRequestStatus(ctx, probe)
			}},
		{"bulk-sign-in", "POST", "/v1/createbatch?TargetRegion=" + region, bulkCmd, "bulk_sign_in_403.json",
			json.RawMessage(`[{"userName":"` + user + `","hardwareId":"` + hardware + `"}]`),
			func(ctx context.Context, s *sdg.Service) (any, *resty.Response, error) {
				return s.CreateBulkSignInRequest(ctx, region, []sdg.BulkSignInItem{{UserName: ptr(user), HardwareID: ptr(hardware)}})
			}},
		{"tagging", "POST", "/devicetagging/tagnewdevice?TargetRegion=" + region, taggingCmd, "tagging_403.json",
			json.RawMessage(`{"HardwareId":"` + hardware + `","TenantId":"` + tenant + `"}`),
			func(ctx context.Context, s *sdg.Service) (any, *resty.Response, error) {
				return s.CreateDeviceTaggingRequest(ctx, region, sdg.DeviceTaggingRequest{HardwareID: ptr(hardware), TenantID: ptr(tenant)})
			}},
		{"transfer", "POST", "/devicetransfer/transfer?SdhDeviceId=" + probe + "&SourceRegion=" + region + "&TargetRegion=" + region, transferCmd, "transfer_403.json", nil,
			func(ctx context.Context, s *sdg.Service) (any, *resty.Response, error) {
				return s.CreateDeviceTransferRequest(ctx, sdg.DeviceTransferRequest{SdhDeviceID: probe, SourceRegion: region, TargetRegion: region})
			}},
	}
}

// Every operation was denied to the lab's app-only Teams Administrator.
func TestUnitSIPDevicesGatewayDenialsUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	for _, c := range calls() {
		denied := readFile(t, "testdata/"+c.fixture)
		run(t, testutil.Case{Name: "actual-" + c.name + "-forbidden",
			Steps: []testutil.Exchange{discovery(t), wire(c.method, c.path, c.command, c.body, 403, denied)},
			Want:  testutil.Expected{Status: 403, Body: denied, APIError: client.NewAPIError(client.TeamsAdmin, 403, nil, denied)},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				value, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.SIPDevicesGateway)
				if !errors.Is(err, auth.ErrForbidden) {
					t.Fatal("403 was not classified as forbidden")
				}
				return testutil.Observation{Value: value, Response: response, Err: err}
			}})
	}
}

func TestUnitSIPDevicesGatewaySuccessShapesPreserveFields(t *testing.T) {
	t.Parallel()
	summary := []byte(`{"requests":[{"batchId":"b1","batchStatus":"Completed","batchItemsCount":2,"userId":"` + probe + `","requestTime":"2026-11-23T00:00:00Z","region":"EMEA","future":1}],"next":null}`)
	status := []byte(`{"batchId":"b1","batchStatus":"Completed","batchItems":[{"userName":"` + user + `","hardwareId":"` + hardware + `","status":"Failed","error":"DeviceNotFound"},{"userName":"u2","hardwareId":"h2","status":"Succeeded","error":null}]}`)
	created := []byte(`{"batchId":"b1","batchItemsCount":1}`)
	tagged := []byte(`{"correlationId":"c1","status":"Accepted"}`)
	moved := []byte(`{"Message":"Transferred"}`)
	replies := map[string][]byte{"summary": summary, "status": status, "bulk-sign-in": created, "tagging": tagged, "transfer": moved}
	for _, c := range calls() {
		reply := replies[c.name]
		run(t, testutil.Case{Name: c.name + "-success",
			Steps: []testutil.Exchange{discovery(t), wire(c.method, c.path, c.command, c.body, 200, reply)},
			Want:  testutil.Expected{Status: 200, Body: reply},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				value, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.SIPDevicesGateway)
				if err != nil {
					t.Fatal(err)
				}
				switch v := value.(type) {
				case *sdg.BulkSignInRequestsSummary:
					if len(v.Requests) != 1 || *v.Requests[0].BatchItemsCount != 2 || v.Requests[0].RawFields["future"] == nil || v.RawFields["next"] == nil {
						t.Fatal("summary fields were lost")
					}
				case *sdg.BulkSignInRequestStatus:
					if len(v.BatchItems) != 2 || !v.BatchItems[0].Failed() || v.BatchItems[1].Failed() {
						t.Fatal("partial batch failure was lost")
					}
					if fmt.Sprintf("%v %+v %#v", *v, *v.BatchItems[0], *v.BatchItems[0]) != "bulk sign-in request status (redacted) bulk sign-in item status (redacted) bulk sign-in item status (redacted)" {
						t.Fatal("status formatting exposed account names")
					}
				case *sdg.BulkSignInResponse:
					if *v.BatchID != "b1" || *v.BatchItemsCount != 1 {
						t.Fatal("created batch was lost")
					}
				case *sdg.DeviceTaggingResponse:
					if *v.CorrelationID != "c1" || *v.Status != "Accepted" {
						t.Fatal("tagging result was lost")
					}
				case *sdg.DeviceTransferResponse:
					if *v.Message != "Transferred" {
						t.Fatal("transfer result was lost")
					}
				default:
					t.Fatalf("unexpected value %T", value)
				}
				return testutil.Observation{Response: response, Err: err}
			}})
	}
}

func TestUnitSIPDevicesGatewayMalformedAndUndeclaredResponsesFail(t *testing.T) {
	t.Parallel()
	for _, c := range calls() {
		for name, reply := range map[string]struct {
			status int
			body   string
			want   error
		}{"null": {200, `null`, client.ErrDecode}, "array": {200, `[]`, client.ErrDecode}, "accepted": {202, `{}`, sdg.ErrResponse}} {
			run(t, testutil.Case{Name: c.name + "-" + name,
				Steps: []testutil.Exchange{discovery(t), wire(c.method, c.path, c.command, c.body, reply.status, []byte(reply.body))},
				Want:  testutil.Expected{Status: reply.status, Error: reply.want},
				Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
					value, response, err := c.invoke(ctx, sdk.TeamsAdminAPI.SIPDevicesGateway)
					if !strings.HasPrefix(fmt.Sprintf("%v", value), "<nil>") {
						t.Fatal("failed response returned a value")
					}
					return testutil.Observation{Response: response, Err: err}
				}})
		}
	}
}

func TestUnitSIPDevicesGatewayInvalidInputFailsBeforeIO(t *testing.T) {
	t.Parallel()
	for name, invoke := range map[string]func(context.Context, *sdg.Service) error{
		"blank-batch": func(ctx context.Context, s *sdg.Service) error {
			_, _, err := s.GetBulkSignInRequestStatus(ctx, " ")
			return err
		},
		"blank-region": func(ctx context.Context, s *sdg.Service) error {
			_, _, err := s.CreateBulkSignInRequest(ctx, "", []sdg.BulkSignInItem{{}})
			return err
		},
		"no-items": func(ctx context.Context, s *sdg.Service) error {
			_, _, err := s.CreateBulkSignInRequest(ctx, region, nil)
			return err
		},
		"tagging-region": func(ctx context.Context, s *sdg.Service) error {
			_, _, err := s.CreateDeviceTaggingRequest(ctx, "", sdg.DeviceTaggingRequest{})
			return err
		},
		"transfer-device": func(ctx context.Context, s *sdg.Service) error {
			_, _, err := s.CreateDeviceTransferRequest(ctx, sdg.DeviceTransferRequest{SourceRegion: region, TargetRegion: region})
			return err
		},
		"transfer-source": func(ctx context.Context, s *sdg.Service) error {
			_, _, err := s.CreateDeviceTransferRequest(ctx, sdg.DeviceTransferRequest{SdhDeviceID: probe, TargetRegion: region})
			return err
		},
		"transfer-target": func(ctx context.Context, s *sdg.Service) error {
			_, _, err := s.CreateDeviceTransferRequest(ctx, sdg.DeviceTransferRequest{SdhDeviceID: probe, SourceRegion: region})
			return err
		},
	} {
		run(t, testutil.Case{Name: name, Want: testutil.Expected{Error: sdg.ErrInput},
			Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				return testutil.Observation{Err: invoke(ctx, sdk.TeamsAdminAPI.SIPDevicesGateway)}
			}})
	}
}

func TestUnitReadBulkSignInItemsMatchesImportCsv(t *testing.T) {
	t.Parallel()
	items, err := sdg.ReadBulkSignInItems(strings.NewReader(string(rune(0xFEFF)) + "Extra,HARDWAREID,username\nx,h1,u1\ny,,u2\nz\n"))
	if err != nil || len(items) != 3 {
		t.Fatalf("csv rows were not parsed: %v", err)
	}
	if *items[0].UserName != "u1" || *items[0].HardwareID != "h1" || *items[1].HardwareID != "" || items[2].UserName != nil || items[2].HardwareID != nil {
		t.Fatal("csv columns were not mapped by name")
	}
	body, err := json.Marshal(items)
	if err != nil || string(body) != `[{"userName":"u1","hardwareId":"h1"},{"userName":"u2","hardwareId":""},{}]` {
		t.Fatalf("items serialized differently: %s", body)
	}
	if fmt.Sprint(items[0]) != "bulk sign-in item (redacted)" {
		t.Fatal("item formatting exposed the account")
	}
	if _, err := sdg.ReadBulkSignInItems(strings.NewReader("")); !errors.Is(err, sdg.ErrInput) {
		t.Fatal("empty csv accepted")
	}
	if items, err := sdg.ReadBulkSignInItems(strings.NewReader("Username,HardwareId\n")); err != nil || len(items) != 0 {
		t.Fatal("header-only csv failed")
	}
}

func TestUnitSIPDevicesGatewayNilContextAndServiceAreRejected(t *testing.T) {
	t.Parallel()
	var s *sdg.Service
	//nolint:staticcheck // A nil context is the guarded input.
	if _, _, err := s.ListBulkSignInRequests(nil); !errors.Is(err, client.ErrNilContext) {
		t.Fatal("nil context accepted")
	}
	if _, _, err := s.GetBulkSignInRequestStatus(t.Context(), probe); !errors.Is(err, client.ErrClosed) {
		t.Fatal("nil service accepted")
	}
	ctx, cancel := context.WithCancel(t.Context())
	cancel()
	if _, _, err := sdg.NewService(nil).CreateDeviceTransferRequest(ctx, sdg.DeviceTransferRequest{}); !errors.Is(err, context.Canceled) {
		t.Fatal("canceled context accepted")
	}
}
