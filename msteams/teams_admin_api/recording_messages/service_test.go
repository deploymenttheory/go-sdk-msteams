package recordingmessages_test

import (
	"context"
	"encoding/json"
	"errors"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/auth"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/recording_messages"
)

func fixtureLocalization() *resource.Localization {
	return &resource.Localization{Language: new("en-US"), InitiatorImplicit: new("fixture initiator"), ParticipantImplicit: new(""), ParticipantExplicitRequested: new("fixture requested"), AgreementDialogue: new("fixture agreement")}
}
func TestUnitRecordingPinnedProviderRequests(t *testing.T) {
	t.Parallel()
	var results []contracttest.Result
	for _, f := range captures(t) {
		if f.Selector == "auto" {
			continue
		} // Random native IDs are checked with a memory sender separately.
		for _, mode := range []auth.Mode{auth.Application, auth.Delegated} {
			x := f.Capture.Request
			var body json.RawMessage
			if x.Body != "" {
				body = json.RawMessage(x.Body)
			}
			set := "ParentAndRelativeKey"
			if f.Selector == "identity" {
				set = "Identity"
			}
			c := contracttest.Case{Name: f.Name + "-" + string(mode), Mode: mode, Mappings: []contracttest.Mapping{{Command: x.CmdletName, ParameterSet: set}}, Steps: []contracttest.Exchange{wire(x.Method, strings.TrimPrefix(x.Url, origin), x.CmdletName, body, f.Status, f.Response)}}
			if x.FailureType != nil {
				c.Want.Error = client.ErrDecode
				if f.Status == 403 {
					c.Want.Error = auth.ErrForbidden
				}
			} else if f.Response != "" {
				c.Want.JSON = json.RawMessage(f.Response)
			}
			c.Invoke = func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				s := sdk.TeamsAdminAPI.RecordingMessages
				controls := policy.Controls{}
				if f.Force {
					controls.Force = new(false)
				}
				id := relative
				if f.Selector == "invalid-id" {
					id = "fixture-non-guid"
				}
				if f.Operation == "list" {
					v, r, e := s.List(ctx)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				if f.Operation == "read" {
					if f.Selector == "identity" {
						v, r, e := s.GetByIdentity(ctx, scoped)
						return contracttest.Observation{Value: v, Response: r, Err: e}
					}
					v, r, e := s.GetByID(ctx, id)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				if f.Operation == "delete" {
					if f.Selector == "identity" {
						v, r, e := s.DeleteByIdentity(ctx, scoped, controls)
						return contracttest.Observation{Value: v, Response: r, Err: e}
					}
					v, r, e := s.DeleteByID(ctx, id, controls)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				settings := resource.Settings{}
				input := resource.CreateRequest{Controls: controls}
				if f.Fields != "none" {
					settings.Description = policy.Set("")
					input.Description = policy.Set("")
				}
				if f.Fields == "all" {
					settings.Localizations = policy.Set([]*resource.Localization{fixtureLocalization()})
					input.Localizations = resource.ReplaceLocalizations(fixtureLocalization())
				}
				if f.Operation == "create" {
					if f.Selector == "identity" {
						v, r, e := s.CreateByIdentity(ctx, scoped, input)
						return contracttest.Observation{Value: v, Response: r, Err: e}
					}
					v, r, e := s.CreateByID(ctx, id, input)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				patch := resource.PatchRequest{Settings: settings, Controls: controls, BodyIdentity: policy.Set("read-model-identity")}
				if f.Selector == "identity" {
					v, r, e := s.PatchByIdentity(ctx, scoped, patch)
					return contracttest.Observation{Value: v, Response: r, Err: e}
				}
				patch.BodyID = policy.Set(relative)
				v, r, e := s.PatchByID(ctx, id, patch)
				return contracttest.Observation{Value: v, Response: r, Err: e}
			}
			results = append(results, run(t, c))
		}
	}
	evidence(t, results)
}

func TestUnitRecordingListFilterKeepsSourceSelection(t *testing.T) {
	t.Parallel()
	const reply = `[{"Identity":"global"},{"Identity":"Tag:Fixture"},{"Identity":"Global"},{"Identity":" "}]`
	var results []contracttest.Result
	for _, pattern := range []string{"", "Fixture*", " "} {
		name, set, want := "list", "Identity", "Global,global,Tag:Fixture, "
		if pattern == "Fixture*" {
			name, set, want = "filter-tag", "Filter", "Tag:Fixture"
		}
		if pattern == " " {
			name, set, want = "filter-whitespace", "Filter", " "
		}
		results = append(results, run(t, contracttest.Case{Name: name, Mappings: []contracttest.Mapping{{Command: "Get-CsTeamsRecordingAndTranscriptionCustomMessage", ParameterSet: set}}, Steps: []contracttest.Exchange{wire("GET", collection, "Get-CsTeamsRecordingAndTranscriptionCustomMessage", nil, 200, reply)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
			var observation contracttest.Observation
			if pattern == "" {
				v, r, e := sdk.TeamsAdminAPI.RecordingMessages.List(ctx)
				observation = contracttest.Observation{Value: v, Response: r, Err: e}
			} else {
				v, r, e := sdk.TeamsAdminAPI.RecordingMessages.ListFiltered(ctx, pattern)
				observation = contracttest.Observation{Value: v, Response: r, Err: e}
			}
			values, _ := observation.Value.([]*resource.Message)
			var ids []string
			for _, v := range values {
				if v == nil || v.Identity == nil {
					ids = append(ids, "missing")
				} else {
					ids = append(ids, *v.Identity)
				}
			}
			if observation.Err == nil && strings.Join(ids, ",") != want {
				observation.Err = errors.New("recording local selection changed")
			}
			return observation
		}}))
	}
	evidence(t, results)
}
