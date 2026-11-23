package audioprompts

import (
	"bytes"
	"encoding/base64"
	"encoding/json"
	"errors"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// Application IDs accepted by the media storage service. The module normalizes
// an empty value and these names case-insensitively to the canonical spelling.
const (
	TenantGlobal     = "TenantGlobal"
	OrgAutoAttendant = "OrgAutoAttendant"
	HuntGroup        = "HuntGroup"
)

// ErrInput identifies a request rejected locally before discovery or credentials.
var ErrInput = errors.New("audio file input is invalid")

// ErrResponse identifies an undeclared successful HTTP status.
var ErrResponse = errors.New("unexpected audio file response")

// AudioFile is the media storage audio file DTO. Content is the decoded audio
// payload; it is empty in list/detail reads and populated only where the service
// returns bytes. Timestamps and Duration keep their service text unchanged.
// A soft-deleted file remains readable with DeletionTimestampOffset set.
type AudioFile struct {
	ApplicationID                    *string                    `json:"ApplicationId"`
	ID                               *string                    `json:"Id"`
	GroupID                          *string                    `json:"GroupId"`
	ContextID                        *string                    `json:"ContextId"`
	OriginalFilename                 *string                    `json:"OriginalFilename"`
	ConvertedFilename                *string                    `json:"ConvertedFilename"`
	Content                          []byte                     `json:"-"`
	Duration                         *string                    `json:"Duration"`
	DownloadURI                      *string                    `json:"DownloadUri"`
	DownloadURIExpiryTimestampOffset *string                    `json:"DownloadUriExpiryTimestampOffset"`
	UploadedTimestampOffset          *string                    `json:"UploadedTimestampOffset"`
	LastAccessedTimestampOffset      *string                    `json:"LastAccessedTimestampOffset"`
	DeletionTimestampOffset          *string                    `json:"DeletionTimestampOffset"`
	RawFields                        map[string]json.RawMessage `json:"-"`
}

// MarkedForDeletion mirrors the module projection: a deletion timestamp is set.
func (v *AudioFile) MarkedForDeletion() bool {
	return v != nil && v.DeletionTimestampOffset != nil && *v.DeletionTimestampOffset != ""
}

func (v *AudioFile) UnmarshalJSON(raw []byte) error {
	type plain AudioFile
	var value plain
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	content, err := decodeContent(value.RawFields["Content"])
	if err != nil {
		return err
	}
	value.Content = content
	*v = AudioFile(value)
	return nil
}

func (AudioFile) String() string     { return "audio file (redacted)" }
func (v AudioFile) GoString() string { return v.String() }

// decodeContent preserves absent/null as nil and "" as a non-nil empty slice.
// The module accepts standard and URL-safe base64, padded or not.
func decodeContent(raw json.RawMessage) ([]byte, error) {
	if raw == nil || bytes.Equal(bytes.TrimSpace(raw), []byte("null")) {
		return nil, nil
	}
	var text string
	if json.Unmarshal(raw, &text) != nil {
		return nil, client.ErrDecode
	}
	text = strings.NewReplacer("-", "+", "_", "/").Replace(text)
	if pad := len(text) % 4; pad != 0 {
		text += strings.Repeat("=", 4-pad)
	}
	decoded, err := base64.StdEncoding.DecodeString(text)
	if err != nil {
		return nil, client.ErrDecode
	}
	return decoded, nil
}

// ImportRequest uploads one audio file. FileName is the original file name the
// service validates (for example a .wav or .mp3 extension); Content is the raw
// file bytes, which are base64-encoded on the wire. An empty ApplicationID means
// TenantGlobal, as in the module.
type ImportRequest struct {
	ApplicationID string
	FileName      string
	Content       []byte
}

func (ImportRequest) String() string     { return "audio file import request (redacted)" }
func (v ImportRequest) GoString() string { return v.String() }

// GetOptions selects optional detail-read parameters. DurationInMinutes is the
// module's DurationInMin (wire durationInMins); nil omits it.
type GetOptions struct {
	DurationInMinutes *int32
}
