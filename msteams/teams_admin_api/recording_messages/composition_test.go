package recordingmessages_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/recording_messages"
)

func TestUnitRecordingComposedUpdateKeepsReadIDAndFailurePreparation(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	const readID = "00000000-0000-4000-8000-000000000032"
	for _, tc := range []struct {
		name            string
		read, patch     int
		readBody, reply string
		accepted        bool
	}{
		{"success", 200, 204, `{"Identity":"read-model-identity","Id":"` + readID + `","Description":"preserved-before","Future":false}`, "", true},
		{"read-denied", 403, 0, `{"Message":"fixture-private"}`, "", false},
		{"read-empty", 204, 0, "", "", false},
		{"read-malformed", 200, 0, "not-json", "", false},
		{"patch-throttled", 200, 429, `{"Id":"` + readID + `"}`, `{"Message":"fixture-private"}`, false},
		{"patch-unavailable", 200, 503, `{"Id":"` + readID + `"}`, "not-json", false},
		{"accepted-malformed", 200, 200, `{"Id":"` + readID + `"}`, "not-json", true},
	} {
		steps := []contracttest.Exchange{wire("GET", item, "Set-CsTeamsRecordingAndTranscriptionCustomMessage", nil, tc.read, tc.readBody)}
		if tc.patch != 0 {
			steps = append(steps, wire("PATCH", item, "Set-CsTeamsRecordingAndTranscriptionCustomMessage", json.RawMessage(`{"Id":"`+readID+`","Description":null}`), tc.patch, tc.reply))
		}
		results = append(results, run(t, contracttest.Case{Name: "update-" + tc.name, Mappings: []contracttest.Mapping{{Command: "Set-CsTeamsRecordingAndTranscriptionCustomMessage", ParameterSet: "ParentAndRelativeKey"}}, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			v, r, e := sdk.TeamsAdminAPI.RecordingMessages.UpdateByID(ctx, relative, resource.UpdateRequest{Description: policy.Null[string]()})
			if v == nil || v.ReadResponse == nil || v.PatchInvoked != (tc.patch != 0) || v.HTTPAccepted != tc.accepted {
				return contracttest.Observation{Err: errors.New("recording update preparation/acceptance lost")}
			}
			if tc.name == "success" {
				if v.Before == nil || v.Before.Description == nil || *v.Before.Description != "preserved-before" || v.After != nil {
					return contracttest.Observation{Err: errors.New("recording read model was modified or fabricated")}
				}
				return contracttest.Observation{Response: r, Err: e}
			}
			if e == nil || tc.accepted && !errors.Is(e, client.ErrDecode) {
				return contracttest.Observation{Err: errors.New("recording partial update error lost")}
			}
			return contracttest.Observation{Response: r}
		}}))
	}
	evidence(t, results)
}

func TestUnitRecordingComposedListEditUsesDTOProjection(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	const before = `{"Identity":"read-model-identity","Id":"00000000-0000-4000-8000-000000000032","Future":true,"RecordingAndTranscriptionLocalizationCustomMessage":[{"Identity":"first","Language":"en-US","InitiatorImplicit":"a"},{"Identity":"second","Language":"en-US","InitiatorImplicit":"a","Future":false}]}`
	const after = `[{"Identity":"second","Language":"en-US","InitiatorImplicit":"a","ParticipantImplicit":null,"InitiatorExplicit":null,"ParticipantExplicitRequested":null,"ParticipantExplicitProvided":null,"AgreementDialogue":null},{"Identity":null,"Language":"fr-FR","InitiatorImplicit":"b","ParticipantImplicit":null,"InitiatorExplicit":null,"ParticipantExplicitRequested":null,"ParticipantExplicitProvided":null,"AgreementDialogue":null}]`
	for _, byID := range []bool{false, true} {
		name, set, path, body := "identity", "Identity", collection+"/configuration/Global", `{"Identity":"read-model-identity","RecordingAndTranscriptionLocalizationCustomMessage":`+after+`}`
		if byID {
			name, set, path, body = "relative-id", "ParentAndRelativeKey", item, `{"Id":"00000000-0000-4000-8000-000000000032","Identity":"read-model-identity","RecordingAndTranscriptionLocalizationCustomMessage":`+after+`}`
		}
		results = append(results, run(t, contracttest.Case{Name: "list-update-" + name, Mappings: []contracttest.Mapping{{Command: "Set-CsTeamsRecordingAndTranscriptionCustomMessage", ParameterSet: set}}, Steps: []contracttest.Exchange{wire("GET", path, "Set-CsTeamsRecordingAndTranscriptionCustomMessage", nil, 200, before), wire("PATCH", path, "Set-CsTeamsRecordingAndTranscriptionCustomMessage", json.RawMessage(body), 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			remove := &resource.Localization{Language: new("en-US"), InitiatorImplicit: new("a")}
			add := &resource.Localization{Language: new("fr-FR"), InitiatorImplicit: new("b")}
			input := resource.UpdateRequest{Localizations: resource.ChangeLocalizations([]*resource.Localization{add}, []*resource.Localization{remove})}
			var observation contracttest.Observation
			var result *resource.UpdateResult
			if byID {
				v, r, e := sdk.TeamsAdminAPI.RecordingMessages.UpdateByID(ctx, relative, input)
				result = v
				observation = contracttest.Observation{Response: r, Err: e}
			} else {
				v, r, e := sdk.TeamsAdminAPI.RecordingMessages.UpdateByIdentity(ctx, "Global", input)
				result = v
				observation = contracttest.Observation{Response: r, Err: e}
			}
			if observation.Err == nil && (result == nil || result.Before == nil || len(result.Before.Localizations) != 2 || result.Before.RawFields["Future"] == nil || result.Before.Localizations[1].RawFields["Future"] == nil || !result.HTTPAccepted) {
				observation.Err = errors.New("read snapshot or acceptance was lost")
			}
			return observation
		}}))
	}
	evidence(t, results)
}

func TestUnitRecordingLocalBuilderHasNoCredentialsOrHTTP(t *testing.T) {
	t.Parallel()
	result := run(t, contracttest.Case{Name: "local-localization-builder", Source: "contracts/endpoints/8.1.0/fixtures/recording-message-models.json", Mappings: []contracttest.Mapping{{Command: "New-CsTeamsRecordingAndTranscriptionLocalizationCustomMessage", ParameterSet: "Identity"}}, Want: contracttest.Expected{JSON: json.RawMessage(`{"Identity":null,"Language":"Future-Language","InitiatorImplicit":null,"ParticipantImplicit":null,"InitiatorExplicit":null,"ParticipantExplicitRequested":null,"ParticipantExplicitProvided":null,"AgreementDialogue":null}`)}, Invoke: func(_ context.Context, _ *msteams.Client) contracttest.Observation {
		v, e := resource.NewLocalization(resource.LocalizationInput{Language: "Future-Language"})
		return contracttest.Observation{Value: v, Err: e}
	}})
	evidence(t, []contracttest.Result{result})
}
