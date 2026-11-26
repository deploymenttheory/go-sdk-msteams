package tenantmigrationconfigurationacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	tenantmigration "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/tenant_migration_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// tenantMigrationSnapshot is the journaled original Global
// MeetingMigrationEnabled value the update test changes.
type tenantMigrationSnapshot struct {
	MeetingMigrationEnabled *bool `json:"meeting_migration_enabled"`
}

func (s tenantMigrationSnapshot) matches(v *tenantmigration.Configuration) bool {
	return v != nil && v.MeetingMigrationEnabled != nil && *v.MeetingMigrationEnabled == *s.MeetingMigrationEnabled
}

// tenantMigrationConfigurationCleanup restores only the journaled Global
// MeetingMigrationEnabled value through the same public SDK service. The update
// test only journals an enabled original, so restoration only re-enables the
// Meeting Migration Service; it never disables it.
type tenantMigrationConfigurationCleanup struct {
	service *tenantmigration.Service
}

func (tenantMigrationConfigurationCleanup) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}

func (tenantMigrationConfigurationCleanup) Delete(context.Context, lab.Entry) error {
	return lab.ErrUnresolved
}

func (h tenantMigrationConfigurationCleanup) Restore(ctx context.Context, entry lab.Entry) error {
	var snapshot tenantMigrationSnapshot
	d := json.NewDecoder(bytes.NewReader(entry.Snapshot))
	d.DisallowUnknownFields()
	if entry.ResourceID != "Global" || d.Decode(&snapshot) != nil || d.Decode(new(any)) != io.EOF || snapshot.MeetingMigrationEnabled == nil || !*snapshot.MeetingMigrationEnabled {
		return lab.ErrUnresolved
	}
	// A previous restoration may have succeeded before its receipt was saved.
	before, _, err := h.service.GetByIdentity(ctx, "Global")
	if err != nil {
		return err
	}
	if snapshot.matches(before) {
		return nil
	}
	if _, err := h.service.UpdateGlobal(ctx, tenantmigration.UpdateRequest{MeetingMigrationEnabled: snapshot.MeetingMigrationEnabled}); err != nil {
		return err
	}
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		v, _, e := h.service.GetByIdentity(ctx, "Global")
		return e == nil && snapshot.matches(v), e
	})
}
