package recordingmessages_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/recording_messages"
)

func TestUnitRecordingUpdatePreparationFailures(t *testing.T) {
	t.Parallel()
	cases := []struct {
		name        string
		read, patch int
		body        string
	}{
		{"read-denied", 403, 0, ""},
		{"read-empty", 204, 0, ""},
		{"read-malformed", 200, 0, "not-json"},
	}
	for _, tc := range cases {
		t.Run(tc.name, func(t *testing.T) {
			steps := []testutil.Exchange{wire("GET", item, "Set-CsTeamsRecordingAndTranscriptionCustomMessage", nil, tc.read, tc.body)}
			if tc.patch != 0 {
				steps = append(steps, wire("PATCH", item, "Set-CsTeamsRecordingAndTranscriptionCustomMessage", json.RawMessage(`{}`), tc.patch, ""))
			}
			run(t, testutil.Case{Name: tc.name, Steps: steps, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.RecordingMessages.UpdateByID(ctx, relative, resource.UpdateRequest{Description: policy.Null[string]()})
				if e == nil || v == nil || v.ReadResponse == nil || v.PatchInvoked != (tc.patch != 0) || v.HTTPAccepted {
					t.Error("failed update lost preparation state or invented acceptance")
				}
				if tc.read == 403 || tc.patch != 0 {
					var apiError *client.APIError
					expected := tc.read
					if tc.patch != 0 {
						expected = tc.patch
					}
					if !errors.As(e, &apiError) || apiError.StatusCode != expected {
						t.Error("failure lost its HTTP status")
					}
				}
				return testutil.Observation{Response: r}
			}})
		})
	}
}

func TestUnitRecordingListEditsStopOnPreparationDenial(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "preparation-denial", Steps: []testutil.Exchange{wire("GET", collection+"/configuration/Global", "Set-CsTeamsRecordingAndTranscriptionCustomMessage", nil, 403, "")}, Want: testutil.Expected{Error: auth.ErrForbidden}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.RecordingMessages.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{Localizations: resource.ChangeLocalizations([]*resource.Localization{{Language: new("fr-FR"), InitiatorImplicit: new("b")}}, nil)})
		if v == nil || v.ReadResponse == nil || v.Before != nil || v.PatchInvoked || v.HTTPAccepted {
			t.Error("failed preparation invented a snapshot, patch or acceptance")
		}
		return testutil.Observation{Response: r, Err: e}
	}})
}

func TestUnitRecordingLocalBuilderHasNoCredentialsOrHTTP(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "local-localization-builder", Want: testutil.Expected{JSON: json.RawMessage(`{"Identity":null,"Language":"Future-Language","InitiatorImplicit":null,"ParticipantImplicit":null,"InitiatorExplicit":null,"ParticipantExplicitRequested":null,"ParticipantExplicitProvided":null,"AgreementDialogue":null}`)}, Invoke: func(_ context.Context, _ *msteams.Client) testutil.Observation {
		v, e := resource.NewLocalization(resource.LocalizationInput{Language: "Future-Language"})
		return testutil.Observation{Value: v, Err: e}
	}})
}
