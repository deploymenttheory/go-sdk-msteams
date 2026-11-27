package policyrp

import (
	"encoding/json"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"math"
)

// ReadUint32 is the pinned nullable UInt32 received-JSON projection. Its own
// observed conversions match ReadUint64 within the checked UInt32 range.
// Request fields remain typed uint32 and are not widened or reinterpreted.
type ReadUint32 struct{ Value *uint32 }

func (v *ReadUint32) UnmarshalJSON(raw []byte) error {
	var wide ReadUint64
	if json.Unmarshal(raw, &wide) != nil {
		return client.ErrDecode
	}
	if wide.Value == nil {
		v.Value = nil
		return nil
	}
	number := *wide.Value
	if number > math.MaxUint32 {
		return client.ErrDecode
	}
	v.Value = new(uint32(number))
	return nil
}
