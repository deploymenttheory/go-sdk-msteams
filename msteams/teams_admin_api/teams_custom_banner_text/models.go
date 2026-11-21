// Package teamscustombannertext manages compiled Teams custom banner text.
package teamscustombannertext

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

var ErrInput = policy.ErrInput

// Banner retains nullable/omitted known values and an owned original snapshot.
// Text can be absent but an explicitly assigned JSON null fails the source setter.
type Banner struct {
	Identity    *string                    `json:"Identity"`
	ID          *string                    `json:"Id"`
	Text        *string                    `json:"Text"`
	Description *string                    `json:"Description"`
	RawFields   map[string]json.RawMessage `json:"-"`
	RawJSON     json.RawMessage            `json:"-"`
}

var modelID = regexp.MustCompile(`^[a-fA-F0-9]{8}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{4}-[a-fA-F0-9]{12}$`)

func validModelID(value string) bool { return modelID.MatchString(strings.TrimSuffix(value, "\n")) }
func (p *Banner) UnmarshalJSON(raw []byte) error {
	var fields map[string]json.RawMessage
	if json.Unmarshal(raw, &fields) != nil || fields == nil {
		return client.ErrDecode
	}
	var next Banner
	decoder := json.NewDecoder(bytes.NewReader(raw))
	if _, err := decoder.Token(); err != nil {
		return client.ErrDecode
	}
	for decoder.More() {
		token, err := decoder.Token()
		if err != nil {
			return client.ErrDecode
		}
		name, ok := token.(string)
		if !ok {
			return client.ErrDecode
		}
		var value json.RawMessage
		if decoder.Decode(&value) != nil {
			return client.ErrDecode
		}
		var target **string
		switch strings.ToLower(name) {
		case "identity":
			target = &next.Identity
		case "id":
			target = &next.ID
		case "text":
			target = &next.Text
		case "description":
			target = &next.Description
		}
		if target != nil {
			if json.Unmarshal(value, target) != nil {
				return client.ErrDecode
			}
			if strings.EqualFold(name, "Text") && next.Text == nil {
				return client.ErrDecode
			}
			if strings.EqualFold(name, "Id") && next.ID != nil && !validModelID(*next.ID) {
				return client.ErrDecode
			}
		}
	}
	next.RawFields = fields
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*p = next
	return nil
}
func (p Banner) MarshalJSON() ([]byte, error) {
	if p.RawJSON != nil {
		return append([]byte(nil), p.RawJSON...), nil
	}
	type snapshot Banner
	return json.Marshal(snapshot(p))
}
func (Banner) String() string     { return "custom banner text (redacted)" }
func (p Banner) GoString() string { return p.String() }
