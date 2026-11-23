package policyassignments

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/url"
	"strconv"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// NgtProvFailoverResult is the provisioning failover service's reply. The
// module decodes the body case-insensitively and prints only ResponseMessage;
// RawFields keeps every property.
type NgtProvFailoverResult struct {
	ResponseMessage *string                    `json:"ResponseMessage"`
	RawFields       map[string]json.RawMessage `json:"-"`
}

func (v *NgtProvFailoverResult) UnmarshalJSON(data []byte) error {
	type plain NgtProvFailoverResult
	var value plain
	if err := snapshot(data, &value, &value.RawFields); err != nil {
		return err
	}
	*v = NgtProvFailoverResult(value)
	return nil
}

// ngtProvRegion accepts only the regions the module's NgtProv DeploymentRegion
// enumeration declares.
func ngtProvRegion(value Region) (Region, bool) {
	selected, ok := normalizeRegion(value)
	if !ok || (selected != RegionNOAM && selected != RegionEMEA && selected != RegionAPAC) {
		return "", false
	}
	return selected, true
}

func ngtProvResult(r *resty.Response) (*NgtProvFailoverResult, error) {
	if r == nil {
		return nil, client.ErrDecode
	}
	if len(bytes.TrimSpace(r.Body())) == 0 {
		return &NgtProvFailoverResult{}, nil
	}
	var value *NgtProvFailoverResult
	if json.Unmarshal(r.Body(), &value) != nil || value == nil {
		return nil, client.ErrDecode
	}
	return value, nil
}

// GetOCENgtProvFailoverStatus reads the provisioning (NgtProv) instance
// failover state of one region (Get-NgtProvInstanceFailOverStatus). region is
// NOAM, EMEA or APAC. This is a Microsoft-internal OCE read.
func (s *Service) GetOCENgtProvFailoverStatus(ctx context.Context, input OCEContext, region Region) (*NgtProvFailoverResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	selected, ok := ngtProvRegion(region)
	if !ok {
		return nil, nil, ErrInput
	}
	call, err := s.prepareOCE(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	r, err := call.do(ctx, http.MethodGet, []string{"Teams.InternalSupport", "NgtProvOperations", "FailoverStatus"}, url.Values{"region": {string(selected)}}, http.StatusOK)
	if err != nil {
		return nil, r, err
	}
	v, err := ngtProvResult(r)
	return v, r, err
}

// MoveOCENgtProvInstance shifts a region's provisioning traffic between its
// two sides (Move-NgtProvInstance): sideAPercent and sideBPercent are the
// traffic percentages sent to each side. This is an irreversible,
// service-wide Microsoft-internal failover action, not a tenant setting. It
// sends exactly one POST with no body and is never replayed. The module
// passes the percentages unvalidated; so does the SDK.
func (s *Service) MoveOCENgtProvInstance(ctx context.Context, input OCEContext, region Region, sideAPercent, sideBPercent int32) (*NgtProvFailoverResult, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	selected, ok := ngtProvRegion(region)
	if !ok {
		return nil, nil, ErrInput
	}
	call, err := s.prepareOCE(ctx, input)
	if err != nil {
		return nil, nil, err
	}
	path := []string{"Teams.InternalSupport", "NgtProvOperations", "Failover", strconv.FormatInt(int64(sideAPercent), 10), strconv.FormatInt(int64(sideBPercent), 10)}
	r, err := call.do(ctx, http.MethodPost, path, url.Values{"region": {string(selected)}}, http.StatusOK)
	if err != nil {
		return nil, r, err
	}
	v, err := ngtProvResult(r)
	return v, r, err
}
