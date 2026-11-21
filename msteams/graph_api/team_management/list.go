package teammanagement

import (
	"context"
	"errors"
	"strconv"
	"strings"
	"time"
	"unicode"
	"unicode/utf8"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
)

const (
	DiscoverGroups  Action = "discover-groups"
	ResolveListUser Action = "resolve-list-user"
	DiscoverJoined  Action = "discover-joined-teams"
	HydrateBatch    Action = "hydrate-team-batch"
)

func normalizeListOptions(value ListOptions) (ListOptions, error) {
	if value.BatchSize < 0 || value.BatchSize > 20 || value.MaxThrottleAttempts < 0 || value.ThrottleDelay < 0 || value.Timeout < 0 {
		return value, client.ErrRequest
	}
	if value.Traversal.Style != 0 && value.Traversal.Style != client.LinkPaging || value.Traversal.Parameter != "" || value.Traversal.ExpectedStatus != 0 && value.Traversal.ExpectedStatus != 200 || value.Traversal.MaxPages < 0 || value.Traversal.MaxItems < 0 {
		return value, client.ErrPagination
	}
	for _, text := range []string{value.DisplayName, value.MailNickname, value.Visibility} {
		if !utf8.ValidString(text) || strings.ContainsFunc(text, unicode.IsControl) {
			return value, client.ErrRequest
		}
	}
	if value.BatchSize == 0 {
		value.BatchSize = 20
	}
	if value.MaxThrottleAttempts == 0 {
		value.MaxThrottleAttempts = 4
	}
	if value.ThrottleDelay == 0 {
		value.ThrottleDelay = 10 * time.Second
	}
	if value.Timeout == 0 {
		value.Timeout = 5 * time.Minute
	}
	if value.Traversal.MaxPages == 0 {
		value.Traversal.MaxPages = 16
	}
	if value.User != nil {
		user := *value.User
		value.User = &user
	}
	if value.Archived != nil {
		archived := *value.Archived
		value.Archived = &archived
	}
	return value, nil
}

func (r *ListResult) record(trace *Result) {
	r.Stages = append(r.Stages, trace.Stages...)
	if trace.Response != nil {
		r.Response = trace.Response
	}
}

func (s *Service) discoverGroups(ctx context.Context, options ListOptions, result *ListResult) ([]group.Group, error) {
	fields := []struct{ name, value string }{{"all", ""}}
	if options.DisplayName != "" {
		fields = []struct{ name, value string }{{"displayName", options.DisplayName}}
	}
	if options.MailNickname != "" {
		if options.DisplayName == "" {
			fields = nil
		}
		fields = append(fields, struct{ name, value string }{"mailNickname", options.MailNickname})
	}
	var selected []group.Group
	for index, field := range fields {
		count := true
		query := group.ListOptions{Version: group.Beta, Filter: "resourceProvisioningOptions/Any(x:x eq 'Team')", Top: 999, Count: &count, EventualConsistency: true}
		if field.value != "" {
			query.Search = strconv.Quote(field.name + ":" + field.value)
		}
		list, err := s.groups.ListAll(ctx, query, options.Traversal)
		result.Discovery = append(result.Discovery, Discovery{Field: field.name, Groups: list.Items, Complete: err == nil, Response: list.Response, Err: err})
		trace := Result{}
		trace.record(DiscoverGroups, list.Response, err)
		result.record(&trace)
		if err != nil {
			return nil, err
		}
		items := uniqueGroups(list.Items)
		if index == 0 {
			selected = items
			continue
		}
		matches := make(map[string]bool, len(items))
		for _, item := range items {
			matches[strings.ToLower(item.ID)] = true
		}
		intersection := make([]group.Group, 0, len(selected))
		for _, item := range selected {
			if matches[strings.ToLower(item.ID)] {
				intersection = append(intersection, item)
			}
		}
		selected = intersection
	}
	result.DiscoveryComplete = true
	return selected, nil
}

func uniqueGroups(items []group.Group) []group.Group {
	seen := make(map[string]bool, len(items))
	result := make([]group.Group, 0, len(items))
	for _, item := range items {
		key := strings.ToLower(item.ID)
		if !seen[key] {
			seen[key] = true
			result = append(result, item)
		}
	}
	return result
}

func filtered(value *ManagedTeam, options ListOptions, user bool) (bool, error) {
	unknown := false
	if user && options.DisplayName != "" {
		if value.DisplayName == nil {
			unknown = true
		} else if *value.DisplayName != options.DisplayName {
			return true, nil
		}
	}
	if user && options.MailNickname != "" {
		if value.MailNickname == nil {
			unknown = true
		} else if !strings.EqualFold(*value.MailNickname, options.MailNickname) {
			return true, nil
		}
	}
	if options.Visibility != "" {
		if value.Visibility == nil {
			unknown = true
		} else if !strings.EqualFold(string(*value.Visibility), options.Visibility) {
			return true, nil
		}
	}
	if options.Archived != nil {
		if value.IsArchived == nil {
			unknown = true
		} else if *value.IsArchived != *options.Archived {
			return true, nil
		}
	}
	if unknown {
		return false, ErrFilterUnavailable
	}
	return false, nil
}

func (s *Service) listForUser(ctx context.Context, options ListOptions, result *ListResult) error {
	resolved, response, err := s.users.Resolve(ctx, *options.User, directoryuser.ResolveOptions{AllowCurrent: true, MaxPages: options.Traversal.MaxPages})
	trace := Result{}
	trace.record(ResolveListUser, response, err)
	result.record(&trace)
	if err != nil {
		return err
	}
	result.UserID, result.UserSource = resolved.ID, resolved.Source
	joined, err := s.teams.ListJoined(ctx, resolved.ID, options.Traversal)
	result.Discovery = append(result.Discovery, Discovery{Field: "joined", Joined: joined.Items, Complete: err == nil, Response: joined.Response, Err: err})
	trace = Result{}
	trace.record(DiscoverJoined, joined.Response, err)
	result.record(&trace)
	if err != nil {
		return err
	}
	result.DiscoveryComplete = true
	seen := map[string]bool{}
	for _, item := range joined.Items {
		key := strings.ToLower(item.ID)
		if !seen[key] {
			seen[key] = true
			result.CandidateIDs = append(result.CandidateIDs, item.ID)
		}
	}
	for _, id := range result.CandidateIDs {
		outcome := TeamOutcome{ID: id}
		if err := ctx.Err(); err != nil {
			outcome.Err = &ItemFailure{ID: id, Stage: "context", Cause: err}
			result.Outcomes = append(result.Outcomes, outcome)
			continue
		}
		if !directoryID.MatchString(id) {
			outcome.Err = &ItemFailure{ID: id, Stage: "identity", Cause: client.ErrDecode}
			result.Outcomes = append(result.Outcomes, outcome)
			continue
		}
		pair := Result{ID: id}
		outcome.Attempts = 1
		err := s.read(ctx, &pair)
		outcome.Value, outcome.Response, outcome.Stages = pair.Value, pair.Response, pair.Stages
		result.record(&pair)
		if err == nil {
			outcome.Filtered, err = filtered(outcome.Value, options, true)
		}
		if err != nil {
			outcome.Err = &ItemFailure{ID: id, Stage: "paired-read-or-filter", Cause: err}
		}
		result.Outcomes = append(result.Outcomes, outcome)
	}
	return nil
}

// List implements Get-Team/Filters with complete bounded discovery followed by
// enriched reads. Tenant queries use beta groups and read-only batch hydration;
// user queries use direct joined teams and paired group/team reads. Incomplete
// discovery never establishes complete intersection/absence. Final errors retain
// successful items, failed identities, source prefixes and per-attempt metadata.
func (s *Service) List(ctx context.Context, options ListOptions) (ListResult, error) {
	result := ListResult{Items: []*ManagedTeam{}, Outcomes: []TeamOutcome{}}
	ctx, stop, err := s.start(ctx)
	if err != nil {
		return result, err
	}
	defer stop()
	options, err = normalizeListOptions(options)
	if err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()
	if options.User != nil {
		err = s.listForUser(ctx, options, &result)
	} else {
		var groups []group.Group
		groups, err = s.discoverGroups(ctx, options, &result)
		if err == nil {
			for _, item := range groups {
				result.CandidateIDs = append(result.CandidateIDs, item.ID)
			}
			err = s.hydrateGroups(ctx, groups, options, &result)
		}
	}
	errorsFound := []error{}
	if err != nil {
		errorsFound = append(errorsFound, err)
	}
	for _, outcome := range result.Outcomes {
		if outcome.Err != nil {
			errorsFound = append(errorsFound, outcome.Err)
		} else if !outcome.Filtered && outcome.Value != nil {
			result.Items = append(result.Items, outcome.Value)
		}
	}
	if len(errorsFound) != 0 {
		return result, errors.Join(append([]error{ErrListingIncomplete}, errorsFound...)...)
	}
	result.Complete = true
	return result, nil
}
