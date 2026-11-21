package team_test

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"net/url"
	"strings"
	"testing"
	"testing/synctest"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
)

func TestCreateBindsUsersAndTemplatesToSelectedCloud(t *testing.T) {
	t.Parallel()
	for _, cloud := range []string{"commercial", "gcc", "gcch", "dod", "china", "france", "germany"} {
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			t.Run(cloud+"/"+string(mode), func(t *testing.T) {
				env, err := auth.ResolveEnvironment(cloud, nil)
				if err != nil {
					t.Fatal(err)
				}
				falseValue, empty := false, ""
				want := `{"template@odata.bind":"` + env.Endpoints.Graph + `/v1.0/teamsTemplates('standard')","displayName":"fixture","description":"","memberSettings":{"allowDeleteChannels":false},"members":[{"@odata.type":"#microsoft.graph.aadUserConversationMember","roles":["owner"],"user@odata.bind":"` + env.Endpoints.Graph + `/v1.0/users('` + userID + `')"}]}`
				sdk, _, calls := fixture(t, cloud, mode, exchange{method: "POST", path: "/v1.0/teams", json: want, contentType: "application/json", status: 202,
					header: http.Header{"Location": {"/teams('" + teamID + "')/operations('" + opID + "')"}}})
				members := []team.InitialMember{{UserID: userID, Owner: true}}
				accepted, err := sdk.GraphAPI.Teams.Create(t.Context(), team.CreateRequest{DisplayName: "fixture", Description: &empty,
					Settings: team.Settings{MemberSettings: &team.MemberSettings{AllowDeleteChannels: &falseValue}}, Members: members})
				if err != nil || accepted.Operation == nil || accepted.TeamID != teamID || accepted.Response.StatusCode() != 202 || *calls != 1 {
					t.Fatal("create was not accepted once")
				}
				snapshot := accepted.Operation.Snapshot()
				location, parseErr := url.Parse(snapshot.Handle.Location)
				if parseErr != nil || snapshot.Status.State != client.OperationAccepted || snapshot.Handle.ID != opID || location.Scheme+"://"+location.Host != env.Endpoints.Graph || location.Path != "/v1.0/teams('"+teamID+"')/operations('"+opID+"')" || members[0].UserID != userID {
					t.Fatal("accepted handle, cloud or caller members changed")
				}
				for _, value := range []any{accepted, accepted.Operation, snapshot} {
					if rendered := fmt.Sprintf("%+v %#v", value, value); strings.Contains(rendered, teamID) || strings.Contains(rendered, "fixture-response") {
						t.Fatal("formatting exposed operation metadata")
					}
				}
			})
		}
	}
}

func TestDelegatedCreateCanOmitMembers(t *testing.T) {
	t.Parallel()
	sdk, _, _ := fixture(t, "commercial", auth.Delegated, exchange{method: "POST", path: "/v1.0/teams", json: `{"template@odata.bind":"https://graph.microsoft.com/v1.0/teamsTemplates('educationClass')","displayName":"fixture"}`, status: 202,
		header: http.Header{"Location": {"https://graph.microsoft.com/v1.0/teams/" + teamID + "/operations/" + opID}}})
	accepted, err := sdk.GraphAPI.Teams.Create(t.Context(), team.CreateRequest{DisplayName: "fixture", TemplateID: "educationClass"})
	if err != nil || accepted.Operation == nil {
		t.Fatal("delegated create unexpectedly required initial members")
	}
}

func TestCreateValidationDoesNotAcquireTokens(t *testing.T) {
	t.Parallel()
	sdk, p, _ := fixture(t, "commercial", auth.Application)
	for _, value := range []team.CreateRequest{
		{}, {DisplayName: " "}, {DisplayName: "fixture"},
		{DisplayName: "fixture", TemplateID: "standard')/users('private", Members: []team.InitialMember{{UserID: userID}}},
		{DisplayName: "fixture", Members: []team.InitialMember{{UserID: "../private"}}},
	} {
		if accepted, err := sdk.GraphAPI.Teams.Create(t.Context(), value); !errors.Is(err, client.ErrRequest) || accepted.Response != nil {
			t.Fatal("invalid create sent")
		}
	}
	if p.calls != 0 {
		t.Fatal("local create validation acquired credentials")
	}
}

func TestAcceptedOperationRejectsUnsafeLocationsWithoutLosingResponse(t *testing.T) {
	t.Parallel()
	for i, location := range []string{
		"", "https://foreign.invalid/v1.0/teams/" + teamID + "/operations/" + opID,
		"//foreign.invalid/v1.0/teams/" + teamID + "/operations/" + opID,
		"http://graph.microsoft.com/v1.0/teams/" + teamID + "/operations/" + opID,
		"/v1.0/users/" + userID,
		"/beta/teams/" + teamID + "/operations/" + opID,
		"/v1.0/teams/" + teamID + "/operations/" + opID + "?private=1",
		"/v1.0/teams/" + teamID + "/operations/" + opID + "#private",
		"/v1.0/teams/" + teamID + "/operations/../users",
		"/v1.0/teams/" + teamID + "/operations/private%2Fsegment",
		"/v1.0/teams/" + userID + "/operations/" + opID,
	} {
		t.Run(fmt.Sprintf("case-%d", i), func(t *testing.T) {
			sdk, _, calls := fixture(t, "commercial", auth.Application, exchange{method: "POST", path: "/v1.0/teams/" + teamID + "/archive", status: 202,
				header: http.Header{"Location": {location}}})
			accepted, err := sdk.GraphAPI.Teams.Archive(t.Context(), teamID, team.ArchiveOptions{})
			if err == nil || accepted.Operation != nil || accepted.TeamID != "" || accepted.Response == nil || accepted.Response.StatusCode() != 202 || *calls != 1 {
				t.Fatal("unsafe handle followed or acceptance lost")
			}
		})
	}
}

func TestArchiveAndUnarchiveWireBodies(t *testing.T) {
	t.Parallel()
	falseValue := false
	header := http.Header{"Location": {"/v1.0/teams/" + teamID + "/operations/" + opID}}
	sdk, _, calls := fixture(t, "commercial", auth.Delegated,
		exchange{method: "POST", path: "/v1.0/teams/" + teamID + "/archive", status: 202, header: header},
		exchange{method: "POST", path: "/v1.0/teams/" + teamID + "/archive", json: `{"shouldSetSpoSiteReadOnlyForMembers":false}`, status: 202, header: header},
		exchange{method: "POST", path: "/v1.0/teams/" + teamID + "/unarchive", status: 202, header: header})
	for _, options := range []team.ArchiveOptions{{}, {ShouldSetSpoSiteReadOnlyForMembers: &falseValue}} {
		accepted, err := sdk.GraphAPI.Teams.Archive(t.Context(), teamID, options)
		if err != nil || accepted.Operation == nil {
			t.Fatal("archive failed")
		}
	}
	accepted, err := sdk.GraphAPI.Teams.Unarchive(t.Context(), teamID)
	if err != nil || accepted.Operation == nil || *calls != 3 {
		t.Fatal("unarchive failed or unexpected implicit read/wait")
	}
}

func TestArchiveSharePointOptionIsRejectedInApplicationMode(t *testing.T) {
	t.Parallel()
	sdk, p, _ := fixture(t, "commercial", auth.Application)
	for _, value := range []bool{true, false} {
		if accepted, err := sdk.GraphAPI.Teams.Archive(t.Context(), teamID, team.ArchiveOptions{ShouldSetSpoSiteReadOnlyForMembers: &value}); !errors.Is(err, team.ErrArchiveOption) || accepted.Response != nil {
			t.Fatal("application SharePoint option accepted")
		}
	}
	if p.calls != 0 {
		t.Fatal("unsupported option acquired tokens")
	}
}

func TestAcceptedOperationHonorsRetryAfterAndDoesNotPollImplicitly(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		statusPath := "/v1.0/teams/" + teamID + "/operations/" + opID
		sdk, _, calls := fixture(t, "commercial", auth.Application,
			exchange{method: "POST", path: "/v1.0/teams/" + teamID + "/archive", status: 202, header: http.Header{"Location": {statusPath}, "Retry-After": {"60"}}},
			exchange{method: "GET", path: statusPath, status: 200, body: `{"id":"` + opID + `","status":"succeeded","targetResourceLocation":"https://foreign.invalid/private"}`})
		accepted, err := sdk.GraphAPI.Teams.Archive(t.Context(), teamID, team.ArchiveOptions{})
		if err != nil || *calls != 1 || accepted.Operation.Snapshot().Status.State != client.OperationAccepted {
			t.Fatal("acceptance caused implicit polling/completion")
		}
		start := time.Now()
		result, err := accepted.Operation.Poll(t.Context())
		if err != nil || time.Since(start) != time.Minute || result.Status.State != client.OperationSucceeded || result.Status.TargetLocation != "https://foreign.invalid/private" || *calls != 2 {
			t.Fatal("cadence, terminal state or opaque target changed")
		}
		if _, err := accepted.Operation.Poll(t.Context()); err != nil || *calls != 2 {
			t.Fatal("terminal operation polled again")
		}
	})
}

func TestResumedOperationWaitUsesGraphCadence(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		path := "/v1.0/teams/" + teamID + "/operations/" + opID
		sdk, _, _ := fixture(t, "commercial", auth.Application,
			exchange{method: "GET", path: path, status: 200, body: `{"id":"` + opID + `","status":"notStarted"}`},
			exchange{method: "GET", path: path, status: 200, body: `{"id":"` + opID + `","status":"inProgress"}`},
			exchange{method: "GET", path: path, status: 200, body: `{"id":"` + opID + `","status":"succeeded"}`})
		op, err := sdk.GraphAPI.Teams.ResumeOperation(teamID, opID)
		if err != nil {
			t.Fatal(err)
		}
		start := time.Now()
		result, err := op.Wait(t.Context(), client.WaitOptions{Timeout: 2 * time.Minute, Interval: time.Millisecond})
		if err != nil || result.Status.State != client.OperationSucceeded || time.Since(start) != 62*time.Second {
			t.Fatal("wait did not use Graph minimum cadence")
		}
	})
}

func TestOperationFailureUnknownStateAndIdentityMismatch(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		want       error
	}{
		{"remote-failed", `{"id":"` + opID + `","status":"failed","error":{"code":"FixtureFailure","message":"private"}}`, client.ErrOperationFailed},
		{"future-state", `{"id":"` + opID + `","status":"futureStatus"}`, client.ErrOperationState},
		{"mismatched-id", `{"id":"different","status":"succeeded"}`, client.ErrOperationState},
		{"malformed", `null`, client.ErrDecode},
	} {
		t.Run(tc.name, func(t *testing.T) {
			sdk, _, _ := fixture(t, "commercial", auth.Application, exchange{method: "GET", path: "/v1.0/teams/" + teamID + "/operations/" + opID, status: 200, body: tc.body})
			op, err := sdk.GraphAPI.Teams.ResumeOperation(teamID, opID)
			if err != nil {
				t.Fatal(err)
			}
			result, err := op.Poll(t.Context())
			if !errors.Is(err, tc.want) || result.Handle.ID != opID || result.Response == nil {
				t.Fatal("operation failed without recoverable response/identity")
			}
			if tc.name == "remote-failed" && (result.Diagnostics == nil || len(result.Diagnostics.Codes) != 1 || result.Diagnostics.Codes[0] != "FixtureFailure") {
				t.Fatal("operation diagnostics lost")
			}
		})
	}
}

func TestCancelledWaitPreservesAcceptedOperation(t *testing.T) {
	synctest.Test(t, func(t *testing.T) {
		sdk, _, calls := fixture(t, "commercial", auth.Application, exchange{method: "POST", path: "/v1.0/teams/" + teamID + "/unarchive", status: 202,
			header: http.Header{"Location": {"/v1.0/teams/" + teamID + "/operations/" + opID}}})
		accepted, err := sdk.GraphAPI.Teams.Unarchive(t.Context(), teamID)
		if err != nil {
			t.Fatal(err)
		}
		result, err := accepted.Operation.Wait(t.Context(), client.WaitOptions{Timeout: time.Second})
		if !errors.Is(err, context.DeadlineExceeded) || result.Handle.ID != opID || result.Response != accepted.Response || *calls != 1 {
			t.Fatal("local timeout lost remote acceptance or polled too early")
		}
	})
}

func TestOperationSubmissionIsNeverReplayed(t *testing.T) {
	t.Parallel()
	for _, action := range []string{"create", "archive", "unarchive"} {
		t.Run(action, func(t *testing.T) {
			path, body := "/v1.0/teams/"+teamID+"/"+action, ""
			if action == "create" {
				path = "/v1.0/teams"
				body = `{"displayName":"fixture","template@odata.bind":"https://graph.microsoft.com/v1.0/teamsTemplates('standard')"}`
			}
			sdk, _, calls := fixture(t, "commercial", auth.Delegated, exchange{method: "POST", path: path, json: body, status: 503, body: `{"error":{"code":"ServiceUnavailable"}}`})
			var accepted team.AcceptedOperation
			var err error
			switch action {
			case "create":
				accepted, err = sdk.GraphAPI.Teams.Create(t.Context(), team.CreateRequest{DisplayName: "fixture"})
			case "archive":
				accepted, err = sdk.GraphAPI.Teams.Archive(t.Context(), teamID, team.ArchiveOptions{})
			case "unarchive":
				accepted, err = sdk.GraphAPI.Teams.Unarchive(t.Context(), teamID)
			}
			var apiErr *client.APIError
			if !errors.As(err, &apiErr) || apiErr.StatusCode != 503 || accepted.Response == nil || accepted.Operation != nil || *calls != 1 {
				t.Fatal("submission replayed or failure metadata lost")
			}
		})
	}
}
