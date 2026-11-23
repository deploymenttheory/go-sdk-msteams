package autoattendantsacc

import (
	"context"
	"errors"
	"regexp"
	"slices"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	aa "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/auto_attendants"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// RecoveryHandlers returns journal recovery handlers for the owned auto
// attendants these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"auto-attendant": attendantCleanup{sdk.TeamsAdminAPI.AutoAttendants},
	}
}

var attendantID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type attendantCleanup struct{ service *aa.Service }

// ownedAttendantFixture accepts only journal fixture names for auto attendants.
func ownedAttendantFixture(name string) bool {
	return strings.HasPrefix(name, "msteams-sdk-") && strings.HasSuffix(name, "-attendant")
}

func ownedAttendant(value *aa.AutoAttendant, fixture string) bool {
	return value != nil && value.Name != nil && *value.Name == fixture && value.ID != nil && attendantID.MatchString(*value.ID)
}

// missing reports the service's answer for an unknown identity, language or
// time zone: HTTP 404 with code 100004.
func missing(err error) bool {
	var api *client.APIError
	return errors.As(err, &api) && api != nil && api.StatusCode == 404 && slices.Contains(api.Codes, "100004")
}

// invalid reports a service-side validation rejection: HTTP 400 with code 100001.
func invalid(err error) bool {
	var api *client.APIError
	return errors.As(err, &api) && api != nil && api.StatusCode == 400 && slices.Contains(api.Codes, "100001")
}

// Resolve reconciles a planned create by its unique attendant name.
func (h attendantCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	if !ownedAttendantFixture(entry.Fixture) {
		return "", lab.ErrUnresolved
	}
	values, _, err := h.service.List(ctx, &aa.ListOptions{NameFilter: &entry.Fixture, ExcludeContent: new(true)})
	if err != nil {
		return "", err
	}
	id := ""
	for _, value := range values {
		if value.Name == nil || *value.Name != entry.Fixture {
			continue
		}
		if id != "" || !ownedAttendant(value, entry.Fixture) {
			return "", lab.ErrUnresolved
		}
		id = *value.ID
	}
	if id == "" {
		// An empty list cannot prove a lost create never happened.
		return "", lab.ErrUnresolved
	}
	return id, nil
}

// Delete removes the owned attendant after confirming its name.
func (h attendantCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !ownedAttendantFixture(entry.Fixture) || !attendantID.MatchString(entry.ResourceID) {
		return lab.ErrUnresolved
	}
	value, _, err := h.service.GetByID(ctx, entry.ResourceID)
	if missing(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !ownedAttendant(value, entry.Fixture) || *value.ID != entry.ResourceID {
		return lab.ErrUnresolved
	}
	if _, err := h.service.DeleteByID(ctx, entry.ResourceID); err != nil && !missing(err) {
		return err
	}
	return h.absent(ctx, entry)
}

func (h attendantCleanup) absent(ctx context.Context, entry lab.Entry) error {
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		_, _, err := h.service.GetByID(ctx, entry.ResourceID)
		if missing(err) {
			return true, nil
		}
		return false, err
	})
}

func (attendantCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }
