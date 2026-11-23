package provisioningdiagnostics

import (
	"context"
	"crypto/rand"
	"encoding/json"
	"fmt"
	"net/http"
	"net/url"
	"regexp"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// The OCE commands (Get-CsUserProvHistory, Invoke-CsGenericNgtProvCommand,
// Invoke-CsDirectoryObjectSync) call Teams.InternalSupport on the ConfigAPI
// origin without forest discovery, identifying the tenant in X-MS-Tenant-Id
// and each command run in X-MS-Correlation-Id. The module only loads them in
// OCE sessions with a Torus user-impersonation token.

var serviceInstance = regexp.MustCompile(`^MicrosoftCommunicationsOnline/.+$`)

func oceHeaders(tenantID string) (http.Header, error) {
	tenant, ok := guid(tenantID)
	if !ok {
		return nil, ErrInput
	}
	var id [16]byte
	if _, err := rand.Read(id[:]); err != nil {
		return nil, client.ErrRequest
	}
	id[6], id[8] = id[6]&0x0f|0x40, id[8]&0x3f|0x80
	correlation := fmt.Sprintf("%x-%x-%x-%x-%x", id[0:4], id[4:6], id[6:8], id[8:10], id[10:16])
	return http.Header{"X-MS-Tenant-Id": {tenant}, "X-MS-Correlation-Id": {correlation}}, nil
}

// oce sends one attempt, as the module's direct sender does.
func (s *Service) oce(ctx context.Context, request client.Request) (*resty.Response, error) {
	response, err := s.base.Do(ctx, request, nil)
	if err != nil {
		return response, err
	}
	if !accepted(response, []int{http.StatusOK, http.StatusCreated, http.StatusAccepted, http.StatusNoContent}) {
		return response, ErrResponse
	}
	return response, nil
}

// GetUserProvisioningHistory reads a user's provisioning version and order
// history (Get-CsUserProvHistory). The module's -Mode only formats output.
func (s *Service) GetUserProvisioningHistory(ctx context.Context, tenantID, userID string, region DeploymentRegion) (*UserHistory, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	user, ok := guid(userID)
	if !ok || (region != RegionNOAM && region != RegionEMEA && region != RegionAPAC) {
		return nil, nil, ErrInput
	}
	h, err := oceHeaders(tenantID)
	if err != nil {
		return nil, nil, err
	}
	response, err := s.oce(ctx, client.Request{Method: http.MethodGet, Path: []string{internalRoute, "AdminStoreOperations", "userHistory", user, string(region)}, Header: h})
	if err != nil {
		return nil, response, err
	}
	value, err := decodeObject[UserHistory](response.Body())
	if err != nil {
		return nil, response, err
	}
	return value, response, nil
}

// InvokeNgtProvHandler runs a named provisioning handler against a target
// (Invoke-CsGenericNgtProvCommand). It is never replayed automatically.
//
// The module's route joins its parameters with "?" into one query value,
// handler=NAME?tenant=ID?target=TARGET?serviceInstance=INSTANCE; the SDK sends
// that same decoded handler value percent-encoded.
func (s *Service) InvokeNgtProvHandler(ctx context.Context, request NgtProvHandlerRequest) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	if _, ok := index(deployments, request.Region); !ok || !validText(request.HandlerName) || !validText(request.Target) || !validBody(request.Payload) ||
		(request.ServiceInstance != "" && !serviceInstance.MatchString(request.ServiceInstance)) {
		return nil, ErrInput
	}
	h, err := oceHeaders(request.TenantID)
	if err != nil {
		return nil, err
	}
	tenant, _ := guid(request.TenantID)
	handler := request.HandlerName + "?tenant=" + tenant + "?target=" + request.Target + "?serviceInstance=" + request.ServiceInstance
	return s.oce(ctx, client.Request{Method: http.MethodPost, Path: []string{internalRoute, "ProvisioningOperations", "runHandler", string(request.Region)},
		Query: url.Values{"handler": {handler}}, Body: request.Payload, Header: h})
}

// SyncDirectoryObject asks directory sync to resynchronize an object or the
// whole tenant (Invoke-CsDirectoryObjectSync). Set IsValidationRequest to
// validate without syncing. It is never replayed automatically.
func (s *Service) SyncDirectoryObject(ctx context.Context, request DirectoryObjectSyncRequest) (*resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, err
	}
	deployment, okDeployment := index(deployments, request.Deployment)
	class, okClass := index(objectClasses, request.ObjectClass)
	if !okDeployment || !okClass || !serviceInstance.MatchString(request.ServiceInstance) ||
		(request.SynchronizeTenantWithAllObjects && request.ObjectClass != ObjectClassCompany) {
		return nil, ErrInput
	}
	// Newtonsoft defaults: declared order, enums as numbers, nulls written.
	body := struct {
		ServiceInstance                 string   `json:"ServiceInstance"`
		DeploymentName                  int      `json:"DeploymentName"`
		ObjectID                        string   `json:"ObjectId"`
		ObjectIDs                       []string `json:"ObjectIds"`
		ObjectClass                     int      `json:"ObjectClass"`
		SynchronizeTenantWithAllObjects bool     `json:"SynchronizeTenantWithAllObjects"`
		IsValidationRequest             bool     `json:"IsValidationRequest"`
		ScenariosToSuppress             *string  `json:"ScenariosToSuppress"`
	}{ServiceInstance: request.ServiceInstance, DeploymentName: deployment, ObjectID: "00000000-0000-0000-0000-000000000000", ObjectClass: class,
		SynchronizeTenantWithAllObjects: request.SynchronizeTenantWithAllObjects, IsValidationRequest: request.IsValidationRequest, ScenariosToSuppress: request.ScenariosToSuppress}
	if len(request.ObjectIDs) == 0 {
		if request.ObjectID != "" {
			id, ok := guid(request.ObjectID)
			if !ok {
				return nil, ErrInput
			}
			body.ObjectID = id
		}
	} else {
		body.ObjectIDs = make([]string, len(request.ObjectIDs))
		for i, value := range request.ObjectIDs {
			id, ok := guid(value)
			if !ok {
				return nil, ErrInput
			}
			body.ObjectIDs[i] = id
		}
	}
	h, err := oceHeaders(request.TenantID)
	if err != nil {
		return nil, err
	}
	raw, err := json.Marshal(body)
	if err != nil {
		return nil, client.ErrRequest
	}
	return s.oce(ctx, client.Request{Method: http.MethodPost, Path: []string{internalRoute, "ProvisioningOperations", "DsRequest"}, Body: json.RawMessage(raw), Header: h})
}
