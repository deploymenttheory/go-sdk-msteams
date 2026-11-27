package recordingmessages_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/recording_messages"
)

func TestUnitRecordingInvalidInputContextAndBackendHaveNoHTTP(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "local-input-guards", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.RecordingMessages
		for _, id := range []string{"", " ", ".", "..", "scope/", "/empty", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid recording identity accepted")}
			}
		}
		for _, pattern := range []string{"", "[broken"} {
			if _, _, e := s.ListFiltered(ctx, pattern); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid recording filter accepted")}
			}
		}
		for _, input := range []resource.PatchRequest{{}, {Settings: resource.Settings{Localizations: policy.Null[[]*resource.Localization]()}}, {Settings: resource.Settings{Description: policy.Set("")}}} {
			if _, _, e := s.PatchByID(ctx, relative, input); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("missing native body ID or invalid settings accepted")}
			}
		}
		if _, _, e := s.PatchByID(ctx, relative, resource.PatchRequest{Settings: resource.Settings{Description: policy.Set("")}, BodyID: policy.Set("invalid")}); !errors.Is(e, resource.ErrInput) {
			return testutil.Observation{Err: errors.New("invalid model body ID accepted")}
		}
		for _, input := range []resource.UpdateRequest{{}, {Controls: policy.Controls{Force: new(true)}}, {Localizations: &resource.LocalizationEdit{}}} {
			if _, _, e := s.UpdateByID(ctx, relative, input); !errors.Is(e, resource.ErrInput) {
				return testutil.Observation{Err: errors.New("invalid recording update performed a read")}
			}
		}
		if _, _, e := s.Create(ctx, resource.CreateRequest{Localizations: &resource.LocalizationEdit{}}); !errors.Is(e, resource.ErrInput) {
			return testutil.Observation{Err: errors.New("invalid auto-create acquired credentials")}
		}
		graph, e := client.NewScope(client.Graph, sdk.TeamsAdminAPI.Scope().Session())
		if e != nil {
			return testutil.Observation{Err: e}
		}
		if _, _, e := resource.NewService(graph).List(ctx); !errors.Is(e, client.ErrDestination) {
			return testutil.Observation{Err: errors.New("wrong recording backend accepted")}
		}
		for _, missing := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := missing.List(ctx); !errors.Is(e, client.ErrClosed) {
				return testutil.Observation{Err: errors.New("missing recording service accepted")}
			}
		}
		canceled, cancel := context.WithCancel(ctx)
		cancel()
		if _, _, e := s.Create(canceled, resource.CreateRequest{}); !errors.Is(e, context.Canceled) {
			return testutil.Observation{Err: errors.New("recording cancellation lost")}
		}
		return testutil.Observation{}
	}})
}

func TestUnitRecordingMutationErrorsKeepStatusAndNeverReplay(t *testing.T) {
	t.Parallel()

	for _, operation := range []string{"create", "patch", "delete"} {
		for _, status := range []int{400, 401, 403, 429, 503} {
			method, command, path, body := "POST", "New-CsTeamsRecordingAndTranscriptionCustomMessage", collection, `{"Identity":"Global/`+relative+`","Id":null}`
			if operation == "patch" {
				method, command, path, body = "PATCH", "Set-CsTeamsRecordingAndTranscriptionCustomMessage", item, `{"Id":"`+relative+`","Description":null}`
			}
			if operation == "delete" {
				method, command, path, body = "DELETE", "Remove-CsTeamsRecordingAndTranscriptionCustomMessage", item, `{"Id":"`+relative+`"}`
			}
			run(t, testutil.Case{Name: fmt.Sprintf("%s-error-%d", operation, status), Steps: []testutil.Exchange{wire(method, path, command, json.RawMessage(body), status, `{"Message":"fixture-private"}`)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				s := sdk.TeamsAdminAPI.RecordingMessages
				var observation testutil.Observation
				switch operation {
				case "create":
					_, r, e := s.CreateByID(ctx, relative, resource.CreateRequest{})
					observation = testutil.Observation{Response: r, Err: e}
				case "patch":
					_, r, e := s.PatchByID(ctx, relative, resource.PatchRequest{Settings: resource.Settings{Description: policy.Null[string]()}, BodyID: policy.Set(relative)})
					observation = testutil.Observation{Response: r, Err: e}
				case "delete":
					_, r, e := s.DeleteByID(ctx, relative, policy.Controls{})
					observation = testutil.Observation{Response: r, Err: e}
				}
				var api *client.APIError
				if !errors.As(observation.Err, &api) || api.StatusCode != status || strings.Contains(fmt.Sprint(observation.Err), "fixture-private") {
					return testutil.Observation{Err: errors.New("recording HTTP status/redaction changed")}
				}
				observation.Err = nil
				return observation
			}})
		}
	}
}
