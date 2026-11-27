package teamscustombannertextacc

import (
	"context"
	"crypto/sha256"
	"fmt"
	"regexp"
	"strings"

	"github.com/deploymenttheory/go-sdk-msteams/internal/lab"
	"github.com/deploymenttheory/go-sdk-msteams/msteams"
	"github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/policy"
	resource "github.com/deploymenttheory/go-sdk-msteams/msteams/teams_admin_api/teams_custom_banner_text"
	"github.com/deploymenttheory/go-sdk-msteams/tests/integration/lab/internal/labkit"
)

var ownedRecordName = regexp.MustCompile(`^msteams-sdk-[a-z2-7]{26,64}-banner-text$`)

// ownedRecordID is stable across recovery of an interrupted submission. The
// journal's random fixture name and the resource description prove ownership.
func ownedRecordID(fixture string) (string, error) {
	if !ownedRecordName.MatchString(fixture) {
		return "", lab.ErrUnresolved
	}
	sum := sha256.Sum256([]byte(fixture))
	sum[6] = sum[6]&0x0f | 0x40
	sum[8] = sum[8]&0x3f | 0x80
	return fmt.Sprintf("%x-%x-%x-%x-%x", sum[:4], sum[4:6], sum[6:8], sum[8:10], sum[10:16]), nil
}

type ownedRecordCleanup struct{ service *resource.Service }

// RecoveryHandlers reconciles only the UUID and ownership marker written by
// these tests. No Global setting, assignment or unrelated record is changed.
func RecoveryHandlers(sdk *msteams.Client) map[string]lab.Handler {
	return map[string]lab.Handler{"banner-text": ownedRecordCleanup{sdk.TeamsAdminAPI.TeamsCustomBannerTexts}}
}

func (h ownedRecordCleanup) Resolve(ctx context.Context, entry lab.Entry) (string, error) {
	id, err := ownedRecordID(entry.Fixture)
	if err != nil {
		return "", err
	}
	identity := "Global/" + id
	if entry.ResourceID != "" && !strings.EqualFold(entry.ResourceID, identity) {
		return "", lab.ErrUnresolved
	}
	value, _, err := h.service.GetByID(ctx, id)
	if labkit.IsHTTP(err, 404) {
		// A denied or inconsistent collection read cannot prove absence.
		values, _, listErr := h.service.List(ctx)
		if listErr != nil {
			return "", listErr
		}
		for _, v := range values {
			if v == nil || v.Identity == nil {
				return "", lab.ErrUnresolved
			}
			if strings.EqualFold(*v.Identity, identity) || v.Description != nil && *v.Description == entry.Fixture {
				return "", lab.ErrUnresolved
			}
		}
		return "", nil
	}
	if err != nil {
		return "", err
	}
	if value == nil || value.Identity == nil || !strings.EqualFold(*value.Identity, identity) || value.Description == nil || *value.Description != entry.Fixture {
		return "", lab.ErrUnresolved
	}
	return identity, nil
}

func (h ownedRecordCleanup) Delete(ctx context.Context, entry lab.Entry) error {
	identity, err := h.Resolve(ctx, entry)
	if err != nil || identity == "" {
		return err
	}
	if _, _, err := h.service.DeleteByIdentity(ctx, identity, policy.Controls{}); err != nil && !labkit.IsHTTP(err, 404) {
		return err
	}
	return labkit.ObserveUntil(ctx, func(ctx context.Context) (bool, error) {
		resolved, err := h.Resolve(ctx, entry)
		return resolved == "", err
	})
}

func (ownedRecordCleanup) Restore(context.Context, lab.Entry) error { return lab.ErrUnresolved }
