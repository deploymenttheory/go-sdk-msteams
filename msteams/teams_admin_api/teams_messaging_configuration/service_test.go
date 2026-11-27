package teamsmessagingconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_messaging_configuration"
)

const collection = "/Skype.Policy/configurations/TeamsMessagingConfiguration"
const selected = collection + "/configuration/Global"
const get = "Get-CsTeamsMessagingConfiguration"
const set = "Set-CsTeamsMessagingConfiguration"

func TestUnitReadAndFiltersUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "get_global.json")
	run(t, testutil.Case{Name: "read", Steps: []testutil.Exchange{compiledWire("GET", selected, get, nil, 200, string(raw))}, Want: testutil.Expected{JSON: raw}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.TeamsMessagingConfigurations.GetByIdentity(ctx, "Global")
		return testutil.Observation{Value: v, Response: r, Err: e}
	}})
	list := readResponseFixture(t, "list.json")
	for _, tc := range []struct {
		name, pattern string
		count         int
	}{{"all", "*", 1}, {"identity", "Glo?al", 1}, {"case", "[gG]lobal", 1}, {"missing", "missing", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: tc.name, Steps: []testutil.Exchange{compiledWire("GET", collection, get, nil, 200, string(list))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsMessagingConfigurations.ListFiltered(ctx, tc.pattern)
				if e == nil && len(v) != tc.count {
					t.Fatal("local filter selection differs")
				}
				return testutil.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitUpdateUsesAcceptancePreparationAndSparsePatch(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	accepted := readResponseFixture(t, "update_204.body")
	run(t, testutil.Case{Name: "update", Steps: []testutil.Exchange{
		compiledWire("GET", selected, set, nil, 200, string(before)),
		compiledWire("PATCH", selected, set, json.RawMessage(`{"Identity":"Global","EnableVideoMessageCaptions":false}`), 204, string(accepted)),
	}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.TeamsMessagingConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{EnableVideoMessageCaptions: policy.Set(false)}})
		if e == nil && (v == nil || v.Before == nil || v.Before.EnableVideoMessageCaptions == nil || !*v.Before.EnableVideoMessageCaptions || v.ReadResponse == nil || v.ReadResponse.StatusCode() != 200 || !v.PatchInvoked || !v.HTTPAccepted) {
			t.Fatal("update preparation and acceptance receipts lost")
		}
		return testutil.Observation{Response: r, Err: e}
	}})
}

func TestUnitListEditsUseTheActualBlockList(t *testing.T) {
	t.Parallel()
	before := readResponseFixture(t, "get_global.json")
	var original resource.Configuration
	if json.Unmarshal(before, &original) != nil || len(original.FileTypeBlockList) < 2 {
		t.Fatal("permanent populated block list unavailable")
	}
	for _, tc := range []struct {
		name string
		edit *policy.ListEdit
		want []*string
	}{{"clear", policy.ReplaceStrings(), []*string{}}, {"remove", policy.ChangeNullableStrings(nil, original.FileTypeBlockList[:1]), original.FileTypeBlockList[1:]}, {"replace", policy.ReplaceNullableStrings(original.FileTypeBlockList[:1]...), original.FileTypeBlockList[:1]}} {
		t.Run(tc.name, func(t *testing.T) {
			body, err := json.Marshal(map[string]any{"Identity": "Global", "FileTypeBlockList": tc.want})
			if err != nil {
				t.Fatal(err)
			}
			run(t, testutil.Case{Name: tc.name, Steps: []testutil.Exchange{compiledWire("GET", selected, set, nil, 200, string(before)), compiledWire("PATCH", selected, set, body, 204, string(readResponseFixture(t, "update_204.body")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				_, r, e := sdk.TeamsAdminAPI.TeamsMessagingConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{FileTypeBlockList: tc.edit})
				return testutil.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitNullableSettingsAndControlsStayBound(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "nullable", Steps: []testutil.Exchange{compiledWire("PATCH", selected, set, json.RawMessage(`{"Identity":"Global","CustomEmojis":null,"MessagingNotes":"","Storyline":0,"Force":"False"}`), 204, string(readResponseFixture(t, "update_204.body")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		_, r, e := sdk.TeamsAdminAPI.TeamsMessagingConfigurations.PatchByIdentity(ctx, "Global", resource.PatchRequest{Settings: resource.Settings{ScalarSettings: resource.ScalarSettings{CustomEmojis: policy.Null[bool](), MessagingNotes: policy.Set(""), Storyline: policy.Set(resource.StorylineDisabled)}}, Controls: policy.Controls{Force: new(false)}})
		return testutil.Observation{Response: r, Err: e}
	}})
}

func TestUnitMalformedReadDoesNotSendAnUpdate(t *testing.T) {
	t.Parallel()
	for _, invalid := range []string{`null`, `[]`, `{"CustomEmojis":"bad"}`, `{"Storyline":"unknown-enum"}`, `{"FileTypeBlockList":{}}`} {
		t.Run(fmt.Sprintf("shape-%x", invalid), func(t *testing.T) {
			run(t, testutil.Case{Name: "bad-read", Steps: []testutil.Exchange{compiledWire("GET", selected, set, nil, 200, invalid)}, Want: testutil.Expected{Error: client.ErrDecode}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsMessagingConfigurations.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{ScalarSettings: resource.ScalarSettings{CustomEmojis: policy.Set(false)}})
				if v != nil && (v.PatchInvoked || v.HTTPAccepted) {
					t.Fatal("malformed read triggered mutation")
				}
				return testutil.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitValidationAndCancellationMakeNoRequests(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.TeamsMessagingConfigurations
		for _, id := range []string{"", " ", "scope/", "/missing", ".", "..", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid identity reached transport")
			}
		}
		if _, _, e := s.UpdateByIdentity(ctx, "Global", resource.UpdateRequest{}); !errors.Is(e, resource.ErrInput) {
			t.Fatal("unbound update accepted")
		}
		for _, pattern := range []string{"", "[bad"} {
			if _, _, e := s.ListFiltered(ctx, pattern); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid filter reached transport")
			}
		}
		canceled, stop := context.WithCancel(ctx)
		stop()
		if _, _, e := s.UpdateByIdentity(canceled, "Global", resource.UpdateRequest{}); !errors.Is(e, context.Canceled) {
			t.Fatal("cancellation lost")
		}
		for _, empty := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := empty.List(ctx); !errors.Is(e, client.ErrClosed) {
				t.Fatal("uninitialized service accepted")
			}
		}
		return testutil.Observation{}
	}})
}
