package teamsmeetingbroadcastconfiguration_test

import (
	"context"
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/testutil"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_broadcast_configuration"
)

const collection = "/Skype.Policy/configurations/TeamsMeetingBroadcastConfiguration"
const selected = collection + "/configuration/Global"
const expose = "?ExposeSDNConfigurationJsonBlob=true"
const get = "Get-CsTeamsMeetingBroadcastConfiguration"
const set = "Set-CsTeamsMeetingBroadcastConfiguration"

func TestUnitReadAndFiltersUseAcceptanceResponses(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "get_global.json")
	run(t, testutil.Case{Name: "read", Steps: []testutil.Exchange{compiledWire("GET", selected, get, nil, 200, string(raw))}, Want: testutil.Expected{JSON: raw}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.TeamsMeetingBroadcastConfigurations.GetByIdentity(ctx, "Global", resource.ReadOptions{})
		return testutil.Observation{Value: v, Response: r, Err: e}
	}})
	list := readResponseFixture(t, "list.json")
	for _, tc := range []struct {
		name, pattern string
		count         int
	}{{"all", " ", 1}, {"identity", "Glo?al", 1}, {"case", "[gG]lobal", 1}, {"missing", "missing", 0}} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: tc.name, Steps: []testutil.Exchange{compiledWire("GET", collection, get, nil, 200, string(list))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsMeetingBroadcastConfigurations.ListFiltered(ctx, tc.pattern, resource.ReadOptions{})
				if e == nil && len(v) != tc.count {
					t.Fatal("local filter selection differs")
				}
				return testutil.Observation{Response: r, Err: e}
			}})
		})
	}
}

// ExposeSDNConfigurationJsonBlob is the pinned wrapper's HttpPipelinePrepend
// query step for reads and lists; the acceptance response is the actual
// exposed Global (an empty provider token in the lab, sanitized otherwise).
func TestUnitExposeSDNAppendsQueryForReadsAndLists(t *testing.T) {
	t.Parallel()
	raw := readResponseFixture(t, "get_global_sdn.json")
	run(t, testutil.Case{Name: "exposed-read", Steps: []testutil.Exchange{compiledWire("GET", selected+expose, get, nil, 200, string(raw))}, Want: testutil.Expected{JSON: raw}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.TeamsMeetingBroadcastConfigurations.GetByIdentity(ctx, "Global", resource.ReadOptions{ExposeSDNConfigurationJsonBlob: true})
		if e == nil && (v.SdnName == nil || v.SdnAPIToken == nil || v.SdnRuntimeConfiguration == nil || v.SdnAttendeeFallbackCount == nil || v.AllowSdnProviderForBroadcastMeeting == nil) {
			t.Fatal("exposed SDN settings not decoded")
		}
		return testutil.Observation{Value: v, Response: r, Err: e}
	}})
	list := "[" + string(raw) + "]"
	run(t, testutil.Case{Name: "exposed-filter", Steps: []testutil.Exchange{compiledWire("GET", collection+expose, get, nil, 200, list)}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		v, r, e := sdk.TeamsAdminAPI.TeamsMeetingBroadcastConfigurations.ListFiltered(ctx, "Global", resource.ReadOptions{ExposeSDNConfigurationJsonBlob: true})
		if e == nil && (len(v) != 1 || v[0].SdnName == nil) {
			t.Fatal("exposed filter selection differs")
		}
		return testutil.Observation{Response: r, Err: e}
	}})
}

func TestUnitUpdateUsesSparsePutAndActualEmptyReceipt(t *testing.T) {
	t.Parallel()
	url, empty, name, license, template, opaque, runtime := "https://support.example.com/broadcast", "", "fixture-provider", "fixture-license", "https://sdn.example.com/{0}", "fixture-provider-value", "{}"
	for _, tc := range []struct {
		name  string
		input resource.UpdateRequest
		body  string
	}{
		{"unbound", resource.UpdateRequest{}, `{"Identity":"Global"}`},
		{"false", resource.UpdateRequest{AllowSdnProviderForBroadcastMeeting: new(false)}, `{"Identity":"Global","AllowSdnProviderForBroadcastMeeting":false}`},
		{"support-url", resource.UpdateRequest{SupportURL: &url}, `{"Identity":"Global","SupportURL":"https://support.example.com/broadcast"}`},
		{"empty-string", resource.UpdateRequest{SupportURL: &empty}, `{"Identity":"Global","SupportURL":""}`},
		{"every-setting", resource.UpdateRequest{AllowSdnProviderForBroadcastMeeting: new(true), SupportURL: &url, SdnProviderName: &name, SdnLicenseID: &license, SdnAPITemplateURL: &template, SdnAPIToken: &opaque, SdnRuntimeConfiguration: &runtime},
			`{"Identity":"Global","AllowSdnProviderForBroadcastMeeting":true,"SupportURL":"https://support.example.com/broadcast","SdnProviderName":"fixture-provider","SdnLicenseId":"fixture-license","SdnApiTemplateUrl":"https://sdn.example.com/{0}","SdnApiToken":"fixture-provider-value","SdnRuntimeConfiguration":"{}"}`},
	} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: tc.name, Steps: []testutil.Exchange{compiledWire("PUT", selected, set, json.RawMessage(tc.body), 204, string(readResponseFixture(t, "update_204.body")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				r, e := sdk.TeamsAdminAPI.TeamsMeetingBroadcastConfigurations.UpdateGlobal(ctx, tc.input)
				return testutil.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitIdentityIsOneEscapedSegment(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "scoped-identity", Steps: []testutil.Exchange{compiledWire("GET", collection+"/configuration/Scope%2Fa%20b%3F%23%25", get, nil, 200, string(readResponseFixture(t, "get_global.json")))}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		_, r, e := sdk.TeamsAdminAPI.TeamsMeetingBroadcastConfigurations.GetByIdentity(ctx, "Scope/a b?#%", resource.ReadOptions{})
		return testutil.Observation{Response: r, Err: e}
	}})
}

func TestUnitMalformedAndUnexpectedSuccessKeepMetadata(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, body string
		status     int
		err        error
	}{{"null", "null", 200, client.ErrDecode}, {"array", "[]", 200, client.ErrDecode}, {"duplicate", `{"Identity":"Global","Identity":"other"}`, 200, client.ErrDecode}, {"nested-case-collision", `{"Future":[{"X":1,"x":2}]}`, 200, client.ErrDecode}, {"unexpected-status", "", 204, resource.ErrResponse}} {
		t.Run(tc.name, func(t *testing.T) {
			run(t, testutil.Case{Name: tc.name, Steps: []testutil.Exchange{compiledWire("GET", selected, get, nil, tc.status, tc.body)}, Want: testutil.Expected{Error: tc.err}, Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
				v, r, e := sdk.TeamsAdminAPI.TeamsMeetingBroadcastConfigurations.GetByIdentity(ctx, "Global", resource.ReadOptions{})
				if v != nil || r == nil || r.StatusCode() != tc.status {
					t.Fatal("failed read lost HTTP response or returned a configuration")
				}
				return testutil.Observation{Response: r, Err: e}
			}})
		})
	}
}

func TestUnitValidationAndCancellationMakeNoRequests(t *testing.T) {
	t.Parallel()
	run(t, testutil.Case{Name: "local", Invoke: func(ctx context.Context, sdk *msteams.Client) testutil.Observation {
		s := sdk.TeamsAdminAPI.TeamsMeetingBroadcastConfigurations
		for _, id := range []string{"", " ", ".", "..", "bad\\id", "bad\nidentity"} {
			if _, _, e := s.GetByIdentity(ctx, id, resource.ReadOptions{}); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid identity reached transport")
			}
			if _, e := s.UpdateByIdentity(ctx, id, resource.UpdateRequest{}); !errors.Is(e, resource.ErrInput) {
				t.Fatal("invalid update identity reached transport")
			}
		}
		if _, _, e := s.ListFiltered(ctx, "[bad", resource.ReadOptions{}); !errors.Is(e, resource.ErrInput) {
			t.Fatal("invalid filter reached transport")
		}
		canceled, stop := context.WithCancel(ctx)
		stop()
		if _, e := s.UpdateGlobal(canceled, resource.UpdateRequest{}); !errors.Is(e, context.Canceled) {
			t.Fatal("cancellation lost")
		}
		if _, _, e := s.GetByIdentity(canceled, "Global", resource.ReadOptions{ExposeSDNConfigurationJsonBlob: true}); !errors.Is(e, context.Canceled) {
			t.Fatal("cancellation lost")
		}
		for _, empty := range []*resource.Service{nil, {}, resource.NewService(nil)} {
			if _, _, e := empty.List(ctx, resource.ReadOptions{}); !errors.Is(e, client.ErrClosed) {
				t.Fatal("uninitialized service accepted")
			}
		}
		return testutil.Observation{}
	}})
}
