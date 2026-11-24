package teamsguestmessagingconfigurationacc

import (
	"bytes"
	"context"
	"encoding/json"
	"io"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	guestmessaging "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_guest_messaging_configuration"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

// guestMessagingSnapshot is the journaled original of the two Global settings
// the update test changes: the AllowMemes Boolean and the GiphyRatingType string.
type guestMessagingSnapshot struct {
	AllowMemes      *bool   `json:"allow_memes"`
	GiphyRatingType *string `json:"giphy_rating_type"`
}

func (s guestMessagingSnapshot) matches(v *guestmessaging.Configuration) bool {
	return v != nil && v.AllowMemes != nil && v.GiphyRatingType != nil && *v.AllowMemes == *s.AllowMemes && *v.GiphyRatingType == *s.GiphyRatingType
}

// guestMessagingConfigurationCleanup restores only the journaled Global
// AllowMemes and GiphyRatingType values through the same public SDK service.
type guestMessagingConfigurationCleanup struct{ service *guestmessaging.Service }

func (guestMessagingConfigurationCleanup) Resolve(context.Context, lab.Entry) (string, error) {
	return "", lab.ErrUnresolved
}

func (guestMessagingConfigurationCleanup) Delete(context.Context, lab.Entry) error {
	return lab.ErrUnresolved
}

func (h guestMessagingConfigurationCleanup) Restore(ctx context.Context, entry lab.Entry) error {
	var snapshot guestMessagingSnapshot
	d := json.NewDecoder(bytes.NewReader(entry.Snapshot))
	d.DisallowUnknownFields()
	if entry.ResourceID != "Global" || d.Decode(&snapshot) != nil || d.Decode(new(any)) != io.EOF || snapshot.AllowMemes == nil || snapshot.GiphyRatingType == nil || *snapshot.GiphyRatingType == "" {
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
	if _, err := h.service.UpdateGlobal(ctx, guestmessaging.UpdateRequest{AllowMemes: snapshot.AllowMemes, GiphyRatingType: snapshot.GiphyRatingType}); err != nil {
		return err
	}
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		v, _, e := h.service.GetByIdentity(ctx, "Global")
		return e == nil && snapshot.matches(v), e
	})
}
