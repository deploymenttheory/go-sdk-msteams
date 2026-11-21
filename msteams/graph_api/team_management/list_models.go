package teammanagement

import (
	"encoding/json"
	"errors"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
	"github.com/go-resty/resty/v2"
)

var (
	ErrListingIncomplete = errors.New("team listing did not complete")
	ErrFilterUnavailable = errors.New("team snapshot lacks a requested filter property")
	ErrHydrationStatus   = errors.New("team hydration returned an unexpected successful status")
)

// ListOptions implements Get-Team/Filters. Nil User selects tenant discovery;
// a non-nil empty User explicitly selects the current delegated user. Tenant
// display/mail searches retain server semantics; user filters are exact matches.
// BatchSize is a request width, not a goroutine count. Zero selects 20; throttle
// attempts include the first request (default four), with a ten-second fallback.
// Timeout defaults to five minutes; Traversal defaults to sixteen pages per read.
type ListOptions struct {
	User                *string
	DisplayName         string
	MailNickname        string
	Visibility          string
	Archived            *bool
	BatchSize           int
	MaxThrottleAttempts int
	ThrottleDelay       time.Duration
	Timeout             time.Duration
	Traversal           client.PagingOptions
}

// Discovery retains the exact source snapshots and its completeness. Incomplete
// search results cannot establish a complete intersection or absence.
type Discovery struct {
	Field    string
	Groups   []group.Group
	Joined   []team.Team
	Complete bool
	Response *resty.Response
	Err      error
}

func (Discovery) String() string     { return "team discovery (redacted)" }
func (d Discovery) GoString() string { return d.String() }

// HydrationResponse is one received batch item. Raw data is retained only for
// explicit inspection, including throttles; target URLs are never followed.
type HydrationResponse struct {
	ID      string
	Status  int
	Headers map[string]string
	Body    json.RawMessage
}

func (HydrationResponse) String() string     { return "team hydration response (redacted)" }
func (r HydrationResponse) GoString() string { return r.String() }

// TeamOutcome preserves input order and a failed identity independently of the
// successful output Items. Value may be present with an error (for example an
// unavailable filter property). Filtered means a known nonmatch, not unknown.
type TeamOutcome struct {
	ID           string
	Value        *ManagedTeam
	Filtered     bool
	Err          error
	Attempts     int
	Subresponses []HydrationResponse
	Response     *resty.Response
	Stages       []Stage
}

func (TeamOutcome) String() string     { return "team listing outcome (redacted)" }
func (o TeamOutcome) GoString() string { return o.String() }

// ItemFailure is a typed per-team failure. Formatting redacts IDs and server
// data; Unwrap preserves context, transport and normalized API classifications.
type ItemFailure struct {
	ID    string
	Stage string
	Cause error
}

func (*ItemFailure) Error() string { return "team listing item failed" }

func (ItemFailure) String() string { return "team listing item failed" }
func (f *ItemFailure) Unwrap() error {
	if f == nil {
		return nil
	}
	return f.Cause
}
func (f ItemFailure) GoString() string { return f.String() }

// ListResult separates partial source discovery, attempted hydration, filtered
// nonmatches and successful matches. Complete requires every selected source and
// outcome to finish without a final error; eventual consistency is not atomicity.
type ListResult struct {
	Items             []*ManagedTeam
	Outcomes          []TeamOutcome
	Discovery         []Discovery
	CandidateIDs      []string
	UserID            string
	UserSource        directoryuser.Source
	DiscoveryComplete bool
	Complete          bool
	Stages            []Stage
	Response          *resty.Response
}

func (ListResult) String() string     { return "team listing result (redacted)" }
func (r ListResult) GoString() string { return r.String() }
