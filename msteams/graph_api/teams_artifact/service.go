// Package teamsartifact finds Teams meeting artifacts (recordings and
// transcripts, Loop meeting notes and whiteboards) stored in OneDrive for
// Business and SharePoint drives, mirroring Get-TeamsArtifacts. It returns the
// artifact metadata, including each item's download URL; it never downloads
// content and writes no file.
package teamsartifact

import (
	"context"
	"net/http"
	"net/url"
	"path"
	"strings"
	"time"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
)

// Service owns no credentials; all requests use the client's Graph scope.
type Service struct{ scope *client.Scope }

// NewService constructs a lazy resource without authentication or HTTP.
func NewService(scope *client.Scope) *Service { return &Service{scope: scope} }

func (s *Service) ready(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil || s.scope == nil || s.scope.Session() == nil {
		return client.ErrClosed
	}
	if s.scope.Backend() != client.Graph {
		return client.ErrDestination
	}
	_, err := s.scope.Session().Info()
	return err
}

func validOptions(o Options) bool {
	if o.OneDrive != "" && (o.SharePoint || strings.TrimSpace(o.OneDrive) == "" || strings.ContainsFunc(o.OneDrive, unicode.IsControl)) {
		return false
	}
	if o.Type != nil && !o.Type.valid() {
		return false
	}
	p := o.Paging
	return (p.Style == 0 || p.Style == client.LinkPaging) && p.Parameter == "" && (p.ExpectedStatus == 0 || p.ExpectedStatus == http.StatusOK)
}

type scan struct {
	s       *Service
	options Options
	result  Result
}

// List scans the selected drives in the module's order: SharePoint sites first
// (unless OneDrive is set), then OneDrive users (unless SharePoint is set).
// A failed search, drive listing or folder read is recorded in Failures and
// skipped, as the module reports and continues; items from pages read before
// the failure are still scanned. The returned error is non-nil
// only for invalid options, a closed client or a done context.
func (s *Service) List(ctx context.Context, options Options) (Result, error) {
	if err := s.ready(ctx); err != nil {
		return Result{}, err
	}
	if !validOptions(options) {
		return Result{}, ErrInput
	}
	options.Paging.Style, options.Paging.ExpectedStatus = client.LinkPaging, http.StatusOK
	run := &scan{s: s, options: options}
	if options.OneDrive == "" {
		if err := run.sharePoint(ctx); err != nil {
			return run.result, err
		}
	}
	if !options.SharePoint {
		if err := run.oneDrive(ctx); err != nil {
			return run.result, err
		}
	}
	return run.result, nil
}

func collect[T any](ctx context.Context, sc *scan, req client.Request) ([]T, error) {
	req.Method, req.Replay = http.MethodGet, client.ReplaySafe
	req.Header = http.Header{"Accept": {"application/json"}}
	pager, err := client.NewPager(sc.s.scope, req, func(data []byte) (client.PageContent[T], error) {
		return client.DecodeCollection[T](data, "value", "@odata.nextLink")
	}, sc.options.Paging)
	if err != nil {
		return nil, err
	}
	result, err := pager.Collect(ctx)
	return result.Items, err
}

// fail records a non-fatal failure, or returns ctx's error when the context
// ended (which stops the whole scan).
func (sc *scan) fail(ctx context.Context, stage, id string, err error) error {
	if ctxErr := ctx.Err(); ctxErr != nil {
		return ctxErr
	}
	sc.result.Failures = append(sc.result.Failures, Failure{Stage: stage, ID: id, Err: err})
	return nil
}

func (sc *scan) sharePoint(ctx context.Context) error {
	sites, listErr := collect[entity](ctx, sc, client.Request{Path: []string{"v1.0", "sites"}, Query: url.Values{"search": {"*"}}})
	for _, site := range sites {
		if site.ID == "" {
			continue
		}
		drives, drivesErr := collect[entity](ctx, sc, client.Request{Path: []string{"v1.0", "sites", site.ID, "drives"}})
		if drivesErr != nil {
			if err := sc.fail(ctx, "site-drives", site.ID, drivesErr); err != nil {
				return err
			}
		}
		for _, drive := range drives {
			if drive.ID == "" {
				continue
			}
			if err := sc.drive(ctx, []string{"drives", drive.ID}); err != nil {
				return err
			}
		}
	}
	if listErr != nil {
		return sc.fail(ctx, "sites", "", listErr)
	}
	return nil
}

func (sc *scan) oneDrive(ctx context.Context) error {
	req := client.Request{Path: []string{"v1.0", "users"}}
	if sc.options.OneDrive != "" {
		req.Query = url.Values{"$filter": {"mail eq '" + strings.ReplaceAll(sc.options.OneDrive, "'", "''") + "'"}}
	}
	users, listErr := collect[entity](ctx, sc, req)
	for _, user := range users {
		if user.ID == "" {
			continue
		}
		if err := sc.drive(ctx, []string{"users", user.ID, "drive"}); err != nil {
			return err
		}
	}
	if listErr != nil {
		return sc.fail(ctx, "users", "", listErr)
	}
	return nil
}

func (sc *scan) drive(ctx context.Context, drive []string) error {
	return sc.traverse(ctx, drive, "root", nil)
}

func (sc *scan) traverse(ctx context.Context, drive []string, itemID string, folder []string) error {
	segments := append(append([]string{"v1.0"}, drive...), "items", itemID, "children")
	items, listErr := collect[DriveItem](ctx, sc, client.Request{Path: segments})
	for index := range items {
		item := &items[index]
		if item.Folder != nil && string(item.Folder) != "null" {
			next := append(append([]string(nil), folder...), item.Name)
			if sc.recurse(next) {
				if err := sc.traverse(ctx, drive, item.ID, next); err != nil {
					return err
				}
			}
			continue
		}
		if sc.output(item, folder) {
			artifact := Artifact{Name: item.Name, Drive: strings.Join(drive, "/"), Folder: append([]string(nil), folder...), Item: item}
			if item.File != nil && item.File.MimeType != nil {
				artifact.MimeType = *item.File.MimeType
			}
			if item.DownloadURL != nil {
				artifact.DownloadURL = *item.DownloadURL
			}
			if item.LastModifiedDateTime != nil {
				artifact.LastModifiedDateTime = *item.LastModifiedDateTime
			}
			sc.result.Artifacts = append(sc.result.Artifacts, artifact)
		}
	}
	if listErr != nil {
		return sc.fail(ctx, "drive", strings.Join(drive, "/"), listErr)
	}
	return nil
}

// recurse is the module's ShouldRecurseFolder: with a type filter, a folder
// named for another artifact type is skipped; folders at depth one and two are
// always entered; deeper, only Recordings/View Only and folders directly
// inside or named for an artifact folder are entered.
func (sc *scan) recurse(folder []string) bool {
	last := folder[len(folder)-1]
	if sc.options.Type != nil {
		if kind, ok := folderType(last); ok && kind != *sc.options.Type {
			return false
		}
	}
	if len(folder) <= 2 {
		return true
	}
	parent := folder[len(folder)-2]
	if strings.EqualFold(parent, "Recordings") && strings.EqualFold(last, "View Only") {
		return true
	}
	_, parentIsArtifact := folderType(parent)
	_, lastIsArtifact := folderType(last)
	return !parentIsArtifact && lastIsArtifact
}

// output is the module's ShouldOutputArtifact: the time window, then the
// extension expected for the containing artifact folder.
func (sc *scan) output(item *DriveItem, folder []string) bool {
	if item.LastModifiedDateTime != nil {
		if modified, err := time.Parse(time.RFC3339Nano, *item.LastModifiedDateTime); err == nil {
			if sc.options.Start != nil && modified.Before(*sc.options.Start) || sc.options.End != nil && modified.After(*sc.options.End) {
				return false
			}
		}
	}
	if len(folder) == 0 {
		return false
	}
	extension := strings.ToUpper(path.Ext(item.Name))
	last := folder[len(folder)-1]
	switch {
	case strings.EqualFold(last, "Recordings") && extension == ".MP4",
		strings.EqualFold(last, "Meetings") && extension == ".LOOP",
		strings.EqualFold(last, "Whiteboards") && extension == ".WHITEBOARD":
		return true
	}
	return len(folder) >= 2 && strings.EqualFold(folder[len(folder)-2], "Recordings") && strings.EqualFold(last, "View Only") && extension == ".MP4"
}
