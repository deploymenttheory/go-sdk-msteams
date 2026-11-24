package teamsguestcallingconfigurationacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	guestcalling "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_guest_calling_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// guestCallingConfigurationCleanup restores only the journaled Global
// AllowPrivateCalling Boolean through the same public SDK service.
type guestCallingConfigurationCleanup struct{ service *guestcalling.Service }

func (guestCallingConfigurationCleanup) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}

func (guestCallingConfigurationCleanup) Delete(context.Context, lab.Entry) error {
	return lab.ErrUnresolved
}

func (h guestCallingConfigurationCleanup) Restore(ctx context.Context, entry lab.Entry) error {
	var snapshot struct {
		Value *bool `json:"value"`
	}
	d := json.NewDecoder(bytes.NewReader(entry.Snapshot))
	d.DisallowUnknownFields()
	if entry.ResourceID != "Global" || d.Decode(&snapshot) != nil || d.Decode(new(any)) != io.EOF || snapshot.Value == nil {
		return lab.ErrUnresolved
	}
	// A previous restoration may have succeeded before its receipt was saved.
	before, _, err := h.service.GetByIdentity(ctx, "Global")
	if err != nil {
		return err
	}
	if before != nil && before.AllowPrivateCalling != nil && *before.AllowPrivateCalling == *snapshot.Value {
		return nil
	}
	if _, err := h.service.UpdateGlobal(ctx, guestcalling.UpdateRequest{AllowPrivateCalling: snapshot.Value}); err != nil {
		return err
	}
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		v, _, e := h.service.GetByIdentity(ctx, "Global")
		return e == nil && v != nil && v.AllowPrivateCalling != nil && *v.AllowPrivateCalling == *snapshot.Value, e
	})
}
