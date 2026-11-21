package channel_test

import (
	"encoding/json"
	"errors"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/channel"
)

func TestChannelModelPreservesPresenceAndUnknownValues(t *testing.T) {
	t.Parallel()
	var value channel.Channel
	data := []byte(`{"id":"fixture","description":null,"membershipType":"future","isArchived":false,"summary":{"future":5},"moderationSettings":{"unknown":true},"future":7}`)
	if json.Unmarshal(data, &value) != nil || value.Description != nil || string(value.RawFields["description"]) != "null" || value.MembershipType == nil || *value.MembershipType != "future" || value.IsArchived == nil || *value.IsArchived || string(value.Summary) != `{"future":5}` || string(value.ModerationSettings) != `{"unknown":true}` || string(value.RawFields["future"]) != "7" {
		t.Fatal("channel snapshot lost presence or unknown data")
	}
	if _, exists := value.RawFields["displayName"]; exists {
		t.Fatal("absent property was synthesized")
	}
}

func TestHostTeamProjectionIsPassiveAndStrict(t *testing.T) {
	t.Parallel()
	for _, tc := range []struct {
		name, address string
		valid         bool
	}{
		{"valid", "https://graph.microsoft.com/beta/tenants/fixture/teams/" + teamID + "/channels/19:fixture%2Fpart@thread.tacv2", true},
		{"opaque-host", "https://metadata.fixture.invalid/beta/tenants/fixture/teams/" + teamID + "/channels/19:fixture%2Fpart@thread.tacv2", true},
		{"ordinary-channel", base + "/channels/19:fixture%2Fpart@thread.tacv2", false},
		{"wrong-channel", "https://graph.microsoft.com/beta/tenants/fixture/teams/" + teamID + "/channels/other", false},
		{"relative", "/beta/tenants/fixture/teams/" + teamID + "/channels/19:fixture%2Fpart@thread.tacv2", false},
		{"missing", "", false},
	} {
		t.Run(tc.name, func(t *testing.T) {
			v := channel.Channel{ID: channelID, ODataID: &tc.address}
			host, e := v.HostTeamID()
			if tc.valid {
				if e != nil || host != teamID {
					t.Fatal("host projection failed")
				}
			} else if !errors.Is(e, channel.ErrProjection) || host != "" {
				t.Fatal("invalid host identity was guessed")
			}
		})
	}
}
