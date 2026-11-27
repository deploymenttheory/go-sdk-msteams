package teamsmeetingbroadcastconfiguration_test

import (
	"context"
	"encoding/json"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/internal/contracttest"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_broadcast_configuration"
)

func TestUnitOwnGeneratedReadValuesKeepDynamicJSON(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../docs/implementation/policies/evidence/pol91-broadcast-generated-read.json")
	if err != nil {
		t.Fatal(err)
	}
	var evidence struct {
		Observations []struct {
			InputJSON       string  `json:"input_json"`
			Error           *string `json:"error"`
			ConversionError *string `json:"conversion_error"`
		} `json:"observations"`
	}
	if err := json.Unmarshal(raw, &evidence); err != nil || len(evidence.Observations) != 125 {
		t.Fatal("own generated source observations unavailable", err)
	}
	for i, row := range evidence.Observations {
		// Nonobject callback behavior remains outside this object-view repair.
		if !strings.HasPrefix(row.InputJSON, "{") {
			continue
		}
		t.Run(strconv.Itoa(i), func(t *testing.T) {
			var value resource.Configuration
			err := json.Unmarshal([]byte(row.InputJSON), &value)
			if row.Error != nil || row.ConversionError != nil {
				if err == nil {
					t.Fatal("source-rejected keys accepted")
				}
				return
			}
			if err != nil {
				t.Fatal("dynamic source object rejected", err)
			}
			encoded, err := json.Marshal(value)
			if err != nil || string(encoded) != row.InputJSON {
				t.Fatal("dynamic tokens changed", err)
			}
			if string(value.RawJSON) != row.InputJSON || value.RawFields == nil {
				t.Fatal("owned raw object missing")
			}
		})
	}
}

func TestUnitDynamicReadUsesRootClientAndTypedViews(t *testing.T) {
	t.Parallel()
	body := `{"Identity":"Global","SupportURL":true,"AllowSdnProviderForBroadcastMeeting":"false","SdnAttendeeFallbackCount":"0","SdnName":"provider","SdnApiToken":{},"Future":[1,null]}`
	run(t, contracttest.Case{Name: "dynamic-root-read", Steps: []contracttest.Exchange{compiledWire("GET", selected, get, nil, 200, body)}, Want: contracttest.Expected{JSON: json.RawMessage(body)}, Invoke: func(ctx context.Context, sdk *msteams.Client) contracttest.Observation {
		v, r, err := sdk.TeamsAdminAPI.TeamsMeetingBroadcastConfigurations.GetByIdentity(ctx, "Global", resource.ReadOptions{})
		if err == nil {
			if v == nil || v.Identity == nil || *v.Identity != "Global" || v.SdnName == nil || *v.SdnName != "provider" {
				t.Fatal("matching typed fields lost")
			}
			if v.SupportURL != nil || v.AllowSdnProviderForBroadcastMeeting != nil || v.SdnAttendeeFallbackCount != nil || v.SdnAPIToken != nil {
				t.Fatal("dynamic values were coerced into typed views")
			}
			if string(v.RawFields["SupportURL"]) != "true" || string(v.RawFields["SdnApiToken"]) != "{}" || string(v.RawJSON) != body {
				t.Fatal("original dynamic values lost")
			}
		}
		return contracttest.Observation{Value: v, Response: r, Err: err}
	}})
}

func TestUnitReadSnapshotsOwnInputAndFailedDecodePreservesValue(t *testing.T) {
	t.Parallel()
	body := []byte(`{"Identity":"Global","SupportURL":{"future":true},"AllowSdnProviderForBroadcastMeeting":false,"SdnAttendeeFallbackCount":0}`)
	var value resource.Configuration
	if err := json.Unmarshal(body, &value); err != nil {
		t.Fatal(err)
	}
	before := string(value.RawJSON)
	body[0] = '['
	if value.AllowSdnProviderForBroadcastMeeting == nil || *value.AllowSdnProviderForBroadcastMeeting || value.SdnAttendeeFallbackCount == nil || *value.SdnAttendeeFallbackCount != 0 {
		t.Fatal("false or zero typed view lost")
	}
	if string(value.RawJSON) != before || string(value.RawFields["SupportURL"]) != `{"future":true}` {
		t.Fatal("input aliases owned read snapshots")
	}
	if err := json.Unmarshal([]byte(`{"Identity":"one","identity":"two"}`), &value); err != client.ErrDecode {
		t.Fatal("case collision accepted", err)
	}
	if string(value.RawJSON) != before {
		t.Fatal("failed decode replaced previous snapshot")
	}
}
