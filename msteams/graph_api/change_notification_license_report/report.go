package licensereport

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strconv"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	graphbatch "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/batch"
)

// GetOptions bounds each collection and the number of distinct license subjects.
// MaxUsers defaults to 10,000; limits are client budgets, not schema restrictions.
type GetOptions struct {
	Activity, Users, Licenses ListOptions
	MaxUsers                  int
}

// Get reads a native report with complete activity/user traversal before resolving
// license subjects. License reads use documented directory IDs, not source UPN
// concatenation. Its delegated-only license mode is explicit and never replaced
// by an automatic identity fallback. Missing or denied data yields partial error.
func (s *Service) Get(ctx context.Context, period Period, options GetOptions) (*ReportResult, error) {
	result := &ReportResult{}
	if e := s.licenseMode(ctx); e != nil {
		return result, e
	}
	result.DocumentedLicenseMode = true
	if !validPeriod(period) {
		return result, ErrInput
	}
	if options.MaxUsers == 0 {
		options.MaxUsers = 10000
	}
	if options.MaxUsers < 1 {
		return result, ErrInput
	}
	activity, e := s.ListActivity(ctx, period, options.Activity)
	if activity != nil {
		result.Activity = *activity
	}
	if e != nil {
		return result, errors.Join(ErrPartial, e)
	}
	result.ActivityComplete = true
	if len(result.Activity.Items) == 0 {
		result.Rows = make([]Row, 0)
		result.Complete = true
		return result, nil
	}
	result.Rows = make([]Row, len(result.Activity.Items))
	for i, a := range result.Activity.Items {
		result.Rows[i].Activity = a
		result.Rows[i].MessageCount, result.Rows[i].Err = total(a)
	}
	users, e := s.ListUsers(ctx, options.Users)
	if users != nil {
		result.Users = *users
	}
	if e != nil {
		return result, errors.Join(ErrPartial, e)
	}
	result.UsersComplete = true
	byID := make(map[string][]*DirectoryUser)
	byUPN := make(map[string][]*DirectoryUser)
	for i := range result.Users.Items {
		u := &result.Users.Items[i]
		key := u.ID
		byID[key] = append(byID[key], u)
		if u.UserPrincipalName != nil {
			byUPN[*u.UserPrincipalName] = append(byUPN[*u.UserPrincipalName], u)
		}
	}
	unique := make(map[string]int)
	var ids []string
	var outcome error
	for i := range result.Rows {
		row := &result.Rows[i]
		user, e := resolveUser(row.Activity, byID, byUPN)
		if e != nil {
			row.Err = errors.Join(row.Err, e)
			outcome = errors.Join(outcome, e)
			continue
		}
		row.DirectoryUser = user
		row.DirectoryUserID = user.ID
		key := user.ID
		if _, found := unique[key]; !found {
			unique[key] = len(ids)
			ids = append(ids, user.ID)
		}
	}
	if len(ids) > options.MaxUsers {
		return result, errors.Join(ErrPartial, client.ErrItemLimit)
	}
	result.Licenses, e = s.readLicenseBatches(ctx, ids, options.Licenses)
	outcome = errors.Join(outcome, e)
	result.LicensesComplete = len(ids) > 0
	for _, license := range result.Licenses {
		if !license.Complete {
			result.LicensesComplete = false
		}
	}
	for i := range result.Rows {
		row := &result.Rows[i]
		if row.DirectoryUserID == "" {
			continue
		}
		index := unique[row.DirectoryUserID]
		if index >= len(result.Licenses) {
			row.Err = errors.Join(row.Err, ErrPartial)
			outcome = errors.Join(outcome, ErrPartial)
			continue
		}
		license := &result.Licenses[index]
		row.Err = errors.Join(row.Err, license.Err)
		row.HasMatchingSourcePlan = license.HasMatchingSourcePlan
		if row.Err != nil {
			outcome = errors.Join(outcome, row.Err)
		}
	}
	result.Complete = outcome == nil
	if outcome != nil {
		return result, errors.Join(ErrPartial, outcome)
	}
	return result, nil
}

func resolveUser(activity Activity, byID, byUPN map[string][]*DirectoryUser) (*DirectoryUser, error) {
	var values []*DirectoryUser
	if activity.UserID != nil && *activity.UserID != "" {
		// A declared directory ID is authoritative across renamed/reused UPNs.
		// Do not fall back to a different UPN owner when that ID cannot be resolved.
		values = byID[*activity.UserID]
	} else if activity.UserPrincipalName != nil {
		values = byUPN[*activity.UserPrincipalName]
	}
	if len(values) != 1 {
		return nil, ErrIdentity
	}
	return values[0], nil
}

func (s *Service) readLicenseBatches(ctx context.Context, ids []string, options ListOptions) ([]LicenseOutcome, error) {
	results := make([]LicenseOutcome, len(ids))
	for i, id := range ids {
		results[i].DirectoryUserID = id
	}
	var combined error
	for start := 0; start < len(ids); start += 20 {
		if e := ctx.Err(); e != nil {
			for i := start; i < len(ids); i++ {
				results[i].Err = e
			}
			return results, errors.Join(combined, e)
		}
		end := min(start+20, len(ids))
		requests := make([]graphbatch.Request, 0, end-start)
		for i := start; i < end; i++ {
			requests = append(requests, graphbatch.Request{ID: strconv.Itoa(i - start), Method: http.MethodGet, URL: "/users/" + url.PathEscape(ids[i]) + "/licenseDetails"})
		}
		replies, response, e := graphbatch.Execute(ctx, s.scope, "v1.0", requests)
		if responseStatus(response) != http.StatusOK && e == nil {
			e = client.ErrPageStatus
		}
		seen := make(map[int]bool)
		for _, reply := range replies {
			offset, parseErr := strconv.Atoi(reply.ID)
			if parseErr != nil || offset < 0 || offset >= end-start {
				combined = errors.Join(combined, graphbatch.ErrResponse)
				continue
			}
			index := start + offset
			seen[index] = true
			result := &results[index]
			result.Response = response
			if reply.Status != http.StatusOK {
				headers := make(http.Header)
				for key, value := range reply.Headers {
					headers.Set(key, value)
				}
				result.Err = client.NewAPIError(client.Graph, reply.Status, headers, reply.Body)
				combined = errors.Join(combined, result.Err)
				continue
			}
			result.Items, result.Pages, result.Err = s.licenseBody(ctx, ids[index], reply.Body, options)
			result.Complete = result.Err == nil
			if result.Complete {
				result.HasMatchingSourcePlan, result.PlanDataComplete, result.Err = matchingPlan(result.Items)
			}
			combined = errors.Join(combined, result.Err)
		}
		for i := start; i < end; i++ {
			if !seen[i] {
				results[i].Response = response
				results[i].Err = errors.Join(e, graphbatch.ErrResponse)
				combined = errors.Join(combined, results[i].Err)
			}
		}
		// Preserve valid individual outcomes on mixed failures. A whole-envelope
		// failure stops subsequent batches; neither denied items nor batches replay.
		combined = errors.Join(combined, e)
		if len(replies) == 0 && e != nil {
			for i := end; i < len(ids); i++ {
				results[i].Err = e
			}
			return results, combined
		}
	}
	return results, combined
}

func (s *Service) licenseBody(ctx context.Context, id string, body []byte, options ListOptions) ([]LicenseDetail, int, error) {
	page, e := client.DecodeCollection[LicenseDetail](body, "value", "@odata.nextLink")
	if e != nil {
		return nil, 0, e
	}
	maxPages, maxItems := options.MaxPages, options.MaxItems
	if maxPages == 0 {
		maxPages = 1000
	}
	if maxItems == 0 {
		maxItems = 100000
	}
	if maxPages < 1 || maxItems < 1 {
		return nil, 0, client.ErrPagination
	}
	if len(page.Items) > maxItems {
		return nil, 1, client.ErrItemLimit
	}
	items := append([]LicenseDetail{}, page.Items...)
	if page.Next == "" {
		return items, 1, nil
	}
	guard, guardErr := s.licenseContinuation(id)
	if guardErr != nil {
		return items, 1, guardErr
	}
	if guardErr = guard(page.Next); guardErr != nil {
		return items, 1, guardErr
	}
	nextURL, parseErr := url.Parse(page.Next)
	if parseErr == nil && nextURL.RawQuery == "" {
		return items, 1, client.ErrContinuationLoop
	}
	if maxPages == 1 {
		return items, 1, client.ErrPageLimit
	}
	if len(items) == maxItems {
		return items, 1, client.ErrItemLimit
	}
	tail, e := s.continueLicenses(ctx, id, page.Next, ListOptions{MaxPages: maxPages - 1, MaxItems: maxItems - len(items)})
	pages := 1
	if tail != nil {
		items = append(items, tail.Items...)
		pages += tail.Pages
	}
	return items, pages, e
}

// Matching is explicitly the pinned source's ordinal string rule. Unknown plan
// data can establish neither a negative match nor entitlement/authorization.
func matchingPlan(items []LicenseDetail) (*bool, bool, error) {
	known := true
	found := false
	for _, license := range items {
		if license.ServicePlans == nil {
			known = false
			continue
		}
		for _, plan := range license.ServicePlans {
			if plan == nil || plan.ID == nil || !planIDPattern.MatchString(*plan.ID) {
				known = false
				continue
			}
			if *plan.ID == SourcePlanID {
				found = true
			}
		}
	}
	if found {
		if !known {
			return &found, false, client.ErrDecode
		}
		return &found, true, nil
	}
	if !known {
		return nil, false, client.ErrDecode
	}
	return &found, true, nil
}
