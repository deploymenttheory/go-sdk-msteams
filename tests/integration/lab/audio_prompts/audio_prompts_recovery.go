package audiopromptsacc

import (
	"context"
	"regexp"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	audio "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/audio_prompts"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

var audioID = regexp.MustCompile(`^[0-9a-f]{32}$`)

type audioCleanup struct{ service *audio.Service }

// ownedAudioFixture accepts only journal fixture names for audio uploads.
func ownedAudioFixture(name string) bool {
	return strings.HasPrefix(name, "msteams-sdk-") && strings.HasSuffix(name, "-audio")
}

func audioFileName(fixture string) string { return fixture + ".wav" }

func ownedAudioFile(value *audio.AudioFile, fixture string) bool {
	return value != nil && value.OriginalFilename != nil && *value.OriginalFilename == audioFileName(fixture) &&
		value.ApplicationID != nil && *value.ApplicationID == audio.TenantGlobal
}

// Resolve reconciles a planned upload by its unique original file name.
func (h audioCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	if !ownedAudioFixture(entry.Fixture) {
		return "", lab.ErrUnresolved
	}
	values, _, err := h.service.List(ctx, audio.TenantGlobal)
	if err != nil {
		return "", err
	}
	id := ""
	for _, value := range values {
		if value.OriginalFilename == nil || *value.OriginalFilename != audioFileName(entry.Fixture) {
			continue
		}
		if id != "" || !ownedAudioFile(value, entry.Fixture) || value.ID == nil || !audioID.MatchString(*value.ID) {
			return "", lab.ErrUnresolved
		}
		id = *value.ID
	}
	if id == "" {
		// An empty list cannot prove a lost upload never happened.
		return "", lab.ErrUnresolved
	}
	return id, nil
}

// Delete marks the owned file for deletion. The service keeps a soft-deleted
// file readable and rejects a second delete, so a set deletion timestamp (or a
// 404) is the absence proof.
func (h audioCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	if !ownedAudioFixture(entry.Fixture) || !audioID.MatchString(entry.ResourceID) {
		return lab.ErrUnresolved
	}
	value, _, err := h.service.GetByID(ctx, audio.TenantGlobal, entry.ResourceID, nil)
	if labkit.IsHTTP(err, 404) {
		return nil
	}
	if err != nil {
		return err
	}
	if !ownedAudioFile(value, entry.Fixture) {
		return lab.ErrUnresolved
	}
	if !value.MarkedForDeletion() {
		if _, err := h.service.DeleteByID(ctx, entry.ResourceID); err != nil && !labkit.IsHTTP(err, 404) {
			return err
		}
	}
	return h.absent(ctx, entry)
}

func (h audioCleanup) absent(ctx context.Context, entry lab.Entry) error {
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		value, _, err := h.service.GetByID(ctx, audio.TenantGlobal, entry.ResourceID, nil)
		if labkit.IsHTTP(err, 404) {
			return true, nil
		}
		if err != nil {
			return false, err
		}
		return value.MarkedForDeletion(), nil
	})
}

func (audioCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }
