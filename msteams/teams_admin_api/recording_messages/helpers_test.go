package recordingmessages_test

import (
	"encoding/json"
	"net/http"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
)

const origin = "https://api.interfaces.records.teams.microsoft.com"
const collection = "/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage"
const relative = "00000000-0000-4000-8000-000000000031"
const scoped = "scope /" + relative + "/ Child"
const item = collection + "/configuration/Global/" + relative

type inputVectors struct {
	Name, Operation, Selector, Fields, Response string
	Status                                      int
	Force                                       bool
	Request                                     struct {
		PreparedIdentity, RelativeId string

		Request struct {
			Method, Url, Body, CmdletName string
			FailureType                   *string
		}
	}
}

func requestInputs(t *testing.T) []inputVectors {
	t.Helper()

	v := struct {
		Cases []inputVectors `json:"cases"`
	}{
		Cases: []inputVectors{
			{
				Name:      "read-identity",
				Operation: "read",
				Selector:  "identity",
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					PreparedIdentity, RelativeId string

					Request struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}
				}{
					PreparedIdentity: "scope /00000000-0000-4000-8000-000000000031/ Child",
					RelativeId:       "00000000-0000-4000-8000-000000000031",

					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "GET",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage/configuration/scope%20/00000000-0000-4000-8000-000000000031%2F%20Child",
						CmdletName: "Get-CsTeamsRecordingAndTranscriptionCustomMessage",
					},
				},
			},
			{
				Name:      "read-relative-id",
				Operation: "read",
				Selector:  "id",
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					PreparedIdentity, RelativeId string

					Request struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}
				}{
					PreparedIdentity: "Global/00000000-0000-4000-8000-000000000031",
					RelativeId:       "00000000-0000-4000-8000-000000000031",

					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "GET",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage/configuration/Global/00000000-0000-4000-8000-000000000031",
						CmdletName: "Get-CsTeamsRecordingAndTranscriptionCustomMessage",
					},
				},
			},
			{
				Name:      "list",
				Operation: "list",
				Selector:  "identity",
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					PreparedIdentity, RelativeId string

					Request struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}
				}{

					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "GET",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage",
						CmdletName: "Get-CsTeamsRecordingAndTranscriptionCustomMessage",
					},
				},
			},
			{
				Name:      "create-identity",
				Operation: "create",
				Selector:  "identity",
				Fields:    "all",
				Status:    400,
				Force:     true,
				Request: struct {
					PreparedIdentity, RelativeId string

					Request struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}
				}{
					PreparedIdentity: "scope /00000000-0000-4000-8000-000000000031/ Child",
					RelativeId:       "00000000-0000-4000-8000-000000000031",

					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method: "POST",
						Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage",
						Body:   "{\n  \"Identity\": \"scope /00000000-0000-4000-8000-000000000031/ Child\",\n  \"RecordingAndTranscriptionLocalizationCustomMessage\": [\n    {\n      \"Identity\": null,\n      \"Language\": \"en-US\",\n      \"InitiatorImplicit\": \"fixture initiator\",\n      \"ParticipantImplicit\": \"\",\n      \"InitiatorExplicit\": null,\n      \"ParticipantExplicitRequested\": \"fixture requested\",\n      \"ParticipantExplicitProvided\": null,\n      \"AgreementDialogue\": \"fixture agreement\"\n    }\n  ],\n  \"Description\": \"\",\n  \"Force\": \"False\",\n  \"Id\": \"00000000-0000-4000-8000-000000000031\"\n}",

						CmdletName: "New-CsTeamsRecordingAndTranscriptionCustomMessage",
					},
				},
			},
			{
				Name:      "create-relative-id",
				Operation: "create",
				Selector:  "id",
				Fields:    "all",
				Status:    400,
				Force:     true,
				Request: struct {
					PreparedIdentity, RelativeId string

					Request struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}
				}{
					PreparedIdentity: "Global/00000000-0000-4000-8000-000000000031",
					RelativeId:       "00000000-0000-4000-8000-000000000031",

					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method: "POST",
						Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage",
						Body:   "{\n  \"Identity\": \"Global/00000000-0000-4000-8000-000000000031\",\n  \"RecordingAndTranscriptionLocalizationCustomMessage\": [\n    {\n      \"Identity\": null,\n      \"Language\": \"en-US\",\n      \"InitiatorImplicit\": \"fixture initiator\",\n      \"ParticipantImplicit\": \"\",\n      \"InitiatorExplicit\": null,\n      \"ParticipantExplicitRequested\": \"fixture requested\",\n      \"ParticipantExplicitProvided\": null,\n      \"AgreementDialogue\": \"fixture agreement\"\n    }\n  ],\n  \"Id\": null,\n  \"Description\": \"\",\n  \"Force\": \"False\"\n}",

						CmdletName: "New-CsTeamsRecordingAndTranscriptionCustomMessage",
					},
				},
			},
			{
				Name:      "create-generated-id",
				Operation: "create",
				Selector:  "auto",
				Fields:    "all",
				Status:    400,
				Force:     true,
				Request: struct {
					PreparedIdentity, RelativeId string

					Request struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}
				}{
					PreparedIdentity: "Global/00000000-0000-4000-8000-000000000031",
					RelativeId:       "00000000-0000-4000-8000-000000000031",

					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method: "POST",
						Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage",
						Body:   "{\n  \"Identity\": \"Global/00000000-0000-4000-8000-000000000031\",\n  \"RecordingAndTranscriptionLocalizationCustomMessage\": [\n    {\n      \"Identity\": null,\n      \"Language\": \"en-US\",\n      \"InitiatorImplicit\": \"fixture initiator\",\n      \"ParticipantImplicit\": \"\",\n      \"InitiatorExplicit\": null,\n      \"ParticipantExplicitRequested\": \"fixture requested\",\n      \"ParticipantExplicitProvided\": null,\n      \"AgreementDialogue\": \"fixture agreement\"\n    }\n  ],\n  \"Description\": \"\",\n  \"Force\": \"False\"\n}",

						CmdletName: "New-CsTeamsRecordingAndTranscriptionCustomMessage",
					},
				},
			},
			{
				Name:      "create-empty-settings",
				Operation: "create",
				Selector:  "id",
				Fields:    "none",
				Status:    400,
				Force:     true,
				Request: struct {
					PreparedIdentity, RelativeId string

					Request struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}
				}{
					PreparedIdentity: "Global/00000000-0000-4000-8000-000000000031",
					RelativeId:       "00000000-0000-4000-8000-000000000031",

					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method: "POST",
						Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage",
						Body:   "{\n  \"Id\": null,\n  \"Force\": \"False\",\n  \"Identity\": \"Global/00000000-0000-4000-8000-000000000031\"\n}",

						CmdletName: "New-CsTeamsRecordingAndTranscriptionCustomMessage",
					},
				},
			},
			{
				Name:      "create-non-guid-id",
				Operation: "create",
				Selector:  "invalid-id",
				Fields:    "description",
				Status:    400,
				Force:     true,
				Request: struct {
					PreparedIdentity, RelativeId string

					Request struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}
				}{
					PreparedIdentity: "Global/fixture-non-guid",
					RelativeId:       "fixture-non-guid",

					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method: "POST",
						Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage",
						Body:   "{\n  \"Id\": null,\n  \"Description\": \"\",\n  \"Force\": \"False\",\n  \"Identity\": \"Global/fixture-non-guid\"\n}",

						CmdletName: "New-CsTeamsRecordingAndTranscriptionCustomMessage",
					},
				},
			},
			{
				Name:      "patch-identity",
				Operation: "patch",
				Selector:  "identity",
				Fields:    "all",
				Status:    400,
				Force:     true,
				Request: struct {
					PreparedIdentity, RelativeId string

					Request struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}
				}{
					PreparedIdentity: "scope /00000000-0000-4000-8000-000000000031/ Child",
					RelativeId:       "00000000-0000-4000-8000-000000000031",

					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method: "PATCH",
						Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage/configuration/scope%20/00000000-0000-4000-8000-000000000031%2F%20Child",
						Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"RecordingAndTranscriptionLocalizationCustomMessage\": [\n    {\n      \"Identity\": null,\n      \"Language\": \"en-US\",\n      \"InitiatorImplicit\": \"fixture initiator\",\n      \"ParticipantImplicit\": \"\",\n      \"InitiatorExplicit\": null,\n      \"ParticipantExplicitRequested\": \"fixture requested\",\n      \"ParticipantExplicitProvided\": null,\n      \"AgreementDialogue\": \"fixture agreement\"\n    }\n  ],\n  \"Description\": \"\",\n  \"Force\": \"False\"\n}",

						CmdletName: "Set-CsTeamsRecordingAndTranscriptionCustomMessage",
					},
				},
			},
			{
				Name:      "patch-relative-id",
				Operation: "patch",
				Selector:  "id",
				Fields:    "all",
				Status:    400,
				Force:     true,
				Request: struct {
					PreparedIdentity, RelativeId string

					Request struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}
				}{
					PreparedIdentity: "Global/00000000-0000-4000-8000-000000000031",
					RelativeId:       "00000000-0000-4000-8000-000000000031",

					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method: "PATCH",
						Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage/configuration/Global/00000000-0000-4000-8000-000000000031",
						Body:   "{\n  \"Identity\": \"read-model-identity\",\n  \"RecordingAndTranscriptionLocalizationCustomMessage\": [\n    {\n      \"Identity\": null,\n      \"Language\": \"en-US\",\n      \"InitiatorImplicit\": \"fixture initiator\",\n      \"ParticipantImplicit\": \"\",\n      \"InitiatorExplicit\": null,\n      \"ParticipantExplicitRequested\": \"fixture requested\",\n      \"ParticipantExplicitProvided\": null,\n      \"AgreementDialogue\": \"fixture agreement\"\n    }\n  ],\n  \"Id\": \"00000000-0000-4000-8000-000000000031\",\n  \"Description\": \"\",\n  \"Force\": \"False\"\n}",

						CmdletName: "Set-CsTeamsRecordingAndTranscriptionCustomMessage",
					},
				},
			},
			{
				Name:      "patch-id-description",
				Operation: "patch",
				Selector:  "id",
				Fields:    "description",
				Status:    400,
				Force:     true,
				Request: struct {
					PreparedIdentity, RelativeId string

					Request struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}
				}{
					PreparedIdentity: "Global/00000000-0000-4000-8000-000000000031",
					RelativeId:       "00000000-0000-4000-8000-000000000031",

					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method: "PATCH",
						Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage/configuration/Global/00000000-0000-4000-8000-000000000031",
						Body:   "{\n  \"Id\": \"00000000-0000-4000-8000-000000000031\",\n  \"Description\": \"\",\n  \"Force\": \"False\"\n}",

						CmdletName: "Set-CsTeamsRecordingAndTranscriptionCustomMessage",
					},
				},
			},
			{
				Name:      "delete-identity",
				Operation: "delete",
				Selector:  "identity",
				Fields:    "none",
				Status:    400,
				Force:     true,
				Request: struct {
					PreparedIdentity, RelativeId string

					Request struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}
				}{
					PreparedIdentity: "scope /00000000-0000-4000-8000-000000000031/ Child",
					RelativeId:       "00000000-0000-4000-8000-000000000031",

					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method: "DELETE",
						Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage/configuration/scope%20/00000000-0000-4000-8000-000000000031%2F%20Child",
						Body:   "{\n  \"Identity\": \"scope /00000000-0000-4000-8000-000000000031/ Child\",\n  \"Force\": \"False\"\n}",

						CmdletName: "Remove-CsTeamsRecordingAndTranscriptionCustomMessage",
					},
				},
			},
			{
				Name:      "delete-relative-id",
				Operation: "delete",
				Selector:  "id",
				Fields:    "none",
				Status:    400,
				Force:     true,
				Request: struct {
					PreparedIdentity, RelativeId string

					Request struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}
				}{
					PreparedIdentity: "Global/00000000-0000-4000-8000-000000000031",
					RelativeId:       "00000000-0000-4000-8000-000000000031",

					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method: "DELETE",
						Url:    "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage/configuration/Global/00000000-0000-4000-8000-000000000031",
						Body:   "{\n  \"Id\": \"00000000-0000-4000-8000-000000000031\",\n  \"Force\": \"False\"\n}",

						CmdletName: "Remove-CsTeamsRecordingAndTranscriptionCustomMessage",
					},
				},
			},
			{
				Name:      "read-invalid-model-id",
				Operation: "read",
				Selector:  "id",
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					PreparedIdentity, RelativeId string

					Request struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}
				}{
					PreparedIdentity: "Global/00000000-0000-4000-8000-000000000031",
					RelativeId:       "00000000-0000-4000-8000-000000000031",

					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "GET",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage/configuration/Global/00000000-0000-4000-8000-000000000031",
						CmdletName: "Get-CsTeamsRecordingAndTranscriptionCustomMessage",
					},
				},
			},
			{
				Name:      "read-null-collection",
				Operation: "read",
				Selector:  "id",
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					PreparedIdentity, RelativeId string

					Request struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}
				}{
					PreparedIdentity: "Global/00000000-0000-4000-8000-000000000031",
					RelativeId:       "00000000-0000-4000-8000-000000000031",

					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "GET",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage/configuration/Global/00000000-0000-4000-8000-000000000031",
						CmdletName: "Get-CsTeamsRecordingAndTranscriptionCustomMessage",
					},
				},
			},
			{
				Name:      "read-denied",
				Operation: "read",
				Selector:  "id",
				Fields:    "none",
				Status:    400,
				Force:     false,
				Request: struct {
					PreparedIdentity, RelativeId string

					Request struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}
				}{
					PreparedIdentity: "Global/00000000-0000-4000-8000-000000000031",
					RelativeId:       "00000000-0000-4000-8000-000000000031",

					Request: struct {
						Method, Url, Body, CmdletName string
						FailureType                   *string
					}{
						Method:     "GET",
						Url:        "https://api.interfaces.records.teams.microsoft.com/Skype.Policy/configurations/TeamsRecordingAndTranscriptionCustomMessage/configuration/Global/00000000-0000-4000-8000-000000000031",
						CmdletName: "Get-CsTeamsRecordingAndTranscriptionCustomMessage",
					},
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
		t.Fatalf("recording contract differs: %v %+v", e, r.Differences())
	}
	return r
}
