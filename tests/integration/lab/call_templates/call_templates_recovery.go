package calltemplatesacc

import (
	"context"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	ct "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/call_templates"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// templateCleanup deletes one owned template of any type. Templates are
// matched only by the journal's unique msteams-sdk-* fixture name.
type templateCleanup[T any] struct {
	service *ct.Templates[T]
	name    func(*T) *string
	id      func(*T) *string
}

func ownedFixture(name string) bool { return strings.HasPrefix(name, "msteams-sdk-") }

func (h templateCleanup[T]) owned(value *T, fixture string) bool {
	return value != nil && h.name(value) != nil && *h.name(value) == fixture
}

// Resolve reconciles a planned create by its unique name. The list route
// returns the whole collection, so no match proves the create never landed.
func (h templateCleanup[T]) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	if !ownedFixture(entry.Fixture) {
		return "", lab.ErrUnresolved
	}
	values, _, err := h.service.List(ctx)
	if err != nil {
		return "", err
	}
	id := ""
	for _, value := range values {
		if !h.owned(value, entry.Fixture) {
			continue
		}
		if id != "" || h.id(value) == nil || !labkit.DirectoryID.MatchString(*h.id(value)) {
			return "", lab.ErrUnresolved
		}
		id = *h.id(value)
	}
	return id, nil
}

// Delete removes the owned template after re-checking its name, then
// observes a 404.
func (h templateCleanup[T]) Delete(ctx context.Context, entry lab.Entry) error {
	if !ownedFixture(entry.Fixture) || !labkit.DirectoryID.MatchString(entry.ResourceID) {
		return lab.ErrUnresolved
	}
	value, _, err := h.service.GetByID(ctx, entry.ResourceID)
	if labkit.IsHTTP(err, 404) {
		return nil
	}
	if err != nil {
		return err
	}
	if !h.owned(value, entry.Fixture) {
		return lab.ErrUnresolved
	}
	if _, err := h.service.DeleteByID(ctx, entry.ResourceID); err != nil && !labkit.IsHTTP(err, 404) {
		return err
	}
	return h.absent(ctx, entry.ResourceID)
}

func (h templateCleanup[T]) absent(ctx context.Context, id string) error {
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		_, _, err := h.service.GetByID(ctx, id)
		if labkit.IsHTTP(err, 404) {
			return true, nil
		}
		return false, err
	})
}

func (templateCleanup[T]) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }
