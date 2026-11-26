// Package callintegration verifies CALL service compositions against explicit
// offline contracts. Synthetic compositions do not claim tenant qualification.
package callintegration

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"os"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	audio "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/audio_prompts"
	aa "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/auto_attendants"
	cq "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/call_queues"
	accounts "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/resource_accounts"
	audioacc "github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/audio_prompts"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const source = "tests/integration/lab/call_integration/contracts_test.go"
const scheduleID = "00000000-0000-4000-8000-000000000001"
const flowID = "00000000-0000-4000-8000-000000000002"
const attendantID = "00000000-0000-4000-8000-000000000003"
const mediaID = "0123456789abcdef0123456789abcdef"
const fixture = "msteams-sdk-aaaaaaaaaaaaaaaaaaaaaaaaaa-audio"

func ptr[T any](value T) *T { return &value }

func discovery(t *testing.T) contracttest.Exchange {
	t.Helper()
	raw, err := os.ReadFile(labkit.Root + "/msteams/teams_admin_api/internal/configapi/testdata/service_discovery.json")
	if err != nil {
		t.Fatal(err)
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: "GET", URL: origin + "/Teams.Tenant/serviceDiscovery", Audience: auth.TeamsAdminAudience}, Reply: contracttest.Reply{Status: 200, Headers: http.Header{"Content-Type": {"application/json"}}, Body: raw}}
}

func exchange(method, path, command, body string, status int, reply string) contracttest.Exchange {
	var input json.RawMessage
	if body != "" {
		input = json.RawMessage(body)
	}
	headers := http.Header{"X-MS-CmdletName": {command}, "X-MS-Target-Uri": {"https://admin.example.invalid/"}, "Accept": {"application/json"}}
	if input != nil {
		headers.Set("Content-Type", "application/json")
	}
	return contracttest.Exchange{Request: contracttest.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: headers, JSON: input}, Reply: contracttest.Reply{Status: status, Headers: http.Header{"Content-Type": {"application/json"}}, Body: []byte(reply)}}
}

func run(t *testing.T, name string, steps []contracttest.Exchange, invoke func(context.Context, *msteams.Client) contracttest.Observation) {
	t.Helper()
	result, err := contracttest.Run(t.Context(), contracttest.Case{Name: name, Family: "call_integration", Scope: "service", Source: source, Steps: append([]contracttest.Exchange{discovery(t)}, steps...), Invoke: invoke})
	if err != nil || !result.Passed() {
		t.Fatalf("CALL composition failed: %v %+v", err, result.Differences())
	}
}

func TestUnitScheduleIdentitySurvivesAttendantComposition(t *testing.T) {
	t.Parallel()
	// These composition responses are synthetic, based on the documented service
	// DTOs. Unknown nested fields must survive the read/replace boundary.
	schedule := `{"Id":"` + scheduleID + `","Name":"owned-weekly","Type":"WeeklyRecurrence","WeeklyRecurrentSchedule":{"MondayHours":[],"IsComplemented":false,"RecurrenceRange":{"Type":"NoEnd"}},"FutureField":{"Enabled":false}}`
	association := `{"Type":"AfterHours","ScheduleId":"` + scheduleID + `","CallFlowId":"` + flowID + `","Enabled":true}`
	draft := `{"CallHandlingAssociation":` + association + `}`
	settings := `{"Name":"owned-attendant","LanguageId":"en-US","TimeZoneId":"UTC","DefaultCallFlow":{"Id":"` + flowID + `"},"Schedules":[` + schedule + `],"CallHandlingAssociations":[` + association + `]}`
	read := `{"AutoAttendant":{"Id":"` + attendantID + `","Name":"owned-attendant","LanguageId":"en-US","TimeZoneId":"UTC","DefaultCallFlow":{"Id":"` + flowID + `"},"Schedules":[` + schedule + `],"CallHandlingAssociations":[` + association + `]}}`
	run(t, "schedule-attendant-identities", []contracttest.Exchange{
		exchange("GET", "/Teams.VoiceApps/schedules/"+scheduleID, "Get-CsOnlineSchedule", "", 200, `{"Schedule":`+schedule+`}`),
		exchange("POST", "/Teams.VoiceApps/auto-attendants/call-handling-associations/draft", "New-CsAutoAttendantCallHandlingAssociation", association, 200, draft),
		exchange("POST", "/Teams.VoiceApps/auto-attendants", "New-CsAutoAttendant", settings, 201, read),
		exchange("PUT", "/Teams.VoiceApps/auto-attendants/"+attendantID, "Set-CsAutoAttendant", `{"Id":"`+attendantID+`",`+settings[1:], 200, read),
	}, func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		schedule, _, err := sdk.TeamsAdminAPI.Schedules.GetByID(ctx, scheduleID)
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		associated, _, err := sdk.TeamsAdminAPI.AutoAttendants.DraftCallHandlingAssociation(ctx, aa.CallHandlingAssociation{Type: ptr(aa.AssociationAfterHours), ScheduleID: schedule.ID, CallFlowID: ptr(flowID)})
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		raw, err := json.Marshal(schedule.RawFields)
		if err != nil {
			t.Fatal(err)
		}
		value, _, err := sdk.TeamsAdminAPI.AutoAttendants.Create(ctx, aa.Settings{Name: ptr("owned-attendant"), LanguageID: ptr("en-US"), TimeZoneID: ptr("UTC"), DefaultCallFlow: &aa.CallFlow{ID: ptr(flowID)}, Schedules: &[]json.RawMessage{raw}, CallHandlingAssociations: &[]aa.CallHandlingAssociation{*associated}})
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		if value.CallHandlingAssociations[0].ScheduleID == nil || *value.CallHandlingAssociations[0].ScheduleID != *schedule.ID {
			t.Fatal("schedule identity changed between services")
		}
		_, _, err = sdk.TeamsAdminAPI.AutoAttendants.UpdateByID(ctx, *value.ID, value.Settings())
		return contracttest.Observation{Err: err}
	})
}

func TestUnitAudioIdentityFeedsAttendantAndQueue(t *testing.T) {
	t.Parallel()
	audioReply := `{"Id":"` + mediaID + `","ApplicationId":"TenantGlobal","OriginalFilename":"owned.wav"}`
	prompt := `{"ActiveType":"AudioFile","AudioFilePrompt":{"Id":"` + mediaID + `","FileName":"owned.wav"}}`
	queue := `{"Name":"owned-queue","WelcomeMusicAudioFileId":"` + mediaID + `"}`
	run(t, "audio-attendant-queue-identities", []contracttest.Exchange{
		exchange("GET", "/Teams.MediaStorage/audiofile/TenantGlobal/"+mediaID, "Get-CsOnlineAudioFile", "", 200, audioReply),
		exchange("POST", "/Teams.VoiceApps/auto-attendants/prompts/draft", "New-CsAutoAttendantPrompt", prompt, 200, `{"Prompt":`+prompt+`}`),
		exchange("POST", "/Teams.VoiceApps/callqueues", "New-CsCallQueue", queue, 201, `{"CallQueue":{"Identity":"`+attendantID+`","Name":"owned-queue","WelcomeMusicResourceId":"`+mediaID+`"}}`),
		exchange("PUT", "/Teams.VoiceApps/callqueues/"+attendantID, "Set-CsCallQueue", queue, 200, `{"Diagnostics":[]}`),
	}, func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		file, _, err := sdk.TeamsAdminAPI.AudioFiles.GetByID(ctx, audio.TenantGlobal, mediaID, nil)
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		prompt, _, err := sdk.TeamsAdminAPI.AutoAttendants.DraftPrompt(ctx, aa.Prompt{ActiveType: ptr(aa.PromptAudioFile), AudioFilePrompt: &aa.AudioFile{ID: file.ID, FileName: file.OriginalFilename}})
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		if prompt.AudioFilePrompt == nil || *prompt.AudioFilePrompt.ID != *file.ID {
			t.Fatal("audio prompt lost uploaded file identity")
		}
		queue, _, err := sdk.TeamsAdminAPI.CallQueues.Create(ctx, cq.Settings{Name: ptr("owned-queue"), WelcomeMusicAudioFileID: file.ID}, nil)
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		_, _, err = sdk.TeamsAdminAPI.CallQueues.UpdateByID(ctx, *queue.Identity, queue.Settings(), nil)
		return contracttest.Observation{Err: err}
	})
}

func TestUnitQueueDenialDoesNotPreventIndependentAudioCleanup(t *testing.T) {
	t.Parallel()
	present := `{"Id":"` + mediaID + `","ApplicationId":"TenantGlobal","OriginalFilename":"` + fixture + `.wav"}`
	deleted := `{"Id":"` + mediaID + `","ApplicationId":"TenantGlobal","OriginalFilename":"` + fixture + `.wav","DeletionTimestampOffset":"2026-11-26T00:00:00Z"}`
	run(t, "authorization-and-independent-cleanup", []contracttest.Exchange{
		exchange("GET", "/Teams.VoiceApps/callqueues/"+attendantID+"?FilterInvalidObos=False", "Get-CsCallQueue", "", 403, `{"code":"Forbidden"}`),
		exchange("GET", "/Teams.MediaStorage/audiofile/TenantGlobal/"+mediaID, "Get-CsOnlineAudioFile", "", 200, present),
		exchange("DELETE", "/Teams.MediaStorage/audiofile/TenantGlobal/"+mediaID, "Remove-CsOnlineAudioFile", "", 204, ""),
		exchange("GET", "/Teams.MediaStorage/audiofile/TenantGlobal/"+mediaID, "Get-CsOnlineAudioFile", "", 200, deleted),
	}, func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		_, response, err := sdk.TeamsAdminAPI.CallQueues.GetByID(ctx, attendantID)
		if !errors.Is(err, auth.ErrForbidden) || response == nil || response.StatusCode() != 403 {
			t.Fatal("queue authorization denial was hidden")
		}
		handler := audioacc.RecoveryHandlers(sdk)["audio-file"]
		if handler == nil {
			t.Fatal("audio recovery handler missing")
		}
		return contracttest.Observation{Err: handler.Delete(context.WithoutCancel(ctx), lab.Entry{Fixture: fixture, ResourceID: mediaID})}
	})
}

func TestUnitQueueAccountAssociationRetainsPerEndpointFailure(t *testing.T) {
	t.Parallel()
	// The successful account read is serializer-derived, not a live account.
	// The failed association reply is a permanent sanitized tenant response.
	reply, err := os.ReadFile(labkit.Root + "/msteams/teams_admin_api/resource_accounts/testdata/associate_missing.json")
	if err != nil {
		t.Fatal(err)
	}
	run(t, "queue-account-association-outcome", []contracttest.Exchange{
		exchange("GET", "/Teams.VoiceApps/callqueues/"+attendantID+"?FilterInvalidObos=False", "Get-CsCallQueue", "", 200, `{"CallQueue":{"Identity":"`+attendantID+`","Name":"owned-queue"}}`),
		exchange("GET", "/Teams.PlatformService/v2/Legacy/ApplicationInstances?identity="+scheduleID, "Get-CsOnlineApplicationInstance", "", 200, `{"objectId":"`+scheduleID+`","displayName":"owned-account"}`),
		exchange("POST", "/Teams.VoiceApps/applicationinstanceassociations", "New-CsOnlineApplicationInstanceAssociation", `{"EndpointsIds":["`+scheduleID+`"],"ConfigurationId":"`+attendantID+`","ConfigurationType":"CallQueue"}`, 200, string(reply)),
	}, func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		queue, _, err := sdk.TeamsAdminAPI.CallQueues.GetByID(ctx, attendantID)
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		account, _, err := sdk.TeamsAdminAPI.ResourceAccounts.GetByID(ctx, scheduleID)
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		results, response, err := sdk.TeamsAdminAPI.ResourceAccounts.CreateAssociations(ctx, accounts.AssociationRequest{EndpointIDs: []string{*account.ObjectID}, ConfigurationID: *queue.Identity, ConfigurationType: "CallQueue"})
		if err != nil {
			return contracttest.Observation{Err: err}
		}
		if response == nil || response.StatusCode() != 200 || len(results) != 1 || results[0].Association != nil || results[0].Status == nil || !results[0].Status.Failed() || results[0].Status.ID == nil || *results[0].Status.ID != *account.ObjectID {
			t.Fatal("HTTP200 association lost the failed endpoint's identity or outcome")
		}
		return contracttest.Observation{}
	})
}
