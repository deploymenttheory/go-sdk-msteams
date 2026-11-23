package callqueuesacc

import (
	"context"
	"errors"
	"regexp"
	"slices"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/client"
	cq "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/call_queues"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// RecoveryHandlers returns journal recovery handlers for the owned call queues
// these acceptance tests create.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{
		"call-queue": queueCleanup{sdk.TeamsAdminAPI.CallQueues},
	}
}

var queueID = regexp.MustCompile(`^[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)

type queueCleanup struct{ service *cq.Service }

// ownedQueueFixture accepts only journal fixture names for call queues.
func ownedQueueFixture(name string) bool {
	return strings.HasPrefix(name, "msteams-sdk-") && strings.HasSuffix(name, "-queue")
}

func ownedQueue(value *cq.CallQueue, fixture string) bool {
	return value != nil && value.Name != nil && *value.Name == fixture && value.Identity != nil && queueID.MatchString(*value.Identity)
}

// missingQueue reports the service's answer for an unknown or deleted
// identity: HTTP 400 with code 100001 (CallQueue_IdentityInvalid), not 404.
func missingQueue(err error) bool {
	var api *client.APIError
	return errors.As(err, &api) && api != nil && api.StatusCode == 400 && slices.Contains(api.Codes, "100001")
}

// Resolve reconciles a planned create by its unique queue name.
func (h queueCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	if !ownedQueueFixture(entry.Fixture) {
		return "", lab.ErrUnresolved
	}
	values, _, err := h.service.List(ctx, &cq.ListOptions{NameFilter: &entry.Fixture})
	if err != nil {
		return "", err
	}
	id := ""
	for _, value := range values {
		if value.Name == nil || *value.Name != entry.Fixture {
			continue
		}
		if id != "" || !ownedQueue(value, entry.Fixture) {
			return "", lab.ErrUnresolved
		}
		id = *value.Identity
	}
	if id == "" {
		// An empty list cannot prove a lost create never happened.
		return "", lab.ErrUnresolved
	}
	return id, nil
}

// Delete removes the owned queue after confirming its name.
func (h queueCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !ownedQueueFixture(entry.Fixture) || !queueID.MatchString(entry.ResourceID) {
		return lab.ErrUnresolved
	}
	value, _, err := h.service.GetByID(ctx, entry.ResourceID)
	if missingQueue(err) {
		return nil
	}
	if err != nil {
		return err
	}
	if !ownedQueue(value, entry.Fixture) || *value.Identity != entry.ResourceID {
		return lab.ErrUnresolved
	}
	if _, err := h.service.DeleteByID(ctx, entry.ResourceID); err != nil && !missingQueue(err) {
		return err
	}
	return h.absent(ctx, entry)
}

func (h queueCleanup) absent(ctx context.Context, entry lab.Entry) error {
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		_, _, err := h.service.GetByID(ctx, entry.ResourceID)
		if missingQueue(err) {
			return true, nil
		}
		return false, err
	})
}

func (queueCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }
