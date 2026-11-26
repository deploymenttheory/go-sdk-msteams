package teamsmeetingbroadcastconfigurationacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	broadcast "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_meeting_broadcast_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// meetingBroadcastSnapshot is the journaled original Global state. Only the
// two qualified settings are recorded; no SDN provider value or token is
// journaled, read for restoration or written.
type meetingBroadcastSnapshot struct {
	Allow      *bool   `json:"allow_sdn_provider"`
	SupportURL *string `json:"support_url"`
}

func (s meetingBroadcastSnapshot) matches(v *broadcast.Configuration) bool {
	return v != nil && v.AllowSdnProviderForBroadcastMeeting != nil && v.SupportURL != nil &&
		*v.AllowSdnProviderForBroadcastMeeting == *s.Allow && *v.SupportURL == *s.SupportURL
}

// meetingBroadcastConfigurationCleanup restores only the journaled Global
// AllowSdnProviderForBroadcastMeeting Boolean and SupportURL string through the
// same public SDK service.
type meetingBroadcastConfigurationCleanup struct{ service *broadcast.Service }

func (meetingBroadcastConfigurationCleanup) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}

func (meetingBroadcastConfigurationCleanup) Delete(context.Context, lab.Entry) error {
	return lab.ErrUnresolved
}

func (h meetingBroadcastConfigurationCleanup) Restore(ctx context.Context, entry lab.Entry) error {
	var snapshot meetingBroadcastSnapshot
	d := json.NewDecoder(bytes.NewReader(entry.Snapshot))
	d.DisallowUnknownFields()
	if entry.ResourceID != "Global" || d.Decode(&snapshot) != nil || d.Decode(new(any)) != io.EOF || snapshot.Allow == nil || snapshot.SupportURL == nil {
		return lab.ErrUnresolved
	}
	// A previous restoration may have succeeded before its receipt was saved.
	before, _, err := h.service.GetByIdentity(ctx, "Global", broadcast.ReadOptions{})
	if err != nil {
		return err
	}
	if snapshot.matches(before) {
		return nil
	}
	if _, err := h.service.UpdateGlobal(ctx, broadcast.UpdateRequest{AllowSdnProviderForBroadcastMeeting: snapshot.Allow, SupportURL: snapshot.SupportURL}); err != nil {
		return err
	}
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		v, _, e := h.service.GetByIdentity(ctx, "Global", broadcast.ReadOptions{})
		return e == nil && snapshot.matches(v), e
	})
}
