package policyassignments

import (
	"context"
	"errors"
	"net/http"
	"net/url"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

func selectedRegions(call *oceCall, region *Region) ([]Region, error) {
	if region == nil {
		return call.profile.regions, nil
	}
	selected, ok := normalizeRegion(*region)
	if !ok {
		return nil, ErrInput
	}
	return []Region{selected}, nil
}

func regionalLookup[T any](ctx context.Context, call *oceCall, id, kind string, region *Region, annotate func(*T, Region)) (*RegionalLookup[T], *resty.Response, error) {
	regions, err := selectedRegions(call, region)
	if err != nil {
		return nil, nil, err
	}
	result := &RegionalLookup[T]{}
	var last *resty.Response
	var failures []error
	for _, selected := range regions {
		if err := ctx.Err(); err != nil {
			return result, last, err
		}
		r, e := call.do(ctx, http.MethodGet, []string{"Skype.Policy", "assignments", "operations", kind, id}, url.Values{"region": {string(selected)}}, http.StatusOK)
		last = r
		attempt := RegionAttempt{Region: selected, Response: r, Err: e}
		if e != nil {
			result.Attempts = append(result.Attempts, attempt)
			failures = append(failures, e)
			continue
		}
		value, e := object[T](r)
		attempt.Err = e
		result.Attempts = append(result.Attempts, attempt)
		if e != nil {
			return result, r, e
		}
		annotate(value, selected)
		result.Value, result.Region = value, selected
		return result, r, nil
	}
	return result, last, errors.Join(failures...)
}

// GetOCEBatchStatus reads the separate regional LROS status protocol. A nil region
// searches the environment's ordered regions and stops at the first HTTP 200.
// Failed attempts remain available even if a later region succeeds. Decode errors
// on a successful response stop the search; they do not become a different job.
func (s *Service) GetOCEBatchStatus(ctx context.Context, input OCEContext, id string, region *Region) (*RegionalLookup[OCEBatchStatus], *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	identity, ok := canonicalGUID(id)
	if !ok {
		return nil, nil, ErrInput
	}
	call, err := s.prepareOCE(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	return regionalLookup(ctx, call, identity, "status", region, func(v *OCEBatchStatus, r Region) { v.Region = r })
}

func (s *Service) GetOCEBatchDefinition(ctx context.Context, input OCEContext, id string, region *Region) (*RegionalLookup[OCEBatchDefinition], *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	identity, ok := canonicalGUID(id)
	if !ok {
		return nil, nil, ErrInput
	}
	call, err := s.prepareOCE(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	return regionalLookup(ctx, call, identity, "definition", region, func(v *OCEBatchDefinition, r Region) { v.Region = r })
}

// ListOCEBatchStatuses gathers all selected regions in order. Successful results
// survive failures elsewhere; errors are returned both jointly and per attempt.
// Nil region selects the complete environment region list, not a single default.
func (s *Service) ListOCEBatchStatuses(ctx context.Context, input OCEContext, region *Region) (*RegionalList[OCEBatchStatus], *resty.Response, error) {
	call, err := s.prepareOCE(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	regions, err := selectedRegions(call, region)
	if err != nil {
		return nil, nil, err
	}
	result := &RegionalList[OCEBatchStatus]{Values: []*OCEBatchStatus{}}
	var last *resty.Response
	var failures []error
	for _, selected := range regions {
		if err := ctx.Err(); err != nil {
			return result, last, err
		}
		r, e := call.do(ctx, http.MethodGet, []string{"Skype.Policy", "assignments", "operations", "status"}, url.Values{"region": {string(selected)}}, http.StatusOK)
		last = r
		attempt := RegionAttempt{Region: selected, Response: r, Err: e}
		if e != nil {
			failures = append(failures, e)
			result.Attempts = append(result.Attempts, attempt)
			continue
		}
		values, e := array[OCEBatchStatus](r)
		if e == nil {
			for _, value := range values {
				if value == nil {
					e = client.ErrDecode
					break
				}
				value.Region = selected
			}
		}
		attempt.Err = e
		result.Attempts = append(result.Attempts, attempt)
		if e != nil {
			return result, r, e
		}
		result.Values = append(result.Values, values...)
	}
	return result, last, errors.Join(failures...)
}

// ReprocessOCEBatch submits once to an explicit region. HTTP 202 and the returned
// Guid do not imply completed processing, user success or effective convergence.
func (s *Service) ReprocessOCEBatch(ctx context.Context, input OCEContext, id string, region Region) (*OCEBatchReprocessResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	identity, validID := canonicalGUID(id)
	selected, validRegion := normalizeRegion(region)
	if !validID || !validRegion {
		return nil, nil, ErrInput
	}
	call, err := s.prepareOCE(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	r, err := call.do(ctx, http.MethodPost, []string{"Skype.Policy", "assignments", "operations", "reprocess", identity}, url.Values{"region": {string(selected)}}, http.StatusAccepted)
	if err != nil {
		return nil, r, err
	}
	result := &OCEBatchReprocessResult{HTTPAccepted: true}
	// The pinned command trims surrounding quotes before Guid construction,
	// accepting a bare canonical Guid as well as the usual JSON string reply.
	body := strings.Trim(string(r.Body()), `"`)
	value, ok := canonicalGUID(strings.TrimSpace(body))
	if !ok {
		return result, r, client.ErrDecode
	}
	result.OperationID = &value
	return result, r, nil
}
