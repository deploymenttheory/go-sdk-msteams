package onlinevoicemailpolicy_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const scoped = "scope /Fixture / Child"
const collection = "/Skype.Policy/configurations/OnlineVoicemailPolicy"
const item = collection + "/configuration/Global"

type inputVectors struct {
	Name, Operation, Response, Duration string
	Status                              int
	Nested, Force                       bool
	Request                             struct {
		Method, Url, Body, CmdletName string
		FailureType                   *string
	}
}

func requestInputs(t *testing.T) []inputVectors {
	t.Helper()

	value := struct {
		Cases []inputVectors `json:"cases"`
	}{
		Cases: []inputVectors{
			{
				Name:      "read",
				Operation: "read",
				Duration:  "normal",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsOnlineVoicemailPolicy",
				},
			},
			{
				Name:      "list",
				Operation: "list",
				Duration:  "normal",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailPolicy",
					CmdletName: "Get-CsOnlineVoicemailPolicy",
				},
			},
			{
				Name:      "create",
				Operation: "create",
				Duration:  "normal",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailPolicy",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"Description\": \"\",\n  \"EnableTranscription\": false,\n  \"ShareData\": \"FutureValue\",\n  \"EnableTranscriptionProfanityMasking\": false,\n  \"EnableEditingCallAnswerRulesSetting\": false,\n  \"MaximumRecordingLength\": \"00:01:30\",\n  \"EnableTranscriptionTranslation\": false,\n  \"PrimarySystemPromptLanguage\": \"FuturePrimary\",\n  \"SecondarySystemPromptLanguage\": \"FutureSecondary\",\n  \"PreambleAudioFile\": \"fixture-audio-reference\",\n  \"PostambleAudioFile\": \"\",\n  \"PreamblePostambleMandatory\": false,\n  \"EnableVoicemailTriage\": false\n}",

					CmdletName: "New-CsOnlineVoicemailPolicy",
				},
			},
			{
				Name:      "patch",
				Operation: "patch",
				Duration:  "normal",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"\",\n  \"EnableTranscription\": false,\n  \"ShareData\": \"FutureValue\",\n  \"EnableTranscriptionProfanityMasking\": false,\n  \"EnableEditingCallAnswerRulesSetting\": false,\n  \"MaximumRecordingLength\": \"00:01:30\",\n  \"EnableTranscriptionTranslation\": false,\n  \"PrimarySystemPromptLanguage\": \"FuturePrimary\",\n  \"SecondarySystemPromptLanguage\": \"FutureSecondary\",\n  \"PreambleAudioFile\": \"fixture-audio-reference\",\n  \"PostambleAudioFile\": \"\",\n  \"PreamblePostambleMandatory\": false,\n  \"EnableVoicemailTriage\": false\n}",

					CmdletName: "Set-CsOnlineVoicemailPolicy",
				},
			},
			{
				Name:      "delete",
				Operation: "delete",
				Duration:  "normal",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "DELETE",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\"\n}",

					CmdletName: "Remove-CsOnlineVoicemailPolicy",
				},
			},
			{
				Name:      "patch-null-duration",
				Operation: "patch",
				Duration:  "null",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "PATCH",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"Description\": \"\",\n  \"EnableTranscription\": false,\n  \"ShareData\": \"FutureValue\",\n  \"EnableTranscriptionProfanityMasking\": false,\n  \"EnableEditingCallAnswerRulesSetting\": false,\n  \"MaximumRecordingLength\": null,\n  \"EnableTranscriptionTranslation\": false,\n  \"PrimarySystemPromptLanguage\": \"FuturePrimary\",\n  \"SecondarySystemPromptLanguage\": \"FutureSecondary\",\n  \"PreambleAudioFile\": \"fixture-audio-reference\",\n  \"PostambleAudioFile\": \"\",\n  \"PreamblePostambleMandatory\": false,\n  \"EnableVoicemailTriage\": false\n}",

					CmdletName: "Set-CsOnlineVoicemailPolicy",
				},
			},
			{
				Name:      "create-one-tick",
				Operation: "create",
				Duration:  "tick",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "POST",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailPolicy",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"Description\": \"\",\n  \"EnableTranscription\": false,\n  \"ShareData\": \"FutureValue\",\n  \"EnableTranscriptionProfanityMasking\": false,\n  \"EnableEditingCallAnswerRulesSetting\": false,\n  \"MaximumRecordingLength\": \"00:00:00.0000001\",\n  \"EnableTranscriptionTranslation\": false,\n  \"PrimarySystemPromptLanguage\": \"FuturePrimary\",\n  \"SecondarySystemPromptLanguage\": \"FutureSecondary\",\n  \"PreambleAudioFile\": \"fixture-audio-reference\",\n  \"PostambleAudioFile\": \"\",\n  \"PreamblePostambleMandatory\": false,\n  \"EnableVoicemailTriage\": false\n}",

					CmdletName: "New-CsOnlineVoicemailPolicy",
				},
			},
			{
				Name:      "delete-force-false",
				Operation: "delete",
				Duration:  "normal",
				Status:    400,
				Force:     true,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "DELETE",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\",\n  \"Force\": \"False\"\n}",

					CmdletName: "Remove-CsOnlineVoicemailPolicy",
				},
			},
			{
				Name:      "read-empty",
				Operation: "read",
				Duration:  "normal",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsOnlineVoicemailPolicy",
				},
			},
			{
				Name:      "read-null-duration",
				Operation: "read",
				Duration:  "normal",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsOnlineVoicemailPolicy",
				},
			},
			{
				Name:      "read-invalid-duration",
				Operation: "read",
				Duration:  "normal",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsOnlineVoicemailPolicy",
				},
			},
			{
				Name:      "list-envelope",
				Operation: "list",
				Duration:  "normal",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailPolicy",
					CmdletName: "Get-CsOnlineVoicemailPolicy",
				},
			},
			{
				Name:      "delete-denied",
				Operation: "delete",
				Duration:  "normal",
				Status:    400,
				Force:     false,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method: "DELETE",
					Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/OnlineVoicemailPolicy/configuration/scope%20/Fixture%20%2F%20Child",
					Body:   "{\n  \"Identity\": \"scope /Fixture / Child\"\n}",

					CmdletName: "Remove-CsOnlineVoicemailPolicy",
				},
			},
		},
	}

	return value.Cases
}
func wire(method, path, command string, body json.RawMessage, status int, reply string) testutil.Exchange {
	h := http.Header{}
	h.Set("MPACmdlet", "true")
	h.Set("X-MS-CmdletName", command)
	h[http.CanonicalHeaderKey("X-MS-Target-Uri")] = nil
	if body != nil {
		h.Set("Content-Type", "application/json; charset=utf-8")
	}
	return testutil.Exchange{Request: testutil.WireRequest{Method: method, URL: origin + path, Audience: auth.TeamsAdminAudience, Headers: h, JSON: body}, Reply: testutil.Reply{Status: status, Headers: http.Header{"Content-Type": {"text/plain; charset=utf-8"}}, Body: []byte(reply)}}
}
func run(t *testing.T, c testutil.Case) testutil.Result {
	t.Helper()

	if len(c.Steps) == 0 {

	} else {

		c.Want.Status = c.Steps[len(c.Steps)-1].Reply.Status
	}
	r, e := testutil.Run(t.Context(), c)
	if e != nil || !r.Passed() {
		t.Fatalf("voicemail contract differs: %v %+v", e, r.Differences())
	}
	return r
}
