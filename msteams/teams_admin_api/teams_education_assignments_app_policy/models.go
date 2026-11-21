// Package teamseducationassignmentsapppolicy manages compiled assignment app policy.
package teamseducationassignmentsapppolicy

import (
	"bytes"
	"encoding/json"
	"regexp"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
)

var ErrInput = policy.ErrInput

// Policy owns its original snapshot, including unknown future fields. Raw
// snapshots and TurnItInAPIKey may contain sensitive tenant configuration.
type Policy struct {
	Identity                *string                    `json:"Identity"`
	ParentDigestEnabledType *string                    `json:"ParentDigestEnabledType"`
	MakeCodeEnabledType     *string                    `json:"MakeCodeEnabledType"`
	TurnItInEnabledType     *string                    `json:"TurnItInEnabledType"`
	TurnItInAPIURL          *string                    `json:"TurnItInApiUrl"`
	TurnItInAPIKey          *string                    `json:"TurnItInApiKey"`
	RawFields               map[string]json.RawMessage `json:"-"`
	RawJSON                 json.RawMessage            `json:"-"`
}

// The source uses explicit ASCII scheme alternatives and .NET's final-LF $.
// Its suffix is opaque text; this is deliberately not URI validation.
var httpsPattern = regexp.MustCompile(`^[hH][tT][tT][pP][sS]://[^\n]+$`)

func validAPIURL(value string) bool { return httpsPattern.MatchString(strings.TrimSuffix(value, "\n")) }

func (p *Policy) UnmarshalJSON(raw []byte) error {
	var rawFields map[string]json.RawMessage
	if json.Unmarshal(raw, &rawFields) != nil || rawFields == nil {
		return client.ErrDecode
	}
	var next Policy
	d := json.NewDecoder(bytes.NewReader(raw))
	if _, e := d.Token(); e != nil {
		return client.ErrDecode
	}
	for d.More() {
		token, e := d.Token()
		if e != nil {
			return client.ErrDecode
		}
		name, ok := token.(string)
		if !ok {
			return client.ErrDecode
		}
		var value json.RawMessage
		if d.Decode(&value) != nil {
			return client.ErrDecode
		}
		var target **string
		switch strings.ToLower(name) {
		case "identity":
			target = &next.Identity
		case "parentdigestenabledtype":
			target = &next.ParentDigestEnabledType
		case "makecodeenabledtype":
			target = &next.MakeCodeEnabledType
		case "turnitinenabledtype":
			target = &next.TurnItInEnabledType
		case "turnitinapiurl":
			target = &next.TurnItInAPIURL
		case "turnitinapikey":
			target = &next.TurnItInAPIKey
		}
		if target != nil {
			if json.Unmarshal(value, target) != nil {
				return client.ErrDecode
			}
			if strings.EqualFold(name, "TurnItInApiUrl") && next.TurnItInAPIURL != nil && !validAPIURL(*next.TurnItInAPIURL) {
				return client.ErrDecode
			}
		}
	}
	next.RawFields = rawFields
	next.RawJSON = append(json.RawMessage(nil), raw...)
	*p = next
	return nil
}

func (p Policy) MarshalJSON() ([]byte, error) {
	if p.RawJSON != nil {
		return append([]byte(nil), p.RawJSON...), nil
	}
	type snapshot Policy
	return json.Marshal(snapshot(p))
}
func (Policy) String() string     { return "education assignments app policy (redacted)" }
func (p Policy) GoString() string { return p.String() }
