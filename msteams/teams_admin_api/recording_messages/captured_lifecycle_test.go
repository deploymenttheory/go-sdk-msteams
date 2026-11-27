package recordingmessages_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"os"
	"slices"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/recording_messages"
)

// TestUnitCapturedOwnedLifecycle reuses the sanitized ID/Identity tenant replies.
// The PATCH acknowledgement is synthetic; read-back, create/delete receipts and
// the typed missing-record response retain the actual captured bytes.
func TestUnitCapturedOwnedLifecycle(t *testing.T) {
	t.Parallel()
	const id = "00000000-0000-4000-8000-000000000031"
	const identity = "Global/" + id
	path := collection + "/configuration/" + identity
	for _, byID := range []bool{true, false} {
		name := "identity"
		if byID {
			name = "id"
		}
		t.Run(name, func(t *testing.T) {
			read := func(stage string) []byte {
				t.Helper()
				b, err := os.ReadFile("testdata/crud_" + name + "_" + stage)
				if err != nil {
					t.Fatal(err)
				}
				return b
			}
			before, after, list, missing := read("create_readback.json"), read("update_readback.json"), read("list.json"), read("get_deleted_404.json")
			createBody := map[string]any{"Identity": identity, "Description": "Fixture description"}
			patchBody := map[string]any{}
			deleteBody := map[string]string{"Identity": identity}
			createBody["RecordingAndTranscriptionLocalizationCustomMessage"] = json.RawMessage(`[{"Identity":null,"Language":"en-US","InitiatorImplicit":null,"ParticipantImplicit":"SDK lab initial recording notice","InitiatorExplicit":null,"ParticipantExplicitRequested":null,"ParticipantExplicitProvided":null,"AgreementDialogue":null}]`)
			patchBody["RecordingAndTranscriptionLocalizationCustomMessage"] = json.RawMessage(`[{"Identity":null,"Language":"en-US","InitiatorImplicit":null,"ParticipantImplicit":"SDK lab updated recording notice","InitiatorExplicit":null,"ParticipantExplicitRequested":null,"ParticipantExplicitProvided":null,"AgreementDialogue":null}]`)
			if byID {
				createBody["Id"] = nil
				patchBody["Id"] = id
				patchBody["Identity"] = identity
				deleteBody = map[string]string{"Id": id}
			} else {
				createBody["Id"] = id
				patchBody["Identity"] = identity
			}
			encode := func(v any) json.RawMessage {
				t.Helper()
				b, err := json.Marshal(v)
				if err != nil {
					t.Fatal(err)
				}
				return b
			}
			steps := []testutil.Exchange{
				wire("POST", collection, "New-CsTeamsRecordingAndTranscriptionCustomMessage", encode(createBody), 201, string(read("create_201.body"))),
				wire("GET", path, "Get-CsTeamsRecordingAndTranscriptionCustomMessage", nil, 200, string(before)),
				wire("GET", path, "Set-CsTeamsRecordingAndTranscriptionCustomMessage", nil, 200, string(before)),
				wire("PATCH", path, "Set-CsTeamsRecordingAndTranscriptionCustomMessage", encode(patchBody), 204, ""),
				wire("GET", path, "Get-CsTeamsRecordingAndTranscriptionCustomMessage", nil, 200, string(after)),
				wire("GET", collection, "Get-CsTeamsRecordingAndTranscriptionCustomMessage", nil, 200, string(list)),
				wire("DELETE", path, "Remove-CsTeamsRecordingAndTranscriptionCustomMessage", encode(deleteBody), 204, string(read("delete_204.body"))),
				wire("GET", path, "Get-CsTeamsRecordingAndTranscriptionCustomMessage", nil, 404, string(missing)),
			}
			for i := range steps {
				steps[i].Reply.Headers.Set("Content-Type", "application/json")
			}
			result, err := testutil.Run(t.Context(), testutil.Case{Name: "captured-" + name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				service := sdk.TeamsAdminAPI.RecordingMessages
				loc, err := resource.NewLocalization(resource.LocalizationInput{Language: "en-US", ParticipantImplicit: new("SDK lab initial recording notice")})
				if err != nil {
					t.Fatal(err)
				}
				input := resource.CreateRequest{Description: policy.Set("Fixture description"), Localizations: resource.ReplaceLocalizations(loc)}
				create, chosen := service.CreateByIdentity, identity
				if byID {
					create, chosen = service.CreateByID, id
				}
				value, response, err := create(ctx, chosen, input)
				if err != nil || value != nil || response == nil || response.StatusCode() != 201 || len(response.Body()) != 0 {
					t.Fatal("empty captured create acceptance became a snapshot")
				}
				beforeValue, response, err := service.GetByIdentity(ctx, identity)
				if err != nil || beforeValue == nil || len(beforeValue.RawFields["ConfigMetadata"]) == 0 || beforeValue.Description == nil || *beforeValue.Description != "Fixture description" || !bytes.Equal(response.Body(), before) {
					t.Fatal("captured initial snapshot lost response or description")
				}
				if len(beforeValue.Localizations) != 1 || beforeValue.Localizations[0] == nil || beforeValue.Localizations[0].ParticipantImplicit == nil || *beforeValue.Localizations[0].ParticipantImplicit != "SDK lab initial recording notice" {
					t.Fatal("captured initial localization changed")
				}
				loc, err = resource.NewLocalization(resource.LocalizationInput{Language: "en-US", ParticipantImplicit: new("SDK lab updated recording notice")})
				if err != nil {
					t.Fatal(err)
				}
				updateInput := resource.UpdateRequest{Localizations: resource.ReplaceLocalizations(loc)}
				update := service.UpdateByIdentity
				if byID {
					update = service.UpdateByID
				}
				changed, response, err := update(ctx, chosen, updateInput)
				if err != nil || changed == nil || !changed.HTTPAccepted || !changed.PatchInvoked || changed.Before == nil || changed.After != nil || response == nil || response.StatusCode() != 204 || changed.ReadResponse == nil || !bytes.Equal(changed.ReadResponse.Body(), before) {
					t.Fatal("prepared update lost its captured before-state or invented a read-back")
				}
				afterValue, response, err := service.GetByID(ctx, id)
				if err != nil || afterValue == nil || len(afterValue.RawFields["ConfigMetadata"]) == 0 || afterValue.Description == nil || *afterValue.Description != "Fixture description" || !bytes.Equal(response.Body(), after) {
					t.Fatal("captured update lost omitted description or response")
				}
				if len(afterValue.Localizations) != 1 || afterValue.Localizations[0] == nil || afterValue.Localizations[0].ParticipantImplicit == nil || *afterValue.Localizations[0].ParticipantImplicit != "SDK lab updated recording notice" || len(changed.Before.Localizations) != 1 || changed.Before.Localizations[0] == nil || changed.Before.Localizations[0].ParticipantImplicit == nil || *changed.Before.Localizations[0].ParticipantImplicit != "SDK lab initial recording notice" {
					t.Fatal("persisted localization or retained before snapshot changed")
				}
				values, _, err := service.List(ctx)
				if err != nil || len(values) != 1 || values[0] == nil || values[0].Identity == nil || *values[0].Identity != identity {
					t.Fatal("captured owned record missing from list")
				}
				remove := service.DeleteByIdentity
				if byID {
					remove = service.DeleteByID
				}
				deleted, response, err := remove(ctx, chosen, policy.Controls{})
				if err != nil || deleted != nil || response == nil || response.StatusCode() != 204 {
					t.Fatal("empty captured delete acceptance lost")
				}
				_, response, err = service.GetByIdentity(ctx, identity)
				var api *client.APIError
				if !errors.As(err, &api) || api.StatusCode != 404 || !slices.Equal(api.ServiceErrorCodes, []int{40407}) || response == nil || !bytes.Equal(response.Body(), missing) {
					t.Fatal("captured absence lost typed status, service code or receipt")
				}
				return testutil.Observation{}
			}})
			if err != nil || !result.Passed() {
				t.Fatalf("captured lifecycle differs: %v %+v", err, result.Differences())
			}
		})
	}
}
