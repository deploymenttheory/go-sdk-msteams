package policyrp

import (
	"bytes"
	"encoding/json"
	"math"
	"strconv"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// ReadUint64 opts into the pinned nullable UInt64 received-JSON conversion.
// Integer tokens stay exact; floating tokens use source double/ties-even rounding.
// It is not used to loosen typed write requests or other numeric DTO types.
type ReadUint64 struct{ Value *uint64 }

func (v *ReadUint64) UnmarshalJSON(raw []byte) error {
	raw = bytes.TrimSpace(raw)
	if bytes.Equal(raw, []byte("null")) {
		v.Value = nil
		return nil
	}
	if bytes.Equal(raw, []byte("true")) {
		v.Value = new(uint64(1))
		return nil
	}
	if bytes.Equal(raw, []byte("false")) {
		v.Value = new(uint64(0))
		return nil
	}
	var text string
	if json.Unmarshal(raw, &text) == nil {
		if text == "" {
			v.Value = nil
			return nil
		}
		// NumberStyles.Integer accepts ASCII whitespace and trailing NULs, not
		// arbitrary Unicode whitespace or an embedded/leading NUL.
		text = strings.TrimRight(text, "\x00")
		text = strings.Trim(text, " \t\n\r\v\f")
		if text == "" {
			return client.ErrDecode
		}
		negative := strings.HasPrefix(text, "-")
		if negative || strings.HasPrefix(text, "+") {
			text = text[1:]
		}
		for _, digit := range text {
			if digit < '0' || digit > '9' {
				return client.ErrDecode
			}
		}
		n, e := strconv.ParseUint(text, 10, 64)
		if e != nil || negative && n != 0 {
			return client.ErrDecode
		}
		v.Value = new(n)
		return nil
	}
	var n json.Number
	if json.Unmarshal(raw, &n) != nil {
		return client.ErrDecode
	}
	token := n.String()
	if !strings.ContainsAny(token, ".eE") {
		if token == "-0" {
			v.Value = new(uint64(0))
			return nil
		}
		value, e := strconv.ParseUint(token, 10, 64)
		if e != nil {
			return client.ErrDecode
		}
		v.Value = new(value)
		return nil
	}
	f, e := strconv.ParseFloat(token, 64)
	if e != nil {
		return client.ErrDecode
	}
	f = math.RoundToEven(f)
	if math.IsInf(f, 0) || math.IsNaN(f) || f < 0 || f >= math.Exp2(64) {
		return client.ErrDecode
	}
	v.Value = new(uint64(f))
	return nil
}
