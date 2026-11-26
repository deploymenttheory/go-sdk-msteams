package teamsguestmeetingconfigurationacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	guestmeeting "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_guest_meeting_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// guestMeetingConfigurationCleanup restores only the journaled Global
// AllowMeetNow Boolean through the same public SDK service.
type guestMeetingConfigurationCleanup struct{ service *guestmeeting.Service }

func (guestMeetingConfigurationCleanup) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}

func (guestMeetingConfigurationCleanup) Delete(context.Context, lab.Entry) error {
	return lab.ErrUnresolved
}

func (h guestMeetingConfigurationCleanup) Restore(ctx context.Context, entry lab.Entry) error {
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
	if before != nil && before.AllowMeetNow != nil && *before.AllowMeetNow == *snapshot.Value {
		return nil
	}
	if _, _, err := h.service.UpdateByIdentity(ctx, "Global", guestmeeting.UpdateRequest{Settings: guestmeeting.Settings{AllowMeetNow: policy.Set(*snapshot.Value)}}); err != nil {
		return err
	}
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		v, _, e := h.service.GetByIdentity(ctx, "Global")
		return e == nil && v != nil && v.AllowMeetNow != nil && *v.AllowMeetNow == *snapshot.Value, e
	})
}
