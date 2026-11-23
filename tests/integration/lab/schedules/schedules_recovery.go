package schedulesacc

import (
	"context"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	sched "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/schedules"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// RecoveryHandlers returns journal recovery handlers for the owned schedules
// these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{"schedule": scheduleCleanup{sdk.TeamsAdminAPI.Schedules}}
}

type scheduleCleanup struct{ service *sched.Service }

// ownedScheduleFixture accepts only journal fixture names for schedules.
func ownedScheduleFixture(name string) bool {
	return strings.HasPrefix(name, "msteams-sdk-") && (strings.HasSuffix(name, "-fixed") || strings.HasSuffix(name, "-weekly"))
}

func ownedSchedule(value *sched.Schedule, fixture string) bool {
	return value != nil && value.Name != nil && *value.Name == fixture && value.ID != nil && labkit.DirectoryID.MatchString(*value.ID)
}

// Resolve reconciles a planned create by its unique schedule name.
func (h scheduleCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	if !ownedScheduleFixture(entry.Fixture) {
		return "", lab.ErrUnresolved
	}
	values, _, err := h.service.List(ctx, nil)
	if err != nil {
		return "", err
	}
	id := ""
	for _, value := range values {
		if value.Name == nil || *value.Name != entry.Fixture {
			continue
		}
		if id != "" || !ownedSchedule(value, entry.Fixture) {
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

// Delete removes the owned schedule; a 404 is the absence proof.
func (h scheduleCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !ownedScheduleFixture(entry.Fixture) || !labkit.DirectoryID.MatchString(entry.ResourceID) {
		return lab.ErrUnresolved
	}
	value, _, err := h.service.GetByID(ctx, entry.ResourceID)
	if labkit.IsHTTP(err, 404) {
		return nil
	}
	if err != nil {
		return err
	}
	if !ownedSchedule(value, entry.Fixture) {
		return lab.ErrUnresolved
	}
	if _, err := h.service.DeleteByID(ctx, entry.ResourceID); err != nil && !labkit.IsHTTP(err, 404) {
		return err
	}
	return h.absent(ctx, entry)
}

func (h scheduleCleanup) absent(ctx context.Context, entry lab.Entry) error {
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		_, _, err := h.service.GetByID(ctx, entry.ResourceID)
		if labkit.IsHTTP(err, 404) {
			return true, nil
		}
		return false, err
	})
}

func (scheduleCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }
