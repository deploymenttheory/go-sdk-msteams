package teamsartifact

import (
	"encoding/json"
	"errors"
	"strings"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// ErrInput identifies options the SDK rejects before sending a request.
var ErrInput = errors.New("invalid teams artifact request")

// Type selects one artifact family, as Get-TeamsArtifacts -ArtifactType does.
type Type int

const (
	// RecordingTranscript selects .mp4 files under a "Recordings" folder
	// (including Recordings/View Only).
	RecordingTranscript Type = iota
	// Notes selects .loop files under a "Meetings" folder.
	Notes
	// Whiteboard selects .whiteboard files under a "Whiteboards" folder.
	Whiteboard
)

func (t Type) valid() bool { return t >= RecordingTranscript && t <= Whiteboard }

// folderType maps an artifact folder name (case-insensitive) to its Type.
func folderType(name string) (Type, bool) {
	switch {
	case strings.EqualFold(name, "Recordings"):
		return RecordingTranscript, true
	case strings.EqualFold(name, "Meetings"):
		return Notes, true
	case strings.EqualFold(name, "Whiteboards"):
		return Whiteboard, true
	}
	return 0, false
}

// Options are Get-TeamsArtifacts' parameters. With neither OneDrive nor
// SharePoint set, every SharePoint site drive and then every user's OneDrive
// is scanned. OneDrive limits the scan to the user whose mail equals the value;
// SharePoint limits it to site drives. Setting both is invalid. Start/End
// omit artifacts last modified before/after the instant. The module's -Usage
// switch only prints help and its EXPORT_TOOL.* environment credentials are
// replaced by the client's Graph session.
type Options struct {
	OneDrive   string
	SharePoint bool
	Type       *Type
	Start      *time.Time
	End        *time.Time
	// Paging bounds each Graph collection read (sites, drives, users, children).
	Paging client.PagingOptions
}

// Artifact is one output record: a drive item's name, file MIME type,
// pre-authenticated download URL and last modification time, plus where it was
// found. DownloadURL is a short-lived credential-bearing URL; treat it as secret.
type Artifact struct {
	Name                 string
	MimeType             string
	DownloadURL          string
	LastModifiedDateTime string
	// Drive is the Graph drive path the item was found under, for example
	// "users/{id}/drive" or "drives/{id}".
	Drive string
	// Folder is the folder path from the drive root.
	Folder []string
	// Item is the complete drive item.
	Item *DriveItem
}

// Failure records one scan step that failed and was skipped. The module prints
// these and continues; so does the SDK.
type Failure struct {
	// Stage is "sites", "site-drives", "users" or "drive".
	Stage string
	// ID is the site, user or drive path the stage addressed ("" for searches).
	ID  string
	Err error
}

// Result is the scan outcome. Failures are non-fatal; inspect them before
// treating Artifacts as complete.
type Result struct {
	Artifacts []Artifact
	Failures  []Failure
}

// DriveItem is the subset of a Graph driveItem the scan reads. RawFields keeps
// the complete object.
type DriveItem struct {
	ID                   string                     `json:"id"`
	Name                 string                     `json:"name"`
	Folder               json.RawMessage            `json:"folder"`
	File                 *DriveFile                 `json:"file"`
	DownloadURL          *string                    `json:"@microsoft.graph.downloadUrl"`
	LastModifiedDateTime *string                    `json:"lastModifiedDateTime"`
	RawFields            map[string]json.RawMessage `json:"-"`
}

// DriveFile is the driveItem file facet.
type DriveFile struct {
	MimeType *string `json:"mimeType"`
}

func (v *DriveItem) UnmarshalJSON(raw []byte) error {
	type plain DriveItem
	var value plain
	if json.Unmarshal(raw, &value) != nil || json.Unmarshal(raw, &value.RawFields) != nil || value.RawFields == nil {
		return client.ErrDecode
	}
	*v = DriveItem(value)
	return nil
}

func (Artifact) String() string     { return "teams artifact (redacted)" }
func (v Artifact) GoString() string { return v.String() }

// entity is a site, user or drive reference: only its id is used.
type entity struct {
	ID string `json:"id"`
}
