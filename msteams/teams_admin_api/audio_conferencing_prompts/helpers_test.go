package audioconferencingprompts_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/TeamsAudioConferencingCustomPromptsConfiguration"
const item = collection + "/configuration/Global"
const scoped = "scope /Fixture / Child"
const setCommand = "Set-CsTeamsAudioConferencingCustomPromptsConfiguration"
const firstID = "11111111-1111-1111-1111-111111111111"
const secondID = "22222222-2222-2222-2222-222222222222"

type inputVectors struct {
	Name, Operation, Fields, Response string
	Status                            int
	ForceMode                         int `json:"force_mode"`
	Request                           struct {
		Method, Url, Body, CmdletName string
		FailureType                   *string
	}
}

func requestInputs(t *testing.T) []inputVectors {
	t.Helper()

	v := struct{ Cases []inputVectors }{
		Cases: []inputVectors{
			{
				Name:      "read",
				Operation: "read",
				Fields:    "all",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAudioConferencingCustomPromptsConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsAudioConferencingCustomPromptsConfiguration",
				},
			},
			{
				Name:      "list-empty",
				Operation: "list",
				Fields:    "all",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAudioConferencingCustomPromptsConfiguration",
					CmdletName: "Get-CsTeamsAudioConferencingCustomPromptsConfiguration",
				},
			},
			{
				Name:      "list-null",
				Operation: "list",
				Fields:    "all",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAudioConferencingCustomPromptsConfiguration",
					CmdletName: "Get-CsTeamsAudioConferencingCustomPromptsConfiguration",
				},
			},
			{
				Name:      "patch-all",
				Operation: "patch",
				Fields:    "all",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "PATCH",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAudioConferencingCustomPromptsConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:       "{\n  \"Identity\": \"read-model-identity\",\n  \"Prompts\": [\n    {\n      \"Identity\": \"prompt-meta\",\n      \"Id\": \"11111111-1111-1111-1111-111111111111\",\n      \"Name\": \"\",\n      \"AudioPrompt\": null,\n      \"TextPrompt\": \"text\",\n      \"Type\": \"FutureType\",\n      \"Locale\": \"xx-future\"\n    }\n  ],\n  \"Packages\": [\n    {\n      \"Identity\": \"package-meta\",\n      \"InboundStartRecordingPrompt\": [\n        \"11111111-1111-1111-1111-111111111111\",\n        \"11111111-1111-1111-1111-111111111111\",\n        null\n      ],\n      \"InboundEndRecordingPrompt\": [\n        \"\"\n      ],\n      \"OutboundStartRecordingPrompt\": [\n        \"future-reference\"\n      ],\n      \"OutboundEndRecordingPrompt\": [\n        \"11111111-1111-1111-1111-111111111111\"\n      ],\n      \"MeetingStartRecordingPrompt\": [\n        \"22222222-2222-2222-2222-222222222222\"\n      ],\n      \"MeetingEndRecordingPrompt\": [],\n      \"Id\": \"22222222-2222-2222-2222-222222222222\",\n      \"Name\": null,\n      \"DefaultLocale\": \"\"\n    }\n  ]\n}",
					CmdletName: "Set-CsTeamsAudioConferencingCustomPromptsConfiguration",
				},
			},
			{
				Name:      "patch-prompts",
				Operation: "patch",
				Fields:    "prompts",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "PATCH",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAudioConferencingCustomPromptsConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:       "{\n  \"Identity\": \"read-model-identity\",\n  \"Prompts\": [\n    {\n      \"Identity\": \"prompt-meta\",\n      \"Id\": \"11111111-1111-1111-1111-111111111111\",\n      \"Name\": \"\",\n      \"AudioPrompt\": null,\n      \"TextPrompt\": \"text\",\n      \"Type\": \"FutureType\",\n      \"Locale\": \"xx-future\"\n    }\n  ]\n}",
					CmdletName: "Set-CsTeamsAudioConferencingCustomPromptsConfiguration",
				},
			},
			{
				Name:      "patch-packages",
				Operation: "patch",
				Fields:    "packages",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "PATCH",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAudioConferencingCustomPromptsConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:       "{\n  \"Identity\": \"read-model-identity\",\n  \"Packages\": [\n    {\n      \"Identity\": \"package-meta\",\n      \"InboundStartRecordingPrompt\": [\n        \"11111111-1111-1111-1111-111111111111\",\n        \"11111111-1111-1111-1111-111111111111\",\n        null\n      ],\n      \"InboundEndRecordingPrompt\": [\n        \"\"\n      ],\n      \"OutboundStartRecordingPrompt\": [\n        \"future-reference\"\n      ],\n      \"OutboundEndRecordingPrompt\": [\n        \"11111111-1111-1111-1111-111111111111\"\n      ],\n      \"MeetingStartRecordingPrompt\": [\n        \"22222222-2222-2222-2222-222222222222\"\n      ],\n      \"MeetingEndRecordingPrompt\": [],\n      \"Id\": \"22222222-2222-2222-2222-222222222222\",\n      \"Name\": null,\n      \"DefaultLocale\": \"\"\n    }\n  ]\n}",
					CmdletName: "Set-CsTeamsAudioConferencingCustomPromptsConfiguration",
				},
			},
			{
				Name:      "patch-empty-lists",
				Operation: "patch",
				Fields:    "empty-lists",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "PATCH",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAudioConferencingCustomPromptsConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:       "{\n  \"Identity\": \"read-model-identity\",\n  \"Prompts\": [],\n  \"Packages\": []\n}",
					CmdletName: "Set-CsTeamsAudioConferencingCustomPromptsConfiguration",
				},
			},
			{
				Name:      "patch-force-false",
				Operation: "patch",
				Fields:    "all",
				Status:    400,
				ForceMode: 0,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "PATCH",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAudioConferencingCustomPromptsConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:       "{\n  \"Identity\": \"read-model-identity\",\n  \"Prompts\": [\n    {\n      \"Identity\": \"prompt-meta\",\n      \"Id\": \"11111111-1111-1111-1111-111111111111\",\n      \"Name\": \"\",\n      \"AudioPrompt\": null,\n      \"TextPrompt\": \"text\",\n      \"Type\": \"FutureType\",\n      \"Locale\": \"xx-future\"\n    }\n  ],\n  \"Packages\": [\n    {\n      \"Identity\": \"package-meta\",\n      \"InboundStartRecordingPrompt\": [\n        \"11111111-1111-1111-1111-111111111111\",\n        \"11111111-1111-1111-1111-111111111111\",\n        null\n      ],\n      \"InboundEndRecordingPrompt\": [\n        \"\"\n      ],\n      \"OutboundStartRecordingPrompt\": [\n        \"future-reference\"\n      ],\n      \"OutboundEndRecordingPrompt\": [\n        \"11111111-1111-1111-1111-111111111111\"\n      ],\n      \"MeetingStartRecordingPrompt\": [\n        \"22222222-2222-2222-2222-222222222222\"\n      ],\n      \"MeetingEndRecordingPrompt\": [],\n      \"Id\": \"22222222-2222-2222-2222-222222222222\",\n      \"Name\": null,\n      \"DefaultLocale\": \"\"\n    }\n  ],\n  \"Force\": \"False\"\n}",
					CmdletName: "Set-CsTeamsAudioConferencingCustomPromptsConfiguration",
				},
			},
			{
				Name:      "patch-force-true",
				Operation: "patch",
				Fields:    "all",
				Status:    400,
				ForceMode: 1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "PATCH",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAudioConferencingCustomPromptsConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					Body:       "{\n  \"Identity\": \"read-model-identity\",\n  \"Prompts\": [\n    {\n      \"Identity\": \"prompt-meta\",\n      \"Id\": \"11111111-1111-1111-1111-111111111111\",\n      \"Name\": \"\",\n      \"AudioPrompt\": null,\n      \"TextPrompt\": \"text\",\n      \"Type\": \"FutureType\",\n      \"Locale\": \"xx-future\"\n    }\n  ],\n  \"Packages\": [\n    {\n      \"Identity\": \"package-meta\",\n      \"InboundStartRecordingPrompt\": [\n        \"11111111-1111-1111-1111-111111111111\",\n        \"11111111-1111-1111-1111-111111111111\",\n        null\n      ],\n      \"InboundEndRecordingPrompt\": [\n        \"\"\n      ],\n      \"OutboundStartRecordingPrompt\": [\n        \"future-reference\"\n      ],\n      \"OutboundEndRecordingPrompt\": [\n        \"11111111-1111-1111-1111-111111111111\"\n      ],\n      \"MeetingStartRecordingPrompt\": [\n        \"22222222-2222-2222-2222-222222222222\"\n      ],\n      \"MeetingEndRecordingPrompt\": [],\n      \"Id\": \"22222222-2222-2222-2222-222222222222\",\n      \"Name\": null,\n      \"DefaultLocale\": \"\"\n    }\n  ],\n  \"Force\": \"True\"\n}",
					CmdletName: "Set-CsTeamsAudioConferencingCustomPromptsConfiguration",
				},
			},
			{
				Name:      "read-empty",
				Operation: "read",
				Fields:    "all",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAudioConferencingCustomPromptsConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsAudioConferencingCustomPromptsConfiguration",
				},
			},
			{
				Name:      "read-null",
				Operation: "read",
				Fields:    "all",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAudioConferencingCustomPromptsConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsAudioConferencingCustomPromptsConfiguration",
				},
			},
			{
				Name:      "read-null-prompt-id",
				Operation: "read",
				Fields:    "all",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAudioConferencingCustomPromptsConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsAudioConferencingCustomPromptsConfiguration",
				},
			},
			{
				Name:      "read-null-package-locale",
				Operation: "read",
				Fields:    "all",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAudioConferencingCustomPromptsConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsAudioConferencingCustomPromptsConfiguration",
				},
			},
			{
				Name:      "list-envelope",
				Operation: "list",
				Fields:    "all",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAudioConferencingCustomPromptsConfiguration",
					CmdletName: "Get-CsTeamsAudioConferencingCustomPromptsConfiguration",
				},
			},
			{
				Name:      "read-denied",
				Operation: "read",
				Fields:    "all",
				Status:    400,
				ForceMode: -1,
				Request: struct {
					Method, Url, Body, CmdletName string
					FailureType                   *string
				}{
					Method:     "GET",
					Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsAudioConferencingCustomPromptsConfiguration/configuration/scope%20/Fixture%20%2F%20Child",
					CmdletName: "Get-CsTeamsAudioConferencingCustomPromptsConfiguration",
				},
			},
		},
	}

	return v.Cases
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
		t.Fatalf("prompts contract differs: %v %+v", e, r.Differences())
	}
	return r
}
