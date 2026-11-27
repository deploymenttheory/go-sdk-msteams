package licensereport_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strconv"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	report "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/change_notification_license_report"
)

const graph = "https://graph.microsoft.com"
const userID = "00000000-0000-4000-8000-000000000001"
const secondID = "00000000-0000-4000-8000-000000000002"
const activity = `{"userId":"` + userID + `","userPrincipalName":"$guest'#EXT#@example.invalid","privateChatMessageCount":2147483647,"teamChatMessageCount":1,"future":[]}`
const users = `{"value":[{"id":"` + userID + `","userPrincipalName":"new-name@example.invalid","displayName":"Fixture user"}]}`
const licenses = `{"value":[{"id":"license-a","servicePlans":[{"servicePlanId":"` + report.SourcePlanID + `","provisioningStatus":"Disabled"}],"future":null}]}`

func activityURL(period string) string {
	return graph + "/beta/reports/" + url.PathEscape("getTeamsUserActivityUserDetail(period='"+period+"')") + "?$format=application/json&$top=2000"
}
func wire(method, url, body string, status int, response string) testutil.Exchange {
	headers := http.Header{}
	var raw json.RawMessage
	if method == http.MethodGet {
		headers.Set("Accept", "application/json")
	}
	if body != "" {
		raw = json.RawMessage(body)
		headers.Set("Content-Type", "application/json")
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: url, Audience: auth.GraphAudience, Headers: headers, JSON: raw}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(response)}}
}
func run(t *testing.T, c testutil.Case) {
	t.Helper()

	if c.Mode == "" {
		c.Mode = auth.Delegated
	}
	r, e := testutil.Run(t.Context(), c)
	if e != nil || !r.Passed() {
		t.Fatalf("license report vector differed: %v %+v", e, r.Differences())
	}
}
func batchRequest(id string) string {
	return `{"requests":[{"id":"0","method":"GET","url":"/users/` + id + `/licenseDetails"}]}`
}
func TestUnitLicenseReportNativeIDCompositionAndUnboundedWireCount(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "native-id-composition", Steps: []testutil.Exchange{
		wire("GET", activityURL("D7"), "", 200, `{"value":[`+activity+`]}`),
		wire("GET", graph+"/v1.0/users?$select=id,displayName,userPrincipalName&$top=999", "", 200, users),
		wire("POST", graph+"/v1.0/$batch", batchRequest(userID), 200, `{"responses":[{"id":"0","status":200,"body":`+licenses+`}]}`),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		result, e := sdk.GraphAPI.ChangeNotificationLicenseReports.Get(ctx, report.Period7Days, report.GetOptions{})
		if e != nil {
			return testutil.Observation{Err: e}
		}
		if result == nil || !result.Complete || !result.ActivityComplete || !result.UsersComplete || !result.LicensesComplete || len(result.Rows) != 1 {
			t.Fatal("composed stages missing")
		}
		row := result.Rows[0]
		if row.DirectoryUserID != userID || row.Activity.UserPrincipalName == nil || *row.Activity.UserPrincipalName != "$guest'#EXT#@example.invalid" || row.DirectoryUser == nil || row.DirectoryUser.UserPrincipalName == nil || *row.DirectoryUser.UserPrincipalName != "new-name@example.invalid" {
			t.Fatal("reported and directory identities were conflated")
		}
		if row.MessageCount == nil || row.MessageCount.String() != "2147483648" || row.HasMatchingSourcePlan == nil || !*row.HasMatchingSourcePlan {
			t.Fatal("native wide count or source-plan predicate changed")
		}
		if !result.Licenses[0].PlanDataComplete {
			t.Fatal("complete plan data not retained")
		}
		return testutil.Observation{}
	}})
}
func TestUnitLicenseReportAllPeriodValuesUseDocumentedJSONRoute(t *testing.T) {
	t.Parallel()
	for _, p := range []struct {
		value report.Period
		text  string
	}{{report.Period7Days, "D7"}, {report.Period30Days, "D30"}, {report.Period90Days, "D90"}, {report.Period180Days, "D180"}} {
		t.Run(p.text, func(t *testing.T) {
			run(t, testutil.Case{Name: strings.ToLower(p.text), Steps: []testutil.Exchange{wire("GET", activityURL(p.text), "", 200, `{"value":[]}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				result, e := sdk.GraphAPI.ChangeNotificationLicenseReports.ListActivity(ctx, p.value, report.ListOptions{})
				if e == nil && (result == nil || result.Pages != 1) {
					t.Fatal("period read completion missing")
				}
				return testutil.Observation{Err: e}
			}})
		})
	}
}
func TestUnitLicenseReportUnsafeContinuationAndPartialReadNeverStartsJoin(t *testing.T) {
	t.Parallel()
	for _, next := range []string{"https://evil.example.invalid/beta/reports/next", graph + "/v1.0/users"} {
		t.Run(next, func(t *testing.T) {
			run(t, testutil.Case{Name: "unsafe-continuation", Steps: []testutil.Exchange{wire("GET", activityURL("D7"), "", 200, `{"value":[`+activity+`],"@odata.nextLink":"`+next+`"}`)}, Want: testutil.Expected{Error: client.ErrDestination}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				result, e := sdk.GraphAPI.ChangeNotificationLicenseReports.Get(ctx, report.Period7Days, report.GetOptions{})
				if !errors.Is(e, client.ErrDestination) || result == nil || result.ActivityComplete || len(result.Activity.Items) != 1 || result.UsersComplete {
					t.Fatal("unsafe continuation lost partial data or authorized a join")
				}
				return testutil.Observation{Err: e}
			}})
		})
	}
}
func TestUnitLicenseReportDeniedLicenseIsUnknownAndRetainsReceipt(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "license-denied", Steps: []testutil.Exchange{
		wire("GET", activityURL("D7"), "", 200, `{"value":[`+activity+`]}`), wire("GET", graph+"/v1.0/users?$select=id,displayName,userPrincipalName&$top=999", "", 200, users), wire("POST", graph+"/v1.0/$batch", batchRequest(userID), 200, `{"responses":[{"id":"0","status":403,"body":{"error":{"code":"Authorization_RequestDenied","message":"private fixture"}}}]}`),
	}, Want: testutil.Expected{Error: auth.ErrForbidden}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		result, e := sdk.GraphAPI.ChangeNotificationLicenseReports.Get(ctx, report.Period7Days, report.GetOptions{})
		if !errors.Is(e, report.ErrPartial) || !errors.Is(e, auth.ErrForbidden) || result == nil || result.Complete || len(result.Rows) != 1 || result.Rows[0].HasMatchingSourcePlan != nil || len(result.Licenses) != 1 || result.Licenses[0].Response == nil {
			t.Fatal("denial became a negative verdict or lost stages")
		}
		return testutil.Observation{Err: e}
	}})
}
func TestUnitLicenseReportDeclaredIDCannotFallBackToDifferentUPNOwner(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "identity-mismatch", Steps: []testutil.Exchange{
		wire("GET", activityURL("D7"), "", 200, `{"value":[`+activity+`]}`), wire("GET", graph+"/v1.0/users?$select=id,displayName,userPrincipalName&$top=999", "", 200, `{"value":[{"id":"`+secondID+`","userPrincipalName":"$guest'#EXT#@example.invalid"}]}`),
	}, Want: testutil.Expected{Error: report.ErrIdentity}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		result, e := sdk.GraphAPI.ChangeNotificationLicenseReports.Get(ctx, report.Period7Days, report.GetOptions{})
		if !errors.Is(e, report.ErrIdentity) || result == nil || result.Complete || len(result.Licenses) != 0 || result.Rows[0].DirectoryUserID != "" || result.Rows[0].HasMatchingSourcePlan != nil {
			t.Fatal("report identity retargeted through reused UPN")
		}
		return testutil.Observation{Err: e}
	}})
}
func TestUnitLicenseReportModeAndInputBoundariesDoNotAcquireCredentials(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "app-mode", Mode: auth.Application, Want: testutil.Expected{Error: report.ErrLicenseMode}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		result, e := sdk.GraphAPI.ChangeNotificationLicenseReports.Get(ctx, report.Period7Days, report.GetOptions{})
		if result == nil || result.DocumentedLicenseMode || result.Complete {
			t.Fatal("unsupported documented mode labeled complete")
		}
		return testutil.Observation{Err: e}
	}})
	run(t, testutil.Case{Name: "bad-period", Want: testutil.Expected{Error: report.ErrInput}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		_, e := sdk.GraphAPI.ChangeNotificationLicenseReports.ListActivity(ctx, 8, report.ListOptions{})
		return testutil.Observation{Err: e}
	}})
}

func TestUnitLicenseReportTwentyItemBatchesCorrelateReorderedReplies(t *testing.T) {
	t.Parallel()
	activities := make([]map[string]any, 0, 21)
	directory := make([]map[string]any, 0, 21)
	ids := make([]string, 21)
	for i := range ids {
		ids[i] = fmt.Sprintf("00000000-0000-4000-8000-%012d", i+1)
		upn := fmt.Sprintf("u%d@example.invalid", i+1)
		activities = append(activities, map[string]any{"userId": ids[i], "userPrincipalName": upn, "privateChatMessageCount": 1, "teamChatMessageCount": 0})
		directory = append(directory, map[string]any{"id": ids[i], "userPrincipalName": upn})
	}
	encode := func(value any) string {
		raw, e := json.Marshal(value)
		if e != nil {
			t.Fatal(e)
		}
		return string(raw)
	}
	steps := []testutil.Exchange{wire("GET", activityURL("D7"), "", 200, encode(map[string]any{"value": activities})), wire("GET", graph+"/v1.0/users?$select=id,displayName,userPrincipalName&$top=999", "", 200, encode(map[string]any{"value": directory}))}
	for start := 0; start < len(ids); start += 20 {
		end := min(start+20, len(ids))
		requests := make([]map[string]any, 0, end-start)
		responses := make([]map[string]any, 0, end-start)
		for i := start; i < end; i++ {
			requests = append(requests, map[string]any{"id": strconv.Itoa(i - start), "method": "GET", "url": "/users/" + ids[i] + "/licenseDetails"})
		}
		for i := end - 1; i >= start; i-- {
			responses = append(responses, map[string]any{"id": strconv.Itoa(i - start), "status": 200, "body": map[string]any{"value": []any{}}})
		}
		steps = append(steps, wire("POST", graph+"/v1.0/$batch", encode(map[string]any{"requests": requests}), 200, encode(map[string]any{"responses": responses})))
	}
	run(t, testutil.Case{Name: "bounded-batches", Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		result, e := sdk.GraphAPI.ChangeNotificationLicenseReports.Get(ctx, report.Period7Days, report.GetOptions{})
		if e == nil {
			if result == nil || !result.Complete || len(result.Licenses) != 21 || len(result.Rows) != 21 {
				t.Fatal("bounded complete outcomes missing")
			}
			for i, row := range result.Rows {
				if row.DirectoryUserID != ids[i] || row.HasMatchingSourcePlan == nil || *row.HasMatchingSourcePlan {
					t.Fatal("batch correlation or complete empty verdict changed")
				}
			}
		}
		return testutil.Observation{Err: e}
	}})
}
func TestUnitLicenseReportLicenseContinuationIsSubjectBoundAndPageBounded(t *testing.T) {
	t.Parallel()
	for _, scenario := range []string{"complete", "page-limit", "other-user"} {
		t.Run(scenario, func(t *testing.T) {
			next := graph + "/v1.0/users/" + userID + "/licenseDetails?$skiptoken=opaque"
			var want error
			options := report.GetOptions{}
			if scenario == "page-limit" {
				options.Licenses.MaxPages = 1
				want = client.ErrPageLimit
			}
			if scenario == "other-user" {
				next = graph + "/v1.0/users/" + secondID + "/licenseDetails?$skiptoken=opaque"
				want = client.ErrDestination
			}
			steps := []testutil.Exchange{wire("GET", activityURL("D7"), "", 200, `{"value":[`+activity+`]}`), wire("GET", graph+"/v1.0/users?$select=id,displayName,userPrincipalName&$top=999", "", 200, users), wire("POST", graph+"/v1.0/$batch", batchRequest(userID), 200, `{"responses":[{"id":"0","status":200,"body":{"value":[],"@odata.nextLink":"`+next+`"}}]}`)}
			if scenario == "complete" {
				steps = append(steps, wire("GET", next, "", 200, licenses))
			}
			run(t, testutil.Case{Name: scenario, Steps: steps, Want: testutil.Expected{Error: want}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				result, e := sdk.GraphAPI.ChangeNotificationLicenseReports.Get(ctx, report.Period7Days, options)
				if result == nil || len(result.Rows) != 1 || len(result.Licenses) != 1 {
					t.Fatal("license stages missing")
				}
				if want != nil {
					if !errors.Is(e, want) || result.Complete || result.Rows[0].HasMatchingSourcePlan != nil {
						t.Fatal("partial/retargeted licenses became a verdict")
					}
				} else if e == nil && (!result.Complete || result.Licenses[0].Pages != 2 || result.Rows[0].HasMatchingSourcePlan == nil || !*result.Rows[0].HasMatchingSourcePlan) {
					t.Fatal("complete continuation evidence missing")
				}
				return testutil.Observation{Err: e}
			}})
		})
	}
}
func TestUnitLicenseReportPartialDirectoryCannotEstablishResolution(t *testing.T) {
	t.Parallel()
	next := graph + "/v1.0/users?$skiptoken=opaque"
	run(t, testutil.Case{Name: "directory-partial", Steps: []testutil.Exchange{wire("GET", activityURL("D7"), "", 200, `{"value":[`+activity+`]}`), wire("GET", graph+"/v1.0/users?$select=id,displayName,userPrincipalName&$top=999", "", 200, `{"value":[{"id":"`+userID+`","userPrincipalName":"a@example.invalid"}],"@odata.nextLink":"`+next+`"}`), wire("GET", next, "", 403, `{"error":{"code":"Forbidden"}}`)}, Want: testutil.Expected{Error: auth.ErrForbidden}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		result, e := sdk.GraphAPI.ChangeNotificationLicenseReports.Get(ctx, report.Period7Days, report.GetOptions{})
		if !errors.Is(e, auth.ErrForbidden) || result == nil || result.UsersComplete || len(result.Users.Items) != 1 || len(result.Licenses) != 0 || result.Rows[0].HasMatchingSourcePlan != nil {
			t.Fatal("partial directory established an identity/license verdict")
		}
		return testutil.Observation{Err: e}
	}})
}

func TestUnitLicenseReportOpaqueDirectoryIDRemainsAuthoritativeAndSubjectBound(t *testing.T) {
	t.Parallel()
	const id = "opaque#User-1"
	const escaped = "opaque%23User-1"
	next := graph + "/v1.0/users/" + escaped + "/licenseDetails?$skiptoken=opaque"
	row := `{"userId":"` + id + `","userPrincipalName":"old@example.invalid","privateChatMessageCount":1,"teamChatMessageCount":0}`
	directory := `{"value":[{"id":"` + id + `","userPrincipalName":"renamed@example.invalid"}]}`
	run(t, testutil.Case{Name: "opaque-directory-id", Steps: []testutil.Exchange{
		wire("GET", activityURL("D7"), "", 200, `{"value":[`+row+`]}`),
		wire("GET", graph+"/v1.0/users?$select=id,displayName,userPrincipalName&$top=999", "", 200, directory),
		wire("POST", graph+"/v1.0/$batch", batchRequest(escaped), 200, `{"responses":[{"id":"0","status":200,"body":{"value":[],"@odata.nextLink":"`+next+`"}}]}`),
		wire("GET", next, "", 200, `{"value":[]}`),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		result, e := sdk.GraphAPI.ChangeNotificationLicenseReports.Get(ctx, report.Period7Days, report.GetOptions{})
		if e == nil && (result == nil || !result.Complete || result.Rows[0].DirectoryUserID != id || result.Licenses[0].Pages != 2 || result.Rows[0].HasMatchingSourcePlan == nil || *result.Rows[0].HasMatchingSourcePlan) {
			t.Fatal("opaque ID was rejected, normalized or retargeted")
		}
		return testutil.Observation{Err: e}
	}})
	run(t, testutil.Case{Name: "opaque-id-primitive", Steps: []testutil.Exchange{
		wire("GET", graph+"/v1.0/users/"+escaped+"/licenseDetails", "", 200, `{"value":[]}`),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		result, e := sdk.GraphAPI.ChangeNotificationLicenseReports.ListLicenseDetailsByUserID(ctx, id, report.ListOptions{})
		if e == nil && (result == nil || len(result.Items) != 0) {
			t.Fatal("opaque-ID primitive result changed")
		}
		return testutil.Observation{Err: e}
	}})
	run(t, testutil.Case{Name: "opaque-id-not-upn", Steps: []testutil.Exchange{
		wire("GET", activityURL("D7"), "", 200, `{"value":[`+row+`]}`),
		wire("GET", graph+"/v1.0/users?$select=id,displayName,userPrincipalName&$top=999", "", 200, `{"value":[{"id":"opaque#user-1","userPrincipalName":"old@example.invalid"}]}`),
	}, Want: testutil.Expected{Error: report.ErrIdentity}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		result, e := sdk.GraphAPI.ChangeNotificationLicenseReports.Get(ctx, report.Period7Days, report.GetOptions{})
		if !errors.Is(e, report.ErrIdentity) || result == nil || len(result.Licenses) != 0 || result.Rows[0].DirectoryUserID != "" {
			t.Fatal("opaque ID was case folded or fell back to UPN")
		}
		return testutil.Observation{Err: e}
	}})
}
