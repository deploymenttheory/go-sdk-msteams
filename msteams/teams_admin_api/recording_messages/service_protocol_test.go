package recordingmessages_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/recording_messages"
)

// Authored DTOs/receipts from the pinned Core recording/localization properties
// and identity/relative-ID bindings; not nonempty tenant captures. The public
// command supports localized messages, while current documentation warns that
// client effect is not fully released. Nothing here qualifies meeting display.
const authoredRecordingLocalization = `{"Identity":null,"Language":"en-US","InitiatorImplicit":"fixture initiator","ParticipantImplicit":"","InitiatorExplicit":null,"ParticipantExplicitRequested":"fixture requested","ParticipantExplicitProvided":null,"AgreementDialogue":"fixture agreement"}`
const authoredRecordingIdentity = "Global/" + relative
const authoredRecordingMessage = `{"Identity":"` + authoredRecordingIdentity + `","Id":"` + relative + `","Description":"","RecordingAndTranscriptionLocalizationCustomMessage":[` + authoredRecordingLocalization + `],"Future":false}`

func verifyAuthoredRecordingMessage(t *testing.T, value *resource.Message) {
	t.Helper()
	if value == nil || value.Identity == nil || *value.Identity != authoredRecordingIdentity || value.ID == nil || *value.ID != relative || value.Description == nil || *value.Description != "" || len(value.Localizations) != 1 || value.Localizations[0] == nil || value.Localizations[0].Language == nil || *value.Localizations[0].Language != "en-US" || value.Localizations[0].ParticipantImplicit == nil || *value.Localizations[0].ParticipantImplicit != "" || value.Localizations[0].InitiatorExplicit != nil || string(value.Localizations[0].RawFields["InitiatorExplicit"]) != "null" || string(value.RawFields["Future"]) != "false" {
		t.Fatal("recording message lost typed identity, localization, empty/null or future fields")
	}
}

func TestUnitRecordingAuthoredSelectorResponses(t *testing.T) {
	t.Parallel()
	const get = "Get-CsTeamsRecordingAndTranscriptionCustomMessage"
	const create = "New-CsTeamsRecordingAndTranscriptionCustomMessage"
	const patch = "Set-CsTeamsRecordingAndTranscriptionCustomMessage"
	const remove = "Remove-CsTeamsRecordingAndTranscriptionCustomMessage"
	for _, tc := range []struct {
		name, operation, selector, method, command, path string
		body                                             json.RawMessage
		status                                           int
	}{
		{"get-identity", "get", "identity", "GET", get, item, nil, 200},
		{"get-id", "get", "id", "GET", get, item, nil, 200},
		{"create-identity", "create", "identity", "POST", create, collection, json.RawMessage(`{"Identity":"` + authoredRecordingIdentity + `","Id":"` + relative + `","Description":"","RecordingAndTranscriptionLocalizationCustomMessage":[` + authoredRecordingLocalization + `]}`), 201},
		{"create-id", "create", "id", "POST", create, collection, json.RawMessage(`{"Identity":"` + authoredRecordingIdentity + `","Id":null,"Description":"","RecordingAndTranscriptionLocalizationCustomMessage":[` + authoredRecordingLocalization + `]}`), 201},
		{"patch-identity", "patch", "identity", "PATCH", patch, item, json.RawMessage(`{"Identity":"` + authoredRecordingIdentity + `","Description":null}`), 204},
		{"patch-id", "patch", "id", "PATCH", patch, item, json.RawMessage(`{"Id":"` + relative + `","Description":null}`), 204},
		{"remove-identity", "remove", "identity", "DELETE", remove, item, json.RawMessage(`{"Identity":"` + authoredRecordingIdentity + `"}`), 204},
		{"remove-id", "remove", "id", "DELETE", remove, item, json.RawMessage(`{"Id":"` + relative + `"}`), 204},
	} {
		t.Run(tc.name, func(t *testing.T) {
			set, reply := "Identity", ""
			if tc.selector == "id" {
				set = "ParentAndRelativeKey"
			}
			want := contracttest.Expected{Body: []byte{}}
			if tc.operation == "get" {
				reply = authoredRecordingMessage
				want = contracttest.Expected{JSON: json.RawMessage(reply), Body: []byte(reply)}
			}
			run(t, contracttest.Case{Name: "typed-" + tc.name, Source: "msteams/teams_admin_api/recording_messages/service_protocol_test.go", Mappings: []contracttest.Mapping{{Command: tc.command, ParameterSet: set}}, Steps: []contracttest.Exchange{wire(tc.method, tc.path, tc.command, tc.body, tc.status, reply)}, Want: want, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.RecordingMessages
				var observed contracttest.Observation
				var value *resource.Message
				switch tc.operation {
				case "get":
					if tc.selector == "id" {
						value, observed.Response, observed.Err = s.GetByID(ctx, relative)
					} else {
						value, observed.Response, observed.Err = s.GetByIdentity(ctx, authoredRecordingIdentity)
					}
				case "create":
					input := resource.CreateRequest{Description: policy.Set(""), Localizations: resource.ReplaceLocalizations(fixtureLocalization())}
					if tc.selector == "id" {
						value, observed.Response, observed.Err = s.CreateByID(ctx, relative, input)
					} else {
						value, observed.Response, observed.Err = s.CreateByIdentity(ctx, authoredRecordingIdentity, input)
					}
				case "patch":
					input := resource.PatchRequest{Settings: resource.Settings{Description: policy.Null[string]()}}
					if tc.selector == "id" {
						input.BodyID = policy.Set(relative)
						value, observed.Response, observed.Err = s.PatchByID(ctx, relative, input)
					} else {
						input.BodyIdentity = policy.Set(authoredRecordingIdentity)
						value, observed.Response, observed.Err = s.PatchByIdentity(ctx, authoredRecordingIdentity, input)
					}
				default:
					if tc.selector == "id" {
						value, observed.Response, observed.Err = s.DeleteByID(ctx, relative, policy.Controls{})
					} else {
						value, observed.Response, observed.Err = s.DeleteByIdentity(ctx, authoredRecordingIdentity, policy.Controls{})
					}
				}
				if observed.Err == nil {
					if tc.operation == "get" {
						verifyAuthoredRecordingMessage(t, value)
					} else if value != nil {
						t.Fatal("empty recording receipt fabricated a readback")
					}
				}
				observed.Value = value
				return observed
			}})
		})
	}
}

func TestUnitRecordingAuthoredNonemptyFilterPreservesTypedItem(t *testing.T) {
	t.Parallel()
	other := strings.ReplaceAll(authoredRecordingMessage, relative, "00000000-0000-4000-8000-000000000032")
	reply := "[" + other + "," + authoredRecordingMessage + "]"
	want := "[" + authoredRecordingMessage + "]"
	run(t, contracttest.Case{Name: "nonempty-recording-filter", Source: "msteams/teams_admin_api/recording_messages/service_protocol_test.go", Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsRecordingAndTranscriptionCustomMessage", ParameterSet: "Filter"}}, Steps: []contracttest.Exchange{wire("GET", collection, "Get-CsTeamsRecordingAndTranscriptionCustomMessage", nil, 200, reply)}, Want: contracttest.Expected{JSON: json.RawMessage(want), Body: []byte(reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		values, response, err := sdk.TeamsAdminAPI.RecordingMessages.ListFiltered(ctx, "*"+relative)
		if err == nil {
			if len(values) != 1 {
				t.Fatal("nonempty recording filter selected the wrong resource")
			}
			verifyAuthoredRecordingMessage(t, values[0])
		}
		return contracttest.Observation{Value: values, Response: response, Err: err}
	}})
}

func TestUnitRecordingAuthoredMalformed200RetainsResponse(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct{ name, reply string }{
		{"invalid-id", `{"Id":"not-a-guid"}`},
		{"null-language", `{"RecordingAndTranscriptionLocalizationCustomMessage":[{"Language":null}]}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, contracttest.Case{Name: tc.name, Source: "msteams/teams_admin_api/recording_messages/service_protocol_test.go", Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsRecordingAndTranscriptionCustomMessage", ParameterSet: "ParentAndRelativeKey"}}, Steps: []contracttest.Exchange{wire("GET", item, "Get-CsTeamsRecordingAndTranscriptionCustomMessage", nil, 200, tc.reply)}, Want: contracttest.Expected{Body: []byte(tc.reply), Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.RecordingMessages.GetByID(ctx, relative)
				if value != nil {
					t.Fatal("malformed recording reply exposed a fabricated result")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
		})
	}
}

func TestUnitRecordingAuthoredComposedDeltaPreservesPreparationAndReadback(t *testing.T) {
	t.Parallel()
	const nextLocalization = `{"Identity":null,"Language":"fr-FR","InitiatorImplicit":null,"ParticipantImplicit":"new message","InitiatorExplicit":null,"ParticipantExplicitRequested":null,"ParticipantExplicitProvided":null,"AgreementDialogue":null}`
	const updated = `{"Identity":"` + authoredRecordingIdentity + `","Id":"` + relative + `","Description":"","RecordingAndTranscriptionLocalizationCustomMessage":[` + nextLocalization + `],"Future":false}`
	// Source comparator removes the first seven-field ordinal match, then adds
	// in order. These are authored protocol values/readback, not persistence
	// evidence from a tenant or proof of the unreleased meeting-client feature.
	steps := []contracttest.Exchange{
		wire("GET", item, "Set-CsTeamsRecordingAndTranscriptionCustomMessage", nil, 200, authoredRecordingMessage),
		wire("PATCH", item, "Set-CsTeamsRecordingAndTranscriptionCustomMessage", json.RawMessage(`{"Identity":"`+authoredRecordingIdentity+`","RecordingAndTranscriptionLocalizationCustomMessage":[`+nextLocalization+`]}`), 204, ""),
		wire("GET", item, "Get-CsTeamsRecordingAndTranscriptionCustomMessage", nil, 200, updated),
	}
	run(t, contracttest.Case{Name: "composed-recording-delta", Source: "msteams/teams_admin_api/recording_messages/service_protocol_test.go", Mappings: []contracttest.Mapping{{Command: "Set-CsTeamsRecordingAndTranscriptionCustomMessage", ParameterSet: "Identity"}, {Command: "Get-CsTeamsRecordingAndTranscriptionCustomMessage", ParameterSet: "Identity"}, {Command: "New-CsTeamsRecordingAndTranscriptionLocalizationCustomMessage", ParameterSet: "Identity"}}, Steps: steps, Want: contracttest.Expected{JSON: json.RawMessage(updated), Body: []byte(updated)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		add, err := resource.NewLocalization(resource.LocalizationInput{Language: "fr-FR", ParticipantImplicit: new("new message")})
		if err != nil {
			t.Fatal(err)
		}
		s := sdk.TeamsAdminAPI.RecordingMessages
		result, response, err := s.UpdateByIdentity(ctx, authoredRecordingIdentity, resource.UpdateRequest{Localizations: resource.ChangeLocalizations([]*resource.Localization{add}, []*resource.Localization{fixtureLocalization()})})
		if err != nil || response == nil || response.StatusCode() != 204 || result == nil || !result.PatchInvoked || !result.HTTPAccepted || result.After != nil || result.ReadResponse == nil || result.ReadResponse.StatusCode() != 200 {
			t.Fatal("composed recording edit lost preparation or confused receipt with readback")
		}
		verifyAuthoredRecordingMessage(t, result.Before)
		value, response, err := s.GetByIdentity(ctx, authoredRecordingIdentity)
		if err == nil && (value == nil || value.Description == nil || *value.Description != "" || len(value.Localizations) != 1 || value.Localizations[0] == nil || value.Localizations[0].Language == nil || *value.Localizations[0].Language != "fr-FR" || value.Localizations[0].ParticipantImplicit == nil || *value.Localizations[0].ParticipantImplicit != "new message") {
			t.Fatal("authored recording readback lost the delta or an omitted scalar")
		}
		verifyAuthoredRecordingMessage(t, result.Before)
		return contracttest.Observation{Value: value, Response: response, Err: err}
	}})
}

func TestUnitRecordingBodyCloseCancellationStopsPreparationBeforePatch(t *testing.T) {
	t.Parallel()
	var stop context.CancelFunc
	step := wire("GET", item, "Set-CsTeamsRecordingAndTranscriptionCustomMessage", nil, 200, authoredRecordingMessage)
	step.Reply.OnBodyClose = func() { stop() }
	run(t, contracttest.Case{Name: "recording-body-close-cancel", Source: "msteams/teams_admin_api/recording_messages/service_protocol_test.go", Mappings: []contracttest.Mapping{{Command: "Set-CsTeamsRecordingAndTranscriptionCustomMessage", ParameterSet: "Identity"}}, Steps: []contracttest.Exchange{step}, Want: contracttest.Expected{Body: []byte(authoredRecordingMessage), Error: context.Canceled}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		operation, cancel := context.WithCancel(ctx)
		defer cancel()
		stop = cancel
		value, response, err := sdk.TeamsAdminAPI.RecordingMessages.UpdateByIdentity(operation, authoredRecordingIdentity, resource.UpdateRequest{Description: policy.Set("changed")})
		// The real shared client gives cancellation priority after body close,
		// before the resource decoder. Preserve its received response, not an
		// invented typed Before or a claim that PATCH was prepared/completed.
		if !errors.Is(err, context.Canceled) || !errors.Is(operation.Err(), context.Canceled) || value == nil || value.Before != nil || value.After != nil || value.PatchInvoked || value.HTTPAccepted || value.ReadResponse == nil || value.ReadResponse.StatusCode() != 200 || response != value.ReadResponse {
			t.Fatal("body-close cancellation lost the actual recording preparation stage")
		}
		return contracttest.Observation{Response: response, Err: err}
	}})
}
