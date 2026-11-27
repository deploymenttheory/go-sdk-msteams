package teamsmessagingconfiguration

import (
	"bytes"
	"encoding/json"
	"strconv"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// readStoryline is a received-JSON profile, separate from the typed write DTO's
// JSON convenience conversion and PowerShell's raw-argument enum binder.
type readStoryline struct{ Value *Storyline }

func (v *readStoryline) UnmarshalJSON(raw []byte) error {
	raw = bytes.TrimSpace(raw)
	if bytes.Equal(raw, []byte("null")) {
		v.Value = nil
		return nil
	}
	if bytes.Equal(raw, []byte("true")) {
		v.Value = new(StorylineEnabled)
		return nil
	}
	if bytes.Equal(raw, []byte("false")) {
		v.Value = new(StorylineDisabled)
		return nil
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		if text == "" {
			v.Value = nil
			return nil
		}
		text = strings.TrimSpace(text)
		if n, e := strconv.ParseInt(text, 10, 32); e == nil {
			v.Value = new(Storyline(n))
			return nil
		}
		var value Storyline
		for _, part := range strings.Split(text, ",") {
			switch strings.ToLower(strings.TrimSpace(part)) {
			case "disabled":
			case "enabled":
				value |= StorylineEnabled
			default:
				return client.ErrDecode
			}
		}
		v.Value = new(value)
		return nil
	}
	var n json.Number
	if json.Unmarshal(raw, &n) != nil || strings.ContainsAny(n.String(), ".eE") {
		return client.ErrDecode
	}
	// JsonConvert passes Int64 integer tokens to Enum.ToObject, which truncates to
	// the underlying Int32. Bigger BigInteger tokens and all float tokens reject.
	value, e := strconv.ParseInt(n.String(), 10, 64)
	if e != nil {
		return client.ErrDecode
	}
	v.Value = new(Storyline(int32(value)))
	return nil
}
