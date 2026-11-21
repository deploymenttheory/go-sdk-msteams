package teammanagement

import (
	"context"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	graphbatch "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/batch"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
)

func hydrationDelay(headers map[string]string, fallback time.Duration) time.Duration {
	var delay time.Duration
	for key, value := range headers {
		if !strings.EqualFold(key, "Retry-After") {
			continue
		}
		value = strings.TrimSpace(value)
		if seconds, err := strconv.ParseUint(value, 10, 63); err == nil && seconds > 0 {
			maximum := uint64((1<<63 - 1) / int64(time.Second))
			if seconds > maximum {
				delay = time.Duration(1<<63 - 1)
			} else {
				delay = max(delay, time.Duration(seconds)*time.Second)
			}
		} else if date, err := http.ParseTime(value); err == nil {
			delay = max(delay, time.Until(date))
		}
	}
	if delay <= 0 {
		return fallback
	}
	return delay
}

func hydrationError(item graphbatch.Response) error {
	if item.Status < 300 {
		return ErrHydrationStatus
	}
	headers := make(http.Header, len(item.Headers))
	for name, value := range item.Headers {
		headers.Add(name, value)
	}
	return client.NewAPIError(client.Graph, item.Status, headers, item.Body)
}

func (s *Service) hydrateChunk(ctx context.Context, groups []group.Group, indices []int, options ListOptions, result *ListResult) error {
	pending := indices
	for attempt := 0; attempt < options.MaxThrottleAttempts; attempt++ {
		if err := ctx.Err(); err != nil {
			for _, index := range pending {
				result.Outcomes[index].Err = &ItemFailure{ID: groups[index].ID, Stage: "context", Cause: err}
			}
			return err
		}
		requests := make([]graphbatch.Request, 0, len(pending))
		correlation := make(map[string]int, len(pending))
		for _, index := range pending {
			id := strconv.Itoa(index + 1)
			requests = append(requests, graphbatch.Request{ID: id, Method: http.MethodGet, URL: "/teams/" + groups[index].ID})
			correlation[id] = index
			result.Outcomes[index].Attempts++
		}
		items, response, batchErr := graphbatch.Execute(ctx, s.scope, "v1.0", requests)
		if response != nil && response.StatusCode() >= 200 && response.StatusCode() < 300 && response.StatusCode() != http.StatusOK {
			items = nil
			batchErr = errors.Join(ErrHydrationStatus, batchErr)
		}
		trace := Result{}
		trace.record(HydrateBatch, response, batchErr)
		result.record(&trace)
		structural := batchErr != nil && (!errors.Is(batchErr, graphbatch.ErrItems) || errors.Is(batchErr, graphbatch.ErrResponse) || errors.Is(batchErr, ErrHydrationStatus))
		seen := make(map[int]bool, len(items))
		next := []int{}
		var delay time.Duration
		for _, item := range items {
			index, known := correlation[strings.ToLower(item.ID)]
			if !known || seen[index] {
				structural = true
				batchErr = errors.Join(batchErr, graphbatch.ErrResponse)
				continue
			}
			seen[index] = true
			outcome := &result.Outcomes[index]
			outcome.Response = response
			outcome.Subresponses = append(outcome.Subresponses, HydrationResponse{ID: item.ID, Status: item.Status, Headers: item.Headers, Body: item.Body})
			if item.Status == http.StatusTooManyRequests && !structural && attempt+1 < options.MaxThrottleAttempts {
				next = append(next, index)
				delay = max(delay, hydrationDelay(item.Headers, options.ThrottleDelay))
				continue
			}
			if item.Status != http.StatusOK {
				outcome.Err = &ItemFailure{ID: outcome.ID, Stage: "team-batch-status", Cause: hydrationError(item)}
				continue
			}
			var value team.Team
			if json.Unmarshal(item.Body, &value) != nil || !strings.EqualFold(value.ID, outcome.ID) {
				outcome.Err = &ItemFailure{ID: outcome.ID, Stage: "team-batch-decode", Cause: client.ErrDecode}
				continue
			}
			groupData := groups[index]
			outcome.Value = &ManagedTeam{ID: groupData.ID, Properties: groupData.Properties, Settings: value.Settings, InternalID: value.InternalID, IsArchived: value.IsArchived, GroupSnapshot: &groupData, TeamSnapshot: &value}
			var err error
			outcome.Filtered, err = filtered(outcome.Value, options, false)
			if err != nil {
				outcome.Err = &ItemFailure{ID: outcome.ID, Stage: "filter", Cause: err}
			}
		}
		for _, index := range pending {
			if !seen[index] {
				cause := batchErr
				if cause == nil {
					cause = graphbatch.ErrResponse
					batchErr = cause
					structural = true
				}
				result.Outcomes[index].Response = response
				result.Outcomes[index].Err = &ItemFailure{ID: groups[index].ID, Stage: "batch-response", Cause: cause}
			}
		}
		if structural {
			return batchErr
		}
		if len(next) == 0 {
			return nil
		}
		timer := time.NewTimer(delay)
		select {
		case <-ctx.Done():
			timer.Stop()
			for _, index := range next {
				result.Outcomes[index].Err = &ItemFailure{ID: groups[index].ID, Stage: "throttle-wait", Cause: ctx.Err()}
			}
			return ctx.Err()
		case <-timer.C:
		}
		pending = next
	}
	return nil
}

func (s *Service) hydrateGroups(ctx context.Context, groups []group.Group, options ListOptions, result *ListResult) error {
	result.Outcomes = make([]TeamOutcome, len(groups))
	for index, item := range groups {
		result.Outcomes[index].ID = item.ID
		if !directoryID.MatchString(item.ID) {
			result.Outcomes[index].Err = &ItemFailure{ID: item.ID, Stage: "identity", Cause: client.ErrDecode}
		}
	}
	for first := 0; first < len(groups); first += options.BatchSize {
		last := min(first+options.BatchSize, len(groups))
		indices := make([]int, 0, last-first)
		for index := first; index < last; index++ {
			if result.Outcomes[index].Err == nil {
				indices = append(indices, index)
			}
		}
		if len(indices) == 0 {
			continue
		}
		if err := s.hydrateChunk(ctx, groups, indices, options, result); err != nil {
			for index := last; index < len(groups); index++ {
				if result.Outcomes[index].Err == nil {
					result.Outcomes[index].Err = &ItemFailure{ID: groups[index].ID, Stage: "not-attempted", Cause: err}
				}
			}
			return err
		}
	}
	return nil
}
