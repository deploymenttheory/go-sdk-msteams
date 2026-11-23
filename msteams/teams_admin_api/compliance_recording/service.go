package compliancerecording

import (
	"context"
	"encoding/json"
	"net/http"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/configapi"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/internal/policyrp"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	"github.com/go-resty/resty/v2"
)

const (
	policyResource      = "TeamsComplianceRecordingPolicy"
	applicationResource = "TeamsComplianceRecordingApplication"
	grantCommand        = "Grant-CsTeamsComplianceRecordingPolicy"
)

// Service manages compliance recording policies, their applications and policy
// grants. Policies and applications use the compiled policy transport; grants
// use the modern ConfigAPI transport. Construction performs no I/O.
type Service struct {
	policies     *policyrp.Resource[Policy]
	applications *policyrp.Resource[Application]
	modern       client.Client
}

// NewService binds both transports to the Teams admin scope.
func NewService(scope *client.Scope) *Service {
	return &Service{
		policies:     policyrp.New(scope, policyResource, func(v *Policy) *string { return v.Identity }),
		applications: policyrp.New(scope, applicationResource, func(v *Application) *string { return v.Identity }),
		modern:       configapi.New(scope),
	}
}

func (s *Service) ready(ctx context.Context) error {
	if s == nil {
		if ctx == nil {
			return client.ErrNilContext
		}
		if err := ctx.Err(); err != nil {
			return err
		}
		return client.ErrClosed
	}
	return s.policies.Ready(ctx)
}

func put(body map[string]json.RawMessage, key string, value any) error {
	raw, err := json.Marshal(value)
	if err != nil {
		return ErrInput
	}
	body[key] = raw
	return nil
}

func validApplications(values []*Application) bool {
	for _, value := range values {
		if value == nil || (value.ID != nil && !validID(*value.ID)) || !validPaired(value.ComplianceRecordingPairedApplications) {
			return false
		}
	}
	return true
}

func validPaired(values []*PairedApplication) bool {
	for _, value := range values {
		if value == nil || (value.ID != nil && !validID(*value.ID)) {
			return false
		}
	}
	return true
}

func policyFields(settings PolicySettings) (map[string]json.RawMessage, error) {
	if apps, ok := settings.ComplianceRecordingApplications.Value(); ok && !validApplications(apps) {
		return nil, ErrInput
	}
	body, err := policyrp.Fields(settings)
	if err != nil {
		return nil, err
	}
	// A bound nil list is an empty list, as the module's list binder sends.
	if apps, ok := settings.ComplianceRecordingApplications.Value(); ok && apps == nil {
		body["ComplianceRecordingApplications"] = json.RawMessage(`[]`)
	}
	return body, nil
}

func applicationFields(settings ApplicationSettings) (map[string]json.RawMessage, error) {
	if paired, ok := settings.ComplianceRecordingPairedApplications.Value(); ok && !validPaired(paired) {
		return nil, ErrInput
	}
	body, err := policyrp.Fields(settings)
	if err != nil {
		return nil, err
	}
	if paired, ok := settings.ComplianceRecordingPairedApplications.Value(); ok && paired == nil {
		body["ComplianceRecordingPairedApplications"] = json.RawMessage(`[]`)
	}
	return body, nil
}

// GetPolicyByIdentity reads one policy, for example "Global" or "Tag:name".
func (s *Service) GetPolicyByIdentity(ctx context.Context, identity string) (*Policy, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.policies.Get(ctx, identity, "Get-CsTeamsComplianceRecordingPolicy")
}

// ListPolicies reads every policy, Global first.
func (s *Service) ListPolicies(ctx context.Context) ([]*Policy, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.policies.List(ctx, "Get-CsTeamsComplianceRecordingPolicy")
}

// ListPoliciesFiltered applies the module's local Identity/Tag wildcard filter.
func (s *Service) ListPoliciesFiltered(ctx context.Context, pattern string) ([]*Policy, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if pattern == "" {
		return nil, nil, ErrInput
	}
	return s.policies.ListFiltered(ctx, pattern, "Get-CsTeamsComplianceRecordingPolicy")
}

// CreatePolicyByIdentity is New-CsTeamsComplianceRecordingPolicy.
func (s *Service) CreatePolicyByIdentity(ctx context.Context, identity string, settings PolicySettings, controls policy.Controls) (*Policy, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	body, err := policyFields(settings)
	if err != nil {
		return nil, nil, err
	}
	if err := policyrp.Identity(body, identity, policy.Field[string]{}, controls); err != nil {
		return nil, nil, err
	}
	return s.policies.Write(ctx, identity, "New-CsTeamsComplianceRecordingPolicy", http.MethodPost, body)
}

// PatchPolicyByIdentity sends one sparse PATCH with the bound settings. At
// least one setting is required; Force alone is not a change.
func (s *Service) PatchPolicyByIdentity(ctx context.Context, identity string, settings PolicySettings, controls policy.Controls) (*Policy, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	body, err := policyFields(settings)
	if err != nil {
		return nil, nil, err
	}
	if len(body) == 0 {
		return nil, nil, ErrInput
	}
	if err := policyrp.Identity(body, identity, policy.Field[string]{}, controls); err != nil {
		return nil, nil, err
	}
	return s.policies.Write(ctx, identity, "Set-CsTeamsComplianceRecordingPolicy", http.MethodPatch, body)
}

// UpdatePolicyByIdentity is Set-CsTeamsComplianceRecordingPolicy: it reads the
// policy, applies any application list edit to the read value and sends one
// sparse PATCH identified by the read Identity.
func (s *Service) UpdatePolicyByIdentity(ctx context.Context, identity string, input PolicyUpdate) (*UpdateResult[Policy], *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	settings := input.PolicySettings
	if input.Applications != nil && settings.ComplianceRecordingApplications.IsSet() {
		return nil, nil, ErrInput
	}
	if probe, err := policyFields(settings); err != nil || (len(probe) == 0 && input.Applications == nil) {
		return nil, nil, ErrInput
	}
	before, read, err := s.policies.Get(ctx, identity, "Set-CsTeamsComplianceRecordingPolicy")
	result := &UpdateResult[Policy]{Before: before, ReadResponse: read}
	if err != nil {
		return result, read, err
	}
	if before == nil || before.Identity == nil {
		return result, read, client.ErrDecode
	}
	if input.Applications != nil {
		settings.ComplianceRecordingApplications = policy.Set(applyEdit(before.ComplianceRecordingApplications, *input.Applications, func(v *Application) *string { return v.ID }))
	}
	result.PatchInvoked = true
	result.After, read, err = s.PatchPolicyByIdentity(ctx, *before.Identity, settings, input.Controls)
	result.HTTPAccepted = read != nil && read.StatusCode() >= 200 && read.StatusCode() < 300
	return result, read, err
}

// DeletePolicyByIdentity is Remove-CsTeamsComplianceRecordingPolicy.
func (s *Service) DeletePolicyByIdentity(ctx context.Context, identity string, controls policy.Controls) (*Policy, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	body := map[string]json.RawMessage{}
	if err := policyrp.Identity(body, identity, policy.Field[string]{}, controls); err != nil {
		return nil, nil, err
	}
	return s.policies.Write(ctx, identity, "Remove-CsTeamsComplianceRecordingPolicy", http.MethodDelete, body)
}

// applyEdit applies a PSListModifier-style edit to a copy of current.
func applyEdit[T any](current []*T, edit ListEdit[T], id func(*T) *string) []*T {
	out := append([]*T{}, current...)
	if edit.Replace != nil {
		out = append([]*T{}, edit.Replace...)
	}
	out = append(out, edit.Add...)
	if len(edit.Remove) == 0 {
		return out
	}
	kept := out[:0]
	for _, item := range out {
		drop := false
		if item != nil {
			if value := id(item); value != nil {
				for _, remove := range edit.Remove {
					if strings.EqualFold(*value, remove) {
						drop = true
					}
				}
			}
		}
		if !drop {
			kept = append(kept, item)
		}
	}
	return kept
}

// applicationIdentity builds the ParentAndRelativeKey identity "{parent}/{id}".
func applicationIdentity(parent, id string) (string, error) {
	if !validID(id) || strings.Contains(parent, "/") || !policyrp.ValidIdentity(parent) {
		return "", ErrInput
	}
	return parent + "/" + id, nil
}

// GetApplicationByIdentity reads one application by "{policy}/{id}".
func (s *Service) GetApplicationByIdentity(ctx context.Context, identity string) (*Application, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.applications.Get(ctx, identity, "Get-CsTeamsComplianceRecordingApplication")
}

// GetApplicationByID reads the application id under the parent policy.
func (s *Service) GetApplicationByID(ctx context.Context, parent, id string) (*Application, *resty.Response, error) {
	identity, err := applicationIdentity(parent, id)
	if err != nil {
		return nil, nil, err
	}
	return s.GetApplicationByIdentity(ctx, identity)
}

// ListApplications reads every application across policies.
func (s *Service) ListApplications(ctx context.Context) ([]*Application, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	return s.applications.List(ctx, "Get-CsTeamsComplianceRecordingApplication")
}

// ListApplicationsFiltered applies the module's local Identity/Tag filter.
func (s *Service) ListApplicationsFiltered(ctx context.Context, pattern string) ([]*Application, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	if pattern == "" {
		return nil, nil, ErrInput
	}
	return s.applications.ListFiltered(ctx, pattern, "Get-CsTeamsComplianceRecordingApplication")
}

// CreateApplicationByIdentity is New-CsTeamsComplianceRecordingApplication
// -Identity "{policy}/{id}". The body carries Identity and the extracted Id.
// The service requires id to be an existing resource account object ID.
func (s *Service) CreateApplicationByIdentity(ctx context.Context, identity string, settings ApplicationSettings, controls policy.Controls) (*Application, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	parent, id, ok := strings.Cut(identity, "/")
	if !ok {
		return nil, nil, ErrInput
	}
	if _, err := applicationIdentity(parent, id); err != nil {
		return nil, nil, err
	}
	body, err := applicationFields(settings)
	if err != nil {
		return nil, nil, err
	}
	if err := put(body, "Id", id); err != nil {
		return nil, nil, err
	}
	if err := policyrp.Identity(body, identity, policy.Field[string]{}, controls); err != nil {
		return nil, nil, err
	}
	return s.applications.Write(ctx, identity, "New-CsTeamsComplianceRecordingApplication", http.MethodPost, body)
}

// CreateApplicationByID is New-CsTeamsComplianceRecordingApplication -Parent -Id.
func (s *Service) CreateApplicationByID(ctx context.Context, parent, id string, settings ApplicationSettings, controls policy.Controls) (*Application, *resty.Response, error) {
	identity, err := applicationIdentity(parent, id)
	if err != nil {
		return nil, nil, err
	}
	return s.CreateApplicationByIdentity(ctx, identity, settings, controls)
}

// PatchApplicationByIdentity sends one sparse application PATCH.
func (s *Service) PatchApplicationByIdentity(ctx context.Context, identity string, settings ApplicationSettings, controls policy.Controls) (*Application, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	body, err := applicationFields(settings)
	if err != nil {
		return nil, nil, err
	}
	if len(body) == 0 {
		return nil, nil, ErrInput
	}
	if err := policyrp.Identity(body, identity, policy.Field[string]{}, controls); err != nil {
		return nil, nil, err
	}
	return s.applications.Write(ctx, identity, "Set-CsTeamsComplianceRecordingApplication", http.MethodPatch, body)
}

// UpdateApplicationByIdentity is Set-CsTeamsComplianceRecordingApplication: a
// read, a paired-list edit applied to the read value, and one sparse PATCH.
func (s *Service) UpdateApplicationByIdentity(ctx context.Context, identity string, input ApplicationUpdate) (*UpdateResult[Application], *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	settings := input.ApplicationSettings
	if input.PairedApplications != nil && settings.ComplianceRecordingPairedApplications.IsSet() {
		return nil, nil, ErrInput
	}
	if probe, err := applicationFields(settings); err != nil || (len(probe) == 0 && input.PairedApplications == nil) {
		return nil, nil, ErrInput
	}
	before, read, err := s.applications.Get(ctx, identity, "Set-CsTeamsComplianceRecordingApplication")
	result := &UpdateResult[Application]{Before: before, ReadResponse: read}
	if err != nil {
		return result, read, err
	}
	if before == nil || before.Identity == nil {
		return result, read, client.ErrDecode
	}
	if input.PairedApplications != nil {
		settings.ComplianceRecordingPairedApplications = policy.Set(applyEdit(before.ComplianceRecordingPairedApplications, *input.PairedApplications, func(v *PairedApplication) *string { return v.ID }))
	}
	result.PatchInvoked = true
	result.After, read, err = s.PatchApplicationByIdentity(ctx, *before.Identity, settings, input.Controls)
	result.HTTPAccepted = read != nil && read.StatusCode() >= 200 && read.StatusCode() < 300
	return result, read, err
}

// UpdateApplicationByID is UpdateApplicationByIdentity for "{parent}/{id}".
func (s *Service) UpdateApplicationByID(ctx context.Context, parent, id string, input ApplicationUpdate) (*UpdateResult[Application], *resty.Response, error) {
	identity, err := applicationIdentity(parent, id)
	if err != nil {
		return nil, nil, err
	}
	return s.UpdateApplicationByIdentity(ctx, identity, input)
}

// DeleteApplicationByIdentity is Remove-CsTeamsComplianceRecordingApplication.
func (s *Service) DeleteApplicationByIdentity(ctx context.Context, identity string, controls policy.Controls) (*Application, *resty.Response, error) {
	if err := s.ready(ctx); err != nil {
		return nil, nil, err
	}
	body := map[string]json.RawMessage{}
	if err := policyrp.Identity(body, identity, policy.Field[string]{}, controls); err != nil {
		return nil, nil, err
	}
	return s.applications.Write(ctx, identity, "Remove-CsTeamsComplianceRecordingApplication", http.MethodDelete, body)
}

// DeleteApplicationByID removes the application id under the parent policy.
func (s *Service) DeleteApplicationByID(ctx context.Context, parent, id string, controls policy.Controls) (*Application, *resty.Response, error) {
	identity, err := applicationIdentity(parent, id)
	if err != nil {
		return nil, nil, err
	}
	return s.DeleteApplicationByIdentity(ctx, identity, controls)
}
