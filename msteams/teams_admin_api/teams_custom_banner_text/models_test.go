package teamscustombannertext_test

import (
	"bytes"
	"encoding/json"
	"errors"
	"os"
	"reflect"
	"testing"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_custom_banner_text"
)

func TestBannerModelSettersMatchPinnedSource(t *testing.T) {
	t.Parallel()
	raw, err := os.ReadFile("../../../" + source)
	if err != nil {
		t.Fatal(err)
	}
	var fixture struct {
		Models []struct {
			Input       string
			Observation struct {
				Failure *string
				Value   json.RawMessage
			}
		}
	}
	if json.Unmarshal(raw, &fixture) != nil || len(fixture.Models) != 17 {
		t.Fatal("invalid model observations")
	}
	for _, m := range fixture.Models {
		t.Run(m.Input, func(t *testing.T) {
			var value resource.Banner
			var err error
			switch m.Input {
			case "constructor":
			case "text-constructor-empty":
				value.Text = new("")
			case "text-constructor-null":
				return
			default:
				err = json.Unmarshal([]byte(m.Input), &value)
			}
			if m.Observation.Failure != nil {
				if !errors.Is(err, client.ErrDecode) {
					t.Fatal("source-invalid model accepted")
				}
				return
			}
			if err != nil {
				t.Fatal("source-valid model rejected")
			}
			type known resource.Banner
			encoded, e := json.Marshal(known(value))
			if e != nil {
				t.Fatal(e)
			}
			var want, got any
			if json.Unmarshal(m.Observation.Value, &want) != nil || json.Unmarshal(encoded, &got) != nil || !reflect.DeepEqual(want, got) {
				t.Fatal("known projections differ from pinned source")
			}
		})
	}
}
func TestBannerRejectsEarlierSetterFailureAndOwnsSnapshot(t *testing.T) {
	t.Parallel()
	for _, raw := range []string{`{"Text":null,"Text":"later"}`, `{"Id":"invalid","Id":"01234567-89ab-cdef-0123-456789abcdef"}`, `{"Id":"01234567-89ab-cdef-0123-456789abcdef\n\n"}`} {
		var value resource.Banner
		if !errors.Is(json.Unmarshal([]byte(raw), &value), client.ErrDecode) {
			t.Fatal("later member bypassed source setter failure")
		}
	}
	raw := []byte(`{"Identity":"Global/fixture","Id":null,"Text":"first","text":"last","Description":null,"Future":{"n":9007199254740993}}`)
	saved := append([]byte(nil), raw...)
	var value resource.Banner
	if json.Unmarshal(raw, &value) != nil || value.Text == nil || *value.Text != "last" {
		t.Fatal("last case-insensitive setter value lost")
	}
	clear(raw)
	value.RawFields["Future"][0] = 'x'
	*value.Text = "changed"
	encoded, e := json.Marshal(value)
	if e != nil || !bytes.Equal(encoded, saved) {
		t.Fatal("snapshot lost bytes/ownership or changed with projection")
	}
}
