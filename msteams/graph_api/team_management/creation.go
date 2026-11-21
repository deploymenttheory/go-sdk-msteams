package teammanagement

import (
	"context"
	"crypto/rand"
	"encoding/hex"
	"encoding/json"
	"errors"
	"net/http"
	"strconv"
	"strings"
	"time"

	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	directoryuser "github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/directory_user"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/group"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/graph_api/team"
	"github.com/go-resty/resty/v2"
)

func normalizeCreationOptions(value CreationOptions) (CreationOptions, error) {
	if value.Timeout < 0 || value.MaxAttempts < 0 || value.Interval < 0 || value.CleanupTimeout < 0 {
		return value, client.ErrRequest
	}
	if value.Timeout == 0 {
		value.Timeout = 5 * time.Minute
	}
	if value.MaxAttempts == 0 {
		value.MaxAttempts = 4
	}
	if value.Interval == 0 {
		value.Interval = 10 * time.Second
	}
	if value.CleanupTimeout == 0 {
		value.CleanupTimeout = 30 * time.Second
	}
	return value, nil
}

func groupTemplate(value GroupTemplate) (GroupTemplate, error) {
	for _, known := range []GroupTemplate{OrdinaryGroup, EducationClass, EducationPLC} {
		if strings.EqualFold(string(value), string(known)) {
			return known, nil
		}
	}
	return "", client.ErrRequest
}

func creationVisibility(value CreateRequest) (group.Visibility, error) {
	if value.Template == EducationClass {
		if value.Visibility != nil && strings.TrimSpace(string(*value.Visibility)) != "" {
			return "", client.ErrRequest
		}
		return group.HiddenMembership, nil
	}
	if value.Visibility == nil {
		return group.Private, nil
	}
	for _, known := range []group.Visibility{group.Private, group.Public, group.HiddenMembership} {
		if strings.EqualFold(string(*value.Visibility), string(known)) {
			return known, nil
		}
	}
	return "", client.ErrRequest
}

func groupCreation(value CreateRequest, alias, owner string) group.CreateRequest {
	visibility, _ := creationVisibility(value) // Validated before any lookup/write.
	request := group.CreateRequest{Properties: group.Properties{DisplayName: &value.DisplayName, Description: value.Description, Classification: value.Classification, MailNickname: &alias, Visibility: &visibility}, GroupTypes: []string{"Unified"}, MailEnabled: true, SecurityEnabled: false, OwnerIDs: []string{owner}, MemberIDs: []string{owner}}
	flags := []string{"ExchangeProvisioningFlags:3552"}
	switch value.Template {
	case EducationClass:
		flags = []string{"ExchangeProvisioningFlags:461", "classAssignments"}
		behaviors := []string{
			"appRoleForSite:8f348934-64be-4bb2-bc16-c54c96789f43:fullcontrol",
			"appRoleForSite:2d4d3d8e-2be3-4bef-9f87-7875a61c29de:fullcontrol",
			"appRoleForSite:13291f5a-59ac-4c59-b0fa-d1632e8f3292:fullcontrol",
			"appRoleForSite:c9a559d2-7aab-4f13-a6ed-e7e9c52aec87:fullcontrol",
			"appRoleForSite:22d27567-b3f0-4dc2-9ec2-46ed368ba538:fullcontrol",
		}
		request.ResourceBehaviorOptions = &behaviors
		request.Extensions = map[string]json.RawMessage{"extension_fe2174665583431c953114ff7268b7b3_Education_ObjectType": json.RawMessage(`"Section"`)}
	case EducationPLC:
		flags = []string{"PLC"}
		behaviors := []string{"appRoleForSite:13291f5a-59ac-4c59-b0fa-d1632e8f3292:fullcontrol", "appRoleForSite:2d4d3d8e-2be3-4bef-9f87-7875a61c29de:fullcontrol"}
		request.ResourceBehaviorOptions = &behaviors
	}
	request.CreationOptions = &flags
	return request
}

func (s *Service) mailNickname(ctx context.Context, value CreateRequest, result *CreationResult) error {
	if value.MailNickname != nil {
		result.MailNickname = *value.MailNickname
		return nil
	}
	result.GeneratedMailNickname = true
	for range 10 {
		var bytes [3]byte
		if _, err := rand.Read(bytes[:]); err != nil {
			return err
		}
		alias := "msteams_" + hex.EncodeToString(bytes[:])
		result.MailNickname = alias
		list, err := s.groups.ListAll(ctx, group.ListOptions{Filter: "groupTypes/any(c:c eq 'Unified') and mailNickname eq '" + alias + "'", Select: []string{"id"}}, client.PagingOptions{MaxPages: 16})
		result.record(CheckMailNickname, list.Response, err)
		if err != nil {
			return err
		}
		if len(list.Items) == 0 {
			return nil
		}
	}
	return ErrMailNicknameLimit
}

func graphStatus(err error, status int) bool {
	var apiError *client.APIError
	return errors.As(err, &apiError) && apiError.Backend == client.Graph && apiError.StatusCode == status
}

func waitPropagation(ctx context.Context, interval time.Duration, response *resty.Response) error {
	if response != nil {
		header := strings.TrimSpace(response.Header().Get("Retry-After"))
		if seconds, err := strconv.ParseUint(header, 10, 63); err == nil {
			maximum := uint64((1<<63 - 1) / int64(time.Second))
			if seconds > maximum {
				interval = time.Duration(1<<63 - 1)
			} else {
				interval = max(interval, time.Duration(seconds)*time.Second)
			}
		} else if date, err := http.ParseTime(header); err == nil {
			interval = max(interval, time.Until(date))
		}
	}
	timer := time.NewTimer(interval)
	defer timer.Stop()
	select {
	case <-ctx.Done():
		return ctx.Err()
	case <-timer.C:
		return nil
	}
}

func (s *Service) teamify(ctx context.Context, result *CreationResult, settings team.Settings, options CreationOptions) (bool, error) {
	var response *resty.Response
	var err error
	for attempt := 0; attempt < options.MaxAttempts; attempt++ {
		var value *group.Group
		value, response, err = s.groups.GetByID(ctx, result.ID, group.ReadOptions{})
		if err == nil && !strings.EqualFold(value.ID, result.ID) {
			err = client.ErrDecode
		}
		result.record(GroupAvailability, response, err)
		if err == nil {
			break
		}
		if !graphStatus(err, 404) || attempt+1 == options.MaxAttempts {
			return false, err
		}
		if err := waitPropagation(ctx, options.Interval, response); err != nil {
			return false, err
		}
	}
	for attempt := 0; attempt < options.MaxAttempts; attempt++ {
		result.teamificationRejected = false
		result.Teamified, response, err = s.teams.CreateFromGroup(ctx, result.ID, completeSettings(settings))
		if err == nil && !strings.EqualFold(result.Teamified.ID, result.ID) {
			err = client.ErrDecode
		}
		result.record(TeamifyGroup, response, err)
		result.teamificationRejected = rejectedTeamification(err)
		if err == nil {
			return true, nil
		}
		if !graphStatus(err, 404) || attempt+1 == options.MaxAttempts {
			return true, err
		}
		if err := waitPropagation(ctx, options.Interval, response); err != nil {
			return true, err
		}
	}
	return true, err
}

func rejectedTeamification(err error) bool {
	var apiError *client.APIError
	return errors.As(err, &apiError) && apiError.Backend == client.Graph && apiError.StatusCode >= 400 && apiError.StatusCode < 500 && apiError.StatusCode != 408
}

func (s *Service) compensate(ctx context.Context, result *CreationResult, attempted bool, options CreationOptions) {
	ctx, cancel := context.WithTimeout(context.WithoutCancel(ctx), options.CleanupTimeout)
	defer cancel()
	if attempted {
		value, response, err := s.teams.GetByID(ctx, result.ID)
		if err == nil && !strings.EqualFold(value.ID, result.ID) {
			err = client.ErrDecode
		}
		result.Cleanup.ReconciliationResponse = response
		result.Cleanup.Team = value
		if err == nil {
			result.record(ReconcileTeam, response, nil)
			result.Cleanup.Disposition = CleanupTeamExists
			return
		}
		if !graphStatus(err, 404) || !result.teamificationRejected {
			result.record(ReconcileTeam, response, err)
			result.Cleanup.Disposition = CleanupUnresolved
			result.Cleanup.Error = errors.Join(ErrCreationReconciliation, err)
			return
		}
		// A received rejection plus this absence observation allows compensation
		// of our new group. A lone 404 after an ambiguous write does not.
		result.record(ReconcileTeam, response, nil)
	}
	response, err := s.groups.DeleteByID(ctx, result.ID)
	result.record(DeleteCreatedGroup, response, err)
	result.Cleanup.Response, result.Cleanup.Error = response, err
	result.Cleanup.Disposition = CleanupDeleted
	if err != nil {
		result.Cleanup.Disposition = CleanupDeleteUnconfirmed
	}
}

func (s *Service) creationFailure(ctx context.Context, result CreationResult, err error, attempted, retain bool, options CreationOptions) (CreationResult, error) {
	result.PrimaryError, result.PrimaryResponse = err, result.Response
	switch {
	case !result.CreatedGroupOwned:
		if result.ID != "" {
			result.Cleanup.Disposition = CleanupSuppliedGroup
		} else if result.CreatedGroupResponse != nil || len(result.Stages) != 0 && result.Stages[len(result.Stages)-1].Action == CreateGroup {
			if !rejectedTeamification(err) {
				result.Cleanup.Disposition = CleanupUnresolved
				result.Cleanup.Error = ErrCreationReconciliation
			}
		}
	case retain:
		result.Cleanup.Disposition = CleanupRetainedByRequest
	default:
		s.compensate(ctx, &result, attempted, options)
	}
	return result, errors.Join(err, result.Cleanup.Error)
}

// Create implements New-Team/CreateTeam as owner resolution, group creation,
// bounded propagation, teamification and paired readback. Only this call's newly
// created group can be compensated. Ambiguous writes are never blindly replayed
// or converted into absence; failed cleanup is returned with the primary error.
func (s *Service) Create(ctx context.Context, value CreateRequest) (CreationResult, error) {
	result := CreationResult{Cleanup: CreationCleanup{Disposition: CleanupNotRequired}}
	ctx, stop, err := s.start(ctx)
	if err != nil {
		return result, err
	}
	defer stop()
	value.Options, err = normalizeCreationOptions(value.Options)
	if err != nil {
		return result, err
	}
	value.Template, err = groupTemplate(value.Template)
	if err != nil {
		return result, err
	}
	if _, err := creationVisibility(value); err != nil {
		return result, err
	}
	if strings.TrimSpace(value.DisplayName) == "" || value.MailNickname != nil && strings.TrimSpace(*value.MailNickname) == "" {
		return result, client.ErrRequest
	}
	ctx, cancel := context.WithTimeout(ctx, value.Options.Timeout)
	defer cancel()
	resolved, response, err := s.users.Resolve(ctx, value.Owner, directoryuser.ResolveOptions{AllowCurrent: true, MaxPages: 16})
	result.record(ResolveOwner, response, err)
	if err != nil {
		return s.creationFailure(ctx, result, err, false, value.RetainCreatedGroup, value.Options)
	}
	result.OwnerID, result.OwnerSource = resolved.ID, resolved.Source
	if err := s.mailNickname(ctx, value, &result); err != nil {
		return s.creationFailure(ctx, result, err, false, value.RetainCreatedGroup, value.Options)
	}
	created, response, err := s.groups.Create(ctx, groupCreation(value, result.MailNickname, result.OwnerID))
	if err == nil && !directoryID.MatchString(created.ID) {
		err = client.ErrDecode
	}
	result.record(CreateGroup, response, err)
	result.CreatedGroupResponse = response
	if err != nil {
		return s.creationFailure(ctx, result, err, false, value.RetainCreatedGroup, value.Options)
	}
	result.ID, result.CreatedGroupOwned, result.CreatedGroup = created.ID, true, created
	attempted, err := s.teamify(ctx, &result, value.Settings, value.Options)
	if err != nil {
		return s.creationFailure(ctx, result, err, attempted, value.RetainCreatedGroup, value.Options)
	}
	if err := s.read(ctx, &result.Result); err != nil {
		// A validated 201 already established team creation. Failed enrichment
		// cannot make a successful team eligible for compensation.
		result.PrimaryError, result.PrimaryResponse = err, result.Response
		result.Cleanup.Disposition = CleanupTeamExists
		return result, err
	}
	return result, nil
}

// MigrateGroup implements New-Team/MigrateGroup. It exposes only the bound group
// ID and wait controls, not CreateTeam-only owner/template/settings parameters.
// The supplied group's ownership is never inferred and it is never deleted.
func (s *Service) MigrateGroup(ctx context.Context, id string, options CreationOptions) (CreationResult, error) {
	result := CreationResult{Result: Result{ID: id}, Cleanup: CreationCleanup{Disposition: CleanupSuppliedGroup}}
	ctx, stop, err := s.begin(ctx, id)
	if err != nil {
		return result, err
	}
	defer stop()
	options, err = normalizeCreationOptions(options)
	if err != nil {
		return result, err
	}
	ctx, cancel := context.WithTimeout(ctx, options.Timeout)
	defer cancel()
	_, err = s.teamify(ctx, &result, team.Settings{}, options)
	if err == nil {
		err = s.read(ctx, &result.Result)
	}
	if err != nil {
		result.PrimaryError, result.PrimaryResponse = err, result.Response
	}
	return result, err
}
