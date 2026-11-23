// Package policyrp implements the compiled policy provider's direct ConfigAPI
// protocol. It never wraps the generated client's forest-discovery adapter.
package policyrp

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"slices"
	"strings"
	"unicode"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyfilter"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policytransport"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

type Resource[T any] struct {
	scope    *client.Scope
	name     string
	identity func(*T) *string
}

func New[T any](scope *client.Scope, name string, identity func(*T) *string) *Resource[T] {
	return &Resource[T]{scope: scope, name: name, identity: identity}
}

func (s *Resource[T]) Ready(ctx context.Context) error {
	if ctx == nil {
		return client.ErrNilContext
	}
	if err := ctx.Err(); err != nil {
		return err
	}
	if s == nil || s.scope == nil || s.scope.Session() == nil {
		return client.ErrClosed
	}
	if s.scope.Backend() != client.TeamsAdmin {
		return client.ErrDestination
	}
	if s.identity == nil {
		return client.ErrRequest
	}
	switch s.name {
	case "TeamsVoiceApplicationsPolicy", "TeamsVdiPolicy", "TeamsVirtualAppointmentsPolicy", "TeamsUpgradeConfiguration", "TeamsTenantAbuseConfiguration", "TeamsUpdateManagementPolicy", "TeamsTemplatePermissionPolicy", "TeamsSipDevicesConfiguration", "TeamsShiftsPolicy", "TeamsSharedCallingRoutingPolicy", "TeamsPersonalAttendantPolicy", "TeamsRecordingRollOutPolicy", "TeamsNotificationAndFeedsPolicy", "TeamsMobilityPolicy", "TeamsMultiTenantOrganizationConfiguration", "TeamsMessagingConfiguration", "TeamsMeetingTemplateConfiguration", "TeamsMeetingTemplatePermissionPolicy", "TeamsMeetingPolicy", "TeamsMessagingPolicy", "ExternalAccessPolicy", "TeamsAcsFederationConfiguration", "TeamsExternalAccessConfiguration", "TeamsMeetingBrandingPolicy", "OnlineVoicemailPolicy", "OnlineVoicemailValidationConfiguration", "PrivacyConfiguration", "TeamsRecordingAndTranscriptionCustomMessage", "TeamsAppPermissionPolicy", "TeamsAppSetupPolicy", "TeamsAudioConferencingCustomPromptsConfiguration", "TeamsCallHoldPolicy", "TeamsCallingPolicy", "TeamsChannelsPolicy", "TeamsClientConfiguration", "CustomBannerText", "TeamsEducationAssignmentsAppPolicy", "TeamsEducationConfiguration", "TeamsEventsPolicy", "TeamsFeedbackPolicy", "TeamsFilesPolicy", "TeamsComplianceRecordingPolicy", "TeamsComplianceRecordingApplication", "TeamsRemoteLogCollectionConfiguration", "TeamsRemoteLogCollectionDevice":
	case "TeamsMediaConnectivityPolicy":
	default:
		return client.ErrDestination
	}
	return nil
}

func ValidIdentity(identity string) bool {
	return strings.TrimSpace(identity) != "" && identity != "." && identity != ".." && !strings.ContainsRune(identity, '\\') && !strings.ContainsFunc(identity, unicode.IsControl)
}

func (s *Resource[T]) path(identity string) ([]string, error) {
	p := []string{"Skype.Policy", "configurations", s.name}
	if identity == "" {
		return p, nil
	}
	if !ValidIdentity(identity) {
		return nil, policy.ErrInput
	}
	p = append(p, "configuration")
	before, after, scoped := strings.Cut(identity, "/")
	if scoped {
		if before == "" || after == "" || before == "." || before == ".." || after == "." || after == ".." {
			return nil, policy.ErrInput
		}
		return append(p, before, after), nil
	}
	return append(p, identity), nil
}

func (s *Resource[T]) Do(ctx context.Context, command, method string, path []string, body any, replay client.ReplaySafety) (*resty.Response, error) {
	if err := s.Ready(ctx); err != nil {
		return nil, err
	}
	return policytransport.New(s.scope).Do(ctx, command, method, path, nil, body, replay)
}

func Decode[T any](response *resty.Response) (*T, error) {
	if response == nil {
		return nil, client.ErrDecode
	}
	if len(bytes.TrimSpace(response.Body())) == 0 {
		return nil, nil
	}
	var value *T
	if json.Unmarshal(response.Body(), &value) != nil {
		return nil, client.ErrDecode
	}
	return value, nil
}

func (s *Resource[T]) Get(ctx context.Context, identity, command string) (*T, *resty.Response, error) {
	if err := s.Ready(ctx); err != nil {
		return nil, nil, err
	}
	if !ValidIdentity(identity) {
		return nil, nil, policy.ErrInput
	}
	p, err := s.path(identity)
	if err != nil {
		return nil, nil, err
	}
	r, err := s.Do(ctx, command, http.MethodGet, p, nil, client.ReplaySafe)
	if err != nil {
		return nil, r, err
	}
	v, err := Decode[T](r)
	return v, r, err
}

func (s *Resource[T]) List(ctx context.Context, command string) ([]*T, *resty.Response, error) {
	if err := s.Ready(ctx); err != nil {
		return nil, nil, err
	}
	p, err := s.path("")
	if err != nil {
		return nil, nil, err
	}
	r, err := s.Do(ctx, command, http.MethodGet, p, nil, client.ReplaySafe)
	if err != nil {
		return nil, r, err
	}
	var values []*T
	if len(bytes.TrimSpace(r.Body())) > 0 && json.Unmarshal(r.Body(), &values) != nil {
		return nil, r, client.ErrDecode
	}
	global := func(v *T) bool {
		if v == nil || s.identity == nil {
			return false
		}
		id := s.identity(v)
		return id != nil && *id == "Global"
	}
	slices.SortStableFunc(values, func(a, b *T) int {
		left, right := global(a), global(b)
		if left == right {
			return 0
		}
		if left {
			return -1
		}
		return 1
	})
	return values, r, nil
}

func (s *Resource[T]) ListFiltered(ctx context.Context, pattern, command string) ([]*T, *resty.Response, error) {
	if err := s.Ready(ctx); err != nil {
		return nil, nil, err
	}
	if pattern == "" {
		return s.List(ctx, command)
	}
	plain, err := policyfilter.Compile(pattern)
	if err != nil {
		return nil, nil, policy.ErrInput
	}
	tagged, err := policyfilter.Compile("Tag:" + pattern)
	if err != nil {
		return nil, nil, policy.ErrInput
	}
	items, r, err := s.List(ctx, command)
	if err != nil {
		return nil, r, err
	}
	if items == nil {
		return nil, r, nil
	}
	values := make([]*T, 0, len(items))
	for _, item := range items {
		if item == nil {
			continue
		}
		id := s.identity(item)
		if id == nil {
			continue
		}
		match, err := plain.Match(ctx, *id)
		if err != nil {
			return nil, r, err
		}
		if !match {
			match, err = tagged.Match(ctx, *id)
		}
		if err != nil {
			return nil, r, err
		}
		if match {
			values = append(values, item)
		}
	}
	return values, r, nil
}

func Fields(value any) (map[string]json.RawMessage, error) {
	raw, err := json.Marshal(value)
	if err != nil {
		return nil, policy.ErrInput
	}
	var body map[string]json.RawMessage
	if json.Unmarshal(raw, &body) != nil || body == nil {
		return nil, policy.ErrInput
	}
	return body, nil
}

func Identity(body map[string]json.RawMessage, identity string, override policy.Field[string], controls policy.Controls) error {
	var value any = identity
	if override.IsSet() {
		if override.IsNull() {
			value = nil
		} else {
			value, _ = override.Value()
		}
	}
	raw, err := json.Marshal(value)
	if err != nil {
		return policy.ErrInput
	}
	body["Identity"] = raw
	if controls.Force != nil {
		force := "False"
		if *controls.Force {
			force = "True"
		}
		body["Force"], err = json.Marshal(force)
		if err != nil {
			return policy.ErrInput
		}
	}
	return nil
}

func (s *Resource[T]) Write(ctx context.Context, identity, command, method string, body map[string]json.RawMessage) (*T, *resty.Response, error) {
	if err := s.Ready(ctx); err != nil {
		return nil, nil, err
	}
	if !ValidIdentity(identity) {
		return nil, nil, policy.ErrInput
	}
	var p []string
	var err error
	if method == http.MethodPost {
		p, err = s.path("")
	} else {
		p, err = s.path(identity)
	}
	if err != nil {
		return nil, nil, err
	}
	r, err := s.Do(ctx, command, method, p, body, client.ReplayNever)
	if err != nil {
		return nil, r, err
	}
	v, err := Decode[T](r)
	return v, r, err
}
