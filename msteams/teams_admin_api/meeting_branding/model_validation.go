package meetingbranding

import (
	"bytes"
	"encoding/json"
	"strings"
)

// Missing properties keep their received absence. Explicit nulls on the pinned
// non-null setters, and invalid non-null HTTPS setters, fail native decoding.
func validReadAsset(raw []byte, nonnull, uris []string) bool {
	d := json.NewDecoder(bytes.NewReader(raw))
	if token, err := d.Token(); err != nil || token != json.Delim('{') {
		return false
	}
	for d.More() {
		token, err := d.Token()
		key, ok := token.(string)
		if err != nil || !ok {
			return false
		}
		var encoded json.RawMessage
		if d.Decode(&encoded) != nil {
			return false
		}
		null := bytes.Equal(bytes.TrimSpace(encoded), []byte("null"))
		for _, field := range nonnull {
			if strings.EqualFold(key, field) && null {
				return false
			}
		}
		for _, field := range uris {
			if !strings.EqualFold(key, field) || null {
				continue
			}
			var uri string
			if json.Unmarshal(encoded, &uri) != nil || !httpsPattern.MatchString(uri) {
				return false
			}
		}
	}
	_, err := d.Token()
	return err == nil
}

// Write projections include CLR getter defaults, even when a caller started
// from a sparse read snapshot. Decode copies without re-running read setters on
// default null fields that a default model constructor itself can serialize.
func decodeProjection[T Asset](raw json.RawMessage) (*T, error) {
	if bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}
	var result T
	var err error
	switch value := any(&result).(type) {
	case *BackgroundImage:
		type projection BackgroundImage
		err = json.Unmarshal(raw, (*projection)(value))
	case *AssuranceSlate:
		type projection AssuranceSlate
		err = json.Unmarshal(raw, (*projection)(value))
	case *Reaction:
		type projection Reaction
		err = json.Unmarshal(raw, (*projection)(value))
	case *BrandingTheme:
		type projection BrandingTheme
		var data struct {
			projection
			Reactions []json.RawMessage `json:"MeetingReactions"`
		}
		err = json.Unmarshal(raw, &data)
		if err == nil {
			*value = BrandingTheme(data.projection)
			value.MeetingReactions = []*Reaction{}
			for _, item := range data.Reactions {
				r, e := decodeProjection[Reaction](item)
				if e != nil {
					return nil, e
				}
				value.MeetingReactions = append(value.MeetingReactions, r)
			}
		}
	}
	if err != nil {
		return nil, ErrInput
	}
	return &result, nil
}
