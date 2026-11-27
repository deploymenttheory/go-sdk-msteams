package customappsettings

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"sort"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/go-resty/resty/v2"
)

// WriteProjection exposes the pinned serializer's body and its changes to the
// read snapshot. These changes describe serialization, not server persistence.
type WriteProjection struct {
	Body         json.RawMessage
	OmittedPaths []string
	ChangedPaths []string
}

func (WriteProjection) String() string     { return "custom app settings projection [redacted]" }
func (v WriteProjection) GoString() string { return v.String() }
func (v WriteProjection) clone() WriteProjection {
	return WriteProjection{Body: append(json.RawMessage(nil), v.Body...), OmittedPaths: append([]string(nil), v.OmittedPaths...), ChangedPaths: append([]string(nil), v.ChangedPaths...)}
}

// ProjectSideloadedAppsInteraction reproduces the pinned Set wrapper's local
// read-model and request-model projection. It performs no authentication or I/O.
// Unknown fields and malformed nullable scalars are omitted, with visible paths.
func ProjectSideloadedAppsInteraction(raw json.RawMessage, enabled bool) (WriteProjection, error) {
	var before map[string]json.RawMessage
	if json.Unmarshal(raw, &before) != nil || before == nil {
		return WriteProjection{}, client.ErrDecode
	}
	body := map[string]json.RawMessage{"appSettingsList": json.RawMessage("[]")}
	flag, _ := json.Marshal(enabled)
	body["isSideloadedAppsInteractionEnabled"] = flag
	for _, key := range []string{"isAppsEnabled", "isAppsPurchaseEnabled", "isExternalAppsEnabledByDefault", "isLicenseBasedPinnedAppsEnabled", "isTenantWideAutoInstallEnabled"} {
		var value *bool
		if json.Unmarshal(before[key], &value) == nil && value != nil {
			body[key], _ = json.Marshal(*value)
		}
	}
	for _, key := range []string{"LobTextColor", "LobBackground", "LobLogo", "LobLogomark"} {
		var value *string
		if json.Unmarshal(before[key], &value) == nil && value != nil {
			body[key], _ = json.Marshal(*value)
		}
	}
	var nested map[string]json.RawMessage
	_ = json.Unmarshal(before["appAccessRequestConfig"], &nested)
	projected := make(map[string]json.RawMessage)
	for _, key := range []string{"adminInstructionMessage", "approvalPortalUrl"} {
		var value *string
		if json.Unmarshal(nested[key], &value) == nil && value != nil {
			projected[key], _ = json.Marshal(*value)
		}
	}
	if len(projected) != 0 {
		body["appAccessRequestConfig"], _ = json.Marshal(projected)
	}
	result := WriteProjection{}
	for key, old := range before {
		next, exists := body[key]
		if !exists {
			result.OmittedPaths = append(result.OmittedPaths, key)
		} else if !sameJSON(old, next) {
			result.ChangedPaths = append(result.ChangedPaths, key)
		}
	}
	for key := range body {
		if _, exists := before[key]; !exists {
			result.ChangedPaths = append(result.ChangedPaths, key)
		}
	}
	for key := range nested {
		if _, exists := projected[key]; !exists {
			result.OmittedPaths = append(result.OmittedPaths, "appAccessRequestConfig."+key)
		}
	}
	sort.Strings(result.OmittedPaths)
	sort.Strings(result.ChangedPaths)
	result.Body, _ = json.Marshal(body)
	return result, nil
}

func sameJSON(a, b []byte) bool {
	var left, right bytes.Buffer
	return json.Compact(&left, a) == nil && json.Compact(&right, b) == nil && bytes.Equal(left.Bytes(), right.Bytes())
}

// SideloadedUpdate freezes the source projection and binds it to the preparing
// service. Inspect Projection before applying; inspection returns owned copies.
type SideloadedUpdate struct {
	owner        *Service
	before       json.RawMessage
	projection   WriteProjection
	readResponse *resty.Response
}

func (SideloadedUpdate) String() string                  { return "prepared custom app settings update [redacted]" }
func (v SideloadedUpdate) GoString() string              { return v.String() }
func (v SideloadedUpdate) Before() json.RawMessage       { return append(json.RawMessage(nil), v.before...) }
func (v SideloadedUpdate) Projection() WriteProjection   { return v.projection.clone() }
func (v SideloadedUpdate) ReadResponse() *resty.Response { return v.readResponse }

// PrepareSideloadedAppsInteraction reads with the private child Get's source
// metadata, then freezes the exact source-compatible Set projection. It sends no
// PUT. The existing enhanced convenience method remains a separate extension.
func (s *Service) PrepareSideloadedAppsInteraction(ctx context.Context, enabled bool) (*SideloadedUpdate, *resty.Response, error) {
	if err := s.check(ctx); err != nil {
		return nil, nil, err
	}
	response, err := s.read(ctx, "Get-CsTeamsSettingsCustomApp_Get")
	if err != nil {
		return nil, response, err
	}
	projection, err := ProjectSideloadedAppsInteraction(response.Body(), enabled)
	if err != nil {
		return nil, response, err
	}
	return &SideloadedUpdate{owner: s, before: append(json.RawMessage(nil), response.Body()...), projection: projection, readResponse: response}, response, nil
}

// SideloadedUpdateResult retains independent read and write receipts. HTTPAccepted
// means a 2xx receipt, independently of the source's required 204 outcome. A
// transport error can leave an ambiguous write: never replay it blindly.
type SideloadedUpdateResult struct {
	ReadResponse   *resty.Response
	UpdateResponse *resty.Response
	UpdateInvoked  bool
	HTTPAccepted   bool
	projection     WriteProjection
}

func (SideloadedUpdateResult) String() string                { return "custom app settings update result [redacted]" }
func (v SideloadedUpdateResult) GoString() string            { return v.String() }
func (v SideloadedUpdateResult) Projection() WriteProjection { return v.projection.clone() }

// UpdateSideloadedAppsInteraction sends one PUT with a frozen source projection. The
// request must belong to this exact service; it cannot be retargeted to another
// tenant/client. It never retries or rereads the snapshot automatically.
func (s *Service) UpdateSideloadedAppsInteraction(ctx context.Context, prepared *SideloadedUpdate) (*SideloadedUpdateResult, *resty.Response, error) {
	if prepared == nil || prepared.owner != s || s == nil {
		return nil, nil, ErrInput
	}
	result := &SideloadedUpdateResult{ReadResponse: prepared.readResponse, projection: prepared.projection.clone()}
	if err := s.check(ctx); err != nil {
		return result, nil, err
	}
	req := request(http.MethodPut, setCommand)
	req.Replay = client.ReplayNever
	req.Body = append(json.RawMessage(nil), prepared.projection.Body...)
	result.UpdateInvoked = true
	response, err := s.client.Do(ctx, req, nil)
	result.UpdateResponse = response
	result.HTTPAccepted = response != nil && response.StatusCode() >= 200 && response.StatusCode() < 300
	if err != nil {
		return result, response, err
	}
	if response == nil || response.StatusCode() != http.StatusNoContent {
		return result, response, ErrResponse
	}
	return result, response, nil
}
