package teamsmeetingconfiguration_test

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_configuration"
)

// The own pinned helper/serializer probe supplies these authored wire branches;
// none is a captured service acceptance or a tenant persistence claim.
func TestUnitContextTypedWriterDropsRawFieldsAndPreservesReadView(t *testing.T) {
	t.Parallel()
	raw := []byte(`{"Id":"c1","PublishedName":37,"Published":1,"Future":true}`)
	var item resource.PublishedEntraAuthenticationContext
	if json.Unmarshal(raw, &item) != nil {
		t.Fatal("own accepted read was rejected")
	}
	body := json.RawMessage(`{"Identity":"Global","PublishedEntraAuthenticationContexts":[{"Identity":null,"Id":"c1","PublishedName":"37","PublishedDescription":null,"EntraName":null,"EntraDescription":null,"Published":true},null]}`)
	run(t, contracttest.Case{Name: "typed-context-write", Source: "msteams/teams_admin_api/teams_meeting_configuration/context_write_test.go", Steps: []contracttest.Exchange{compiledWire("PATCH", selected, set, body, 204, "")}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		value, response, err := sdk.TeamsAdminAPI.TeamsMeetingConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{PublishedEntraAuthenticationContexts: policy.Set([]*resource.PublishedEntraAuthenticationContext{&item, nil})}})
		if err == nil && (value != nil || !bytes.Equal(item.RawJSON, raw)) {
			t.Fatal("typed wire mutated raw read view or invented a204model")
		}
		return contracttest.Observation{Response: response, Err: err}
	}})
}

func TestUnitContextModifiersKeepNullMembersAndHelperStageFailures(t *testing.T) {
	t.Parallel()
	id := "c1"
	entry := &resource.PublishedEntraAuthenticationContext{ID: &id}
	for _, tc := range []struct {
		name, current, sent string
		edit                *resource.AuthenticationContextEdit
		want                error
	}{
		{"null-add", "[]", "[null]", resource.AddAuthenticationContexts(nil), nil},
		{"null-replace", "[]", "[null]", resource.ReplaceAuthenticationContexts(nil), nil},
		{"remove-empty", "[]", "[]", resource.ChangeAuthenticationContexts(nil, []*resource.PublishedEntraAuthenticationContext{entry}), nil},
		{"null-remove-empty", "[]", "", resource.ChangeAuthenticationContexts(nil, []*resource.PublishedEntraAuthenticationContext{nil}), resource.ErrInput},
		{"remove-populated", `[{"Id":"c1"}]`, "", resource.ChangeAuthenticationContexts(nil, []*resource.PublishedEntraAuthenticationContext{entry}), resource.ErrInput},
	} {
		t.Run(tc.name, func(t *testing.T) {
			steps := []contracttest.Exchange{compiledWire("GET", selected, set, nil, 200, withContexts(t, tc.current))}
			if tc.want == nil {
				steps = append(steps, compiledWire("PATCH", selected, set, json.RawMessage(`{"Identity":"Global","PublishedEntraAuthenticationContexts":`+tc.sent+`}`), 204, ""))
			}
			run(t, contracttest.Case{Name: tc.name, Source: "msteams/teams_admin_api/teams_meeting_configuration/context_write_test.go", Steps: steps, Want: contracttest.Expected{Error: tc.want}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
				value, response, err := sdk.TeamsAdminAPI.TeamsMeetingConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{PublishedEntraAuthenticationContexts: tc.edit})
				if value == nil || value.Before == nil || value.ReadResponse == nil || value.ReadResponse.StatusCode() != 200 {
					t.Fatal("helper outcome lost actual preparation")
				}
				if tc.want != nil {
					if !errors.Is(err, tc.want) || value.PatchInvoked || value.HTTPAccepted || value.After != nil {
						t.Fatal("helper failure submittedPATCH or lost its stage")
					}
				} else if err != nil || !value.PatchInvoked || !value.HTTPAccepted || value.After != nil {
					t.Fatal("accepted modifier lost receipt semantics")
				}
				return contracttest.Observation{Response: response, Err: err}
			}})
		})
	}
}

func TestUnitContextEditSnapshotsOwnPointersAndRawBytes(t *testing.T) {
	t.Parallel()
	id, name := "c1", "original"
	flag := true
	item := &resource.PublishedEntraAuthenticationContext{ID: &id, PublishedName: &name, Published: &flag, RawJSON: json.RawMessage(`{"Id":"c1"}`)}
	edit := resource.ReplaceAuthenticationContexts(item)
	name = "changed"
	flag = false
	item.RawJSON[0] = ' '
	values, err := edit.Apply(nil)
	if err != nil || len(values) != 1 || *values[0].PublishedName != "original" || !*values[0].Published || values[0].RawJSON[0] != '{' {
		t.Fatal("edit retained caller aliases")
	}
	*values[0].PublishedName = "result mutation"
	values[0].RawJSON[0] = ' '
	again, err := edit.Apply(nil)
	if err != nil || *again[0].PublishedName != "original" || again[0].RawJSON[0] != '{' {
		t.Fatal("result mutated edit snapshot")
	}
}
