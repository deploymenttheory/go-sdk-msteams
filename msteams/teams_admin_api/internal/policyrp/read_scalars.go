package policyrp

import (
	"bytes"
	"encoding/json"
	"errors"
	"strconv"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// ReadString is an opt-in projection of the pinned provider's nullable string
// reader. It preserves numeric lexemes without a floating-point conversion.
// Resource DTO setters can impose additional validation after this projection.
type ReadString struct{ Value *string }

func (v *ReadString) UnmarshalJSON(raw []byte) error {
	var value *string
	if json.Unmarshal(raw, &value) == nil {
		v.Value = value
		return nil
	}
	raw = bytes.TrimSpace(raw)
	if bytes.Equal(raw, []byte("true")) || bytes.Equal(raw, []byte("false")) {
		v.Value = new(string(raw))
		return nil
	}
	var number json.Number
	if json.Unmarshal(raw, &number) != nil {
		return client.ErrDecode
	}
	v.Value = new(number.String())
	return nil
}

// ReadBool reproduces the pinned provider's nullable Boolean read conversion.
// An empty string is null; other strings must parse as Boolean words. Numeric
// input uses the source's double conversion, including overflow and underflow.
type ReadBool struct{ Value *bool }

func (v *ReadBool) UnmarshalJSON(raw []byte) error {
	raw = bytes.TrimSpace(raw)
	if bytes.Equal(raw, []byte("null")) {
		v.Value = nil
		return nil
	}
	if bytes.Equal(raw, []byte("true")) || bytes.Equal(raw, []byte("false")) {
		v.Value = new(bytes.Equal(raw, []byte("true")))
		return nil
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		if text == "" {
			v.Value = nil
			return nil
		}
		text = strings.TrimFunc(text, func(r rune) bool { return r == 0 || unicode.IsSpace(r) })
		switch {
		case strings.EqualFold(text, "true"):
			v.Value = new(true)
		case strings.EqualFold(text, "false"):
			v.Value = new(false)
		default:
			return client.ErrDecode
		}
		return nil
	}
	var number json.Number
	if json.Unmarshal(raw, &number) != nil {
		return client.ErrDecode
	}
	value, err := strconv.ParseFloat(number.String(), 64)
	if err != nil && !errors.Is(err, strconv.ErrRange) {
		return client.ErrDecode
	}
	// Source nonzero integers are true even if their floating approximation
	// overflows. The same is true for overflowing nonzero JSON float tokens.
	v.Value = new(value != 0)
	return nil
}
